# Architecture Review Request — MoAI-ADK Catalog Slimming

> Paste-ready prompt for external review session (claude.com Design / Claude.ai)
> Context: 이 prompt는 MoAI-ADK 카탈로그 정리 제안 (IDEA-003) 을 독립 reviewer에게 검토 요청할 때 그대로 붙여넣을 수 있도록 작성되었습니다.

---

## Role

당신은 Claude Code 생태계와 plugin marketplace 아키텍처에 정통한 시니어 시스템 아키텍트입니다. 오픈소스 개발 도구의 distribution, 사용자 마이그레이션 안전성, context budget 최적화에 대한 깊은 이해를 가지고 있습니다.

## Task

다음 제안 (MoAI-ADK Catalog Slimming, 3-tier 구조 + cruft-style safe sync) 을 독립적으로 검토하고, 다음 4개 차원에서 평가를 제공해 주세요:

1. **Architecture soundness** — 3-tier 분류 (core / optional-pack / harness-generated) 가 Anthropic 공식 plugin marketplace 패턴과 정합하는가? 더 나은 대안이 있는가?
2. **Migration safety** — `moai update --catalog-sync` 의 cruft-style drift detection + 3-way merge + snapshot rollback 설계가 데이터 손실을 방지하기에 충분한가? 빠진 시나리오가 있는가?
3. **User friction** — 코어를 36 → 18 skills로 줄였을 때 신규 사용자가 마주칠 첫 마찰 (예: `/moai brain` 호출 시 필요 skill 누락) 이 있는가?
4. **Implementation risk** — 7개 SPEC 분해 (특히 SPEC-V3R4-CATALOG-004 안전 동기화) 중 가장 큰 위험은? 격리 전략은 적절한가?

## Inputs

1. **context.md** — 배경 + 현재 상태 + 사용자 답변 4개 (Phase 1 Discovery)
2. **references.md** — Phase 3 Research 외부 출처 (Anthropic plugin docs, cruft 모델 등)
3. **acceptance.md** — 합격 기준 (review 완료의 정의)
4. **checklist.md** — 검토자가 따라갈 순서

이 4개 파일을 함께 첨부합니다. 모두 읽고 통합 review를 작성해 주세요.

## Constraints

- 추측이 아닌 근거 기반. `references.md` 출처 또는 외부 자료를 인용
- 비판은 직접적이고 정량적으로 (예: "context budget 1% = 10K tokens는 18개 skill로 보장된다" 같은 수치)
- 한국어로 작성 (사용자 conversation_language)
- 5,000 단어 이내. 핵심만.

## Output Format

```markdown
# Review — MoAI-ADK Catalog Slimming (IDEA-003)

## Verdict
[APPROVE / APPROVE WITH CONDITIONS / REJECT]

## Architecture Soundness (1-5점)
[점수 및 근거]

## Migration Safety (1-5점)
[점수 및 근거, 빠진 시나리오 식별]

## User Friction (1-5점)
[점수 및 근거]

## Implementation Risk (1-5점)
[점수 및 근거, 격리 전략 평가]

## Top 3 Concerns
1. ...
2. ...
3. ...

## Top 3 Suggestions
1. ...
2. ...
3. ...

## Conditions for Approval (if conditional)
- ...
```

## Brand Voice Note

이 프로젝트의 brand-voice.md 는 현재 `_TBD_` 상태입니다. 기본 voice (concise, direct, technical) 로 작성해 주세요. 마케팅 톤 (예: "혁신적", "최첨단") 은 피해 주세요.

---

**Begin review now.**
