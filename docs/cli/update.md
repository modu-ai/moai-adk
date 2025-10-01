# moai update

## 1. 서론 — 실패한 바이브 코딩 vs 에이전틱 코딩

현장에서 가장 흔한 실패는 "일단 vibing 하다가" 운 좋게 맞추려는 방식입니다. 템플릿이 바뀌었는지, 백업을 남겼는지, 업데이트가 성공했는지에 대한 근거가 없기 때문에 한번 빗나가면 되돌리기 어렵습니다. 아래 표는 MoAI-ADK가 지향하는 에이전틱(Agentic) 코딩과 대비됩니다.

| 구분 | 바이브 코딩 | 에이전틱 코딩 |
|------|--------------|----------------|
| 의사결정 | 느낌/즉흥 | 근거 기반 / 로그 추적 |
| 실패 원인 | 무백업, 옵션 의미 미파악 | 체크리스트 기반 동작 실패 예방 |
| `moai update` 활용 | 거의 없음 | 업데이트 루틴을 통한 지속적 동기화 |

MoAI-ADK는 “입력 → 처리 → 반환”이 명확한 도구를 제공해 팀이 에이전틱 코딩을 실천하도록 돕습니다. 실패하지 않는 흐름을 만들려면 (1) 지금 실행하려는 명령이 무엇을 하는지 이해하고, (2) 성공/실패를 즉시 관찰하며, (3) 되돌릴 수 있는 안전장치를 준비해야 합니다. `moai update` 문서는 이 세 가지를 기준으로 다시 쓴다.

---

## 2. `moai update` 개요

- **목적**: MoAI-ADK 패키지와 프로젝트 템플릿(`.moai/`, `.claude/`, `CLAUDE.md`)을 최신 버전으로 맞추는 자동화 스크립트.
- **현재 구현** (`src/cli/commands/update.ts`, `src/core/update/update-orchestrator.ts`):
  1. 버전 확인 (`getCurrentVersion`, `checkLatestVersion`).
  2. 백업(기본) 혹은 백업 생략(`--no-backup`).
  3. `npm install moai-adk@latest` 실행 후 템플릿 파일 목록을 단순 복사(덮어쓰기).
  4. 핵심 파일 존재 여부를 짧게 검증하고 종료.
- **중요 제약**: 충돌 병합(ConflictResolver), 마이그레이션 프레임워크, 세부 로깅은 아직 구현되지 않았습니다.

---

## 3. 옵션 요약

| 옵션 | 지원 | 동작 | 비고 |
|------|------|------|------|
| `--check` | O | 최신 버전 여부만 확인, 파일 변경 없음 | 성공 시 종료 코드 0, 업데이트 필요 시 메시지 표시 |
| `--no-backup` | O | 백업 단계를 건너뛰고 바로 업데이트 | 위험 플래그, `UpdateConfiguration.force=true` 전달 |
| `--verbose`, `-v` | △ | 로거의 verbose 모드만 켜짐 | 현재 구현에서는 추가 정보가 거의 없음 |
| `--package-only` | X(미동작) | CLI에 정의돼 있지만 사용 안 함 | 향후 패키지만 갱신하는 기능 계획 |
| `--resources-only` | X(미동작) | CLI에 정의돼 있지만 사용 안 함 | 향후 템플릿만 재설치 기능 계획 |
| `--project-path <path>` | △ | 옵션 구조에 존재하지만 `run()`에서 전달되지 않음 | 현재 명령 실행 디렉터리만 업데이트됨 |

> **로드맵**: ConflictResolver, MigrationFramework, 실제 `--project-path` 지원은 문서 하단 “8. 로드맵 & 제한 사항”에 정리했습니다.

---

## 4. 업데이트 파이프라인 (실제 구현)

```mermaid
flowchart TD
    A[moai update 실행] --> B{--check?}
    B -->|예| C[버전 비교 후 종료]
    B -->|아니오| D{--no-backup?}
    D -->|예| E[백업 생략]
    D -->|아니오| F[.moai-backup/<ts> 생성]
    E --> G
    F --> G
    G[패키지 업데이트 (npm install moai-adk@latest)] --> H[템플릿 파일 덮어쓰기]
    H --> I[핵심 파일 존재 검증]
    I --> J[결과 로그 후 종료]
```

에이전틱 코딩 체크포인트:
1. **상태 파악** — `--check`로 먼저 버전 확인.
2. **안전망** — 기본 백업 유지, 생략 시 Git 커밋 선행.
3. **검증** — 실행 후 `moai status` 또는 테스트 스위트로 검증.

---

## 5. 단계별 상세 동작

### 5.1 버전 확인
- `getCurrentVersion()`으로 설치된 CLI 버전, `checkLatestVersion()`으로 npm 최신 버전 확인.
- 최신일 경우: `"✅ 최신 버전을 사용 중입니다"` 메시지 후 즉시 종료.
- 업데이트 필요: 최신 버전 정보 출력 후 다음 단계 진행.

### 5.2 백업 처리
- 기본은 `.moai-backup/<ISO 타임스탬프>/` 폴더에 `.moai`, `.claude`, `CLAUDE.md`를 복제.
- `--no-backup` 플래그가 있으면 백업 건너뛰고 경고 메시지 출력.
- 백업 실패 시 에러 메시지 후 종료 (되돌릴 여지 확보).

### 5.3 패키지 & 템플릿 업데이트
- `npm install moai-adk@latest` 시도. 지역 `package.json` 없으면 글로벌 설치로 fallback.
- 템플릿 복사 대상: `.claude/commands/moai`, `.claude/agents/moai`, `.claude/hooks/moai`, `.moai/memory/development-guide.md`, `.moai/project/{product,structure,tech}.md`, `CLAUDE.md`.
- 충돌 검사 없이 기존 파일을 덮어씌움.

