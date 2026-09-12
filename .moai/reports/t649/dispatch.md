# t649 작업 기록

- card: t649
- source_session_id: 01a08e7b-6aa0-7361-ab7e-ea8da1f02228
- SPEC: SPEC-MOAI-GATEWAY-001
- 작업 트리: `.claude/worktrees/moai-proxy-unified`
- 브랜치: `WT-unified-gateway`
- 시작 HEAD: `81c1d58f9`
- 상태: 착수. 구현 완료 및 실사용 검증 판정은 아직 하지 않았다.

## Claim

사용자의 “카드 발행해서 진행해줘” 지시에 따라 기존 gateway 후속 작업을 t649로 발행하고 착수 상태로 기록했다. 기존 미커밋 변경을 보존하며 이어간다.

## Evidence

`moai todo add '<승인된 gateway 제공자별 모델 제한·reasoning 왕복·실사용 검증 범위>' --pick` 출력:

```text
picked t649 [GATEWAY-PROVIDER-MODEL-20260912] 기존 moa...
```

`moai todo next t649 --spec SPEC-MOAI-GATEWAY-001` 이후 `moai todo list --json`에서 해당 카드의 id/state/spec_id를 추출한 출력:

```json
[{"id": "t649", "state": "picked", "spec_id": "SPEC-MOAI-GATEWAY-001"}]
```

`git -C .claude/worktrees/moai-proxy-unified rev-parse --short HEAD`:

```text
81c1d58f9
```

`git -C .claude/worktrees/moai-proxy-unified branch --show-current`:

```text
WT-unified-gateway
```

`/tmp/moai-gateway-current gpt status`:

```text
GPT: logged in
```

## Baseline-attribution

위 출력은 2026-09-12 이 세션에서 직접 확인했다. 카드 상태는 저장소의 공식 `moai todo` CLI로 조회했다. 기존 바이너리의 로그인 상태는 새 구현의 모델 응답 성공 증거가 아니다.

## 작업 범위와 소유

- 구현 담당: gateway·관련 CLI·회귀 시험. 기존 코드 변경을 보존한다.
- SPEC 담당: 새 제공자 제한 계약 및 인수 조건을 기존 SPEC 문서에 반영한다.
- E2E 담당: 카드 보고서 폴더 안의 실제 Claude Code 실행용 검증 도구를 준비한다.
- 영실: 작업 조정, 증거 확인, 새 빌드 이후 실제 답변·도구·재개 판정.

## Gaps

아직 이 카드의 최종 빌드, 실제 GPT-6 Astra 답변, 도구 호출, 재개 성공을 판정하지 않았다. 세 launcher의 실제 제품 경로와 모델 선택 목록을 모두 검증해야 한다. Windows 실행 증거는 승인된 GitHub CI 경로에서 별도로 확인해야 한다.

## Residual-risk

카드의 `picked`는 작업 착수만 뜻한다. 목록 제한만으로 요청 제한이 보장되지 않으므로 gateway에서도 제공자를 검증해야 한다. reasoning carrier의 형식 검증만으로 소유권을 인정하지 않으며, 승인된 세션과 receipt에 결합해야 한다. push·PR·병합·워크트리 제거는 별도 지시 범위로 유지한다.
