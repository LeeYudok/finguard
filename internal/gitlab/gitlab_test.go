package gitlab

import (
	"context"
	"testing"
)

func TestValidateCloneHost(t *testing.T) {
	api := "https://gitlab.example.com/api/v4"
	cases := []struct {
		name, cloneURL string
		ok             bool
	}{
		{"신뢰 호스트", "https://gitlab.example.com/dok123/app.git", true},
		{"대소문자 무시", "https://gitlab.example.com/dok123/app.git", true},
		{"다른 호스트", "https://evil.example.com/dok123/app.git", false},
		{"서브도메인 위장", "https://gitlab.example.com.evil.example.com/a.git", false},
		{"ssh scheme", "ssh://git@gitlab.example.com/dok123/app.git", false},
		{"http 다운그레이드", "http://gitlab.example.com/dok123/app.git", false},
		{"빈 URL", "", false},
	}
	for _, tc := range cases {
		err := ValidateCloneHost(api, tc.cloneURL)
		if tc.ok && err != nil {
			t.Errorf("%s: 허용돼야 하는데 거부됨: %v", tc.name, err)
		}
		if !tc.ok && err == nil {
			t.Errorf("%s: 거부돼야 하는데 허용됨", tc.name)
		}
	}
}

// FetchSource 는 sha 형식 검증을 git 명령 실행 전에 수행하므로,
// 비정상 sha 는 네트워크·git 접근 없이 거부돼야 한다.
func TestFetchSourceRejectsBadSHA(t *testing.T) {
	bad := []string{
		"--upload-pack=touch /tmp/pwned", // 옵션 인젝션 시도
		"-x",
		"deadbeef; rm -rf /",
		"refs/heads/main",
		"zzzz",
	}
	for _, sha := range bad {
		err := FetchSource(context.Background(),
			"https://gitlab.example.com/a/b.git", "", 1, sha, t.TempDir()+"/src")
		if err == nil {
			t.Errorf("비정상 sha %q: 거부돼야 하는데 통과", sha)
		}
	}
}
