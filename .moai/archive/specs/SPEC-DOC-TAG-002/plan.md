# @SPEC:DOC-TAG-002 구현 계획

## 개요 (Overview)

**Phase 2 목표**: Phase 1 라이브러리를 MoAI-ADK 워크플로우에 통합

**핵심 활동**:
- doc-syncer 에이전트 수정 (Phase 2.5 추가)
- /alfred:3-sync 명령 확장 (Phase 1.5 추가)
- moai-foundation-tags Skill 업데이트
- CLI 유틸리티 개발 (선택사항)

**예상 기간**: 2-3시간 (CLI 포함 시 3-4시간)

**성과물**:
- 4개 파일 수정 (doc-syncer, 3-sync, SKILL.md, examples.md)
- 2개 파일 생성 (선택: tag_commands.py, test_tag_commands.py)
- 통합 테스트 통과
- 문서 업데이트 완료

---

## 상세 구현 계획

### 1. doc-syncer 에이전트 수정 (30분)

**파일**: `.claude/agents/alfred/doc-syncer.md`

#### 수정 내용

**1.1 Phase 2.5 섹션 추가**

```markdown
## Workflow Phase 2.5: @DOC TAG 자동 생성

**목표**: 태그 없는 문서 파일에 @DOC TAG 자동 생성

**실행 조건**:
- Phase 2 (문서 동기화) 완료 후
- 사용자가 Phase 1.5에서 TAG 생성 승인 시

**단계**:

1. **태그 없는 파일 스캔**
   - `docs/` 디렉토리 전체 스캔
   - `DocumentParser.extract_tags()`로 기존 @DOC TAG 확인
   - 태그 없는 파일 목록 생성

2. **TAG ID 생성**
   - `DocTagGenerator.generate_doc_tag(file_path, domain_hint=None)`
   - 파일명/디렉토리에서 도메인 추출 (예: `docs/auth/login.md` → `AUTH`)
   - 도메인별 시퀀스 번호 할당 (예: `AUTH-001`)

3. **SPEC-DOC 매핑**
   - `SpecDocMapper.find_related_spec(doc_path, spec_dir=".moai/specs")`
   - 도메인 일치 확인 (예: `SPEC-AUTH-001` ↔ `docs/auth/`)
   - 신뢰도 점수 계산 (0.0~1.0)
   - 신뢰도별 분류:
     - **HIGH (0.8-1.0)**: 도메인 완전 일치 → 자동 제안
     - **MEDIUM (0.5-0.8)**: 부분 일치 → 제안 + 확인 요청
     - **LOW (<0.5)**: 매핑 불확실 → 수동 개입 요청

4. **사용자 승인 확인** (Phase 1.5에서 이미 승인됨)
   - Phase 1.5에서 승인 받았으므로 바로 진행

5. **TAG 삽입**
   - `TagInserter.insert_tag_to_markdown(file_path, tag, chain_ref=None)`
   - 파일 백업 생성: `{file}.backup`
   - 마크다운 헤더에 TAG 삽입:
     ```markdown
     # @DOC:AUTH-001 | Chain: @SPEC:AUTH-001 -> @DOC:AUTH-001

     # 원본 제목
     ```
   - 삽입 성공 시: 원본 삭제, 백업 보관
   - 삽입 실패 시: 백업 복원, 에러 로그

6. **TAG 인벤토리 업데이트**
   - `TagRegistry.add_tag(tag_id, file_path, tag_type="DOC")`
   - `.moai/memory/tag-inventory.json` 업데이트
   - 트레이서빌리티 체인 기록

**출력**:
- TAG 삽입 성공 목록
- 실패한 파일 목록 (에러 메시지 포함)
- TAG 인벤토리 업데이트 요약

**에러 처리**:
- 백업 생성 실패 → TAG 삽입 중단
- 신뢰도 < 0.5 → 수동 개입 요청
- 파일 쓰기 실패 → 백업 복원
```

**1.2 워크플로우 다이어그램 업데이트**

