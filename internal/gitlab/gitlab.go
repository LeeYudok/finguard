// Package gitlab 은 사내 GitLab 에서 점검 대상 소스를 확보한다.
// 외부 API 호출 없음 — 사내 GitLab 만 접근한다.
package gitlab

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os/exec"
	"regexp"
	"strings"
)

// shaPattern 은 git 커밋 해시(축약 포함, sha-1/sha-256) 형식이다.
// checkout 인자로 넘기기 전 payload 의 sha 가 이 형식인지 검증해,
// `-` 로 시작하는 값이 git 옵션으로 해석되는 인자 인젝션을 차단한다.
var shaPattern = regexp.MustCompile(`^[0-9a-fA-F]{7,64}$`)

// ValidateCloneHost 는 webhook payload 의 clone URL 이 신뢰 GitLab(apiV4URL 의
// 호스트)을 가리키는지 검증한다. payload 는 위조 가능하므로, 검증 없이 clone 하면
// API 토큰이 임의 호스트로 전송될 수 있다.
func ValidateCloneHost(apiV4URL, cloneURL string) error {
	au, err := url.Parse(apiV4URL)
	if err != nil {
		return fmt.Errorf("API URL 파싱 실패: %w", err)
	}
	cu, err := url.Parse(cloneURL)
	if err != nil {
		return fmt.Errorf("clone URL 파싱 실패: %w", err)
	}
	// https 만 허용한다. http 를 허용하면 FetchSource 가 URL 에 심는 oauth2
	// 토큰이 평문으로 전송될 수 있다(위조 payload 로 다운그레이드 가능).
	if cu.Scheme != "https" {
		return fmt.Errorf("clone URL scheme 불허(https 만 허용): %s", cu.Scheme)
	}
	if !strings.EqualFold(au.Hostname(), cu.Hostname()) {
		return fmt.Errorf("clone URL 호스트 불일치: %s (신뢰 호스트 %s)", cu.Hostname(), au.Hostname())
	}
	return nil
}

// FetchSource 는 cloneURL 레포에서 MR ref(refs/merge-requests/<iid>/head)를
// destDir 에 얕게 받아 sha 로 체크아웃한다. shallow clone 은 --single-branch 가
// 암묵 적용되어 기본 브랜치만 받으므로, MR 브랜치 커밋은 MR ref 를 직접 fetch 한다.
// token 이 있으면 HTTP 클론 URL 에 oauth2 자격으로 심는다.
func FetchSource(ctx context.Context, cloneURL, token string, mrIID int, sha, destDir string) error {
	cu := cloneURL
	if token != "" {
		u, err := url.Parse(cloneURL)
		if err != nil {
			return fmt.Errorf("clone URL 파싱 실패: %w", err)
		}
		u.User = url.UserPassword("oauth2", token)
		cu = u.String()
	}
	steps := [][]string{
		{"git", "init", "--quiet", destDir},
		{"git", "-C", destDir, "remote", "add", "origin", cu},
		{"git", "-C", destDir, "fetch", "--quiet", "--depth", "50", "origin",
			fmt.Sprintf("refs/merge-requests/%d/head", mrIID)},
	}
	// sha 는 payload 에서 온 위조 가능한 값이다. 형식이 커밋 해시가 아니면
	// FETCH_HEAD 로 폴백하지 말고 거부한다 — 잘못된 커밋을 조용히 체크아웃하는
	// 대신 명시적으로 실패하는 편이 안전하다.
	co := "FETCH_HEAD"
	if sha != "" {
		if !shaPattern.MatchString(sha) {
			return fmt.Errorf("checkout sha 형식 불허: %q", sha)
		}
		co = sha
	}
	// ref 뒤에 "--" 를 두어 pathspec 이 없음을 못박는다(뒤에 붙이는 이유: `--` 앞이
	// pathspec 이 아니라 커밋으로 해석되어야 하므로). ref 는 위에서 형식 검증된 값이다.
	steps = append(steps, []string{"git", "-C", destDir, "checkout", "--quiet", co, "--"})
	for _, s := range steps {
		cmd := exec.CommandContext(ctx, s[0], s[1:]...)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			// 토큰이 URL 에 있으므로 커맨드라인은 로그에 남기지 않는다.
			msg := stderr.String()
			if len(msg) > 1000 {
				msg = msg[:1000]
			}
			return fmt.Errorf("git %s 실패: %w: %s", gitVerb(s), err, msg)
		}
	}
	return nil
}

// gitVerb 는 로그용 git 서브커맨드 이름을 뽑는다 ("-C <dir>" 프리픽스 무시).
func gitVerb(s []string) string {
	for i := 1; i < len(s); i++ {
		if s[i] == "-C" {
			i++
			continue
		}
		return s[i]
	}
	return s[0]
}
