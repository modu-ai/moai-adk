# t458 판정문 — 잔재 WT-* 브랜치 정리 + BranchGuard 조회-과다매칭 판정

- 카드: t458 (Class B — plan 생략, run → sync)
- 브랜치: `WT-guard-query-doc` (워크트리 `.claude/worktrees/t458`, develop 팁 `d63dee78d`에서 ff 흡수)
- 세션: ccc2c966 (lane-7) · 측정일: 2026-09-03
- 기준 트리: 이 커밋 직전 HEAD `d63dee78d` (브랜치 팁 SHA는 완료 보고에 기재)

---

## 주장 (Claim)

1. **축⑴**: 잔재 WT-* 브랜치 11건(lane-6 3 + lane-10 7 + t224 제외 해제 1)이 전부 안전 삭제됐다.
2. **축⑵**: "BranchGuard가 읽기 조회까지 막는다"는 카드 전제는 **측정으로 기각** — 설치 바이너리는 이미 요구 상태(primary에서 조회 허용 · 상태변경 거부)다. **코드 변경 0건.**
3. **기록 정정**: 오기의 재발원인 룰 스텁을 1.3.1 → 1.3.2로 정정했다(금지표가 `git branch` 전체를 금지로 읽히는 행 + 조회/뮤테이션 판별 설명 부재).
4. **인접 발견(측정)**: `git branch -f`는 가드를 **통과**한다(언더매치) — 후속 카드 권고.
5. 재발 방지 규율 (a)(b)를 이 판정문에 기록했다(§ 재발 방지).

## 증거 (Evidence)

### 축⑴ — 브랜치 11건 삭제

사전 조상검증(10건, `git merge-base --is-ancestor <branch> develop; echo rc=$?`, develop=`5107bbfff` 시점) — **전부 rc=0**:

```
chain-node-spawn rc=0        req-parser-widen(t385) rc=0   fold-ac-retrofit(t390) rc=0
navigator-hook-restore rc=0  audit-evidence-store(t386) rc=0  ble4-vacuity-count(t396) rc=0
team-ac-verify-wiring rc=0   audit-advice-integrity(t387) rc=0  syncsha-void-axes(t397) rc=0
                             e5-residue-verdict(t434) rc=0
```

사전 팁 확인(`git show-ref --heads | grep WT-`) — 리드가 건넨 7개 SHA(`19b745cf`·`d537dfa4`·`a73708b9`·`64ad1a68`·`3ea189cc`·`6e70768c`·`e1e78cb2`)와 전부 일치. lane-6 3건의 팁(`793800986`·`44a428a3c`·`cf96d0f0a`)도 lane-6 메모리와 일치.

조상 밖 1건(`WT-lane-spawn-authority`@`8b6c30f77`) 트리 비교 — **내 측정**:

```
git rev-parse 9dde832d8^{tree}                 → 614fae83140a21bd696aabeaf4b9eaf5ebf170c5
git rev-parse WT-lane-spawn-authority^{tree}   → 614fae83140a21bd696aabeaf4b9eaf5ebf170c5  (바이트 동일)
git diff 9dde832d8 WT-lane-spawn-authority --stat | wc -l → 0
git merge-base --is-ancestor WT-lane-spawn-authority develop → rc=1 (조상 밖 확인)
```

삭제 11건 — 10건 `git branch -d`, 1건 `git branch -D WT-lane-spawn-authority`(트리 동등성이 근거). 전부 `Deleted branch ... (was <sha>)`였고 **was 값이 사전 팁과 11/11 일치**.

사후 검증:

```
git show-ref --heads | grep -c 'refs/heads/WT-'  → 72   (삭제 전 83)
git show-ref --heads | wc -l                     → 97   (삭제 전 108; 중간에 타 레인 신규 생성분 있음)
잔재 앵커검사(grep -E '...(11개 이름)$')          → rc=1 (11건 전부 부재)
git show-ref --verify refs/heads/WT-doctor-freshness-reds → 86862826a (보존 확인)
타 레인 브랜치 8건 스팟검사(goal-prose-arm 등)     → 8/8 존재
```

