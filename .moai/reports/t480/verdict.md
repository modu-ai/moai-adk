# t480 판정 — `.claude/settings.json` 블록 "유실" 조사

> 카드: t480 (배차 lane-7, 2026-09-04) · 브랜치: `WT-settings-block-loss` · 워크트리: `.claude/worktrees/t480`
> 원 전제(lane-12 → lead 전달): "병합 후 네 블록(respectGitignore / skillListingBudgetFraction / env / permissions)이 현재 트리 어디에도 없다"

## 판정 요약

**유실 없음. 전제는 반증됐다.** 네 블록은 템플릿이 배포하는 관리 콘텐츠이며, 측정한 모든 트리에 존재한다.
병합이 폐기한 dirty 워킹 사본은 관리본보다 **모든 차원에서 낡은 렌더링**이었고, 되살릴 값은 없었다.

## Claim (주장)

1. 네 블록은 로컬 전용 설정이 아니라 **템플릿 관리 콘텐츠**다 (→ 질문 1: "잔재 vs 의도" 프레임 자체가 성립하지 않음 — 블록은 애초에 유실되지 않았다)
2. 네 블록은 현재 **측정한 5개 위치 전부**에 존재한다 — "현재 트리 어디에도 없다"는 전제는 거짓
3. 병합이 폐기한 dirty 사본에는 **되살릴 가치가 있는 값이 없다** — 유일한 dirty 고유 내용은 구형(더 좁은) 매처 문자열 1줄뿐
4. `settings.local.json`으로 옮길 대상 없음 (→ 질문 2 해당 없음) · 복구 대상 없음 (→ 질문 3 해당 없음)

## Evidence (증거 — 명령 + 실측 출력, 2026-09-04 lane-7 실행)

### E1. 네 블록의 현재 존재 (5개 트리)

| 트리 | 측정 | 네 블록 |
|---|---|---|
| origin/develop 팁 `25a3212a9` (t480 워크트리) | Read `.claude/settings.json` | 400행 `skillListingBudgetFraction` · 404-409행 `env` · 410-586행 `permissions` · 592행 `respectGitignore` ✓ |
| 로컬 develop 워크트리 `a825183dd` | Read 동일 파일 | 동일 내용 ✓ |
| primary 체크아웃 (md5 `437834679fcc…`, 21,846B) | `grep -n` | 359행 `skillListingBudgetFraction` · 363행 `env` · 369행 `permissions` · 551행 `respectGitignore` ✓ |
| t452 워크트리 (병합이 일어난 그 트리) | `md5` = `568a3d3a…` = develop 커밋판과 바이트 동일 ✓ | ✓ |
| 템플릿 `internal/template/templates/.claude/settings.json.tmpl` | `grep -n` | 387행 · 392행 · 403행 · 588행 ✓ |

### E2. 보존 사본 2개의 실측 (lane-12 스크래치패드)

```
/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go--claude-worktrees-t452/
  9f51bf2c-3c20-47c2-9f34-d9bac5bfba83/scratchpad/
    settings.json.develop         24,399B  md5 568a3d3a32a360731d1d02f686f27d24
    settings.json.worktree-dirty  24,667B  md5 b669972dc738d1bf925281dcc90f152e
```

- `settings.json.develop`의 md5는 **현재 develop 커밋판과 정확히 일치** — develop은 t452 창 이후 settings.json을 바꾼 적 없음
- 네 블록은 **두 사본 모두에 존재** (위치만 달랐다 — dirty판은 파일 앞쪽, develop판은 뒤쪽. 이 위치 차이가 원 보고의 "유실" 오판 근원으로 추정)

### E3. 정규화 내용 비교 (키 정렬 `jq -S` — 순서 효과 제거)

```
jq -S 'del(.hooks)' 각 사본 | md5
  develop 보존 사본   = 70414c9c7efe5bc9ccdb01255fe186fa
  dirty 보존 사본     = f19b70fe6dcb5bbf85d94eaba0e26209   ← 진짜 내용 차이 존재
  develop 커밋판      = 70414c9c7efe5bc9ccdb01255fe186fa   (= develop 보존 사본과 동일)
```

정규화 디프 전체 (비-훅 영역 — 단일 훙크):

```diff
134,139c134
<       "Bash(rm:*)",
<       "Bash(sudo:*)",
<       "Bash(chmod:*)",
<       "Bash(chown:*)",
<       "Read(./.env)",
<       "Read(./.env.*)"
---
>       "Bash(sudo:*)"
```

→ **dirty 사본이 5개 ask 엔트리가 "부족한" 쪽.** 병합 결과(develop판)가 더 완전하다.

훅 영역 정규화 디프: develop판에만 `status-transition-ownership.sh` 조건부 3엔트리 + `async`/`if` 플래그 존재(t216 parity 산물). dirty 고유 줄은 전체 파일에서 **1줄**:

```
>       "matcher": "Write|Edit|MultiEdit"
```

(develop판의 확장형 `"Write|Edit|MultiEdit|EnterWorktree|ExitWorktree"`에 대한 구형 값)

### E4. dirty 사본 = 이 리포의 어떤 상태와도 불일치 (소출처 소거)

