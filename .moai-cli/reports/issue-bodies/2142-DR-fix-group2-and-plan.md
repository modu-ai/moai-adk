# [DR][06] 하위 6편 수정안 — Core Brain·SYNC·실행 소유권·검증 감사·판단 backend + RC1 온라인 우선 최종 계획

## 배경
종합 보고서 §3 (`review-group2-core-brain-sync-owner.md` ①~⑤, `review-group1-rc1-plan.md`). 대상: [136c] Core Brain, [0d45] SYNC·Issues 보고, [0259] 실행 소유권·복구, [c4f9] 검증 감사·보안, [e361] 판단 backend·이미지, [43fa] RC1 온라인 우선 최종 계획. ChatGPT Space 페이지 → **적용 가능한 수정안 문서**.

## 범위
- `docs/design-review/fixes/group2-<pageid>.md` 5개 + `fixes/plan-43faf222.md`: 각 검토의 "개선 제안(영향 순)" 전부를 (위치, 현재, 교체, 근거) 표로. 우선: Core Brain L4436-4450 TOON 결론(S6 — 두 안(제외/채택) 모두 교체문 제시, 결정은 사용자), [0259] L69·[43fa] L3781 claim 권위(S1), [43fa] L5164 tag·§20 경로·§18 명령(S2·S3·S11), [0d45] 효과 상태 어휘·outbox 경로, [c4f9] preflight·verdict 형식, [e361] vendor 정책 날짜·URL(없으면 "미결로 환원" 교체문)
- 근거 행은 현재 tip에서 재확인
- 입력: `references/chatgpt-space/page_{136c8203…,0d4538a2…,0259e595…,c4f994be…,e3612b0b…,43faf222…}.md` 읽기 전용(43fa는 완전본 L1753 이후 행 기준)

## 수용 기준
- AC-1: 6편 모두 ≥ 5항목, (위치, 현재, 교체, 근거) 채움, tip SHA 기록
- AC-2: S6 TOON은 결정 대안 2개를 각각 완성문으로; 사용자 결정 전 단일안으로 적지 않음
- AC-3: 외부 사실(vendor credit·Decisions GA)은 공식 URL+확인일을 붙이거나 "미결" 교체문 — 모델 기억으로 채우지 않음
- AC-4: 가설 근거 단독 0, 개인 절대 경로·비밀 0건, check set 실측 exit

## 사용자 결정
- 2026-11-09 00:1x 보고서 컨펌
