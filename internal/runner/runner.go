// Package runner 는 reviewdog 바이너리를 파이프로 실행한다.
// reviewdog 은 포크/수정 없이 외부 프로세스로만 호출한다.
package runner

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// Env 는 reviewdog gitlab-mr-discussion reporter 가 요구하는 환경변수 묶음이다.
// CI 밖에서 돌므로 finguard 이 webhook payload 에서 뽑아 채운다.
type Env struct {
	APIToken  string // REVIEWDOG_GITLAB_API_TOKEN
	APIV4URL  string // CI_API_V4_URL
	ProjectID string // CI_PROJECT_ID
	MRIID     string // CI_MERGE_REQUEST_IID
	CommitSHA string // CI_COMMIT_SHA
}

// Reviewdog 은 목 가능하도록 인터페이스로 둔다.
type Reviewdog interface {
	Post(ctx context.Context, rdjsonl io.Reader, env Env) error
}

// CLI 는 reviewdog 바이너리를 호출하는 기본 구현이다.
type CLI struct {
	Bin string // 비면 "reviewdog"
}

func (c CLI) Post(ctx context.Context, rdjsonl io.Reader, env Env) error {
	bin := c.Bin
	if bin == "" {
		bin = "reviewdog"
	}
	cmd := exec.CommandContext(ctx, bin,
		"-f=rdjsonl",
		"-name=finguard",
		"-reporter=gitlab-mr-discussion",
		"-filter-mode=added",
		"-fail-level=error",
	)
	cmd.Stdin = rdjsonl
	cmd.Env = append(os.Environ(),
		"REVIEWDOG_GITLAB_API_TOKEN="+env.APIToken,
		"CI_API_V4_URL="+env.APIV4URL,
		"CI_PROJECT_ID="+env.ProjectID,
		"CI_MERGE_REQUEST_IID="+env.MRIID,
		"CI_COMMIT_SHA="+env.CommitSHA,
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := stderr.String()
		if len(msg) > 2000 {
			msg = msg[:2000]
		}
		return fmt.Errorf("reviewdog 실행 실패: %w: %s", err, msg)
	}
	return nil
}
