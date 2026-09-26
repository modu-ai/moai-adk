# moai update와 로컬 전용 파일 — 관리 대상 삭제의 실측과 생존 규칙

> CLAUDE.local.md §2.3 에서 이관했다(card t750, 2026-09-14). **로컬 전용 문서** — 템플릿에 미러하지 않는다(내부 카드 id·SPEC id·내부 날짜를 포함하므로 템플릿 중립성 §25 의 금지 클래스에 해당). 이 문서의 정본은 develop 트리의 이 사본이다(CLAUDE.local.md §0.1 판별식 준용) — 갱신은 카드 워크트리에서 한 뒤 develop 으로 병합한다.
> CLAUDE.local.md 쪽에는 이 요지만 남는다: `CleanMoaiManagedPaths` 가 관리 대상 뿌리를 통째로 삭제하며 보호 목록이 없다는 사실, 로컬 전용 파일 배치 [HARD], update 후 검증·git-strategy 재적용 [HARD].

---

**위 Local-Only 목록에 적혀 있다는 사실만으로는 파일이 보호되지 않는다.** 2026-08-15 `moai update --yes` 실행에서 이 목록에 명시된 파일 12개가 실제로 삭제됐다 — 목록은 사람과 AI가 읽는 문서일 뿐, 삭제를 수행하는 Go 코드는 이 파일을 읽지 않는다.

**삭제 주체**: `CleanMoaiManagedPaths` (`internal/cli/update/deploy/deploy.go:107`; [2026-08-27 감사 정정]). diff 기반이 아니다. 템플릿 재배포 **전에** 아래 뿌리를 `os.RemoveAll`로 **통째 삭제**한 뒤 임베드 템플릿에 있는 것만 다시 깐다. 템플릿에 대응 파일이 없으면 복구되지 않는다.

```
.claude/settings.json      .claude/commands/moai     .claude/agents/moai
.claude/skills/moai*(글롭) .claude/rules/moai         .claude/output-styles/moai
.claude/hooks/moai         .moai/config              (deploy.go:187-190; [2026-08-27 감사 정정])
```

`.moai-skip-cleanup` 마커는 이 함수가 **참조하지 않는다**(참조 0회 — v2→v3 clean-reinstall 경로 전용). 사용자가 편집할 수 있는 보호 목록 설정은 **존재하지 않는다**.

**[HARD] 규율 — 로컬 전용 파일을 위 뿌리 안에 두지 않는다.** 새 로컬 전용 파일을 만들 때 위치부터 정한다:

| 용도 | 금지 (삭제됨) | 안전 |
|---|---|---|
| 룰 | `.claude/rules/moai/**` | `.claude/rules/` 하위의 **비-moai** 디렉터리 |
| 스킬 | `.claude/skills/moai*` | `moai`로 **시작하지 않는** 이름 (글롭 회피) |
| ast-grep 룰셋 | `.moai/config/astgrep-rules/` | `.moai/` 하위 **`config/` 밖** + `gate.yaml`의 `ast_grep_gate.rules_dir` 지정 (빈 값이면 기본 경로 폴백 없음 — t50, `internal/cli/astgrep.go:69,105`; [2026-08-27 감사 정정]) |
| 하네스 | — | `.claude/skills/hns-*`, `.claude/agents/harness/`, `.claude/commands/harness/`, `.moai/harness/` (`IsUserOwnedNamespace` 백업 대상) |

**[HARD] update 실행 후 매번 검증한다.** 전제: 실행 **전** 추적 파일 수정이 **§0.4 가 규정한 ` M CLAUDE.local.md` 한 건뿐**이어야 diff 귀속이 가능하다. primary 체크아웃에서 그 표식은 영구적이므로 **0 이 되는 일은 없다** — 0 을 전제로 읽고 그 한 건을 없애려 들면 §0.4 가 막은 회귀로 떠밀린다. 다른 파일이 함께 수정돼 있으면 그때는 귀속이 불가능하니, update 전에 그쪽을 먼저 정리한다.

```bash
git status --porcelain | grep -v '^??' | wc -l        # 실제 변경 수 — primary 의 baseline 은 0 이 아니라 1(§0.4)
git status --porcelain | grep '^ D'                   # 삭제된 파일 — 0이어야 정상
# 삭제가 있으면 (전부 추적 파일이므로 git이 안전망):
git status --porcelain | grep '^ D' | sed 's/^...//' | tr '\n' '\0' | xargs -0 git restore --
```

