//go:build semgrep_integration

// 실제 semgrep 바이너리를 돌려 룰이 취약 패턴을 검출하는지 확인하는 통합 테스트.
// 폐쇄망·CI 에는 semgrep 이 없을 수 있어 기본 go test 에서 제외한다.
//
//	실행: go test -tags semgrep_integration ./internal/scanner/
//
// semgrep 이 PATH 에 없으면 실패가 아니라 skip 한다.
package scanner

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func rulesDir(t *testing.T) string {
	t.Helper()
	// internal/scanner/ → 저장소 루트의 rules/
	p, err := filepath.Abs("../../rules")
	if err != nil {
		t.Fatalf("rules 경로 확인 실패: %v", err)
	}
	return p
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

func requireSemgrep(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("semgrep"); err != nil {
		t.Skip("semgrep 바이너리가 없어 통합 테스트를 건너뛴다")
	}
}

// ── 기대값 마커 ──
//
// 픽스처는 검출이 기대되는 줄 **바로 위**에 마커 주석을 단다 (#44):
//
//	# EXPECT: finguard.python.cleartext-websocket
//	REALTIME_FEED_URL = "ws://ops.example-broker.co.kr:21000"
//
// 기대값이 테스트가 아니라 픽스처 안에 있으므로 블록을 어디에 추가하든 따라 움직인다.
// 라인 번호를 테스트에 하드코딩하던 구조는 픽스처를 건드리는 모든 PR 을 서로
// 충돌시켰고, 병합 후 조용히 어긋나기까지 했다.
var expectMarker = regexp.MustCompile(`(?:#|//)\s*EXPECT:\s*(\S+)\s*$`)

// expectation 은 픽스처 한 줄에 기대되는 검출이다.
type expectation struct {
	file   string // 픽스처 파일의 베이스명
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

// parseExpectations 는 픽스처 디렉터리를 훑어 마커에서 기대 집합을 만든다.
func parseExpectations(t *testing.T, dir string) map[expectation]bool {
	t.Helper()
	want := map[expectation]bool{}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("픽스처 디렉터리 읽기 실패: %v", err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatalf("픽스처 읽기 실패 %s: %v", e.Name(), err)
		}
		lines := strings.Split(string(raw), "\n")
		for i, l := range lines {
			m := expectMarker.FindStringSubmatch(l)
			if m == nil {
				continue
			}
			if i+1 >= len(lines) {
				t.Errorf("%s:%d 마커가 파일 마지막 줄에 있어 대상 줄이 없다", e.Name(), i+1)
				continue
			}
			// lines 는 0-based, 파일 라인은 1-based → 마커 다음 줄 = i+2
			want[expectation{file: e.Name(), line: i + 2, ruleID: m[1]}] = true
		}
	}
	if len(want) == 0 {
		t.Fatal("픽스처에서 EXPECT 마커를 하나도 찾지 못했다 — 마커 문법이 깨졌거나 경로가 틀렸다")
	}
	return want
}

// TestFixtureExpectationsMatchScan 은 픽스처 마커와 실제 스캔 결과가 정확히 일치하는지 본다.
//
// 양방향으로 검사한다.
//   - 마커가 있는데 검출이 없다 → 미탐 회귀
//   - 마커가 없는데 검출이 있다 → 오탐 회귀 (safe_* 픽스처의 "검출 0건"이 여기에 포함된다)
func TestFixtureExpectationsMatchScan(t *testing.T) {
	requireSemgrep(t)
	dir := fixturesDir(t)

	want := parseExpectations(t, dir)

	findings, err := CLI{}.Scan(context.Background(), rulesDir(t), dir)
	if err != nil {
		t.Fatalf("스캔 실패: %v", err)
	}

	got := map[expectation]bool{}
	for _, f := range findings {
		// semgrep 이 디렉터리 config 에서 "rules." 접두어를 붙이므로 잘라낸다.
		id := f.RuleID
		if i := strings.LastIndex(id, "finguard."); i >= 0 {
			id = id[i:]
		}
		got[expectation{file: filepath.Base(f.Path), line: f.StartLine, ruleID: id}] = true
	}

	var missing, unexpected []string
	for e := range want {
		if !got[e] {
			missing = append(missing, e.String())
		}
	}
	for e := range got {
		if !want[e] {
			unexpected = append(unexpected, e.String())
		}
	}
	sort.Strings(missing)
	sort.Strings(unexpected)

	for _, m := range missing {
		t.Errorf("미탐 회귀 — 마커가 있는데 검출되지 않았다: %s", m)
	}
	for _, u := range unexpected {
		t.Errorf("오탐 회귀 — 마커가 없는데 검출됐다: %s", u)
	}

	t.Logf("기대 %d건 · 검출 %d건 · 전부 일치", len(want), len(got))
}

// TestSafeFixturesHaveNoExpectations 는 safe_ 접두 픽스처가 마커를 갖지 않음을 강제한다.
//
// safe_* 는 "어떤 룰에도 걸리면 안 되는" 대조군이다. 여기에 마커가 생기면 위 테스트가
// 그 검출을 정당한 것으로 받아들여 대조군의 의미가 사라진다.
func TestSafeFixturesHaveNoExpectations(t *testing.T) {
	dir := fixturesDir(t)
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("픽스처 디렉터리 읽기 실패: %v", err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasPrefix(e.Name(), "safe_") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatalf("픽스처 읽기 실패 %s: %v", e.Name(), err)
		}
		for i, l := range strings.Split(string(raw), "\n") {
			if expectMarker.MatchString(l) {
				t.Errorf("%s:%d 대조군 픽스처에 EXPECT 마커가 있다 — safe_* 는 검출 0건이어야 한다", e.Name(), i+1)
			}
		}
	}
}

