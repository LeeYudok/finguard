// 룰 회귀 픽스처의 기대값 마커 — 문법 정의와 픽스처 트리 순회.
//
// **이 파일에는 빌드 태그가 없다.** 기본 `go test` 와 `-tags semgrep_integration`
// 양쪽에서 컴파일되므로, 태그 뒤에 있는 integration_test.go 도 여기 정의를 그대로 쓴다.
// semgrep 바이너리가 필요 없는 검사(마커 문법·순회·대조군)는 전부 여기 둔다.
package scanner

import (
	"bytes"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// ── 기대값 마커 — 단일 정의 ──
//
// 픽스처는 검출이 기대되는 줄 **바로 위**에 마커 주석을 단다 (#44):
//
//	# EXPECT: finguard.python.cleartext-websocket
//	REALTIME_FEED_URL = "ws://ops.example-broker.co.kr:21000"
//
// 기대값이 테스트가 아니라 픽스처 안에 있으므로 블록을 어디에 추가하든 따라 움직인다.
// 라인 번호를 테스트에 하드코딩하던 구조는 픽스처를 건드리는 모든 PR 을 서로
// 충돌시켰고, 병합 후 조용히 어긋나기까지 했다.
//
// 예전에는 같은 문법을 integration_test.go 에도 복제해 두었다. 둘이 어긋나면 픽스처
// 마커가 조용히 무시되고 통합 테스트가 "기대 0건 · 검출 0건" 으로 통과해버렸다
// (#60·#61). 정의를 하나로 합쳐 그 어긋남 자체를 구조적으로 없앤다.
//
// 지원하는 주석 형태:
//
//	# EXPECT: <룰ID>                  파이썬·셸·YAML·properties
//	// EXPECT: <룰ID>                  Go·Java·Kotlin·Swift·TS
//	<!-- EXPECT: <룰ID> -->            XML·plist (#61)
//
// 룰ID 문자는 `[A-Za-z0-9_.-]` 로 제한한다. XML 닫기(`-->`)가 룰ID 에 섞여 들어가는
// 것을 막고, 마커 뒤에 설명을 덧붙이면 매칭 자체가 실패하게 하는 것이 목적이다.
// 조용히 무시되는 대신 통합 테스트가 "마커 없는데 검출됨(오탐 회귀)" 으로 즉시
// 실패해야 작성자가 알아챈다.
var expectMarker = regexp.MustCompile(
	`(?:(?:#|//)\s*EXPECT:\s*([\w.-]+)|<!--\s*EXPECT:\s*([\w.-]+)\s*-->)\s*$`)

// expectMarkerRuleID 는 한 줄에서 마커의 룰ID 를 뽑는다. 마커가 아니면 "" 를 준다.
//
// 정규식이 주석 형태별로 캡처 그룹을 따로 갖기 때문에 호출부가 그룹 번호를 알 필요가
// 없도록 여기서 흡수한다.
func expectMarkerRuleID(line string) string {
	m := expectMarker.FindStringSubmatch(line)
	if m == nil {
		return ""
	}
	if m[1] != "" {
		return m[1]
	}
	return m[2]
}

func TestExpectMarkerSyntax(t *testing.T) {
	tests := []struct {
		name string
		line string
		want string // 빈 문자열이면 매칭되지 않아야 한다
	}{
		// 매칭돼야 하는 형태 — 줄 주석
		{"파이썬/셸 주석", "# EXPECT: finguard.python.weak-hash", "finguard.python.weak-hash"},
		{"TS/Java 주석", "// EXPECT: finguard.ts.eval", "finguard.ts.eval"},
		{"들여쓰기", "    # EXPECT: finguard.python.eval-exec", "finguard.python.eval-exec"},
		{"콜론 뒤 공백 여러 개", "#   EXPECT:   finguard.go.weak-crypto", "finguard.go.weak-crypto"},
		{"주석기호 뒤 공백 없음", "#EXPECT: finguard.swift.http-url", "finguard.swift.http-url"},
		{"줄 끝 공백", "# EXPECT: finguard.java.xxe   ", "finguard.java.xxe"},

		// 매칭돼야 하는 형태 — XML/plist 주석 (#61)
		{"XML 한 줄 주석", "<!-- EXPECT: finguard.plist.ats-arbitrary-loads -->", "finguard.plist.ats-arbitrary-loads"},
		{"XML 들여쓰기", "\t\t<!-- EXPECT: finguard.config.plist-secret -->", "finguard.config.plist-secret"},
		{"XML 닫기 앞 공백 없음", "<!-- EXPECT: finguard.config.debug-logging-->", "finguard.config.debug-logging"},
		{"XML 여는 태그 뒤 공백 없음", "<!--EXPECT: finguard.plist.ats-partial-exception-->", "finguard.plist.ats-partial-exception"},
		{"XML 줄 끝 공백", "<!-- EXPECT: finguard.java.xxe -->  ", "finguard.java.xxe"},
		{"XML 요소 뒤에 붙은 마커", "<key>X</key> <!-- EXPECT: finguard.config.plist-secret -->", "finguard.config.plist-secret"},

		// 매칭되면 안 되는 형태 — 줄 주석
		{"평범한 주석", "# 이 줄은 API 키를 담는다", ""},
		{"EXPECT 언급만", "# 이 룰은 EXPECT 마커로 검증한다", ""},
		{"콜론 없음", "# EXPECT finguard.python.weak-hash", ""},
		{"룰ID 없음", "# EXPECT:", ""},
		{"룰ID 자리가 공백뿐", "# EXPECT:    ", ""},
		{"주석 아닌 코드", `EXPECT = "finguard.python.weak-hash"`, ""},
		{"뒤에 다른 내용이 붙음", "# EXPECT: finguard.python.weak-hash 이유: 테스트", ""},

		// 매칭되면 안 되는 형태 — XML/plist (#61)
		{"XML 마커 뒤 설명", "<!-- EXPECT: finguard.plist.ats-arbitrary-loads 이유: ATS -->", ""},
		{"XML 닫기 뒤 내용", "<!-- EXPECT: finguard.plist.ats-arbitrary-loads --><key>X</key>", ""},
		{"XML 주석 미종료", "<!-- EXPECT: finguard.plist.ats-arbitrary-loads", ""},
		{"XML 평범한 주석", "<!-- 대조군 — INFO 는 대상이 아니다 -->", ""},
		{"XML 룰ID 없음", "<!-- EXPECT: -->", ""},
		{"XML 여는 기호 앞에 설명", "<!-- 참고 EXPECT: finguard.plist.ats-arbitrary-loads -->", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := expectMarkerRuleID(tc.line)
			if got != tc.want {
				t.Errorf("룰ID 추출 결과 %q, 기대 %q (입력 %q)", got, tc.want, tc.line)
			}
		})
	}
}

