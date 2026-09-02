# t264 — 잔여 로컬 브랜치·워크트리 정리: 조사 보고 및 실행 계획

- 카드: t264 (Tier S) · 조사 단계 산출물 — **삭제 0건, 승인 대기**
- 세션: `WT-stale-branch-sweep` (`.claude/worktrees/t264`) · 2026-09-02
- 기준선: `origin/develop` = `ad272be20` (fetch 후 직접 확인 — 리드 통보치와 일치)
- 범위 확장 접수: 원격 `WT-*` 잔재 포함 (리드 통보, 운영자 판정 2026-09-02)

## 1. 실측 요약

| 측정 대상 | 값 | 근거 명령 |
|---|---|---|
| 로컬 브랜치 | **450** | `git for-each-ref refs/heads` |
| ├ origin/develop 조상(병합) | 264 | `git for-each-ref --merged=origin/develop` |
| └ 미병합(고유 커밋 보유) | 186 | `git for-each-ref --no-merged=origin/develop` |
| 워크트리 등록 | **182** (primary 포함, gone 0) | `git worktree list --porcelain` |
| 프로세스 cwd 점유 트리 | 7 | `lsof -d cwd` |
| 원격 heads (실측) | **92** | `git ls-remote --heads origin` |
| ├ develop 병합 | 14 (residue 접두사 11) | `git for-each-ref --merged` × heads 교차 |
| └ 미병합 | 78 (residue 접두사 19) | 〃 |
| stale tracking refs | **66** (tracking 158 − 실측 92) | tracking 전수 vs ls-remote 교차 |
| 고아 트리 디렉터리 | **3** (t336·t392·t413) | porcelain 부재 + 디렉터리·lsof 존재 |

**리드 관측 검증**: 카드 실측(08-25) "WT-* 129" → 오늘 **450**으로 증가(약 3.5배). 오늘-병합 3건
(`WT-integration-lock-atomic`·`WT-stress-invariant-guard`·`WT-graph-report`) + 이전 보고 4건
(`WT-concurrency-stress`·`WT-tempdir-race`·`WT-codemaps-orphan-stamp`·`WT-todo-fail-loud`)
전부 **무트리·develop 병합 완료** → Stage A 삭제후보 확정. 원격에는 오늘-병합 3건 부재 확인
(grep exit 1) — WT push 금지 규율 준수의 증거.

**신규 발견 3건**:
1. **t430(활성 리드 트리)의 브랜치 `WT-lead-batch-push`는 병합 상태** — naive 판별식이 활성
   리드 트리를 폐기 대상에 넣는 함정. manifest에서 제외 완료.
2. **고아 트리 디렉터리 3개** (t336·t392·t413): git 등록부에서는 이미 소멸했는데 디렉터리와
   프로세스 cwd가 잔존. 별도 처리 클래스 필요.
3. `moai session list --json` = `[]` 이었지만 lsof는 7개 트리 점유를 관측 — **레지스트리 빈값은
   "세션 없음"의 근거가 아님**. 점유 판정은 lsof로 수행.

## 2. 판별식 (삭제 안전성 3축)

- **축 1 — 조상 판정 (기계적)**: `git merge-base --is-ancestor <tip> origin/develop` rc=0
  ⇔ 브랜치 내용이 push된 develop에 전부 존재 → 이 축 하나로 삭제 근거 충분.
- **축 2 — 저작경로 공집합 (미병합 전용)**: 축 1 실패 시, 브랜치의 저작 경로
  (`git diff --name-only <merge-base>..<tip>`)에 대해 `git diff origin/main..<tip> -- <paths>`가
  공집합이면 squash 착지로 판정 → 삭제 가능. 비공집합이면 유일본 → 보존.
  (`feedback_landing_proof_needs_authored_paths`. `--merged` 목록 단독 신뢰 = 카드가 지적한
  대표 mutant. 리드 지목 원격 4건이 정확히 이 클래스: `WT-astgrep-16-langs`·`WT-hook-wiring-drift`·
  `WT-main-stamp-repair`·`WT-precommit-vet` — develop 미병합이라 축 1로는 못 지움.)
- **축 3 — 점유·미커밋**: 워크트리 점유 브랜치는 트리 폐기 선행. 프로세스 cwd 점유(lsof) 트리는
  제외. 트리 내 미커밋은 비강제 `git worktree remove`가 기계적으로 거부하는 네이티브 안전망으로
  상호 검증.

## 3. 단계별 실행 계획 (승인 후)

