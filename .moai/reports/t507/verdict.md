# t507 — codex 플러그인 매니페스트 수용 형태 재측정 (codex-cli 0.153.4)

| 항목 | 값 |
|---|---|
| 카드 | t507 · Tier S · Class B · 재측정 전용 |
| 트리 | `.claude/worktrees/t507` · 브랜치 `WT-codex-plugin-remeasure` · HEAD `ace1c5440` (`origin/develop`과 델타 `0 0`) |
| 측정 대상 | **codex-cli 0.153.4** (`/Users/goos/.local/bin/codex --version`) |
| 측정 일자 | 2026-09-07 |
| 격리 | `CODEX_HOME=<scratchpad>/…` — 사용자 실제 `~/.codex`를 어떤 프로브에도 넘기지 않았다. 백업이 불필요했다(무접촉) |
| 원문 로그 | `.moai/reports/t507/logs/probe-raw.txt` · 픽스처 `.moai/reports/t507/fixtures/` |
| 범위 | 재측정까지. 패키징 구현·SPEC 작성 없음 |

> **이 문서의 판정 지위** — 아래 결론은 **권고**이고 판정이 아니다. 최종 판정은 리드가 재측정해서 낸다. 그 이유는 이 카드에 구조적 편향이 있기 때문이다: t494 §7에서 M2(플러그인 패키징)를 권고한 것이 나이고, 그 권고를 뒤집을 수 있는 측정을 같은 사람이 했다. 그래서 리드가 건 조건대로 **결론은 대조군 출력이 만들고**, 불리한 관측을 맨 앞에 둔다.

---

## 0. 먼저 — 내 t494 §7이 오해를 부를 수 있었던 지점 (불리한 관측)

t494 판정서 §7은 플러그인 매니페스트를 `.codex-plugin/plugin.json`으로 적었다. 그 자체는 문서 인용으로서 정확했다. 그러나 **마켓플레이스 루트를 그 경로로 만들면 거부된다**:

```
$ CODEX_HOME=<probe> codex plugin marketplace add <fixtures>/f1-codexplugin-mkt
Error: invalid marketplace file `…/f1-codexplugin-mkt`: marketplace root does not contain a supported manifest
EXIT=1
```

`.codex-plugin/plugin.json`을 마켓플레이스 루트로 둔 f4도 같은 거부다(EXIT=1). 즉 §7만 읽고 "`.codex-plugin/`에 매니페스트를 두면 된다"고 착수했다면 **첫 명령에서 막혔을 것**이다. §7은 플러그인과 마켓플레이스라는 두 층을 구분해 적지 않았고, 이 카드가 그 공백을 메운다.

**그리고 t90의 남은 논거는 이 측정으로 무너지지 않는다** — §5에서 따로 다룬다. 매니페스트 경로가 무엇이든 그 논거는 독립적으로 살아 있다.

---

## 1. 카드의 질문과 답

> **0.153.4가 실제로 받는 매니페스트 경로는 무엇인가.**

**두 층을 나눠 답해야 한다. 충돌하던 두 기록은 서로 다른 층을 재고 있었다.**

| 층 | 수용 | 거부 |
|---|---|---|
| **마켓플레이스 루트** | `.claude-plugin/marketplace.json` · **`.agents/plugins/marketplace.json`** | `.codex-plugin/marketplace.json` · `.codex-plugin/plugin.json` · `marketplace.json` · `codex-marketplace.json` · `.agents/marketplace.json` |
| **플러그인 매니페스트** (수용된 마켓플레이스 안) | **`.codex-plugin/plugin.json`** · `.claude-plugin/plugin.json` | (없음) → `Error: missing plugin.json` |

**그래서 t90과 t494는 둘 다 참이고, 애초에 충돌하지 않았다.**

- t90은 `.codex-plugin`을 **마켓플레이스 루트**로 재서 거부를 관측했다 — 0.153.4에서도 그대로 거부된다. t90의 실측은 **유지된다.**
- t494가 문서에서 읽은 `.codex-plugin/plugin.json`은 **플러그인 매니페스트**다 — 0.153.4에서 실제로 **수용된다.** t494의 문서 판독도 그 층에서는 옳았다.

겉보기 충돌의 정체는 버전 차이가 아니라 **층 혼동**이었다. 카드는 "버전이 달라 둘 다 참일 수 있다"고 적었는데, 실제로는 **버전과 무관하게 둘 다 참**이다.

