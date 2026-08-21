---
name: project-finguard-scan-corpus
description: 룰 검증용 실코드 코퍼스 고르는 법 — 국내 금융 Swift/Kotlin 고스타 레포는 존재하지 않음, 대안 목록과 스캔 운용 요령
metadata:
  type: project
---

finguard 룰을 실코드로 검증할 때 쓸 코퍼스 선정 기준. 2026-08-17 GitHub 전수 탐색으로 확인했다.

**"국내 + 금융 + Swift/Kotlin + 고스타" 교집합은 사실상 없다.** 한글 금융 키워드(`금융`/`주식`/`증권`/`가계부`/`은행`)로 Swift·Kotlin 을 검색하면 **최고 스타가 17**이고 부트캠프 과제·습작이 대부분이다. 반대로 국내 조직 고스타 레포(toss 11k, naver 7k, line 5k, kakao 1.4k)는 전부 TS/Python 이고 금융 도메인이 아니다. 셋 중 둘까지만 만족시킬 수 있으니 무엇을 우선할지 먼저 사용자에게 확정받아라.

**Why**: "스타 많은 국내 금융 Swift 레포 찾아줘" 를 그대로 수행하려 하면 검색만 반복하다 빈손이 된다. 전제 자체가 성립하지 않는다는 사실을 먼저 보고하는 게 맞다.

**How to apply**:

- 금융 도메인 우선으로 갈 때 실제로 쓸 만한 10개(2026-08-17 기준 스타):
  `koreainvestment/open-trading-api`(1563, Py) · `samchon/payments`(361, TS) ·
  `DingdingKim/CoinNow`(219, Swift) · `Gosrock/DuDoong-Backend`(151, Kotlin, 토스페이먼츠 연동) ·
  `iamport/iamport-rest-client-python`(133) · `banksalad/AXSnapshot`(111, Swift) ·
  `namjug-kim/reactive-crypto`(101, Kotlin) · `sharebook-kr/pybithumb`(95) ·
  `iamport/iamport-rest-client-java`(91) · `techinpark/upbitBar`(22, Swift)
- 검색 축은 한글 키워드보다 **국내 PG·증권·거래소 SDK 이름**(iamport/portone, tosspayments, bootpay, upbit, bithumb, 한국투자 KIS)이 훨씬 잘 걸린다.
- **스캔 결과의 1차 필터는 항상 테스트 코드다.** orca 스캔에서 202건 중 182건(90%)이 `.test.`·`fixtures/` 의 더미 값이었다. 트리아지 전에 경로로 먼저 갈라야 실제 신호가 보인다.
- **백그라운드 에이전트가 이 레포를 만지는 동안 스캔하려면 룰을 커밋에서 떠서 고정한다** — 워킹트리의 `rules/` 를 그대로 쓰면 에이전트가 픽스처를 쓰는 중간 상태가 섞인다.
  `git archive <sha> rules mapping | tar -x -C <tmpdir>` 후 `--rules=<tmpdir>/rules` 로 실행.
- **⚠️ 코퍼스를 `/private/tmp` 스크래치패드에 두면 증발한다.** macOS 가 주기적으로 비운다. 무서운 점은 **조용히 실패한다**는 것 — 디렉터리 껍데기는 남고 파일만 사라져서, semgrep 이 `Ran N rules on 0 files: 0 findings` 를 반환하고 그게 "깨끗한 레포" 와 구분되지 않는다. 2026-08-22 에 실제로 r1 스캔이 0건으로 나와 발견했다(직전까지 8건이었음).
  스캔 결과를 신뢰하기 전에 **대상 파일 수를 먼저 확인**해라: `semgrep ... --json` 의 `paths.scanned` 길이 또는 `find <dir> -name '*.<ext>' | wc -l`. 0 이면 결과가 아니라 사고다. 며칠 걸치는 캠페인이면 코퍼스를 `/private/tmp` 밖(예: 프로젝트 내 gitignore 디렉터리)에 두는 편이 낫다.
- 관련: [[project-gh-merge-branch-cleanup]], [[project-adversarial-verify-recheck]]
