---
name: gitlab-mirror-review-bot
description: gitlab.doksam.com busan/finguard 는 review-bot 2차 리뷰용 미러 — SSOT 는 GitHub, 머지 후 git push gitlab main, 리뷰 원하면 GitLab 에 MR
metadata:
  type: project
---

`gitlab.doksam.com/busan/finguard`(project id 151)는 2026-08-25 에 만든 **review-bot 실험·2차 리뷰용 미러**다. 로컬 리모트 이름 `gitlab`(`ssh://git@gitlab.doksam.com:2222/busan/finguard.git`).

- SSOT 는 여전히 GitHub `LeeYudok/finguard`. 이슈·PR·머지는 GitHub 에서만.
- GitHub 머지 후 `git push gitlab main` 으로 미러 갱신. 미러 main 에 직접 커밋 금지.
- review-bot 리뷰를 받고 싶으면 브랜치를 `gitlab` 에 푸시하고 **GitLab 에 MR** 을 연다(시스템 훅이 전역이라 별도 설정 없음). 리뷰 코멘트는 `<!-- review-bot:v2 -->` 마커, commit status `review-bot`. 약 1분.
- `.reviewbot.yml` 의 `context_files: [AGENTS.md, CONTEXT.md]` 가 있어야 프로젝트 제약을 근거로 P0 를 잡는다 — main 반영은 GitHub 이슈 참조.
- 실험 MR !1(`test/reviewbot-trial`)은 머지하지 않는다. review-bot 의 "P0 두 건 병합" 미탐은 busan/review-bot 이슈로 등록.

**Why:** finguard 는 GitHub 에 있어 사내 review-bot 이 못 보므로 미러가 필요했고, CONTEXT.md 를 리뷰 근거로 쓰는 실험이 성공해 유지하기로 함.
**How to apply:** 머지 마무리 루틴에 `git push gitlab main` 추가. 미러 상태가 의심되면 `git ls-remote --heads gitlab`.
