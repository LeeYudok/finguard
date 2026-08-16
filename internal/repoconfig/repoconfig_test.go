package repoconfig

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestIgnored(t *testing.T) {
	cfg := Config{Ignore: []string{"*.sql", "vendor/*", "docs/*.md"}}
	cases := []struct {
		name string
		path string
		want bool
	}{
		{"슬래시 없는 패턴은 base 매칭", "db/schema/init.sql", true},
		{"루트 파일도 매칭", "init.sql", true},
		{"디렉터리 패턴 매칭", "vendor/lib.go", true},
		{"디렉터리 패턴은 하위 디렉터리 못 넘음", "vendor/a/b.go", false},
		{"경로 있는 패턴", "docs/guide.md", true},
		{"경로 있는 패턴은 base 매칭 안 함", "src/guide.md", false},
		{"무관 파일", "src/main/java/UserDao.java", false},
	}
	for _, tc := range cases {
		if got := cfg.Ignored(tc.path); got != tc.want {
			t.Errorf("%s: Ignored(%q) = %v, 원하는 값 %v", tc.name, tc.path, got, tc.want)
		}
	}
	if (Config{}).Ignored("a.go") {
		t.Error("빈 설정: Ignored = true, 원하는 값 false")
	}
}

func TestEffectiveBlockOn(t *testing.T) {
	global := []string{"ERROR"}
	cases := []struct {
		name string
		cfg  Config
		want []string
	}{
		{"필드 없음 → 전역 기본", Config{}, []string{"ERROR"}},
		{"명시적 빈 배열 → off", Config{BlockOn: &[]string{}}, []string{}},
		{"오버라이드 + 정규화", Config{BlockOn: &[]string{" warning ", "ERROR"}}, []string{"WARNING", "ERROR"}},
	}
	for _, tc := range cases {
		if got := EffectiveBlockOn(global, tc.cfg); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s: got %v, 원하는 값 %v", tc.name, got, tc.want)
		}
	}
}

func TestGlobalBlockOn(t *testing.T) {
	os.Unsetenv("FINGUARD_BLOCK_ON")
	if got := GlobalBlockOn(); !reflect.DeepEqual(got, []string{"ERROR"}) {
		t.Errorf("미설정: got %v, 원하는 값 [ERROR]", got)
	}
	t.Setenv("FINGUARD_BLOCK_ON", "")
	if got := GlobalBlockOn(); len(got) != 0 {
		t.Errorf("빈 문자열 명시: got %v, 원하는 값 빈 슬라이스", got)
	}
	t.Setenv("FINGUARD_BLOCK_ON", "error, warning")
	if got := GlobalBlockOn(); !reflect.DeepEqual(got, []string{"ERROR", "WARNING"}) {
		t.Errorf("쉼표 구분: got %v, 원하는 값 [ERROR WARNING]", got)
	}
}

func TestGate(t *testing.T) {
	cases := []struct {
		name    string
		sevs    []string
		blockOn []string
		want    string
	}{
		{"차단 대상 있음", []string{"ERROR", "WARNING"}, []string{"ERROR"}, "failed"},
		{"차단 대상 없음", []string{"WARNING", "INFO"}, []string{"ERROR"}, "success"},
		{"finding 없음", nil, []string{"ERROR"}, "success"},
		{"게이트 off", []string{"ERROR"}, []string{}, "success"},
		{"WARNING 까지 차단", []string{"WARNING"}, []string{"ERROR", "WARNING"}, "failed"},
	}
	for _, tc := range cases {
		if state, _ := Gate(tc.sevs, tc.blockOn); state != tc.want {
			t.Errorf("%s: state = %s, 원하는 값 %s", tc.name, state, tc.want)
		}
	}
}

func TestLoad(t *testing.T) {
	dir := t.TempDir()

	cfg, found, err := Load(dir)
	if err != nil || found {
		t.Fatalf("파일 없음: found=%v err=%v, 원하는 값 false/nil", found, err)
	}
	if len(cfg.Ignore) != 0 || cfg.BlockOn != nil {
		t.Fatalf("파일 없음: zero value 기대, got %+v", cfg)
	}

	yml := "ignore:\n  - \"*.sql\"\nblock_on: [warning]\n"
	if err := os.WriteFile(filepath.Join(dir, FileName), []byte(yml), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, found, err = Load(dir)
	if err != nil || !found {
		t.Fatalf("정상 파일: found=%v err=%v", found, err)
	}
	if !cfg.Ignored("a/b.sql") {
		t.Error("ignore 미적용")
	}
	if cfg.BlockOn == nil || len(*cfg.BlockOn) != 1 || (*cfg.BlockOn)[0] != "warning" {
		t.Errorf("block_on = %v, 원하는 값 [warning]", cfg.BlockOn)
	}

	if err := os.WriteFile(filepath.Join(dir, FileName), []byte(":\tbroken"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Load(dir); err == nil {
		t.Error("깨진 yaml: err = nil, 원하는 값 에러")
	}
}
