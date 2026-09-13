# Todo runtime-store + identity post-document sync audit

## 1. Input Scope Reviewed

**Overall Verdict: FAIL — 85/100**

검토한 승인 문서 delta는 정확히 다음 10개 경로다.

1. `CHANGELOG.md`
2. `docs-site/content/ko/utility-commands/moai-todo.md`
3. `docs-site/content/en/utility-commands/moai-todo.md`
4. `docs-site/content/ja/utility-commands/moai-todo.md`
5. `docs-site/content/zh/utility-commands/moai-todo.md`
6. `.moai/specs/SPEC-TODO-RUNTIME-STORE-001/spec.md`
7. `.moai/specs/SPEC-TODO-RUNTIME-STORE-001/progress.md`
8. `.moai/specs/SPEC-TODO-IDENTITY-001/spec.md`
9. `.moai/specs/SPEC-TODO-IDENTITY-001/progress.md`
10. `.moai/reports/SPEC-TODO-IDENTITY-001/stacked-sync.md`

`.moai/backups/sync-20260912T110913Z/`의 승인 대상 14개와 현재 파일을 `cmp`한 결과 9개가 변경됐고 `stacked-sync.md` 1개가 새로 추가됐다. `README.md`, 두 child의 `plan.md`, 두 child의 `acceptance.md`는 backup과 동일했다.

근거 문서:

- Runtime 독립 감사 `.moai/reports/t648/runtime-store-sync-audit.md`, SHA-256 `4b84d5a5ced5b29f58f99b71d04d58cb5793a7f7b4ef7b5b41673615e00f3932`.
- Identity 독립 재감사 `.moai/reports/SPEC-TODO-IDENTITY-001/sync-audit-iter2.md`, SHA-256 `77282220ba97fc741bfeaf455669e507bc4c1ceee46e4377ec0e1badff0c655d`.
- Sync 전 backup `.moai/backups/sync-20260912T110913Z/`, timestamp 식별자 `20260912T110913Z`.
- 현재 측정 기준은 `WT-todo-unified` / HEAD `a315dad9af0d3a0e04862e6106b3993d9a3812f7` / `origin/main...HEAD = 0 3005`다.

주관적 표현은 별도로 판정할 항목이 없었다. 공개 문서의 기능·호환성·완료범위 주장을 원자 단위로 분리해 아래에서 판정했다.

## 2. Claim Assessments

| # | Atomic claim | Label | Reason | Evidence anchor or what is missing |
|---:|---|---|---|---|
| 1 | runtime run/assignment 최신 기록은 Todo SQLite DB에 보존된다. | verified | — | Runtime 감사 PASS 100/100 및 hash 일치. |
| 2 | 별도 Factory 저장소에 신규 runtime 기록을 쓰지 않는다. | verified | — | Runtime 감사 AC-TRS-001 및 stacked evidence anchor. |
| 3 | Git/비Git 프로젝트 모두 canonical lowercase UUIDv7 identity를 발급한다. | verified | — | Identity 재감사 PASS 98/100 및 hash 일치. |
| 4 | legacy pure read는 UUID 키를 생략하지 않고 `null`로 반환하며 영속 상태를 바꾸지 않는다. | verified | — | Identity 재감사 AC-TID-002, 4-locale JSON null contract 검사. |
| 5 | 승인된 writer만 lock 아래 같은 transaction에서 runtime/identity를 기록한다. | verified | — | 두 독립 감사의 transaction/rollback 근거와 stacked evidence anchor. |
| 6 | public JSON의 `project_uuid`, `card_uuid`, `runtime.runs`, `runtime.assignments` 구조가 4개 locale에서 같다. | verified | — | 4개 JSON parse 성공, object-key path 32개씩 동일, null contract PASS. |
| 7 | Runtime child는 독립 감사 PASS 100/100, 4 AC PASS다. | verified | — | runtime audit hash 및 문서 존재 검사 일치. |
| 8 | Identity child는 독립 재감사 PASS 98/100, 7 AC PASS다. | verified | — | identity iter2 audit hash 및 문서 존재 검사 일치. |
| 9 | Identity plan audit FAIL 0.75와 BYPASSED 이력은 유지된다. | verified | — | identity spec/progress/stacked/CHANGELOG가 모두 FAIL/BYPASSED를 보존함. |
| 10 | 이 delivery는 t648, 자동 picked→done, receipt/recovery, hooks, Graph/UI, 운영 migration을 완료하지 않았다. | verified | — | CHANGELOG와 두 progress 및 stacked excluded scope가 모두 명시적으로 PENDING/NOT_RUN/false를 기록함. |
| 11 | 두 child spec은 모두 `in-progress → completed`로 전이했다. | unsupported | contradicted | Runtime backup `spec.md:5`는 `status: draft`, 현재 `spec.md:5`는 `completed`; runtime `progress.md:54`와 `stacked-sync.md:33`은 `in-progress_to_completed`라고 기록함. Identity는 실제 `in-progress → completed`. |
| 12 | 통합 branch CI와 git commit은 아직 완료 전이다. | verified | — | 두 progress §E.4와 stacked baseline이 `PENDING_INTEGRATION_BRANCH_CI` / `PENDING_BACKFILL`을 기록함. |
| 13 | 승인 문서 경로 10개 밖의 backup 대상 문서는 바뀌지 않았다. | verified | — | backup/current `cmp`: changed 9, unchanged 5, new stacked 1, delta paths 10. |
| 14 | 공개 문서의 local link target이 존재한다. | verified | — | Hugo-route-aware local link 검사 289개 PASS; Hugo panic-on-warning build PASS. |
| 15 | 변경 문서에는 credential-shaped secret가 없다. | verified | — | 제한된 secret pattern scan `SECRET_PATTERN_MATCHES=0`. |

