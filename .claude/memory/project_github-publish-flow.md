---
name: project-github-publish-flow
description: 2026-08-17부터 GitHub 가 개발 SSOT — PR 플로우·noreply 커밋 신원·브랜치 보호. 구 fresh-history force 퍼블리시는 폐기됨
metadata:
  type: project
---

**2026-08-17 부로 GitHub(`LeeYudok/claude-scaffold`)가 개발 SSOT 다** (Contributors 유치 결정, A안). 이슈·PR·CI·릴리즈 전부 GitHub 에서. 내부 GitLab(busan/claude-scaffold)은 아카이브/백업 — 새 작업을 그쪽 main 에 쌓지 말 것.

**Why**: 구 구조(내부 GitLab SSOT + fresh-history 스냅샷 force 퍼블리시)는 기여자 커밋이 재퍼블리시마다 증발해 Contributors 그래프가 영구 공백이었다. 스냅샷 커밋 a9fd324 를 루트로 히스토리 누적으로 전환.

**How to apply**:
- 작업 브랜치는 `github/main` 기반으로 생성(`git switch -C <branch> github/main`), 푸시는 `github` 리모트로, 머지는 GitHub PR 로. `origin`(GitLab) main 으로 신규 작업 푸시 금지.
- **커밋 신원**: GitHub 로 나가는 워크트리에는 `git config user.name leeyudok`, `user.email leeyudok@users.noreply.github.com` 설정 — 개인 이메일 유출 방지.
- **main 보호 활성**(2026-08-17): force push·삭제 차단, `bats` 체크 필수. 직접 push 대신 PR.
- `scripts/publish-github.sh` 는 **폐기**(내부 GitLab 트리에만 존재). 다시 쓰지 말 것 — force push 는 보호 규칙에도 막힌다.
- 리뷰봇 워크플로(claude-review/assistant)는 repo 변수 `CLAUDE_REVIEW_ENABLED == 'true'` 일 때만 실행 — 켜려면 `ANTHROPIC_API_KEY` 시크릿 + 그 변수 설정(비용 발생, 사용자 결정).
- 내부 전용 정보(고객사명·내부 호스트)는 이제 커밋 전 스크립트 가드가 없다 — 작성 단계에서 차단이 유일한 방어. 메모리 `user_*` 규칙 유지.
- 과거 사고 교훈(2026-08-16, 구조 전환 전): 내부 main 을 github 에 직push 시도(비FF 거부가 방어였음), publish 트리 제외 파일의 테스트가 공개 CI 를 죽임(#46 skip 가드).
