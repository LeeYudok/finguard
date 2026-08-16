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

func (c CLI) Scan(ctx context.Context, rulesPath, targetDir string) ([]Finding, error) {
	bin := c.Bin
	if bin == "" {
		bin = "semgrep"
	}
	cmd := exec.CommandContext(ctx, bin, "scan",
		"--sarif", "--quiet", "--metrics=off", "--disable-version-check",
		"--config", rulesPath, targetDir)
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
