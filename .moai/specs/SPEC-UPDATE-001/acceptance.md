# SPEC-UPDATE-001: 수락 기준

---
tags: SPEC-UPDATE-001
---

## 1. 버전 확인 (Checker)

### AC-UPD-001: 최신 버전 조회 성공

```gherkin
Given GitHub Releases API에 v1.2.0 릴리즈가 존재하고
  And 네트워크가 정상적으로 연결되어 있을 때
When Checker.CheckLatest(ctx)를 호출하면
Then VersionInfo가 반환되어야 한다
  And VersionInfo.Version은 "v1.2.0"이어야 한다
  And VersionInfo.URL은 유효한 HTTPS URL이어야 한다
  And VersionInfo.Checksum은 SHA-256 해시 형식이어야 한다
  And VersionInfo.Date는 릴리즈 날짜여야 한다
```

### AC-UPD-002: 업데이트 가용성 판단 -- 업데이트 있음

```gherkin
Given 현재 바이너리 버전이 "v1.1.0"이고
  And 최신 릴리즈 버전이 "v1.2.0"일 때
When Checker.IsUpdateAvailable("v1.1.0")을 호출하면
Then (true, *VersionInfo, nil)이 반환되어야 한다
  And VersionInfo.Version은 "v1.2.0"이어야 한다
```

### AC-UPD-003: 업데이트 가용성 판단 -- 이미 최신

```gherkin
Given 현재 바이너리 버전이 "v1.2.0"이고
  And 최신 릴리즈 버전이 "v1.2.0"일 때
When Checker.IsUpdateAvailable("v1.2.0")을 호출하면
Then (false, nil, nil)이 반환되어야 한다
```

### AC-UPD-004: 네트워크 오류 처리

```gherkin
Given GitHub API에 접근할 수 없을 때 (네트워크 타임아웃)
When Checker.CheckLatest(ctx)를 호출하면
Then nil과 오류가 반환되어야 한다
  And 오류 메시지에 타임아웃 또는 네트워크 관련 정보가 포함되어야 한다
```

### AC-UPD-005: 컨텍스트 취소 처리

```gherkin
Given 이미 취소된 context가 있을 때
When Checker.CheckLatest(cancelledCtx)를 호출하면
Then nil과 context.Canceled 오류가 반환되어야 한다
```

---

## 2. 바이너리 다운로드 (Updater.Download)

### AC-UPD-006: 플랫폼 바이너리 다운로드 성공

```gherkin
Given 유효한 VersionInfo (v1.2.0, darwin/arm64 URL)가 있고
  And 다운로드 서버가 정상 응답할 때
When Updater.Download(ctx, versionInfo)를 호출하면
Then 임시 파일 경로가 반환되어야 한다
  And 해당 파일이 존재해야 한다
  And 파일 크기가 0보다 커야 한다
  And 파일의 SHA-256 체크섬이 VersionInfo.Checksum과 일치해야 한다
```

### AC-UPD-007: 체크섬 불일치 시 실패

```gherkin
Given 다운로드된 바이너리의 체크섬이 VersionInfo.Checksum과 다를 때
When Updater.Download(ctx, versionInfo)를 호출하면
Then "" 빈 문자열과 ErrChecksumMismatch 오류가 반환되어야 한다
  And 임시 다운로드 파일이 정리(삭제)되어야 한다
```

### AC-UPD-008: 다운로드 중 네트워크 중단

```gherkin
Given 바이너리 다운로드 중 네트워크 연결이 끊길 때
When Updater.Download(ctx, versionInfo)가 실패하면
Then ErrDownloadFailed 오류가 반환되어야 한다
  And 부분 다운로드된 임시 파일이 정리(삭제)되어야 한다
```

### AC-UPD-009: 다운로드 타임아웃

```gherkin
Given 300초 타임아웃이 설정된 context가 있고
  And 다운로드 서버 응답이 300초를 초과할 때
When Updater.Download(ctx, versionInfo)를 호출하면
Then context.DeadlineExceeded 오류가 반환되어야 한다
  And 임시 파일이 정리되어야 한다
```

---

## 3. 원자적 바이너리 교체 (Updater.Replace)

