---
paths:
  - "rules/**"
  - "testdata/rule-fixtures/**"
  - "internal/scanner/**"
  - ".semgrepignore"
  - ".github/secret_scanning.yml"
---

# Semgrep 룰 · 회귀 픽스처 규약 (포인터)

규약 본문은 디렉터리 소유 문서 **[`rules/AGENTS.md`](../../rules/AGENTS.md)** 에 있다
(Claude 외 에이전트도 읽도록 `.claude/` 밖에 둔다). 룰·픽스처·`internal/scanner` 를
편집하기 전에 그 파일을 읽을 것. 이 스텁은 Claude Code 가 경로 매칭 시 자동 로드하기 위한 것이다.
