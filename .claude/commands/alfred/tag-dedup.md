# Alfred TAG De-duplication Command

## Command: /alfred:tag-dedup

**Purpose**: GPT-5 Pro 분석 기반 TAG 중복 자동 정리 시스템 실행

**Usage**:
```bash
/alfred:tag-dedup [options]
```

**Options**:
- `--scan-only`: 중복 탐지만 수행, 수정하지 않음
- `--dry-run`: 시뮬레이션 모드, 수정 계획만 보여줌
- `--apply`: 실제 수정 적용 (백업 필수)
- `--output <path>`: 결과 리포트 저장 경로
- `--backup`: 수정 전 자동 백업 생성
- `--whitelist <path>`: 화이트리스트 파일 지정

**Examples**:
```bash
# 중복 탐지만 수행
/alfred:tag-dedup --scan-only

# 시뮬레이션으로 수정 계획 확인
/alfred:tag-dedup --dry-run

# 실제 수정 적용 (백업 포함)
/alfred:tag-dedup --apply --backup

# 특정 결과 파일 저장
/alfred:tag-dedup --dry-run --output .moai/reports/my-dedup.json
```

**Workflow**:
1. **Detection**: 전체 프로젝트 스캔으로 중복 TAG 식별
2. **Analysis**: 권한 기반 Primary/Related 분석
3. **Planning**: 자동 수정 계획 생성 (topline 중복 우선)
4. **Validation**: TAG 체인 무결성 검증
5. **Execution**: 승인 후 자동 수정 적용
6. **Verification**: 수정 결과 검증 및 리포트 생성

**Algorithm**:
```
1. Scan all eligible files for @TAG patterns
2. Group by TAG:TYPE-DOMAIN-ID pattern
3. Identify duplicates (>1 occurrence per pattern)
4. Select primary candidate (highest authority score):
   - .moai/specs/*/spec.md (100 points)
   - src/**/*.py, tests/**/*.py (80 points)
   - docs/api/**/*.md, docs/reference/**/*.md (60 points)
   - docs/guides/**, examples/** (40 points)
5. Plan corrections:
   - Topline duplicates: renumber with new ID
   - Reference duplicates: update to primary
   - Whitelisted paths: track only
6. Apply corrections with validation
7. Generate comprehensive report
```

**Exit Codes**:
- `0`: 성공, 치명적 중복 없음
- `1`: 치명적 중복 발견 또는 수정 실패
- `2`: 설정 오류 또는 시스템 오류

**Safety Features**:
- Automatic backup before changes
- Chain integrity validation
- Confidence threshold checking
- Rollback capability
- Dry-run mode default

**Integration**:
- PreToolUse 훅과 통합하여 실시간 검증
- Git pre-commit 훅으로 중복 방지
- Alfred workflow와 통합된 자동화

**Related Commands**:
- `/alfred:tag-audit`: TAG 시스템 전체 감사
- `/alfred:tag-renumber`: 특정 도메인 TAG 재번호화
- `/alfred:tag-reserve`: TAG 예약 및 관리