### 축⑵ — 가드 전제 판정 (설치 바이너리: `v3.2.0-rc.0` · `v3.1.2-1490-ge79c010b8` · built 2026-09-03T00:15:26Z)

라이브 Bash 경로(primary에서 실행):

- `git branch -vv | head -5` → **실행됨**(WT-* 목록 출력 반환 — 가드가 통과시킴)
- `git branch --show-current` → **`main` 반환**

합성 PreToolUse 페이로드(`... | moai hook pre-tool`, 실제 git 실행 없음):

| 페이로드 command | 판정 |
|---|---|
| `git branch -vv` | `{"permissionDecision":"allow"}` rc=0 |
| `git checkout -b t458-probe` | **deny** — `BRANCH_GUARD_VIOLATION: git checkout <branch/-b> in primary checkout ...` |
| `git branch -D WT-req-parser-widen` | **deny** — `BRANCH_GUARD_VIOLATION: git branch in primary checkout ...` |
| `git branch -f t458-probe-head HEAD~1` | **allow (언더매치 — 인접 발견)** |

소스·설정 근거(HEAD `d63dee78d` 기준 판독):

- `internal/hook/branch_guard.go:120` — 실제 패턴은 `\bgit\s+branch\s+(-[dDmMcC]\s+)?[^\s-]`. 뮤테이션 플래그 그룹 + 첫 비플래그 토큰 규칙으로 조회(`-vv`, `--list`, `--show-current`, bare)를 통과시킨다. 리드 배차문이 인용한 무차별 `\bgit\s+branch\b`는 소스에 존재한 적 없음 — t42(2026-08-15) pickaxe 기록과 일치.
- `.claude/settings.json:402` `"Bash(git branch:*)"` — **allow 목록**(397~415행 git 명령 allow 블록) 소속. deny 아님.
- `.moai/config/sections/workflow.yaml` — `branch_guard.enabled: true`(로컬 도그푸드 옵트인; 배포 기본 false).

### 기록 정정 (축⑵의 산출물)

- `internal/template/templates/.claude/rules/moai/workflow/main-checkout-branch-guard.md` + 로컬 미러 같은 편집: 금지표 행을 뮤테이션 형태로 한정(`git branch <name>` / `-d|-D|-m|-M|-c|-C`), 허용 목록에 조회 패밀리 명시, "Query-vs-mutate discrimination" 불릿 신설(`-f`/`-u` 언더매치 수용 잔여 포함), Version 1.3.1 → 1.3.2.
- 검증: `TestRuleTemplateMirrorDrift` PASS · `TestLateBranchTemplateMirror` PASS · `TestRuleDateProvenance` PASS · `TestRuleProvenanceAudit` PASS.

## Baseline 귀속 (Baseline-attribution)

- 축⑴: 본 세션(lane-7), 워크트리 `.claude/worktrees/t458`, HEAD `d63dee78d`, 2026-09-03에 직접 측정. 조상검증은 develop=`5107bbfff` 시점, 삭제는 HEAD=`d63dee78d` 시점 — 조상 관계는 단조적이라 이동이 판정을 무효화하지 않는다.
- 축⑵: 설치 바이너리 `~/go/bin/moai`(위 버전)에 대한 라이브 + 합성 관측. 소스 판독은 HEAD `d63dee78d`.
- 패턴 이름(`refs/heads/WT-` 접두 for-each-ref 빈 출력) 사건: 슬래시 경계 매칭 한계로 첫 나열이 빈 출력을 냈고 `git show-ref`로 재측정해 해소 — 부재 판정은 나열 도구 형태를 검증한 뒤에만 유효.

## 미검증 (Gaps)

