# SPEC-DEEPIV-001 Acceptance Criteria

## AC-DEEPIV-001: Auto-trigger on ambiguous input
Given user runs `/moai plan "기능 개선"`
When clarity score is evaluated
Then score is >= 4 (ambiguous)
And Deep Interview Phase 0.3 starts automatically

## AC-DEEPIV-002: Skip on clear input
Given user runs `/moai plan "JWT refresh token rotation API 추가 with RS256 signing"`
When clarity score is evaluated
Then score is <= 3 (clear)
And Phase 0.3 is skipped, proceeding to Phase 0.5

## AC-DEEPIV-003: AskUserQuestion format
Given Deep Interview round starts
When AskUserQuestion is presented
Then there are exactly 3 recommended options + 1 free-text option ("Type your answer")
And first option is marked "(Recommended)"
And no emoji in question text

## AC-DEEPIV-004: Max rounds respected
Given Deep Interview in `/moai plan` context
When 5 rounds complete without reaching clarity <= 3
Then interview ends and proceeds to Phase 0.5

## AC-DEEPIV-005: Interview artifact created
Given Deep Interview completes with 3 rounds of Q&A
When results are saved
Then `.moai/specs/SPEC-{ID}/interview.md` exists with all Q&A pairs

## AC-DEEPIV-006: Project workflow integration
Given user runs `/moai project` on a new project
When Phase 0.3 Deep Interview runs
Then it replaces the old 4-question sequence
And max 3 rounds

## AC-DEEPIV-007: Skip flag
Given user runs `/moai plan --skip-interview "description"`
When processing starts
Then Deep Interview is bypassed entirely
And Phase 0.5 proceeds directly

## AC-DEEPIV-008: Free-text input works
Given user selects "Type your answer" option
When user provides custom text
Then the custom text is captured as the answer
And next round uses this answer as context