### Stage A — 기계적 정리 (축 1 와결, `-d` safe 모드)
- **트리 폐기 142**: `data/stageA-attached-final.txt` (병합∧트리점유 144 − 자기 트리 t264 − 활성 리드 트리 t430)
- **브랜치 삭제 256** = 트리점유 142 + 무트리 114 (`data/stageA-bare.txt`)
- 방식: `git worktree remove <path>` (비강제) → `git branch -d <n1> <n2> …` 배치 삭제.
  `-d`는 develop 기반 HEAD에 미병합이면 거부하므로 강제 플래그 불필요 — 판별식과 독립된 두 번째 안전망.
  실행은 본 워크트리 세션에서 pilot 1건으로 가드 판정을 확인한 뒤 일괄 진행
  (워크트리 세션 가드의 `git worktree remove <타경로>` 허용 여부 미확인 — primary 세션 경유가
  폴백: primary는 `git branch`가 차단되므로 트리 폐기만 primary에서 수행 가능).

### Stage B — 미병합 186건: 축 2 검증 후 삭제/보존
- **보존 확정**: `release/v3.1.4`·`release/v3.0.2-prep` (릴리즈 라인), 활성 점유 4
  (t223→`WT-agent-memory-drain`·t278→`WT-ci-flake-series`·t371→`WT-lint-shallow-clone`·
  t395→`WT-stale-backlog-json` — 전부 미병합 유일본), 보관 브랜치 10
  (`backup/*` 3, `save-*` 4, `stash-keep-*` 2, `GoosLab/main-fork` — 운영자 결정 항목)
- 나머지 ~170: 브랜치당 2명령(저작경로 추출 → 공집합 판정)으로 분류. 공집합 → 근거 로그와 함께
  `-D`. 비공집합 → 보존. 예상: 구세대 feat/chore/docs/sync PR 브랜치 대부분 squash 착지(공집합),
  일부 유일본 잔존.

### Stage C — 원격 정리 (실행 주체: 리드 — WT push 금지 규율, lane은 push하지 않음)
- **원격 삭제 후보 11** (현재 heads 교차 확정): `WT-card-landing-state`·`WT-codex-launcher`·
  `WT-freshness-sync`·`WT-gate-three-axes`·`WT-ossdocs-v311`·`WT-t69`·`WT-version-stamp-predicate`·
  `WT-wscfg-graph`·`WT-wscfg-worktree`·`worktree-agent-a06d06a2f644438e7`·
  `worktree-agent-ab699739d0b242c89` — 전부 develop 병합 확인.
- **원격 미병합 residue 19**: 축 2 검증 대상 (리드 지목 4건 포함).
- **stale tracking refs 66**: `git fetch --prune origin` (로컬 메타데이터 정리, 안전).
- 원격 삭제(`git push origin --delete`)는 리드 병합 창 경유로만 — lane 직접 push 없음.

### Stage D — 특수 클래스
- **고아 트리 디렉터리 3** (t336·t392·t413): 실행 시점 lsof 재측정으로 프로세스 소멸 확인 후
  `rm -rf`. 불확실하면 보존·보고.
- **release/* 전체**: 이 스윕 제외 (릴리즈 라이프사이클 소관).
- **보관 브랜치 10**: KEEP 기본 — 운영자가 명시적으로 정리 원할 때 별도 지시.

## 4. 승인 요청 사항 (리드 → 운영자)

1. Stage A 실행 — 트리 142 폐기 + 브랜치 256 삭제 (`-d` safe)
2. Stage B 실행 — 축 2 통과분만 `-D`, 비통과분 보존
3. Stage C 원격 11건 삭제 — 리드 창 경유
4. Stage D 고아 디렉터리 처리 — 프로세스 소멸 확인 조건부
5. 보관 브랜치 10건 운명 (KEEP 기본)

## 5. 미관측(Gaps)·잔여 위험

- **트리별 미커밋 상태 미측정**: 워크트리 세션 가드가 타 트리 `git -C`를 거부해 조사 시점
  직접 측정 불가. 대신 Stage A의 비강제 `worktree remove`가 더티 트리를 거부하는 네이티브
  안전망에 의탁 — 측정 갭을 안전망으로 보완하되 갭 자체는 명시.
- **축 2 판정 미수행**: 브랜치당 2명령 ≈ 340회 — 실행 시점 수행 (이번 조사 범위 밖).
- **lsof 점유 프로세스의 신원 미판별** (활성 세션 vs 좀비 셸): 실행 시점 재측정.
  t430은 확실한 활성 리드 트리 — 최종 제외 완료.
- **원격 삭제의 비가역성**: Stage C 실행 전 로컬에 동명 브랜치가 남아 있는지 대조 권장
  (로컬 병합셋/미병합셋에 동명 존재 여부 대조는 실행 시점 수행).
- 원데이터: `data/` 하위 (브랜치·워크트리·원격·조인 중간 산출물 전부 보존).