- `git branch -u` / `--set-upstream-to`의 라이브 판정은 측정하지 않았다(`-f`만 측정). 패턴 정독상 통과 추정이나 **미측정**이다.
- exotic 결합 플래그(`git branch -vD x`) 라이브 미측정 — 소스 주석의 문서화된 수용 잔여.
- "조회가 막혔던" 과거 관측의 원 바이너리는 확보하지 못했다. t42 pickaxe("무차별 패턴은 커밋 이력 전체에 존재한 적 없음")에 근거한 이진 지연/이격 귀속이 기록상 유일한 설명이지만, 당시 바이너리 자체는 재현 불가.
- 본 세션은 bypass 환경이라 settings.json allow 목록의 프롬프트 절감 효과는 관측 불가.
- CLAUDE.local.md §4.1 마찰 메모(“BranchGuard가 --list/-vv까지 막는다”) 정정은 **미수행** — 해당 파일에 타 행위자 미커밋 수정(`M CLAUDE.local.md`)이 있어 이번 커밋에서 손대지 않았다. 다음 편집자가 이 판정문을 참조해 정정할 것.

## 잔여 위험 (Residual-risk)

- `git branch -f` 통과는 유지 중 — primary에서 강제 브랜치 이동이 가드를 우회한다. 후속 카드 권고: 플래그 클래스 확장(`[dDmMcC]` → `+f`)과 `-u`/`--set-upstream-to` 처리 + 테스트. (문서는 이번에 실제 동작과 일치시켰으므로 문서-코드 불일치는 없음)
- 카탈로그 해시 적색(`sync-auditor` CATALOG_HASH_UNSTABLE)과 agents-emit `.toml` 드리프트는 본 브랜치와 무관하게 develop에 열려 있다(소관: 운영자 큐 `4244c4a06` / t443-t444 계열). 수리 전까지 `make build`는 agents-emit-check에서 중단된다 — 본 카드는 catalog.yaml이 룰 문서를 해시 추적하지 않음(grep 0매치)을 확인해 재생산 의무 없음을 확인했다.
- 문서 정정에도 불구 운영자의 과거 세션 기억이 같은 카드를 재발행할 수 있다 — 재발 시 이 판정문의 레시피(합성 페이로드 4건)로 즉시 재판정 가능.

## 재발 방지 (리드 지시 반영)

- **(a) 레인 규율**: develop 병합 착지 후에는 카드 브랜치에 어떤 커밋도 얹지 않는다. 흡수 누락을 뒤늦게 발견하면 리드 보고가 우선 — 조용한 추가 커밋이 조상 밖 커밋을 다시 만든다.
- **(b) 잔재 정리 규율**: 조상 밖 브랜치는 조상 검사만으로 삭제 판정을 내리지 않는다. `git rev-parse <병합>^{tree}` vs `<브랜치팁>^{tree}` **트리 비교**가 먼저다(동일=삭제 안전, 상이=콘텐츠 존재). t460 판정서 §3 레시피와 동일.
- **대조군**: `WT-doctor-freshness-reds`의 흡수 `86862826a`는 병합 **이전** 흡수라 develop 계열 안(→ 보존). t224의 `8b6c30f77`는 병합 **이후** 역산 흡수라 계열 밖(→ 트리 동등 확인 후 삭제). 한 카드에서 두 형태가 나온 판별 사례.

## 삭제 대상 최종 목록

| 브랜치 | was | 근거 |
|---|---|---|
| WT-chain-node-spawn | 793800986 | 조상 rc=0, `-d` |
| WT-navigator-hook-restore | 44a428a3c | 조상 rc=0, `-d` |
| WT-team-ac-verify-wiring | cf96d0f0a | 조상 rc=0, `-d` |
| WT-req-parser-widen (t385) | 19b745cfc | 조상 rc=0, `-d` |
| WT-audit-evidence-store (t386) | d537dfa49 | 조상 rc=0, `-d` |
| WT-audit-advice-integrity (t387) | a73708b99 | 조상 rc=0, `-d` |
| WT-fold-ac-retrofit (t390) | 64ad1a681 | 조상 rc=0, `-d` |
| WT-ble4-vacuity-count (t396) | 3ea189cc9 | 조상 rc=0, `-d` |
| WT-syncsha-void-axes (t397) | 6e70768c1 | 조상 rc=0, `-d` |
| WT-e5-residue-verdict (t434) | e1e78cb2c | 조상 rc=0, `-d` |
| WT-lane-spawn-authority (t224) | 8b6c30f77 | 조상 밖(rc=1) + 트리 동등(614fae831…) → `-D` |

보존: `WT-doctor-freshness-reds`@`86862826a` (t444 미푸시 유일본).
