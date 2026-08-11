---
title: 자주 묻는 질문
weight: 100
draft: false
---

MoAI-ADK 를 처음 쓰면서 자주 겪는 혼동 포인트를 모았습니다. 대부분은 "두 가지를 헷갈렸다" 에서 출발합니다 — 터미널 명령과 슬래시 명령을 바꿔 쓰거나, statusline 표시의 의미를 잘못 읽거나, 설정 파일이 어디에 저장되는지 몰라 헤매는 식입니다. 이 페이지는 그런 첫 질문들을 한자리에 둡니다.

각 질문은 독립적으로 읽어도 이해되도록 답을 썼습니다. 위에서부터 순서대로 읽지 않아도 됩니다 — 필요한 질문으로 바로 건너뛰세요. 더 깊은 배경이 필요한 답에는 관련 페이지 링크를 달아 두었습니다.

MoAI-ADK 의 설계는 "세 가지 핵심(토크노믹스 · 에이전틱 루프 엔지니어링 · 에이전틱 하네스)이 하나의 계통 안에서만 작동한다" 는 전제 위에 서 있습니다. 이 아래의 질문들도 결국 이 세 가지 중 하나로 돌아갑니다 — 비용 통제는 토크노믹스의 결과이고, 설정 보존과 학습 축적은 루프 엔지니어링의 재료이며, 품질 게이트는 하네스가 책임지는 영역입니다.


---

## Q: `moai`와 `/moai`는 뭐가 다른가요?

완전히 다른 두 가지입니다. 가장 흔한 혼동이니 먼저 짚고 갑니다.

| | `moai` (터미널 CLI) | `/moai` (슬래시 서브커맨드) |
|---|---|---|
| **실행 위치** | 터미널 셸 | Claude Code 대화창 |
| **정체** | Go 바이너리 | Claude Code 스킬 호출 |
| **용도** | 프로젝트 설정, 템플릿 배포 | AI 에이전트 개발 워크플로우 |
| **예시** | `moai init my-project` | `/moai plan "인증 기능"` |

- 터미널에서 `moai plan`을 실행하면 동작하지 않습니다 — `/moai plan`은 Claude Code 안에서만 유효합니다.
- Claude Code에서 `/moai init`을 입력해도 동작하지 않습니다 — `moai init`은 터미널 명령입니다.

---

## Q: statusline의 버전 표시는 무엇을 의미하나요?

MoAI statusline은 버전 정보와 업데이트 알림을 함께 표시합니다:

```
🗿 v3.0.0 ⬆️ v3.0.1
```

- **`v3.0.0`**: 현재 설치된 버전
- **`⬆️ v3.0.1`**: 업데이트 가능한 새 버전

최신 버전을 사용 중일 때는 버전 번호만 표시됩니다:

```
🗿 v3.0.1
```

**업데이트 방법**: `moai update` 실행 시 업데이트 알림이 사라집니다.

{{< callout type="info" >}}
**참고**: Claude Code의 빌트인 버전 표시(`🔅 v2.1.172`)와는 다릅니다. MoAI 표시는 MoAI-ADK 버전을 추적하며, Claude Code는 자체 버전을 별도로 표시합니다.
{{< /callout >}}

---

## Q: statusline에 표시되는 세그먼트를 커스터마이징하려면?

statusline은 세그먼트 단위로 켜고 끌 수 있습니다. 각 세그먼트를 따로 토글해 원하는 정보만 남기세요. 디스플레이 프리셋은 따로 없고, 테마와 세그먼트 두 가지만으로 구성합니다.

`moai init` 또는 `moai update -c` 마법사에서 설정하거나, `.moai/config/sections/statusline.yaml`을 직접 편집합니다:

```yaml
statusline:
  segments:
    model: true
    context: true
    output_style: false
    directory: false
    git_status: true
    claude_version: false
    moai_version: false
    git_branch: true
```

`segments:` 블록이 없으면 기본적으로 모든 세그먼트가 활성화됩니다.

