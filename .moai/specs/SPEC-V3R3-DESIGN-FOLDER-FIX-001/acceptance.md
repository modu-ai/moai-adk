# SPEC-V3R3-DESIGN-FOLDER-FIX-001 — Acceptance Criteria

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.1.0   | 2026-04-26 | manager-spec | Initial acceptance criteria. 6 ACs covering update path warning, scaffold strict mode, message content, file preservation. |

---

## 1. Acceptance Criteria Overview

| AC ID    | REQ Coverage                                  | Priority | Test File / Method |
|----------|------------------------------------------------|----------|---------------------|
| AC-DFF-01 | REQ-DFF-001, REQ-DFF-002, REQ-DFF-004         | P0       | `TestDesignFolderUpdate_ReservedExact_WarnsButContinues` |
| AC-DFF-02 | REQ-DFF-001, REQ-DFF-002                      | P0       | `TestDesignFolderUpdate_ReservedGlob_WarnsButContinues` |
| AC-DFF-03 | REQ-DFF-003, REQ-DFF-008                      | P0       | `TestDesignFolderScaffold_ReservedExact_StillErrors` |
| AC-DFF-04 | REQ-DFF-005, REQ-DFF-007                      | P1       | `TestDesignFolderUpdate_WarningIncludesGuidance` |
| AC-DFF-05 | REQ-DFF-006, REQ-DFF-004                      | P0       | `TestDesignFolderUpdate_ReservedNotModified` (강화) |
| AC-DFF-06 | REQ-DFF-001, REQ-DFF-002                      | P1       | `TestDesignFolderUpdate_MultipleReservedConflicts` |

---

## 2. AC-DFF-01 — Update Path Reserved Exact Warns But Continues

### 2.1 REQ Coverage
REQ-DFF-001 (Update path warning), REQ-DFF-002 (Other files sync continuation), REQ-DFF-004 (User file preservation)

### 2.2 Given-When-Then

**Given**:
- `root` 임시 디렉터리에 `.moai/design/` 생성됨
- `scaffoldDesignDir(root, &warnBuf)`로 templates 정상 deploy됨 (README.md, research.md, system.md, spec.md)
- 사용자가 `.moai/design/tokens.json`을 내용 `{"primary": "#ff0000"}`로 작성
- 사용자가 `.moai/design/README.md`를 수정하여 user-edit 마커 추가

**When**:
- `updateDesignDir(root, &errBuf)` 호출

**Then**:
- 함수 반환값: `nil` (error 없음)
- `errBuf`에 "warning" 키워드 포함
- `errBuf`에 "tokens.json" 키워드 포함
- `errBuf`에 "preserved" 또는 동등 보존 안내 키워드 포함
- `.moai/design/tokens.json` 내용 변경 없음 (`{"primary": "#ff0000"}`)
- `.moai/design/README.md` user-edit 보존됨 (REQ-005 user-edit preservation 정상 동작)
- `.moai/design/research.md`, `system.md`, `spec.md`는 canonical hash 일치 시 정상 update (또는 user-modified면 보존)

### 2.3 Test Method

```go
func TestDesignFolderUpdate_ReservedExact_WarnsButContinues(t *testing.T) {
    t.Parallel()
    root := t.TempDir()
    designDir := filepath.Join(root, ".moai", "design")
    // ... scaffold + create tokens.json + modify README ...
    var errBuf strings.Builder
    err := updateDesignDir(root, &errBuf)
    if err != nil { t.Fatalf("expected nil, got %v", err) }
    // assert errBuf contains "warning" + "tokens.json" + "preserved"
    // assert tokens.json unchanged
    // assert README.md user edit preserved
}
```

### 2.4 Pass Criteria

- [ ] 모든 7개 assertion pass
- [ ] race detection 통과
- [ ] golangci-lint clean

---

## 3. AC-DFF-02 — Update Path Reserved Glob Warns But Continues

