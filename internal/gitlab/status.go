package gitlab

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// SetCommitStatus 는 GitLab commits API 로 name=finguard 의 commit status 를 게시한다.
// state 는 GitLab 이 받는 값(success·failed 등) 그대로 넘긴다.
// apiV4URL 은 CI_API_V4_URL 형식(https://gitlab.example.com/api/v4)이다.
func SetCommitStatus(ctx context.Context, apiV4URL, token string, projectID int, sha, state, description string) error {
	endpoint := fmt.Sprintf("%s/projects/%d/statuses/%s",
		strings.TrimRight(apiV4URL, "/"), projectID, url.PathEscape(sha))
	form := url.Values{
		"state":       {state},
		"name":        {"finguard"},
		"description": {description},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint,
		strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("commit status 요청 생성 실패: %w", err)
	}
	req.Header.Set("PRIVATE-TOKEN", token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	hc := &http.Client{Timeout: 30 * time.Second}
	resp, err := hc.Do(req)
	if err != nil {
		return fmt.Errorf("commit status 게시 실패: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("commit status 게시 실패: status %d", resp.StatusCode)
	}
	return nil
}