```markdown
## doc-syncer 전체 워크플로우

Phase 1: 상태 분석
    ├─ 프로젝트 문서 검색 (.moai/project/, docs/)
    ├─ SPEC 문서 검색 (.moai/specs/)
    └─ 현재 상태 분석 리포트 생성
    ↓
Phase 2: 문서 동기화
    ├─ 프로젝트 문서 ↔ SPEC 동기화
    ├─ SPEC ↔ 구현 코드 동기화
    └─ 불일치 항목 식별
    ↓
Phase 2.5: @DOC TAG 자동 생성 (NEW)
    ├─ 태그 없는 파일 스캔
    ├─ TAG ID 생성 (도메인 기반)
    ├─ SPEC-DOC 매핑 (신뢰도 점수)
    ├─ TAG 삽입 (백업 생성)
    └─ TAG 인벤토리 업데이트
    ↓
Phase 3: 품질 검증
    ├─ TAG 체인 무결성 검증
    ├─ 문서 완전성 검증
    └─ 최종 리포트 생성
```

#### 테스트 계획

- **단위 테스트**: Phase 1 라이브러리 테스트 재사용 (이미 90.5% 커버리지)
- **통합 테스트**: doc-syncer 전체 워크플로우 테스트
- **E2E 테스트**: /alfred:3-sync 명령으로 전체 플로우 테스트

---

### 2. /alfred:3-sync 명령 수정 (30분)

**파일**: `.claude/commands/alfred/3-sync.md`

#### 수정 내용

**2.1 Phase 1.5 섹션 추가**

```markdown
## Workflow Phase 1.5: TAG 할당 체크

**목표**: 태그 없는 문서 파일 확인 및 TAG 생성 제안

**실행 위치**: Phase 0 (프로젝트 분석) 완료 후, Phase 1 (doc-syncer 호출) 이전

**단계**:

1. **태그 없는 파일 스캔**
   - `docs/` 디렉토리의 모든 마크다운 파일 스캔
   - `DocumentParser.extract_tags()`로 @DOC TAG 존재 확인
   - 태그 없는 파일 목록 생성

2. **TAG 제안 생성**
   - 각 파일에 대해 `suggest_tag_for_file(file_path)` 호출:
     - `DocTagGenerator.generate_doc_tag()` → TAG ID 생성
     - `SpecDocMapper.find_related_spec()` → SPEC 매핑
     - 신뢰도 점수 계산
   - 제안 목록 생성:
     ```
     태그 없는 파일: 10개

     HIGH 신뢰도 (8개):
     - docs/auth/login.md → @DOC:AUTH-001 (SPEC:AUTH-001, 신뢰도: 0.92)
     - docs/auth/register.md → @DOC:AUTH-002 (SPEC:AUTH-001, 신뢰도: 0.89)
     ...

     MEDIUM 신뢰도 (2개):
     - docs/guides/setup.md → @DOC:GUIDE-001 (신뢰도: 0.65)
     - docs/guides/deploy.md → @DOC:GUIDE-002 (신뢰도: 0.58)

     LOW 신뢰도 (0개):
     (수동 개입 필요한 파일 없음)
     ```

3. **사용자 상호작용**
   - `AskUserQuestion` 도구 사용:
     ```
     질문: "10개 파일에 @DOC TAG를 생성하시겠습니까?"

     옵션:
     1) 예, 모두 생성 (Y)
     2) 아니오, 건너뛰기 (n)
     3) HIGH 신뢰도만 생성 (h)
     4) 수동으로 선택 (m)

     기본값: Y
     ```

4. **승인 처리**
   - **Y (모두 생성)**: Phase 1 진행 → doc-syncer Phase 2.5 실행
   - **n (건너뛰기)**: Phase 2.5 생략, 기존 워크플로우 계속
   - **h (HIGH만)**: HIGH 신뢰도 파일만 Phase 2.5 처리
   - **m (수동 선택)**: 파일별 선택 UI 표시 (선택사항)

**출력**:
- TAG 제안 목록 (신뢰도별 분류)
- 사용자 선택 결과
- 다음 단계 안내

**에러 처리**:
- 파일 스캔 실패 → 경고 표시, 기존 워크플로우 계속
- 신뢰도 계산 실패 → 해당 파일 제외, 나머지 진행
```

**2.2 워크플로우 다이어그램 업데이트**

