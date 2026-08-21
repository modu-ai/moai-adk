# t159 — ci-autofix-protocol.md 두 판본 판정

베이스: `main` @ `4b2f203fe` · 워크트리 `.claude/worktrees/t159` · 브랜치 `WT-autofix-docs` · Tier S

**판정: (a)(b)(c) 중 어느 것도 아니다. 정리할 문제가 없다 — 카드의 핵심 전제가 반증됐다.**
두 판본은 완결된 설계 결정의 산물이고, §2.3 경로 충돌도 이미 해소돼 있다. 남은 것은 그 사실을 반영하지
못한 **문서 3곳**이었고, 그중 2곳을 고쳤다.

---

## 1. 카드 전제 반증 — "두 판본이 함께 로드된다"는 거짓

카드는 이렇게 적었다: *"로딩은 파일 존재로 결정되므로 두 판본이 함께 로드된다."*
**로딩은 파일 존재가 아니라 `paths:` 매칭으로 결정된다.** 이 리포 자신의 규정:

- `.claude/rules/moai/development/coding-standards.md:28` — *"path-scoped rules (`.claude/rules/` with
  `paths:` frontmatter) so it **loads only when matching files are touched**"*
- `.claude/rules/moai/development/rule-authoring.md:29` — always-loaded 표면을 *"a rule file **without**
  top-level `paths:`"* 로 **정의**한다. 즉 `paths:` 가 있으면 always-loaded 가 아니다.

두 판본의 `paths:` 는 **서로소**다:

| 판본 | `paths:` | 지배 대상 |
|---|---|---|
| (1) `.claude/rules/local/ci-autofix-protocol.md` 7,154B, **추적됨** | `.claude/skills/hns-workflow-ci-loop/SKILL.md` | dev 전용 스킬 + `scripts/ci-autofix/` (log-fetch.sh·classify.sh 를 전제조건 3·4로 명시) |
| (2)=(3) 배포판 7,037B (`moai/workflow/` 미추적 사본 = `internal/template/templates/` 추적본과 **byte-identical**) | `.claude/agents/moai/manager-develop.md`, `manager-develop-prompt-template.md`, `.github/workflows/**` | manager-develop `cycle_type=autofix` |

교집합이 0이므로 **둘을 동시에 매칭시키는 파일이 없다** → 중복 로드는 일어나지 않는다.
카드가 (c) 로 얻으려던 "중복 로드 제거"는 **이미 성립해 있다**. 갈라져 있지 않은 것은 범위가 아니라 이름뿐이고,
이름이 같은 것은 같은 독트린의 두 주체를 가리키므로 의도된 것이다.

## 2. 두 판본은 완결된 설계 결정이다 (미해결 상태가 아니다)

`CHANGELOG.md:340` (SPEC-CI-LOOP-DEVONLY-001 sync-phase close) 가 명시적으로 적어 둔 결정:

> **Intentional template/mirror asymmetry**: `ci-autofix-protocol.md` was rewritten in the template to
> drop its script dependency but preserved verbatim in the dev-repo mirror, while `ci-watch-protocol.md`
> stopped shipping entirely — one principle covers both, *a rule belongs only where the thing it governs
> exists*.

그 원칙 기준으로 두 판본 모두 **지배 대상이 실재**한다 — 실측:

- 로컬판: `scripts/ci-autofix/{classify.sh,log-fetch.sh}` + `test/` 2개 = 4파일 전부 추적됨,
  `.claude/skills/hns-workflow-ci-loop/SKILL.md` 9,334B 실재(템플릿에는 없음 = dev 전용 확인)
- 배포판: `manager-develop.md:51` 과 `manager-develop-prompt-template.md:20` 이 **정본으로 인용**하고,
  docs-site 4로케일 × 3페이지(`guides/ci-autonomy`, `workflow-commands/moai-sync`, `utility-commands/moai-e2e`)가
  `.claude/rules/moai/workflow/ci-autofix-protocol.md` 를 SSoT 로 지목한다

따라서 **(a) 는 살아 있는 dev 룰을 지우고**(scripts/ci-autofix 의 지배자가 사라진다 — 카드 자신도 이 위험을
적었다), **(b) 는 살아 있는 사용자 독트린을 철회한다**(인용 12곳이 끊긴다). 둘 다 완결된 결정을 되돌리는 일이다.

## 3. §2.3 경로 충돌도 이미 해소됐다

