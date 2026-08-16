package scanner

import (
	"strings"
	"testing"
)

func TestScanArgs(t *testing.T) {
	args := scanArgs("/rules", "/target", []string{"node_modules", "*.min.js"})
	joined := strings.Join(args, " ")

	for _, want := range []string{
		"scan", "--sarif", "--quiet", "--metrics=off", "--disable-version-check",
	} {
		if !contains(args, want) {
			t.Errorf("필수 인자 누락: %s (전체: %s)", want, joined)
		}
	}

	// --config 는 룰 경로를 값으로 받아야 한다
	if i := indexOf(args, "--config"); i < 0 || i+1 >= len(args) || args[i+1] != "/rules" {
		t.Errorf("--config 뒤에 룰 경로가 오지 않았다: %s", joined)
	}

	// 대상 디렉터리는 마지막 인자 — 그래야 semgrep 이 플래그가 아닌 대상으로 읽는다
	if args[len(args)-1] != "/target" {
		t.Errorf("대상 디렉터리가 마지막 인자가 아니다: %s", joined)
	}
}

func TestScanArgsPassesEveryExclude(t *testing.T) {
	excludes := []string{"node_modules", "target", "*.min.js"}
	args := scanArgs("/r", "/t", excludes)

	for _, ex := range excludes {
		found := false
		for i := 0; i+1 < len(args); i++ {
			if args[i] == "--exclude" && args[i+1] == ex {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("제외 경로가 --exclude 로 전달되지 않았다: %s", ex)
		}
	}

	if got, want := countOf(args, "--exclude"), len(excludes); got != want {
		t.Errorf("--exclude 개수 불일치: got %d, want %d", got, want)
	}
}

// 기본 제외 목록이 비면 노이즈가 그대로 돌아오므로, 대표 항목이 남아 있는지 지킨다.
func TestDefaultExcludesCoversBuildOutputAndDeps(t *testing.T) {
	for _, want := range []string{"node_modules", "vendor", "dist", "build", "target"} {
		if !contains(DefaultExcludes, want) {
			t.Errorf("기본 제외 목록에 %q 가 없다", want)
		}
	}
}

func contains(ss []string, s string) bool { return indexOf(ss, s) >= 0 }

func indexOf(ss []string, s string) int {
	for i, v := range ss {
		if v == s {
			return i
		}
	}
	return -1
}

func countOf(ss []string, s string) int {
	n := 0
	for _, v := range ss {
		if v == s {
			n++
		}
	}
	return n
}
