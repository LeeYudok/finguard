---
name: project-gh-merge-branch-cleanup
description: 이 레포에서 gh pr merge --delete-branch 는 원격 브랜치를 안 지운다 — 머지 후 항상 수동 삭제 + prune 확인
metadata:
  type: project
---

`gh pr merge <N> --squash --delete-branch` 를 써도 **원격 브랜치가 남는다.** 2026-08-17 PR #21·#22 에서 연속 재현했다. 로컬 브랜치 삭제만 시도하고(워크트리가 물고 있으면 그것마저 실패), 원격은 그대로다.

**Why**: 머지 후 `git branch -a` 에 남은 `remotes/origin/<branch>` 를 stale ref 로 오해해 `git remote prune origin` 만 돌리면 아무 일도 일어나지 않는다. `gh api repos/<owner>/<repo>/branches` 로 확인하면 원격에 실제로 살아 있다. 쌓이면 다음 세션이 "브랜치 정리" 부터 해야 한다 — 실제로 이 세션이 머지된 브랜치 9개를 치우며 시작했다.

**How to apply**: 머지 직후 항상 이 순서로 마감한다.

```bash
gh pr merge <N> --squash --delete-branch          # 실패해도 계속
git worktree remove <path>                        # 워크트리가 브랜치를 물고 있으면 먼저
git branch -D <branch>
git push origin --delete <branch>                 # 이게 실제 삭제
git remote prune origin && git branch -a          # 남은 게 없는지 눈으로 확인
```

- `git branch -a` 확인까지 해야 끝이다. prune 만으로는 원격 실물이 안 지워진다.
- 서브에이전트가 워크트리에서 작업했으면 브랜치 삭제 전에 `git worktree remove` 가 선행돼야 한다(`cannot delete branch ... used by worktree` 로 막힘).
- 관련: [[project-finguard-scan-corpus]]
