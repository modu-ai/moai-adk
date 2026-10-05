# S5 — acceptEdits 정규화 고지문 현지화 증거 (REQ-TRI-006 / AC-TRI-007)

- 수리 전(베이스): 고지문은 영어 고정 — `internal/cli/profile_setup.go:31` 의 단일 상수.
- 수리 후: `emitAcceptEditsConfirmation(out, locale)` 이 위저드 종료 로케일로 출력. 미지 로케일은 영어 폴백.
- 앵커 토큰 계약(REQ-CCI-006, 소유 SPEC: SPEC-V3R6-CLI-CONFIG-INTEGRITY-001):
  `acceptEdits` · `settings.local.json` 토큰은 4개 로케일 모두 번역문 안에 **그대로 보존**된다.
- 검증 명령: `go test ./internal/cli/ -run 'TestEmitAcceptEditsConfirmationAnchor' -count=1 -v` → PASS (run-phase M3, 이 트리)
- 호출부: `runProfileSetup` 이 `result.ConversationLang` 을 전달 (profile_setup.go 수리 커밋).

## 로케일별 출력 (acceptEditsConfirmationTexts, 그대로 인용)

| locale | 고지문 |
|---|---|
| en | `Note: "acceptEdits" is the project default, so no settings.local.json defaultMode override will be written.` |
| ko | `참고: "acceptEdits"는 프로젝트 기본값이므로 settings.local.json에 defaultMode 재정의를 기록하지 않습니다.` |
| ja | `注意: "acceptEdits"はプロジェクトのデフォルトのため、settings.local.jsonにはdefaultModeの上書きを書き込みません。` |
| zh | `注意: "acceptEdits"是项目默认值，因此不会向 settings.local.json 写入 defaultMode 覆盖。` |

ko 프레임 잔존 판정(AC-TRI-006): en 원문 단편(`project default`, `will be written`)은 ko/ja/zh 출력에
존재하지 않음 — `TestEmitAcceptEditsConfirmationAnchor` 의 enResidues 단언이 기계적으로 잠근다.