### 3.1 REQ Coverage
REQ-DFF-001, REQ-DFF-002

### 3.2 Given-When-Then

**Given**:
- `.moai/design/` scaffold 완료됨
- `.moai/design/brief/BRIEF-LOCAL.md` 사용자 작성 (내용 "local design brief notes")

**When**:
- `updateDesignDir(root, &errBuf)` 호출

**Then**:
- 반환값 `nil`
- `errBuf`에 "warning" + "BRIEF" + "preserved" 포함
- `BRIEF-LOCAL.md` 내용 변경 없음
- 다른 templates (README/research/system/spec) sync 정상 진행

### 3.3 Test Method

```go
func TestDesignFolderUpdate_ReservedGlob_WarnsButContinues(t *testing.T) {
    t.Parallel()
    // ... scaffold + create brief/BRIEF-LOCAL.md ...
    var errBuf strings.Builder
    err := updateDesignDir(root, &errBuf)
    if err != nil { t.Fatalf("expected nil, got %v", err) }
    // assert warning content + file preservation
}
```

### 3.4 Pass Criteria

- [ ] 5개 assertion pass

---

## 4. AC-DFF-03 — Scaffold Path Reserved Still Errors

### 4.1 REQ Coverage
REQ-DFF-003 (Scaffold path strict mode), REQ-DFF-008 (Path mode dispatch)

### 4.2 Given-When-Then

**Given**:
- `.moai/design/tokens.json` 사용자 작성 (이론적 신규 프로젝트 케이스)
- 다른 design files 없음 또는 일부 존재

**When**:
- `checkReservedCollision(root, &errBuf, true)` 직접 호출 (scaffold strict mode 시뮬레이션)

**Then**:
- 반환값: error (non-nil)
- error 메시지에 "reserved filename" + "tokens.json" 포함
- `errBuf`에 "error: reserved filename: tokens.json" 포함 (warning이 아닌 error 키워드)
- tokens.json 내용 변경 없음 (REQ-DFF-004 always honored)

### 4.3 Test Method

```go
func TestDesignFolderScaffold_ReservedExact_StillErrors(t *testing.T) {
    t.Parallel()
    root := t.TempDir()
    designDir := filepath.Join(root, ".moai", "design")
    os.MkdirAll(designDir, 0o755)
    os.WriteFile(filepath.Join(designDir, "tokens.json"), []byte(`{"x":1}`), 0o644)

    var errBuf strings.Builder
    err := checkReservedCollision(root, &errBuf, true) // strict=true
    if err == nil { t.Fatal("expected error in strict mode") }
    if !strings.Contains(err.Error(), "reserved filename") {
        t.Errorf("error must mention reserved filename, got %v", err)
    }
    // assert tokens.json unchanged
}
```

### 4.4 Pass Criteria

- [ ] 4개 assertion pass
- [ ] strict mode와 update mode 분리 검증

---

## 5. AC-DFF-04 — Warning Includes Guidance

### 5.1 REQ Coverage
REQ-DFF-005 (Warning guidance), REQ-DFF-007 (No silent failure)

### 5.2 Given-When-Then

**Given**:
- `.moai/design/components.json` 사용자 작성

**When**:
- `updateDesignDir(root, &errBuf)` 호출

**Then**:
- `errBuf`에 다음 모두 포함:
  - "warning" 키워드 (silent failure 방지, REQ-DFF-007)
  - 파일 경로 ("components.json")
  - "preserved" 또는 보존 안내 키워드
  - 우회 절차 힌트 ("rename" 또는 "canonical templates" 키워드)

### 5.3 Test Method

