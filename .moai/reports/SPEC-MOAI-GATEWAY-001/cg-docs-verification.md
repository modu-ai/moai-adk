# CG 폐기 문서 동기화 검증

## Claim

README 네 언어, docs-site의 현재 안내와 메뉴를 CG 폐기·명시적 이전 계약에 맞추어 수정했다. 소유 범위의 변경 파일은 156개이며 정확한 SHA-256은 `cg-docs-source-manifest.json`에 있다. 기존 `/{ko,en,ja,zh}/multi-llm/cg-mode/` URL은 유지하고 설정 이전 안내로 다시 작성했다. 명시 앵커 `tmux-env-security`도 보존했다.

이전 명령은 기본 preview인 `moai migrate cg`, 역할 변화 수락을 포함한 `moai migrate cg --target claude-only --apply --accept-role-change`이다. `claude-only`가 GLM 팀원 자동 배정을 제거함을 명시했다. `claude-glm`은 미리보기만 가능하고 실제 TEAMMATE 게이트 전에는 적용·실행할 수 없다고 적었다. CG를 cc/GPT의 자동 별칭으로 설명하지 않았고, GPT gateway를 사용 가능하다고 새로 안내하지 않았다.

현재 비용·역할 분담 보장, CG 실행 예시, 프로필/작업트리의 세 런처 설명, 폐기된 backend 거절 사유를 정리했다. 단일 GLM 세션의 웹 도구·추론·일반 tmux 자격증명 설명에서는 CG 대안만 제거했다. 정적 Agent Teams 폐기의 과거 설명과 현재 CG 폐기를 구분했다. 역사 release/audit 파일과 이미지 원본은 편집하지 않았다.

## Evidence

모든 명령은 아래 Baseline-attribution의 worktree에서 실행했다.

### 최종 Hugo 빌드

```bash
hugo --source /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified/docs-site --destination /tmp/cg-docs-render-20260911 --minify --cacheDir /tmp/cg-docs-hugo-cache-20260911 > /tmp/cg-docs-build-20260911.log 2>&1
```

exit 0. 최종 로그 끝부분 원문:

```text
Total in 2724 ms
```

버전 출력 원문:

```text
hugo v0.160.1+extended+withdeploy darwin/arm64 BuildDate=2026-04-08T14:02:42Z VendorInfo=Homebrew
```

전체 로그에는 Pages KO187/EN185/JA185/ZH185가 기록돼 있다. 초기 편집 중 callout 종료 태그 두 개가 빠져 빌드가 실패했고 원문과 다시 대조하여 복구했다. 위 결과는 그 수정과 마지막 문구 정리 이후의 최종 빌드다. `--gc`와 배포는 실행하지 않았다.

### 네 언어 정합성

```bash
DOCS_I18N_STRICT=1 bash scripts/docs-i18n-check.sh > /tmp/cg-docs-i18n-20260911.log 2>&1
```

exit 0. 최종 로그 끝부분 원문:

```text
=== Check 4: Glossary term preservation (canonical: ko) ===

=== Summary ===
Errors:   0
Warnings: 0
OK: all 4 locales pass parity, frontmatter, H1, and glossary checks.
```

각 언어 `.md` 153개, 파일 경로·title·H1·용어 정합성 검사다. 문장 의미의 완전성이나 브라우저 시각 검사를 뜻하지 않는다.

### 생성 HTML의 링크·실행 예시

`python3 /tmp/cg-docs-check.py`: exit 0. Python HTMLParser로 생성 파일을 읽어 CG 페이지로 향하는 site-local 링크의 경로와 fragment를 실제 HTML id에 대조했다. pre 블록에서 폐기된 `moai cg` 실행 줄을 검사했다. 최종 원문은 `/tmp/cg-docs-html-check-20260911.json`이다.

```json
{
  "html_pages": 657,
  "cg_inbound_links": 729,
  "cg_inbound_errors": [],
  "retired_launch_codeblocks": [],
  "migration_pages": {
    "ko": {"preview": true, "apply": true, "hybrid_preview": true, "security_anchor": true},
    "en": {"preview": true, "apply": true, "hybrid_preview": true, "security_anchor": true},
    "ja": {"preview": true, "apply": true, "hybrid_preview": true, "security_anchor": true},
    "zh": {"preview": true, "apply": true, "hybrid_preview": true, "security_anchor": true}
  }
}
```