```markdown
## /alfred:3-sync 전체 워크플로우

Phase 0: 프로젝트 분석
    ├─ 프로젝트 모드 확인 (Personal/Team)
    ├─ Git 상태 확인
    └─ 문서 구조 파악
    ↓
Phase 1.5: TAG 할당 체크 (NEW)
    ├─ 태그 없는 파일 스캔
    ├─ TAG 제안 생성 (신뢰도 점수)
    ├─ 신뢰도별 분류 (HIGH/MEDIUM/LOW)
    ├─ 제안 목록 표시
    └─ AskUserQuestion: "TAG 생성?"
         ├─ Y → Phase 1 (Phase 2.5 포함)
         ├─ n → Phase 2 (Phase 2.5 생략)
         ├─ h → Phase 1 (HIGH만 Phase 2.5)
         └─ m → 수동 선택 UI
    ↓
Phase 1: doc-syncer 호출
    ├─ doc-syncer Phase 1: 상태 분석
    ├─ doc-syncer Phase 2: 문서 동기화
    ├─ doc-syncer Phase 2.5: @DOC TAG 자동 생성 (Phase 1.5 승인 시)
    └─ doc-syncer Phase 3: 품질 검증
    ↓
Phase 2: TAG 검증
    ├─ tag-agent 호출
    ├─ TAG 체인 무결성 검증
    └─ 검증 리포트 생성
    ↓
Phase 3: 최종 리포트
    ├─ 동기화 결과 요약
    ├─ TAG 생성 결과 (Phase 2.5 실행 시)
    └─ 다음 단계 안내
```

#### 테스트 계획

- **단위 테스트**: Phase 1.5 로직 테스트
- **통합 테스트**: Phase 1.5 → doc-syncer Phase 2.5 플로우
- **E2E 테스트**: /alfred:3-sync 전체 명령 테스트

---

### 3. moai-foundation-tags Skill 업데이트 (20분)

#### 3.1 SKILL.md 업데이트

**파일**: `.claude/skills/moai-foundation-tags/SKILL.md`

**추가 섹션**:

```markdown
## @DOC TAG: 문서 태그 자동 생성

### 개요

@DOC TAG는 문서 파일에 고유 식별자를 부여하여 SPEC ↔ DOC 트레이서빌리티를 구축합니다.

**Phase 2 통합**:
- `/alfred:3-sync` Phase 1.5: TAG 할당 체크
- `doc-syncer` Phase 2.5: @DOC TAG 자동 생성

### @DOC TAG 포맷

**기본 형식**:
```markdown
# @DOC:DOMAIN-NNN

# 문서 제목
```

**SPEC 체인 참조 형식**:
```markdown
# @DOC:DOMAIN-NNN | Chain: @SPEC:SOURCE-ID -> @DOC:DOMAIN-NNN

# 문서 제목
```

**예시**:
```markdown
# @DOC:AUTH-001 | Chain: @SPEC:AUTH-001 -> @DOC:AUTH-001

# 사용자 인증 가이드

이 문서는 사용자 인증 시스템의 구현 가이드입니다.
```

### SPEC-DOC 매핑

**매핑 원칙**:
1. **도메인 일치**: SPEC과 DOC의 도메인이 일치해야 함
   - `@SPEC:AUTH-001` ↔ `@DOC:AUTH-001`
2. **파일 위치 힌트**: 파일 경로에서 도메인 추출
   - `docs/auth/login.md` → `AUTH` 도메인
3. **신뢰도 점수**: 매핑 확실성을 0.0~1.0 점수로 표현

**신뢰도 점수 기준**:

| 점수 범위 | 레벨 | 의미 | 처리 방식 |
|-----------|------|------|-----------|
| 0.8 - 1.0 | HIGH | 도메인 완전 일치, 파일명 일치 | 자동 제안 (사용자 승인 필요) |
| 0.5 - 0.8 | MEDIUM | 도메인 부분 일치, 키워드 일치 | 제안 + 확인 요청 |
| < 0.5 | LOW | 매핑 불확실 | 수동 개입 요청 |

**매핑 예시**:

| SPEC | DOC 파일 | 도메인 | 신뢰도 | 결과 |
|------|---------|--------|--------|------|
| @SPEC:AUTH-001 | docs/auth/login.md | AUTH | 0.95 | HIGH - 자동 제안 |
| @SPEC:AUTH-001 | docs/guides/authentication.md | GUIDE | 0.68 | MEDIUM - 확인 요청 |
| @SPEC:AUTH-001 | docs/misc/notes.md | MISC | 0.32 | LOW - 수동 개입 |

### Phase 1.5/2.5 통합

**Phase 1.5: TAG 할당 체크** (in `/alfred:3-sync`)

1. `docs/` 디렉토리 스캔
2. `suggest_tag_for_file()` 호출
3. 신뢰도별 분류 (HIGH/MEDIUM/LOW)
4. 제안 목록 표시
5. `AskUserQuestion`: "TAG 생성?"

**Phase 2.5: @DOC TAG 자동 생성** (in `doc-syncer`)

1. 태그 없는 파일 스캔
2. `generate_doc_tag()` → TAG ID 생성
3. `find_related_spec()` → SPEC-DOC 매핑
4. `insert_tag_to_markdown()` → TAG 삽입
5. 백업 관리 (생성/삭제/보관)
6. TAG 인벤토리 업데이트

### TAG 자동 생성 Best Practices

1. **사용자 승인 우선**: 자동 생성 전 항상 사용자 승인 받기
2. **파일 백업**: TAG 삽입 전 반드시 백업 생성
3. **신뢰도 확인**: 신뢰도 < 0.5인 경우 수동 개입 요청
4. **체인 참조**: SPEC 매핑 시 Chain 참조 포함
5. **인벤토리 업데이트**: TAG 생성 후 즉시 인벤토리 업데이트
```