### AC-UPD-010: 바이너리 교체 성공

```gherkin
Given 유효한 새 바이너리 파일이 임시 경로에 존재하고
  And 새 바이너리가 "moai version" 서브커맨드를 정상 실행할 수 있을 때
When Updater.Replace(ctx, newBinaryPath)를 호출하면
Then nil이 반환되어야 한다
  And 현재 바이너리 경로에 새 바이너리가 위치해야 한다
  And 새 바이너리에 실행 권한(0755)이 설정되어야 한다
```

### AC-UPD-011: 검증 실패 시 교체 거부

```gherkin
Given 새 바이너리 파일이 손상되어 실행할 수 없을 때
When Updater.Replace(ctx, corruptBinaryPath)를 호출하면
Then ErrReplaceFailed 오류가 반환되어야 한다
  And 기존 바이너리가 변경되지 않아야 한다
  And 손상된 임시 파일이 정리되어야 한다
```

### AC-UPD-012: 파일시스템 권한 부족

```gherkin
Given 현재 바이너리 위치에 쓰기 권한이 없을 때
When Updater.Replace(ctx, newBinaryPath)를 호출하면
Then 권한 관련 오류가 반환되어야 한다
  And 기존 바이너리가 변경되지 않아야 한다
```

---

## 4. 롤백 메커니즘 (Rollback)

### AC-UPD-013: 백업 생성 성공

```gherkin
Given 현재 바이너리가 /usr/local/bin/moai에 존재할 때
When Rollback.CreateBackup()을 호출하면
Then 백업 경로가 반환되어야 한다 (예: /usr/local/bin/moai.backup.1706918400)
  And 백업 파일이 원본과 동일한 내용이어야 한다
  And 백업 파일에 실행 권한이 보존되어야 한다
```

### AC-UPD-014: 백업에서 복원 성공

```gherkin
Given 유효한 백업 파일이 존재하고
  And 현재 바이너리가 손상되었을 때
When Rollback.Restore(backupPath)를 호출하면
Then nil이 반환되어야 한다
  And 현재 바이너리 위치에 백업 내용이 복원되어야 한다
  And 복원된 바이너리에 실행 권한(0755)이 설정되어야 한다
```

### AC-UPD-015: 존재하지 않는 백업에서 복원 시도

```gherkin
Given 백업 파일이 존재하지 않을 때
When Rollback.Restore("/nonexistent/path")를 호출하면
Then ErrRollbackFailed 오류가 반환되어야 한다
  And 오류 메시지에 백업 경로가 포함되어야 한다
```

### AC-UPD-016: 업데이트 실패 시 자동 롤백

```gherkin
Given 업데이트 오케스트레이션 중 바이너리 교체가 실패하고
  And 사전에 백업이 생성되어 있을 때
When Orchestrator.Update(ctx)가 Replace 단계에서 실패하면
Then 시스템은 자동으로 Rollback.Restore를 호출해야 한다
  And 원래 바이너리가 복원되어야 한다
  And UpdateResult는 nil이어야 하고 오류에 롤백 완료 정보가 포함되어야 한다
```

---

## 5. 3-Way Merge: LineMerge 전략

### AC-MRG-001: 한쪽만 변경 -- 자동 머지 성공

```gherkin
Given base 파일이 "line1\nline2\nline3"이고
  And current 파일이 "line1\nline2\nline3" (변경 없음)이고
  And updated 파일이 "line1\nline2_modified\nline3"일 때
When Engine.ThreeWayMerge(base, current, updated)를 호출하면
Then MergeResult.Content는 "line1\nline2_modified\nline3"이어야 한다
  And MergeResult.HasConflict는 false여야 한다
  And MergeResult.Strategy는 LineMerge여야 한다
```

### AC-MRG-002: 양쪽 변경 비충돌 -- 자동 머지 성공

```gherkin
Given base 파일이 "A\nB\nC\nD"이고
  And current 파일이 "A\nB_user\nC\nD" (2번째 줄 변경)이고
  And updated 파일이 "A\nB\nC\nD_template" (4번째 줄 변경)일 때
When Engine.ThreeWayMerge(base, current, updated)를 호출하면
Then MergeResult.Content는 "A\nB_user\nC\nD_template"이어야 한다
  And MergeResult.HasConflict는 false여야 한다
```

