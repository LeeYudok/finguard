package scanner

import (
	"os"
	"path/filepath"
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

// 벤더 경로는 전역이 아니라 룰별 paths.exclude 가 소유한다 (#75).
//
// 전역 `--exclude` 는 semgrep 의 타겟 선정 단계에서 걸러버려 룰의 paths.include 보다
// 먼저 작동한다. 그래서 여기에 "Pods" 를 되돌려 놓으면, 벤더 경로를 **의도적으로**
// 점검하는 finguard.swift.insecure-trust-vendor 의 include 가 다시 도달 불가가 된다
// (벤더 Swift 코드의 인증서 검증 무력화는 최종 바이너리에 그대로 실려 나간다).
//
// 일반 Swift 룰 9개의 벤더 제외는 rules/swift.yaml 의 룰별 exclude 로 이관돼 있다.
func TestDefaultExcludesLeavesVendorScopeToRules(t *testing.T) {
	for _, notWant := range []string{"Pods", "Carthage"} {
		if contains(DefaultExcludes, notWant) {
			t.Errorf("기본 제외 목록에 %q 가 있다 — 벤더 경로를 점검하는 룰의 include 가 도달 불가가 된다 (#75). "+
				"제외가 필요하면 rules/swift.yaml 의 룰별 paths.exclude 에 넣어라", notWant)
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

// semgrep 내장 기본 무시목록이 test/·tests/·*_test.go 를 통째로 지우는 문제 때문에
// 스캔 대상 루트에 .semgrepignore 를 심는다 (#25).
func TestEnsureSemgrepIgnore(t *testing.T) {
	dir := t.TempDir()

	created, err := ensureSemgrepIgnore(dir)
	if err != nil {
		t.Fatalf("생성 실패: %v", err)
	}
	if !created {
		t.Error("우리가 생성했는데 created=false 다 — 호출부가 스캔 후 되돌리지 못한다")
	}
	got, err := os.ReadFile(filepath.Join(dir, ".semgrepignore"))
	if err != nil {
		t.Fatalf(".semgrepignore 가 생성되지 않았다: %v", err)
	}
	if !strings.Contains(string(got), "finguard") {
		t.Errorf("finguard 가 생성했다는 표시가 없다: %q", string(got))
	}
	// 아무 경로도 무시하지 않아야 한다 — 제외 정책은 --exclude 와 .finguard.yml 소관이다.
	for _, line := range strings.Split(string(got), "\n") {
		s := strings.TrimSpace(line)
		if s != "" && !strings.HasPrefix(s, "#") {
			t.Errorf("무시 패턴이 들어 있다: %q", s)
		}
	}
}

// 레포가 자기 .semgrepignore 를 갖고 있으면 그쪽 의도를 존중해 덮어쓰지 않는다.
func TestEnsureSemgrepIgnoreKeepsExisting(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, ".semgrepignore")
	if err := os.WriteFile(p, []byte("generated/\n"), 0o644); err != nil {
		t.Fatalf("사전 준비 실패: %v", err)
	}
	created, err := ensureSemgrepIgnore(dir)
	if err != nil {
		t.Fatalf("실행 실패: %v", err)
	}
	if created {
		t.Error("레포의 기존 파일인데 created=true 다 — 호출부가 남의 파일을 지우게 된다")
	}
	got, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("읽기 실패: %v", err)
	}
	if string(got) != "generated/\n" {
		t.Errorf("레포의 기존 .semgrepignore 를 덮어썼다: %q", string(got))
	}
}
