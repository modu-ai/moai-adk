# Sync-audit 판정 — SPEC-SYNC-GATE-FAILSTATE-001 (card t624, Tier M, lens `--security`)

- 판정: **PASS-WITH-DEBT**
- 점수: 조화평균 **85.6** / 100 (참고: 가중 평균 86.2)
- 필수 통과(Functionality, Security): **둘 다 통과** — Critical/High 발견 0건, AC 15/15 재측정 통과
- 발견: 차단(blocking) 0건 / 선택(optional) 9건 (Low 7, Info 2)
- 측정 트리: `.claude/worktrees/t624`, 브랜치 `WT-sync-gate-failstate`, HEAD `9af4acb92`, 감사 시작 시 `git status --short` 빈 출력
- 감사자는 판정만 한다. 수정은 하지 않았고 추적 파일도 건드리지 않았다(mutant 재적용 없음, 커밋 없음).

---

## 1. 주장 (Claim)

| 차원 | 점수 | 판정 | 요지 |
|---|---|---|---|
| Functionality (40%) | 90 | PASS | 셀렉터 13개 상위 테스트 전부 통과, AC013 하위 D1/S1/S2/S3 존재, 템플릿 패키지 통과, 훅 사본 바이트 동일, AC-010 앵커 해시 성립, mutant 증거 표본 일치. CHANGELOG에 경미한 귀속 부정확 1건(F7) |
| Security (25%) | 80 | PASS | 새로 생긴 표면(저장된 stdout 재방출, 상태 파일 쓰기)에서 Low 4건 관측. 모두 `.moai/state`(gitignore 대상)에 대한 로컬 쓰기 권한을 전제로 하며, 그 권한이면 훅 스크립트 자체를 고칠 수 있으므로 High 이상이 아니다 |
| Craft (20%) | 85 | PASS | gofmt/vet/golangci-lint 0건, mutant 27/27 적색 기록, 부분 쓰기 경로(S2/S3) 보강. 심볼릭 링크·디렉터리·미래 mtime·로그 쓰기 불가 행은 테스트에 없음 |
| Consistency (15%) | 88 | PASS | 훅 사본·문서 앵커·catalog 해시·sync 커밋 범위(3파일, `status`+`updated`만) 모두 규약과 일치. stderr 누수 1건(F6), 타 문서의 낡은 서술은 t644 소관 |

## 2. 증거 (Evidence)

### 2.1 트리 확인

```
$ git rev-parse --show-toplevel && git rev-parse --short HEAD && git branch --show-current && git status --short
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t624
9af4acb92
WT-sync-gate-failstate
(빈 출력)
```

`git log --oneline 3f5dc3f8c..HEAD`: 13개 커밋, 첫 run 커밋 `1a3b12ce5`, 마지막 `9af4acb92`. 모든 커밋 제목에 `t624` 또는 SPEC ID 포함.

### 2.2 기능 재측정

셀렉터 — 전문: `.moai/reports/t624/sync-audit-selector.txt`

```
$ unset MOAI_SYNC_GATE_BLOCKING MOAI_AUTONOMY_TIER && go test ./internal/hook/ -run '^TestSyncGateFailState' -count=1 -v > .moai/reports/t624/sync-audit-selector.txt 2>&1
exit=0
상위 테스트 RUN 수: 13
'no tests to run' 수: 0
--- PASS: TestSyncGateFailState_AC013_RetryByDeletionNoStaleAuxState (6.31s)
    --- PASS: .../D1 (1.08s)   --- PASS: .../S1 (1.59s)
    --- PASS: .../S2 (1.78s)   --- PASS: .../S3 (1.87s)
...
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook	100.287s
```

집계한 상위 13개: AC001, AC002, AC003, AC004, AC005_UnknownAndLegacyRecordsRegate, AC005_TornWriteNeverSilentPass, AC006_RunningRecordStaleWindow, AC007, AC008, AC013, AC014, AC006c, AC015 — 전부 `--- PASS`, `--- FAIL` 0건.

템플릿 패키지 — 전문: `.moai/reports/t624/sync-audit-template.txt`