**[HARD] update 후 `git-strategy.yaml`의 git-flow 키를 반드시 재적용한다.** `.moai/config`는 위 wipe 대상이므로, `moai update` 는 `.moai/config/sections/git-strategy.yaml` 을 템플릿 기본값(`workflow: github-flow`, develop/release 키 없음)으로 되돌린다. 이 파일은 **템플릿에 미러하지 않는다** — 미러하면 16개 언어 배포판 전체에 이 프로젝트의 사설 워크플로가 실려 나간다(§15). 그러니 매 update 후 로컬에서 다시 넣는다:

```bash
# 확인 — 두 키를 함께 본다. 하나만 보면 나머지가 되돌아간 것을 놓친다 (card t1159)
grep -n 'workflow: git-flow' .moai/config/sections/git-strategy.yaml || echo 'REVERTED(workflow) — 재적용 필요'
grep -n 'worktree_base_branch: develop' .moai/config/sections/git-strategy.yaml || echo 'REVERTED(worktree_base_branch) — 재적용 필요'
# 재적용 — **--source=develop 이다. HEAD 가 아니다** (card t1159)
git restore --source=develop -- .moai/config/sections/git-strategy.yaml
```

**[HARD] `--source=HEAD` 를 쓰지 않는다.** primary 체크아웃은 `main` 에 체크아웃돼 있고, `main` 커밋본에는 `worktree_base_branch` 키가 **아예 없으며** 세 블록 모두 `workflow: github-flow` 다(2026-09-24 실측: `git show main:.moai/config/sections/git-strategy.yaml`). `HEAD` 에서 복원하면 두 키가 함께 되돌아간다 — 확인 grep 은 실패하는데 복원은 고쳐 주지 않는 순환이 된다. 정본은 `develop` 이다(§0.1 과 같은 판별식: 레인이 분기하는 트리가 지배한다).

`git restore` 가 통하지 않는 상황(커밋 전 상태)이면 `git_strategy` 아래를 손으로 되돌린다 — **네 줄이 아니라 다섯 줄이다**:

- `git_strategy.manual`: `workflow: git-flow` [2026-08-27 감사 정정], 그리고 `main_branch:` 바로 아래에 `develop_branch: develop` / `release_branch_prefix: release/` / `rc_version_format: vX.Y.Z-rc.N` 세 줄.
- `git_strategy` 최상위: `worktree_base_branch: develop` — **이 줄이 목록에서 빠져 있어 2026-09-24 에 카드 트리 6개가 develop 이 아니라 main 에서 났다**(t1154·t1153·t1075·t1157·t1158·t1159). 빈 값은 `SPEC-WORKTREE-BASEREF-001` 의 중립 기본값이라 `moai worktree new` 가 base 오퍼랜드 없이 `git worktree add` 를 돌리고, git 은 호출 트리의 HEAD(= primary 의 `main`)에서 판다. 손실이 조용하다 — 확인 grep 이 `workflow` 만 보면 이 되돌림은 통과한다. 근거: `.moai/reports/t1159/measurement.md`.

**[HARD] AC 스냅숏 커밋 가드의 무장 상태는 세션 시작마다 읽는다 (card t1161).** t1150 이
넣은 가드(`git config hook.ac-baseline-guard.{event,command}` + `scripts/ac-baseline/check-staged.sh`)는
**꺼져도 조용하다** — 키 삭제·git < 2.54·체커 부재 세 경우 모두 커밋이 그냥 통과하고, 그 통과는
"가드가 돌아서 아무것도 못 찾았다"와 출력이 같다. 침묵이 정보를 담지 않으므로 말해 주는 표면이
따로 있어야 한다:

```bash
# 읽기 전용. 무장이면 침묵(stdout `ARMED`), 아니면 고장마다 stderr 1줄. 항상 exit 0
sh scripts/ac-baseline/check-armed.sh
```

배선: `.claude/settings.local.json` 의 `SessionStart` 에 항목 하나를 더한다(§28 의 lsel 2항목과
같은 자리·같은 이유 — tracked `settings.json` 은 `moai update` 가 통째 재배포해 배선이 매번
유실된다). **`moai doctor` 에는 넣지 않는다**: doctor 체크는 Go 제품 코드라 전 사용자 배포판에
실려 나가고, 이 가드는 이 저장소 전용 도구다(템플릿 미러 없음).

```jsonc
// .claude/settings.local.json  .hooks.SessionStart[0].hooks[] 에 추가
{ "type": "command", "command": "bash", "timeout": 30,
  "args": ["-c", "[ -f \"$0\" ] && exec sh \"$0\"; exit 0",
           "${CLAUDE_PROJECT_DIR}/scripts/ac-baseline/check-armed.sh"] }
```

