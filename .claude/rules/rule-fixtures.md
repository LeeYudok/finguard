---
paths:
  - "rules/**"
  - "testdata/rule-fixtures/**"
  - "internal/scanner/**"
  - ".semgrepignore"
  - ".github/secret_scanning.yml"
---

# Semgrep 룰 · 회귀 픽스처 규약

새 룰을 만들거나 기존 룰을 고치면 **정탐 픽스처와 오탐 방지 픽스처를 함께** 넣는다.

## 픽스처 작성 규칙

- 위치는 **`testdata/rule-fixtures/`** — `rules/` 안에 두면 안 된다. semgrep 은 `--config <디렉터리>`
  하위의 모든 `.yml`/`.yaml` 을 룰 파일로 재귀 파싱하므로, yaml 픽스처가 하나라도 섞이면
  룰셋 전체가 `0 rule(s)` 로 로드 실패한다 (#43, 회귀 테스트 `TestNoYAMLFixturesUnderRules`).
- 기대값은 테스트가 아니라 **픽스처 안의 마커 주석**에 적는다 (#44). 검출이 기대되는 줄
  **바로 위**에 한 줄:

  ```python
  # EXPECT: finguard.python.cleartext-websocket
  REALTIME_FEED_URL = "ws://ops.example-broker.co.kr:21000"
  ```

  마커 뒤에 설명을 덧붙이면 매칭되지 않는다(룰ID 만 온다).
- **주석 형태 3종** — 파일 문법에 맞는 것을 쓴다.

  | 형태 | 대상 |
  | :--- | :--- |
  | `# EXPECT: <룰ID>` | 파이썬·셸·YAML·properties |
  | `// EXPECT: <룰ID>` | Go·Java·Kotlin·Swift·TS |
  | `<!-- EXPECT: <룰ID> -->` | XML·plist (#61) |

  XML 은 한 줄로 여닫는다. 주석을 다음 줄에서 닫거나 `<!--` / `# EXPECT:` / `-->` 로
  쪼개던 우회책은 폐기했다 (#61). 닫기(`-->`) 앞뒤에 설명을 붙이면 다른 형태와 똑같이
  매칭되지 않는다.
- **하위 디렉터리도 순회한다** (#60). 마커 파서(`walkFixtures`)와 스캔(`CLI.Scan`)의 순회
  범위가 같으므로, `paths.include` 로 경로 구조 자체를 판정 기준으로 삼는 룰
  (`finguard.swift.insecure-trust-vendor` 등)도 벤더 경로 하위에 픽스처를 두고 검증한다.
  기대값 키는 픽스처 루트 기준 **상대경로**라 다른 디렉터리의 동명 파일이 섞이지 않는다.
  - 단 `DefaultExcludes`(`internal/scanner/exclude.go`)에 걸리는 경로는 스캔 자체에서
    빠진다 — `Pods/` 하위 픽스처는 `--exclude Pods` 때문에 검출되지 않으므로
    `Carthage/` 를 쓴다.
  - 바이너리(NUL 바이트 포함 파일)는 마커 순회에서 건너뛴다.
- **`safe_` 접두 파일은 대조군**이라 마커를 가질 수 없다 — 어떤 룰에도 걸리지 않아야 한다.
  하위 디렉터리의 `safe_*` 도 같은 기준으로 검사한다.
- `TestFixtureExpectationsMatchScan` 이 양방향으로 검사한다. 마커 있는데 미검출 = 미탐 회귀,
  마커 없는데 검출 = 오탐 회귀. 라인 번호를 테스트에 적지 않으므로 블록을 어디에 추가하든
  기대값이 따라 움직이고, 픽스처를 건드리는 PR 끼리 충돌하지 않는다.
- 마커 문법·순회 구현은 **`internal/scanner/expect_marker_test.go` 한 곳**에만 있다.
  이 파일에는 빌드 태그가 없어 기본 `go test` 와 `-tags semgrep_integration` 양쪽에서
  컴파일되고, 태그 뒤의 `integration_test.go` 가 그 정의를 그대로 쓴다. 예전처럼 문법을
  양쪽에 복제하면 어긋난 순간 마커가 조용히 무시되고 통합 테스트가 "기대 0건 · 검출 0건"
  으로 통과한다 — 복제하지 말 것 (#60·#61).
- 통합 테스트 실행: `go test -race -mod=vendor -tags semgrep_integration ./internal/scanner/`
- 레포 루트의 **`.semgrepignore` 를 지우지 말 것** (#25). semgrep 은 이 파일이 없으면
  내장 기본 무시목록을 쓰는데 거기에 `test/`·`tests/`·`*_test.go` 가 들어 있어,
  `*_test.go`·`*Test.java` 픽스처와 `finguard.<lang>.hardcoded-secret-test` 룰이
  룰 설정과 무관하게 통째로 사라진다. 운영 스캔에서는 `internal/scanner` 가 대상 루트에
  같은 파일을 심는다(레포가 자기 것을 갖고 있으면 건드리지 않는다).

## 시크릿 형태 픽스처는 외부 스캐너를 트립시킨다

시크릿 탐지 룰의 정탐 픽스처는 정의상 "시크릿처럼 생긴 문자열"이고, GitHub 시크릿
스캐닝의 탐지기도 값의 유효성이 아니라 **형태만** 보므로 반드시 함께 걸린다 (#82).

- **픽스처 값의 형태를 바꿔 회피하지 마라.** 형태가 깨지면 우리 룰도 못 잡아 정탐
  픽스처의 목적이 사라진다.
- 대신 `.github/secret_scanning.yml` 의 `paths-ignore` 가 `testdata/rule-fixtures/**`
  를 제외한다. 픽스처는 반드시 그 아래에 둔다.
- 픽스처에 **실제 자격증명을 넣지 마라.** 합성값을 쓰고, 형태만 룰이 요구하는 대로
  맞춘다(예: 문자 클래스를 규칙적으로 교대시키면 사람이 봐도 합성임이 드러난다).
- 운영 코드·설정 경로는 제외 대상이 아니다 — 거기서의 실제 유출 탐지는 유지된다.

## 스코프는 룰 파일이 소유한다

`DefaultExcludes`(`internal/scanner/exclude.go`)의 전역 `--exclude` 는 semgrep 의 **타겟
선정 단계**에서 작동해 룰의 `paths.include` 보다 먼저 파일을 걷어낸다. 그래서 전역에서
제외한 경로는 어떤 룰도 볼 수 없다 — 그 경로를 의도적으로 점검하는 룰이 있어도 마찬가지다.

- 특정 경로를 **의도적으로 점검하는 룰**(예: `finguard.swift.insecure-trust-vendor` 는
  벤더 코드의 인증서 검증 무력화를 본다)이 존재하면, 그 경로를 `DefaultExcludes` 에
  넣으면 안 된다. 제외가 필요한 다른 룰들은 각자 `paths.exclude` 에 적는다 (#75).
- `Pods`·`Carthage` 가 이 경우다. `DefaultExcludes` 에 넣으면
  `TestDefaultExcludesLeavesVendorScopeToRules` 가 실패한다.
- 전역 제외에 새 항목을 추가할 때는 **그 경로를 대상으로 하는 룰이 없는지** 먼저 확인하라.
