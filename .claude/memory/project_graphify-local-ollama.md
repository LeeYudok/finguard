---
name: graphify-local-ollama
description: graphify 를 로컬 Ollama 로 돌릴 때의 함정 — [ollama] extra 필요, 리즈닝 모델(qwen3.6)은 hollow 응답으로 실패, vendor/testdata 는 .graphifyignore 로 제외
metadata:
  type: project
---

graphify(`uv tool install graphifyy`) 로컬 실측 (2026-08-25, #84):

- Ollama 백엔드는 `uv tool install "graphifyy[ollama]" --force` 로 **openai 패키지 extra** 를 깔아야 동작한다. 기본 설치만 하면 "the 'openai' package is required" 로 모든 청크 실패.
- 맥북 로컬 `qwen3.6:35b`(리즈닝 모델)는 청크마다 **hollow 응답**(노드 0개, output_tokens 2500+)을 내고 무한 재시도한다 — 사고 토큰만 뱉고 JSON 을 안 낸다. 문서 의미 추출은 비리즈닝 모델(pig Ollama 의 `qwen2.5-coder:32b` 등, `--backend ollama --model ...`)로.
- 코드만이면 `--code-only` 가 LLM 없이 즉시 끝난다. `.graphifyignore` 없이 돌리면 `vendor/`·`testdata/` 가 그래프의 80% 를 차지해 질의가 yaml.v3 내부로 새므로 반드시 제외.

**Why:** 장기 기억 도구를 graphify 하나로 통일했는데(AGENTS.md), 첫 빌드에서 위 세 가지에 연달아 막혀 30분 소모.
**How to apply:** 새 머신/새 레포에서 graphify 세팅 시 extra 설치 → `.graphifyignore` → `--code-only` 순으로. 문서 추출은 모델 먼저 확인.