#### 3.2 examples.md 업데이트

**파일**: `.claude/skills/moai-foundation-tags/examples.md`

**추가 예제**:

```markdown
## @DOC TAG 자동 생성 예제

### 예제 1: 도메인 기반 매핑

**입력**:
- SPEC: `@SPEC:AUTH-001` (사용자 인증 시스템)
- 문서: `docs/auth/login.md` (로그인 가이드, TAG 없음)

**처리**:
1. 파일 경로에서 도메인 추출: `docs/auth/` → `AUTH`
2. TAG ID 생성: `@DOC:AUTH-001` (AUTH 도메인의 첫 번째 문서)
3. SPEC 매핑: `@SPEC:AUTH-001` ↔ `@DOC:AUTH-001` (도메인 일치)
4. 신뢰도 점수: 0.95 (HIGH - 도메인 완전 일치)

**출력** (`docs/auth/login.md`):
```markdown
# @DOC:AUTH-001 | Chain: @SPEC:AUTH-001 -> @DOC:AUTH-001

# 로그인 가이드

사용자 로그인 기능의 구현 가이드입니다.
```

---

### 예제 2: 신뢰도 낮은 경우 수동 매핑

**입력**:
- SPEC: `@SPEC:AUTH-001` (사용자 인증 시스템)
- 문서: `docs/misc/old-notes.md` (구 메모, TAG 없음)

**처리**:
1. 파일 경로에서 도메인 추출: `docs/misc/` → `MISC`
2. TAG ID 생성: `@DOC:MISC-001`
3. SPEC 매핑: `@SPEC:AUTH-001` ↔ `@DOC:MISC-001` (도메인 불일치)
4. 신뢰도 점수: 0.28 (LOW - 매핑 불확실)

**출력**:
- 자동 매핑 생략
- 사용자에게 수동 개입 요청:
  ```
  ⚠️  LOW 신뢰도 파일 발견:
  - docs/misc/old-notes.md → @DOC:MISC-001 (신뢰도: 0.28)

  수동으로 SPEC을 매핑하거나 TAG를 직접 생성해주세요.
  ```

---

### 예제 3: /alfred:3-sync에서의 TAG 제안 플로우

**시나리오**: 10개의 문서 파일에 TAG가 없는 상태

**Phase 1.5 실행**:

1. **파일 스캔 결과**:
   ```
   태그 없는 파일: 10개

   HIGH 신뢰도 (7개):
   - docs/auth/login.md → @DOC:AUTH-001 (SPEC:AUTH-001, 0.95)
   - docs/auth/register.md → @DOC:AUTH-002 (SPEC:AUTH-001, 0.92)
   - docs/auth/password-reset.md → @DOC:AUTH-003 (SPEC:AUTH-001, 0.89)
   - docs/api/endpoints.md → @DOC:API-001 (SPEC:API-001, 0.91)
   - docs/api/authentication.md → @DOC:API-002 (SPEC:API-001, 0.88)
   - docs/deploy/production.md → @DOC:DEPLOY-001 (SPEC:DEPLOY-001, 0.94)
   - docs/deploy/staging.md → @DOC:DEPLOY-002 (SPEC:DEPLOY-001, 0.87)

   MEDIUM 신뢰도 (2개):
   - docs/guides/setup.md → @DOC:GUIDE-001 (0.65)
   - docs/guides/troubleshooting.md → @DOC:GUIDE-002 (0.58)

   LOW 신뢰도 (1개):
   - docs/misc/notes.md → @DOC:MISC-001 (0.32)
   ```

2. **사용자 상호작용**:
   ```
   질문: "10개 파일에 @DOC TAG를 생성하시겠습니까?"

   옵션:
   1) 예, 모두 생성 (Y)
   2) 아니오, 건너뛰기 (n)
   3) HIGH 신뢰도만 생성 (h)
   4) 수동으로 선택 (m)

   기본값: Y
   ```

3. **사용자 선택: "h" (HIGH만)**

4. **Phase 2.5 실행**:
   - 7개 HIGH 신뢰도 파일에 TAG 삽입
   - 2개 MEDIUM + 1개 LOW는 수동 개입 대기

5. **결과**:
   ```
   ✅ TAG 생성 완료: 7개
   ⚠️  수동 개입 필요: 3개

   생성된 TAG:
   - @DOC:AUTH-001, @DOC:AUTH-002, @DOC:AUTH-003
   - @DOC:API-001, @DOC:API-002
   - @DOC:DEPLOY-001, @DOC:DEPLOY-002

   수동 개입 필요:
   - docs/guides/setup.md (MEDIUM, 0.65)
   - docs/guides/troubleshooting.md (MEDIUM, 0.58)
   - docs/misc/notes.md (LOW, 0.32)
   ```

---

### 예제 4: 일반적인 문제 해결

#### 문제 1: 고아 DOC (Orphaned DOC)

**상황**: @DOC TAG가 있지만 관련 SPEC이 없음

**검증**:
```bash
moai-adk tag-validate docs/
```

**출력**:
```
⚠️  고아 DOC 발견:
- @DOC:FEATURE-001 (docs/feature/guide.md)
  → 관련 SPEC 없음