// TestNoYAMLFixturesUnderRules 는 #43 회귀 방어다.
//
// rules/ 하위에 .yml/.yaml 픽스처가 생기면 semgrep 이 그것을 룰 파일로 파싱해
// 룰셋 전체가 로드 실패한다. 룰 파일 자신(rules/*.yaml)만 허용한다.
func TestNoYAMLFixturesUnderRules(t *testing.T) {
	root := rulesDir(t)
	err := filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(p))
		if ext != ".yml" && ext != ".yaml" {
			return nil
		}
		if filepath.Dir(p) != root {
			t.Errorf("rules/ 하위 디렉터리에 YAML 이 있다: %s — semgrep 이 룰 파일로 파싱해 룰셋 전체가 로드 실패한다 (#43). 픽스처는 testdata/rule-fixtures/ 에 둔다", p)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("rules/ 순회 실패: %v", err)
	}
}

// 회귀(#25): 스캔이 대상 레포에 잔여물을 남기지 않아야 한다.
//
// semgrep 내장 무시목록을 대체하려면 대상 루트에 `.semgrepignore` 를 심어야 하는데,
// `scan --dir=<작업 사본>` 로 자기 레포를 점검하는 로컬 모드에서 그게 남으면
// 사용자가 의도치 않게 커밋한다. finguard 는 점검 도구이지 대상을 바꾸는 도구가 아니다.
func TestScanLeavesNoResidueInTarget(t *testing.T) {
	requireSemgrep(t)

	t.Run("우리가 심은 파일은 스캔 후 사라진다", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "x.py"), []byte("A = 1\n"), 0o644); err != nil {
			t.Fatalf("사전 준비 실패: %v", err)
		}
		if _, err := (CLI{}).Scan(context.Background(), rulesDir(t), dir); err != nil {
			t.Fatalf("스캔 실패: %v", err)
		}
		if _, err := os.Stat(filepath.Join(dir, ".semgrepignore")); !os.IsNotExist(err) {
			t.Error("스캔 후 .semgrepignore 가 대상 레포에 남았다 — 사용자가 커밋하게 된다")
		}
	})

	t.Run("레포 자신의 파일은 보존된다", func(t *testing.T) {
		dir := t.TempDir()
		p := filepath.Join(dir, ".semgrepignore")
		if err := os.WriteFile(p, []byte("generated/\n"), 0o644); err != nil {
			t.Fatalf("사전 준비 실패: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, "x.py"), []byte("A = 1\n"), 0o644); err != nil {
			t.Fatalf("사전 준비 실패: %v", err)
		}
		if _, err := (CLI{}).Scan(context.Background(), rulesDir(t), dir); err != nil {
			t.Fatalf("스캔 실패: %v", err)
		}
		got, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("레포의 .semgrepignore 가 지워졌다: %v", err)
		}
		if string(got) != "generated/\n" {
			t.Errorf("레포의 .semgrepignore 가 변조됐다: %q", string(got))
		}
	})
}
