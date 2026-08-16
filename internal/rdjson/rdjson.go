// Package rdjson 은 SARIF Finding 과 매핑 테이블을 합쳐 reviewdog rdjsonl 을 만든다.
package rdjson

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/leeyudok/finguard/internal/mapping"
	"github.com/leeyudok/finguard/internal/scanner"
)

type source struct {
	Name string `json:"name"`
}

type position struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

type rng struct {
	Start position `json:"start"`
	End   position `json:"end"`
}

type location struct {
	Path  string `json:"path"`
	Range rng    `json:"range"`
}

// Line 은 rdjsonl 한 줄(= MR 코멘트 하나)이다.
type Line struct {
	Source   source   `json:"source"`
	Severity string   `json:"severity"`
	Location location `json:"location"`
	Message  string   `json:"message"`
}

func severityKo(s string) string {
	switch s {
	case "ERROR":
		return "상"
	case "WARNING":
		return "중"
	default:
		return "하"
	}
}

// Message 는 코멘트 문구 규칙(CLAUDE.md)대로 본문을 만든다.
// 근거 문구는 매핑 테이블 값만 사용한다.
func Message(r mapping.Rule) string {
	var b strings.Builder
	fmt.Fprintf(&b, "[%s] %s (%s)\n\n", r.Code, r.Title, r.CWE)
	basis := r.Basis
	if r.KisaItem != "" {
		basis += " / 금보원 평가항목 " + r.KisaItem
	}
	fmt.Fprintf(&b, "▸ 근거: %s\n▸ 심각도: %s\n\n%s\n", basis, severityKo(r.Severity), strings.TrimSpace(r.Explain))
	if fix := strings.TrimSpace(r.FixExample); fix != "" {
		fmt.Fprintf(&b, "\n수정 예시:\n%s\n", fix)
	}
	return b.String()
}

// Build 는 Finding 하나를 rdjsonl 한 줄로 변환한다.
func Build(f scanner.Finding, r mapping.Rule) Line {
	return Line{
		Source:   source{Name: "finguard"},
		Severity: r.Severity,
		Location: location{
			Path: f.Path,
			Range: rng{
				Start: position{Line: f.StartLine, Column: f.StartCol},
				End:   position{Line: f.EndLine, Column: f.EndCol},
			},
		},
		Message: Message(r),
	}
}

// Emit 은 매핑에 있는 Finding 만 rdjsonl 로 내보낸다.
// 매핑에 없는 룰ID 는 코멘트를 생략하고 개수만 반환한다(임의 문구 생성 금지).
func Emit(w io.Writer, findings []scanner.Finding, tbl mapping.Table) (written, skipped int, err error) {
	enc := json.NewEncoder(w)
	for _, f := range findings {
		r, ok := tbl.Lookup(f.RuleID)
		if !ok {
			skipped++
			continue
		}
		if err := enc.Encode(Build(f, r)); err != nil {
			return written, skipped, err
		}
		written++
	}
	return written, skipped, nil
}