```

**해결**:
1. SPEC 생성: `/alfred:1-plan "FEATURE 기능"`
2. SPEC-DOC 매핑 업데이트: `moai-adk tag-map FEATURE-001 docs/feature/`

---

#### 문제 2: 중복 TAG ID

**상황**: 동일한 @DOC TAG가 여러 파일에 존재

**검증**:
```bash
moai-adk tag-validate docs/
```

**출력**:
```
❌ 중복 TAG 발견:
- @DOC:AUTH-001
  - docs/auth/login.md
  - docs/auth/guide.md (중복!)
```

**해결**:
1. 한 파일의 TAG 수동 변경: `@DOC:AUTH-001` → `@DOC:AUTH-002`
2. TAG 인벤토리 업데이트: `moai-adk tag-generate docs/ --update-inventory`

---

#### 문제 3: 신뢰도 점수 낮음

**상황**: SPEC-DOC 매핑 신뢰도가 0.5 이하

**자동 처리**:
- 자동 매핑 생략
- 사용자에게 수동 개입 요청

**수동 해결**:
1. 파일에 TAG 직접 추가:
   ```markdown
   # @DOC:CUSTOM-001

   # 문서 제목
   ```

2. 또는 SPEC 매핑 직접 지정:
   ```markdown
   # @DOC:CUSTOM-001 | Chain: @SPEC:CORE-001 -> @DOC:CUSTOM-001

   # 문서 제목
   ```

3. TAG 인벤토리 업데이트:
   ```bash
   moai-adk tag-generate docs/ --update-inventory
   ```
```

---

### 4. CLI 유틸리티 개발 (선택사항, 30분)

**파일**: `src/moai_adk/cli/tag_commands.py` (~200 LOC)

#### 구현 내용

