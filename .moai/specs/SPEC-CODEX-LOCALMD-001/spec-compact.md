# spec-compact.md — SPEC-CODEX-LOCALMD-001 (auto-generated digest)

## 요구 모듈 (5)

1. **듀얼 적재·합성** (REQ-LMD-001, 003): CLAUDE.local.md + AGENTS.local.md → 단일 `developer_instructions` 오버라이드(`-c` + key=value 2요소 쌍 1개), 연결 후 marshal, CLAUDE first 순서, 핀된 출처 구분자(`<!-- source: <filename> -->` 헤더 라인).
2. **경로 가드·fail-closed** (REQ-LMD-002, 010): 파일별 IsRegular 가드 + `codexPathGuardError`/`codexModeName` 어휘; AGENTS.md/CLAUDE.md 무변경, symlink 금지, 3-path 테이블 불변(exactly-3 오류 경로 테스트는 plan M1의 신설 RED 과제).
3. **부재/빈 매트릭스** (REQ-LMD-004): 빈 파일 = 부재로 취급(per-file empty-as-absent); 비어있지 않은 콘텐츠의 연결을 주입, 비어있지 않은 것이 없으면 argv 불변 — 8셀 전부 핀.
4. **충돌·오버플로 방어** (REQ-LMD-005, 007): tail 내 `developer_instructions=`로 시작하는 `-c` 토큰과 합성값 공존 시 명시적 거부(합성값 없으면 정상 런치); 최종 JSON-escape 토큰 바이트 길이 > 상한(기본 131,072 − 4,096 = 126,976, `internal/config/defaults.go` 단일 상수)이면 launch 전 fail-closed.
5. **균일성·무절단·신선 판독·표면 정합** (REQ-LMD-006, 008, 009): 6 런치 경로 동일 페이로드(단일 funnel), 61KiB 바이트 보존 + 매 런치 신선 판독(캐시 금지), 헬프 카피 정합(핀 테스트가 헬프 본문만으로 통과).

## 수용 기준 (12)

AC-LMD-001 (3 verb 동일) · 002 (spawn/-w/factory 동일) · 003 (순서 + 구분자 fixture) · 004 (absent/empty 매트릭스 8셀) · 005 (파일별 거부 — socket/device 포함, Unix-gate note) · 006 (61,360-byte round-trip) · 007 (`-c` 충돌 거부 + 음성 대조 2셀) · 008 (argv 초과 fail-closed — 상한 126,976 재기술) · 009 (입력 비가공) · 010 (재시작 재판독 — REQ-LMD-006 신선 판독 조항이 앵커) · 011 (헬프·핀 정합) · 012 (LIVE 실세션 nonce — codex CLI 세션 로그 표면 명명, NOT_RUN이면 전체 un-PASS).

## 수정 대상 파일

**Run phase (in-scope)**:
- `internal/cli/codex_launcher.go` — 2-파일 로더, 합성, 충돌·오버플로 가드, 헬프 카피(:334-335)
- `internal/cli/codex_local_instructions_test.go` — 신규/확장 테스트 + 핀 토큰 갱신
- `internal/cli/codex_contract.go` — `codexClaudeLocalName` 상수 추가 (3-path 테이블 불변)

**Template (R1 운영자 결정 후에만)**:
- `internal/template/templates/AGENTS.md.tmpl` — §8 문구 (M4 중립성 긴장)

**Sync phase (manager-docs)**:
- `CHANGELOG.md`, `docs-site/content/{en,ko,ja,zh}/advanced/codex-dual-harness.md`, 리포 `AGENTS.md` §8

## Exclusions

- Factory broker scope / idle-wake (t1074/t1075 소관)
- AGENTS.md/CLAUDE.md 수정, symlink/import 체인 생성
- `codexInstructionRelPathsFn` 테이블 확장
- Codex CLI 외부 동작 변경
- sync-phase 문서의 run-phase 선실행
