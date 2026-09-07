# SPEC-CODEX-SKILL-DISABLE-001 — 구현 계획

카드: **t502** · Tier M · Class C · cycle_type **tdd**

## §A 맥락

`~/.codex/config.toml` 의 `[[skills.config]]` 는 이미 두 손이 닿아 있다 — 읽는 손(`internal/codexwiring/skills.go`, t451)과 지우는 손(`internal/cli/codex_skills_prune.go`, t506). 이 카드는 **쓰는 손 하나**를 더 만든다: 사용자가 이름으로 지목한 스킬 하나를 codex에서만 끄는 엔트리 발행기.

새 손이므로 결함 수정이 아니다. 그리고 사용자 HOME에 쓰는 손이므로, 안전 성질 셋(기본 비활성 · 명시 opt-in · 멱등 비파괴)이 기능 자체보다 먼저 온다.

## §B 미해결 질문 — 실행 전 결론이 필요

### 미러는 한 모양이 아니다 (측정된 전제)

먼저 전제를 바로잡는다. 초기 배차문은 미러를 「심링크」 하나로 서술했으나, `internal/template/skill_mirror.go` 의 `mirrorOneSkill`(:198-245)은 **네 갈래**로 끝난다.

| 결과 | 조건 | 미러 경로에 심링크가 있는가 |
|---|---|---|
| `MirrorModeSymlink` | 정상 경로. 이미 올바른 링크가 있는 멱등 재실행도 여기 | 있다 |
| `MirrorModeCopy` | `os.Symlink` 실패 시 명시적 폴백(권한 없는 Windows, 이색 파일시스템, 샌드박스) — 실제 사본 | **없다** (진짜 디렉터리) |
| `MirrorModeSkipped` | 심링크가 아닌 항목이 이미 그 자리를 차지(사용자 것일 수 있음). `Lstat`으로 구분해 손대지 않고 경고만 | 그 자리는 사용자 것 |
| `MirrorModeFailed` | 낡은 링크 제거 실패, 미러 디렉터리 생성 실패, 사본 실패 | 쓸 수 있는 미러 항목이 없다 |

그래서 `<root>/.agents/skills/<name>/SKILL.md` 는 앞 세 경우 모두 **해석되지만**, 경로 안에 심링크가 있는 것은 첫 번째뿐이다.

두 번째 측정 사실: **이 저장소에는 `.agents/skills/` 가 아예 없다** — t502 워크트리와 primary 체크아웃 양쪽에서 `ls` exit 1(직접 실행해 확인). 미러는 배포 실행이 만드는 산물이지 체크아웃이 만드는 것이 아니다. 따라서 「스킬 이름이 미러된 `SKILL.md` 로 항상 해석된다」는 전제는 성립하지 않고, **미러 부재는 평범한 상태**다 — 이것이 REQ-CSD-011(부재는 에러가 아니라 다른 실패와 구별되는 거절 사유)의 근거다.

### [NEEDS CLARIFICATION: Q-a — 게이트는 로드한 문자 그대로의 경로에 묶이는가, 심링크 해석된 경로에 묶이는가]

t504는 게이트를 **실제 디렉터리**(`$CODEX_HOME/skills/t504probe/SKILL.md`)에 대해서만 측정했다. 심링크가 경로에 끼어 있을 때 codex가 무엇과 비교하는지는 어디에도 측정돼 있지 않다. 후보 표기:

| 후보 표기 | 예 |
|---|---|
| 심링크 경로(문자 그대로) | `<project>/.agents/skills/<name>/SKILL.md` |
| 해석된 실경로 | `<project>/.claude/skills/<name>/SKILL.md` |
| codex home 컨벤션 | `~/.codex/skills/<name>/SKILL.md` |

틀리면 결과는 조용한 무효 엔트리 — 사용자는 껐다고 믿고 스킬은 계속 노출된다. 추측으로 기본값을 정하지 않는다.

### [NEEDS CLARIFICATION: Q-b — 미러 모드가 바뀌어도 같은 엔트리가 계속 묶이는가]

Q-a가 「해석한다」로 나오더라도 질문이 하나 더 남는다. 심링크 모드 프로젝트에서 발행한 엔트리가, 같은 프로젝트가 **사본 폴백(copy)** 으로 배포된 뒤에도 여전히 묶이는가 — 그리고 그 반대는. **한 모드에서 유효하고 다른 모드에서 관성인 경로 모양은 조용한 실패 표면**이고, 그것이 t504가 닫으려는 바로 그 부류다. REQ-CSD-013이 이 요구를 문장으로 들고 있다.

주의: 이것은 「사용자가 미러 모드를 바꾼다」는 가정이 아니다. 같은 이름의 스킬이 기계마다 다른 모드로 배포된다는 사실만으로 충분하다 — 발행된 엔트리는 사용자 계층(`~/.codex/config.toml`)에 남고 프로젝트 배포는 그와 무관하게 다시 돌기 때문이다.

**해소 방법(run-phase M1, 코드 작성 전)**: t504 양식대로 `CODEX_HOME` 격리 셀을 돌린다. 최소 셀 집합 —