// TestExpectMarkerRejectsTrailingText 는 "뒤에 다른 내용이 붙음" 케이스의 의도를 못박는다.
//
// 마커 줄에 설명을 덧붙이면 조용히 무시되는 대신 아예 매칭되지 않아야 한다.
// 그래야 통합 테스트가 "마커 없는데 검출됨(오탐 회귀)" 으로 즉시 실패해 작성자가 알아챈다.
// XML 주석도 같은 규칙을 따른다 — 닫기(`-->`) 앞이든 뒤든 설명이 붙으면 거부된다 (#61).
func TestExpectMarkerRejectsTrailingText(t *testing.T) {
	lines := []string{
		"# EXPECT: finguard.python.weak-hash — MD5 사용",
		"// EXPECT: finguard.python.weak-hash — MD5 사용",
		"<!-- EXPECT: finguard.python.weak-hash — MD5 사용 -->",
		"<!-- EXPECT: finguard.python.weak-hash --> MD5 사용",
	}
	for _, l := range lines {
		if got := expectMarkerRuleID(l); got != "" {
			t.Errorf("마커 뒤 설명을 허용하면 룰ID 에 설명이 섞여 들어간다 — 거부돼야 한다: %q → %q", l, got)
		}
	}
}

// TestExpectMarkerXMLCloserNotInRuleID 는 트레일링 `-->` 가 룰ID 에 섞이지 않음을 못박는다 (#61).
//
// 예전 문법(`(\S+)\s*$`)을 그대로 두고 `<!--` 만 주석 기호에 추가하면 `\S+` 가 `-->` 까지
// 삼켜 룰ID 가 "finguard.plist.ats-arbitrary-loads-->" 가 된다. 그러면 마커는 매칭되지만
// 어떤 검출과도 일치하지 않아 미탐 회귀로 오인된다.
func TestExpectMarkerXMLCloserNotInRuleID(t *testing.T) {
	const want = "finguard.plist.ats-arbitrary-loads"
	for _, l := range []string{
		"<!-- EXPECT: " + want + " -->",
		"<!-- EXPECT: " + want + "-->",
	} {
		if got := expectMarkerRuleID(l); got != want {
			t.Errorf("룰ID 추출 결과 %q, 기대 %q (입력 %q)", got, want, l)
		}
	}
}