### AC-MRG-003: 양쪽 변경 충돌 -- 충돌 감지

```gherkin
Given base 파일이 "A\nB\nC"이고
  And current 파일이 "A\nB_user\nC" (2번째 줄을 B_user로 변경)이고
  And updated 파일이 "A\nB_template\nC" (2번째 줄을 B_template로 변경)일 때
When Engine.ThreeWayMerge(base, current, updated)를 호출하면
Then MergeResult.HasConflict는 true여야 한다
  And MergeResult.Conflicts 길이는 1이어야 한다
  And Conflicts[0].Current는 "B_user"를 포함해야 한다
  And Conflicts[0].Updated는 "B_template"를 포함해야 한다
```

### AC-MRG-004: 동일한 변경 -- 충돌 없이 머지

```gherkin
Given base 파일이 "A\nB\nC"이고
  And current 파일이 "A\nB_same\nC"이고
  And updated 파일이 "A\nB_same\nC" (양쪽이 동일하게 변경)일 때
When Engine.ThreeWayMerge(base, current, updated)를 호출하면
Then MergeResult.Content는 "A\nB_same\nC"이어야 한다
  And MergeResult.HasConflict는 false여야 한다
```

---

## 6. 3-Way Merge: YAMLDeep 전략

### AC-MRG-005: YAML 사용자 키 보존 + 새 키 추가

```gherkin
Given base YAML이 {a: 1, b: 2}이고
  And current YAML이 {a: 1, b: 2, user_key: "custom"} (사용자 키 추가)이고
  And updated YAML이 {a: 1, b: 3, c: 4} (b 변경, c 추가)일 때
When Engine.MergeFile(ctx, "config.yaml", base, current, updated)를 호출하면
Then 결과 YAML은 {a: 1, b: 3, c: 4, user_key: "custom"}이어야 한다
  And MergeResult.HasConflict는 false여야 한다
  And MergeResult.Strategy는 YAMLDeep이어야 한다
```

### AC-MRG-006: YAML 중첩 구조 deep merge

```gherkin
Given base YAML이 {server: {host: localhost, port: 8080}}이고
  And current YAML이 {server: {host: localhost, port: 9090}} (port 변경)이고
  And updated YAML이 {server: {host: localhost, port: 8080, timeout: 30}} (timeout 추가)일 때
When Engine.MergeFile(ctx, "config.yaml", base, current, updated)를 호출하면
Then 결과 YAML은 {server: {host: localhost, port: 9090, timeout: 30}}이어야 한다
  And 사용자의 port 변경이 보존되어야 한다
  And 템플릿의 timeout 추가가 반영되어야 한다
```

### AC-MRG-007: YAML 동일 키 양쪽 변경 충돌

```gherkin
Given base YAML이 {version: "1.0"}이고
  And current YAML이 {version: "1.1"} (사용자가 변경)이고
  And updated YAML이 {version: "2.0"} (템플릿이 변경)일 때
When Engine.MergeFile(ctx, "config.yaml", base, current, updated)를 호출하면
Then MergeResult.HasConflict는 true여야 한다
  And Conflicts에 "version" 키에 대한 충돌이 기록되어야 한다
```

---

## 7. 3-Way Merge: JSONMerge 전략

### AC-MRG-008: JSON 객체 머지

```gherkin
Given base JSON이 {"key1": "a", "key2": "b"}이고
  And current JSON이 {"key1": "a", "key2": "b", "user": true}이고
  And updated JSON이 {"key1": "a", "key2": "c", "key3": "d"}일 때
When Engine.MergeFile(ctx, "settings.json", base, current, updated)를 호출하면
Then 결과 JSON은 {"key1": "a", "key2": "c", "key3": "d", "user": true}이어야 한다
  And 결과는 json.Valid()를 통과해야 한다
  And MergeResult.Strategy는 JSONMerge여야 한다
```

### AC-MRG-009: JSON 배열 합집합 머지