```python
"""
CLI commands for @DOC TAG management.
"""

import click
from pathlib import Path
from moai_adk.tags import (
    DocumentParser,
    DocTagGenerator,
    SpecDocMapper,
    TagInserter,
    TagRegistry,
)


@click.group()
def tag():
    """@DOC TAG management commands."""
    pass


@tag.command()
@click.argument("docs_dir", type=click.Path(exists=True))
@click.option("--update-inventory", is_flag=True, help="Update TAG inventory after generation")
def generate(docs_dir: str, update_inventory: bool):
    """
    Scan for untagged markdown files and generate @DOC TAGs.

    Example:
        moai-adk tag-generate docs/
    """
    docs_path = Path(docs_dir)
    parser = DocumentParser()
    generator = DocTagGenerator()
    mapper = SpecDocMapper()
    inserter = TagInserter()
    registry = TagRegistry()

    # Scan for untagged files
    untagged = []
    for md_file in docs_path.rglob("*.md"):
        tags = parser.extract_tags(md_file)
        if not any(tag.startswith("@DOC:") for tag in tags):
            untagged.append(md_file)

    click.echo(f"Found {len(untagged)} untagged files")

    # Generate TAG suggestions
    suggestions = []
    for file_path in untagged:
        tag_id = generator.generate_doc_tag(file_path)
        spec_mapping = mapper.find_related_spec(file_path)
        suggestions.append({
            "file": file_path,
            "tag": tag_id,
            "spec": spec_mapping["spec_id"] if spec_mapping else None,
            "confidence": spec_mapping["confidence"] if spec_mapping else 0.0,
        })

    # Classify by confidence
    high = [s for s in suggestions if s["confidence"] >= 0.8]
    medium = [s for s in suggestions if 0.5 <= s["confidence"] < 0.8]
    low = [s for s in suggestions if s["confidence"] < 0.5]

    click.echo(f"\nHIGH confidence ({len(high)}):")
    for s in high:
        click.echo(f"  - {s['file']} → {s['tag']} (SPEC:{s['spec']}, {s['confidence']:.2f})")

    click.echo(f"\nMEDIUM confidence ({len(medium)}):")
    for s in medium:
        click.echo(f"  - {s['file']} → {s['tag']} ({s['confidence']:.2f})")

    click.echo(f"\nLOW confidence ({len(low)}):")
    for s in low:
        click.echo(f"  - {s['file']} → {s['tag']} ({s['confidence']:.2f}) [manual intervention needed]")

    # Ask for confirmation
    if not click.confirm("\nGenerate TAGs for all files?", default=True):
        click.echo("Aborted.")
        return

    # Insert TAGs
    success = 0
    failed = []
    for s in suggestions:
        chain_ref = f"@SPEC:{s['spec']} -> {s['tag']}" if s["spec"] else None
        result = inserter.insert_tag_to_markdown(s["file"], s["tag"], chain_ref)
        if result["success"]:
            success += 1
            if update_inventory:
                registry.add_tag(s["tag"], s["file"], "DOC")
        else:
            failed.append((s["file"], result["error"]))

    click.echo(f"\n✅ TAG generation completed: {success}/{len(suggestions)}")
    if failed:
        click.echo(f"❌ Failed: {len(failed)}")
        for file, error in failed:
            click.echo(f"  - {file}: {error}")


@tag.command()
@click.argument("docs_dir", type=click.Path(exists=True))
def validate(docs_dir: str):
    """
    Validate existing @DOC TAGs in markdown files.

    Example:
        moai-adk tag-validate docs/
    """
    docs_path = Path(docs_dir)
    parser = DocumentParser()
    registry = TagRegistry()

    # Scan for tagged files
    tagged = []
    for md_file in docs_path.rglob("*.md"):
        tags = parser.extract_tags(md_file)
        doc_tags = [tag for tag in tags if tag.startswith("@DOC:")]
        if doc_tags:
            tagged.append((md_file, doc_tags))

    click.echo(f"Found {len(tagged)} tagged files")

    # Validate TAGs
    orphaned = []
    duplicates = {}
    for file_path, tags in tagged:
        for tag in tags:
            # Check for orphaned DOCs
            if not registry.find_related_spec(tag):
                orphaned.append((tag, file_path))

            # Check for duplicates
            if tag in duplicates:
                duplicates[tag].append(file_path)
            else:
                duplicates[tag] = [file_path]

    # Report orphaned DOCs
    if orphaned:
        click.echo(f"\n⚠️  Orphaned DOCs ({len(orphaned)}):")
        for tag, file_path in orphaned:
            click.echo(f"  - {tag} ({file_path})")
            click.echo(f"    → No related SPEC found")

    # Report duplicates
    dup_tags = {tag: files for tag, files in duplicates.items() if len(files) > 1}
    if dup_tags:
        click.echo(f"\n❌ Duplicate TAGs ({len(dup_tags)}):")
        for tag, files in dup_tags.items():
            click.echo(f"  - {tag}")
            for file_path in files:
                click.echo(f"    - {file_path}")

    if not orphaned and not dup_tags:
        click.echo("\n✅ All TAGs are valid!")


@tag.command()
@click.argument("spec_id")
@click.argument("docs_dir", type=click.Path(exists=True))
def map(spec_id: str, docs_dir: str):
    """
    Display SPEC-DOC mapping for a given SPEC ID.

    Example:
        moai-adk tag-map AUTH-001 docs/
    """
    docs_path = Path(docs_dir)
    mapper = SpecDocMapper()

    # Find related DOCs
    related = mapper.find_related_docs(f"SPEC-{spec_id}")

    if not related:
        click.echo(f"No related DOCs found for SPEC-{spec_id}")
        return

    click.echo(f"SPEC-DOC mapping for @SPEC:{spec_id}:")
    for doc in related:
        click.echo(f"  - {doc['file']}")
        click.echo(f"    TAG: {doc['tag']}")
        click.echo(f"    Confidence: {doc['confidence']:.2f}")
        if doc.get("chain"):
            click.echo(f"    Chain: {doc['chain']}")
```