---

## 2. 1단계 — 마켓플레이스 루트 매니페스트 (7형태)

명령 형태: `CODEX_HOME=<probe home> codex plugin marketplace add <fixture dir>` — 각 픽스처는 매니페스트 파일을 **정확히 1개만** 갖는다.

| # | 배치한 파일 | 결과 | EXIT |
|---|---|---|---|
| f1 | `.codex-plugin/marketplace.json` | `Error: … does not contain a supported manifest` | 1 |
| f2 | `.claude-plugin/marketplace.json` | `Added marketplace 't507probe' …` | **0** |
| f3 | `marketplace.json` (루트) | `Error: … does not contain a supported manifest` | 1 |
| f4 | `.codex-plugin/plugin.json` | `Error: … does not contain a supported manifest` | 1 |
| f5 | `codex-marketplace.json` | `Error: … does not contain a supported manifest` | 1 |
| f6 | `.agents/marketplace.json` | `Error: … does not contain a supported manifest` | 1 |
| f7 | **`.agents/plugins/marketplace.json`** | `Added marketplace 't507probe' …` | **0** |

**f1 대 f2가 깨끗한 대조쌍이다** — 내용이 byte 동일한 같은 `marketplace.json`을 디렉터리만 바꿔 넣었고, 결과가 갈렸다. 그러므로 갈림의 원인은 내용이 아니라 **경로**다.

### f7 — t90이 재지 않은 형태

t90은 `.agents/marketplace.json`(f6)을 쟀고 거부를 얻었다. 그러나 0.153.4 공식 문서가 지목하는 경로는 한 단계 더 깊은 **`.agents/plugins/marketplace.json`**이다:

> **Marketplace file locations** — Repo: `$REPO_ROOT/.agents/plugins/marketplace.json` · Personal: `~/.agents/plugins/marketplace.json`
> — <https://developers.openai.com/plugins/build/plugins> (t494에서 본문 확인)

이 형태는 **수용된다.** t90의 프로브에는 없던 칸이고, t90의 "`.codex-plugin/`은 어느 형태로도 인식되지 않는다"는 결론과 모순되지도 않는다 — `.agents/plugins/`는 `.codex-plugin/`이 아니다.

---

## 3. 2단계 — 플러그인 매니페스트 (3팔, 음성 대조군 포함)

마켓플레이스 파일을 `.agents/plugins/marketplace.json`으로 **고정**하고, 플러그인 자신의 매니페스트 위치만 바꿨다. 전 구간(`marketplace add` → `plugin list` → `plugin add`)을 돌렸다.

| 팔 | 플러그인 매니페스트 | `plugin add` 결과 | EXIT |
|---|---|---|---|
| g1 | `.codex-plugin/plugin.json` | `Added plugin 'probe-plugin' from marketplace 't507probe'.` | **0** |
| g2 | `.claude-plugin/plugin.json` | `Added plugin 'probe-plugin' from marketplace 't507probe'.` | **0** |
| **g3** | **없음** (음성 대조군) | `Error: missing plugin.json` | **1** |

### 음성 대조군이 여기서 하는 일

g1과 g2가 **둘 다 성공**했다. 이것만 보면 "둘 다 수용"과 "매니페스트를 아예 안 읽는다"를 구분할 수 없다 — 카드가 [HARD]로 경고한 바로 그 모호성이다.

g3이 그 모호성을 깬다. 매니페스트를 전부 치우자 설치가 **실패**하고(`Error: missing plugin.json`, EXIT=1), 그러므로 **매니페스트는 실제로 읽힌다.** g1·g2의 성공은 무시가 아니라 **측정된 수용**이다.

---

## 4. 3단계 — 설치가 실제로 배달하는 것

`plugin add`가 성공하면 플러그인 루트가 통째로 복사되고, **`skills/`가 함께 들어온다**:

```
$ find <probe home>/plugins/cache -type f
<plugin-root>/.codex-plugin/plugin.json
<plugin-root>/skills/probe-skill/SKILL.md
```

설치 후 `config.toml`에 남는 것은 두 테이블뿐이다:

```toml
[marketplaces.t507probe]
source_type = "local"
source = "<marketplace root>"

[plugins."probe-plugin@t507probe"]
enabled = true
```

t90의 E2b 관측(0.150.1)과 같은 모양이다 — 이 축은 세 판(0.150.1 → 0.153.4) 사이에 변하지 않았다.

