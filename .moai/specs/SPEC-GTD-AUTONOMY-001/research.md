---
id: SPEC-GTD-AUTONOMY-001
created: 2026-09-15
updated: 2026-09-15
---

# SPEC-GTD-AUTONOMY-001 조사 기록

## 1. 조사 기준

이 문서는 구현 전 계획 기준을 기록한다. 코드 사실은 develop 계획 보고서가 확인한 HEAD `b45c813493751106cd6619dce60fba1c61e70583`와 그 실행에서 읽은 파일에 귀속한다. 현재 worktree에서는 Xcode license 미동의로 Git 실행이 막혀 HEAD를 독립 재확인하지 못했으므로 구현자는 M1 직전에 다시 측정해야 한다.

## 2. 확인된 현재 계약

원본 계획 보고서가 현재 develop에서 관측한 사실은 다음과 같다.

- todo 저장소는 `~/.moai/db/<project-key>/todo/backlog.db`의 SQLite 큐다.
- live table은 meta/items/findings, archive table은 archived_items/archived_findings다.
- queue state는 queued/picked/dropped이며 done은 보관, undone은 복원이다.
- `internal/cli/todo.go`는 하나의 부모 명령에 하위 명령을 등록한다.
- add `--pick`은 발행과 선택을 한 잠금 변경으로 수행하며 next `--expect`는 카드 본문 prefix를 검사한다.
- `internal/graph`의 repo graph는 codemaps/mx-index/specs/reports fingerprint에서 재생성하는 `edges.jsonl`이다.
- foreman은 picked card만 배차하며 발행·선택 권한을 갖지 않는다.
- 기존 goal은 condition 평가와 반복 진행 기능이며 카드 발행, PR, merge 권한을 뜻하지 않는다.
- `super-advisor`는 비구속 조언자이고 감사 verdict나 구현을 소유하지 않는다.
- command emitter는 file stem 기반으로 사용자 command를 만들며 별도 alias model이 없다.
- repo-local HARD 정책은 카드 worktree를 local `develop`에서 시작하고, lane push와 card-level PR을 금지하며, 단일 local develop integration worktree에서 `--no-ff` merge한 뒤 lead가 `origin/develop`에 batch-push하도록 요구한다. `main`은 검증된 develop에서 분기한 release PR로만 진입한다.
- 현재 todo 명령 표면은 19개 verb(`add/list/done/undone/next/unpick/edit/move/drop/undrop/analyze/relate/unrelate/why/pr/landed/auto-done/export-json/history`)이며 주요 flag는 `add --pick/--force`, `list --json/--dropped/--limit`, `done --expect/--require-landed`, `next --spec/--expect`, `edit --expect`, `move --top/--bottom/--before/--after`, `drop|undrop --expect`, `relate --relation/--note`, `pr --json`, `landed --sha/--ref/--clear`, `auto-done --fetch/--dry-run/--json`, `history --limit`이다.
- 기존 relation vocabulary는 `contains/absorbs/replaces/conflicts`이며 신규 관계로 자동 치환할 수 없다.

## 3. 현재 사실이 아닌 설계 제안

다음은 아직 구현·실측되지 않았다.

- `moai gtd` 정식 표면과 todo 호환 alias
- GTD table, relation assertion, event, policy, decision, mission, operation receipt
- 비공개 `gtd-edges.jsonl`과 logical fingerprint
- 사전 위임 validator와 autonomous publication/selection/dispatch
- `/moai:goal --auto`, `mission_mode=auto`, `mission-governor`
- durable MissionRuntime와 `active-session-only` fallback
- 자율 commit, release integration, batch PR merge와 SHA chain

초안의 카드→release→target 모델은 이 저장소의 local develop 통합 정책과 충돌했다. 본 revision에서는 local develop→launcher-entered WT→single local develop `--no-ff`→lead batch push→origin/develop CI→release branch→main release PR로 대체한다.

따라서 이 항목은 AC를 통과하기 전 완료된 capability로 문서화하면 안 된다.

## 4. GTD 근거의 사용 범위

- 공식 GTD 자료는 Capture → Clarify → Organize → Reflect → Engage라는 정리 절차를 설명한다.
- 공식 Weekly Review checklist는 inbox, next actions, waiting, projects, someday/maybe의 정기 검토 관점을 제공한다.
- 이 자료는 MoAI에서의 생산성 증가, LLM 자율성의 안전, 개인 지식 그래프의 추가 효과를 입증하지 않는다.
- W3C SHACL과 PROV-O는 관계 제약과 provenance를 설계할 때 참고할 수 있으나 본 제품의 저장 schema를 규정하지 않는다.