JSON은 동일 값을 줄 접기만 바꾸어 실었다. 729개에는 공통 메뉴의 반복 링크가 포함된다. 고유 문서 729개를 뜻하지 않는다. Hugo Pages 집계와 실제 html 파일 수는 서로 다른 측정 대상이다.

`git diff --check -- README.md README.ko.md README.ja.md README.zh.md docs-site/content docs-site/data/menu/main.yaml`: exit 0, 출력 없음.

### 역사 보존 표본

`cg-docs-handoff.md`의 여섯 SHA-256과 현재 파일 bytes를 Python hashlib로 비교했다. 출력 원문:

```text
UNCHANGED 0da2986a18540093d014cb6f09b354dc31e865612667f24dd7e64ec62614c767 CHANGELOG.md
UNCHANGED 04137784fe991d7e77a485b0ffd808ba06c52c4da4e2ee5479f3ad96fcd32395 docs-site/content/ko/changelog/_index.md
UNCHANGED e80c44498fb0e751314d8622e54ddb3e6ab599535344ae99032ea072cb3a6806 docs-site/content/en/changelog/_index.md
UNCHANGED f46622c6817924a7fa59a3e2e7147569e79ecaaea19dd4f46f59ec47f2339df0 docs-site/content/ja/changelog/_index.md
UNCHANGED 4517e52e625e1187a311e1c0bebc4f044e1ee42d994619736d0856b85bc51bcb docs-site/content/zh/changelog/_index.md
UNCHANGED e095e800f0d55d5a2d4f331d22dcbd69928e4a7eb4b47efa4fb5148f2e53b40b .moai/reports/SPEC-MOAI-GATEWAY-001/plan-audit-iter5.md
```

## Baseline-attribution

- 2026-09-11, `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`.
- 직접 실행한 `git rev-parse --short HEAD`, `git branch --show-current` 출력: `81c1d58f9`, `WT-unified-gateway`.
- 입력: CG-RETIRE 0.1.0 여섯 문서, `cg-migration-verification.md`, `cg-retirement-runtime-verification.md`, `cg-docs-handoff.md`.
- runtime 사실은 위 구현 보고와 SPEC 계약을 입력으로 삼았다. 구현 보고의 Go 시험을 이 문서 작업에서 재실행했다고 주장하지 않는다.
- `moai-workflow-project` 스킬의 문서/언어/검토 범위를 적용했다. checkpoint는 `.moai/state/checkpoints/docs/cg-retirement-20260911.json`이다.
- 코드·template·SPEC 본문·상태·CHANGELOG를 편집하지 않았다. commit/push/PR/merge/publish는 하지 않았다.

## Gaps

- 브라우저 화면·접근성·스크린샷은 검사하지 않았다. Hugo 성공은 HTML 생성 근거이며 Mermaid의 브라우저 렌더 성공 증거가 아니다. 독립 Mermaid parser와 별도 Markdown linter는 실행하지 않았다.
- HTML 링크 검사는 CG 이전 페이지로 향하는 site-local 링크만 대상으로 했다. 외부 사이트나 모든 다른 문서 링크의 HTTP 상태를 검사하지 않았다. 현재 사이트에서 사용하지 않는 외부 과거 heading fragment까지 보존한다고 주장하지 않는다.
- template 배포·실제 CLI help·runtime·TEAMMATE 게이트는 부모/구현 담당의 별도 범위다. 이 보고만으로 AC-CR-006 전체 또는 CG-RETIRE SPEC 전체 PASS/완료를 표시하지 않는다.
- 여섯 역사 표본 외 모든 역사·사용자 파일의 전후 hash를 측정한 것은 아니다. 사용자 설정은 읽거나 수정하지 않았다.
- 독립 원어민 의미 감사는 수행하지 않았다. 이전 CG 비용 수치는 새 gateway 실계정 결과로 전용하지 않았다.

## Residual-risk

지원 런처의 실제 동작과 TEAMMATE 게이트가 달라지면 안내의 제한도 갱신해야 한다. 기존 페이지의 다른 기능은 CG 폐기 범위 밖이므로 재검증하지 않았다. 156개 파일의 넓은 문맥 변경은 독립 검토에서 추가 확인할 수 있도록 파일별 해시와 생성 HTML을 보존했다. 문서의 입력 계약과 실제 제품 활성화 조건을 혼동하여 전체 gateway 출시 완료로 읽어서는 안 된다.