| 셀 | 미러 모드 | 엔트리 path 표기 | `enabled` | 읽는 것 |
|---|---|---|---|---|
| S+ | symlink | 심링크 경로 | true | 양성 통제(로드되는지) |
| S-lit | symlink | 심링크 경로 | false | Q-a: 문자 그대로 묶이나 |
| S-res | symlink | 해석 실경로 | false | Q-a: 해석 후 묶이나 |
| C+ | copy | 사본 경로 | true | 양성 통제 |
| C-x | copy | S 셀에서 묶인 것과 **같은** 표기 | false | Q-b: 모드 교차에서도 묶이나 |

양성 통제가 노출되지 않는 셀이 하나라도 있으면 그 행렬은 채택하지 않는다(공허한 음성 방지). 어느 표기도 두 모드 모두에서 묶지 못하면, **이 기능은 미러 스킬에 대해 성립하지 않는다**는 판정이며 범위를 재협상해야 한다 — 그 경우 `$CODEX_HOME/skills/` 컨벤션으로 로드되는 스킬만 대상으로 좁히는 것이 후보다.

### [NEEDS CLARIFICATION: 이름 해석 후보 루트의 순서와 집합]

한 스킬 이름이 최대 세 곳에서 발견될 수 있다(Q-a 표). 어느 루트를 후보로 삼고 어떤 순서로 볼지는 M1 결과에 종속된다 — 게이트가 묶이지 않는 표기를 후보에 넣는 것은 무효 엔트리를 발행할 길을 열어두는 것이기 때문이다. M1 판정 이후 확정한다. 다중 매치 시 거절(REQ-CSD-011)이 기본 태도다.

## §C CLI 배치 결정

**결정: 새 루트 커맨드 `moai skills`, 하위 동사 `disable`, 계층 지정 플래그 `--codex` 를 필수로.**

```
moai skills disable <skill-name> --codex [--force]
```

`--codex` 를 필수로 두는 것이 곧 REQ-CSD-001(명시 opt-in)의 기계적 표현이다. 사용자의 HOME에 쓰는 순간, 그 쓰기 대상 계층을 사용자가 호출문에 적어 넣게 만든다.

### 검토한 대안과 기각 사유

| 안 | 내용 | 판정 |
|---|---|---|
| A. `moai codex skills disable` | 런처 `codexCmd` 에 동사 추가 | **기각.** `codexCmd` 는 `DisableFlagParsing: true` 인 런처이고(`internal/cli/codex_launcher.go:293` 부근), 동사 조회가 `runCodex` 안의 손수 만든 `codexVerbRouting` 맵 + `len(head) > 1` 이면 usage 실패라는 **한 토큰 형태**에 묶여 있다(`codex_launcher.go:344-368`). `disable <name>` 은 두 토큰이라 그 형태를 깨고, `--force` 도 손으로 파싱해야 한다. 게다가 진단 출력이 바이트 정확히 시험되고 있다(`SilenceErrors`/`SilenceUsage` 주석이 그 이유를 명시). 새 기능 하나를 위해 회귀 위험이 가장 높은 파일을 여는 거래 — **받지 않는다.** |
| B. `moai clean --codex-skills` 확장 | 쌍둥이와 같은 집 | **기각.** `clean` 의 계약은 **제거**다(`newCleanCmd` 는 스코프 하나만 받고 `--home` 과 상호배타). 비활성화는 제거가 아니라 **저작(엔트리 추가·갱신)** 이다. 같은 파일을 만진다는 이유로 반대 의미의 동사를 제거 커맨드에 넣으면, 다음 독자가 `clean` 이 파일에 내용을 **쓴다**는 것을 알 길이 없다. |
| C. `moai config ...` | 설정 커맨드에 태우기 | **기각.** `moai config` 는 moai 자신의 `.moai/config/` 계층을 다룬다. 사용자 HOME의 codex 파일을 여기에 넣으면 두 계층이 한 동사 뒤에서 섞인다. |
| D. 프로젝트 config 키(`codex.disabled_skills`) | 선언형 | **운영자가 이미 기각.** 프로젝트 설정이 사용자 HOME에 쓰기를 유발하면 「그 순간 사용자가 요청하지 않은 쓰기」가 되어 성질 (a)·(b)와 정면 충돌하고, 이 머신의 유령 49건과 같은 부채를 다시 만든다. |

### 배치 결정의 잔여 위험

`skills` 라는 루트 이름은 넓다 — 장래에 Claude Code 쪽 스킬 조작 동사가 생기면 같은 이름 아래로 들어오게 된다. 그것은 결함이 아니라 **의도한 자리**로 본다(계층은 언제나 플래그로 명시되므로 `moai skills disable X --codex` 와 가상의 `--claude` 가 한 지붕 아래 공존할 수 있다). 다만 이 카드는 `--codex` 하나만 구현하고, 다른 계층 플래그는 만들지 않는다.

## §D 기술 접근

### 파일 배치

