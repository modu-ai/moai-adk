# Decision Index — SPEC-V3R6-DOCS-PREREWRITE-FLAG-001

Authoring: 2026-10-19, manager-spec, card t1726 (plan phase). 부모 인덱스(SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001, decision-index.md revision 0.3.1)가 Q18의 완전한 행을 이 SPEC으로 넘겼다(부모 스텁이 이 SPEC을 소유자로 이름 붙인다). 이 파일은 세 행을 담는다. Q18(이 SPEC 고유의 결정, 열림)과, 이 SPEC의 파일에 적용되는 공통 행 Q13·Q17 — 둘 다 운영자 결정 기록 d-20261019T015908Z-220a(2026-10-19, 리더 전달)를 실은 적용본이다.

라벨 규칙: DECIDED와 POLICY-COVERED는 권위 등록부(authority register)의 커밋된 앵커를 요구한다. 아래 어느 행에도 그런 앵커는 없다. d-20261019T015908Z-220a 보드 기록은 운영자 쪽 기록으로 커밋된 트리에서 읽을 수 없으므로 여기서 핀할 수 없고, 핀할 수 없는 보드 인용은 행을 DECIDED나 POLICY-COVERED로 다시 붙이지 못한다. 결정 기록이 실린 두 행은 부모 인덱스의 판정 기록 관례(그 Q20·Q24 행)를 따라 FOUNDER 라벨에 `Operator verdict:`를 채운 형태로 남는다.

### Q13: 각 낡은 인용은 삭제되는가, 치환되는가?
Label: FOUNDER
Authority anchor: 커밋된 트리에 없음(운영자 보드 기록 d-20261019T015908Z-220a — 이 트리에서 읽을 수 없어 여기선 핀 불가)
Why unresolved: 부모 인덱스의 공통 행(부모 decision-index Q13)이며 각 자식이 자기 파일에 적용한다. 2026-10-19 운영자 결정으로 답이 정해졌고, 이 행은 이 SPEC의 파일에 대한 그 답을 기록한다.
Class: implementation-level
Operator verdict: 교체가 될 현재 문서·명령이 존재하면 인용을 현재 메커니즘을 가리키도록 REPLACE하고, 존재하지 않으면 해당 행을 DELETE한다 — operator ruling d-20261019T015908Z-220a (2026-10-19), 리더 전달

이 SPEC에의 적용: STALE-1(README 루프 인용)과 STALE-2(factory-mode relaunch 셀 서술) 양쪽 모두 현재 교체 대상이 존재한다 — 통합 `moai todo --auto` 엔진이 큐를 소비하는 단일 레인 세션(`internal/cli/factory_card.go`의 `factoryRelaunchSupersededNote`, card t1554). 따라서 기본 처분은 둘 다 치환이다.

### Q17: 같은 편집이 en, ja, zh에도 같은 변경으로 들어가는가?
Label: FOUNDER
Authority anchor: 커밋된 트리에 없음(운영자 보드 기록 d-20261019T015908Z-220a; `.moai/docs/docs-site-i18n-rules.md` 61·65행의 4 로캘 의무도 권위 등록부 밖이므로 POLICY-COVERED의 앵커가 못 된다)
Why unresolved: 부모 인덱스의 공통 행(부모 decision-index Q17); 로캘 파일을 바꾸는 각 자식이 자기 파일에 적용하고 행 사본을 품는다. 2026-10-19 운영자 결정으로 답이 정해졌다.
Class: implementation-level
Operator verdict: ko/en/ja/zh 네 로캘을 같은 카드·같은 변경으로 함께 고친다 — 로캘별 분산 편집 없음; 런 페이즈 커밋 하나로 충분 — operator ruling d-20261019T015908Z-220a (2026-10-19), 리더 전달

### Q18: 각 `--factory-run`·`--clear-policy` 인용을 유지하되 help에 플래그를 추가하는가, 제거하는가?
Label: FOUNDER
Authority anchor: 커밋된 트리에 없음
Why unresolved: 부모 revision 0.3.0에서 이 SPEC으로 넘어온 행(원문 commit bdb7b764). help-text 추가 선택지는 `internal/`·`cmd/`를 바꾸므로 이 문서 전용 SPEC의 범위 밖이고 별도 SPEC이 필요하다. 운영자 결정 d-20261019T015908Z-220a는 Q13과 Q17만 정한다 — Q18은 이 결정으로 정해지지 않았으며 이 행은 열려 있다. 이 행의 `Operator verdict:`가 채워질 때까지 자율 Kickoff는 막힌다 — 판정은 리더가 운영자에게 라우팅하며, planner 기본값 (a)는 잠정 계획 입력일 뿐 자율 진입 승인이 아니다(iter-1 D1, plan.md F.1 M1 선결 조건과 동일 문구). GATE(card t1717, 계획 전 — SPEC-V3R6-DOCS-PREREWRITE-GATE-001은 `.moai/specs/`에 아직 없고 그 플랜 페이즈가 이 참조를 소유한다)가 이 답을 입력으로 받는다(iter-1 D6 표기).
Class: product-level
Options: (a) 대체된 `--clear-policy relaunch` 카드별 소비를 서술하는 인용만 정리하고 살아 있는 동작을 서술하는 인용은 모두 유지한다; (b) 두 플래그의 인용을 살아 있는 것까지 모두 제거한다; (c) 인용을 모두 유지하고 help 텍스트에 플래그를 추가한다(이 SPEC 범위 밖 — 별도 SPEC).
Planner default applied: (a) — 잠정값이며 판정이 아니다
Default: (a) (rule: 더 작은 사용자 가시 표면 — 살아 있는 두 플래그의 정확한 인용을 모두 보존하고 죽은 정책의 서술만 없앤다; 되돌리기는 이 SPEC 자체 문서 커밋의 단일 revert로 끝난다)
Alternate: (b)
Operator verdict:
