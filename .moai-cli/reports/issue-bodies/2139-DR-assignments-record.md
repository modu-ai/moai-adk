# [DR][03] 배정·진행 기록 파일(`assignments.json`)과 리더 감독 절차 (F-B·F-C)

## 배경
종합 보고서 §10.2 F-B·F-C, §10.4 self-claim: 에이전트가 미배차 Issue 목록을 로컬에서 갱신하고 claim한 뒤, 리더가 배정·진행 기록을 대조·감독한다. 지금은 `.moai-cli/manifests/github-issues.json`의 dispatch_journal·session_notes에 리더가 손으로 적는다. 기록 형식과 감독 절차를 정한다. 로컬 파일은 **같은 호스트 캐시·직렬화용**이며 lock이 아니다(소유권은 Issue 댓글 claim).

## 범위
- `.moai-cli/manifests/assignments.json` 스키마 문서(`docs/design-review/assignments-schema.md`) + 초기 파일: `{schema_version, refreshed_at, unassigned:[issue#], assigned:[{issue, agent, claim_ref:"<issue>#<comment-id>", epoch, stage: claimed|in_progress|pr_open|audited|merged|lost, worktree, pr, updated_at}]}`
- 갱신 규칙: `unassigned`는 원격(`gh issue list --milestone`)에서 재계산한 스냅샷(쓰기 주체 누구나, 충돌 시 원격이 정본); `assigned` 항목은 claim 댓글 id가 있을 때만 추가; `lost`는 최소 id 경합에서 진 기록
- 리더 감독 절차를 `AGENTS.local.md` § Lane loop 아래 신규 절 "Assignment record and supervision"으로: 5분 틱마다 ① PR 상태 읽기 ② claim 댓글 ↔ assignments 대조(불일치는 원격 우선, 기록은 Issue 댓글에) ③ 수리 필요 시 원 작성자 아닌 레인에 `assignee-hint` ④ 목록 보충
- 코드 변경 없음(스키마·문서·초기 JSON만). JSON은 `python3 -I -c "import json;json.load(open(...))"`로 검증

## 수용 기준
- AC-1: 스키마 문서의 모든 필드가 초기 JSON 예시에 존재하고 JSON 파싱 통과
- AC-2: "로컬 파일은 lock이 아니다 / 소유권은 claim 댓글" 문장이 스키마 문서와 AGENTS.local.md 절에 각각 있음
- AC-3: 현재 열린 Issue(RC1 #2014·#1984·#1975 + design-docs-r2 milestone 전부)가 초기 `unassigned`/`assigned`에 정확히 반영(원격 대조 명령·출력 PR 본문 기록)
- AC-4: 개인 절대 경로·비밀 0건, check set 실측 exit

## 사용자 결정
- 2026-11-09 00:3x: 로컬 배정 파일 + Issue 댓글 + 상태 변경으로 세션별 개별 진행(리더 판단: 중재는 댓글 id, 파일·label은 보조)