---

## 5. t90의 남은 논거 재평가 (카드 [HARD] 요구)

> "플러그인 없이 이미 스킬·MCP 둘 다 닿으므로 패키징은 배포 채널이지 사용 가능성이 아니다."

**이 논거는 살아 있다. 이번 측정은 그것을 흔들지 못했고, 오히려 한 곳에서 강화한다.**

근거는 세 가지다.

1. **닿는 경로가 이미 있다.** moai는 `.agents/skills/<name>` 미러로 스킬을(`internal/template/skill_mirror.go:52`), `[mcp_servers.moai]` 등록으로 MCP를(`internal/codexwiring/configtoml.go:11-22`) 이미 codex에 노출한다. 두 경로 모두 플러그인 없이 동작하며, 이 카드는 그 사실을 뒤집는 관측을 하나도 내지 않았다.
2. **플러그인이 배달하는 것도 결국 같은 자산이다.** §4가 보인 대로 설치 결과물은 `skills/`가 담긴 디렉터리 복사다. **새 능력이 아니라 같은 능력의 다른 전달 방식**이다.
3. **오히려 논거가 강화되는 지점** — 플러그인 경로는 마켓플레이스라는 **추가 인프라**(리포 또는 로컬 루트 + `marketplace.json` + 버전 관리)를 요구한다. 지금의 직접 배선은 그것 없이 닿는다. 즉 패키징은 순비용을 먼저 지불하고 배포 편의를 얻는 거래다.

**따라서 이 카드의 결과는 "M2를 해도 된다"이지 "M2를 해야 한다"가 아니다.** 경로 판정(할 수 있는가)과 수요 판정(할 이유가 있는가)은 다른 질문이고, 이 카드는 앞쪽만 답했다. t90이 건 수요 게이트는 여전히 열려 있으며 **그 판단은 운영자 것**이다.

한 가지만 갱신을 권한다: t90 카드 본문의 `.codex-plugin/plugin.json` 서술은 **마켓플레이스 층에서는 여전히 틀렸고 플러그인 층에서는 옳다**. 카드를 언젠가 다시 연다면 본문을 그 두 층으로 나눠 적어야 한다 — t90 판정서가 "본문을 고쳐야 한다"고 적은 지적은 유효하되, 고칠 방향이 "`.claude-plugin`으로 바꿔라"가 아니라 "층을 나눠라"이다.

---

## 6. 권고 (판정 아님)

| # | 권고 | 근거 |
|---|---|---|
| R1 | **M2를 착수한다면 마켓플레이스는 `.agents/plugins/marketplace.json`으로 만든다.** `.claude-plugin/`도 수용되지만 그것은 Claude 규약이고, `.agents/plugins/`가 codex 공식 문서가 지목하는 자리다 | §2 f7 · f2 |
| R2 | **플러그인 자신의 매니페스트는 `.codex-plugin/plugin.json`으로 둔다.** 둘 다 수용되므로 선택은 규약 문제이고, codex 문서의 형태를 따르는 쪽이 낫다 | §3 g1 · g2 |
| R3 | **M2 착수 여부는 이 측정으로 결정되지 않는다.** 수요 게이트는 t90이 걸었고 아직 운영자 판단이다 | §5 |
| R4 | t494 판정서 §7에 **층 구분 한 줄을 덧붙이는 것**을 권한다. 원문을 고치는 것이 아니라 정정을 나란히 붙이는 형태로 — 지금 §7만 읽으면 `.codex-plugin/`에 마켓플레이스를 만들려다 막힌다 | §0 |

---

## 7. 판정서 (Claim / Evidence / Baseline / Gaps / Residual-risk)

### Claim
1. 0.153.4에서 마켓플레이스 루트로 수용되는 형태는 **`.claude-plugin/marketplace.json`과 `.agents/plugins/marketplace.json` 둘**이다. `.codex-plugin/`은 어느 형태로도 거부된다.
2. 0.153.4에서 플러그인 매니페스트로 수용되는 형태는 **`.codex-plugin/plugin.json`과 `.claude-plugin/plugin.json` 둘**이다.
3. 그 수용은 공허하지 않다 — 매니페스트를 치우면 설치가 실패한다.
4. **t90과 t494는 서로 다른 층을 재고 있었고, 둘 다 참이다.** 충돌은 버전 차이가 아니라 층 혼동이었다.
5. t90이 재지 않은 `.agents/plugins/marketplace.json`이 수용된다 — t90 프로브의 공백이었다.
6. 설치는 플러그인 루트를 통째로 복사하며 `skills/`가 따라온다.
7. **t90의 "패키징은 배포 채널이지 사용 가능성이 아니다"는 논거는 이 측정 뒤에도 유지된다.**

