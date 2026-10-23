package cli

// m7_gate51_test.go — gate round 51:
//
//   - 51-2 [P2]: the checkpoint OUTLIVES the asset install — a transient
//     failure in a LATER post-step must leave the next run a resume route
//     (the resume-pending marker), and the marker retires on completion.
//   - 51-3 [P2]: the deployed LAYOUT is the interrupted --llm record — a
//     resume over a deployed GPT layout restores the gpt wiring instead of
//     recording the claude default.