// fixturesDir 는 룰 회귀 픽스처 디렉터리다.
//
// rules/ 바깥에 두는 것이 필수다 (#43) — semgrep 은 `--config <디렉터리>` 하위의 모든
// .yml/.yaml 을 룰 파일로 재귀 파싱하므로, rules/ 안에 yaml 픽스처가 하나라도 있으면
// 룰셋 전체가 "0 rule(s)" 로 로드 실패한다.
func fixturesDir(t *testing.T) string {
	t.Helper()
	p, err := filepath.Abs("../../testdata/rule-fixtures")
	if err != nil {
		t.Fatalf("픽스처 경로 확인 실패: %v", err)
	}
	return p
}

// expectation 은 픽스처 한 줄에 기대되는 검출이다.
type expectation struct {
	file   string // 픽스처 루트 기준 상대경로 (슬래시 구분)
	line   int    // 검출이 기대되는 줄 (마커 다음 줄)
	ruleID string // 기대 룰 ID (semgrep 접두어 없는 suffix)
}

func (e expectation) String() string {
	return e.file + ":" + itoa(e.line) + " " + e.ruleID
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

// walkFixtures 는 픽스처 트리를 **재귀로** 훑어 텍스트 파일마다 fn 을 부른다.
//
// 재귀가 핵심이다 (#60). 스캔(`CLI.Scan` → semgrep)은 대상 디렉터리를 재귀로 훑는데
// 마커 파서만 `os.ReadDir` 로 최상위 파일을 보고 있었다. 그래서 하위 디렉터리 픽스처는
// 검출은 되는데 기대값이 안 읽혀 정상 픽스처가 "오탐 회귀" 로 실패했고, 경로 구조 자체를
// 판정 기준으로 삼는 룰(`paths.include` 를 쓰는 벤더 경로 룰 등)은 회귀 픽스처를
// 아예 만들 수 없었다. 순회 범위를 스캔과 일치시킨다.
//
// rel 은 픽스처 루트 기준 상대경로다. 베이스명을 쓰면 서로 다른 디렉터리의 동명 파일이
// 같은 키로 뭉개져 기대값이 섞인다.
//
// 확장자 화이트리스트 대신 NUL 바이트 유무로 바이너리를 거른다. 목록 방식은 새 형식의
// 픽스처가 들어올 때 마커를 **조용히** 누락시키는데, 그건 이 인프라에서 가장 위험한
// 실패 양상이다(기대값이 사라져도 테스트는 통과한다).
func walkFixtures(t *testing.T, dir string, fn func(rel string, lines []string)) {
	t.Helper()
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		raw, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if bytes.IndexByte(raw, 0) >= 0 {
			return nil // 바이너리 — 마커가 있을 수 없다
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		fn(filepath.ToSlash(rel), strings.Split(string(raw), "\n"))
		return nil
	})
	if err != nil {
		t.Fatalf("픽스처 순회 실패: %v", err)
	}
}

