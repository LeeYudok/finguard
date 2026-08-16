// Package mapping 은 Semgrep 룰ID 를 금보원/개발보안가이드 근거 항목으로 잇는
// 매핑 테이블을 로드한다. 이 테이블이 finguard 의 고유 자산이며, 코멘트에 쓰이는
// 근거 문구는 전부 이 테이블 값을 그대로 사용한다.
package mapping

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Rule 은 매핑 한 건이다. mapping/rules.yaml 의 항목 하나와 대응한다.
type Rule struct {
	RuleID     string `yaml:"rule_id"`
	Code       string `yaml:"code"`
	CWE        string `yaml:"cwe"`
	Title      string `yaml:"title"`
	Severity   string `yaml:"severity"`
	Basis      string `yaml:"basis"`
	KisaItem   string `yaml:"kisa_item"`
	Explain    string `yaml:"explain"`
	FixExample string `yaml:"fix_example"`
}

// Table 은 rule_id 를 키로 하는 매핑 테이블이다.
type Table map[string]Rule

// Lookup 은 Semgrep 이 보고한 룰ID 로 매핑을 찾는다.
// Semgrep 은 --config 로 준 경로를 룰ID 앞에 점 구분으로 붙이므로
// (예: rules.finguard.go.sql-sprintf) 정확 일치가 없으면 ".<rule_id>" suffix 로 재시도한다.
func (t Table) Lookup(ruleID string) (Rule, bool) {
	if r, ok := t[ruleID]; ok {
		return r, true
	}
	for key, r := range t {
		if strings.HasSuffix(ruleID, "."+key) {
			return r, true
		}
	}
	return Rule{}, false
}

var validSeverity = map[string]bool{"ERROR": true, "WARNING": true, "INFO": true}

// Load 는 YAML 매핑 파일을 읽어 Table 로 만든다.
// rule_id/code/severity 가 비거나 severity 가 허용값 밖이거나 rule_id 가 중복이면 실패한다.
func Load(path string) (Table, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("매핑 파일 읽기 실패: %w", err)
	}
	var rules []Rule
	if err := yaml.Unmarshal(raw, &rules); err != nil {
		return nil, fmt.Errorf("매핑 YAML 파싱 실패(%s): %w", path, err)
	}
	t := make(Table, len(rules))
	for i, r := range rules {
		if r.RuleID == "" || r.Code == "" || r.Title == "" {
			return nil, fmt.Errorf("매핑 %d번째 항목: rule_id/code/title 은 필수", i)
		}
		if !validSeverity[r.Severity] {
			return nil, fmt.Errorf("매핑 %s: severity 는 ERROR/WARNING/INFO 중 하나여야 함 (현재 %q)", r.RuleID, r.Severity)
		}
		if _, dup := t[r.RuleID]; dup {
			return nil, fmt.Errorf("매핑 rule_id 중복: %s", r.RuleID)
		}
		t[r.RuleID] = r
	}
	return t, nil
}