```
$ unset MOAI_SYNC_GATE_BLOCKING MOAI_AUTONOMY_TIER && go test ./internal/template/ -count=1 > .moai/reports/t624/sync-audit-template.txt 2>&1
exit=0
ok  	github.com/modu-ai/moai-adk/internal/template	35.887s
```

훅 사본 동일성:

```
$ cmp internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh .claude/hooks/moai/sync-phase-quality-gate.sh && echo CMP-IDENTICAL
CMP-IDENTICAL
```

AC-010 앵커(두 문서 사본; `fa96fe644` 판은 `git show`로 scratch에 반출 후 비교). macOS `sed`가 `--`를 파일명으로 읽어 `sed: --: No such file or directory`를 stderr에 찍었으나 실제 파일은 처리됐다 — 해시는 실제 파일 내용이다.

```
앵커 줄까지(template, HEAD)  3c5846648cc8d0dd7475ce86d2cb923e503470a4eca244d66d6f6e6c187dce9c
앵커 줄까지(local,    HEAD)  3c5846648cc8d0dd7475ce86d2cb923e503470a4eca244d66d6f6e6c187dce9c   → 두 사본 동일
앵커 이후(template, HEAD)    27b3f818e518ae746ea82456191514e697ba05b39d0cac32bab81c266575b01d
앵커 이후(template, fa96fe644) 27b3f818e518ae746ea82456191514e697ba05b39d0cac32bab81c266575b01d   → 불변
앵커 이후(local,    HEAD)    38ac05381d74cb699280a21fbb3c61e1ae671abd9d324861a7e9b0d034549428
앵커 이후(local,    fa96fe644) 38ac05381d74cb699280a21fbb3c61e1ae671abd9d324861a7e9b0d034549428   → 불변
```

앵커 줄 번호: 템플릿 160행. 두 사본의 diff 헝크는 113행·137행 부근뿐이며 모두 앵커 위다.

mutant 증거 표본(커밋된 파일을 읽었고 재적용하지 않음):

| 파일 | 대상 행 | 관측 |
|---|---|---|
| `m4-mutant-M06.txt` | AC-006c | `run 3 invoked the stub 2 time(s); want 0`, `--- FAIL: TestSyncGateFailState_AC006c_RetryBoundAndNotice` |
| `m4-mutant-M18.txt` | AC-013 S1(1차) | D1·S1 `--- PASS`, `ok` — 기록대로 생존 |
| `m4-mutant-M18-rerun.txt` | AC-013 S2·S3 | S2 `2 of 2 stub invocation(s) during call 2 saw the payload file present`, S3 `a payload file exists after call 2; want none` / `call 3 stub count unchanged (delta 0)` — 적색. 훅 372행 `rm -f "$PAYLOAD_FILE"`와 대응 |
| `m4-mutant-M20.txt` | AC-006 b70 | b61·b70 `stub invoked 0 time(s)`, notice 출력 — 적색. 355행 비교식과 대응 |
| `m4-mutant-M24.txt` | AC-005 TB2 | `stub invoked 0 time(s) on this call`, `stdout=""` — 적색 |

### 2.3 품질 도구

```
$ gofmt -l internal/hook/sync_gate_failstate_test.go internal/hook/sync_gate_failstate_unix_test.go; echo "gofmt_rc=$?"
gofmt_rc=0
$ go vet ./internal/hook/ > .moai/reports/t624/sync-audit-vet.txt 2>&1; echo "vet_rc=$?"; wc -c < .moai/reports/t624/sync-audit-vet.txt
vet_rc=0
       0
$ golangci-lint run ./internal/hook/... > .moai/reports/t624/sync-audit-lint.txt 2>&1; echo "lint_rc=$?"
lint_rc=0
0 issues.
$ command -v shellcheck || echo "shellcheck: absent"
shellcheck: absent
```

### 2.4 CHANGELOG 사실 대조