## 5. 이름 변경 선택지

### 선택 A — 내부까지 전면 rename

DB path, package, table, type을 한 번에 바꾸면 사용자 용어는 단순해지지만 compatibility와 migration blast radius가 가장 크다. 카드 identity와 archive semantics 보존 목표에 맞지 않아 기각한다.

### 선택 B — canonical surface만 GTD로 전환

CLI·slash·skill·docs는 GTD를 canonical로 만들고 todo가 같은 handler를 쓰게 한다. 내부 이름과 물리 path는 유지한다. 첫 구현의 선택안이다.

### 선택 C — GTD를 별도 제품·DB로 추가

두 queue와 두 identity source가 생겨 발행·선택·보관의 단일 원본 계약을 깨므로 기각한다.

## 6. 자율 결정 선택지

### 선택 A — super-advisor에 모든 권한 부여

조언, 결정, 실행, 자기검증이 한 역할에 집중되고 기존 permission contract를 깨므로 기각한다.

### 선택 B — governor와 deterministic executor 분리

governor는 구조화된 선택을 하고 executor가 sealed policy·freshness·evidence·idempotency를 검사한다. auditor와 owning agent를 유지하므로 선택한다.

### 선택 C — 질문 없는 무조건 완료

범위 밖 권한을 추론하거나 안전 gate를 우회해야 하므로 기각한다. 질문 없는 실행은 승인 범위 안에서만 적용하고 그 밖은 blocked다.

## 7. 런타임 미검증 사항

구현 전에 다음을 기계적으로 조사한다.

- 지원되는 GPT session start/reconnect/replace API와 headless contract
- session credential expiration과 runtime process identity
- 다중 manager lease를 위한 기존 persistence helper
- queue transaction과 외부 side effect 사이 crash-cut에 재사용 가능한 receipt pattern
- GitHub PR/CI/review/merge readback의 현재 provider contract
- old binary가 unknown SQLite table/row를 보존하는 실제 왕복 결과

지원 API가 확인되지 않으면 durable supervisor를 추정 구현하지 않고 `active-session-only`로 제한한다.

## 8. 보안 위협

| 위협 | 경계 | 요구 대응 |
|---|---|---|
| 외부 문서 prompt injection | Capture/Clarify | content와 authority 분리, policy 변경 금지 |
| mission shell injection | `--auto` parser | 자연어 전용 schema, shell/condition parser 미사용 |
| stale Decision | governor→executor | snapshot hash, policy version, expiry 검증 |
| 중복 side effect | crash/restart | stable operation ID, receipt, authoritative readback |
| 개인 메모 유출 | graph/template/git | private projection, privacy fail-closed, repo export 제외 |
| 자기 승인 | agent 역할 | advisor/governor/auditor/executor 소유권 분리 |
| stale CI 재사용 | merge | 현재 SHA별 audit/CI/review 검증 |

## 9. 구현 전 검증 질문

아래 항목은 사용자 선택을 다시 요구하는 질문이 아니라 구현자가 현재 코드로 답해야 할 probe다.

- gtd와 todo를 Cobra alias로 안전하게 공유할 수 있는가, 아니면 같은 constructor의 두 command가 필요한가?
- 구버전 SQLite writer가 신규 table과 row를 그대로 보존하는가?
- 기존 graph freshness helper 중 logical revision에 재사용 가능한 부분은 무엇인가?
- 현재 GPT launcher가 durable reconnect를 공식 지원하는가?
- release branch와 batch PR 통합에 이미 존재하는 lease·receipt·readback helper는 무엇인가?

## 10. 참고 자료

- 개발 계획: `.moai/reports/todo-gtd-autonomy/plan.md`
- GTD 다섯 단계: `https://gettingthingsdone.com/what-is-gtd/`
- Weekly Review checklist: `https://gettingthingsdone.com/wp-content/uploads/2014/10/Weekly_Review_Checklist.pdf`
- SHACL: `https://www.w3.org/TR/shacl/`
- PROV-O: `https://www.w3.org/TR/prov-o/`
- repo-local delivery: `.claude/rules/local/repo-local-pr-policy.md`
- verification adoption: `.claude/rules/moai/development/verification-completeness.md §2`
