# t517 Verdict — moai web 무저장 config 재작성 (SPEC-WEB-WRITE-SAFETY-001)

> 3-phase close 완료. 본 문서는 카드 t517의 증거 정본이다. 측정 트리: HEAD `0b9f881b0`, 브랜치 `WT-web-write-safety` (로컬 develop 대비 13커밋 + 본 문서 커밋).

## Claim

`moai web`의 settings 콘솔이 값이 바뀌지 않은 제출(무접촉 저장 포함)로 추적 config를 재작성하던 결함을 수리했다. SPEC-WEB-WRITE-SAFETY-001 plan→run→sync 3-phase close 완료, AC-WWS-001..008 전부 PASS, sync-audit 판정 PASS(F1·F1-D 수리 후, 가중 91.3).

## Evidence (커맨드 → 관측, 본 런·본 트리)

- **M-a 결론**: 무저장 쓰기 경로는 부재 — 기동/GET 9라우트/폴링 3단계 격리 재현 diff 0. 결함 실체는 **값-불변 제출의 무차별 기록**(값-불변 POST /save 1회로 리드 관측 O1 서명 완전 재현: feedback 빈 줄 1행 삭제 + git-strategy 키 이동·3키 추가). progress.md §E.2 M1 표.
- **M-b 결론**: git-strategy 우회 = `applyTypedEdits` 무조건 SetSection → dirty → 전체 재마샬. feedback 빈 줄 = yaml.v3 전체 재직렬화. progress.md §E.2 M3.
- **수리 7건**: 값-불변 seam edit 제거 · typed 실변경 게이트 · yamlpatch 라인 스플라이싱 · 중복 폼값 reject · nested 실변경 게이트 · **absent 극성 선언(`AbsentDefault`)+유효 기본값 인지 게이트** · **렌더 checked 산출에 AbsentDefault 반영**(`146372d4b`, `6639e7ecb`).
- **RED-first 진위**: 동작 게이트 미수리 상태에서 3건 FAIL exit 1(`RED-f1-absent-polarity.log`, 444행) · 렌더-루프 가드 FAIL exit 1(`RED-f1d-render-polarity.log`, 오버레이 실험 B 재채득 — 커맨드·exit·트리 상태 헤더 포함) · 뮤턴트 4종(전부-default-off·전부-default-on·극성 반전·렌더-only) 각 RED 확인 후 복원.
- **최종 검증**: `go test -count=1 ./internal/settings/... ./internal/web/...` ok ×4 · `golangci-lint` 0 issues · `gofmt -l` 무출력 · 빌드 darwin/windows exit 0. settings 커버리지 90.2%.
- **감사**: plan-audit iter2 PASS 1.00 · sync-audit FAIL(81.3, F1) → F1 수리 → 델타 FAIL(F1-D) → F1-D 수리 → 한정 재감사 **PASS**(91.3) → minor 2건 위생 커밋(`0b9f881b0`).

## Baseline-attribution

본 트리 HEAD `0b9f881b0` (커밋 13건: `3b11b3ed9` draft → `4e94f9607` 0.1.1 → `2031ccf7f` audit-ready → run 4건 → sync 2건 → F1/F1-D 수리 2건 → 위생 1건 + 본 문서). RED 측정의 working-tree 중간 상태는 커밋 불가 특성상 각 로그 헤더와 progress.md 트리 귀속 노트로 기술.

## Gaps

- **CI 전체 스위트 판정 미관측** — develop push는 리드 일괄 소관. 로컬은 스코프 테스트(darwin)만 측정됐다. darwin/windows 크로스빌드는 통과했으나 windows 테스트 실행은 CI 몫.
- 실물 브라우저(헤드리스 외부 프로세스) 재현은 미수행 — httptest 전체 루프 + 격리 `/tmp` 실서버 기동으로 대체.
- `effort` 소실 2건의 기제는 미해결로 기록(progress.md §E.2).
- mcp.yaml 부재 시 Save 500(`atomicWrite` os.Stat, `yamlpatch.go:190-193`) — 신규 관측, 본 카드 미수리.

## Residual-risk

- `TestManifestHashFormat`는 본 base `0b1e27877`에서 적색이나 develop tip에서는 수리됨(`072bc9f60`, t526) — develop 흡수 시 소멸 예정.
- 감사자 잔여 노트: absent default-on 렌더 극성 발산은 본 카드에서 수리됐으나, 렌더 name-유일성 고정 테스트 부재는 후속 카드 후보로 남는다.
- post-close 수리 커밋 4건(`146372d4b`·`a8042c630`·`6639e7ecb`·`0b9f881b0`)이 sync 커밋 이후 트리를 움직인다 — CHANGELOG·§E.4 기술은 수리 후 상태와 일치하도록 정정·보강됐으나(t500 선례), 최종 일치 판정은 develop CI가 한다.
