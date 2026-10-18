# Decision Index — SPEC-TUXIU-GOLDEN-REFRESH-001

카드 t1681 지시에 따른 기초 질문 Q1-Q4와 plan 조립 중 표면화된 Q1건. 각 행은 Detect → Explain → Ask를 따르며 판단 호출에 대한 내장 추천을 지니지 않는다. `Default:`는 공표된 기본값 규칙의 적용(정책 적용)이지 추천이 아니다.

### Q1: U-08 archive 요약 제거를 AC-016 데이터 불변 계약에 어떤 방식으로 편입하는가 — 문서화된 예외 목록(정규화 계층 처리)인가, 두 기준선 재기준선인가?

Label: DECIDED
Authority anchor: `internal/cli/tuxiu_characterization_test.go:14-24` ("Immutable" 선언 + 두 기준선 설계 문서)와 `.moai/specs/SPEC-UPDATE-REPAIR-RC30-001/acceptance.md` AC-RC30-09 행(`--include='*.go'` 판정 범위) — 커밋 트리에서 확인 가능.
Why unresolved: (해소 — 본 행은 결정 기록) 두 기준선 모두의 재기준선은 현재 바이너리가 logo-era 프레젠테이션을 출력하기 때문에 pre-logo 기준선의 존재 이유를 소멸시켜 구조적으로 불가능하다. 따라서 예외는 (a) 계약 문서에 기록되고, (b) pre-logo 디스크 불변을 유지한 채 (c) DATA 필터의 정규화 계층에서 처리되며, postm4 집합만 재생성된다. archived 스킬 수 자체는 예외 대상이 아니며 exact-line 단정으로 계속 보호된다.
Operator verdict: DECIDED 2026-10-19 — spec.md §A.2/§B REQ-TGR-001/003/008로 편입. 예외 제2 항목(archive 부분집합)은 Q6에서 추가 결정.

### Q2: 저장소 내 재생성 하네스의 형식 — 환경변수 게이트형 go test + make 타깃 래퍼인가, 독립 스크립트(testdata 인접 셸/pty 스크립트)인가?

Label: FOUNDER
Class: implementation-level
Authority anchor: 없음 — 형식 선택을 결정하는 선완재 SPEC·설정·헌장 조항이 없음(M4 스크래치패드 하네스는 유실되어 그 형식을 앵커로 쓸 수 없음).
Why unresolved: 카드는 재생성 명령의 존재와 make 타깃 형태(`tuxiu-golden-capture`)를 정했지만 내부 형식은 정하지 않았다. tty 변형의 실빌드+pty 제약(M4 기록)은 두 형식 모두에서 충족 가능하다.
Default: 환경변수 게이트형 go test + make 타깃 래퍼 (rule: preserves current behavior — 기존 `make tui-snapshot`의 `UPDATE_GOLDEN=1 go test` 패턴을 따르며 새 스크립트 계열을 도입하지 않음)
Alternate: testdata 인접 독립 스크립트(capture 전용 셸/pty 래퍼)
Operator verdict: DEFAULT-APPLIED 2026-10-19 manager-spec (lane-11)

### Q3: 재생성 대상 범위 — postm4 전체 12파일 재기록인가, stale 3파일만인가? (카드 "골든 3개"와 RC30 기록 6파일의 상충 해소 포함)

Label: EVIDENCE-NEEDED
Authority anchor: 해당 없음(증거 대기 행).
Why unresolved: 측정으로 이미 소멸한 부분 — 리터럴 보유는 6파일이고(pre-logo 3 + postm4 3, spec.md §A.1 실측), 내용 변경은 postm4 3파일이며, pre-logo 3파일은 예외 문서화로 유지한다. 남은 미결 증거: 하네스로 postm4의 update stderr 3 + init 6 파일을 재캡처했을 때 바이트 동일성이 유지되는지 — 이는 하네스가 존재해야 관측 가능하다(plan.md M3, AC-TGR-006). (iter-1 D2 갱신) 3 update stdout 파일은 literal 접미사 소멸에 더해 Q6의 archive DATA 편입으로 materially 변경이 확정 시사되므로, 수렴 기대는 위 9파일로 한정된다.
Operator verdict: (비움 — run 단계 M2/M3 실측으로 소거)

### Q4: 제거된 리터럴에 대해 어느 쪽이 옳은가 — 코드인가 골든인가?

