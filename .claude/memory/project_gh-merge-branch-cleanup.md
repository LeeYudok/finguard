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
gh pr merge <N> --squash
sleep 3 && gh pr view <N> --json state -q .state   # ★ MERGED 를 확인한 뒤에만 다음 줄로
git push origin --delete <branch>                  # 이게 실제 삭제
git worktree remove <path>                         # 워크트리가 물고 있으면 로컬 삭제 전에
git branch -D <branch>
git remote prune origin && git branch -a           # 남은 게 없는지 눈으로 확인
```

- `git branch -a` 확인까지 해야 끝이다. prune 만으로는 원격 실물이 안 지워진다.
- 서브에이전트가 워크트리에서 작업했으면 브랜치 삭제 전에 `git worktree remove` 가 선행돼야 한다(`cannot delete branch ... used by worktree` 로 막힘).

**★ 머지 성공을 확인하기 전에 원격 브랜치를 지우지 마라.** 2026-08-21 에 PR #56·#57 을 이렇게 날려먹었다(2회 반복). `gh pr merge` 가 실패했는데도 뒤이은 `git push origin --delete` 가 실행돼 **PR 이 닫히고 re-open 도 거부된다**(`Cannot change the base branch of a closed pull request`). 복구는 로컬 ref 로 브랜치를 되살려 **새 PR 을 다시 여는 것뿐**이다. `--delete-branch` 플래그는 이 레포에서 원격을 안 지우므로(위 본문) 어차피 수동 삭제가 필요한데, 그 수동 삭제를 조건부로 걸어야 한다.

**스택드 PR 은 머지 전에 base 를 `main` 으로 돌려놔라.** 서브에이전트가 `A ← B ← C` 로 쌓아 올린 PR 은, A 를 머지하며 A 브랜치를 지우는 순간 B 의 base 가 사라져 `CONFLICTING/DIRTY` 가 된다. 머지 시작 전에 `gh pr edit <B> --base main` 으로 전부 재지정한 뒤 순서대로 처리하면 이 문제가 사라진다. 각 단계에서 `git rebase origin/main` + 매핑 파일 tail 충돌 해소가 필요하다.

- 관련: [[project-finguard-scan-corpus]], [[project-adversarial-verify-recheck]]
