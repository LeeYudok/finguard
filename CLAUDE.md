# finguard

은행 소스코드 취약점 점검 봇. 특정 은행에 종속되지 않는 은행권 범용 도구다. 개발 GitLab의 MR에 취약점 근거와 수정 가이드를 인라인 코멘트로 단다.

## 아키텍처

```
개발 GitLab (MR webhook)
  → finguard: Semgrep 실행 → SARIF 수신
  → finguard: 룰ID로 금보원 매핑 조회 → rdjsonl 생성
  → reviewdog: diff 필터 · 중복 제거 · MR discussion POST
  → 개발 GitLab MR 인라인 코멘트

```

finguard은 보안 담당 조직의 리눅스 서버에서 상시 구동된다. CI 러너 안이 아니다. reviewdog은 외부 바이너리로 호출한다. **포크하거나 수정하지 않는다.**

## 역할 경계 (중요)

finguard이 담당하는 것:

- webhook 수신 및 검증
- 대상 소스 확보 (GitLab API)
- Semgrep 실행 및 SARIF 파싱
- **룰ID → 금보원/개발보안가이드 매핑 (이 프로젝트의 고유 자산)**
- rdjsonl 생성 및 stdout 출력
- reviewdog 프로세스 실행

finguard이 하지 않는 것:

- diff 라인 필터링 (reviewdog `-filter-mode=added`가 처리)
- 중복 코멘트 제거 (reviewdog이 처리)
- GitLab Discussion API 직접 호출 (reviewdog이 처리)
- 코드 자동 수정 (가이드만 제시, 수정은 개발자 몫)

## 기술 스택

- Go 1.22+
- 외부 의존성 최소화. 표준 라이브러리 우선
- Semgrep: OSS 버전, 커스텀 룰 YAML
- reviewdog: 바이너리 호출 (`os/exec`)

## 디렉토리

```
cmd/finguard/main.go
internal/
  webhook/   MR 이벤트 수신 · secret token 검증
  gitlab/    diff 조회 · 소스 clone
  scanner/   semgrep 실행 · SARIF 파싱
  mapping/   룰ID → 금보원 항목 (YAML 로드)
  rdjson/    SARIF + 매핑 → rdjsonl 출력
  runner/    reviewdog 프로세스 파이프
mapping/rules.yaml

```

## reviewdog 호출 규약

```bash
finguard scan --mr-iid=$MR_IID --project=$PROJECT_ID \
  | reviewdog -f=rdjsonl \
      -name="finguard" \
      -reporter=gitlab-mr-discussion \
      -filter-mode=added \
      -fail-level=error

```

필수 환경변수:

- `REVIEWDOG_GITLAB_API_TOKEN` — 봇 계정 PAT (api 스코프, Developer 이상)
- `CI_API_V4_URL`, `CI_PROJECT_ID`, `CI_MERGE_REQUEST_IID`, `CI_COMMIT_SHA`

CI 밖에서 실행하므로 위 변수는 finguard이 webhook payload에서 뽑아 세팅한다.

## rdjsonl 출력 형식

한 줄에 JSON 하나, 코멘트 하나. `message`는 마크다운을 그대로 받는다.

```json
{"source":{"name":"finguard"},"severity":"ERROR","location":{"path":"src/main/java/UserDao.java","range":{"start":{"line":42,"column":9},"end":{"line":42,"column":68}}},"message":"..."}

```

severity는 `ERROR` / `WARNING` / `INFO` 중 하나.

## 매핑 테이블 스키마

`mapping/rules.yaml`:

```yaml
- rule_id: java.lang.security.audit.sqli.jdbc-sqli   # Semgrep 룰 ID
  code: FIN-SQLI-001
  cwe: CWE-89
  title: SQL 삽입
  severity: ERROR
  basis: 개발보안가이드 「입력데이터 검증 및 표현」
  kisa_item: (금보원 점검항목)
  explain: 사용자 입력이 문자열 연결로 쿼리에 삽입되고 있습니다.
  fix_example: |
    ```java
    PreparedStatement ps = conn.prepareStatement(
        "SELECT * FROM USER WHERE ID = ?");
    ps.setString(1, userId);
    ```

```

