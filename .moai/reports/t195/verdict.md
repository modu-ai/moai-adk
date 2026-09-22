# t195 — Claude Cloud 환경 moai 바이너리 시딩

- 브랜치: `WT-cloud-seed` (worktree `.claude/worktrees/t195`), base `origin/main` (`76b2c4ece`)
- 결론: **no-code 채택** — 문서(가이드 4로케일)만. 템플릿 setup 스크립트 불필요(§3).
- 카드 지시 레시피는 **실측으로 반증됨** — `go install …@latest`는 동작하지 않는다(§2).

## 1. 근거 (공식 문서 실측, code.claude.com/docs/en/cloud-environments)

| 사실 | 인용 절 |
|---|---|
| 세션은 저장소를 새로 클론한 Ubuntu 24.04 x86_64 VM에서 시작, 커밋된 `.claude/**`·`.mcp.json`·`CLAUDE.md`는 전부 실림 | What carries over from your setup |
| setup script는 bash·root·Claude Code 기동 **전** 1회, exit 0 필수, 약 5분 내 완료, 설치엔 네트워크 필요 | Setup scripts / Script requirements |
| 완료 후 파일시스템 스냅샷 → 이후 세션은 setup script 생략. 스크립트/허용호스트 변경 또는 약 7일 만료 시 재실행 | Environment caching |
| 기본 **Trusted** 허용 도메인에 `github.com`·`raw.githubusercontent.com`·`objects.githubusercontent.com`·`release-assets.githubusercontent.com`·`proxy.golang.org`·`sum.golang.org` 포함 | Default allowed domains |
| **`adk.mo.ai.kr`은 허용 목록에 없음** | 같은 절 (부재로 확인) |
| Go는 사전 설치되어 있으나 버전은 표에 없음("Go with module support") | Installed tools |

## 2. `go install` 실측 — 카드 지시 레시피는 실패한다

```
$ go install github.com/modu-ai/moai-adk/cmd/moai@latest
go: github.com/modu-ai/moai-adk/cmd/moai@latest: module github.com/modu-ai/moai-adk@latest
    found (v1.14.5), but does not contain package github.com/modu-ai/moai-adk/cmd/moai

$ go install github.com/modu-ai/moai-adk/cmd/moai@v3.1.2
go: invalid version: module contains a go.mod file, so module path must match
    major version ("github.com/modu-ai/moai-adk/v3")
```

원인: `go.mod`의 모듈 경로가 `github.com/modu-ai/moai-adk`로 `/v3` 접미사가 없다. Go의 시맨틱
임포트 버저닝이 메이저 2 이상에 접미사를 요구하므로 `@latest`는 접미사 없는 경로가 가질 수 있는
최신 태그(`v1.14.5`)로 해소되고, v3 태그는 지정해도 거부된다.

`@main`은 동작한다(측정):

| 경로 | 소요 | 산출물 | 판본 표시 |
|---|---|---|---|
| `go install …@main` | **1m42s** | 76,142,962 B | `moai-adk v3.1.2` (컴파일 기본값 — ldflags 없음, commit/date 없음) |
| `install.sh --install-dir <dir>` | **2.4s** | 35,970,082 B | `3.1.2  4b2f203fe  built 2026-08-21T06:47:28Z` |

`@main`을 권장하지 않는 이유 3: (a) 트리 전체 컴파일 (b) VM의 Go가 이 모듈 `go` 지시자
(현재 `go 1.26.4`)를 만족해야 함 — VM Go 버전 미공개 (c) 판본 스탬프 부재로 사후 추적 불가.

## 3. 채택 레시피와 그 근거

```bash
#!/bin/bash
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh \
  | bash -s -- --install-dir /usr/local/bin || true
moai --version || true
```

