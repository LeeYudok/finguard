# CONTEXT.md — finguard 의 존재 이유와 굳힌 결정

`AGENTS.md` 가 "어떻게 일하나"라면 이 파일은 **"왜 이렇게 생겼나"** 다. 구현 전에 한 번 읽고,
아래 결정을 뒤집고 싶으면 코드가 아니라 이 파일부터 고친다(이슈 + PR). 형식은 `/grill-me`
의 "단단해진 스펙" 을 따른다 — 결정 사항 / 비목표 / 수용한 리스크 / 인수 기준.

> 상태: 2026-08-25 초안. 기존 AGENTS.md·이슈 이력에서 역추출했고 `/grill-me` 라운드는 아직
> 돌리지 않았다. 다음 기능 착수 전 한 번 심문해 갱신할 것.

## 목적

은행권 개발 GitLab 의 MR 에 **감사 대응이 가능한 근거**(금보원 점검항목·개발보안가이드
조항)와 수정 예시를 인라인 코멘트로 단다. 사용자는 보안 담당 조직이며, 수신자는 MR 을 올린
개발자다. 없으면 보안팀이 MR 을 수작업으로 훑고 근거 문서를 손으로 찾아 붙인다.

## 결정 사항

| 질문 | 결정 | 이유 |
| :--- | :--- | :--- |
| 탐지 엔진을 직접 만드나? | **아니오. Semgrep OSS + 커스텀 룰 YAML** | 탐지기는 상품이 아니다. 자산은 룰과 **룰ID → 금보원 매핑**(`mapping/rules.yaml`) |
| MR 코멘트를 직접 POST 하나? | **아니오. reviewdog 이 diff 필터·중복 제거·Discussion API 담당** | GitLab API 세부와 중복 코멘트 관리는 이미 풀린 문제. 포크·수정 금지 |
| 코멘트 근거 문구를 LLM 이 쓰나? | **절대 아니오. 매핑 테이블 값 그대로** | 감사 자료다. 출처가 확정돼야 하고, 매핑 없는 룰은 코멘트 생략 |
| 어디서 돌아가나? | **보안팀 리눅스 서버 상시 구동, webhook 수신** | CI 러너 안이 아니다 — 대상 레포의 파이프라인에 손대지 않고 붙이기 위함 |
| 머지를 막나? | **commit status `finguard` 로 게이트, 기본 `ERROR` 만 차단** | 강제는 GitLab 의 "Pipelines must succeed" 설정에 위임. `FINGUARD_BLOCK_ON`·레포별 `.finguard.yml` 로 조정 |
| 의존성 정책 | **표준 라이브러리 우선, `vendor/` 커밋** | 폐쇄망. `go get` 불가 |
| 특정 은행 종속? | **아니오. 은행권 범용** | 은행별 차이는 `.finguard.yml`·매핑 테이블로 흡수 |
| 로그에 소스 남기나? | **경로 + 라인번호까지만** | 점검 대상은 고객사 코드 |
| 룰 기대값은 어디에? | **픽스처 파일 안 `EXPECT:` 마커** (`rules/AGENTS.md`) | 라인번호를 테스트에 적으면 픽스처 수정마다 충돌 (#44) |

## 비목표 (v1)

- 코드 자동 수정 — 가이드만 제시, 수정은 개발자 몫
- diff 라인 필터링·중복 코멘트 제거·GitLab Discussion API 직접 호출 — reviewdog 영역
- Semgrep Pro·타 SAST 엔진 지원
- GitLab 외 forge(GitHub·Bitbucket) 대상 스캔
- 대시보드·통계·이력 저장 — 스캔은 무상태, 결과는 MR 에만 남는다
- 외부 API 호출 전부 (LLM 포함)

## 수용한 리스크

- 매핑에 없는 룰의 finding 은 **조용히 사라진다** — 로그로만 남고 코멘트 없음. 매핑 커버리지는 사람이 관리.
- webhook 은 항상 200 을 돌려주므로 GitLab 쪽에선 실패를 모른다 — 실패 감지는 서버 로그 모니터링에 의존.
- 동일 MR 스캔 중 재푸시는 스킵된다(inflight) — 다음 푸시에서 다시 잡힌다는 전제.
- 시크릿 탐지 정탐 픽스처는 형태상 진짜 시크릿과 구분이 안 된다 — `.github/secret_scanning.yml` 제외에 의존 (#82).

## 인수 기준 (항상 참이어야 하는 것)

- `go build -mod=vendor ./...` · `go test -race -mod=vendor ./...` 통과
- `semgrep --validate --config rules/` 가 룰 전부를 로드한다 (`0 rule(s)` 금지, #43)
- `go test -tags semgrep_integration ./internal/scanner/` — 마커 있는데 미검출 0, 마커 없는데 검출 0
- 코멘트 본문의 `basis`·`kisa_item` 이 `mapping/rules.yaml` 과 바이트 단위로 같다
- Draft MR·라벨만 바뀐 update·비-MR 이벤트에 스캔이 돌지 않는다
