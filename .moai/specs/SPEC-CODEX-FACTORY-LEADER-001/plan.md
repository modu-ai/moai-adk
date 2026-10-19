# SPEC-CODEX-FACTORY-LEADER-001 — 구현 계획

## §A 맥락과 확정한 선택

- 카드 t1774와 운영자 결정 `d-20261015T000351Z-0dfd`가 Codex 리더 허용의 근거다.
- Tier M: 기존 런·CLI·운영 안내·회귀 테스트를 함께 변경하지만 별도 아키텍처나 저장소를 만들지 않는다.
- 리더 진입은 값 없는 `-f` / `--factory`다. 레인은 기존 `-l` 경로를 유지한다.
- t1745 재개 여부: **이번 카드에서 재개하지 않는다.** 기존 수명주기·런 등록은 재사용하되 bare -f에 필요하지 않은 네이티브 resume 파서와 재채택을 추가하지 않는다. 새 리더는 새 런을 생성한다.
- Codex 리더 기록은 `codex`다. 기존 레인의 `gpt` 백엔드 의미를 리더에 복사하지 않는다.

## §B 확인한 기준과 알려진 한계

작성 기준: `8249bff8c3933d647b4f99e59048595f9808e83a`, `WT-codex-factory-leader`.

- `moai todo list --json`에서 t1774 전체 본문과 결정 id를 확인했다.
- `internal/cli/codex_launcher.go`의 리더 거부와 `internal/cli/codex_factory.go`의 기존 리더 분기를 읽었다. 기존 run 등록·환경 전달·anchor 경로를 활용한다.
- t1745의 반려 문서 두 개만 `git show a9f53c169:<path>`로 복사한다. 코드·다른 기록은 병합하지 않는다.
- 저장 형식 명확화: `gtd.go`의 `RecordFactoryCardAssignment` 호출과 `todo_runtime.go`의 큐 SQLite 기록, `factory_mirror.go`의 팩토리 SQLite 배차 mirror를 소스로 확인했다. 카드의 “dispatch 파일”은 기존 SQLite 파일에 저장된 배차 기록을 뜻한다. 새 sidecar writer는 추가하지 않는다. 검증은 launcher → 실제 `todo engage --dispatch` → 두 저장소 재열기 → 실제 저장 상태 조회로 연결한다.
- 레인 완료 범위: 기존 `factory next`·`factory stage` 경로와 정당한 검증 증거로 저장된 `merge-ready`를 리더가 읽는다. 큐 `CardDone`이나 실제 병합·착지는 요구하지 않으며, 더 넓은 t1645의 착지 동작은 제외한다.
- 이번 기준의 실행 테스트·실제 모델 실행·전체 CI는 아직 관측하지 않았다. 정적 확인과 실행 증거를 구분한다.
- 카드 임대는 별도 운영 상태다. SPEC 작성과 유효한 작업트리 존재를 큐 임대·완료 증거로 해석하지 않는다.

## §C 사전 조건

1. SPEC ID Bash 정규식 검사 결과 `PASS`; 신규 id의 기존 경로 검색 결과 없음.
2. canonical 12-field schema에 맞춘 초안과 stateless plan·acceptance를 작성한다.
3. 독립 plan audit와 확정한 artifact hash를 확인한 뒤 실행을 시작한다. 이번 작업의 우선 착수와 로컬 RC 설치는 운영자가 명시적으로 지시했다.

## §D 제약

- `quality.yaml`의 `development_mode: tdd`: 실제 assertion으로 실패하는 RED를 먼저 기록하고 최소 GREEN 후 회귀를 확인한다.
- 기존 helper 재사용을 우선한다. 훅·메시징의 새 전송 기능은 추가하지 않는다.
- 하나의 작업트리에는 한 명의 구현 writer만 둔다. 이 계획 작성자는 SPEC 본문만 수정한다.
- 테스트는 `t.TempDir()` 기반의 격리된 루트와 홈을 사용한다. 실제 사용자 큐·설정은 변경하지 않는다.
- 반복되는 검증은 `moai verify run` 정책을 따르고 재사용 결과는 재관측으로 주장하지 않는다.

## §E 자체 검증과 위험

- malformed 형태는 상태 생성 전에 거부되어야 한다. 허용 분기만 넓혀 legacy 값까지 통과시키는 회귀를 검사한다.
- launcher stub은 child 경계만 대체한다. 큐 조회·기존 SQLite dispatch 저장·저장 완료 기록 읽기는 실제 경로를 실행하고 두 저장소를 다시 열어 확인한다.
- model·effort는 설정 상수를 재사용하고 plain 세션에 리더 기본값이 새지 않는지 검사한다.
- 살아 있는/판별 불가능한 writer를 anchor가 계속 거부하는지 확인한다.
- Codex의 PermissionRequest·PreCompact 부재와 메시지 nudge 부재를 운영 안내에 명시한다.
- 전 저장소 로컬 suite는 실행하지 않는다. 변경 영향 범위의 테스트·vet와 독립 감사 결과를 보고한다.

## §F 단계

### High — 사용자 동작과 검증 경계

1. M1 RED: bare -f 리더 시작, malformed 무부작용, 새 런·백엔드, no-nudge dispatch/완료 경로의 실패 재현을 별도 테스트로 남긴다.
2. M2 GREEN: 기존 Codex 리더 분기를 연결하고 기존 런·hook·broker 수명주기를 재사용한다. 리더 기본 모델·effort와 anchor를 유지한다.
3. M3 회귀: Claude/GLM 리더, Codex -l, plain Codex, legacy 오류, 혼합 레인 합류와 교차 리더 인수 거부를 검증한다.

### Medium — 안내와 동기화

4. M4 factory-dispatch 안내와 README를 갱신한다. 배포 템플릿에 포함된 안내는 template-first 및 생성물 검사 절차를 따른다.
5. M5 독립 실행 감사 후 manager-docs가 진행·상태와 동기화 증거를 기록한다.

### Low — 로컬 RC 전달

6. 운영자가 지시한 다음 로컬 RC 번호를 version-management runbook으로 측정한다. 승인된 통합 main에서 정식 build provenance를 넣어 빌드·복사하고 설치 검증 스크립트로 식별자·실행 경로를 확인한다. 태그·배포·전체 CI 성공을 추정하지 않는다.

## §G 금지하는 우회

- t1745 전체 코드 병합, 새 resume 파서, 레인 진입 개편.
- 메시지가 도착했다는 이유만으로 완료 처리하거나 저장 기록 검사를 생략하는 구현.
- 테스트용 문자열·fixture dispatch만 만들어 실제 큐 경로를 우회하는 검증.
- 로컬 focused PASS를 전체 CI·실제 Codex 모델 검증으로 보고하는 것.

## §H 참조

- spec.md §0·§B·§E, acceptance.md, progress.md.
- 선행 SPEC 두 개의 개정 HISTORY와 보존한 SPEC-FACTORY-CODEX-LEADER-RESUME-001.
- AGENTS.md·AGENTS.local.md, `.moai/docs/version-management.md`.