### 5.4 검증 및 로그
- `.moai/memory/development-guide.md`, `CLAUDE.md`, `.claude` 하위 디렉터리 존재 여부만 확인.
- 성공 시:
  - `✅ 설치 완료` 같은 로그와 함께 백업 경로(있다면) 표시.
- 실패 시:
  - `❌ 업데이트 실패: <에러>` 출력, 백업 사용 안내 없음 → 사용자가 직접 `moai restore` 수행 필요.

---

## 6. 옵션별 사용 시나리오

### 6.1 기본 업데이트 (백업 포함)
```bash
moai update
```
예상 출력:
```
🔍 MoAI-ADK 업데이트 확인 중...
📦 현재 버전: v0.0.1
⚡ 최신 버전: v0.0.2
💾 백업 생성 중...
   → <프로젝트>/.moai-backup/2025-03-15-14-30-00
📦 패키지 업데이트 중...
   ✅ 패키지 업데이트 완료
📄 템플릿 파일 복사 중...
   ✅ 8개 파일 복사 완료
🔍 검증 중...
   ✅ 검증 완료
✨ 업데이트 완료!
```

### 6.2 사전 점검 (`--check`)
```bash
moai update --check
```
예상 출력:
```
🔍 MoAI-ADK 업데이트 확인 중...
📦 현재 버전: v0.0.1
⚡ 최신 버전: v0.0.2
✅ 업데이트 가능 (파일 변경 없음)
```

### 6.3 백업 생략 (`--no-backup`)
```bash
# Git에 변경 사항을 커밋한 뒤 사용하는 것을 강력히 권장합니다.
moai update --no-backup
```
출력의 첫 줄에 `⚠️ Backup disabled - proceeding without safety net` 경고가 등장하고 그 외 흐름은 기본과 동일합니다.

### 6.4 Verbose 로깅 (`--verbose`)
```bash
moai update --verbose
```
현재 구현에서는 verbose 레벨이 켜지지만, 추가 상세 로그는 아직 준비돼 있지 않습니다. 실패 진단에는 도움되지 않으니 추후 개선까지는 참고용으로만 사용하세요.

---

## 7. 실패 모드 & 복구 전략

| 증상 | 원인 | 해결 방법 |
|------|------|-----------|
| `⚠️ This doesn't appear to be a MoAI-ADK project` | `.moai/` 디렉터리 없음 | `moai status`로 상태 확인 → 필요 시 `moai init` 실행 |
| `npm install` 관련 오류 | 네트워크/권한/레지스트리 문제 | 사내 레지스트리 사용 여부 확인, `npm config get registry`, VPN/프록시 점검 |
| `EACCES` 권한 오류 | 프로젝트 파일 소유권 불일치 | `chmod -R u+w .moai .claude`, `chown`으로 소유자 수정 후 재시도 |
| 백업 실패 | 디스크 부족 또는 경로 권한 문제 | 용량 확보 → 다시 실행, 임시로 `--no-backup` 사용 시 Git 커밋 필수 |
| 업데이트 후 동작 이상 | 덮어쓰기 때문에 사용자 변경 사라짐 | 백업에서 `moai restore <backup-path>` 수행 또는 Git으로 롤백 |
| `--project-path`가 먹지 않음 | 옵션은 정의됐지만 전달되지 않음 | 해당 기능은 로드맵, 현재는 대상 디렉터리에서 직접 실행 |

바이브 코딩은 “실패하면 다시 해보면 되지” 수준의 낙관을 전제로 합니다. 에이전틱 코딩은 **실패 가능성을 줄이기 위한 사전 점검**(테이블 7)과 **되돌릴 방법 확보**(백업 / Git)를 항상 포함합니다.

---

## 8. 로드맵 & 제한 사항

현재 문서는 구현된 기능을 기준으로 작성했습니다. 향후 계획과 제한 사항은 아래와 같습니다.

| 항목 | 상태 | 메모 |
|------|------|------|
| ConflictResolver, MigrationFramework | 미구현 | 문서에서 “계획”으로만 언급되던 기능. 설계 초안은 남아 있지만 코드 반영 전입니다. |
| `--project-path <path>` | 미지원 | CLI 옵션에 노출돼 있지만 `UpdateCommand.run`에서 사용하지 않습니다. |
| `--package-only`, `--resources-only` | 미동작 | 현재는 로그만 남기고 아무 동작도 하지 않습니다. |
| Verbose 세부 로그 | 제한적 | Console 레벨만 바뀌어서 실질적인 분석에는 부족. |

구현 완료 후 문서 업데이트 시에는 이 섹션을 “실제 동작”으로 옮기고, 테스트 커버리지와 회귀 테스트 계획을 함께 명시하는 것이 바람직합니다.

---

## 9. 관련 명령어 & 참고 자료

- [`moai init`](./init.md) — 프로젝트 초기화
- [`moai status`](./status.md) — 템플릿/버전 상태 확인
- [`moai doctor`](./doctor.md) — 시스템 요구사항 점검
- [`moai restore`](./restore.md) — 백업 복원
- 사내 가이드: 에이전틱 코딩 체크리스트, 업데이트 전후 테스트 시나리오

---

**요약**: `moai update`는 단순하지만 강력한 도구입니다. 에이전틱 코딩 방식으로 사용하면 업데이트 실패와 데이터 손실을 예방할 수 있고, 향후 기능 확장을 위해 필요한 근거도 확보할 수 있습니다. 먼저 `--check`, 백업 유지, 실행 후 검증이라는 세 단계를 습관화하세요.