| 선택 | 근거 (실측) |
|---|---|
| `raw.githubusercontent.com` 경유 | `adk.mo.ai.kr`이 Trusted 목록에 없음. raw URL 응답 http=200 size=12636이고 저장소 루트 `install.sh`와 **바이트 동일**(`diff -q` 일치) |
| `--install-dir` 플래그 (환경변수 아님) | `install.sh` main()이 `INSTALL_DIR=""`로 초기화(코드 확인) → env 값 무시. 기본값 경로는 `$GOPATH/bin` 또는 `~/.local/bin`이라 PATH 보장 없음 |
| `\|\| true` | exit≠0이면 세션 시작 자체가 실패(Script requirements) |
| 최신 릴리즈(비고정) | 캐시가 7일/스크립트 변경 시 재빌드되므로 재실행마다 현재 판본을 집음. 고정 필요 시 `--version 3.1.2` |

**템플릿 setup 스크립트를 만들지 않은 이유**: setup script는 저장소 파일이 아니라
claude.ai/code의 환경 설정 대화상자 필드에 입력한다. 저장소에 스크립트를 두어도 클라우드가
자동으로 읽지 않으므로, 배포물이 아니라 붙여넣기 레시피가 맞는 형태다. (SessionStart 훅으로
설치하는 대안은 공식 문서가 "프로젝트 의존성"용으로 구분하고 VM 프로비저닝은 setup script
소관이라 명시 — Setup scripts vs. SessionStart hooks.)

## 4. 산출물

- `docs-site/content/{en,ko,ja,zh}/guides/claude-cloud.md` (신규 4파일)
- `docs-site/content/{en,ko,ja,zh}/guides/_meta.yaml` — 네비 항목 추가
- `docs-site/content/{en,ko,ja,zh}/guides/_index.md` — 가이드 목록 줄 추가

## 5. 검증 (관측된 출력)

| 명령 | 결과 |
|---|---|
| `bash scripts/docs-i18n-check.sh` | `Errors: 0 / Warnings: 0`, 4로케일 각 150 .md 파일 파리티 |
| `hugo` (docs-site) | rc 0, WARN/ERROR 0건, KO 184 / EN·JA·ZH 182 페이지 |
| 렌더 산출 | `public/{en,ko,ja,zh}/guides/claude-cloud/index.html` 4개 생성, mermaid 블록 렌더 확인 |
| Mermaid 방향 | `flowchart LR/RL/BT` grep 0건 (TD-only 준수) |
| URL | 사용 URL 3종 = `raw.githubusercontent.com/...`, `claude.ai/code`, `adk.mo.ai.kr/install.sh`(README 대비 설명용). 금지 URL(`docs.moai-ai.dev`·`adk.moai.com`·`adk.moai.kr`) 0건 |

## 6. 미검증 (Gaps)

- **실제 Claude Cloud VM에서 레시피를 실행해 보지 않았다.** 모든 근거는 (a) 공식 문서 인용과
  (b) 로컬 macOS/arm64 실측이다. VM은 Ubuntu 24.04 x86_64이며 `install.sh`의 플랫폼 감지가
  `linux_amd64`를 고르는 것은 코드로만 확인했다.
- **클론과 setup script의 선후 관계**를 문서가 명시하지 않는다. 레시피는 저장소 파일에
  의존하지 않으므로 어느 순서든 무해하지만, 저장소 스크립트를 호출하는 형태였다면 이 불확실성이
  결정적이었을 것이다.
- VM의 Go 버전 미공개 — `@main` 경로의 실패 가능성은 추정이지 실측이 아니다.
- 세션 사용자가 root인지 여부 미확인. `/usr/local/bin`을 고른 것은 이 불확실성에 대한 방어다.

## 7. 부수 관측 (카드 밖)

`install.sh` 실행 로그 끝의 판본 줄은 **이미 설치돼 있던 `~/go/bin/moai`**(`v3.1.3-rc.0`)를
PATH에서 찾아 출력한 것이었고, 새로 설치된 바이너리는 `3.1.2 4b2f203fe`로 정상이었다.
릴리즈 자산 결함이 아니다 — 초기 오독을 정정해 기록해 둔다.
