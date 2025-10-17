# @DOC:CMD-ALFRED-001 | Chain: @SPEC:DOCS-003 -> @DOC:CMD-001

# Alfred Commands

Alfred SuperAgent 명령어 레퍼런스입니다.

## /alfred:0-project

프로젝트 초기화 및 관리:

```bash
/alfred:0-project "새 프로젝트 초기화"
```

project-manager 에이전트가 처리합니다.

---

## /alfred:1-spec

SPEC 문서 작성:

```bash
/alfred:1-spec "사용자 인증 기능 구현"
```

spec-builder 에이전트가 EARS 방식의 SPEC을 생성합니다.

---

## /alfred:2-build

TDD 기반 코드 구현:

```bash
/alfred:2-build "SPEC-AUTH-001"
```

code-builder 에이전트가 RED-GREEN-REFACTOR 사이클을 수행합니다.

---

## /alfred:3-sync

문서 동기화 및 TAG 체인 검증:

```bash
/alfred:3-sync
```

doc-syncer 에이전트가 문서를 동기화하고 TAG 체인을 검증합니다.

---

**다음**: [Agent Commands →](agent-commands.md)