| 파일 | 역할 |
|---|---|
| `internal/cli/codex_skills_disable.go` (신규) | 순수 병합 함수 + 판정 타입 + 러너 |
| `internal/cli/skills.go` (신규) | `moai skills` 루트 커맨드 + `disable` 하위 커맨드 배선 |
| `internal/cli/root.go` | `AddCommand` 한 줄 |
| `internal/codexwiring/skills.go` | **수정 없음** — 읽기 전용 유지(REQ-CSD-050) |
| `internal/cli/codex_skills_prune.go` | **수정 없음**(REQ-CSD-051) |

### 순수 함수 시그니처(형태만, 확정은 run-phase)

```
// content in → content out + 항목별 판정
func upsertCodexSkillDisable(content []byte, skillPath string) ([]byte, codexSkillDisableVerdict)
```

`codexwiring.SplitConfigLines` → `ParseSkillEntries` 로 기존 엔트리를 읽고, 경로 일치 엔트리가 있으면 그 구간의 `enabled` 줄만 교체, 없으면 말미에 두 키를 가진 엔트리를 덧붙인 뒤 `JoinConfigLines` 로 되돌린다. 이미 `false` 면 **입력 바이트를 그대로 반환**한다(prune의 `if !removed { return content, ... }` 와 같은 규율 — 왕복 결함이 no-op으로 위장하지 못하게).

### 러너

`runCleanCodexSkills` 의 골격을 그대로 따른다: `resolveCodexHomeDir()` → config 경로 조립 → 읽기 실패는 fail-open 안내 → dry-run 보고 → `--force` 시 선백업(`.bak-<UTC>`, 0600) + sha256 출력 → 쓰기 → 결과 보고.

## §E 마일스톤 (되돌리기 어려운 결정 먼저)

| # | 내용 | 산출 |
|---|---|---|
| **M1** | **게이트 경로 모양 측정 (Q-a + Q-b, AC-CSD-050)** — `CODEX_HOME` 격리 5셀(§B 표) + 양성 통제. 이 판정이 발행 경로 모양과 후보 루트 집합을 확정한다. 코드 작성 **전**에 끝낸다 | `.moai/reports/t502/gate-path-shape.md` |
| **M2** | 이름 → `SKILL.md` 해석기 (RED→GREEN). 후보 루트는 M1 판정에서 도출. 미해석·다중 매치는 거절 | `internal/cli/codex_skills_disable.go` 일부 + 테스트 |
| **M3** | 순수 병합 함수 — 신규 추가 / `enabled` 갱신 / 이미 false 시 바이트 동일 / 미인식 줄 보존 | 위 파일 + 테스트 |
| **M4** | CLI 배치 + 러너 — `moai skills disable --codex`, dry-run 기본, `--force`, 선백업, 항목별 보고 | `internal/cli/skills.go`, `root.go` 한 줄 |
| **M5** | 문서 — `moai doctor` 안내문과의 정합, README/도움말 문구. 기계적 마감 | 도움말 문자열 |

## §F 위험

| 위험 | 완화 |
|---|---|
| `enabled` 누락 엔트리 발행 → 사용자 codex 전면 장애 | 발행 경로를 함수 하나로 좁히고, 그 함수의 출력에 `enabled` 가 없으면 실패하는 테스트를 둔다(AC-CSD-001). 뮤턴트로 가드의 물성을 증명 |
| 경로 표기 오판(Q-a) → 조용한 무효 엔트리 | M1을 코드보다 앞에 둔다. 판정 전 어떤 경로 모양도 코드에 박지 않는다 |
| 미러 모드 교차(Q-b) → 한 기계에서만 유효한 엔트리 | M1 행렬에 copy 모드 셀을 포함(C+/C-x). 두 모드 모두에서 묶이는 표기가 없으면 범위를 좁힌다 |
| 사용자 손 편집 config 파괴 | 미인식 줄 보존(REQ-CSD-033), 선백업, dry-run 기본 |
| codex 버전 드리프트 — t504는 0.153.4 한정 | 판정서에 측정 버전을 스탬프하고, 발행 동사의 도움말에 「이 동작은 codex가 이 키를 어떻게 읽느냐에 달려 있다」는 사실을 남긴다 |
| 런처 회귀 | 배치 안 A 기각으로 회피 — `codex_launcher.go` 를 열지 않는다 |

## §G 안티패턴

- 이 기능을 「스킬 등록」·「프로비저닝」으로 서술하거나 그렇게 시험하는 것 — t504 V5/V7이 무력을 실측했다.
- 디렉터 모양 경로 발행 — D1d가 무효를 실측했다.
- prune과 코드를 공유하려고 `judgeCodexSkillEntry` 를 일반화하는 것 — 두 동사는 판정 기준이 반대다(prune은 부재를 증명해야 지우고, disable은 실존을 증명해야 쓴다).

## §H 교차 참조

- `.moai/reports/t504/skills-config-path-shape.md` — 채택 행렬
- `.moai/specs/SPEC-CODEX-GHOST-SKILLS-PRUNE-001/` — 쌍둥이 쓰기 경로
- `internal/codexwiring/skills.go` — 소비할 파서