| CHANGELOG 주장 | 대조 | 결과 |
|---|---|---|
| 옛 훅은 checks 전에 HEAD SHA만 `.last`에 기록하고 같은 HEAD에서 조용히 통과 | 기준 훅 diff: `-# ... The SHA is recorded BEFORE the checks run` / `-if [ -n "$HEAD_SHA" ] && [ -f "$SENTINEL_FILE" ] && [ "$(cat ...)" = "$HEAD_SHA" ]; then exit 0` | 일치 |
| `<sha> running\|pass\|fail` 기록 + payload + retry 표식 | 훅 199-202, 310-379, 563-570행 | 일치 |
| stop_hook_active면 유예(버리지 않음), advisory 해석에서는 재전달 안 함 | 343-346행, 339-342행; AC004·AC008·S1 통과 | 일치 |
| running 60초 이하 → notice, 초과 → HEAD당 1회 재실행, 이후 notice | 353-366행, `SYNC_GATE_STALE_WINDOW=60`; AC006·AC006c·AC014 통과 | 일치 |
| 새 플래그·환경변수 없음 | 훅이 읽는 `MOAI_*`는 `MOAI_SYNC_GATE_BLOCKING`, `MOAI_AUTONOMY_TIER`뿐 | 일치 |
| 문서와 **훅 주석**이 "a dependency vulnerability scan", "audits every manifest at the project root"라고 주장 | 기준 훅 주석은 `dependency manifest audit`(3행, 287행)뿐. 두 문구는 기준 문서 142·145행에 있고, 뒤 문구는 인용부호 안의 의역 | **부분 부정확 → F7** |
| 두 문서 사본·두 훅 사본 수정, Phase 8은 추가 렌즈 | local 사본 diff가 템플릿 diff와 같은 문장; `cmp` 동일 | 일치 |
| AC 15개(acceptance.md §D), mutant 27개(§D.16) | §D 표 15행, §D.16 표 M1–M27 27행(M22는 두 절로 측정) | 일치 |
| 증거 `.moai/reports/t624/` | `ls` 결과 m1-*~m4-* 존재 | 일치 |
| sync 커밋: `status in-progress → completed`, `updated`만 | `git show 9af4acb92 -- spec.md`: `-status: in-progress +status: completed`, `-updated: 2026-09-10 +updated: 2026-09-11` 두 줄뿐. `--stat`: progress.md·spec.md·CHANGELOG.md 3파일 | 일치 |
| `sync_commit_sha: pending-backfill` | progress.md §E.4 | 일치 |

### 2.5 보안 프로브 (scratch 저장소, 추적 파일 무관)

primitive — 훅 `write_state_file` 214-215행과 같은 명령을 macOS에서 실행:

```
(a) 기록 경로가 디렉터리를 가리키는 심볼릭 링크
    mv -f .../state/.sync-quality-gate.last.tmp.1 .../state/sync-quality-gate.last  → mv_rc=0
    outside/ 안에 .sync-quality-gate.last.tmp.1 생성, state/ 에는 링크만 남음      → 상태 디렉터리 밖으로 쓰기
(b) 기록 경로가 파일을 가리키는 심볼릭 링크
    mv_rc=0, victim.txt = ORIGINAL (불변), 링크가 일반 파일로 교체됨                 → 안전
(c) 기록 경로가 실제 디렉터리
    mv_rc=0, 디렉터리 안에 .t 로 들어감, 기록 파일은 생기지 않음
(d) 예측 가능한 임시 이름에 미리 심어둔 심볼릭 링크
    printf 'abc fail\n' > .../state/.sync-quality-gate.last.tmp.4242  → write_rc=0
    victim.txt = abc fail                                                          → 링크 대상 덮어씀
```

stdin 판별 정규식(훅은 `[ \t]`, 프로브는 동치인 `[[:blank:]]`, `/usr/bin/grep -E`):

```
1:{"stop_hook_active":true}                                  → 일치
2:{"stop_hook_active" :  true}                               → 일치
3:{"m":"quoting \"stop_hook_active\": true here"}           → 불일치 (문자열 안 이스케이프 제외, 의도대로)
4:{"m":"ends with backslash \\","stop_hook_active":true}    → 일치 (실제 키, 올바름)
5:{"a":{"stop_hook_active":true}}                            → 일치 (중첩 키 — §D.17 선언된 갭)
6:{"stop_hook_active":false}                                 → 불일치
7:{"session_id":"x","stop_hook_active":true}                 → 일치
grep_rc=0
```

훅 전 구간 실행 — scratch 저장소 `p1`(go.mod + 컴파일 실패 main.go, 제목 `docs(p): sync-phase probe`), `CLAUDE_PROJECT_DIR`·모드 변수 unset, 템플릿 훅 경로 직접 실행:

