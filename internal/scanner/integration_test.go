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
	"os/exec"
	"path/filepath"
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

// 회귀: es-toolkit 에서 제보한 template() variable 코드 주입 패턴을
// FIN-INJ-001(finguard.ts.eval)이 반드시 검출해야 한다.
func TestDetectsTemplateVariableInjection(t *testing.T) {
	requireSemgrep(t)
	dir, _ := filepath.Abs("../../rules/testdata/vuln-samples")

	findings, err := CLI{}.Scan(context.Background(), rulesDir(t), dir)
	if err != nil {
		t.Fatalf("스캔 실패: %v", err)
	}

	var hitInjection bool
	for _, f := range findings {
		// semgrep 이 디렉터리 config 에서 "rules." 접두어를 붙이므로 suffix 로 본다
		if strings.HasSuffix(f.RuleID, "finguard.ts.eval") &&
			filepath.Base(f.Path) == "template_variable_injection.ts" {
			hitInjection = true
		}
		// 대조 픽스처는 어떤 룰에도 걸리지 않아야 한다
		if filepath.Base(f.Path) == "safe_template.ts" {
			t.Errorf("안전한 대조 픽스처가 %s 로 오검출됐다 (line %d)", f.RuleID, f.StartLine)
		}
	}

	if !hitInjection {
		t.Error("template variable 코드 주입(FIN-INJ-001)이 검출되지 않았다 — 회귀")
	}
}

// 회귀(#19): curl|sh 룰이 실행되는 명령만 잡고, 안내용 문자열 리터럴은 잡지 않아야 한다.
// 실코드 스캔에서 검출 7건 중 3건이 `info "설치 중: curl ... | sh"` 형태의 오탐이었다.
func TestCurlPipeShellSkipsStringLiterals(t *testing.T) {
	requireSemgrep(t)
	dir, _ := filepath.Abs("../../rules/testdata/vuln-samples")

	findings, err := CLI{}.Scan(context.Background(), rulesDir(t), dir)
	if err != nil {
		t.Fatalf("스캔 실패: %v", err)
	}

	// 정탐 픽스처의 실행 라인 — 하나라도 빠지면 recall 회귀다.
	wantLines := map[int]bool{7: false, 10: false, 13: false, 16: false, 19: false, 22: false, 25: false}

	for _, f := range findings {
		base := filepath.Base(f.Path)
		// 안내 문자열만 담은 픽스처는 어떤 룰에도 걸리면 안 된다.
		if base == "safe_curl_pipe_shell_string.sh" {
			t.Errorf("문자열 리터럴 픽스처가 %s 로 오검출됐다 (line %d)", f.RuleID, f.StartLine)
			continue
		}
		if base != "curl_pipe_shell.sh" {
			continue
		}
		if !strings.HasSuffix(f.RuleID, "finguard.shell.curl-pipe-shell") {
			t.Errorf("정탐 픽스처가 예상 밖 룰 %s 로 검출됐다 (line %d)", f.RuleID, f.StartLine)
			continue
		}
		if _, ok := wantLines[f.StartLine]; !ok {
			t.Errorf("정탐 픽스처의 비대상 라인 %d 이 검출됐다", f.StartLine)
			continue
		}
		wantLines[f.StartLine] = true
	}

	for line, hit := range wantLines {
		if !hit {
			t.Errorf("curl_pipe_shell.sh:%d 이 검출되지 않았다 — 회귀", line)
		}
	}
}

// 회귀(#20): 파이썬 룰 8종이 각각 정탐 픽스처를 정확히 검출하고, 오탐 방지
// 픽스처는 어떤 룰에도 걸리지 않아야 한다. "실코드 스캔에서 파이썬 검출 0건"이
// 나왔을 때 그것이 코드가 깨끗해서인지 룰이 죽어서인지 이 테스트가 판별한다.
func TestPythonRulesDetectVulnsAndSkipSafeFixtures(t *testing.T) {
	requireSemgrep(t)
	dir, _ := filepath.Abs("../../rules/testdata/vuln-samples")

	findings, err := CLI{}.Scan(context.Background(), rulesDir(t), dir)
	if err != nil {
		t.Fatalf("스캔 실패: %v", err)
	}

	// 룰ID(suffix) → python_vulns.py 에서 정탐이 기대되는 라인 집합.
	// tls-verify-disabled 는 형태가 여럿이라 라인이 여러 개다 (#26):
	// requests verify=False / 전역 기본 컨텍스트 교체 / verify_mode 직접 해제 / httpx.
	wantLines := map[string][]int{
		"finguard.python.hardcoded-secret":    {17},
		"finguard.python.weak-hash":           {22},
		"finguard.python.http-url":            {26},
		"finguard.python.sql-format":          {31},
		"finguard.python.subprocess-shell":    {37},
		"finguard.python.eval-exec":           {42},
		"finguard.python.tls-verify-disabled": {47, 52, 58, 64},
		"finguard.python.cleartext-websocket": {68},
		"finguard.python.yaml-unsafe-load":    {73},
	}
	hit := map[string]map[int]bool{}
	for id, lines := range wantLines {
		hit[id] = make(map[int]bool, len(lines))
		for _, ln := range lines {
			hit[id][ln] = false
		}
	}

	for _, f := range findings {
		base := filepath.Base(f.Path)

		if base == "safe_python.py" {
			t.Errorf("오탐 방지 픽스처가 %s 로 오검출됐다 (line %d)", f.RuleID, f.StartLine)
			continue
		}
		if base != "python_vulns.py" {
			continue
		}

		var matched string
		for ruleID := range wantLines {
			if strings.HasSuffix(f.RuleID, ruleID) {
				matched = ruleID
				break
			}
		}
		if matched == "" {
			t.Errorf("python_vulns.py 가 예상 밖 룰 %s 로 검출됐다 (line %d)", f.RuleID, f.StartLine)
			continue
		}
		if _, ok := hit[matched][f.StartLine]; !ok {
			t.Errorf("%s 가 예상 밖 라인 %d 에서 검출됐다 (기대: %v)", matched, f.StartLine, wantLines[matched])
			continue
		}
		hit[matched][f.StartLine] = true
	}

	for ruleID, lines := range wantLines {
		for _, ln := range lines {
			if !hit[ruleID][ln] {
				t.Errorf("python_vulns.py:%d 에서 %s 가 검출되지 않았다 — 회귀", ln, ruleID)
			}
		}
	}
}
