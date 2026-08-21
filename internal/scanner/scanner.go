// Package scanner 는 Semgrep 을 실행하고 SARIF 출력을 파싱한다.
package scanner

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

// semgrepIgnoreBody 는 스캔 대상 루트에 심는 `.semgrepignore` 내용이다.
//
// semgrep 은 대상 루트에 `.semgrepignore` 가 없으면 **내장 기본 무시목록**을 쓰는데,
// 거기에 `test/`·`tests/`·`*_test.go` 가 들어 있다 (semgrep 1.169.0 실측). 그 결과
// `src/test/java/**` 같은 경로가 룰의 paths 설정과 무관하게 조용히 사라진다 —
// 테스트 코드에 복붙된 실키를 보겠다는 #25 의 룰이 통째로 무효화되는 경로다.
//
// 제외 정책은 finguard 가 `--exclude`(DefaultExcludes)와 `.finguard.yml` 로 이미
// 소유하고 있으므로, 여기서는 아무것도 무시하지 않는 파일을 심어 내장 기본값을 대체한다.
// 레포가 자기 `.semgrepignore` 를 갖고 있으면 그쪽 의도를 존중해 건드리지 않는다.
const semgrepIgnoreBody = `# finguard 가 생성한 파일 (#25).
# semgrep 내장 기본 무시목록(test/·tests/·*_test.go 포함)을 대체한다.
# 제외 경로는 finguard 의 --exclude 와 .finguard.yml 의 ignore 가 담당한다.
`

// ensureSemgrepIgnore 는 대상 루트에 `.semgrepignore` 가 없을 때만 생성하고,
// 자기가 만들었는지를 돌려준다(호출부가 스캔 후 되돌리기 위함).
// 실패해도 스캔 자체는 계속한다 — 무시목록은 정확도 문제이지 중단 사유가 아니다.
func ensureSemgrepIgnore(targetDir string) (created bool, err error) {
	p := filepath.Join(targetDir, ".semgrepignore")
	if _, err := os.Stat(p); err == nil {
		return false, nil // 레포가 자기 것을 갖고 있다 — 그 의도를 존중한다
	} else if !os.IsNotExist(err) {
		return false, err
	}
	if err := os.WriteFile(p, []byte(semgrepIgnoreBody), 0o644); err != nil {
		return false, err
	}
	return true, nil
}

func (c CLI) Scan(ctx context.Context, rulesPath, targetDir string) ([]Finding, error) {
	bin := c.Bin
	if bin == "" {
		bin = "semgrep"
	}
	// 우리가 심은 경우에만 스캔 후 되돌린다. `scan --dir=<작업 사본>` 로 자기 레포를
	// 점검하는 로컬 모드에서 잔여물이 남으면 사용자가 의도치 않게 커밋하게 된다.
	// finguard 는 점검 도구이지 대상 레포를 바꾸는 도구가 아니다.
	created, err := ensureSemgrepIgnore(targetDir)
	if err != nil {
		// 경로·라인만 남기는 로깅 규약상 여기서는 조용히 진행한다.
		_ = err
	}
	if created {
		defer func() { _ = os.Remove(filepath.Join(targetDir, ".semgrepignore")) }()
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