### Finding

- **F1 [High] [blocking] [confidence: 1.00]** `.moai/specs/SPEC-TODO-RUNTIME-STORE-001/progress.md:54`, `.moai/reports/SPEC-TODO-IDENTITY-001/stacked-sync.md:33`, 비교 기준 `.moai/backups/sync-20260912T110913Z/.moai/specs/SPEC-TODO-RUNTIME-STORE-001/spec.md:5` — Runtime child의 실제 sync 전 상태는 `draft`이고 현재는 `completed`인데 두 문서는 `in-progress → completed`라고 기록한다. 영향: lifecycle provenance와 frontmatter transition audit trail이 실제 delta와 다르며, 요청상 새 문서 finding은 차단이다. Required fix: 실제 전이를 `draft → completed`로 정직하게 기록하거나, workflow가 반드시 `in-progress` 중간 상태를 요구한다면 sync close를 되돌린 뒤 권한 있는 phase owner가 누락된 transition을 올바른 순서와 근거로 수행하고 progress/stacked 기록을 그 실제 이력에 맞춰 갱신하라.

Lifecycle 불일치 외 추가 blocking finding은 발견하지 않았다.

## 3. Boundary Notes

### Claim

문서 렌더링·구조·JSON·링크·secret·경로 경계는 PASS했지만, Runtime lifecycle에 새 차단 불일치 F1이 있으므로 Functionality must-pass가 실패하고 전체 판정은 FAIL이다. 기존 markdownlint 14건은 동일 문구/규칙이 sync backup에도 존재하는 baseline gap으로 귀속했으며 새 finding으로 세지 않았다.

### Dimension Scores

| Dimension | Score | Verdict | Evidence |
|---|---:|---|---|
| Functionality (40%) | 75/100 | **FAIL (must-pass)** | Runtime 실제 `draft → completed`와 문서 주장 `in-progress → completed`가 모순됨. |
| Security (25%) | 100/100 | PASS | secret pattern 0, local link targets PASS, 외부/운영 완료를 허위 주장하지 않음. |
| Craft (20%) | 95/100 | PASS | strict SPEC lint, Prettier, Hugo, JSON parse/parity 모두 PASS. markdownlint 14건은 backup에 동일하게 존재. |
| Consistency (15%) | 75/100 | FAIL | Runtime progress와 stacked-sync의 lifecycle 값이 backup/current 사실과 불일치. Identity lifecycle과 나머지 cross-doc claim은 일치. |

가중치: `75×0.40 + 100×0.25 + 95×0.20 + 75×0.15 = 85.25`, 표시 점수 85/100. Functionality must-pass 실패가 전체 FAIL을 강제한다.

### Evidence

```text
$ moai spec lint .moai/specs/SPEC-TODO-RUNTIME-STORE-001 --strict --json
[]
$ moai spec lint .moai/specs/SPEC-TODO-IDENTITY-001 --strict --json
[]
exit=0

$ npx --yes prettier --check docs-site/content/{ko,en,ja,zh}/utility-commands/moai-todo.md
Checking formatting...
All matched files use Prettier code style!
exit=0

$ hugo --source docs-site --destination <temporary-directory> --gc --minify --panicOnWarning
Pages        │ 187 │ 185 │ 185 │ 185
Total in 5039 ms
exit=0

JSON_OK ko paths=32 null_contract=PASS
JSON_OK en paths=32 null_contract=PASS
JSON_OK ja paths=32 null_contract=PASS
JSON_OK zh paths=32 null_contract=PASS
FOUR_LOCALE_FIELD_PARITY=PASS

LOCAL_LINKS_CHECKED=289
LOCAL_LINK_TARGETS=PASS
SECRET_PATTERN_MATCHES=0

BACKUP_CHANGED=9
BACKUP_UNCHANGED=5
APPROVED_DELTA_PATHS=10
```

Lifecycle counterevidence:

