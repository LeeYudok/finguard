package rdjson

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/leeyudok/finguard/internal/mapping"
	"github.com/leeyudok/finguard/internal/scanner"
)

var testRule = mapping.Rule{
	RuleID:     "finguard.swift.hardcoded-secret",
	Code:       "FIN-KEY-001",
	CWE:        "CWE-798",
	Title:      "하드코드된 중요정보",
	Severity:   "ERROR",
	Basis:      "개발보안가이드 「보안기능」",
	Explain:    "소스코드에 비밀정보가 하드코드되어 있습니다.",
	FixExample: "```swift\nlet key = Keychain.load(\"api_key\")\n```",
}

func TestEmitSkipsUnmapped(t *testing.T) {
	findings := []scanner.Finding{
		{RuleID: "finguard.swift.hardcoded-secret", Path: "a.swift", StartLine: 1, StartCol: 1, EndLine: 1, EndCol: 5},
		{RuleID: "unknown.rule", Path: "b.swift", StartLine: 2, StartCol: 1, EndLine: 2, EndCol: 2},
	}
	tbl := mapping.Table{testRule.RuleID: testRule}
	var buf bytes.Buffer
	written, skipped, err := Emit(&buf, findings, tbl)
	if err != nil {
		t.Fatal(err)
	}
	if written != 1 || skipped != 1 {
		t.Fatalf("written=%d skipped=%d", written, skipped)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("rdjsonl 1줄 기대, got %d", len(lines))
	}
	var l Line
	if err := json.Unmarshal([]byte(lines[0]), &l); err != nil {
		t.Fatal(err)
	}
	if l.Source.Name != "finguard" || l.Severity != "ERROR" || l.Location.Path != "a.swift" {
		t.Fatalf("unexpected line: %+v", l)
	}
	for _, want := range []string{"[FIN-KEY-001] 하드코드된 중요정보 (CWE-798)", "▸ 근거: 개발보안가이드 「보안기능」", "▸ 심각도: 상", "수정 예시:"} {
		if !strings.Contains(l.Message, want) {
			t.Fatalf("message 에 %q 누락:\n%s", want, l.Message)
		}
	}
}

func TestMessageAppendsKisaItem(t *testing.T) {
	r := testRule
	r.KisaItem = "112(프로그램 통제)"
	msg := Message(r)
	if !strings.Contains(msg, "금보원 평가항목 112(프로그램 통제)") {
		t.Fatalf("kisa_item 미반영:\n%s", msg)
	}
}
