# decision-index.md — SPEC-CODEX-LOCALMD-001

decision_gate: ON. 각 행은 Detect → Explain → Ask 구조이며 선호 답안을 내장하지 않는다. Implementation Kickoff Approval 전에 운영자 결정이 필요하다.

## R1 — 템플릿 §8 문구 vs SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001 M4 긴장

- **Detect**: run phase의 헬프 카피·계약 역전이 `internal/template/templates/AGENTS.md.tmpl` §8(문단 :268-273; "…are not forwarded to Codex." 문장 :273 — 본 트리 직접 판독)의 부정확성을 만든다. 그러나 M4는 템플릿 공개 콘텐츠에서 CLAUDE.local.md 참조 17건을 제거한 상태다.
- **Explain**: 선택지 — (i) CLAUDE.local.md를 명명하지 않고도 정확한 일반 문구로 교정, (ii) "파일이 존재하면 런처가 Codex 세션에도 주입한다"는 사실 문장으로 파일명 명시(M4 위반 재유입), (iii) 템플릿 무변경 + 런처 헬프에만 문서화. 리포 `AGENTS.md` §8은 템플릿 재렌더 산물이므로 (iii)는 리포 문서도 원상태의 부정확 문구를 유지하게 된다. sync-phase의 CHANGELOG·docs-site 4-locale 갱신은 어느 선택과도 무관하게 필요하다.
- **Ask**: 템플릿 §8을 (i)/(ii)/(iii) 중 어느 방향으로 처리할까요?

## R2 — `moai codex status` readout에 local-instruction 행 추가 여부

- **Detect**: `codex_readout` 6행에 local-instruction 상태가 없어 symlink 등 거부 조건이 런치 시도 전에는 보이지 않는다(research.md §6 GAP).
- **Explain**: 추가하면 거부 조건의 가시성이 생기지만 신규 readout 행 = 헬프/문서/테스트 표면 확장이고, 거부는 런치 오류로 이미 명시적으로 보고된다. in-scope(이번 SPEC) 편입과 별도 후속 카드 분리가 가능하다.
- **Ask**: readout 행을 이번 SPEC 범위에 넣을까요, 후속 카드로 남길까요, 아니면 만들지 않을까요?

## R3 — argv-overflow 방어 깊이 (TOCTOU 봉쇄 포함 여부)

- **Detect**: 요구되는 최소 방어는 인코딩 길이 사전 검사(launch 전 fail-closed)다. 별도로 Lstat→ReadFile TOCTOU 창이 존재하며(research.md C4 — `ReadFile`은 symlink를 따름), 파일 수가 2개가 되며 창이 2배가 된다.
- **Explain**: 선택지 — (a) size-check만 (기존 AGENTS.local.md와 동일한 노출 수준 유지, 변경 최소), (b) size-check + open-with-`O_NOFOLLOW`+fstat로 TOCTOU 봉쇄 (Unix 전용 빌드 태그 분기 필요 — Windows 대응 부담, E2 크로스 플랫폼 빌드 검증 대상). (b)는 기존 1-파일 경로까지 같이 봉쇄할지의 범위 문제도 함께 끌어온다.
- **Ask**: argv-overflow 방어를 (a) size-check만 할까요, (b) TOCTOU 봉쇄까지 포함할까요? (b)라면 기존 AGENTS.local.md 경로에도 동일 적용할까요?

## R4 — AC-LMD-012 LIVE 수용의 실행 환경·주체

- **Detect**: LIVE AC는 실제 codex 바이너리 + auth가 있는 별도 세션을 요구한다. CI 환경에서의 재현성이 보장되지 않고, auth 키는 로컬 전용이다.
- **Explain**: 코드 레벨 AC(001~011)만으로는 주입 내용의 실세션 도달을 관측하지 못한다(research.md Gaps — `developer_instructions`와 파일 discovery의 조합 관계 미검증). 선택지 — (i) 운영자가 로컬에서 직접 실행하고 증거를 SPEC 디렉터리에 반출, (ii) auth가 있는 레인이 run phase 마지막에 실행, (iii) LIVE를 별도 검증 카드로 분리. 어느 쪽이든 NOT_RUN이면 전체 판정 un-PASS라는 AC-LMD-012 조항은 유지된다.
- **Ask**: LIVE 수용을 누가 어느 환경에서 실행할까요?
