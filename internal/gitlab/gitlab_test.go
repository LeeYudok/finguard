package gitlab

import "testing"

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
