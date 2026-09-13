# GPT AUTH broker 계획 감사

Iteration: 1 — AUTH wrapper 설계와 코어 연결 경계만 판정
Verdict: PASS
Overall Score: 1.00 (계획 범위 점수; Tier M 기준 0.80)

Reasoning context ignored per M1 Context Isolation.

## Claim

지정한 AUTH 계획은 mocked broker와 로컬 저장/송신 wrapper 구현을 준비할 수 있다. operation/state lock 분리, 세대 CAS와 logout tombstone, broker 종료 뒤 scratch 검증, refresh 증거 조건, first-header-write와 logout 직렬화 계약에서 구현 착수를 막는 결함은 찾지 못했다. 실제 broker·새 로그인·refresh·구독 endpoint와 transport callback 호환성은 아직 미검증이다.

읽기 전에 둔 실패 가설은 장기 operation lock이 logout을 막는 경우, 지연 결과가 tombstone을 덮는 경우, 실행 중 broker의 부분 파일 공개, RPC success를 refresh 성공으로 간주하는 경우, 헤더 쓰기를 멈추지 못했는데 logout 완료를 반환하는 경우, 기존 사용자 credential 복사, 코어 CredentialRef만으로 송신 경합을 막았다고 주장하는 경우였다.

## Must-Pass Results

- **MP-1 PASS:** REQ-GA-001~010이 연속이며 AC 10개가 같은 번호 REQ를 추적한다(E2).
- **MP-2 PASS:** `spec.md:31-49`의 요구사항은 When/While/Where 또는 The … shall 형식을 따른다. `acceptance.md:3-25`의 Given-When-Then은 검증 계층으로 판정했다.
- **MP-3 PASS:** `spec.md:2-15`에서 canonical frontmatter 12필드, `version: "0.1.0"`, `status: draft`, `tier: M`을 직접 읽었다. 현재 lint 출력은 E1.
- **MP-4 N/A:** 단일 Go 저장소 인증 wrapper이며 범용 언어 template 변경이 아니다.
- **MP-5 PASS:** 참조된 코어 SPEC은 현재 존재하며 상태는 `draft`다(E2). AUTH plan C가 인터페이스 연결과 추가 송신 경계를 구분한다. 형제 구현을 이미 완료했다고 전제하지 않았다.
- **MP-6 PASS:** AUTH spec의 syscall 언급은 0개(E2). `plan.md:40-44`는 Unix/Windows 잠금 구현을 분리하고 기존 Windows process-local mutex를 프로세스 간 보장의 근거로 사용하지 않는다.
- **MP-7 PASS:** AUTH plan의 unresolved clarification 표지는 0개(E2). runtime preflight·deadline·취소 전달의 후속 고정은 `plan.md:61`, `:69-75`에 명시된 구현/측정 과제다.
- **MP-8 N/A:** AUTH acceptance에는 release-blocking RED-now 인용 셀이 없다(E2 및 본문 판독). 예정된 실제 인증 시험을 이번에 재실행하지 않았다.

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|---|---:|---:|---|
| Clarity | 1.00 | 1.00 | `plan.md:10-19`, `:23-38`, `:52-61`: 소유권·publish·송신/로그아웃 완료 시점을 구분 |
| Completeness | 1.00 | 1.00 | `plan.md:25-38`, `:40-44`: 두 잠금·세대·tombstone·scratch·crash/실패·플랫폼 경계를 포함 |
| Testability | 1.00 | 1.00 | `acceptance.md:30-43`: 별도 프로세스, late completion, first-write 전/중/후, 멈추지 않는 SSE, timeout, RPC 거짓 양성의 판정 |
| Traceability | 1.00 | 1.00 | 10개 REQ와 10개 AC의 번호 추적 일치(E2), 코어 네 메서드와 추가 SendAuthorized의 명시적 연결(`plan.md:48-56`) |

이는 AUTH 계획의 제한된 검토 점수다. 제품 코드·공식 broker 내부 PKCE·실제 transport 안전성의 점수가 아니다.

## Defects Found

No defects found in the authorized plan scope.

## 세부 경계 판정