```gherkin
Given base JSON 배열이 ["a", "b"]이고
  And current JSON 배열이 ["a", "b", "user_item"]이고
  And updated JSON 배열이 ["a", "b", "new_item"]일 때
When 배열 머지를 수행하면
Then 결과는 ["a", "b", "user_item", "new_item"]이어야 한다
  And 중복이 제거되어야 한다
```

---

## 8. 3-Way Merge: SectionMerge 전략

### AC-MRG-010: CLAUDE.md 사용자 섹션 보존

```gherkin
Given base CLAUDE.md가 "## Section A\ncontent_a\n## Section B\ncontent_b"이고
  And current CLAUDE.md가 "## Section A\ncontent_a\n## Section B\ncontent_b\n## My Custom\nmy_content" (사용자 섹션 추가)이고
  And updated CLAUDE.md가 "## Section A\ncontent_a_new\n## Section B\ncontent_b\n## Section C\ncontent_c" (A 변경, C 추가)일 때
When Engine.MergeFile(ctx, "CLAUDE.md", base, current, updated)를 호출하면
Then 결과에 "## Section A\ncontent_a_new"가 포함되어야 한다 (템플릿 변경 반영)
  And 결과에 "## Section C\ncontent_c"가 포함되어야 한다 (새 섹션 추가)
  And 결과에 "## My Custom\nmy_content"가 포함되어야 한다 (사용자 섹션 보존)
  And MergeResult.Strategy는 SectionMerge여야 한다
```

### AC-MRG-011: CLAUDE.md 동일 섹션 내 충돌

```gherkin
Given base CLAUDE.md의 "## Config" 섹션 내용이 "default"이고
  And current에서 "custom_user_config"로 변경되었고
  And updated에서 "new_template_config"로 변경되었을 때
When Engine.MergeFile(ctx, "CLAUDE.md", base, current, updated)를 호출하면
Then MergeResult.HasConflict는 true여야 한다
  And "## Config" 섹션에 대한 충돌이 기록되어야 한다
```

---

## 9. 3-Way Merge: EntryMerge 전략

### AC-MRG-012: .gitignore 엔트리 머지

```gherkin
Given base .gitignore가 "*.pyc\n__pycache__/\n.env"이고
  And current .gitignore가 "*.pyc\n__pycache__/\n.env\nmy_secret.txt" (사용자 추가)이고
  And updated .gitignore가 "*.pyc\n__pycache__/\n.env\n.moai/cache/" (템플릿 추가)일 때
When Engine.MergeFile(ctx, ".gitignore", base, current, updated)를 호출하면
Then 결과에 "*.pyc"가 포함되어야 한다
  And 결과에 "__pycache__/"가 포함되어야 한다
  And 결과에 ".env"가 포함되어야 한다
  And 결과에 "my_secret.txt"가 포함되어야 한다 (사용자 추가 보존)
  And 결과에 ".moai/cache/"가 포함되어야 한다 (템플릿 추가 반영)
  And 중복 엔트리가 없어야 한다
  And MergeResult.HasConflict는 false여야 한다
  And MergeResult.Strategy는 EntryMerge여야 한다
```

### AC-MRG-013: .gitignore 사용자 삭제 엔트리 미복원

```gherkin
Given base .gitignore가 "*.pyc\n*.log\n.env"이고
  And current .gitignore가 "*.pyc\n.env" (사용자가 *.log 삭제)이고
  And updated .gitignore가 "*.pyc\n*.log\n.env\n.cache/"일 때
When Engine.MergeFile(ctx, ".gitignore", base, current, updated)를 호출하면
Then 결과에 "*.log"가 포함되지 않아야 한다 (사용자 삭제 존중)
  And 결과에 ".cache/"가 포함되어야 한다 (신규 추가 반영)
```

---

## 10. 3-Way Merge: Overwrite 전략

### AC-MRG-014: 전체 교체 + 백업

```gherkin
Given 바이너리 형식 파일(비머지 대상)이 존재하고
  And 사용자의 현재 파일 내용이 있을 때
When Overwrite 전략이 적용되면
Then 결과 Content는 updated 내용과 동일해야 한다
  And 원본 파일이 ".backup" 확장자로 보존되어야 한다
  And MergeResult.HasConflict는 false여야 한다
  And MergeResult.Strategy는 Overwrite여야 한다
```

