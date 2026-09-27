# t1226 — SPEC-ALWAYS-LOADED-HEADROOM-001 판정서 (뼈대 — run 단계가 채운다)

레인 백엔드: _미관측_ (출처: _미관측_)
카드: t1226 · Tier M · 클래스 C · 워크트리 `.claude/worktrees/t1226` · 브랜치 `WT-always-loaded-headroom`
기준 커밋: `7fe658815eb0d4110b9acadad56e5a85bee3ed3f` (`git rev-parse HEAD`, 깨끗한 트리 — 오케스트레이터 실측)
기준선(`S_live`): `199111 total` — 150,000 에 대한 잔여 49,111
동결 다중집합 sha256(`S_live`, 170줄, 기준 커밋): `d97b33d960c9801d4ec145ca263ed788425b337f43c585594c8d527c1318c6c3`
폐기 수치(다른 트리의 값, 산술에 쓰지 않음): 197,897(`3a48485af`) · 198,361(`8e50ef148`)

> 이 뼈대에는 `acceptance.md` 가 요구하는 기계 줄(`charmap`·`build_head`·`harness_sha256`·`total_*`·`hash_*`·`current_*`·`A_adm_*`·`R_*`·`U_*`·`T_min_*`·`verdict_*`·`P절 재조정`·`F(172ef22eb)`·`runtime_*`)이 하나도 없다. run 단계가 측정 뒤에 적는다. 레인 백엔드 줄은 run 레인이 자기 서빙 모델을 관측해 채운다.

## 측정 명령 (고정)

```bash
locale charmap
wc -m CLAUDE.md AGENTS.md \
  .moai/config/sections/user.yaml .moai/config/sections/language.yaml \
  .claude/rules/moai/core/agent-common-protocol.md \
  .claude/rules/moai/core/askuser-protocol.md \
  .claude/rules/moai/core/moai-constitution.md \
  .claude/rules/moai/core/moai-mcp-tools.md \
  .claude/rules/moai/core/native-idiom-and-register.md \
  .claude/rules/moai/core/verification-claim-integrity.md \
  .claude/rules/moai/workflow/cache-aware-execution.md \
  .claude/rules/moai/workflow/context-window-management.md \
  .claude/rules/moai/workflow/cross-session-messaging.md \
  .claude/rules/moai/workflow/goal-directive.md \
  .claude/rules/moai/workflow/kanban-dispatch.md \
  .claude/rules/moai/workflow/main-checkout-branch-guard.md \
  .claude/rules/moai/workflow/session-handoff.md \
  .claude/rules/moai/workflow/skill-routing.md | tail -1
```

## S_init

_미측정 — run 단계 M1: 격리 하네스 실행 출력(`harness-run.txt`, 첫 줄 `harness_head`), 프로젝트 트리 전체 내보내기, 18경로 존재 대조, 계수 집합(`count-set-init.txt`), `total_init`, `hash_init`, 격리 런타임 계수 관측 또는 `runtime_observed = no` 와 17집합 줄._

## S_live

_미측정 — run 단계 M1: `total_live`(기준선 199,111 재현 여부), `hash_live`._

## P절 재조정

_미측정 — run 단계 M2: `sec-172ef22eb.txt` 재실행 출력, `J_includes_kanban_scope`, 재조정 선택 줄, 원 트리 서술값 F 줄._

## A_adm 후보 표

_미측정 — `candidates-init.tsv` · `candidates-live.tsv` · `dest-sizes.tsv` · `pointers-*.tsv` (열 규약: `SPEC-ALWAYS-LOADED-HEADROOM-001/acceptance.md §D`). 요약은 run 단계가 채운다._

## T_min Claim

_미측정 — 표면마다 기계 줄을 명령 귀속과 함께 적는다._

## 판정

_미측정 — 표면별 판정 토큰(ACHIEVABLE / STRUCTURALLY-INFEASIBLE-UNDER-FREEZE / UNDETERMINED)._

## 동결 해제 상신 절차

_해당 여부 미정 — `S_init` 판정이 STRUCTURALLY-INFEASIBLE-UNDER-FREEZE 또는 UNDETERMINED 일 때 (a) 해제 후보 줄·묶인 자수, (b) 탐욕 해제 집합, (c) 위험, (d) 결정권자(운영자, 리드가 AskUserQuestion 으로 상신) 네 항목을 채운다. 레인의 결론은 RECOMMEND 접두 문장으로만 쓴다._

## Gaps

_미측정 — run 단계가 채운다. plan 단계 시점의 알려진 Gap: `A_adm` 전체, `P절` 재조정, `S_init` 합계와 계수 집합, 런타임 계수 집합, M2 허용 자수, `R`._

## Residual-risk

_미측정 — run 단계가 채운다._