#### 테스트 파일

**파일**: `tests/cli/test_tag_commands.py` (~150 LOC)

```python
"""
Tests for CLI tag commands.
"""

import pytest
from click.testing import CliRunner
from pathlib import Path
from moai_adk.cli.tag_commands import tag


@pytest.fixture
def runner():
    return CliRunner()


@pytest.fixture
def temp_docs(tmp_path):
    """Create temporary docs directory with test files."""
    docs = tmp_path / "docs"
    docs.mkdir()

    # Untagged file
    (docs / "auth.md").write_text("# Login Guide\n\nThis is a guide.")

    # Tagged file
    (docs / "api.md").write_text("# @DOC:API-001\n\n# API Guide\n\nThis is an API guide.")

    return docs


def test_tag_generate(runner, temp_docs):
    """Test tag-generate command."""
    result = runner.invoke(tag, ["generate", str(temp_docs)], input="y\n")
    assert result.exit_code == 0
    assert "Found 1 untagged files" in result.output


def test_tag_validate(runner, temp_docs):
    """Test tag-validate command."""
    result = runner.invoke(tag, ["validate", str(temp_docs)])
    assert result.exit_code == 0
    assert "Found 1 tagged files" in result.output


def test_tag_map(runner, temp_docs):
    """Test tag-map command."""
    result = runner.invoke(tag, ["map", "AUTH-001", str(temp_docs)])
    assert result.exit_code == 0
```

---

## 폴더 구조 (수정/생성 파일)

```
.claude/
├── agents/alfred/
│   └── doc-syncer.md (수정: + Phase 2.5 섹션)
├── commands/alfred/
│   └── 3-sync.md (수정: + Phase 1.5 섹션)
└── skills/moai-foundation-tags/
    ├── SKILL.md (수정: + @DOC 섹션)
    └── examples.md (수정: + @DOC 예제)

src/moai_adk/cli/
└── tag_commands.py (생성: ~200 LOC, 선택사항)

tests/cli/
└── test_tag_commands.py (생성: ~150 LOC, 선택사항)
```

---

## 기술 스택

**재사용 라이브러리** (Phase 1):
- `DocumentParser`: 마크다운 파싱 및 TAG 추출
- `DocTagGenerator`: 도메인 기반 TAG ID 생성
- `SpecDocMapper`: SPEC-DOC 매핑 및 신뢰도 점수
- `TagInserter`: 마크다운 헤더에 TAG 삽입
- `TagRegistry`: TAG 인벤토리 관리

**사용자 상호작용**:
- `AskUserQuestion`: TUI 기반 사용자 승인/거부 도구

**Skill 시스템**:
- `moai-foundation-tags`: TAG 체계 및 Best Practice 문서

**CLI 프레임워크** (선택사항):
- `click`: Python CLI 라이브러리

---

## 일정 및 마일스톤

### Primary Goal (필수)

