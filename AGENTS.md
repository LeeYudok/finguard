# AGENTS.md — finguard

이 파일은 이 저장소에서 작업할 때 AI 에이전트(Claude Code·Gemini CLI·Codex 등)가 따르는
**단일 진실원천(SSOT)** 이다. 프로젝트 브레인. `CLAUDE.md`·`GEMINI.md` 는 이 파일을 참조한다.

@.claude/memory/MEMORY.md

## 메모리 경로 오버라이드

이 프로젝트의 auto-memory SSOT는 `.claude/memory/` 이다.
- 시스템 기본 경로(`~/.claude/projects/.../memory/`)는 사용하지 않는다.
- 모든 메모리 읽기/쓰기는 `.claude/memory/` 하위에서 수행한다.
- `MEMORY.md` 가 인덱스(단일), 타입접두 `project_`/`feedback_`/`reference_`/`user_`.
- `user_*.md` 만 개인(gitignore), 그 외는 팀 공유.

## .claude/ 인프라

개요는 [.claude/README.md](.claude/README.md), 각 하위 디렉터리 README 에 작성 골격과
컨벤션이 있다.

| 디렉터리 | 역할 | 상세 |
| :--- | :--- | :--- |
| `agents/` | 서브에이전트 정의 (code-reviewer, security-audit, db-migration, sdlc-*, agent-evolve) | [README](.claude/agents/README.md) |
| `commands/` | 커스텀 슬래시 커맨드 (fix-issue, sdlc-cycle, sonar, knowledge-graph) | [README](.claude/commands/README.md) |
| `hooks/` | 강제 게이트 — pre-commit, 자동 포맷, observe-lite, 메모리 리마인드 | [README](.claude/hooks/README.md) |
| `memory/` | 프로젝트 메모리 SSOT — MEMORY.md 인덱스 + 타입접두 파일 | [README](.claude/memory/README.md) |
| `rules/` | 맥락 인지 룰 — `paths:` 스코프 조건부 로드 | [README](.claude/rules/README.md) |
| `skills/` | 상황별 절차 — review, status, search-first, memory-factcheck, security-precheck, grill-me 등 | [README](.claude/skills/README.md) |
| `workflows/` | 저장형 Workflow 오케스트레이션 스크립트(`*.js`) — rules-audit 예제 | [README](.claude/workflows/README.md) |
| `scripts/` | 레포 로컬 헬퍼 — knowledge_graph.py(문서 그래프 + 링크 체커) | [README](.claude/scripts/README.md) |

`settings.json` 에 deny 규칙과 훅 바인딩이 있다. 문서 정합성 점검·온보딩은
`/knowledge-graph` 실행.

## 프로젝트 개요

은행 소스코드 취약점 점검 봇. 특정 은행에 종속되지 않는 은행권 범용 도구다. 개발 GitLab의 MR에 취약점 근거와 수정 가이드를 인라인 코멘트로 단다.

## 스택

- Go 1.22+ · 표준 라이브러리 우선, 외부 의존성 최소화 (현재 `gopkg.in/yaml.v3` 1개)
- Semgrep OSS(커스텀 룰 YAML) · reviewdog — 둘 다 외부 바이너리 호출, 포크·수정하지 않는다
- 폐쇄망 전제: `go get` 불가, 의존성은 `vendor/` 커밋

## 명령

```bash
go build -mod=vendor ./...     # 빌드
go test -race -mod=vendor ./... # 테스트 (레이스 감지 포함)
make build-linux               # 폐쇄망 반입용 정적 바이너리
semgrep --validate --config rules/   # 룰 문법 검증
```

## 코드 탐색 — CodeGraph

이 레포는 CodeGraph 로 인덱싱돼 있다(`.codegraph/`, 전역 gitignore 로 커밋 제외).
코드 위치를 찾거나 흐름을 파악할 때 grep/find 보다 먼저 쓴다.

```bash
codegraph explore "<질문 또는 심볼명>"   # 관련 심볼 소스 + 호출 경로 한 번에
codegraph sync .                        # 코드 변경 후 인덱스 갱신
```

Semgrep·reviewdog 은 인터페이스로 추상화돼 있어 호출 흐름이 grep 으로 안 잡힌다
(`scanner.Semgrep` → `CLI`, `runner.Reviewdog` → `CLI`). CodeGraph 는 이 동적 디스패치
연결과 "테스트가 덮지 않는 심볼"까지 같이 알려준다.

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

### 룰 회귀 픽스처

새 룰을 만들거나 기존 룰을 고치면 **정탐 픽스처와 오탐 방지 픽스처를 함께** 넣는다.

