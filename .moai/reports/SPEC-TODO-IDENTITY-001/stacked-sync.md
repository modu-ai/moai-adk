# Todo runtime store + UUIDv7 identity stacked sync

## Claim

`SPEC-TODO-RUNTIME-STORE-001`과 `SPEC-TODO-IDENTITY-001`의 검증된 child 범위를 하나의 문서 동기화 단위로 닫았다. 공개 Todo JSON 문서는 실행·할당 최신 기록과 프로젝트·카드 UUIDv7 identity를 설명하며, Git 저장소와 Git 없는 일반 폴더에 같은 계약을 적용한다. 기존 identity 미발급 데이터의 순수 조회는 UUID 키를 `null`로 반환하고 영속 상태를 바꾸지 않으며, 승인된 writer만 잠금 아래 같은 SQLite transaction에서 runtime 기록과 identity backfill을 처리한다.

## Evidence

- Runtime 독립 감사: `.moai/reports/t648/runtime-store-sync-audit.md`, PASS 100/100, 4 AC PASS / 0 FAIL, SHA256 `4b84d5a5ced5b29f58f99b71d04d58cb5793a7f7b4ef7b5b41673615e00f3932`.
- Identity 독립 재감사: `.moai/reports/SPEC-TODO-IDENTITY-001/sync-audit-iter2.md`, PASS 98/100, 7 AC PASS / 0 FAIL, SHA256 `77282220ba97fc741bfeaf455669e507bc4c1ceee46e4377ec0e1badff0c655d`.
- B12 pre-emission: 두 SPEC 모두 CHANGELOG 기존 참조 0건. 공식 AC counter는 runtime 4, identity 7을 반환했으며 0 또는 ambiguous가 아니었다. CHANGELOG가 지칭하는 SPEC·감사·문서 경로는 모두 존재한다.
- Phase 12 backup: `.moai/backups/sync-20260912T110913Z/`; 대상 14개 파일을 복사한 뒤 원본과 `cmp` 일치를 확인했다. 백업은 커밋 대상이 아니다.
- Runtime lifecycle: Phase 12 최초 백업의 `spec.md`는 `draft`였고, 그 뒤 manager-develop이 run-phase 선행 보수로 `draft → in-progress`를 수행했으며 manager-docs가 sync-phase에서 `in-progress → completed`로 닫았다. 따라서 backup-to-final 관측 delta는 `draft → completed`지만 직접 `draft → completed` 전이를 뜻하지 않는다.
- 공개 문서: `docs-site/content/{ko,en,ja,zh}/utility-commands/moai-todo.md`의 JSON field set과 설명을 같은 구조로 동기화했다.

## Baseline-attribution

```text
measured_at_utc: 2026-09-12T11:14:40Z
worktree: /Users/goos/MoAI/moai-adk-go/.claude/worktrees/todo-unified
branch: WT-todo-unified
HEAD: a315dad9af0d3a0e04862e6106b3993d9a3812f7
origin/main...HEAD: 0 3005
```

두 child의 `sync_commit_sha`는 자기 자신을 참조할 수 없으므로 `pending-backfill-sync`로 기록했다. 통합 branch commit과 원격 CI 결과는 이 문서 작성 시점에 존재하지 않는다.

## Included scope

- `CHANGELOG.md` `[Unreleased] / Fixed` stacked delivery 항목 1개
- Todo 공개 JSON의 `project_uuid`, 카드별 `card_uuid`, `runtime.runs`, `runtime.assignments`
- legacy `null` 순수 조회와 writer-only transactional backfill 의미
- Git·비Git 프로젝트 동일 identity 계약
- Runtime `spec.md`: manager-develop/run-phase 선행 보수의 `draft → in-progress`, 이어서 manager-docs/sync-phase의 `in-progress → completed`; Phase 12 backup-to-final 관측 delta는 `draft → completed`
- Identity `spec.md`: 기존대로 manager-docs/sync-phase의 `in-progress → completed`
- 두 child `progress.md` §E.4 sync signal

## Excluded scope and Gaps

- `SPEC-TODO-UNIFIED-001`과 카드 t648 완료: 제외; 계속 `picked`·진행 중
- 자동 `picked → done`, completion receipt/recovery, hooks, Graph/UI: 제외
- 운영 Todo DB migration, 설치, 배포, 카드 상태 변경: NOT_RUN
- Git stage/commit/push/merge와 통합 branch CI: PENDING
- README와 `.moai/project/*`: 직접 stale claim이 관측되지 않아 변경하지 않음
- runtime UUID의 물리 column: 도입하지 않았으며 문서도 Todo identity의 논리 projection으로만 설명함
- Identity plan audit: iteration 2 FAIL 0.75와 사용자 승인 BYPASSED 이력을 PASS로 바꾸지 않음

## Residual-risk

고정된 JSON 키 개수나 byte snapshot에 의존하는 외부 소비자는 additive field를 처리하도록 갱신해야 할 수 있다. 운영 데이터는 아직 이전하지 않았고 원격 CI도 실행 전이므로, local child 감사 PASS는 운영 적용이나 통합 branch PASS를 대신하지 않는다. Runtime `reported_state`는 최신 보고일 뿐 카드 완료 권위가 아니며, 자동 완료 신뢰 경계는 후속 receipt/recovery child가 별도로 닫아야 한다.

## Documentation validation

```text
git diff --check
exit=0

moai spec lint .moai/specs/SPEC-TODO-RUNTIME-STORE-001 --strict --json
[]
exit=0

moai spec lint .moai/specs/SPEC-TODO-IDENTITY-001 --strict --json
[]
exit=0

npx --yes prettier --check docs-site/content/{ko,en,ja,zh}/utility-commands/moai-todo.md
All matched files use Prettier code style!
exit=0

hugo --source docs-site --destination <temporary-directory> --gc --minify --panicOnWarning
Pages: KO 187 / EN 185 / JA 185 / ZH 185
HUGO_EXIT=0

four-locale JSON parse and recursive field-path comparison
JSON_OK: 4 files, 32 paths each
FOUR_LOCALE_FIELD_PARITY=PASS
```

`markdownlint-cli2`는 4개 페이지에서 14건을 반환했다. 전부 이번 변경 구간 밖의 기존 명령 표 두 행(각 locale의 현재 226행과 231행)에 있는 MD038/MD056이며, 신규 JSON·identity/runtime 설명 구간에서는 finding이 없었다. 승인 범위 밖의 오래된 표 문장을 고치지 않았고, Prettier와 Hugo의 현재 변경 페이지 검사는 통과했다.