```
RUN1  .moai/logs 를 일반 파일로 둠
      exit=1, stdout 237바이트(차단 JSON), stderr: mkdir: .../.moai/logs: File exists
      .last = 8c3a95ca57f39261dbbbda8760e504525747893a fail
      payload 1행 = 8c3a95ca57f39261dbbbda8760e504525747893a block 1 1
RUN2  같은 상태로 재실행
      exit=0, stdout 237바이트, IDENTICAL-TO-RUN1
      stderr: ...sync-phase-quality-gate.sh: line 225: .../.moai/logs/sync-quality-gate.log: Not a directory
RUN3  payload 본문을 {"decision":"approve","reason":"INJECTED-BY-THIRD-PARTY"} 로 교체(헤더 SHA 일치)
      exit=0, stdout = {"decision":"approve","reason":"INJECTED-BY-THIRD-PARTY"}
RUN4  .last = <HEAD> running, payload 삭제, mtime = 2030-01-01
      exit=0, stdout = {"systemMessage":"sync-phase quality gate: the previous gate run for this HEAD has not completed yet, so no checks ran this turn. To force a new gate run, delete .moai/state/sync-quality-gate.last."}
      .last 불변 (8c3a95c... running)
```

gitignore:

```
$ git check-ignore -v .moai/state/sync-quality-gate.last .moai/state/sync-quality-gate.payload
.gitignore:354:.moai/state/	.moai/state/sync-quality-gate.last
.gitignore:354:.moai/state/	.moai/state/sync-quality-gate.payload
templates/.gitignore:240:.moai/state/
```

## 3. 발견 목록 (structured defect-list)

| ID | 심각도 | 분류 | 위치 | 내용 | 확신 | 권장 수정 |
|---|---|---|---|---|---|---|
| F1 | Low | optional | `sync-phase-quality-gate.sh:347` (검증 `:316-329`) | 저장된 payload 본문을 `tail -n +2`로 그대로 재방출한다. `PAYLOAD_VALID`는 헤더(SHA=HEAD, 종류 `block\|advisory`, 두 정수, 추가 필드 없음)만 본다. 헤더만 맞춘 조작 본문이 임의의 Stop JSON으로 나간다(RUN3 관측) | 높음 | 헤더 필드(kind, C1, C2)로 JSON을 다시 만들거나, 본문이 이 훅의 차단 템플릿 접두부로 시작하는지 검사 |
| F2 | Low | optional | `:214-215` `write_state_file` | 목적지가 디렉터리를 가리키는 심볼릭 링크이면 `mv -f`가 임시 파일을 링크 대상 디렉터리 안으로 옮기고 rc 0을 낸다 → `.moai/state` 밖에 점파일 생성, 기록은 안 됨. 실제 디렉터리도 같다. 결과는 매 턴 재게이트(조용한 통과 아님) | 높음 | 쓰기 전에 `[ -d "$1" ]`면 쓰지 않고, `[ -L "$1" ]`면 링크를 먼저 제거 |
| F3 | Low | optional | `:214` | 임시 이름이 `.<name>.tmp.$$`로 예측 가능하고 `cat >`가 심볼릭 링크를 따라간다 → 미리 심은 링크 대상이 게이트 내용으로 덮어써진다(프로브 d). PID 추측 필요 | 중간 | `mktemp "$STATE_DIR/.${1##*/}.XXXXXX"` 사용, 또는 해당 리다이렉트만 noclobber |
| F4 | Low | optional | `:253-261`, `:355` | 기록 mtime이 미래이면 나이가 음수라 `-le 60`을 통과한다 → mtime+60초가 지날 때까지 매 턴 notice만 내고 checks를 돌리지 않는다(RUN4 관측). notice는 보이므로 조용한 통과는 아니다 | 높음 | 음수 나이는 stale로 취급 |
| F5 | Low | optional (기존 결함) | `:574-576` | 마지막 로그 두 줄에 가드가 없어 `set -e`에서 `.moai/logs`가 디렉터리가 아니면 차단 JSON을 찍은 **뒤** exit 1로 끝난다(RUN1). stdout JSON은 exit 0에서만 채택되므로 그 턴의 차단은 사라진다. 다음 턴은 저장된 payload로 exit 0 재전달된다(RUN2, 바이트 동일) — 이번 설계가 영향을 한 턴으로 줄였다. spec.md §4 "Always exit 0"과 이 카드가 새로 쓴 `:211-212` 주석 "the gate exits 0 on every path"가 이 경로에서는 거짓 | 높음 | 두 줄을 `log_gate_event`와 같은 가드(`2>/dev/null \|\| true`)로 감싸거나 `log_gate_event` 재사용; 주석을 사실에 맞춤. 기준 훅 391-393행부터 있던 경로라 후속 카드 적합 |
| F6 | Info | optional | `:225` | `>> file 2>/dev/null` 순서 때문에 리다이렉트 실패 메시지가 stderr로 샌다(RUN2 stderr). 동작 영향 없음 | 높음 | `2>/dev/null`을 `{ echo …; } 2>/dev/null >> file` 형태로 앞에 적용 |
| F7 | Low | optional | `CHANGELOG.md:16` | 인용한 두 문구를 "the sync workflow document and the hook's own comments"에 함께 귀속하지만, 훅 주석은 `dependency manifest audit`만 말했고(기준 3·287행) 두 문구는 문서(기준 142·145행)에 있다. "audits every manifest at the project root"는 인용부호 안의 의역 | 높음 | 표면별로 귀속을 나누고 의역에는 인용부호를 쓰지 않음 |
| F8 | Info | optional | `:246` | 중첩 `stop_hook_active` 키도 일치(프로브 5행). §D.17에 선언된 미검증 갭이며 Claude Code Stop stdin에는 해당 중첩 객체가 없다 | 높음 | 조치 불요(선언 유지) |
| F9 | Info | optional (범위 밖, t644) | `.claude/rules/moai/core/agent-common-protocol.md` § Hook Invocation Surface 외 | 이 훅을 "lint + test + coverage delta"로 서술하는 표면들은 spec.md §5("other claims", "other hook scripts") 밖이며 리드가 t644로 이관. 추가로 같은 절이 "advisory unless `MOAI_SYNC_GATE_BLOCKING=1`"이라 쓰는데 훅 13-19행은 기본 차단이다 — 이 역시 기존 서술, t644에 합류 권장 | 중간 | t644에서 함께 정정 |

