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
	// "Pods" 는 여기서 빼고 rules/swift.yaml 의 룰별 paths.exclude 로 옮겼다 (#75).
	// 전역에서 제외하면 semgrep 이 타겟 단계에서 걸러버려, 벤더 경로를 **의도적으로**
	// 점검하는 룰(finguard.swift.insecure-trust-vendor — 벤더 Swift 코드의 인증서 검증
	// 무력화는 최종 바이너리에 그대로 실려 나간다)의 include 가 도달하지 못했다.
	// Carthage 가 원래 여기 없었던 것과의 비대칭도 이 이관으로 해소된다.
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
