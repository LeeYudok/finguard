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

func requireSemgrep(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("semgrep"); err != nil {
		t.Skip("semgrep 바이너리가 없어 통합 테스트를 건너뛴다")
	}
}

// 마커 문법(`expectMarkerRuleID`)과 픽스처 순회(`walkFixtures`·`parseExpectations`),
// 그리고 대조군 검사(`TestSafeFixturesHaveNoExpectations`)는 빌드 태그가 없는
// expect_marker_test.go 에 있다. 이 파일은 semgrep_integration 태그 뒤에 있어
// 기본 `go test` 에서 컴파일되지 않으므로, 마커 문법을 여기 복제하면 양쪽이 어긋나도
// 아무도 모른다 — 그 순간 이 테스트는 "기대 0건 · 검출 0건" 으로 조용히 통과한다
// (#60·#61). 정의는 한 곳에만 두고, semgrep 이 필요한 검사만 여기 남긴다.

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
		// f.Path 는 ParseSARIF 가 스캔 루트(= 픽스처 루트) 기준 상대경로로 만들어 준다.
		got[expectation{file: filepath.ToSlash(f.Path), line: f.StartLine, ruleID: id}] = true
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