차단 발견이 없으므로 FAIL 전환 사유는 없다. F5는 SPEC이 명시한 제약(§4 "Always exit 0")과 어긋나는 경로지만, 이 카드가 바꾸지 않은 기존 줄이고 도달 조건이 인위적이며 이번 변경으로 영향이 줄었으므로 선택으로 분류했다. 리드가 달리 판단하면 차단으로 올릴 수 있다.

## 4. 보안 렌즈 답변 (1–6)

1. **상태 경로.** 모든 경로는 `"${CLAUDE_PROJECT_DIR:-$PWD}/.moai/state/..."`로 따옴표 안에서 조합되므로 공백이 있는 프로젝트 루트도 안전하다. `.moai/state` 자체가 링크이면 쓰기는 그 대상으로 가지만, 이는 사용자가 구성한 것이다. 상태 파일 경로에 **파일을 가리키는 링크**가 있으면 `mv -f`가 링크만 교체한다(대상 불변, 프로브 b). **디렉터리를 가리키는 링크**나 실제 디렉터리가 있으면 임시 파일이 그 안으로 들어가 상태 디렉터리 밖에 쓴다(F2, 프로브 a·c). 미리 심은 임시 이름 링크는 `cat >`가 따라가 대상을 덮는다(F3).
2. **저장 내용 신뢰.** 헤더만 검증하고 본문은 검증하지 않아, 헤더를 맞춘 조작 본문이 임의의 `decision`/`reason`/`systemMessage`로 나간다(F1, RUN3). 현실적 심각도는 Low다. 상태 파일은 gitignore 대상이라 clone으로 전달되지 않고, 헤더 SHA는 HEAD와 같아야 하는데 그 파일을 커밋에 담으면 HEAD SHA가 바뀌므로 커밋으로 심을 수 없다. 남는 경로는 사용자 권한으로 저장소에 쓰는 로컬 프로세스뿐인데, 그런 프로세스는 이미 `.claude/hooks`나 `CLAUDE.md`를 고칠 수 있다. 그래도 `reason`은 모델에 입력으로 들어가므로, 재방출 대신 헤더 필드로 재구성하는 편이 공격면을 없앤다.
3. **부분 쓰기 실패.** payload 쓰기 실패 + fail 기록 성공 → payload 없음 → 재게이트(AC-013 S3 통과, M18 적색). 기록 쓰기 실패 → 이전 `running`이 남아 notice/1회 재실행/소진 notice, 또는 기록 없음 → 재게이트. `.moai/state` 쓰기 불가 → 기록이 안 남아 매 턴 재게이트(TA/TB 행). 디스크 가득 참 → `cat`이 실패해 원자 교체가 일어나지 않음(코드 판독). 조용한 통과 경로는 관측되지 않았다. 다만 exit 0이 **모든** 경로에서 성립하지는 않는다: `.moai/logs`가 디렉터리가 아니면 기존 마지막 로그 줄에서 exit 1(F5). `.moai/state`만 불가하고 `.moai/logs`는 가능한 경우는 AC-005/AC-007 행이 exit 0을 단언한다.
4. **임시 파일과 권한.** 이름 `.<name>.tmp.$$`는 PID로 동시 실행 간 충돌은 없지만 예측 가능하다(F3). 실패 시 `rm -f`로 정리하나, 런타임이 `cat`과 `mv` 사이에서 SIGKILL하면 잔여 파일이 남는다(관측 안 함). 생성 모드는 umask를 따르며 프로브에서 `-rw-r--r--`(0644). world-writable 생성은 없다. `mktemp -d` 작업 디렉터리는 0700.
5. **입력 파싱.** `stop_hook_active`는 줄 단위 grep이다. 이스케이프된 문자열 안의 키는 제외하고, 공백·탭을 허용하며, 중첩 키는 구분하지 못한다(F8, 선언된 갭). 여러 줄로 나뉜 JSON은 놓치지만 Claude Code가 한 줄 JSON을 보내므로 영향이 낮다. `HEAD_SHA`는 `git rev-parse` 출력이고 `case` 패턴 안에서 따옴표로 쓰여 리터럴 비교다. 헤더 정수는 `*[!0-9]*`로 검증된다. 이 카드가 추가한 줄에서 공격자 영향 문자열이 따옴표 없이 전개되는 곳은 없다. 기존 줄 `trap "rm -rf $GATE_TMPDIR"`(385행)과 `git diff ... -- $DEPS_MANIFESTS`(504행)는 따옴표 없는 전개가 있으나 값이 `mktemp -d` 결과와 고정 목록이다. `mv`를 이름으로 찾는 것은 PATH 탈취를 전제로 하는데, 훅은 이미 `git`·`cat`·`grep`·`go`를 같은 방식으로 호출하므로 새 표면이 아니다(S3 shim이 이 사실을 테스트 장치로 쓴다).
6. **피드백 영구 차단.** 조용한 영구 침묵에 이르는 경로는 셋이다. ① 같은 HEAD에서 advisory payload가 저장된 뒤 모드가 차단으로 바뀌면 그 HEAD에서는 다시 알리지 않는다 — AC-008 A6이 의도로 고정한 동작이고, 변경 전(두 번째 턴부터 무조건 침묵)보다 느슨하지 않다. ② `<HEAD> pass`를 위조하면 침묵한다(F1과 같은 로컬 쓰기 전제). ③ 커밋 없이 계속 60초를 넘기는 빌드는 두 번 죽은 뒤 소진 notice만 반복한다 — checks는 더 돌지 않지만 매 턴 보이며 삭제 방법을 알려준다. 미래 mtime은 시계가 따라잡을 때까지 같은 notice 상태를 만든다(F4). 어느 경로도 기록 삭제로 복구되고, notice는 차단 한도를 소모하지 않는다(AC-015).