```go
func TestDesignFolderUpdate_WarningIncludesGuidance(t *testing.T) {
    t.Parallel()
    root := t.TempDir()
    designDir := filepath.Join(root, ".moai", "design")
    os.MkdirAll(designDir, 0o755)
    os.WriteFile(filepath.Join(designDir, "components.json"), []byte("{}"), 0o644)

    var errBuf strings.Builder
    err := updateDesignDir(root, &errBuf)
    if err != nil { t.Fatalf("unexpected error: %v", err) }

    msg := errBuf.String()
    requiredKeywords := []string{"warning", "components.json", "preserved", "rename"}
    for _, kw := range requiredKeywords {
        if !strings.Contains(strings.ToLower(msg), strings.ToLower(kw)) {
            t.Errorf("warning must contain %q, got: %s", kw, msg)
        }
    }
}
```

### 5.4 Pass Criteria

- [ ] 모든 4 keyword 포함 확인
- [ ] 메시지가 빈 문자열이 아님

---

## 6. AC-DFF-05 — Reserved File Not Modified

### 6.1 REQ Coverage
REQ-DFF-006 (No auto-overwrite), REQ-DFF-004 (User file preservation)

### 6.2 Given-When-Then

**Given**:
- `.moai/design/components.json` 사용자 작성, 내용 `{"user": "data"}`
- `.moai/design/import-warnings.json` 사용자 작성, 내용 `{"warning": "test"}`

**When**:
- `updateDesignDir(root, &errBuf)` 호출

**Then**:
- 반환값 `nil`
- 두 파일 모두 내용 변경 없음
- 두 파일 모두 size 변경 없음 (`os.Stat` 비교)
- mtime은 변경될 수 있음 (구현 디테일)

### 6.3 Test Method

```go
func TestDesignFolderUpdate_ReservedNotModified(t *testing.T) {
    t.Parallel()
    root := t.TempDir()
    designDir := filepath.Join(root, ".moai", "design")
    os.MkdirAll(designDir, 0o755)

    components := []byte(`{"user": "data"}`)
    imports := []byte(`{"warning": "test"}`)
    os.WriteFile(filepath.Join(designDir, "components.json"), components, 0o644)
    os.WriteFile(filepath.Join(designDir, "import-warnings.json"), imports, 0o644)

    var errBuf strings.Builder
    if err := updateDesignDir(root, &errBuf); err != nil {
        t.Fatalf("expected nil, got %v", err)
    }

    // assert byte-identical content for both files
    got1, _ := os.ReadFile(filepath.Join(designDir, "components.json"))
    if !bytes.Equal(got1, components) { t.Error("components.json modified") }
    got2, _ := os.ReadFile(filepath.Join(designDir, "import-warnings.json"))
    if !bytes.Equal(got2, imports) { t.Error("import-warnings.json modified") }
}
```

### 6.4 Pass Criteria

- [ ] 두 파일 모두 byte-identical 보존
- [ ] 함수 nil 반환

---

## 7. AC-DFF-06 — Multiple Reserved Conflicts

### 7.1 REQ Coverage
REQ-DFF-001, REQ-DFF-002

### 7.2 Given-When-Then

**Given**:
- `.moai/design/tokens.json` 사용자 작성
- `.moai/design/brief/BRIEF-X.md` 사용자 작성
- `.moai/design/components.json` 사용자 작성

**When**:
- `updateDesignDir(root, &errBuf)` 호출

**Then**:
- 반환값 `nil`
- `errBuf`에 세 파일 이름 모두 포함 ("tokens.json", "BRIEF-X", "components.json")
- 세 파일 모두 내용 보존
- 다른 templates (README 등) sync 정상

### 7.3 Test Method