카드는 *"다음 update 가 템플릿판을 다시 깔면"* 을 위험으로 적었으나, 그 위험은 이미 제거돼 있다.
`ed04e40e6` (#1557, `chore(local): move local-only files out of the update-managed roots`) 이
**한 커밋 안에서** 로컬 원본을 `.claude/rules/local/` 로 옮기고 `.claude/rules/moai/workflow/` 의 추적본을
삭제했다(`git log --diff-filter=A` / `--diff-filter=D` 양쪽 모두 이 커밋을 가리킨다).

`.claude/rules/local/` 은 `CleanMoaiManagedPaths` 의 wipe 뿌리 밖이므로 로컬 원본은 더 이상 덮어써질 수 없다.
§2.3 이 처방한 "관리 대상 밖 배치"가 그대로 적용된 상태다.

## 4. 그래서 진짜 남아 있던 것 — 문서 3곳

정리할 파일이 아니라 **파일을 잘못 가리키는 문서**가 남아 있었다.

### 4.1 고침 — dev 스킬이 자기 쌍둥이가 아닌 배포판을 인용 (2건)

`.claude/skills/hns-workflow-ci-loop/SKILL.md` §Works Well With 가 두 줄 다 틀려 있었다:

- `.claude/rules/moai/workflow/ci-watch-protocol.md` → **죽은 경로**. `ci-watch-protocol.md` 는 템플릿에서
  아예 빠졌고 현재 트리에는 `.claude/rules/local/` 에만 있다(`find` 로 확인 — `moai/workflow/` 매치 3건은
  전부 #1557 이전에 만들어진 스테일 워크트리 사본이다)
- `.claude/rules/moai/workflow/ci-autofix-protocol.md` → **틀린 쌍둥이**. 이 스킬은 `scripts/ci-autofix` 를
  돌리는데, 배포판은 그 스크립트를 의도적으로 들어낸 판본이라 스킬의 실제 계약을 서술하지 않는다

둘 다 `.claude/rules/local/` 로 고치고, 왜 같은 이름의 다른 파일을 가리키면 안 되는지 4줄로 적었다.

### 4.2 고침 — 이 카드를 만들어 낸 스테일 한 줄

`CLAUDE.local.md:130` 이 *"어느 쪽을 살릴지 정리 필요"* 로 미해결을 선언하고 있었다. 이 줄이 t159 의
출처다. 해소된 상태(의도적 쌍둥이 / `paths:` 서로소 / #1557 로 경로 충돌 해소)로 교체했다.

### 4.3 고치지 않음 — 사고 기록은 보존

`CLAUDE.local.md:181` 의 *"`ci-autofix-protocol.md`(dev 원본 vs 템플릿 script-free판)가 그렇게 유실됐다"*
는 2026-08-15 사고의 **착지 시점 기록**이다. 현재 동작에 맞춰 고치면 기록이 거짓이 된다. 손대지 않았다.

## 5. 판정이 필요한 잔여 1건 (조치하지 않음 — 운영자/리드 몫)

배포판 사본이 `.claude/rules/moai/workflow/ci-autofix-protocol.md` 에 **미추적으로** 남아 있고,
`.claude/rules/` 전체에서 **유일한 미추적 파일**이다(`git status --porcelain .claude/rules/` 실측 1건).
`moai update` 가 매번 다시 깔므로 영구적인 `git status` 노이즈다. 세 갈래인데 셋 다 부작용이 있다:

| 선택 | 결과 |
|---|---|
| 그대로 둔다 | 노이즈 유지. 다음 감사가 다시 "미추적 이상"으로 집어 t160 을 연다(이 카드가 그렇게 태어났다) |
| `.gitignore` 에 추가 | 노이즈 제거. 다만 `.claude/rules/` 에 ignore 선례가 없어(패턴 0건) 새 관행이 된다 |
| 다시 추적 | #1557 이 **의도적으로** 지운 것을 되돌리는 것 — 데브 저장소 추적 트리에 배포 쌍둥이가 되살아난다 |

기술적 실측은 위가 전부이고, 어느 쪽이 관행상 맞는지는 저장소 소유자 판정이라 손대지 않았다.

## 6. Template-First — **이번 카드는 해당 없음** (실측 근거)

지시받은 항목이라 명시한다. 고친 두 파일 모두 템플릿 미러가 **없다**:

- `.claude/skills/hns-workflow-ci-loop/` → `internal/template/templates/.claude/skills/` 에 부재(실측).
  SPEC-CI-LOOP-DEVONLY-001 이 배포를 중단시킨 dev 전용 스킬이다
- `CLAUDE.local.md` → `internal/template/templates/CLAUDE.local.md` 부재(실측). 템플릿에 있는 것은
  별개 파일인 `CLAUDE.md` 다

따라서 미러 커밋도 `make build` 도 owed 되지 않는다. 배포되는 `ci-autofix-protocol.md` 는 **건드리지 않았다**.

## 7. 검증

| 검사 | 명령 | 결과 |
|---|---|---|
| 배포 사본 == 템플릿본 | `diff -q` | `IDENTICAL` (7,037B) — 배포판은 템플릿의 산출물이 맞음 |
| 로컬판 추적 여부 | `git ls-files --error-unmatch` | 추적됨 |
| 배포 사본 추적 여부 | `git status --porcelain` | `??` 미추적, `git check-ignore` 무매치 |
| `paths:` 서로소 | 두 파일 frontmatter 직접 읽음 | 교집합 0 |
| 지배 대상 실재 | `ls scripts/ci-autofix`, `ls SKILL.md`, agent/template 인용 grep | 양쪽 다 실재 |
| ci-watch 경로 실재 | `find -name ci-watch-protocol.md` | 현재 트리엔 `rules/local/` 뿐(=죽은 인용 확인) |
| 미러 부재 | `ls internal/template/templates/…` 2건 | 둘 다 부재 → Template-First 해당 없음 |
| 템플릿 중립성·parity | `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/... -run 'Leak\|Neutral\|Parity'` | `ok` |

## 8. 잔여 위험

- **동작 시험 0.** `paths:` 가 로딩을 제한한다는 판정의 근거는 이 리포의 규정 문서 2곳(coding-standards.md:28,
  rule-authoring.md:29)이지 로더 실행 관측이 아니다. 룰 로딩은 Claude Code 런타임 기능이라 이 저장소에
  구현이 없고, 두 판본을 실제로 매칭시켜 로드 여부를 눈으로 본 것은 아니다. **리포 규정이 런타임과 어긋나면
  판정도 어긋난다** — 그 경우 (c) 가 되살아난다.
- §5 잔여 1건을 그대로 두면 다음 미추적 감사가 같은 카드를 다시 연다. 이 verdict 가 그때 읽히도록
  §1·§2 를 근거와 함께 남겼다.
- 스테일 워크트리 3개(`web-shell-chrome`, `v31-m4-ko-content`, `release-v310`)가 `moai/workflow/` 경로에
  옛 `ci-watch-protocol.md` 를 들고 있다. 이 카드 범위 밖이고, 워크트리 폐기 시 함께 사라진다.
