# CG 문서·runtime 연결 읽기 전용 인계

## Claim

현재 docs-site는 Hugo Geekdoc이다. README 네 언어와 CG 페이지·template의 아래 항목은 실제 문맥상 현재 실행 안내다. CHANGELOG의 버전 항목과 기존 감사 보고는 역사 자료다. 사이트·README·template 내용은 수정하지 않았으며 이 인계 파일만 작성했다.

## Evidence — 빌드·경로·검증 도구

| 파일/행 | 실제 읽은 설정/코드 | 다음 작업에 주는 조건 |
|---|---|---|
| docs-site/vercel.json:4-11 | framework=hugo, buildCommand=`hugo --minify --gc`, outputDirectory=public, devCommand=`hugo server`, HUGO_VERSION=0.160.1 | Nextra/npm 빌드가 아니다. 읽기 전용 조사에서는 gc/install/deploy를 실행하지 않았다. |
| docs-site/hugo.toml:1-9 | baseURL=https://adk.mo.ai.kr/, 기본 ko, defaultContentLanguageInSubdir=true, theme=hugo-geekdoc | 모든 언어가 URL prefix를 쓴다. |
| docs-site/hugo.toml:92-124 | languages ko/en/ja/zh, contentDir=content/<locale> | 네 CG 파일의 경로는 `/{ko,en,ja,zh}/multi-llm/cg-mode/`다. 이는 설정·파일 대응이며 실제 HTTP 응답 검증은 아니다. |
| docs-site/hugo.toml:67-70 | weight 기반 자동 sidebar | CG 파일을 남기고 title/description을 폐기·이전 안내로 바꾸면 같은 URL을 사용할 수 있다. 메뉴 수동 복제만 바꾸면 충분하다고 가정하지 않는다. |
| scripts/docs-i18n-check.sh:1-15,26-29,44 | canonical ko; 파일 수/경로, title, H1, glossary 검사; 기본 strict | `DOCS_I18N_STRICT=1 bash scripts/docs-i18n-check.sh`. HTML 링크/anchor/렌더 검사는 아니다. |
| .github/workflows/docs-i18n-check.yml:71-99 | PR/push는 strict=false, 수동 strict 가능 | CI 초록만으로 parity 무오류를 주장하지 않는다. |
| scripts/docs-version-snapshot/main.go:1-15 | release 시 content/<locale>/v<previous> 복사 | 역사 버전 snapshot 생성 도구이며 CG 문서 수정 전 해시 snapshot 도구가 아니다. 이번에는 실행하지 않았다. |

실제 실행한 도구 확인:

```text
$ command -v hugo
/opt/homebrew/bin/hugo
$ hugo version
hugo v0.160.1+extended+withdeploy darwin/arm64 BuildDate=2026-04-08T14:02:42Z VendorInfo=Homebrew
```

부모의 로컬 렌더 검증 후보 명령: `hugo --source <WT>/docs-site --destination /tmp/<고유검증경로> --minify`. 배포 설정의 `--gc`는 cache 정리를 하므로 읽기 전용 조사의 명령으로 사용하지 않았다. 빌드 후 생성된 `<destination>/<locale>/multi-llm/cg-mode/index.html`과 inbound 링크/anchor를 실제 확인해야 한다. 이 조사는 빌드나 HTTP/브라우저 검사를 실행하지 않았다.

## 문맥별 구분과 수정 전제

