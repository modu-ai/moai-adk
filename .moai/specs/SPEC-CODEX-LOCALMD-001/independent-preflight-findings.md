# t1078 독립 사전검토 — 수정 전 기준선

## Claim

전용 plan-auditor 생성은 `agent thread limit reached`로 실패했다. 기존 독립 에이전트 `/root/t1074_spec_fast`의 읽기 전용 검토에서 아래 수정 항목을 받았다. 이는 전용 감사 PASS나 구현 승인 기록이 아니다.

## Baseline-attribution

- tree: `WT-codex-local-md @ 1e00e35f8` + 문서 변경
- spec.md SHA-256: `4e23712dadbfc8e2cee04868ae950fbbfec19f3aff7c427a2d9db303664c4179`
- plan.md SHA-256: `d3ff0b28e7f33ff3cf6ae850f5458ea2edc6ad0c5e0d672d590ea65ed903a128`
- acceptance.md SHA-256: `c2089a7807082e1128b08c39c0ed284f4d30c5be882410bf72c677b5edea8ba1`
- 검토자가 최초·종료 해시 일치를 보고했다. 이후 수정본에는 이 결과를 그대로 적용하지 않는다.

## Findings

| ID | 수준 | 위치(수정 전) | 필요한 수정 |
|---|---|---|---|
| F1 | High | spec.md:50, acceptance.md:64 | provenance가 포함된 전체 payload와 원본 body를 동일하다고 비교하지 않고 body slice 또는 명시적 전체 기대값을 검증 |
| F2 | High | spec.md:100,107, plan.md:75 | 런처/LIVE의 입력 비가공과 sync 문서 갱신을 구분; 사용자 로컬 입력은 모든 단계 보존 |
| F3 | Medium | spec.md:72, plan.md:64, acceptance.md:94 | `Codex-only` 단어 자체를 금지하지 않고 AGENTS 전용·CLAUDE 추가 적재를 정확히 설명 |
| F4 | Medium | acceptance.md:57 | Unix-only와 Windows symlink 실행 요구의 모순 제거, 환경별 증거/Gap 정의 |
| F5 | Medium | acceptance.md:100, plan.md:44 | LIVE 재현 명령·timeout/cleanup·nonce/출처 성공 판독 조건 고정 |
| F6 | High | progress.md:5, decision-index.md:5, spec-compact.md:6 | 이전 감사 PASS와 수정본을 구분하고 결정 기록·축약본을 현재 본문에 정합 |

F1~F5는 작성자에게 delta 수정으로 전달했다. F6의 ancillary 정합 작업도 진행 중이다. 모든 항목은 수정 후 재확인 전까지 미해소다.

## Evidence

검토자가 반환한 명령과 결과:

```text
git diff --check -- <세 본문>
stdout 없음, exit 0

GOCACHE=/tmp/t1078-spec-audit-cache go run ./cmd/moai spec lint SPEC-CODEX-LOCALMD-001 --strict --json
[]
exit 0
```

형식 lint 통과는 의미 모순의 부재를 증명하지 않는다.

## Gaps

전용 plan-auditor 판정, 수정 후 delta 재검토, 구현, 실제 Codex LIVE, Windows 런타임 검증은 이 기록에서 입증하지 않았다.

## Residual-risk

파일 수정이 진행 중이므로 행 번호는 위 해시의 수정 전 문서에만 유효하다. 독립 검토의 발견 목록이 모든 결함을 포괄한다고 주장하지 않는다.