## 5. 기준 귀속 (Baseline-attribution)

- 모든 명령은 이번 실행에서 `.claude/worktrees/t624`, HEAD `9af4acb92`(clean)에 대해 실행했다.
- Go 테스트·vet은 이 트리에서 새로 컴파일한 바이너리로 판정했다. 훅 프로브는 이 트리의 템플릿 훅 파일을 경로로 직접 실행했다(설치본 아님).
- `golangci-lint`는 `/opt/homebrew/bin/golangci-lint` 설치본이며 버전은 기록하지 않았다(§6).
- 이전 증거(`m4-*`)는 재측정 대상이 아니라 교차 확인 대상으로만 인용했다. 셀렉터·템플릿 패키지·cmp·앵커·gofmt·vet·lint는 이번 실행의 새 측정이다.
- mutant 판정은 커밋된 `m4-mutant-*.txt`를 읽은 것이다(재적용 금지 지시에 따름).

## 6. 미검증 (Gaps)

- `shellcheck` 부재로 셸 정적 분석은 실행하지 못했다.
- 셸 스크립트의 커버리지 수치는 측정 수단이 없다. mutant 27건 기록으로 대신 판단했다.
- mutant 증거는 M06, M18(1차·재실행), M20, M24 다섯 파일만 다시 읽었다. 나머지 23개 파일은 progress.md 표에 의존했다.
- Windows에서는 실행하지 않았다(S3는 Windows에서 skip).
- 프로브는 macOS의 BSD `mv`/`stat`에서만 했다. GNU `mv`(`-T` 없음)의 디렉터리 링크 동작은 같다고 알려져 있으나 관측하지 않았다.
- 정규식 프로브는 리터럴 탭 대신 동치인 `[[:blank:]]`를 썼다.
- 디스크 가득 참과 SIGKILL 중간 종료는 코드 판독만 했고 재현하지 않았다.
- t644 이관 표면(`sync.md:43`, `agent-common-protocol-reference.md:222`, `archived-agent-rejection.md` 81·88·99행, docs-site en/ko)은 한 줄씩 열어보지 않았다. 범위 밖 판정은 spec.md §5 문구에 근거한다.
- 평가 프로필 파일은 로드하지 않았다. 리드가 지정한 4차원·조화평균을 따랐다.
- 교차 모델 감사(`mcp__moai__audit_multi` 등)는 호출하지 않았다.
- `golangci-lint` 판정 빌드 버전은 기록하지 않았다.