```text
$ nl -ba .moai/backups/sync-20260912T110913Z/.moai/specs/SPEC-TODO-RUNTIME-STORE-001/spec.md | sed -n '1,10p'
5  status: draft

$ nl -ba .moai/specs/SPEC-TODO-RUNTIME-STORE-001/spec.md | sed -n '1,10p'
5  status: completed

.moai/specs/SPEC-TODO-RUNTIME-STORE-001/progress.md:54:  spec_md: in-progress_to_completed
.moai/reports/SPEC-TODO-IDENTITY-001/stacked-sync.md:33:- 두 child `spec.md`의 `in-progress → completed` 전이
```

Markdownlint attribution:

```text
CURRENT: Summary: 14 issues in 4 files
rules: MD038 / MD056, current lines 226 and 231

BASELINE: the same 14 MD038 / MD056 contexts exist at backup lines 200 and 205
```

현재 14건의 exact contexts(`The same`, 각 locale의 `없으면`/동등 문구, `[DROPPED — ...] `, table pipe parsing)은 backup에도 동일하다. Prettier로 주변 표 간격과 line 위치가 바뀌었지만 신규 identity/runtime JSON 및 설명 문구에서 발생한 finding은 아니다.

첫 local-link probe는 `/ko/...` 같은 Hugo route를 raw filesystem path로 취급해 16개를 잘못 실패시킨 setup 오류였다. `.md`/`_index.md` route mapping을 적용한 교정 검사에서 289개 local target이 모두 PASS했고 Hugo도 warning 없이 exit 0이었다. 최초 probe는 semantic FAIL 근거로 사용하지 않았다.

### Baseline-attribution

```text
worktree: /Users/goos/MoAI/moai-adk-go/.claude/worktrees/todo-unified
branch: WT-todo-unified
HEAD: a315dad9af0d3a0e04862e6106b3993d9a3812f7
origin/main...HEAD: 0 3005
sync backup: .moai/backups/sync-20260912T110913Z/
runtime audit SHA256: 4b84d5a5ced5b29f58f99b71d04d58cb5793a7f7b4ef7b5b41673615e00f3932
identity audit SHA256: 77282220ba97fc741bfeaf455669e507bc4c1ceee46e4377ec0e1badff0c655d
```

승인 경계는 backup에 포함된 14개 파일과 새 stacked report를 비교해 측정했다. production/test의 기존 dirty 변경은 이번 문서 delta의 변경으로 귀속하지 않았고 건드리지 않았다.

### Gaps

- `moai todo list --json`은 운영 Todo DB/card mutation 금지 경계를 지키기 위해 호출하지 않았다. 따라서 t648의 현재 운영 state는 이번 감사에서 독립 재조회하지 않았고, 문서/dispatch의 `picked_in_progress` 근거만 확인했다.
- 외부 HTTP 링크의 live reachability는 검사하지 않았다. local link 289개와 Hugo route/build만 검증했다.
- markdownlint 전체 baseline은 273건이고 current는 14건이다. 이번 판정은 current 14개의 동일 rule/context가 backup에도 존재함을 확인한 attribution이며, 기존 273건 전체를 품질 PASS로 바꾸지 않는다.
- `moai-workflow-docs-claim-check`의 claim triage를 공개 CHANGELOG/Todo 문서에 적용했지만, 이번 사용자 요청이 별도의 기계 검증을 명시했기 때문에 Prettier/Hugo/lint 등의 명령은 일반 sync-auditor 검증 단계에서 실행했다. 따라서 그 스킬의 문자 그대로인 “no commands executed” certification은 하지 않는다.
- Git stage/commit/push/merge, 운영 migration, DB/card 변경은 NOT RUN이다.

### Residual-risk

- F1을 문구만 실제 `draft → completed`로 고칠지, 누락된 lifecycle phase를 권한 있는 owner가 재수행할지는 workflow owner가 결정해야 한다. 감사자는 어느 쪽도 수정하지 않았다.
- Hugo 성공은 외부 URL의 현재 응답이나 배포된 사이트를 증명하지 않는다.
- 4-locale parity는 JSON field path와 null contract를 검사한다. 번역의 모든 뉘앙스가 동일하다는 의미는 아니다.
- child local audit PASS는 t648/Graph/hooks/receipt/recovery/운영 migration/원격 CI 완료를 의미하지 않는다. 현재 문서는 이 경계를 올바르게 명시한다.

### Final verdict

**FAIL — 85/100.** 새 차단 finding은 F1 하나다. Runtime child의 실제 lifecycle은 backup 기준 `draft → completed`인데 `progress.md`와 `stacked-sync.md`가 `in-progress → completed`로 기록했다. 그 외 strict lint, Prettier, Hugo panic-on-warning, 4-locale JSON parse/path/null parity, local links, secret scan, audit hash 및 승인 changed-path 경계는 PASS했다. 기존 MD038/MD056 14건은 backup-attributable gap이다.