```go
func TestDesignFolderUpdate_MultipleReservedConflicts(t *testing.T) {
    t.Parallel()
    root := t.TempDir()
    designDir := filepath.Join(root, ".moai", "design")
    briefDir := filepath.Join(designDir, "brief")
    os.MkdirAll(briefDir, 0o755)

    files := map[string][]byte{
        filepath.Join(designDir, "tokens.json"):     []byte(`{"a":1}`),
        filepath.Join(briefDir, "BRIEF-X.md"):       []byte("brief X"),
        filepath.Join(designDir, "components.json"): []byte(`[]`),
    }
    for path, content := range files {
        os.WriteFile(path, content, 0o644)
    }

    var errBuf strings.Builder
    if err := updateDesignDir(root, &errBuf); err != nil {
        t.Fatalf("expected nil, got %v", err)
    }

    msg := errBuf.String()
    for _, kw := range []string{"tokens.json", "BRIEF-X", "components.json"} {
        if !strings.Contains(msg, kw) {
            t.Errorf("warning must mention %q, got: %s", kw, msg)
        }
    }
    // assert all three files preserved
    for path, want := range files {
        got, _ := os.ReadFile(path)
        if !bytes.Equal(got, want) {
            t.Errorf("%s modified", path)
        }
    }
}
```

### 7.4 Pass Criteria

- [ ] 세 파일 이름 모두 warning 메시지에 포함
- [ ] 세 파일 모두 byte-identical 보존
- [ ] 함수 nil 반환

---

## 8. Edge Cases

### 8.1 errOut == nil

**시나리오**: caller가 `updateDesignDir(root, nil)` 호출

**Expected**: panic 없음. warning은 silent (현재 코드도 동일하게 nil check 있음). 함수는 nil 반환.

**검증**: 단위 테스트 `TestDesignFolderUpdate_NilErrOut`가 panic 없이 통과.

### 8.2 Empty .moai/design/ Directory

**시나리오**: `.moai/design/`은 존재하지만 파일 없음

**Expected**: reserved 충돌 없음 (warning도 출력 안 됨). templates 정상 deploy. nil 반환.

**Existing Coverage**: `TestDesignFolderSkip_EmptyDir`에서 이미 검증됨.

### 8.3 Reserved Match in Subdirectory Outside brief/

**시나리오**: `.moai/design/wireframes/tokens.json` (wireframes는 reserved subdir 아님)

**Expected**: `tokens.json`은 exact match 정책상 `.moai/design/tokens.json`에서만 검사. wireframes/ 하위는 무시. (기존 동작 유지)

**검증**: 신규 테스트 추가 가능하나 본 SPEC 범위 외. 기존 정책 그대로.

---

## 9. Definition of Done

### 9.1 Code

- [ ] `internal/cli/design_folder.go`: `checkReservedCollision` strict 파라미터 추가, `updateDesignDir` 호출 변경
- [ ] `internal/cli/design_folder_test.go`: 6개 AC 테스트 모두 추가/갱신
- [ ] 모든 신규/갱신 테스트 GREEN

### 9.2 Documentation

- [ ] `.claude/rules/moai/design/constitution.md` v3.3.1 amendment
- [ ] `internal/template/templates/.claude/rules/moai/design/constitution.md` mirror
- [ ] HISTORY entry 추가됨

### 9.3 Quality

- [ ] `go test ./...` 전체 통과
- [ ] `go test -race ./internal/cli/...` 통과
- [ ] `golangci-lint run` 0 warnings
- [ ] `go vet ./...` clean
- [ ] `make build` 성공

### 9.4 Verification

- [ ] 사용자 보고 시나리오 (`mo.ai.kr/.moai/design/tokens.json`) 수동 재현 → warning + 다른 sync 정상 확인
- [ ] commit message 한국어 body + Conventional Commits

---

## 10. Quality Gate Mapping

| TRUST Pillar | 검증 |
|--------------|------|
| Tested       | 6 ACs, 모든 REQ-DFF-* 커버, race detection 포함 |
| Readable     | strict bool 파라미터로 의도 명확, warning 메시지 자명 |
| Unified      | 기존 design_folder.go 코드 스타일 준수, gofmt |
| Secured      | 사용자 데이터 수정 절대 없음 (REQ-DFF-004/006) |
| Trackable    | SPEC ID + commit message + HISTORY entry로 traceable |

---

Version: 0.1.0
Last Updated: 2026-04-26
REQ coverage: REQ-DFF-001 ~ REQ-DFF-008 (full)
AC count: 6
