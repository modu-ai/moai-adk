# GPT 인증 broker 설계 검토안

## Claim

공식 설치 Codex app-server를 새 MoAI 소유 scratch CODEX_HOME의 인증 broker로 재사용한다. 이는 제안 architecture이며
구현·로그인·갱신 PASS가 아니다. 기존 사용자 로그인이나 등록 client ID를 복사하지 않고 새로 로그인한다.

## 경계

- broker: PKCE·callback·로그인/취소/갱신/logout. gateway 추론은 별도 Responses HTTP.
- MoAI: canonical 원자 저장·세대·tombstone·cross-process operation/state lock·송신 first-write와 logout 직렬화.
- broker 파일: scratch 전용. broker 종료 확인 후 검증하여 시작 세대가 변하지 않았을 때만 canonical publish.
- logout: tombstone 먼저, in-flight 취소 시도, remote revoke 결과 별도. 실패한 revoke로 credential 복원 금지.

## 실제 관측과 기준선

공개 source checkout `/tmp/openai-codex-audit.zbvNXB`에서 `git rev-parse --short HEAD` 출력은 `5a9eb14`다.
`login/src/auth/storage.rs` 직접 열람 출력에 `options.truncate(true).write(true).create(true);`가 있다.
`account_processor.rs` 직접 열람 출력에 `self.refresh_token_if_requested(do_refresh).await;`가 반환값 사용 없이 있다.
이는 source 동작 판독이며 설치 Codex 0.154.0과 동일하다는 증명도 실제 refresh 실패 관측도 아니다.
공식 endpoint·file storage 문서를 열어 확인했으며 AUTH plan D에 링크와 한계를 기록했다.

## 검토 항목

state lock은 response stream 전체를 잡지 않는다. 최초 HTTP header write 완료/실패까지 직렬화하며 bounded write deadline을
적용한다. 그 직전 세대 검사 없이 enqueue만 하는 방식은 logout 이후 stale send를 막지 못한다. 설치 transport가 이
callback을 제공하는지와 프로세스 간 crash·취소 정리는 run에서 검증한다. 해당 경계 없는 CredentialRef 연결만으로는 미완료다.

## Gaps 및 잔여 위험

설치본 RPC protocol, file store 적용과 관리 정책, 새 로그인·실제 refresh, 구독 endpoint 필수 header/field·네 GPT 접근은
아직 미실행이다. remote logout의 서버 revoke 의미도 확인 전 주장하지 않는다. mock broker 시험은 MoAI wrapper만
검증하며 Codex PKCE unit coverage가 아니다. 실제 auth 시험은 2026-09-11 19:00 Asia/Seoul 이후다.


## 구조 검사와 재사용 근거

지정 WT에서 `/tmp/moai-gateway-81c1d58f9 spec lint SPEC-MOAI-GPT-AUTH-001`을 실행했다.

```text
✓ No findings — all SPEC documents are valid
```

Exit code 0. 기존 바이너리를 사용했으며 이번 작성자는 재빌드하지 않았다. AUTH는 REQ10·AC10, version 0.1.0 draft다.
atomicfile.Replace 및 lockfile의 Windows in-process 범위 주석, homestate의 LockFileEx 호출을 직접 읽었다.
AUTH 전용 최소 lock과 기존 atomic replace를 재사용하는 계획이며 해당 플랫폼 실행은 아직 하지 않았다.
