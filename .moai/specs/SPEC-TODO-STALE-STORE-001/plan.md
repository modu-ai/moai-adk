# plan.md — SPEC-TODO-STALE-STORE-001 (카드 t1307)

## §A Context

정식 큐는 홈 DB `~/.moai/db/<project-key>/todo/backlog.db` 하나다
(`internal/kanban/state_dir.go` `StateDirForRoot`, `internal/homestate/paths.go:222`
`BacklogDBPath`). primary 체크아웃에는 롤백 스냅샷인 프로젝트-로컬 레거시 스토어가
`.moai/state/todo/`(`projectStateDirForRoot`)와 `.moai/state/kanban/`
(`LegacyStateDirForRoot`)에 남을 수 있고, `moai todo`는 이제 어느 쪽도 읽지 않는다.
그런데 읽기 표면의 유일한 고지(`internal/cli/todo_disclosure.go`)는 State D
(`backlog.json` 병존)만 다루므로 스테일 SQLite 스토어는 완전히 침묵한다. 2026-09-29
리드가 `queued=72`(seq 661)를 살아 있는 큐로 오독한 사고가 이 침묵의 피해다.

## §B Known Issues

- 유령 스토어 `.moai/state/todo/backlog.db` — seq 661, "queued" 72행, 홈 DB(last_seq
  1305)와 644 seq 격차. 고지·마커·발산 검출 모두 없음.
- 기존 stderr 고지는 `kanban.InspectBacklogArchiveVouch`
  (`internal/kanban/backlog_archive_vouch.go:55`)의 `NonAuthoritativeJSON` 한 가지 사실만
  전한다.
- 0바이트 쌍 `.moai/backlog.db`, `.moai/state/backlog.db` — 카드 t472가 2026-09-03부터
  플래그, 처분 안 됨.
- 마이그레이션 백업 `.moai/state/todo-merge-backup-20260913/store-{1,2}/` — 보존 필요성
  미측정.
- develop 브랜치 저장소 문서는 "홈 DB 백업/검증까지 유지"라고 하나 격리 마커도 스테일
  고지도 발산 검출도 없다.

## §C Pre-flight (조사 결과 — 구현이 인용할 파일)

| 대상 | 위치 |
|---|---|
| 홈 DB 경로 해석 | `internal/kanban/state_dir.go` `StateDirForRoot`; `internal/homestate/paths.go` `ProjectKey`/`BacklogDBPath` |
| 레거시 로컬 스토어 경로 | `internal/kanban/state_dir.go:57-63` — `projectStateDirForRoot`(`.moai/state/todo`), `LegacyStateDirForRoot`(`.moai/state/kanban`) |
| `meta.last_seq` 판독 | `internal/kanban/backlog_sqlite.go:65` `backlogMetaKeyLastSeq`; `internal/kanban/backlog_migrate.go:305` `readLastSeq`(비공개 — 검출기는 별도 읽기전용 판독이 필요) |
| 기존 고지 표면 | `internal/cli/todo_disclosure.go` — `discloseNonAuthoritativeBacklogJSON` + `discloseQueueLayout`; 호출부 `internal/cli/todo.go:573`, `todo_pr.go:179`, `todo_why.go:28`, `todo_history.go:157` |
| vouch 팩트 구조 | `internal/kanban/backlog_archive_vouch.go` `BacklogArchiveVouch`/`InspectBacklogArchiveVouch` |
| doctor 점검 등록 | `internal/cli/doctor.go:187` `runGroupedChecksObserved` — `{상수, func(v bool) DiagnosticCheck}` 행 추가. 선례 상수: `hookMissingLogCheckName`(t1251), `servedModelCheckName`(t1282), `factoryRunCheckName`(t1256) |
| binary_lag 쌍 | `internal/cli/binary_lag_test.go` `namesAddedAfterBaseline` 맵 + `TestBinaryLag_AllowlistKeysAreLiveNames`(키 모양 검증 — 상수 등록이면 bare 식별자 키) |
| doctor 골든 | `internal/cli/doctor_golden_test.go` + `testdata/*.golden` — `UPDATE_GOLDEN=1` 재생성 |
| 읽기-경로 순수성 선례 | `internal/kanban/todo_root.go`(ResolveTodoQueueRoot PURE), REQ-WTQ-001 |

## §D Constraints

- 개발 모드: DDD(quality.yaml). 읽기 경로에 쓰기 금지 — 검출기는 read-only로만.
- 테스트는 `t.TempDir()` + 기존 시임(`kanban.HomeDirFn`, `paths.EnvHome` override)으로
  홈 DB 격리. 실제 `~/.moai/db/...`와 primary 체크아웃 `.moai/` 건드리기 금지.
- `moai doctor` 신규 점검은 binary_lag 쌍이 HARD — 쌍 없는 착지는 결함.
- 사용자 대상 stderr 문안 영어, 기존 고지 문안 어조("NOT the queue")와 일관.

## §E Self-Verification

M1..M3 각 종료 시 본인 변경 패키지 테스트 + 아래 AC 재측정. 전체 스위트는 CI 몫
(CLAUDE.local.md §4 — 로컬 `go test ./...` 금지).

## §F Milestones (결정 변동 가능성 순 — 설계 결정이 위, 기계 작업이 아래)

### M1 (High) — 스테일 로컬 스토어 검출기 + stderr 고지 (항목 ①)

기술 접근:

1. `internal/kanban`에 **하나의 읽기전용 검출기**를 둔다(REQ-TSS-004). 입력은 큐 루트,
   출력은 팩트 구조: 레거시 스토어 경로 유무, 양쪽 `meta.last_seq`, 발산 여부.
   `meta.last_seq` 판독은 기존 `backlog_sqlite.go`의 meta 키 상수를 재사용하는 별도
   read-only 오픈(마이그레이션·lock·DDL 없음 — REQ-TSS-013 성격 공유). 스토어 부재·
   비SQLite 파일·0바이트 파일은 "판독 불가"로 발산과 구분해 보고한다.
2. `internal/cli/todo_disclosure.go`의 고지 경로가 이 팩트를 추가로 전한다 — 기존
   `discloseQueueLayout` 진입점에 한 줄 확장(verb 접두어, stderr 전용, stdout 무변경:
   REQ-TSS-001/002). 기존 backlog.json 고지 문안과 별개 줄로, 서로 덮지 않는다.
3. 검출기에 레거시 디렉터리 둘(`todo`/`kanban`)을 모두 보게 한다.

테스트: `internal/kanban`(검출기 단위 — 발산/일치/부재/0바이트) + `internal/cli`(고지
stderr 문안, stdout 바이트 동일성, 스토어 mtime·해시 불변 = REQ-TSS-003).

AC: AC-TSS-001, AC-TSS-002, AC-TSS-003, AC-TSS-004.

### M2 (High) — doctor 발산 점검 + binary_lag 쌍 (항목 ②)

기술 접근:

1. `internal/cli/doctor_todo_store.go`에 점검을 두고 이름은 **상수**로 등록한다
   (`runGroupedChecksObserved` 행 추가 — §C 표의 선례 상수 3건과 같은 모양).
2. 판정 3상태: 발산 → 실패(양쪽 last_seq를 메시지에), 홈 DB 부재 → 실패, 레거시 스토어
   부재 → PASS. M1 검출기 재사용(REQ-TSS-004), 읽기 전용(REQ-TSS-013).
3. **binary_lag 쌍 — 같은 커밋에**: `internal/cli/binary_lag_test.go`
   `namesAddedAfterBaseline`에 상수 식별자를 bare 키로 추가하고
   `TestBinaryLag_AllowlistKeysAreLiveNames`를 녹인다(키 모양 오류 방지 — 상수 등록은
   bare, 문자열 리터럴 등록만 따옴표 포함).
4. doctor 골든 스냅샷이 새 점검 이름을 포함하면 `UPDATE_GOLDEN=1 go test`로 재생성해
   같은 커밋에 넣는다(REQ-TSS-012).

테스트: 점검 단위(임시 트리에서 3상태 table-driven) + `go test ./internal/cli/ -run
TestBinaryLag` + 골든 재생성 후 `go test ./internal/cli/ -run '^TestDoctorGolden$'`.

AC: AC-TSS-010, AC-TSS-011, AC-TSS-012, AC-TSS-013, AC-TSS-014.

### M3 (Medium) — 잔존 저장소 처분: 측정 먼저, 확인 게이트, 증거 (항목 ③)

절차(운영 단계 — 코드 변경 없음, primary 체크아웃에서 수행):

1. **측정**(REQ-TSS-020): 대상 4개 — `.moai/backlog.db`, `.moai/state/backlog.db`
   (0바이트 확인: `ls -la`/`stat`), `.moai/state/todo-merge-backup-20260913/store-{1,2}/`
   (크기, 내부 스키마, 최종 수정 시각). 각각에 대해 "홈 DB 백업/검증이 이 스냅샷을
   필요로 하는가"를 판정한다(참조 검색: 저장소 문서·코드 경로 grep).
2. **확인 게이트**(REQ-TSS-021): 삭제 후보 목록과 측정 결과를 리드에게 보고하고
   **운영자/리드의 명시적 확인을 전제조건으로 받는다**. 확인이 없으면 여기서 멈추고
   측정 기록만 남긴다. 자동 삭제 경로를 만들지 않는다.
3. **처분 + 증거**(REQ-TSS-022): 확인된 대상만 삭제(또는 `.moai/reports/t1307/` 아래
  보존 이동)하고 — 경로, 삭제 전 sha256/크기, 확인 주체·시점 — 을 SPEC progress 기록
  §E.2에 남긴다.

[NEEDS CLARIFICATION: 0바이트 쌍과 백업 쌍의 삭제 승인 주체가 운영자 직답인지 리드 대행인지 — 실행 시점에 확인 게이트에서 확정]

AC: AC-TSS-020, AC-TSS-021, AC-TSS-022.

## §G Anti-Patterns

- 검출을 위해 유령 스토어를 열 때 migration/lock 경로(`openEngine`)를 재사용하는 것 —
  읽기 경로가 쓰게 된다(REQ-TSS-003 위반). 반드시 read-only 오픈.
- doctor 점검을 문자열 리터럴로 등록하는 것 — allowlist 키 모양이 갈라진다(t1251 교훈).
- stdout에 고지를 섞는 것 — foreman 루프 기계 표면 오염(REQ-BJD-004).
- 삭제를 "측정과 동시에" 진행하는 것 — REQ-TSS-021 위반, 게이트 없는 파괴.
- 테스트가 실제 홈 DB를 보는 것 — `t.TempDir()` + 시임 주입 필수.

## §H Cross-References

- spec.md §B 요구사항, acceptance.md AC 전수, `.moai/reports/todo-logic-review-20260929.md`(primary 체크아웃, 읽기전용 참고)
- `.claude/rules/moai/core/verification-claim-integrity.md` — 처분 증거 기록의 근거 규율
