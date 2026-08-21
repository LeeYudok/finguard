package scanner

import (
	"regexp"
	"testing"
)

// expectMarkerPattern 은 integration_test.go 의 expectMarker 와 같은 문법이다.
// 그쪽은 semgrep_integration 빌드 태그 뒤에 있어 기본 `go test` 에서 컴파일되지 않으므로,
// 마커 문법 자체의 회귀는 태그 없는 이 테스트가 지킨다 (#44).
//
// 두 정규식이 어긋나면 픽스처 마커가 조용히 무시되고, 그 순간 통합 테스트는
// "기대 0건 · 검출 0건" 으로 통과해버린다 — 그래서 문법을 여기에 고정한다.
const expectMarkerPattern = `(?:#|//)\s*EXPECT:\s*(\S+)\s*$`

func TestExpectMarkerSyntax(t *testing.T) {
	re := regexp.MustCompile(expectMarkerPattern)

	tests := []struct {
		name string
		line string
		want string // 빈 문자열이면 매칭되지 않아야 한다
	}{
		// 매칭돼야 하는 형태
		{"파이썬/셸 주석", "# EXPECT: finguard.python.weak-hash", "finguard.python.weak-hash"},
		{"TS/Java 주석", "// EXPECT: finguard.ts.eval", "finguard.ts.eval"},
		{"들여쓰기", "    # EXPECT: finguard.python.eval-exec", "finguard.python.eval-exec"},
		{"콜론 뒤 공백 여러 개", "#   EXPECT:   finguard.go.weak-crypto", "finguard.go.weak-crypto"},
		{"주석기호 뒤 공백 없음", "#EXPECT: finguard.swift.http-url", "finguard.swift.http-url"},
		{"줄 끝 공백", "# EXPECT: finguard.java.xxe   ", "finguard.java.xxe"},

		// 매칭되면 안 되는 형태
		{"평범한 주석", "# 이 줄은 API 키를 담는다", ""},
		{"EXPECT 언급만", "# 이 룰은 EXPECT 마커로 검증한다", ""},
		{"콜론 없음", "# EXPECT finguard.python.weak-hash", ""},
		{"룰ID 없음", "# EXPECT:", ""},
		{"룰ID 자리가 공백뿐", "# EXPECT:    ", ""},
		{"주석 아닌 코드", `EXPECT = "finguard.python.weak-hash"`, ""},
		{"뒤에 다른 내용이 붙음", "# EXPECT: finguard.python.weak-hash 이유: 테스트", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := re.FindStringSubmatch(tc.line)
			if tc.want == "" {
				if m != nil {
					t.Errorf("매칭되면 안 되는데 %q 로 매칭됐다: %q", m[1], tc.line)
				}
				return
			}
			if m == nil {
				t.Fatalf("매칭돼야 하는데 안 됐다: %q", tc.line)
			}
			if m[1] != tc.want {
				t.Errorf("룰ID 추출 결과 %q, 기대 %q (입력 %q)", m[1], tc.want, tc.line)
			}
		})
	}
}

// TestExpectMarkerRejectsTrailingText 는 "뒤에 다른 내용이 붙음" 케이스의 의도를 못박는다.
//
// 마커 줄에 설명을 덧붙이면 조용히 무시되는 대신 아예 매칭되지 않아야 한다.
// 그래야 통합 테스트가 "마커 없는데 검출됨(오탐 회귀)" 으로 즉시 실패해 작성자가 알아챈다.
func TestExpectMarkerRejectsTrailingText(t *testing.T) {
	re := regexp.MustCompile(expectMarkerPattern)
	if re.MatchString("# EXPECT: finguard.python.weak-hash — MD5 사용") {
		t.Error("마커 뒤 설명을 허용하면 룰ID 에 설명이 섞여 들어간다 — 거부돼야 한다")
	}
}