매핑에 없는 룰ID가 나오면 코멘트를 생략한다. 임의 문구를 지어내지 않는다.

## webhook 스킵 규칙

다음 이벤트는 스캔하지 않고 사유를 로그로 남긴다.

- `merge_request` 가 아닌 이벤트, `open`/`reopen`/`update` 외의 action
- `update` 인데 `object_attributes.oldrev` 가 비어 있는 경우 (라벨/제목/설명 변경 — 코드 푸시가 아님)
- Draft/WIP MR (`draft` 또는 `work_in_progress`)
- 동일 `projectID/iid` 스캔이 이미 진행 중인 경우 (inflight 중복 방지 — 다음 푸시 webhook이 다시 온다)

## commit status 게이트

스캔이 정상 완료되면 GitLab commits API(`POST /projects/:id/statuses/:sha`)로 name=`finguard` 의 commit status를 게시한다.

- 매핑된 finding 중 차단 대상 severity가 있으면 `failed`, 없으면 `success`
- clone/스캔/reviewdog 실패 시에는 status를 게시하지 않는다 (로그만 남김)
- 차단 대상 severity는 env `FINGUARD_BLOCK_ON` (쉼표 구분, 예: `ERROR,WARNING`)
  - 미설정 시 기본값 `ERROR`
  - 빈 문자열로 명시하면 전역 게이트 off (항상 `success`)
- GitLab 프로젝트 설정에서 "Pipelines must succeed"를 켜면 `failed` status가 머지를 차단한다

## `.finguard.yml` 레포별 오버라이드

점검 대상 레포 루트에 `.finguard.yml` 을 두면 레포별로 동작을 조정할 수 있다. 파일이 없으면 전역 기본만 적용되고, 파싱에 실패하면 무시하고 전역 기본으로 진행한다.

```yaml
# 점검에서 제외할 파일 glob (path.Match 문법).
# 슬래시 없는 패턴은 디렉터리 무관하게 파일명에 매칭된다.
ignore:
  - "*.sql"
  - "vendor/*"
  - "generated/**"

# 차단 게이트 severity 오버라이드.
# 필드를 생략하면 전역(FINGUARD_BLOCK_ON)을 따르고,
# 빈 배열([])로 명시하면 이 레포만 게이트 off.
block_on: [ERROR, WARNING]
```

- `ignore` 는 rdjsonl 생성 전에 finding 파일 경로(레포 상대)에 적용된다 — 걸린 파일은 코멘트도, 게이트 집계도 제외
- `block_on` 은 대소문자 무관하게 정규화된다

## 코멘트 문구 규칙

```
[FIN-SQLI-001] SQL 삽입 (CWE-89)

▸ 근거: 개발보안가이드 「입력데이터 검증 및 표현」
▸ 심각도: 상

{explain}

수정 예시:
{fix_example}

```

근거(`basis`, `kisa_item`)는 반드시 매핑 테이블 값을 그대로 쓴다. LLM이나 추론으로 생성하지 않는다. 감사 대응 자료이므로 출처가 확정되어야 한다.

## 제약

- 폐쇄망. `go get` 불가. 의존성은 vendor 디렉토리로 커밋한다.
- 외부 API 호출 없음. 사내 GitLab만 접근.
- 로그에 소스코드 본문을 남기지 않는다. 파일경로와 라인번호까지만.
- 에러 시 조용히 죽지 말 것. webhook은 200을 반환하고 실패는 로그로 남긴다.

## 테스트

- Semgrep과 reviewdog은 인터페이스로 추상화해 목(mock) 가능하게 한다.
- SARIF 샘플과 기대 rdjsonl을 `testdata/`에 두고 골든 파일 테스트.
- 실제 GitLab 호출은 통합 테스트로 분리. 기본 `go test`에서 제외.

## 작업 방식

- 한 번에 한 패키지씩. 완료 후 빌드와 테스트 통과를 확인하고 다음으로 넘어간다.
- 스펙이 불확실하면 추측하지 말고 질문한다. 특히 SARIF 필드명, rdjsonl 필드명, GitLab API 응답 구조는 실물 확인 없이 지어내지 말 것.

