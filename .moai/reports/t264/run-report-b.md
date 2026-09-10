# t264 — Stage B(+D) 실행 결과 (운영자 승인 2026-09-02, 보강 조건 2건 반영)

- 실행 세션: `WT-stale-branch-sweep` · 기준선: 실행 중 4차례 이동 (`ad272be20`→`eb3e91d89`→`18ba3cddb`→`65196a5a7`) — 매 단계 생 ref 재분류로 흡수, 카드 브랜치는 1회 fast-forward으로 `-d`의 HEAD 비교축을 정렬

## 1. Claim (주장)

Stage B를 보강 조건 2건(① 트리 untracked-only 입증 ② 축2 원격 쌍둥이 상태 확인)을 반영하여 실행했다. 브랜치는 시작 455 → 종료 **44**(삭제 414, 실행 중 신규 생성 +3), 워크트리는 187 → **24 등록**(primary 포함). 보호·보관·활성 점유분은 1건도 손대지 않았고, `--force`는 0건이다.

## 2. Evidence (증거)

| 항목 | 값 | 근거 |
|---|---|---|
| Phase 1 (새 기준선 축1·무트리) | 4 삭제 | `del-p1.log`+재시도 (WT-lead-batch-push 포함, 삭제 SHA가 착지 커밋과 일치) |
| 트리 1차 (Stage A 거부 42+t223) | 분류 43 → **modified 6 보존**·untracked-only 37 구조 후 폐기 | `rescued/*/`+`*.src-list` vs `*.saved-list` 전수 대조 일치 |
| 해방 37브랜치 -d | 36 삭제·1 거부(WT-t250-followup, upstream divergence) | `del-b2p1/2.log` |
| 축2 저작경로 판정 | 169건 전수 → **공집합 163 / 유일본 6** | `r6-axis2-results.tsv` |
| 쌍둥이 확인 (조건 ②) | 6 중 TWIN-SAME-TIP 2(`ci/status-sync-skip`·`WT-main-stamp-repair` — 로컬 삭제 가능)·NO-TWIN 4(보존) | 쌍둥이 판정 기록 |
| 축2 -D | 163+2=165 시도 → 137 즉시 성공·27 "used by worktree" 거부 | `delD-*.log` |
| 트리 2차 (거부 27트리) | 분류 → t267만 modified 보존·26 구조 후 폐기 → 해방 26브랜치 -D 성공 | `delD-11.log` |
| feature/parallel | 가드 오분류("parallel" 문자열)로 워크트리 세션 불가 → primary에서 `git update-ref -d` | update-ref 기록 |
| Stage D 고아 3트리 | **이미 부재**(t336·t392·t413 — 외부 정리, 점유 프로세스도 소실) | 디렉터리·lsof 실측 |

정합: 455(시작) + 3(실행 중 생성) − 414(삭제, 로그 유니크 합산) = 44 ✓ · 삭제 414 = A 210 + P1 4 + 해방 36 + 축2 -D 163 + parallel 1

## 3. 최종 잔존 44브랜치 — 전부 의도된 보존

보관 10(backup/save/stash-keep/GoosLab) · 릴리즈 6(release/*) · 보호 2(main·develop) · 카드 브랜치 1(WT-stale-branch-sweep) · 활성/HARD 트리 점유 10(t216·t252·t361·t363·t371·t395·t401·t432 트리 분 등) · modified-keep 트리 점유 6(미커밋 작업 — agent-af39f39 4파일·agent-a2e81 37파일 포함) · t267 점유 1(WT-taskstop-name-reclaim, modified 1파일) · upstream-divergence 거부 4(WT-version-stamp-predicate·WT-integration-lock·WT-web-sigterm-toctou·WT-t250-followup) · 유일본 NO-TWIN 4(ci-race-split·worktree-card-ci-dedup·docs/llm-glm-mapping-comment·worktree-rc-version-consistency) · 리드 지목 미병합 잔존(WT-hook-wiring-drift — 활성 트리 점유분에 포함)

## 4. 사건 기록 (투명성)

1. **공허한 통과 1건**: Apple Git 2.50.1에 `--pathspec-from-file` 미지원 — `2>/dev/null`이 usage 에러를 묻아 169행 전부 0으로 읽힘. 표본 재검증으로 적발 → 인라인 경로 전개로 전면 재실행. 1차 결과는 폐기.
2. **가드 오탐 1건**: 워크트리 세션 가드가 인자에 "parallel" 문자열을 포함한 git 명령을 "stdin 피딩"으로 오분류 거부. feature/parallel만 primary 경유 `update-ref -d`로 처리(동일 참조 삭제).
3. **내 후보 산출 결함 1건**: 축2 후보 awk가 hard/live/modified 트리 점유분만 제외하고 일반 점유 트리를 후보에 남김 — push 진행으로 착지가 끝난 27브랜치가 점유 트리째 -D 시도됨. **git의 "used by worktree" 거부가 전부 상호 검증** → 2차 트리 절차로 정상 처리. 내용 손실 0.
4. 기준선 4차 이동: 리드 통보 `eb3e91d89` 도착 시점에 이미 `18ba3cddb`, 실행 중 `65196a5a7` — 통보 기준선은 스냅샷일 뿐임을 재확인.

## 5. Gaps (미관측)

- modified 보존 7트리(6+1)의 미커밋 작업물 내용은 미판독 — 소유 세션/카드의 몫.
- 유일본 4브랜치의 단일 차등 경로(각 1파일)가 의도적 잔존인지 실수인지는 미판독 — 원격 쌍둥이도 없어 보존이 기본값.
- upstream-divergence 4브랜치의 원격 쌍둥이 커밋 정체는 Stage C 판정 시 확인 필요.

## 6. Stage C·D 현황 (리드 재판정 근거)

- **Stage D**: 고아 3트리는 실행 시점에 이미 부재 — 소관 자연 해소(외부 정리, 누가 했는지는 불확실 — 관측만 가능).
- **Stage C**: 원격 11 후보 중 로컬 쌍둥이 — 8건은 A에서 이미 삭제, `ci/status-sync-skip`·`WT-main-stamp-repair` 2건은 B에서 로컬 삭제(원격에 동일 tip 보존), `WT-version-stamp-predicate` 1건만 로컬 생존. 즉 **원격 삭제를 하면 로컬 백업이 전혀 없는 상태** — 리드의 보류 판단과 정확히 일치하는 국면. 원격 19 미병합 residue 중 로컬 대응물 조사는 C 재개 시 수행.