| 위치 | 실제 문맥 | 분류 / 후속 조건 |
|---|---|---|
| docs-site/content/en/multi-llm/cg-mode.md:115-121 | Step 3 launch CG, `moai cg` 실행과 자동 Claude 시작 설명 | 현재 실행 안내. 같은 페이지를 retirement/migration 안내로 전환하는 것이 URL 유지에 가장 작은 변경이다. 새 migration 문서 URL이 이미 있다는 근거는 발견하지 못했다. |
| ko/multi-llm/cg-mode.md:8-24,109-115; ja:8-15; zh:8-15,79-85 | 현재형 역할/비용절감 설명과 실행 절차 | 네 언어 모두 역사 기록으로 분리되지 않은 현재 제품 설명. 비용절감 수치를 새 gateway의 실증 결과로 재사용하면 안 된다. |
| en/multi-llm/cg-mode.md:171-182 | cc/glm/cg 비교표 및 CG 권장 설명 | 현재 선택 안내. CG를 GPT로 단순 치환하면 기존 혼합 역할이 보존되는 것처럼 읽히므로 금지. |
| en/multi-llm/cg-mode.md:184-200 | `--team` flag는 v3.0에 retired라는 역사 설명 + 현재 teammateMode 표·CG 설정 동작 | 한 절 안에 역사와 현재 설명이 섞인다. 과거 flag 설명을 사실 없이 삭제할 필요는 없지만, 지금 가능한 teammate 역할은 실제 capability 게이트와 분리해야 한다. |
| en/cli-reference/launchers.md:89-103,117 | 모든 세 launcher profile/worktree 지원, cg -w 예시, CG 링크 | 현재 실행 안내. 현재 지원 명령과 unverified gateway 활성화 상태를 구분하며 링크 라벨은 migration 안내로 바꿀 수 있다. |
| en/multi-llm/_index.md:103; en/multi-llm/model-policy.md:413 | CG 아키텍처/비용절감 페이지 링크 | 기존 URL inbound. ko 대응은 _index.md:178,208 및 model-policy.md:382. 네 언어 다른 inbound도 빌드 후 확인 필요. |
| README.md/README.ko.md/README.ja.md/README.zh.md:571 및 685 부근 | 실행 예시·현재 모드 비교표 | 현재 안내. :80의 factory 진입 이력은 과거 flag 유래와 현행 cg 거절을 함께 서술하므로 문단 전체 제거가 아닌 해당 의미 수정이 필요하다. |
| internal/template/templates/.claude/rules/moai/core/glm-web-tooling.md:8,25,44,54-75 | 운영 SSOT, GLM backend 표, CG leader 예외, active hybrid 메커니즘 | 사용자 프로젝트에 전달되는 현재 규칙. 'retired static Agent Teams prose'라는 문장이 있어도 CG 자체가 과거라는 뜻은 아니다. 실제 GLM 도구 경로/일반 GLM 지원은 유지해야 한다. |
| internal/template/templates/AGENTS.md:284-292 | 현재 명령 표에 cc/glm/cg, generated help 참고 | 현재 안내. 표에서 CG 실행을 제거하고 실제 지원·gate를 설명해야 한다. |
| internal/template/templates/CLAUDE.md:127 인근 | 현재 web policy와 CG pane 구분 설명 | 현재 운영 규칙. 긴 한 줄에 타 정책이 섞여 있으므로 전체 줄 삭제 금지. |
| CHANGELOG.md:1043, heading `[3.1.0] - 2026-08-15` | SPEC-INFINITE-GOAL-001 당시 변경·검증·커밋 기록 안 CG 참조 | 역사 보존. 현재 CLI 안내와 동일 검색어라도 재작성 대상 아님. |
| docs-site/content/<locale>/changelog/_index.md:1-14 | GitHub Releases/CHANGELOG로 이동 안내 | 역사 문서 연결. CG 전환 때문에 파일 변경할 이유를 관측하지 않았다. |
| .moai/reports/SPEC-MOAI-GATEWAY-001/plan-audit-iter5.md:1-8 | 당시 iter5 FAIL0.83, 당시 HEAD/바이트 기록 | 감사 증거 원문 보존. 현재 코드와 다르다는 이유로 인용/판정을 고치지 않는다. |

