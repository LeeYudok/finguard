package mapping

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// ruleIDs 는 rules/*.yaml 에 정의된 Semgrep 룰ID 전체를 모은다.
func ruleIDs(t *testing.T, rulesDir string) map[string]bool {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(rulesDir, "*.yaml"))
	if err != nil || len(files) == 0 {
		t.Fatalf("룰 파일 목록 실패: %v (files=%d)", err, len(files))
	}
	ids := map[string]bool{}
	for _, f := range files {
		raw, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("%s 읽기 실패: %v", f, err)
		}
		var doc struct {
			Rules []struct {
				ID string `yaml:"id"`
			} `yaml:"rules"`
		}
		if err := yaml.Unmarshal(raw, &doc); err != nil {
			t.Fatalf("%s 파싱 실패: %v", f, err)
		}
		for _, r := range doc.Rules {
			if r.ID == "" {
				t.Fatalf("%s: id 없는 룰 존재", f)
			}
			if ids[r.ID] {
				t.Fatalf("룰ID 중복: %s", r.ID)
			}
			ids[r.ID] = true
		}
	}
	return ids
}

// TestEveryRuleMapped 는 rules/ 의 모든 룰ID 가 mapping/rules.yaml 에 있고,
// 반대로 매핑에만 남은 죽은 룰ID 가 없음을 보장한다. 룰 신설 시 매핑 누락을 CI 에서 차단한다.
func TestEveryRuleMapped(t *testing.T) {
	tbl, err := Load(filepath.Join("..", "..", "mapping", "rules.yaml"))
	if err != nil {
		t.Fatalf("매핑 로드 실패: %v", err)
	}
	ids := ruleIDs(t, filepath.Join("..", "..", "rules"))

	for id := range ids {
		if _, ok := tbl.Lookup(id); !ok {
			t.Errorf("매핑 누락: 룰 %s 이 mapping/rules.yaml 에 없음", id)
		}
	}
	for id := range tbl {
		if !ids[id] {
			t.Errorf("죽은 매핑: %s 에 대응하는 룰이 rules/ 에 없음", id)
		}
	}
}
