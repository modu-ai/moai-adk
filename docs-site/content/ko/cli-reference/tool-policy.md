---
title: moai tool-policy 도구 정책
weight: 88
draft: false
---

`moai tool-policy` 는 도구/권한 정책 SSOT를 관리합니다. `.moai/config/sections/tool-policy.yaml` 이 단일 진실 공급원이며, 여기서 `settings.json` 의 permissions 블록을 생성(codegen)하고 정책 항목을 조회합니다.

## 하위 명령어

| 명령어 | 설명 |
|--------|------|
| `moai tool-policy build` | tool-policy.yaml에서 settings.json permissions 블록 재생성 |
| `moai tool-policy list` | tool-policy 항목 나열 (thin query) |

## moai tool-policy build

```bash
moai tool-policy build
moai tool-policy build --local-only
```

로컬 `.claude/settings.json` 과 템플릿 `settings.json.tmpl` 의 permissions 블록을 재생성합니다.

| 플래그 | 설명 |
|--------|------|
| `--repo-root <path>` | 저장소 루트 (기본: cwd) |
| `--policy <path>` | tool-policy.yaml 경로 (기본: `<repo-root>/.moai/config/sections/tool-policy.yaml`) |
| `--local-only` | 로컬 `.claude/settings.json` 만 재생성 (템플릿 .tmpl 생략) |
| `--template-only` | 템플릿 settings.json.tmpl만 재생성 (로컬 생략) |
| `--default-mode <mode>` | `permissions.defaultMode` 재정의 (기본: 기존 값 보존) |
| `--json` | 결과를 JSON으로 출력 |

## moai tool-policy list

```bash
moai tool-policy list
moai tool-policy list --risk-tier irreversible --decision deny
```

| 플래그 | 설명 |
|--------|------|
| `--risk-tier <read\|write\|irreversible>` | 위험 티어로 필터 |
| `--decision <allow\|deny\|ask>` | 결정으로 필터 |
| `--tool <name>` | 도구 이름으로 필터 (정확 일치) |
| `--format <text\|json>` | 출력 형식 |
| `--repo-root <path>` | 저장소 루트 (기본: cwd) |
| `--policy <path>` | tool-policy.yaml 경로 |

## 관련 문서

- [settings.json 가이드](/ko/advanced/settings-json) — permissions 블록 상세
- [config 섹션 레퍼런스](/ko/advanced/config-sections)
- [CLI 개요](/ko/getting-started/cli)