## 7. 잔여 위험 (Residual-risk)

- F1은 로컬 쓰기 권한이 있는 프로세스가 모델에 넘어가는 `reason` 텍스트를 바꿀 수 있는 통로다. 위협 모델상 새 권한을 주지는 않지만, 재방출 설계를 유지하는 한 이 통로는 남는다.
- F5 경로에서는 첫 실패 턴의 차단이 한 번 사라진다. 사용자가 그 턴에 다음 작업을 이어가면 다음 Stop에서야 차단을 본다.
- S3는 `mv`를 이름으로 호출한다는 사실에 기대므로, 향후 절대경로 `mv`로 바꾸면 S3는 해석 불가(갭)가 된다. 설계상 통과로 읽히지는 않는다.
- 두 세션이 같은 루트의 `.moai/state`를 공유하면 기록과 payload가 서로 다른 실행에서 올 수 있다. spec.md §5가 잠금을 범위 밖으로 둔 부분이다.
- 프로브 산출물(`p1-run*.out/err`, `prim/*`)은 세션 scratchpad에만 있고 반출하지 않았다. 판정 근거로 필요한 부분은 이 문서 §2.5에 원문으로 옮겼으며, 나머지는 인용하지 않는다.

## 8. 권고

- 후속 카드(선택): F5 가드와 주석 정정, F2·F3을 `mktemp` + 디렉터리/링크 거부로 함께 처리, F4 음수 나이 처리, F1 재방출을 헤더 기반 재구성으로 교체. 이 네 건에 해당하는 테스트 행(심볼릭 링크, 디렉터리, 로그 쓰기 불가, 미래 mtime, 조작 본문)을 추가.
- F7은 다음 CHANGELOG 편집 때 표면별 귀속으로 정정.
- F9 두 번째 항목(`advisory unless MOAI_SYNC_GATE_BLOCKING=1`)을 t644 범위에 합류.

---

적용한 정책 조항: `verification-claim-integrity.md` §1.1 surface 3, §2, §3 (5절 증거 형식); `verification-completeness.md` §1.1 (셀렉터 스윕 수를 판정 전에 확인); sync-auditor 정의의 Finding-consumption discipline (차단/선택 분류).