---

## 11. 전략 선택기 (StrategySelector)

### AC-MRG-015: 파일 확장자별 전략 매핑

```gherkin
Given StrategySelector가 초기화되어 있을 때
When 다음 파일 경로로 SelectStrategy를 호출하면
Then 올바른 전략이 반환되어야 한다:
  | 파일 경로                    | 예상 전략     |
  | config.yaml                  | YAMLDeep      |
  | .moai/config/sections/user.yml | YAMLDeep    |
  | settings.json                | JSONMerge     |
  | manifest.json                | JSONMerge     |
  | CLAUDE.md                    | SectionMerge  |
  | .gitignore                   | EntryMerge    |
  | README.md                    | LineMerge     |
  | agents/expert-backend.md     | LineMerge     |
  | unknown.bin                  | Overwrite     |
  | image.png                    | Overwrite     |
```

---

## 12. 충돌 파일 생성

### AC-MRG-016: .conflict 파일 생성

```gherkin
Given 3-Way Merge에서 충돌이 감지되었을 때
When conflict.go가 충돌 파일을 생성하면
Then "{원본경로}.conflict" 파일이 생성되어야 한다
  And 파일 내용에 Git 스타일 충돌 마커가 포함되어야 한다:
    <<<<<<< current (사용자)
    B_user
    =======
    B_template
    >>>>>>> updated (템플릿)
  And 원본 파일은 수정되지 않아야 한다
```

### AC-MRG-017: 다중 충돌 영역 처리

```gherkin
Given 하나의 파일에서 3개의 충돌 영역이 감지되었을 때
When conflict.go가 충돌 파일을 생성하면
Then .conflict 파일에 3개의 충돌 마커 세트가 포함되어야 한다
  And MergeResult.Conflicts 길이는 3이어야 한다
  And 각 Conflict의 StartLine과 EndLine이 올바르게 설정되어야 한다
```

---

## 13. 통합: 오케스트레이션 전체 흐름

### AC-INT-001: 정상 업데이트 전체 흐름

```gherkin
Given 현재 버전이 "v1.1.0"이고 최신 버전이 "v1.2.0"이고
  And 프로젝트에 10개의 template_managed 파일과 3개의 user_modified 파일이 있고
  And 1개의 user_created 파일이 있을 때
When Orchestrator.Update(ctx)를 호출하면
Then UpdateResult가 반환되어야 한다
  And UpdateResult.PreviousVersion은 "v1.1.0"이어야 한다
  And UpdateResult.NewVersion은 "v1.2.0"이어야 한다
  And UpdateResult.FilesUpdated는 10이어야 한다 (template_managed 덮어쓰기)
  And UpdateResult.FilesMerged는 3이어야 한다 (user_modified 머지)
  And UpdateResult.FilesSkipped는 1이어야 한다 (user_created 건너뜀)
  And RollbackPath에 유효한 백업 경로가 포함되어야 한다
```

### AC-INT-002: 매니페스트 기반 provenance 처리

```gherkin
Given 매니페스트에 다음 파일이 등록되어 있을 때:
  | 파일 경로               | provenance       | 해시 변경 |
  | agents/backend.md       | template_managed | 없음     |
  | agents/custom.md        | user_created     | -        |
  | rules/core.md           | user_modified    | -        |
  | skills/old-skill.md     | deprecated       | -        |
When 업데이트 오케스트레이션이 각 파일을 처리하면
Then agents/backend.md는 안전 덮어쓰기되어야 한다
  And agents/custom.md는 건너뛰어야 한다 (절대 수정하지 않음)
  And rules/core.md는 3-way merge를 수행해야 한다
  And skills/old-skill.md는 사용자에게 알림 후 유지해야 한다
```

### AC-INT-003: template_managed 파일 해시 변경 시 승격

```gherkin
Given 매니페스트에서 agents/backend.md가 template_managed이고
  And 현재 파일 해시가 deployed_hash와 다를 때 (사용자가 수정)
When 업데이트 오케스트레이션이 해당 파일을 처리하면
Then 해당 파일의 provenance가 user_modified로 승격되어야 한다
  And 3-way merge가 수행되어야 한다
```

