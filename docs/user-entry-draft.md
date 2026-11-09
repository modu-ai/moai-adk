# 사용자 첫 진입 개편안 (08-D #2015 초안 — 게시 전)

목표: "한 번 추가하고, 기존 앱에서 쓴다"(설계 00-page §22). 자연어 진입이 기본이고 수동 명령은 명시적 fallback이다. 3.x Go 세대 명령을 차세대 필수 단계로 소개하지 않는다(REQ-08D-06). docs-site 배포는 범위 밖 — 이 파일은 README·문서 사이트가 채택할 초안이다.

## 초안 본문

### MoAI를 프로젝트에 한 번 추가하기

설치 후 프로젝트 루트에서 호스트 세션을 열고 이렇게 말합니다:

- Claude Code: `/moai 이 프로젝트 현재 상태를 읽고 설정해 줘`
- Codex: `$moai 이 프로젝트 현재 상태를 읽고 설정해 줘`

SETUP이 끝나면 같은 진입으로 READ → PLAN → RUN → VERIFY → AUDIT → SYNC가 이어집니다. 매번 절차를 외울 필요가 없습니다 — 요청하면 워크플로가 필요한 만큼만 로딩됩니다.

### 설치

```sh
# RC1 게시 후 — GitHub Releases의 `moai-cli-v<version>` prerelease 자산
# (installer가 ~/.moai-cli/bin/moai-cli 설치와 PATH 안내를 담당합니다)
```

- 기존 Go 세대(`moai` 3.x) 사용자: `~/.moai/` 작업 데이터는 그대로 보존되며 두 세대는 서로의 업데이트 대상이 아닙니다. 전환 절차는 별도 안내(04-D)가 제공될 때까지 기존 대로 `moai`를 쓰면 됩니다.
- 서명 검증: release artifact는 GitHub artifact attestation으로 검증합니다(`gh attestation verify <asset> -R modu-ai/moai-adk`). checksum은 무결성 재확인용이며 서명을 대체하지 않습니다.

### 요청 예시

| 하고 싶은 것 | 말하면 |
| --- | --- |
| 프로젝트 등록·설정 | `/moai 이 프로젝트 설정해 줘` |
| 현황 읽기 | `/moai 지금 저장소 상태 요약해 줘` |
| 요구 정리 | `/moai 이 기능 요구사항 정리해 줘` |
| 구현·검증 | `/moai 이 SPEC 구현하고 검증까지 해 줘` |

### 호스트 세션과 역할

호스트 세션은 역할 없이 엽니다 — `moai-cli cc|glm|codex`는 평범한 host 세션이고, 세션 안에서 `/moai --mode leader|agent`로 MoAI 역할을 정합니다. 리더·에이전트 론처 플래그 `-l|-a`(`moai-cli -l|-a`, `moai-cli cc|glm|codex -l|-a`)는 deprecated입니다: 사용하면 경고가 뜨고 기존처럼 동작하지만, 새 세션은 역할 없이 열고 `/moai --mode`를 쓰는 것을 권합니다.

### 수동 명령 (명시적 fallback)

자연어 진입이 불가능한 환경에서만 내부 실행 도구를 직접 씁니다: `moai-cli generation-plan`, `moai-cli update-plan` 등(각 명령은 plan/apply 2단계이며 plan digest 없이는 아무것도 쓰지 않습니다). 일상 사용 경로가 아닙니다.

### 3.x 명령에 대해 (fallback)

기존 `moai init`·`moai update`(Go 3.x)는 구 세대의 명령입니다. RC1을 이미 설치했다면 3.x 명령을 새 절차로 쓰지 마세요 — 두 세대의 업데이트 통로는 분리되어 있습니다(지원표와 릴리스 체크리스트 참조). 마이그레이션 도구는 별도 제공 예정입니다.

## 채택 시 확인 항목

- [x] 3.x 명령을 필수 단계로 안내하는 문장 0건(`grep -nE 'moai (init|update)\b' docs/*.md README.md | grep -v 'fallback\|3\.x'` = 0)
- [x] `/moai`·`$moai` 예시가 진입의 기본
- [ ] README 본문 반영 — README는 기존 사용자 영역 유지(이 초안 채택 시 별도 편집)