| 단계 | 활동 | 예상 시간 | 성과물 |
|------|------|-----------|--------|
| 1 | doc-syncer 수정 | 30분 | Phase 2.5 섹션 추가 |
| 2 | /alfred:3-sync 수정 | 30분 | Phase 1.5 섹션 추가 |
| 3 | Skill 업데이트 | 20분 | SKILL.md + examples.md 업데이트 |
| 4 | 통합 테스트 | 30분 | 전체 워크플로우 E2E 테스트 |

**합계**: 2시간

### Secondary Goal (선택사항)

| 단계 | 활동 | 예상 시간 | 성과물 |
|------|------|-----------|--------|
| 5 | CLI 유틸리티 개발 | 30분 | tag_commands.py (~200 LOC) |
| 6 | CLI 테스트 작성 | 30분 | test_tag_commands.py (~150 LOC) |

**합계**: 1시간

**전체 합계**: 2-3시간 (또는 CLI 포함 시 3-4시간)

---

## 리스크 및 대응

### 리스크 1: 기존 워크플로우 호환성 문제

**리스크**: Phase 1.5/2.5 추가로 기존 워크플로우가 깨질 수 있음

**대응**:
- 모든 수정은 기존 코드에 **추가만**, 제거 없음
- Phase 1.5/2.5는 **선택적 단계** (사용자 거부 시 생략)
- 기존 Phase 0-3 워크플로우는 **그대로 유지**
- 통합 테스트로 호환성 검증

**완화 조치**:
- Phase 1.5에서 사용자 선택 "n" → Phase 2.5 생략, 기존 플로우 계속
- doc-syncer Phase 2.5 실패 시 → Phase 3로 바로 진행

---

### 리스크 2: 사용자 상호작용으로 워크플로우 지연

**리스크**: `AskUserQuestion`으로 사용자 승인 대기 시간 발생

**대응**:
- 승인 단계 제공 (Y/n/h/m)
- 기본값 제공 (Y - 모두 생성)
- 자동 모드 옵션 (선택사항): `--auto-approve` 플래그

**완화 조치**:
- Phase 1.5에서 한 번만 승인 받고, Phase 2.5는 자동 진행
- 신뢰도 HIGH만 자동 처리 옵션 제공 ("h")

---

### 리스크 3: SPEC 매핑 오류

**리스크**: 잘못된 SPEC-DOC 매핑으로 트레이서빌리티 손상

**대응**:
- 신뢰도 점수 기반 필터링 (< 0.5 → 수동 개입)
- 사용자 승인 단계에서 제안 검토
- 수동 확인 옵션 제공 ("m")

**완화 조치**:
- Phase 2 TAG 검증 단계에서 고아 DOC 탐지
- CLI `tag-validate` 명령으로 사후 검증
- 잘못된 매핑 발견 시 수동 수정 가이드 제공

---

### 리스크 4: 파일 백업 실패

**리스크**: TAG 삽입 중 파일 백업 실패로 데이터 손실 가능

**대응**:
- 백업 생성 실패 시 TAG 삽입 **즉시 중단**
- 에러 로그 기록 및 사용자 알림
- Git 커밋 전 백업 필수

**완화 조치**:
- 백업 디렉토리 권한 확인 (사전 검증)
- 디스크 공간 부족 시 경고 표시
- Git 이력으로 복구 가능 (최후의 수단)

---

## 문서화

### 업데이트 대상

1. **`.claude/agents/alfred/doc-syncer.md`**: Phase 2.5 섹션 추가
2. **`.claude/commands/alfred/3-sync.md`**: Phase 1.5 섹션 추가
3. **`.claude/skills/moai-foundation-tags/SKILL.md`**: @DOC 섹션 추가
4. **`.claude/skills/moai-foundation-tags/examples.md`**: @DOC 예제 추가

### 새로 생성 (선택사항)

5. **`docs/cli-reference.md`**: CLI 명령 참고 문서
6. **`docs/tag-generation-guide.md`**: TAG 자동 생성 가이드

---

## 다음 단계 (Phase 3)

Phase 2 완료 후:

1. **Phase 3 계획**: E2E 워크플로우 자동화 (SPEC → CODE → DOC → TAG 전체 체인)
2. **Phase 4 계획**: 성능 최적화 및 대규모 프로젝트 지원 (100+ 파일)
3. **Pre-commit Hook**: TAG 자동 생성을 pre-commit에 통합
4. **CI/CD 통합**: GitHub Actions에서 TAG 검증 자동화

---

**END OF PLAN**