| 후보 상태 | ask 목록 실측 | dirty([sudo] 1개)와 일치? |
|---|---|---|
| tracked `.claude/settings.json` 전 이력 | `git log -S 'Bash(rm:*)'` → 초기 커밋 `9110b273c`(2026-02-03) 1건뿐 — 존내내 6엔트리 | ✗ |
| 템플릿 생성 전환 `e21d85d8d`(2026-02-07) 이전 생성기판 | `git show e21d85d8d^:… \| jq '.permissions.ask'` → 6엔트리 | ✗ |
| 릴리스 템플릿 v3.0.0 / v3.1.0 / v3.1.3 / v3.1.4 / v3.2.0-rc.0 | 각 워크트리·tag에서 `grep -A8 '"ask"'` → 전부 6엔트리 | ✗ |
| 전역 `~/.claude/settings.json` / 프로젝트 `settings.local.json` | `.permissions.ask` = null (항목 없음) | ✗ |
| 설치 바이너리 `~/go/bin/moai` | v3.2.0-rc.0 (2026-09-03 빌드 — 최신) | 작성자 가설 기각 |

dirty 사본의 훅 구조는 v3.1.4보다도 구형(status-transition 배선 부재)이고 매처는 v3.1.4와 같은 좁은 형태 — 즉 **여러 축에서 서로 다른 시대의 구형 모양이 섞인, 어느 단일 릴리스와도 일치하지 않는 렌더링**.

## Baseline-attribution (귀속)

- 모든 측정은 2026-09-04, lane-7 세션, t480 워크트리(`25a3212a9` 베이스)에서 실행한 위 명령들의 실측 출력이다.
- 보존 사본 원본은 lane-12 세션 스크래치패드에 있으며 lane-7은 복사만 하고 원본을 변경하지 않았다. 바이트 동일 반출본: `.moai/reports/t480/preserved-copies/` (md5 원본 일치 확인).
- 단, "이 사본들이 lane-12가 t452 창에서 보존한 것"이라는 귀속은 파일명(`.develop`/`.worktree-dirty`)·t452 스크래치패드 경로·mtime(2026-09-04 03:52)의 정합에 근거한 **추론**이다 — lane-12 세션은 `/clear` 후라 본인도 보증하지 않는다고 명시했다.

## Gaps (미검증)

1. **dirty 사본을 그 워크트리에 쓴 작성자/프로세스는 미식별.** repo 이력·생성기·릴리스 템플릿 5종·전역/로컬 설정을 모두 소거했으나 일치 후보가 없다. 작성 시각(워크트리 생성 2026-09-03 13:40 ~ 병합 2026-09-04 03:52 사이)만 확보.
2. lane-12 원 보고("워킹 사본에만 네 블록이 더 있었다")와 보존 사본 실측(양쪽 모두 존재)의 어긋남 경로는 lane-12 세션이 `/clear` 되어 직접 확인 불가 — 위치 차이를 내용 차이로 오독했을 가능성이 유력 추정이나, 추정이다.
3. 같은 덮어쓰기가 **다른 워크트리**에도 있었는지는 스윕하지 않았다 (t480 소관 밖 — 별도 조사 카드 후보).

## Residual-risk (잔여 위험)

- 원본 보존 사본은 `/tmp` 아래라 재부팅 시 소실 — 본 반출(.moai/reports/t480/preserved-copies/)이 이를 방지한다.
- "작성자 미식별" 상태이므로 같은 유형의 덮어쓰기가 재발할 수 있다. 다만 재발해도 병합 전 pre-check + 본 판정의 절차(사본 보존 → 정규화 비교)가 같은 결론에 도달하게 만들어 두었다.
- lane-12 원 보고의 다른 서술(병합 거부 메시지 등)은 본 조사에서 재검증하지 않았다 — 다만 병합 거부 자체는 t452 워크트리의 현재 상태(restore 후 develop판 착지, md5 일치)와 정합이다.

## 질문 3개에 대한 최종 답

| 질문 | 답 |
|---|---|
| 1. 네 블록은 의도된 로컬 설정인가, update 잔재인가? | **둘 다 아니다** — 템플릿 관리 제품 설정이며 유실 자체가 없었다. 잔재는 오히려 폐기된 dirty 사본 쪽(출처 미식별의 구형 렌더링) |
| 2. 의도된 것이라면 settings.local.json으로 가야 하나? | **해당 없음** — 옮길 로컬 값이 없다. (블록은 템플릿 의도 배포: CLAUDE.md §13 skillListingBudgetFraction · §15 agent-teams env와 정합) |
| 3. 복구가 필요한가, 필요하다면 어느 값으로? | **불필요** — 병합 결과가 최신·최완전. 되살릴 값 0개 |

## 운영자 결정 항목 (lane-7 → lead 경윤)

1. 본 판정(복구 불필요) 확인 — 동의 시 카드는 조사 완료로 종결 가능 (run 단계에서 구현할 것이 없다)
2. (선택) Gap 1·3의 후속: "워크트리 tracked settings.json을 덮어쓰는 미식별 작성자" 조사/방어 카드화 여부
