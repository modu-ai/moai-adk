---
title: 내비게이터 바인딩 토큰
weight: 25
draft: false
---
# 내비게이터 바인딩 토큰

코드와 문서가 서로를 가리키게 만들면, 에이전트가 한쪽을 고쳤을 때 다른 쪽의 맥락을 곧바로 끌어올 수 있습니다. **내비게이터 바인딩 토큰** (Navigator Binding Tokens) 은 설계 결정·코드 심볼·SPEC 이 하나의 주소 가능한 그래프로 이어지도록 짓는 작성용 토큰 세 개입니다. 이 토큰들이 모여 `.moai/project/navigator/nav-graph.json` 이라는 단일 산출물을 만듭니다.

## 토큰 세 개

내비게이터 통합 계층은 세 가지 바인딩 토큰 집안을 하나의 그래프로 합칩니다.

| 집안 | 토큰 형태 | 작성 자리 | 역할 |
|------|----------|-----------|------|
| `NAV:DEC` | `@NAV:DEC-<id>` | 설계 문서 (`.moai/project/*.md`, `.moai/docs/**/*.md`) | 설계 결정을 SPEC이나 심볼에 연결 |
| `NAV:SYM` | `@NAV:SYM:<symbol>` | 코드 주석 + 설계 문서 | 문서 자리를 코드의 이름 붙은 심볼에 연결 |
| `MX:SPEC` | `@MX:SPEC:<SPEC-ID>` | 코드 주석 (`@MX:` 태그의 하위 줄) | 코드 자리를 SPEC에 연결 |

`MX:SPEC` 은 이미 [MX 태그 시스템](/ko/advanced/mx-tags/) 이 다루고 있습니다. 내비게이터 계층은 MX 스캐너의 `SpecAssociator` 출력을 **소비** 할 뿐 다시 스캔하지 않습니다. 그러니 이 토큰을 새로 짓지 말고 기존 MX 태그 규칙을 따르세요.

## 토큰을 언제 작성하는가

### `@NAV:DEC-<id>` 를 작성할 때

- `.moai/project/tech.md`, `structure.md`, `product.md` 나 `.moai/docs/` 아래 설계 문서의 한 결정이 특정 SPEC이나 코드 심볼에 대응할 때.
- 나중에 코드를 고칠 때 그 결정의 맥락이 다시 떠오르기를 바랄 때.

### `@NAV:SYM:<symbol>` 를 작성할 때

- 문서 자리나 코드 주석이 이름 붙은 코드 심볼에 묶여야, 그래프를 읽는 사람이 문서에서 코드로 (혹은 심볼에서 심볼로) 이동할 수 있을 때.

`@MX:SPEC:` 은 이곳에서 짓지 않습니다. 이미 mx-scanner 표면입니다. 다시 짓는 것은 불필요합니다.

## 토큰 문법

두 토큰 모두 빈 값이 들어가면 안 됩니다. 스캐너는 빈 값을 만나면 진단 경고를 `.moai/logs/navigator-sync.log` 에 남기고 그 항목을 건너뛰되, 그래프 빌드 전체를 중단하지는 않습니다 (fail-open).

### `@NAV:DEC-<id>`

`<id>` 는 `[A-Z][A-Z0-9-]*` 여야 합니다. 대문자 ASCII 와 숫자, 그리고 내부 하이픈만 허용됩니다. SPEC-ID 도메인 토큰과 일관된 규칙입니다. `@NAV:DEC-` 접두사가 모호하지 않은 판별자이므로, id 자체는 접두사 없이 등장하지 않습니다.

### `@NAV:SYM:<symbol>`

`<symbol>` 은 `[A-Za-z_][A-Za-z0-9_.]*` 여야 합니다. 식별자 모양이면 되고, 언어 중립적입니다. 패키지 한정형 (`pkg.ParseHeader`) 이 관례이고, 짧은형 (`ParseHeader`) 도 받아들여 기존 심볼 집합에 대한 접미 일치로 해결합니다.

## 스캔 뿌리

내비게이터 통합 계층은 다음 표면을 스캔합니다.

