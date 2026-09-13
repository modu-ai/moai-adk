# t748 plan-audit — SPEC-CODEMAPS-FOLD-GUARD-001 · iter 1/1 (Tier S ceiling S=1, harness.yaml:76)

판정: **PASS** · 점수 **0.86** (조화평균, Tier S 임계 0.75) · 측정 트리: worktree t748 @ `146faed9d` (재확인) · plan-auditor(opus/high) 수행
audit_model 미설정 — Claude 전용 감사. 감사관이 디스패치 전달 사항을 근거로 취급하지 않고 적대 주장으로 재측정(작성자 신뢰 없음, 전부 자기 실행 명령으로 검증).

## 필수 통과 (7/7)

- MP-1 REQ/AC 번호 일관성 PASS — REQ-CFG-001~005·AC-CFG-001~005 연속 무갭 (spec.md:110-118, :124-138)
- MP-2 GEARS 형식 PASS — 5개 REQ 전부 M3 패턴 준수(라벨 니트 D3)
- MP-3 프론트매터 PASS — 12 정식 필드·거부 별칭 없음·era/tier/depends_on 공인 옵션; Artifact Statelessness 준수; 기계 교차검증 `spec_audit` → modern_era_clean 1·drift 0
- MP-4 언어 중립성 N/A — 단일 리포지토리 Go 가드
- MP-5 D7 교차-SPEC PASS — 참조 5 SPEC 전원 status: completed, reconciliation 의무 없음
- MP-6 D8 교차 플랫폼 PASS — syscall 0, 부정 주장+E2 GOOS=windows 면제 명시
- MP-7 확인 게이트 PASS — NEEDS CLARIFICATION 0, research.md 불요

## 카테고리

명확성 **1.0** · 완전성 0.75 · 테스트 가능성 **1.0** · 추적성 0.75

## 감사 질문 6개 응답 요지 (전부 자기 실행 검증)

1. 보호 집합 이중 잠금 건전 — fold-judgments.txt 데이터 주도(재측정 fold 5/omission 15) + 5단위 floor 상수(기록 편집으로 축소 불가, floor-행 손실→FAIL); 손-열거 재발(REFRESH-002 plan-audit D1) 차단.
2. 위반 규약 거짓-부정 — 사고 서명(전체 경로 표기 6곳, t475 §④-b:558)은 잡음; `core/git` 단축형 예외가 살아있는 fold 상태 보존(overview.md:103·modules.md:171 문자 확인); 잔여(토큰 없는 재작성)는 진실하나 미명시 → D5.
3. 스탬프 독립성 — 숨은 의존성 0; t688 G3 확인(fold-judgments.txt·provenance.json 모두 생성기-5 집합 밖 → 재스탬프가 가드 입력 미접촉); graph check 구조적 사각 주장 일관 유지.
4. RED 비공허 — Green 경로가 기본 스위트의 실제 `.moai/project/codemaps/`를 매번 실행(실제 표면), fixture는 RED 준비 전용, frozen 표면 불손상(M2 git-status 변조-0 확인); 잔여: 단일-판독기 고정 누락 → D4.
5. Tier S 정당화 검사 통과(테스트 1파일·프로덕션 0줄·8/8 천장 내) + spec_audit 기계 교차검증.
6. 3축 실제 분리 — 축1 기록(AC-CM2-007 문자 확인):163-165+spec.md:236, 축2 lint(lint_req_widen.go 6 미캡처 형상 :15-20·패턴 :59) 판정 기반 미수입+수리 범위 밖, 축3 비침묵(verification-completeness §1.1·t747 F5 검증) — 축 붕괴 없음.

## 결함

- **D1 [경미·차단]** generator-doc 무결성 미명시 — REQ-CFG-002(:112)는 토큰 포함만, REQ-CFG-003(:114)은 기록 층만 커버; 5개 명명 문서 중 하나가 부재/판독 불가면 처리되지 않은 입력 → 읽기 오류를 통과로 처리하는 구현은 공허 green(축-3 비침묵 원칙의 빈 스윕 형태). 수정: REQ-CFG-002 또는 003에 절 추가 — 5개 생성기 문서 각각 존재·판독 가능, 위반 시 FAIL(plan §D 루트-미스 실패 금지의 문서 수준 확장). manager-spec 경유, run 전.
- D2 [선택] AC-CFG-001/003/004의 REQ 참조가 실질적만(토큰 없음) — AC 헤더에 REQ 토큰 추가.
- D3 [선택] "Event-detected" 라벨 → 정식 "(Event-driven)".
- D4 [선택] plan M1.4에 단일-판독기 문장 추가(기본 스위트와 주입 실행이 동일 함수 경로 — AC-CFG-002 RED 증거 전송의 기반).
- D5 [선택] 탐지 한계 명시(토큰 없는 산문 재작성은 규약 회피 — 잔여 수용 가능하나 정직성 문장 요구).
- D6 [선택] §D 전문 열거에 AC-CFG-003 누락 → 002/003/004/005로 수정.

## 권장

**PASS.** 라우팅 2건: (1) D1 — run 전 manager-spec 경유 절 수정; (2) D2-D6 — 같은 편집으로 무료 수리. Kickoff 게이트는 본 판정 영향 없음(리드 소관).

— 보고서는 plan-auditor 전달문을 lane이 `.moai/reports/t748/plan-audit.md`로 보존(감사자 미작성 규약).
