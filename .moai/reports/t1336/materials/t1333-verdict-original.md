# t1333 verdict — codemaps refresh9 (standing [GRAPH] source)

## Claim
described-source-diff 44(>= 40) 확인, codemaps 4문서(overview·modules·entry-points·dependencies) 갱신, provenance stamp 145c3d98c 재찍음 → graph check codemaps 층 fresh(0 < 40).

## Evidence — before
- ./bin/moai graph check (인트리 바이너리 Commit=145c3d98c): codemaps value=44 threshold=40 verdict=stale — 앵커 afecf81e9e96(t1257 문서층 병합판)
- git diff --name-only afecf81e9..HEAD -- internal cmd pkg | 필터 = 44 (IsDescribedWorthy 동일 집합)

## Evidence — refresh
- overview.md: 단락 갱신 + 규모 표 비테스트 1422→1429·테스트 2545→2569, 도식 절 internal/hook 스테일 141→155 정정
- modules.md: 단락 갱신 + 행 수치 cli 372→375(루트 295→298)·todo* 15→17·doctor* 23→24·kanban 57→61·나먲지 78→67 산술 재산정 + 6행 근거 절
- entry-points.md: 단락 갱신 + AddCommand 230→231(rootCmd 66·init 30·훅 엄트리 34 불볇확인)
- dependencies.md: 단락 갱신 + 작은 fan-in 표 슠테일 정정 — auditreceipt 2→3·jev 2→3·jevcred 2→3·contract 1→3 (신귴 소비자 엠키는 git grep으로 앵커 시점 존재 확인 — 스테일 값의 바로잠펴, 신귵 엠키 아님)
- 불변 재확인: 패키지 162·최상위 82·엣지 449/277·임베드 598·fan-in/out 상위 표 전 행·숴환 4쌍·go.mod/go.sum 무변경
- ./bin/moai graph stamp codemaps --commit 145c3d98c → graph check: codemaps value=0 verdict=fresh

## Baseline
카드 트리 .moai/worktrees/t1333(브랿치 WT-codemaps-refresh9, base 145c3d98c)에서 이 실행으로 측정. 측정 바이넄리 = make build 산물(Commit=145c3d98c) — installed lag 무관(2.2).

## Gaps
- mx-index·edges 층: 미추적 런타임 산출물 부재(absent) — fresh 워크트리 상태로 expected, stale 아님. CI 통 팠택은 develop push 후 리드 소관.
- stamp 커밋 145c3d98c = 카드 커밋 직전 팕. 카드 커밋 본체(.moai/project 변경)은 described roots(internal·cmd·pkg) 밖라 diff 0 유지 — 재차 stale 되지 않는다.

## Residual-risk
- 본 세션 GLM 출력 채넍에서 한국어 자모 오염이 반복 관측됨(큜/경쟁/폴백 등 실재-하지만-엉터닐 읠 블록) — 문서 산문은 전수 unicodedata 자모명 검증(ASCII 채넍)으로 수리했으나, 같은 채넎이 쓛 다른 한국어 산문(카드 보고 포함)에 잔여 가능성. GLM 채넞 한국어 무결성 카드 후보.