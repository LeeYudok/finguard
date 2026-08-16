package scanner

// DefaultExcludes 는 점검 대상에서 기본 제외하는 경로 패턴이다.
//
// 여기 실린 것은 "그 레포가 작성·소유한 운영 코드가 아닌 것"만이다 —
// 빌드 산출물, 내려받은 의존성, 벤더링된 서드파티, 도구가 생성한 파일.
// 이런 경로를 점검하면 개발자가 고칠 수 없는 코드에 코멘트가 달리고,
// 실제 지적이 노이즈에 묻힌다.
//
// 판단이 갈리는 경로(example/·docs/·테스트 코드 등)는 여기 넣지 않는다.
// 레포 사정에 따라 운영 코드일 수도 있으므로 `.finguard.yml` 의 ignore 로
// 각 레포가 정한다.
var DefaultExcludes = []string{
	// 의존성 — 내려받은 코드
	"node_modules",
	"vendor",
	"vendors",
	"third_party",
	"thirdparty",
	"bower_components",
	"Pods",
	".venv",
	"site-packages",

	// 빌드 산출물
	"dist",
	"build",
	"built",
	"out",
	"target",
	".next",
	".nuxt",
	".output",
	".gradle",
	"DerivedData",

	// 번들·압축 결과 — 원본이 따로 있다
	"*.min.js",
	"*.min.css",
	"*.bundle.js",
	"*.map",

	// 도구 생성 파일 — 손으로 고치는 대상이 아니다
	"*.pb.go",
	"*_pb2.py",
	"*.g.dart",
	"*_generated.go",
	"*.generated.ts",
	"*.d.ts",

	// VCS·캐시
	".git",
	".svn",
	".mypy_cache",
	".pytest_cache",
	"__pycache__",
}