{{< callout type="info" >}}
자세한 내용은 [SPEC-STATUSLINE-001](https://github.com/modu-ai/moai-adk/blob/main/.moai/specs/SPEC-STATUSLINE-001/spec.md)을 참조하세요.
{{< /callout >}}

---

## Q: 모델 정책을 어떻게 선택하나요?

MoAI-ADK는 Claude Code 구독 요금제에 맞춰 에이전트에 최적의 AI 모델을 할당합니다. 요금제의 사용량 한도 안에서 품질을 최대한 끌어올리는 토크노믹스 장치입니다.

### 티어 비교

| 티어 | 특징 |
|------|------|
| **high** | 최고 품질 — 호출 빈도가 가장 낮은 두 에이전트에 `max` 추론 깊이 |
| **medium** (기본값) | 품질과 비용의 균형 — 비용/점수 곡선의 무릎 |
| **low** | 작업당 최저 비용 — 에이전틱 에이전트가 Opus `low` effort로 내려감 |

{{< callout type="warning" >}}
**왜 중요한가요?** 티어를 낮춘다는 것은 모델 클래스가 아니라 *추론 깊이*를 낮춘다는 뜻입니다. 오래 이어지는 에이전틱 작업에서는 Opus의 `low` effort가 어떤 effort의 Sonnet보다도 점수가 높고 작업당 비용도 낮습니다. 청구액을 좌우하는 것은 토큰당 단가가 아니라, 모델이 작업을 끝낼 때까지 밟은 스텝 수이기 때문입니다. 그래서 `low`는 Opus 안에서 아끼고, 여러 스텝을 밟다 실패할 일이 없는 단발성 행 (`manager-git`, `Explore`) 에서만 Sonnet을 씁니다.
{{< /callout >}}

### 티어별 에이전트 모델 배정

**11개 에이전트 카탈로그** (10 MoAI 커스텀 + 1 Anthropic 빌트인 `Explore`) 가운데 MoAI 커스텀 에이전트는 티어에 따라 모델이 정해집니다. 과거의 12개 보관 에이전트 (archived agents) 는 쓸 수 없습니다.

#### Manager Agents (5개)

| 에이전트 | high | medium | low |
|---------|------|--------|-----|
| manager-spec | opus / high | opus / medium | opus / low |
| manager-develop | opus / max | opus / medium | opus / low |
| manager-docs | opus / medium | opus / low | sonnet / low |
| manager-git | sonnet / low | sonnet / low | sonnet / low |
| manager-design | opus / high | opus / medium | opus / low |

#### Evaluator · Builder · Advisor · Specialist Agents (5개)

| 에이전트 | high | medium | low |
|---------|------|--------|-----|
| plan-auditor | opus / high | opus / medium | opus / low |
| sync-auditor | opus / high | opus / medium | opus / low |
| builder-harness | opus / high | opus / medium | opus / low |
| super-advisor | opus / max | opus / high | opus / medium |
| e2e-tester | opus / medium | opus / low | sonnet / low |

빌트인 `Explore`는 모든 열에서 `sonnet / low`로 해석됩니다. 디스크에 고정해 둘 에이전트 파일이 없어서, 호출하는 시점에 이 기본값이 적용됩니다.

### 설정 방법

```bash
# 프로젝트 초기화 시
moai init my-project          # 대화형 마법사에서 모델 정책 선택

# 기존 프로젝트 재설정
moai update -c                # 설정 마법사 재실행
```

{{< callout type="info" >}}
기본 티어는 `medium` 입니다. `moai update -c`로 설정 마법사를 다시 실행하여 변경할 수 있습니다.
{{< /callout >}}

---

## Q: "Allow external CLAUDE.md file imports?" 경고가 나타납니다

프로젝트를 열 때 Claude Code가 외부 파일 import 관련 보안 프롬프트를 띄울 수 있습니다:

```
External imports:
  /Users/<user>/.moai/config/sections/quality.yaml
  /Users/<user>/.moai/config/sections/user.yaml
  /Users/<user>/.moai/config/sections/language.yaml
```

{{< callout type="info" >}}
**권장 조치:** **"No, disable external imports"** 를 선택하세요.
{{< /callout >}}

**이유:**
- 프로젝트의 `.moai/config/sections/`에 이미 이 파일들이 존재합니다
- 프로젝트별 설정이 전역 설정보다 우선 적용됩니다
- 필수 설정은 이미 CLAUDE.md 텍스트에 포함되어 있습니다
- 외부 import를 꺼 두는 편이 더 안전하고, 기능에도 아무 영향이 없습니다

**파일 설명:**
- `quality.yaml`: TRUST 5 프레임워크 및 개발 방법론 설정
- `language.yaml`: 언어 설정 (대화, 코멘트, 커밋)
- `user.yaml`: 사용자 이름 (선택 사항, Co-Authored-By 표시용)

---

## Q: TDD와 DDD 방법론의 차이는 무엇인가요?

MoAI-ADK v2.5.0+는 방법론을 **TDD 또는 DDD 둘 중 하나**로만 고릅니다. 명확성과 일관성을 위해 하이브리드 모드는 없앴습니다.

TDD는 테스트를 먼저 쓰고 그 테스트를 통과시키는 순서라 신규 개발에 맞고, DDD는 기존 동작을 특성 테스트로 붙잡아 둔 뒤 조금씩 손보는 순서라 테스트가 거의 없는 코드에 맞습니다. 각 사이클의 단계별 절차는 [SPEC 기반 개발](/ko/core-concepts/spec-based-dev)과 [DDD](/ko/core-concepts/ddd)에서 다룹니다.

### 방법론 선택 표

| 프로젝트 상태 | 테스트 커버리지 | 권장 방법론 | 이유 |
|--------------|---------------|-------------|------|
| 신규 프로젝트 | N/A | TDD | 테스트 우선 개발 |
| 기존 프로젝트 | 50%+ | TDD | 테스트 기반이 있음 |
| 기존 프로젝트 | 10-49% | TDD | 테스트 확장 가능 |
| 기존 프로젝트 | < 10% | DDD | 점진적 특성 테스트 필요 |

### 설정 방법

```bash
# 프로젝트 초기화 시 자동 감지
moai init my-project          # --mode <ddd|tdd> 플래그로 지정 가능

# 수동 설정
# .moai/config/sections/quality.yaml 편집
development_mode: tdd         # 또는 ddd
```


---

## Q: 내 코드에 @MX 태그가 없는 이유는?

**아주 정상**입니다. @MX 태그 시스템은 AI가 먼저 살펴야 할 가장 위험하고 중요한 코드에만 표시를 남기도록 만들었습니다.

| 질문 | 답변 |
|------|------|
| 태그가 없으면 문제인가요? | **아닙니다.** 대부분의 코드에는 태그가 필요 없습니다. |
| 태그는 언제 추가되나요? | **높은 fan_in** (호출자 >= 3), **복잡한 로직** (복잡도 >= 15), **위험 패턴** (context 없는 고루틴) 에만 추가됩니다. |
| 모든 프로젝트가 비슷한가요? | **네.** 모든 프로젝트에서 대부분의 코드에는 태그가 없습니다. |

### 태그 우선순위

| 우선순위 | 조건 | 태그 유형 |
|---------|------|----------|
| **P1 (치명적)** | fan_in >= 3 | `@MX:ANCHOR` |
| **P2 (위험)** | 고루틴, 복잡도 >= 15 | `@MX:WARN` |
| **P3 (컨텍스트)** | 매직 상수, godoc 없음 | `@MX:NOTE` |
| **P4 (누락)** | 테스트 파일 없음 | `@MX:TODO` |

코드베이스를 @MX 태그로 스캔하려면:

```bash
/moai mx --all        # 전체 스캔
/moai mx --dry        # 미리보기
/moai mx --priority P1  # 치명적 항목만
```

---

## 더 많은 질문이 있으신가요?

- [GitHub Discussions](https://github.com/modu-ai/moai-adk/discussions) — 질문, 아이디어, 피드백
- [Issues](https://github.com/modu-ai/moai-adk/issues) — 버그 리포트, 기능 요청
- [Discord 커뮤니티](https://discord.gg/Z7E7Mdc5aN) — 실시간 소통, 팁 공유