### Evidence
전 명령·출력·종료코드가 `.moai/reports/t507/logs/probe-raw.txt`에 원문 그대로 있다. 본문 §2·§3·§4가 그 요약이며, 픽스처는 `.moai/reports/t507/fixtures/`에 커밋돼 재현 가능하다.
대조 설계 2건: **f1↔f2**(내용 동일·경로만 다름 → 갈림의 원인이 경로임을 고정) · **g3 음성 대조군**(매니페스트 제거 → 실패 확인, g1·g2의 성공이 공허하지 않음을 고정).

**픽스처 커밋에 관한 주의 [기록]**: `.agents/` 하위 픽스처 5개는 `.gitignore:133`(`.agents/` — 스킬 미러용 규칙)에 걸려 일반 `git add`로는 커밋되지 않았다. `git add -f`로 강제 추가했다. 이름을 바꾸는 대신 강제 추가를 택한 이유는 **경로 자체가 이 실험의 시험 변수**이기 때문이다 — `.agents/plugins/`를 다른 이름으로 바꾸면 f7이 재현하는 대상이 더 이상 f7이 아니다. 다만 이 5개는 리포의 무시 규칙과 충돌하는 상태로 남으므로, 훗날 `.agents/` 관련 도구가 이 경로를 스캔 대상으로 오인할 여지가 있다(관측된 문제는 아니고, 남겨 두는 주의사항이다).

### Baseline-attribution
`codex --version` → `codex-cli 0.153.4`. 트리 `.claude/worktrees/t507` HEAD `ace1c5440`, `origin/develop`과 델타 `0 0`. 모든 프로브는 `CODEX_HOME` 격리 하에 이번 런에서 실행했다. t90 값(0.150.1)은 인용이며 이번 런에서 재실행하지 않았다 — 비교 대상이지 내 측정이 아니다.

### Gaps — 관측하지 않은 것
1. **`~/.agents/plugins/marketplace.json`(사용자 홈 스코프)**는 재지 않았다. 리포 스코프만 쟀다.
2. **Git 마켓플레이스**(`codex plugin marketplace add owner/repo`)는 재지 않았다. 로컬 루트만 쟀다.
3. **설치된 플러그인의 스킬이 실제 세션에서 로드되는지**는 재지 않았다. 파일이 복사된 것까지만 봤다 — 로드 여부는 별개 질문이다.
4. **`.mcp.json` / `hooks/hooks.json` 번들**은 재지 않았다. 플러그인이 MCP와 훅도 같이 나른다는 문서 서술은 이번에 실증되지 않았다.
5. **`plugin.json`의 필드 검증 수준**은 재지 않았다. 파일 존재만 확인됐고, 필수 필드를 비웠을 때의 동작은 모른다.
6. **`.claude-plugin`과 `.agents/plugins`가 항구 계약인지 과도기 호환인지** — t90이 남긴 같은 미확인이 그대로 남는다. 문서는 `.agents/plugins`만 적고 `.claude-plugin` 수용은 언급하지 않는다.
7. **t90 판정서는 untracked 파일**(`.moai/reports/t90/verdict.md`, primary 체크아웃)이라 이 브랜치에 없다. 인용은 그 사본을 읽은 것이고, 이 트리에서 재현되지 않는다.

### Residual-risk
- codex-cli는 빠르게 움직인다. t90 본문이 "M0 시점 `plugin_hooks` 제거가 이미 한 번 전제를 뒤집었다"고 적었고, 이번 카드가 그 경고대로 재측정한 것이다. **0.153.4 이후 판에서 또 바뀔 수 있다** — 착수 시점에 다시 재야 한다.
- `.claude-plugin` 수용이 문서에 없다는 것은 **그것이 비공식 호환일 가능성**을 남긴다. 그래서 R1은 문서에 적힌 `.agents/plugins/` 쪽을 권한다.
- 이 측정은 이 머신 한 대, macOS(darwin 25.6.0)에서 났다. 다른 OS의 경로 해석은 확인하지 않았다.
