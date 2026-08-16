package mapping

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "rules.yaml")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestLoadOK(t *testing.T) {
	p := writeTemp(t, `
- rule_id: finguard.swift.hardcoded-secret
  code: FIN-KEY-001
  cwe: CWE-798
  title: 하드코드된 중요정보
  severity: ERROR
  basis: 개발보안가이드 「보안기능」
  explain: 설명
  fix_example: 예시
`)
	tbl, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	r, ok := tbl["finguard.swift.hardcoded-secret"]
	if !ok || r.Code != "FIN-KEY-001" || r.Severity != "ERROR" {
		t.Fatalf("unexpected: %+v", r)
	}
}

func TestLoadRejectsBadSeverity(t *testing.T) {
	p := writeTemp(t, `
- rule_id: a
  code: C1
  title: t
  severity: CRITICAL
`)
	if _, err := Load(p); err == nil || !strings.Contains(err.Error(), "severity") {
		t.Fatalf("severity 검증 실패를 기대: %v", err)
	}
}

func TestLoadRejectsDup(t *testing.T) {
	p := writeTemp(t, `
- {rule_id: a, code: C1, title: t, severity: ERROR}
- {rule_id: a, code: C2, title: t, severity: ERROR}
`)
	if _, err := Load(p); err == nil || !strings.Contains(err.Error(), "중복") {
		t.Fatalf("중복 검증 실패를 기대: %v", err)
	}
}

func TestLookupSuffixMatch(t *testing.T) {
	tbl := Table{"finguard.go.sql-sprintf": {RuleID: "finguard.go.sql-sprintf", Code: "FIN-SQLI-002"}}
	if _, ok := tbl.Lookup("finguard.go.sql-sprintf"); !ok {
		t.Fatal("정확 일치 실패")
	}
	r, ok := tbl.Lookup("rules.finguard.go.sql-sprintf")
	if !ok || r.Code != "FIN-SQLI-002" {
		t.Fatalf("suffix 매칭 실패: ok=%v r=%+v", ok, r)
	}
	if _, ok := tbl.Lookup("other.rule"); ok {
		t.Fatal("무관한 룰ID 가 매칭되면 안 됨")
	}
	if _, ok := tbl.Lookup("go.sql-sprintf"); ok {
		t.Fatal("부분 suffix(짧은 쪽)가 매칭되면 안 됨")
	}
}