Label: DECIDED
Authority anchor: `.moai/specs/SPEC-UPDATE-REPAIR-RC30-001/spec.md` REQ-RC30-011 행("Every count and status line ... shall be a value measured in the same run. No literal count string shall appear in a format string. Items: U-08, U-18")과 acceptance.md AC-RC30-09 행 — 커밋 트리에서 확인 가능.
Why unresolved: (해소 — 확정 전제로 운반) AC-RC30-09의 판정으로 코드 측(measured-count-only 출력)이 옳고 골든이 stale임은 이미 결정되었다. 본 SPEC은 이 판정을 여는 질문으로 다루지 않는다(spec.md §A.3, Out of Scope — 프레젠테이션 코드 변경).
Operator verdict: DECIDED 2026-10-19 — 확정 전제로 spec.md §A에 기록.

### Q5: 재생성 명령에 no-write verify 모드(커밋 골든 vs 현재 바이너리 출력 diff)를 포함하는가?

Label: FOUNDER
Class: implementation-level
Authority anchor: 없음 — verify 모드 포함 여부를 결정하는 선완재 SPEC·설정·헌장 조항이 없음.
Why unresolved: 카드의 납품물은 재생성 명령 정의와 실행이며 verify 모드는 부가 기능이다. 포함 시 현행 4테스트의 골든-대-골든 성격이 코드-대-골든으로 확장되는 범위 판단이 따른다.
Default: capture-only — verify 모드 미포함 (rule: preserves current behavior — 현행 테스트 구조를 변경하지 않으며, 되돌리기는 본 SPEC 커밋의 단일 revert)
Alternate: capture + verify 모드(`tui-snapshot-verify` 선례의 형제 타깃)
Operator verdict: DEFAULT-APPLIED 2026-10-19 manager-spec (lane-11)

### Q6: 기록 골든의 archive DATA(부활 3 ID)는 현재 바이너리가 재현할 수 없다 — 재생성 시딩과 AC-016 비교 범위를 어떻게 재결정하는가?

Label: FOUNDER
Class: implementation-level
Authority anchor: 없음 — archive DATA 불일치의 처리를 결정하는 선완재 SPEC·설정·헌장 조항이 없음. 근거 레코드: `internal/cli/update_archive.go:36-39`(세 부활 ID의 `legacySkillIDs` 제외 명시, SPEC-UPDATE-LEGACY-SKILL-LIST-001 참조)와 골든 `postm4/update.notty.stdout.golden:57-59`(`✓ archive: moai-domain-backend/frontend/database`) — 커밋 트리에서 본인 실측으로 확인(2026-10-19; plan-audit iter-1 D2와 일치).
Why unresolved: 기록 baselines의 archive DATA는 부활로 제외된 3 ID를 지니므로 어떤 시딩으로도 교차-일치할 수 없다. `legacySkillIDs` 수정은 프로덕션 코드 변경으로 AC-RC30-09 동결 하에 경로가 아니고, 양 기준선 재기준선은 Q1과 동일하게 구조적으로 불가능하다. 남은 선택은 시딩 유무와 AC-016 비교 범위의 조합이다.
Default: 시딩 + 교차-기준선 예외 제2 항목 — 하네스가 `legacySkillIDs`(`internal/cli/update_archive.go:45-71`)에서 선정한 archive-대상 2 ID를 시딩하고, 재생성 postm4는 실측 archive DATA(`total: 2 skills archived` + 시딩 ID의 `✓ archive:` 행)를 기록하며, archive 부분집합(항목 행 + 요약 행)은 pre-logo/postm4 교차 비교에서 제외된다. 이후 archive DATA는 원시-라인 exact-line 단정(REQ-TGR-007)으로 계속 보호된다(요약 행과 시딩 ID 항목 행 모두 — 항목 행 삭제 mutant도 단정에서 실패해야 한다). (rule: preserves current behavior — archive DATA가 골든에 기록되고 단정된다는 현행 가드 의미를 유지)
Alternate: 무시딩 + 범위 축소 — `total: 0 skills archived`를 있는 그대로 기록하고 archive 부분집합을 교차 비교와 단정 양쪽에서 완전히 떨어뜨린다(가드 대상에서 archive DATA 소멸).
Operator verdict: DEFAULT-APPLIED 2026-10-19 manager-spec (lane-11)
