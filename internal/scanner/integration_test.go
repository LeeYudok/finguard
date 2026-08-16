//go:build semgrep_integration

// 실제 semgrep 바이너리를 돌려 룰이 취약 패턴을 검출하는지 확인하는 통합 테스트.
// 폐쇄망·CI 에는 semgrep 이 없을 수 있어 기본 go test 에서 제외한다.
//   실행: go test -tags semgrep_integration ./internal/scanner/
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
