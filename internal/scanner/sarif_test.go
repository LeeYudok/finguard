package scanner

import (
	"os"
	"testing"
)

func TestParseSARIF(t *testing.T) {
	raw, err := os.ReadFile("testdata/sample.sarif")
	if err != nil {
		t.Fatal(err)
	}
	got, err := ParseSARIF(raw, "repo")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("위치 있는 결과 1건을 기대, got %d", len(got))
	}
	f := got[0]
	if f.RuleID != "finguard.swift.hardcoded-secret" ||
		f.Path != "Sources/Login.swift" ||
		f.StartLine != 42 || f.StartCol != 9 || f.EndLine != 42 || f.EndCol != 68 {
		t.Fatalf("unexpected finding: %+v", f)
	}
}

func TestParseSARIFClampsZero(t *testing.T) {
	raw := []byte(`{"runs":[{"results":[{"ruleId":"r","message":{"text":"m"},"locations":[{"physicalLocation":{"artifactLocation":{"uri":"a.swift"},"region":{}}}]}]}]}`)
	got, err := ParseSARIF(raw, "")
	if err != nil {
		t.Fatal(err)
	}
	f := got[0]
	if f.StartLine != 1 || f.StartCol != 1 || f.EndLine != 1 || f.EndCol != 1 {
		t.Fatalf("0 값 보정 실패: %+v", f)
	}
}
