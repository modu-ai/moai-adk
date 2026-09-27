---
title: moai contract 자율 실행 계약
weight: 100
draft: false
---

`moai contract` 는 SPEC 의 자율 실행 계약(`.moai/specs/<SPEC-ID>/contract.yaml`)을 검증하고, 내용을 보여 주고, 서명합니다. 계약에는 에이전트가 사람 확인 없이 해도 되는 작업(actions), 써도 되는 경로와 절대 건드리면 안 되는 경로(ownership), 멈추고 사람에게 넘겨야 하는 조건(escalate_on), 예산(budget)이 적히고, 서명은 그 내용과 `acceptance.md` 를 해시로 묶어 둡니다. 서명 뒤에 계약이나 인수 조건이 바뀌면 `verify` 가 바로 불일치를 보고합니다.

{{< callout type="info" >}}
현재 버전에서 서명은 **구현 착수 승인(Implementation Kickoff Approval)을 대체하지 않습니다.** 계약은 기록하고 검증하는 수단이며, 사람 승인 게이트는 그대로 남습니다.
{{< /callout >}}

## 하위 명령어

| 명령어 | 설명 | 종료 코드 |
|--------|------|-----------|
| `moai contract verify <SPEC-ID>` | 계약을 검사합니다. 읽기 전용이라 훅에서 호출해도 안전합니다 | 0 유효 · 1 무효 · 2 사용법/입출력 오류 |
| `moai contract show <SPEC-ID>` | 계약 섹션, 서명 상태, 파생 집합(실효 금지 경로 등)을 출력합니다 | 0 출력 · 2 사용법/입출력 오류 |
| `moai contract sign <SPEC-ID>...` | 계약에 서명합니다 | 0 서명 · 1 거부 · 2 사용법/입출력 오류 |

`verify` 와 `show` 는 `--json` 플래그로 기계가 읽는 JSON 객체를 출력합니다. `verify` 의 무효 사유는 닫힌 코드 집합(`unsigned`, `acceptance_hash_mismatch`, `contract_digest_mismatch` 등)으로만 보고됩니다.

## moai contract sign

```bash
moai contract sign SPEC-AUTH-001
moai contract sign SPEC-AUTH-001 --resign
moai contract sign SPEC-AUTH-001 --signer llm \
  --receipt .moai/specs/SPEC-AUTH-001/kickoff-receipt.json
```

| 플래그 | 설명 |
|--------|------|
| `--signer <human\|llm\|llm+jev>` | 서명 주체. 기본값은 사람(`human`) 경로입니다 |
| `--receipt <path>` | 착수 영수증 경로(프로젝트 루트 기준). `llm` · `llm+jev` 서명에 필요합니다 |
| `--resign` | 이미 서명된 계약을 `acceptance.md` 가 바뀐 뒤 다시 서명합니다. 이전 서명은 `supersedes` 로 남습니다 |

**사람 경로.** 대화형 터미널에서만 동작하며, 서명 요약을 보여 준 뒤 확인 토큰(SPEC ID, 여러 개를 한꺼번에 서명할 때는 `sign N contracts`)을 직접 입력해야 서명합니다. 에이전트 실행 표식 환경변수가 잡히거나 터미널이 아니면 프롬프트를 띄우기 전에 거부합니다. 여러 SPEC 을 한 번에 서명하려면 `workflow.autonomy.contract.batch_sign` 이 켜져 있어야 합니다.

**영수증 경로.** `--signer llm` 또는 `--signer llm+jev` 와 `--receipt` 를 함께 주면 터미널 확인 없이 착수 영수증을 근거로 서명합니다. 이 경로는 `workflow.autonomy.mode` 가 `contract` 일 때만 열리고, 한 번에 SPEC 하나만 서명합니다.

**거부.** 서명이 거부되면 파일은 하나도 바뀌지 않고 `refused <code> (<SPEC-ID>): <사유>` 한 줄이 출력됩니다. 코드는 닫힌 집합(`not_tty`, `confirmation_mismatch`, `already_signed`, `plan_audit_not_passing`, `verify_failed` 등)입니다. 서명할 파일은 쓰기 전에 다시 검증해 유효한 서명이 확인된 경우에만 기록되며, 작성자의 주석과 빈 줄은 보존됩니다.

## 관련 문서

- [config 섹션 레퍼런스 — workflow.yaml autonomy](/ko/advanced/config-sections/#workflowyaml--autonomy)
- [moai spec 문서 관리](/ko/cli-reference/spec)
- [자율성 티어 (MOAI_AUTONOMY_TIER)](/ko/advanced/autonomy-tier) — 이름은 비슷하지만 서로 무관한 설정입니다
