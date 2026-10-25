# t1959 card-review

## Round 1 — PASS @ c1c9c26d98

- backend: codex (`moai verify codex-review --project-root …/worktrees/t1930`, CLI 폴백 — 세션 도구 목록에 MCP `codex_review` 부재; fallback 사유: `mcp: server=<실행 서버> cli=<CLI> fallback=CLI`, 카드 인테이크 비교 규정 준수)
- head: c1c9c26d9891845fb1c11fb1a3df1a26ccb6b4ce · tree_digest 261ab9e326f0f6f2 · tool_version codex-cli 0.163.0 · exit_code 0 · verdict **pass**
- 검토 범위 (도구 자기 서술): 정적 코드 검토 — 테스트·빌드·재현 명령 미실행

### Findings

- 없음 (verdict pass, findings 미보고 — 243바이트 결과 JSON에 findings 필드 없음)

### Dispositions

- 처분 대상 없음. 레인 독립 검증은 별도 기록(커밋 전 gofmt·build·scoped test ok 21.569s·신규 3케이스 PASS·vet 0·lint 0 issues).
- 감사 상한: 리더 지시대로 독립 감사 1회(head 결속) — `.moai/reports/t1959/verdict.md`.

## Round 2 — FAIL (advisory) @ 309b2670b6

- backend: codex (`moai verify codex-review --project-root …/worktrees/t1930`, CLI 폴백 — MCP `codex_review` 부재, round 1과 동일 경로)
- head: 309b2670b6cc92e603dd8bee4e6a8162f952bbb9 · tree_digest 09d11b9f40f28e4a · tool_version codex-cli 0.163.0 · exit_code 1 · verdict **fail**
- F1 수리 커밋(레코드 SpecID 라우팅, RED-first 회귀 포함) 이후 재검토

### Findings

- **[P1] internal/cli/factory_merge.go:127** — SPEC-less `audited_sha`가 작업 트리 HEAD에 결속되는데 나머지 조건은 `--branch` 후보를 검사: 감사받은 커밋 체크아웃 + 미감사 브랜치 지정 시 후보의 충돌·트리·재측정 조건만으로 cleared+창 획득 가능. 처방: 감사 SHA를 후보 팁에 결속하거나 후보-감사 일치 확인. (정적 확인만 수행 — 테스트·빌드·재현 미실행, 도구 자기 서술)

### Dispositions

- 이 finding은 **독립 감사 iteration 1의 F3와 동일 클래스**(감사자도 codex P1-2로 동일 지적) — 리더 판정(10-25): "선천적 구조·기본 경로 도달 불가 → 수리 안 함, verdict Gaps/Residual에 그대로 두고 **후속 카드 후보로 회신**". 본 PR 수리 대상 아님 — r2 FAIL은 F3의 재확인(corroboration)으로 기록하고 레인은 리더 회신에 그대로 전달. 재검토 라운드로의 강제 복귀 없음(리더 처분이 선행).
- F1 수리 자체는 r2 findings에 포함되지 않음(수리 후 소멸 확인).
