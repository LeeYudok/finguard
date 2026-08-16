package scanner

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

// Finding 은 SARIF 결과 한 건을 finguard 이 쓰는 최소 형태로 줄인 것이다.
type Finding struct {
	RuleID    string
	Path      string
	Message   string
	StartLine int
	StartCol  int
	EndLine   int
	EndCol    int
}

// SARIF 2.1.0 중 finguard 이 읽는 필드만 정의한다.
type sarifDoc struct {
	Runs []struct {
		Results []struct {
			RuleID  string `json:"ruleId"`
			Message struct {
				Text string `json:"text"`
			} `json:"message"`
			Locations []struct {
				PhysicalLocation struct {
					ArtifactLocation struct {
						URI string `json:"uri"`
					} `json:"artifactLocation"`
					Region struct {
						StartLine   int `json:"startLine"`
						StartColumn int `json:"startColumn"`
						EndLine     int `json:"endLine"`
						EndColumn   int `json:"endColumn"`
					} `json:"region"`
				} `json:"physicalLocation"`
			} `json:"locations"`
		} `json:"results"`
	} `json:"runs"`
}

// ParseSARIF 는 SARIF JSON 을 Finding 목록으로 변환한다.
// basePath 가 주어지면 결과 경로에서 그 접두를 제거해 레포 상대경로로 만든다.
func ParseSARIF(data []byte, basePath string) ([]Finding, error) {
	var doc sarifDoc
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("SARIF 파싱 실패: %w", err)
	}
	var out []Finding
	for _, run := range doc.Runs {
		for _, r := range run.Results {
			if len(r.Locations) == 0 {
				continue
			}
			loc := r.Locations[0].PhysicalLocation
			p := loc.ArtifactLocation.URI
			if basePath != "" {
				if rel, err := filepath.Rel(basePath, p); err == nil && !strings.HasPrefix(rel, "..") {
					p = rel
				}
			}
			f := Finding{
				RuleID:    r.RuleID,
				Path:      filepath.ToSlash(p),
				Message:   r.Message.Text,
				StartLine: loc.Region.StartLine,
				StartCol:  loc.Region.StartColumn,
				EndLine:   loc.Region.EndLine,
				EndCol:    loc.Region.EndColumn,
			}
			// reviewdog rdjsonl 은 1 미만 라인/컬럼을 받지 않으므로 보정한다.
			if f.StartLine < 1 {
				f.StartLine = 1
			}
			if f.StartCol < 1 {
				f.StartCol = 1
			}
			if f.EndLine < f.StartLine {
				f.EndLine = f.StartLine
			}
			if f.EndCol < 1 {
				f.EndCol = f.StartCol
			}
			out = append(out, f)
		}
	}
	return out, nil
}