`scripts/` 는 관리 대상 뿌리 밖이므로 이 스크립트는 `moai update` 가 지우지 않는다 — 재적용
대상이 아니다. 배선만 settings.local.json 에 있고, 그 파일은 런타임이 쓰는 로컬 전용이다.

**[HARD] 보고된 파일 수를 믿지 않는다.** `Updated N files`의 N은 관리 대상 뿌리 **밖** 파일만 센다(`internal/cli/update/plan/plan.go:73` `if IsMoaiManaged(...) { continue }`). 2026-08-15 실측: 보고 32, 실제 175. **삭제는 이 요약에 전혀 나타나지 않는다.**

**삭제만이 손실이 아니다 — 덮어쓰기 2종** (2026-08-15 실측):

- **경로 충돌**: 템플릿과 **내용이 달라야 하는** 로컬 파일이 템플릿 경로에 있으면, 삭제가 아니라 템플릿판으로 **덮어써진다**. `ci-autofix-protocol.md`(dev 원본 vs 템플릿 script-free판)가 그렇게 유실됐다. 파일이 남아 있어 삭제 검사(`git status | grep '^ D'`)로는 안 잡힌다 — **내용이 달라야 하는 파일은 반드시 관리 대상 밖에 둔다.**
- **`.sh` / `.sh.tmpl` 쌍 드리프트**: 배포되는 것은 `.tmpl` 쪽인데 편집이 `.sh` 쪽에만 들어가면, update가 배포본을 **구버전으로 되돌린다**. 실측: `handle-{agent-hook,task-completed,teammate-idle,stop-goal}.sh` 4쌍에서 `.sh`에만 있던 SPEC-STOPCHAIN-TRIM-001 가드(41줄)가 사라짐. `sync-phase-quality-gate.sh`는 `.tmpl`이 없어 무사. **훅 래퍼를 고칠 때는 `.sh`와 `.sh.tmpl`을 함께 고친다.**
  ```bash
  # 쌍 존재 확인 + 드리프트 점검
  for f in internal/template/templates/.claude/hooks/moai/*.tmpl; do b=${f%.tmpl}; [ -f "$b" ] && diff -q "$b" "$f" >/dev/null || echo "DRIFT $(basename $b)"; done
  ```

**참고 — 관련 결함 3건** (별도 카드 소관): ① `CleanMoaiManagedPaths`에 보호 목록 부재(근본 원인) ② `archiveLegacySkills`가 wipe **이후**(`internal/cli/update.go:554`; [2026-08-27 감사 정정])에 호출돼 원본이 이미 없어 `0 archived`로 조용히 통과 ③ `--dry-run`이 `CleanMoaiManagedPaths` 삭제 예정 목록을 미리보기하지 않음.

> **§2.2 astgrep-rules 로컬 전용 예외 (2026-07-02)**: 로컬 `.moai/config/astgrep-rules/`의 언어별 서브디렉터리 트리 + `sgconfig.yml`은 dogfood-experimental(10/17 빈 `.gitkeep` stub, 나머지는 데모성 스캐폴드, 메시지 언어 혼재 ko/en, `sgconfig.yml`이 존재하지 않는 `utils` ruleDir 참조 + SPEC-ID 포함)이라 템플릿에 미러하지 않는다. 배포 사용자는 template-managed `go-hardcoding.yml`(root, SPEC-ID stripped) 1개를 baseline으로 받는다. **(2026-08-01 정정)** 종전 이 절은 "`gate.yaml`/`gate` 로더 부재 → `AstGrepGate.Enabled` 항상 컴파일 기본값 false"라고 적었으나 **두 주장 모두 현재 main에서 거짓**이다: 로더는 `internal/config/loader_gate.go` 에 존재하며 `internal/config/loader.go` 의 `Loader.Load` 에서 호출된다(SPEC-CONFIG-AUDIT-REPAIR-001 M2, PR #1142). 실제 기본값은 `internal/config/defaults.go` 의 `AstGrepGate{Enabled: true, BlockOnError: false, WarnOnlyMode: true}` — 즉 **차단 없는 권고 모드로 켜져 있고**, 차단(blocking)만이 `gate.yaml` opt-in이다. 단, **릴리스된 `v3.0.1` 에는 로더가 없어**(`loadGateSection` 호출 0회) 해당 버전 사용자에게는 종전 서술이 여전히 사실이다 — 이것이 이슈 #1265 의 내용이며, 해결책은 코드 수정이 아니라 릴리스다. 16-언어 정식 룰셋 배포(메시지 영어 통일 + 데모 stub → 실제 패턴 + `utils` 정리 + SPEC-ID strip + `sg` config-mode 검증)는 별도 후속 SPEC 소관.