| 점검 대상 | 판정 근거 | 아직 확인하지 않은 것 |
|---|---|---|
| 기존 사용자 인증과 분리 | `plan.md:10-14`: 빈 로그인 scratch, refresh는 MoAI canonical만 seed, file backend 강제, keyring/사용자 저장소 fallback 금지. `AC-GA-009`와 `acceptance.md:42-43`이 읽기/복사까지 계수 | 설치본이 override·관리 정책을 실제로 적용하는 방식 |
| operation/state lock | `plan.md:25-26`: login/refresh는 operation lock, 짧은 state lock만 사용. `:33` logout은 장기 operation lock을 기다리지 않음 | 별도 OS 프로세스에서 잠금·취소·crash 실행 |
| CAS와 logout tombstone | `:30-35`: 시작 세대와 현재 세대 일치 조건, 증가 세대 원자 publish, 중간 logout 세대 변화 시 지연 결과 폐기. tombstone에서 새 명시 login을 허용하는 예외도 시작 세대 일치를 요구 | 실제 CAS/원자 저장 구현 |
| broker 종료와 scratch 공개 | `:16-18`, `:27-29`, `:37-38`: 종료 확인 후 읽기·형식/방식/계정/만료 검증, canonical 직접 쓰기 금지, crash scratch 재채택 금지 | 실제 broker 종료/notification/파일 형식 호환 |
| refresh 진실성 | `:27-29`, `acceptance.md:33-35`: 새 token/만료·공급자 수용 증거 요구, RPC success·last_refresh만으로 PASS 불가 | 새 로그인·실제 refresh의 공급자 수용 |
| logout와 첫 헤더 쓰기 | `plan.md:52-61`: state lock을 최초 request header write 완료/실패까지 유지. stream 전체에는 유지하지 않음. 종료 확인 불가 시 logout pending/error | transport callback의 실제 의미, deadline과 다른 프로세스의 취소/lease |
| 로컬/원격 logout | `:33-36`, `:58-60`: tombstone publish 전 로컬 완료 금지, in-flight는 취소 시도만, remote revoke 실패로 복원 금지 | 공식 broker logout의 실제 서버 revoke 의미 |
| CredentialRef 호환 | 현재 `internal/gateway/auth/credential.go:24-28`은 Provider/Generation/Apply/Redacted 네 메서드. `plan.md:48-56`은 이를 유지하고 추가 SendAuthorized 연결 없이는 완료하지 않음 | core transport 연결과 그 barrier 시험 |

네트워크 작업을 기다리는 login/refresh와 실제 첫 헤더 write를 기다리는 송신은 서로 다른 상태 잠금 구간이다. 후자는 양수 deadline으로 제한하고, 종료를 확인하지 못하면 완료를 주장하지 않도록 문서에 정했다. 이 계약이 실제 구현에서 지켜졌다는 판정은 하지 않았다.

## Evidence

### E1 — 현재 트리 구조 검사

작업 디렉터리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`에서 직접 실행했다.

```text
$ /tmp/moai-gateway-81c1d58f9 spec lint SPEC-MOAI-GPT-AUTH-001
✓ No findings — all SPEC documents are valid
```

Exit 0. 같은 HEAD에서 앞서 빌드된 바이너리이며 이번 감사에서는 재빌드하지 않았다.

### E2 — 정의·추적·현재 내용

Read-only `python3 - <<'PY'`에서 `^\*\*REQ-GA-(\d{3})\*\*`, AC 정의와 `(REQ-GA-NNN)` 헤더를 추출하고, 연속 번호·양쪽 추적·표지 개수·SHA-256을 출력했다. 관측 stdout(exit 0):

```text
REQ=10 sequential=True
AC=10 trace_equal=True
spec_syscall=0 plan_clarification=0 release_blocking=0
.moai/specs/SPEC-MOAI-GPT-AUTH-001/spec.md 40299c36ce0d05f1897c38329f319e148e86328a60beba5c67ba11266ed2f696
.moai/specs/SPEC-MOAI-GPT-AUTH-001/plan.md 4823044f533912e6ec2889a22d110f067751075cf9657a0addc4f8ccb14acc79
.moai/specs/SPEC-MOAI-GPT-AUTH-001/acceptance.md 3754a9aab2c77e960bc5050770d7d3afce51d465812754e2975945b64d9e517b
.moai/specs/SPEC-MOAI-GPT-AUTH-001/progress.md bb49dca33471fae059a32376375de4e979b09644572464dd0e7df93a066fca59
.moai/reports/SPEC-MOAI-GATEWAY-001/auth-broker-design.md e3a8d995de7bbbf4daa2174d6e2f06a6447baacbdc0463d803de19514d6ffdda
internal/gateway/auth/credential.go 36d3ae6e9030652908f2ae235cce020e8129d141b5b4159595524d6929103e64
core_status=draft
```

### E3 — 공개 source 기준선과 직접 판독

```text
$ git -C /tmp/openai-codex-audit.zbvNXB rev-parse --short HEAD
5a9eb14
```

Exit 0. 아래는 해당 checkout에서 `nl -ba`와 범위를 지정한 `sed`로 직접 읽은 코드의 발췌다.

```text
codex-rs/login/src/auth/storage.rs:214
        options.truncate(true).write(true).create(true);