- **설계 문서** — `.moai/project/{product,structure,tech}.md` 와 `.moai/docs/**/*.md`.
- **코드** (`@NAV:SYM` 만) — `*_test.go` 와 `vendor/` 를 제외한 Go `*.go` 파일. 설계 문서 표면도 함께.

다음은 스캔하지 **않습니다**.

- `.moai/specs/` — 이미 mx-scanner가 본문 기반 연관으로 덮습니다.
- `.moai/reports/`, `.moai/state/` — 일시적이거나 실행 시점 상태.
- 기존 세 내비게이터 체인의 소스 코드 (소비 전용).

## 산출물 — `nav-graph.json`

`.moai/project/navigator/nav-graph.json` 하나의 파일로 떨어집니다. 모양은 다음과 같습니다.

```json
{
  "provenance": { "extract_commit_sha": "...", "captured_at": "..." },
  "nodes": [
    { "entity_type": "decision", "identifier": "...", "display_name": "..." }
  ],
  "edges": [
    { "edge_type": "dec-edge", "source_node": "...", "target_node": "...", "source_path": "...", "line_number": 0 }
  ]
}
```

`entity_type` 은 `decision | spec | symbol` 셋 중 하나이고, `edge_type` 은 `dec-edge | spec-edge | sym-edge` 셋 중 하나입니다.

이 산출물은 **바이트 안정적**입니다. 같은 git HEAD 에서 두 번 돌리면 바이트 단위로 같은 결과가 나옵니다. 벽시간 타임스탬프를 찍지 않기 때문에, 누가 언제 돌렸는지와 무관하게 결과가 같습니다. 감사와 재현 가능성이 이 성질 위에 섭니다.

{{< callout type="info" >}}
**fail-open** — 그래프 빌드는 언제나 종료 코드 0 을 냅니다. 잘못된 토큰이 있어도 중단하지 않고, 진단 경고만 남긴 채 건전한 부분의 그래프를 만듭니다.
{{< /callout >}}

## 작성 예시

설계 문서에서 결정과 심볼을 가리키고, 코드 주석에서 같은 결정과 심볼을 받아 적는 가장 단순한 모습입니다.

설계 문서 (`tech.md`):

```markdown
# Tech

세션 계층은 위임 접근을 위해 OAuth2 를 채택한다.

결정 @NAV:DEC-AUTH-STRATEGY: 클라이언트 자격증명(client-credentials) 방식의 OAuth2.

헤더 파서 (see @NAV:SYM:pkg.ParseHeader) 가 Bearer 토큰을 뽑아낸다.
```

코드 (`auth/auth.go`):

```go
package auth

// @NAV:DEC-AUTH-STRATEGY: OAuth2 client-credentials 흐름을 구현한다.
// @NAV:SYM:auth.ParseBearer 가 Authorization 헤더에서 Bearer 토큰을 뽑아낸다.
func ParseBearer(h string) string { ... }
```

이 두 파일에서 그래프는 세 노드 (결정 `AUTH-STRATEGY`, 심볼 `pkg.ParseHeader`, 심볼 `auth.ParseBearer`) 와 그 사이의 간선을 만듭니다. 그래프를 읽는 사람은 설계 문서에서 코드로, 코드에서 설계 근거로 자유롭게 오갈 수 있습니다.

## 앞으로 호환성

토큰 문법, 바인딩 레코드의 5-필드 모양, 그래프 스키마는 모두 앞으로 호환 (추가 전용) 입니다. 뒤따른 마일스톤이 필드를 추가할 수는 있어도, 기존 필드의 이름과 모양을 바꾸지는 않습니다. 한 번 짓은 토큰은 장기적으로 유효합니다.

## 관련 문서

- [MX 태그 시스템](/ko/advanced/mx-tags/) — `@MX:SPEC` 토큰의 원천 규칙. 내비게이터 계층은 이 출력을 소비합니다.
- [SPEC 기반 개발](/ko/core-concepts/spec-based-dev/) — SPEC 라이프사이클과 `@MX:SPEC` 의 상위 맥락.
- [에이전트 가이드](/ko/advanced/agent-guide/) — 에이전트가 코드 주석과 설계 문서를 어떻게 오가는가.
