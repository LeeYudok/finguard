# finguard

은행 소스코드 취약점 점검 봇. 특정 은행에 종속되지 않는 은행권 범용 도구다. 개발 GitLab의 MR에 취약점 근거와 수정 가이드를 인라인 코멘트로 단다.

보안 담당 조직의 리눅스 서버에서 상시 구동되며(CI 러너 안이 아님), Semgrep으로 스캔한 결과를 룰ID→금보원/개발보안가이드 매핑으로 변환해 reviewdog을 통해 MR discussion으로 게시한다.

```
개발 GitLab (MR webhook)
  → finguard: Semgrep 실행 → SARIF 수신
  → finguard: 룰ID로 금보원 매핑 조회 → rdjsonl 생성
  → reviewdog: diff 필터 · 중복 제거 · MR discussion POST
  → 개발 GitLab MR 인라인 코멘트
```

## 디렉토리

```
cmd/finguard/main.go
internal/
  webhook/     MR 이벤트 수신 · secret token 검증
  gitlab/      diff 조회 · 소스 clone · commit status
  scanner/     semgrep 실행 · SARIF 파싱
  mapping/     룰ID → 금보원 항목 (YAML 로드)
  rdjson/      SARIF + 매핑 → rdjsonl 출력
  runner/      reviewdog 프로세스 파이프
  repoconfig/  .finguard.yml 레포별 오버라이드
mapping/rules.yaml   룰ID → 금보원 매핑 테이블 (이 프로젝트 고유 자산)
rules/               Semgrep 커스텀 룰 (언어별 YAML)
```

## 빌드 · 테스트

```bash
go build -mod=vendor ./...
go test -mod=vendor ./...
```

폐쇄망 전제로 의존성은 `vendor/`에 커밋한다 (`go get` 불가).

## 실행 규약

```bash
# 로컬 디렉터리 점검 (사람이 읽는 리포트 / rdjsonl)
finguard scan --dir <소스경로> [--format report|rdjsonl] [--rules <경로>] [--mapping <경로>]

# MR webhook 서버 — reviewdog·commit status 게시까지 내부에서 수행
FINGUARD_WEBHOOK_SECRET=... REVIEWDOG_GITLAB_API_TOKEN=... CI_API_V4_URL=... \
  finguard serve --addr :8480
```

serve 환경변수: `FINGUARD_BLOCK_ON`(차단 severity, 기본 `ERROR`), `FINGUARD_MAX_SCANS`(동시 스캔 상한, 기본 2).
`--rules`·`--mapping` 기본값은 실행파일 기준 상대 경로다 (cwd 무관).

상세 아키텍처·매핑 스키마·코멘트 규칙은 [`CLAUDE.md`](./CLAUDE.md) 참고.

## 한계 (반드시 읽을 것)

**finguard 통과가 "취약점 없음"을 의미하지 않는다.** 이 도구는 MR 단위 실시간 게이트이며, 정기 종합 취약점 진단(상용/인증 SAST, 모의해킹, SCA)을 대체하지 않는다. 이중 체계의 보완재로 운용할 것.

- **diff 라인 한정 점검**: reviewdog `-filter-mode=added` 로 MR에서 추가·변경된 라인만 코멘트한다. 변경 라인 밖의 기존 취약점은 잡지 않는다. 레포 전체 점검은 `finguard scan --dir` 로 별도 수행해야 한다.
- **탐지 방식의 한계**: 룰 상당수가 정규식 기반이라 데이터 흐름을 추적하지 못하며, 오탐(예제·테스트 문자열)과 미탐(변수 경유·문자열 조립)이 공존한다. taint(오염 추적) 룰은 Java·Go·TypeScript 일부에만 있고, Semgrep OSS 특성상 **파일 단위 추적**이다 — Controller→Service→DAO 로 계층이 분리된 코드에서는 추적이 파일 경계에서 끊긴다.
- **커버리지 범위**: 매핑된 룰이 다루는 취약점 유형만 점검한다. 인가·세션 관리·역직렬화 등 미커버 영역이 존재하며, 커버 범위는 `mapping/rules.yaml` 이 전부다.
- **SCA 미포함**: 오픈소스 의존성의 알려진 취약점(CVE)은 점검 대상이 아니다. 별도 도구가 필요하다.
- **언어 범위**: `rules/` 에 룰이 있는 언어만 점검한다. COBOL·PL/SQL·JSP 등 레거시 스택은 Semgrep OSS 가 지원하지 않는다.

## 라이선스

finguard 는 [MIT License](./LICENSE) 로 배포한다.

서드파티:

- [`gopkg.in/yaml.v3`](https://github.com/go-yaml/yaml) — MIT / Apache-2.0. `vendor/` 에 원본 라이선스와 함께 포함.
- [Semgrep OSS](https://github.com/semgrep/semgrep) — LGPL-2.1. 이 저장소에 포함되지 않으며 외부 바이너리로 호출한다.
- [reviewdog](https://github.com/reviewdog/reviewdog) — MIT. 이 저장소에 포함되지 않으며 외부 바이너리로 호출한다.

Semgrep 룰(`rules/`)은 전부 자체 작성이며 Semgrep Registry 룰을 포함하지 않는다.