CG 이전 안내가 설명해야 하는 현재 구현 계약: preview 기본, claude-only 적용에 `--apply --accept-role-change` 필요, GLM 팀원 자동배정 제거를 명시, claude-glm 실제 TEAMMATE capability 없으면 적용/launch 모두 거부, 사용자 verified 필드 우회 불가. gateway 전체 실계정 PASS나 혼합 역할 동등성을 미리 주장하지 않는다.

## 실제 template 시험 경계

- `internal/template/renderer_test.go:12` `TestRendererRender`: fstest.MapFS 합성 템플릿의 strict missing key/치환 등을 판정한다. 실제 CG 문구 부재를 검사하지 않는다.
- `internal/template/embed_test.go:153` `TestEmbeddedTemplates_CLAUDEmd`: 실제 EmbeddedTemplates에서 CLAUDE.md를 읽지만 길이/필수 표제만 검사한다. CG retirement 의미를 통과시켜 주는 시험이 아니다.
- `internal/template/deployer_test.go:45` `TestDeployerDeploy`: `NewDeployer(testFS())` 합성 FS를 격리 프로젝트로 배포한다. 실제 수정된 전체 embedded 파일의 의미 검증을 대신하지 않는다.
- `internal/template/rule_template_mirror_test.go:137` `TestRuleTemplateMirrorDrift`: 명시 allowlist의 source↔template byte parity. :42-65에서 spec-workflow, session-handoff-examples 등이 포함됨을 읽었다. 전체 template 자동 비교가 아니다. 부모가 수정하는 source가 allowlist 밖이어도 실제 mirror 정책을 별도 확인해야 한다.

검증 후보: `go test ./internal/template -run 'TestRendererRender|TestEmbeddedTemplates_CLAUDEmd|TestRuleTemplateMirrorDrift|TestDeployerDeploy'`. 이 조사에서는 실행하지 않았으며, 이것만으로 CR006 PASS는 아니다. 실제 수정한 embedded 파일 추출/렌더와 현재 CG 실행 권고 부재, 역사 hash 보존을 추가 확인해야 한다.

## 역사 파일 hash snapshot

별도 사용자 파일을 생성하지 않고, 이 보고서의 다음 블록을 **현재 읽은 역사 보존 manifest 위치**로 둔다. Python hashlib.sha256로 실제 bytes를 읽어 측정했다. 전체 역사 파일 목록의 완전성은 주장하지 않는다.

```text
0da2986a18540093d014cb6f09b354dc31e865612667f24dd7e64ec62614c767  CHANGELOG.md
04137784fe991d7e77a485b0ffd808ba06c52c4da4e2ee5479f3ad96fcd32395  docs-site/content/ko/changelog/_index.md
e80c44498fb0e751314d8622e54ddb3e6ab599535344ae99032ea072cb3a6806  docs-site/content/en/changelog/_index.md
f46622c6817924a7fa59a3e2e7147569e79ecaaea19dd4f46f59ec47f2339df0  docs-site/content/ja/changelog/_index.md
4517e52e625e1187a311e1c0bebc4f044e1ee42d994619736d0856b85bc51bcb  docs-site/content/zh/changelog/_index.md
e095e800f0d55d5a2d4f331d22dcbd69928e4a7eb4b47efa4fb5148f2e53b40b  .moai/reports/SPEC-MOAI-GATEWAY-001/plan-audit-iter5.md
```

## Baseline-attribution / Gaps / Residual-risk

2026-09-11 같은 WT `moai-proxy-unified`, 기준 HEAD81c1d58f9. 읽기 도구는 rg/sed/Python bytes hashing 및 hugo version이다. docs writer가 병행 착수했으므로 후속 수정 전 각 라인을 다시 읽어야 한다. root README/template/docs body/사용자 설정/git 상태를 수정하지 않았다. 빌드·렌더·링크·parity·template 시험은 이 읽기 전용 조사에서 실행하지 않았다. 위 구분은 직접 읽은 문맥에 한하며 남은 모든 CG 검색 결과의 삭제 안전성이나 문서 전체 완료를 주장하지 않는다.