codex-rs/login/src/auth/storage.rs:220-221
        file.write_all(json_data.as_bytes())?;
        file.flush()?;
codex-rs/app-server/src/request_processors/account_processor.rs:1021
        if do_refresh && let Err(err) = self.auth_manager.refresh_token().await {
codex-rs/app-server/src/request_processors/account_processor.rs:1025
                return RefreshTokenRequestOutcome::FailedTransiently;
codex-rs/app-server/src/request_processors/account_processor.rs:1027
            return RefreshTokenRequestOutcome::FailedPermanently;
codex-rs/app-server/src/request_processors/account_processor.rs:1112
        self.refresh_token_if_requested(do_refresh).await;
```

이 관측은 scratch 격리·별도 publish와 refresh 결과 재검증의 계획 근거를 확인한 것이다. 설치본 0.154.0이 이 코드와 같다는 주장이나 실제 refresh 실패 관측이 아니다. `login/src/auth/manager.rs`의 `refresh_lock: Semaphore`와 생성자도 직접 확인했다.

현재 WT `internal/lockfile/lockfile_windows.go`도 직접 읽었다. 관측 주석은 `Multi-process locking is not supported; see package comment above.`다. AUTH plan은 이 구현을 cross-process 보장의 근거로 쓰지 않으므로 알려진 범위와 충돌하지 않는다.

## Baseline-attribution

`git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified rev-parse --short HEAD` → `81c1d58f9`, exit 0. 대상 AUTH 0.1.0의 네 문서와 broker 보고서, 현재 코어 인터페이스의 내용은 E2 해시로 고정했다. 진행 중인 다른 worker의 코드는 수정하거나 감사 범위에 끌어들이지 않았다.

`git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified check-ignore .moai/reports/SPEC-MOAI-GATEWAY-001/auth-plan-audit-iter1.md` → stdout 빈 출력, exit 1. 추적 가능한 보고서 경로다. 이번 작업에서 작성한 것은 이 보고서뿐이다.

## Gaps

- 인증 네트워크·새 로그인·broker RPC·실제 refresh·remote revoke·일반 또는 인증 테스트를 실행하지 않았다.
- 공개 source `5a9eb14`와 설치 Codex 0.154.0의 빌드 대응은 검증하지 않았다.
- first-write callback이 실제 헤더 쓰기 완료/실패를 의미하는지, 실패 후 transport가 확실히 종료되는지는 실행 증거가 없다.
- 정확한 deadline·프로세스 간 lease 취소·Unix/Windows 잠금·Windows ACL·durable rename 실행은 아직 남아 있다.
- 구독 endpoint의 필수 헤더·지원 필드·실계정 네 GPT 접근과 코어/PICKER 통합은 아직 미측정이다. M0 INCONCLUSIVE는 유지된다.

## Residual-risk

wrapper mock이 통과해도 공식 broker protocol과 token 파일 형식이 다르면 실제 인증은 실패할 수 있다. Generation/Apply 호출만 붙이고 SendAuthorized를 연결하지 않으면 로그아웃 뒤 stale send가 남을 수 있다는 것이 계획 자체의 경고다. 이 경계의 실행 증거 없이 전체 AUTH 완료를 보고해서는 안 된다. 로컬 폐기를 원격 revoke 완료로 설명해서도 안 된다.

## Recommendation

로컬 mock broker wrapper·저장 트랜잭션·별도 프로세스 잠금·transport barrier 시험을 준비할 수 있다. 공식 broker live preflight와 새 인증은 승인된 2026-09-11 19:00 Asia/Seoul 이후에 실시하며, 그 전에 로그인·refresh·제품 인증 경로 성공을 주장하지 않는다. 이 계획 PASS는 실제 AUTH 완료 판정과 별개다.
