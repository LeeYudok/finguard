# memory/ — 프로젝트 메모리 (SSOT)

이 프로젝트 auto-memory **단일 진실원천**. 시스템 기본 경로 미사용, 모든 메모리는 여기서.

## 장기 기억 도구 = graphify (하나만)

코드·문서·이슈 이력의 장기 기억은 **graphify** (`graphify-out/`, gitignore) 로 통일한다 (#84).
빌드·질의는 루트 `AGENTS.md` 「코드 탐색 · 장기 기억」 참조. 이 디렉터리의 파일 메모리는
그래프가 추출할 수 없는 것 — 사고 회고, 사용자 피드백, 외부 포인터 — 만 남긴다. 코드 구조나
"어떤 파일이 뭘 하나"는 여기 쓰지 말고 `graphify query` 로 답한다.

## 규칙

- `MEMORY.md` — 인덱스(단일). 메모리 1개 = 파일 1개, 한 줄 포인터를 인덱스에 추가.
- 타입 접두:
  - `project_*` — 진행 작업·목표·제약(코드/git에서 안 드러나는 것)
  - `feedback_*` — 작업 방식 지침(왜 + 적용법)
  - `reference_*` — 외부 리소스 포인터(URL·대시보드·티켓)
  - `user_*` — 개인 메모리, **gitignore 제외**(팀 비공유). 그 외 팀 공유.
  - `instinct_*` — 세션 관찰에서 추출된 습관(instinct-lite). frontmatter 에
    `trigger`(발동 조건)/`confidence`(0.3 잠정 ~ 0.9 확신)/`evidence`(근거 관찰),
    본문에 행동 1개. 같은 습관 재관찰 시 confidence 상향, 반증 시 하향/삭제.
- 단정형 사실(임계값·active flag)은 날짜를 박고, 행동 전 code/DB로 재검증(decay).
- `observations/` — PostToolUse 훅(observe-lite.sh)의 세션 관찰 JSONL. **gitignore**(비커밋),
  7일 자동 정리. Stop 훅이 이 로그로 instinct 추출을 유도한다.
- 신선도 감사(stale 정정·dead 아카이브)는 `memory-factcheck` 스킬로. 아카이브 파일은
  `archive/` 로 이동(`archived:` frontmatter 추가)하고 인덱스에서 제거 — 삭제 금지.
