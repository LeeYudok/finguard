// Package scanner 는 Semgrep 을 실행하고 SARIF 출력을 파싱한다.
package scanner

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

// Semgrep 은 목(mock) 가능하도록 인터페이스로 둔다.
type Semgrep interface {
	Scan(ctx context.Context, rulesPath, targetDir string) ([]Finding, error)
}

// CLI 는 semgrep 바이너리를 호출하는 기본 구현이다.
type CLI struct {
	Bin string // 비면 "semgrep"
}

// scanArgs 는 semgrep 실행 인자를 조립한다. 외부 프로세스 없이 검증할 수 있도록
// 순수 함수로 분리했다 — 제외 경로가 조용히 누락되면 노이즈가 그대로 돌아온다.
func scanArgs(rulesPath, targetDir string, excludes []string) []string {
	args := []string{"scan",
		"--sarif", "--quiet", "--metrics=off", "--disable-version-check",
		"--config", rulesPath}
	// 운영 코드가 아닌 경로(의존성·빌드 산출물·생성 파일)는 점검하지 않는다.
	// 개발자가 고칠 수 없는 코드에 코멘트가 달리면 실제 지적이 묻힌다.
	for _, ex := range excludes {
		args = append(args, "--exclude", ex)
	}
	return append(args, targetDir)
}

func (c CLI) Scan(ctx context.Context, rulesPath, targetDir string) ([]Finding, error) {
	bin := c.Bin
	if bin == "" {
		bin = "semgrep"
	}
	cmd := exec.CommandContext(ctx, bin, scanArgs(rulesPath, targetDir, DefaultExcludes)...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		// 로그에 소스 본문이 섞이지 않도록 stderr 앞부분만 남긴다.
		msg := stderr.String()
		if len(msg) > 2000 {
			msg = msg[:2000]
		}
		return nil, fmt.Errorf("semgrep 실행 실패: %w: %s", err, msg)
	}
	return ParseSARIF(stdout.Bytes(), targetDir)
}
