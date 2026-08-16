// Package repoconfig 는 점검 대상 레포 루트의 .finguard.yml 오버라이드를 다룬다.
// 소스는 로컬에 clone 되므로 GitLab API 가 아니라 체크아웃 디렉터리에서 읽는다.
package repoconfig

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// FileName 은 대상 레포 루트에서 찾는 설정 파일 이름이다.
const FileName = ".finguard.yml"

// Config 는 .finguard.yml 스키마다. 파일이 없으면 zero value 로 전역 기본만 적용된다.
type Config struct {
	// Ignore 는 점검에서 제외할 파일 glob (path.Match 문법).
	Ignore []string `yaml:"ignore"`
	// BlockOn 은 차단 게이트 대상 severity 오버라이드. 포인터인 이유:
	// "필드 없음"(nil → 전역 기본 적용)과 "명시적 빈 배열"(block_on: [] →
	// 이 레포 차단 opt-out)을 yaml 에서 구분해야 해서다.
	BlockOn *[]string `yaml:"block_on"`
}

// Parse 는 .finguard.yml 바이트를 파싱한다.
func Parse(b []byte) (Config, error) {
	var c Config
	if err := yaml.Unmarshal(b, &c); err != nil {
		return Config{}, fmt.Errorf(".finguard.yml 파싱 실패: %w", err)
	}
	return c, nil
}

// Load 는 체크아웃 루트 dir 의 .finguard.yml 을 읽는다.
// 파일이 없으면 (zero value, false, nil), 읽기·파싱 실패는 에러로 반환한다.
func Load(dir string) (Config, bool, error) {
	b, err := os.ReadFile(filepath.Join(dir, FileName))
	if os.IsNotExist(err) {
		return Config{}, false, nil
	}
	if err != nil {
		return Config{}, false, fmt.Errorf(".finguard.yml 읽기 실패: %w", err)
	}
	c, err := Parse(b)
	if err != nil {
		return Config{}, false, err
	}
	return c, true, nil
}

// Ignored 는 finding 파일 경로(레포 상대, 슬래시 구분)가 ignore 패턴에 걸리는지 판정한다.
// path.Match 의 `*` 는 `/` 를 못 넘으므로, 슬래시 없는 패턴(`*.sql` 등)은
// 디렉터리 무관하게 파일명(base)에도 매칭한다 — 일반적인 gitignore 직관에 맞추기 위함.
func (c Config) Ignored(p string) bool {
	base := path.Base(p)
	for _, pat := range c.Ignore {
		if ok, err := path.Match(pat, p); err == nil && ok {
			return true
		}
		if !strings.Contains(pat, "/") {
			if ok, err := path.Match(pat, base); err == nil && ok {
				return true
			}
		}
	}
	return false
}

// EffectiveBlockOn 은 레포 block_on 이 존재하면(빈 배열 포함) global 을 오버라이드한다.
// 레포 값은 대소문자 관용을 위해 정규화한다.
func EffectiveBlockOn(global []string, c Config) []string {
	if c.BlockOn == nil {
		return global
	}
	out := make([]string, 0, len(*c.BlockOn))
	for _, s := range *c.BlockOn {
		if s = strings.ToUpper(strings.TrimSpace(s)); s != "" {
			out = append(out, s)
		}
	}
	return out
}

// GlobalBlockOn 은 env FINGUARD_BLOCK_ON(쉼표 구분)에서 전역 차단 대상을 읽는다.
// 미설정이면 기본 ["ERROR"], 빈 문자열로 명시하면 전역 게이트 off(빈 슬라이스).
func GlobalBlockOn() []string {
	v, ok := os.LookupEnv("FINGUARD_BLOCK_ON")
	if !ok {
		v = "ERROR"
	}
	return splitBlockOn(v)
}

func splitBlockOn(v string) []string {
	out := []string{}
	for _, s := range strings.Split(v, ",") {
		if s = strings.ToUpper(strings.TrimSpace(s)); s != "" {
			out = append(out, s)
		}
	}
	return out
}

// Gate 는 매핑된 finding 의 severity 목록과 차단 대상으로 commit status 를 판정한다.
func Gate(severities, blockOn []string) (state, desc string) {
	n := 0
	for _, sev := range severities {
		if sevIn(sev, blockOn) {
			n++
		}
	}
	switch {
	case n > 0:
		return "failed", fmt.Sprintf("차단 대상 지적 %d건 (%s)", n, strings.Join(blockOn, ","))
	case len(blockOn) == 0:
		return "success", "차단 게이트 off"
	default:
		return "success", "차단 대상 지적 없음"
	}
}

func sevIn(sev string, set []string) bool {
	u := strings.ToUpper(sev)
	for _, s := range set {
		if u == s {
			return true
		}
	}
	return false
}