// parseExpectations 는 픽스처 트리를 훑어 마커에서 기대 집합을 만든다.
func parseExpectations(t *testing.T, dir string) map[expectation]bool {
	t.Helper()
	want := map[expectation]bool{}

	walkFixtures(t, dir, func(rel string, lines []string) {
		for i, l := range lines {
			id := expectMarkerRuleID(l)
			if id == "" {
				continue
			}
			if i+1 >= len(lines) {
				t.Errorf("%s:%d 마커가 파일 마지막 줄에 있어 대상 줄이 없다", rel, i+1)
				continue
			}
			// lines 는 0-based, 파일 라인은 1-based → 마커 다음 줄 = i+2
			want[expectation{file: rel, line: i + 2, ruleID: id}] = true
		}
	})
	if len(want) == 0 {
		t.Fatal("픽스처에서 EXPECT 마커를 하나도 찾지 못했다 — 마커 문법이 깨졌거나 경로가 틀렸다")
	}
	return want
}

// TestSafeFixturesHaveNoExpectations 는 safe_ 접두 픽스처가 마커를 갖지 않음을 강제한다.
//
// safe_* 는 "어떤 룰에도 걸리면 안 되는" 대조군이다. 여기에 마커가 생기면
// TestFixtureExpectationsMatchScan 이 그 검출을 정당한 것으로 받아들여 대조군의 의미가
// 사라진다.
// 하위 디렉터리도 같은 기준으로 본다 (#60) — 대조군이 중첩 경로에 있다고 해서
// 검사를 빠져나가면 안 된다. 판정은 상대경로의 베이스명 접두로 한다.
func TestSafeFixturesHaveNoExpectations(t *testing.T) {
	dir := fixturesDir(t)
	walkFixtures(t, dir, func(rel string, lines []string) {
		if !strings.HasPrefix(path.Base(rel), "safe_") {
			return
		}
		for i, l := range lines {
			if expectMarkerRuleID(l) != "" {
				t.Errorf("%s:%d 대조군 픽스처에 EXPECT 마커가 있다 — safe_* 는 검출 0건이어야 한다", rel, i+1)
			}
		}
	})
}

// TestParseExpectationsWalksSubdirectories 는 #60 의 핵심을 합성 트리로 못박는다.
//
// 실제 픽스처 트리는 semgrep 이 있어야 검증되지만, 순회 자체는 바이너리 없이 고정할 수
// 있다. 확인하는 것은 셋이다.
//   - 하위 디렉터리의 마커가 읽힌다 (예전 os.ReadDir 구현이 통째로 버리던 지점)
//   - 서로 다른 디렉터리의 **동명 파일**이 상대경로로 구분돼 기대값이 섞이지 않는다
//   - NUL 바이트가 있는 바이너리는 건너뛴다
func TestParseExpectationsWalksSubdirectories(t *testing.T) {
	root := t.TempDir()
	write := func(rel, body string) {
		t.Helper()
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatalf("사전 준비 실패: %v", err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatalf("사전 준비 실패: %v", err)
		}
	}

	write("Trust.swift",
		"// EXPECT: finguard.swift.insecure-trust\nSecTrustSetExceptions(t, b)\n")
	write("Carthage/Checkouts/SDK/Trust.swift",
		"// 벤더 사본 — 최상위와 베이스명이 같다\n// EXPECT: finguard.swift.insecure-trust-vendor\nSecTrustSetExceptions(t, b)\n")
	write("Nested/Info.plist",
		"<!-- EXPECT: finguard.plist.ats-arbitrary-loads -->\n<key>NSAllowsArbitraryLoads</key>\n")
	write("blob.bin", "\x00# EXPECT: finguard.python.weak-hash\nx\n")

	got := parseExpectations(t, root)
	want := map[expectation]bool{
		{file: "Trust.swift", line: 2, ruleID: "finguard.swift.insecure-trust"}:                               true,
		{file: "Carthage/Checkouts/SDK/Trust.swift", line: 3, ruleID: "finguard.swift.insecure-trust-vendor"}: true,
		{file: "Nested/Info.plist", line: 2, ruleID: "finguard.plist.ats-arbitrary-loads"}:                    true,
	}
	for e := range want {
		if !got[e] {
			t.Errorf("기대값이 읽히지 않았다: %s", e)
		}
	}
	for e := range got {
		if !want[e] {
			t.Errorf("읽히면 안 되는 기대값이 나왔다: %s", e)
		}
	}
}