---

## 14. 통합: 오류 및 롤백 시나리오

### AC-INT-004: 다운로드 실패 시 롤백

```gherkin
Given 업데이트 오케스트레이션 중 Download 단계에서 실패할 때
When Orchestrator.Update(ctx)가 오류를 반환하면
Then 자동 롤백이 실행되어야 한다
  And 원래 바이너리가 보존되어야 한다
  And 오류 메시지에 "download failed" 정보가 포함되어야 한다
```

### AC-INT-005: 머지 중 충돌 발생 시 계속 진행

```gherkin
Given 업데이트 중 일부 파일에서 머지 충돌이 발생할 때
When Orchestrator.Update(ctx)를 호출하면
Then 충돌 파일에 대해 .conflict 파일을 생성하고
  And 나머지 파일의 처리를 계속해야 한다
  And UpdateResult.FilesConflicted에 충돌 파일 수가 기록되어야 한다
  And 바이너리 업데이트 자체는 정상 완료되어야 한다
```

### AC-INT-006: 롤백 실패 시 사용자 안내

```gherkin
Given 업데이트 실패 후 롤백도 실패할 때
When 시스템이 복구를 시도하면
Then 오류 메시지에 백업 파일 경로가 포함되어야 한다
  And "수동으로 {backup_path}에서 복원하세요" 형태의 안내가 출력되어야 한다
```

---

## 15. Diff 생성기

### AC-MRG-018: Unified diff 형식 출력

```gherkin
Given base가 "A\nB\nC"이고 current가 "A\nB_mod\nC"일 때
When differ.Diff(base, current)를 호출하면
Then 결과는 unified diff 형식이어야 한다
  And "-B" (삭제)와 "+B_mod" (추가) 라인이 포함되어야 한다
  And 컨텍스트 라인("A", "C")이 포함되어야 한다
```

### AC-MRG-019: 동일 파일 diff

```gherkin
Given base와 current가 동일한 내용일 때
When differ.Diff(base, current)를 호출하면
Then 빈 diff(변경 없음)가 반환되어야 한다
```

---

## 16. 비기능 수락 기준

### AC-NFR-001: 성능

```gherkin
Given 200개의 템플릿 파일이 포함된 프로젝트일 때
When 전체 3-Way Merge를 수행하면
Then 총 소요 시간은 5초 미만이어야 한다
```

### AC-NFR-002: 단일 파일 머지 성능

```gherkin
Given 1,000라인의 Markdown 파일일 때
When ThreeWayMerge를 수행하면
Then 소요 시간은 100ms 미만이어야 한다
```

### AC-NFR-003: 테스트 커버리지

```gherkin
Given internal/merge/ 및 internal/update/ 패키지가 구현되었을 때
When go test -coverprofile을 실행하면
Then 커버리지가 85% 이상이어야 한다
```

### AC-NFR-004: 임시 파일 정리

```gherkin
Given 업데이트가 성공하거나 실패한 후에
When 프로세스가 종료되면
Then 임시 디렉토리에 다운로드된 파일이 남아있지 않아야 한다
```

---

## 17. Definition of Done

모든 수락 기준(AC-UPD-001 ~ AC-UPD-016, AC-MRG-001 ~ AC-MRG-019, AC-INT-001 ~ AC-INT-006, AC-NFR-001 ~ AC-NFR-004)이 통과하면 SPEC-UPDATE-001은 완료된 것으로 간주한다.

**필수 통과 항목:**

- [ ] 모든 Gherkin 시나리오에 대응하는 Go 테스트 작성 및 통과
- [ ] `go test -race ./internal/merge/... ./internal/update/...` 통과
- [ ] 테스트 커버리지 >= 85%
- [ ] `golangci-lint run ./internal/merge/... ./internal/update/...` 경고 0건
- [ ] 벤치마크 테스트 성능 목표 달성
- [ ] 6가지 머지 전략 각각에 대한 testdata/ 픽스처 포함
- [ ] 크로스 플랫폼 빌드 성공 (darwin, linux, windows)
- [ ] Orchestrator 통합 테스트 (모킹 기반) 통과