- 위치는 **`testdata/rule-fixtures/`** — `rules/` 안에 두면 안 된다. semgrep 은 `--config <디렉터리>`
  하위의 모든 `.yml`/`.yaml` 을 룰 파일로 재귀 파싱하므로, yaml 픽스처가 하나라도 섞이면
  룰셋 전체가 `0 rule(s)` 로 로드 실패한다 (#43, 회귀 테스트 `TestNoYAMLFixturesUnderRules`).
- 기대값은 테스트가 아니라 **픽스처 안의 마커 주석**에 적는다 (#44). 검출이 기대되는 줄
  **바로 위**에 한 줄:

  ```python
  # EXPECT: finguard.python.cleartext-websocket
  REALTIME_FEED_URL = "ws://ops.example-broker.co.kr:21000"
  ```

  마커 뒤에 설명을 덧붙이면 매칭되지 않는다(룰ID 만 온다). `//` 주석도 같은 문법.
- **`safe_` 접두 파일은 대조군**이라 마커를 가질 수 없다 — 어떤 룰에도 걸리지 않아야 한다.
- `TestFixtureExpectationsMatchScan` 이 양방향으로 검사한다. 마커 있는데 미검출 = 미탐 회귀,
  마커 없는데 검출 = 오탐 회귀. 라인 번호를 테스트에 적지 않으므로 블록을 어디에 추가하든
  기대값이 따라 움직이고, 픽스처를 건드리는 PR 끼리 충돌하지 않는다.
- 통합 테스트 실행: `go test -race -mod=vendor -tags semgrep_integration ./internal/scanner/`
- 레포 루트의 **`.semgrepignore` 를 지우지 말 것** (#25). semgrep 은 이 파일이 없으면
  내장 기본 무시목록을 쓰는데 거기에 `test/`·`tests/`·`*_test.go` 가 들어 있어,
  `*_test.go`·`*Test.java` 픽스처와 `finguard.<lang>.hardcoded-secret-test` 룰이
  룰 설정과 무관하게 통째로 사라진다. 운영 스캔에서는 `internal/scanner` 가 대상 루트에
  같은 파일을 심는다(레포가 자기 것을 갖고 있으면 건드리지 않는다).

## 작업 방식

- 한 번에 한 패키지씩. 완료 후 빌드와 테스트 통과를 확인하고 다음으로 넘어간다.
- 스펙이 불확실하면 추측하지 말고 질문한다. 특히 SARIF 필드명, rdjsonl 필드명, GitLab API 응답 구조는 실물 확인 없이 지어내지 말 것.

## 컨벤션

- skill/agent 신규 생성 시 `finguard-` prefix 네임스페이스
- 세부 규약은 `.claude/rules/` 의 paths 스코프 룰 참조
- 스택별 세부 규약 → 아래 우선순위 체계 참조

## 우선순위 체계 (P0/P1/P2)

### P0 — 절대 규칙 (AI/사람 모두, 예외 없음)
P0 위반 시 즉시 작업 중단 + 사용자 에스컬레이션.

- **보안**: 시크릿/토큰/비밀번호를 코드·로그·이슈에 노출 금지
- **데이터**: 프로덕션 DB에 `DELETE/DROP/TRUNCATE` 전 사용자 명시 동의
- **git**: `force push` / `reset --hard` 전 확인. `.env` 스테이징 금지
- **인증**: 인증 없는 API 엔드포인트 신규 추가 금지
- **스택별 P0**: 각 `.claude/rules/<stack>.md` 의 `## P0` 섹션 참조

### P1 — 필수 (AI 자율 실행 범위, 위반 시 PR 차단)

- 이슈 번호를 브랜치명·커밋·PR/MR 제목에 반드시 포함
- 커밋 전 타입체크·린트 통과 (`.claude/hooks/pre-commit.sh` 자동 게이트)
- 새 기능에 최소 1개 테스트 동반
- `main`/`develop` 직접 커밋 금지 → 항상 feature/fix/chore 브랜치

### P2 — 권장 (리뷰 지적 사항, 예외 협의 가능)

- 함수당 인지 복잡도(CC) 15 이하 (`.claude/hooks/cc-check.py` 경고)
- 파일 1개 = 단일 책임 (300줄 초과 시 분리 검토)
- TODO/FIXME 에 이슈 번호 병기

## 워크플로

1. **이슈 등록** → 2. **브랜치 생성** (`feat/issue-<N>-<slug>`) → 3. **구현** →
4. **pre-commit 자동 게이트 통과** → 5. **PR/MR 생성** → 6. **리뷰** → 7. **머지 + 이슈 클로즈**

이슈 클로즈 규약은 forge 별로 다르다 → `.claude/rules/forge.md` 참조.
GitHub = PR 본문 `Closes #N` 으로 머지 시 자동 클로즈. GitLab 19 = `Closes #N` 자동 클로즈 실측 동작 — 단 머지 후 `glab issue view <N>` 로 확인하고, `opened` 로 남은 경우에만 수동 클로즈.

## 멀티 에이전트 · 병렬 세션

이 레포를 동시에 만지는 모든 워커(세션·서브에이전트·페르소나)는 **각자의 git worktree**
로 격리한다 — 이 섹션은 공유 체크아웃을 쓰던 병렬 세션 두 개가 서로의 작업을
교차오염시킨 사고에서 나왔다.

- 선제 격리: `git worktree add ../finguard-<slug> -b <type>/issue-<N>` —
  1세션 = 1worktree = 1이슈 = 1브랜치.
- 정식 클론은 default 브랜치 미러로 유지(pull·읽기만 — 거기서 `checkout`/`switch`
  **금지**; 공유 폴더에서 브랜치를 갈아타는 순간 다른 세션의 발밑이 바뀐다).
- `git add` 는 명시 파일만, 디렉터리·`-A` 금지 — 다른 세션의 미커밋 작업을 흡수하지 않기 위함.
- `git status` 에 내가 만들지 않은 변경이 보이면 진행 전에 병렬 세션 여부부터 확인.
- 병렬 서브에이전트가 파일을 동시에 수정하면 `isolation: "worktree"` 필수.
- 머지 후: worktree 제거 + 로컬 브랜치 삭제를 그 자리에서 항상 수행.
- worktree 오케스트레이터(예: Orca) 사용 시: worktree 생성·정리는 해당 도구에 위임 —
  오케스트레이터가 소유한 worktree 를 수동 `git worktree add`/`remove` 로 만지지 않는다
  (도구 상태와 어긋남). 격리 원칙(1세션 = 1worktree = 1이슈 = 1브랜치)은 동일하게
  적용되며, "머지 후 제거" 규칙은 오케스트레이터의 자체 정리로 충족한다.

역할별 에이전트: `.claude/agents/sdlc-*.md` (developer/tester/verifier).
SDLC 자동화: `/sdlc-cycle` 명령 참조.
