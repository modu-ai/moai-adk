---
title: moai web 웹 콘솔
weight: 50
draft: false
---

`moai web` 는 브라우저 기반 설정 편집기인 **MoAI Web Console** 를 실행합니다. 터미널 프로필 마법사(`moai profile`)와 동일한 검증·저장 로직을 재사용하며, 프로필 선호도와 프로젝트의 user / language / statusline 섹션을 웹 UI에서 편집합니다.

## 개요

```bash
moai web [OPTIONS]
```

콘솔은 **루프백(127.0.0.1)에만 바인딩** 됩니다. 외부 데이터베이스, 인증, 네트워크 노출이 전혀 없습니다. 기본적으로 대상 포트를 오래된 moai 인스턴스가 점유하고 있으면 해당 인스턴스를 종료하고 재바인딩합니다. moai 가 아닌 외부 프로세스는 절대 종료하지 않으며, 이 경우 오류를 보고하고 `--port` 사용을 제안합니다.

## 플래그

| 플래그 | 설명 |
|--------|------|
| `--port <N>` | 127.0.0.1 에 바인딩할 TCP 포트 (기본값: `3041`) |
| `--no-open` | 브라우저를 자동으로 열지 않음 |
| `--no-reuse` | 오래된 moai 인스턴스로부터 포트를 회수하지 않고, 포트 충돌 시 실패 |

## 예시

```bash
moai web                 # 127.0.0.1:3041 에 바인딩하고 브라우저를 엶
moai web --port 9000     # 다른 포트에 바인딩
moai web --no-open       # 브라우저를 열지 않고 시작
moai web --no-reuse      # 포트가 사용 중이면 회수 대신 실패
```

## 편집 대상

웹 콘솔은 다음을 편집합니다.

- **프로필 선호도** — 모델·언어·표시 설정 등 프로필별 설정
- **프로젝트 설정** — `.moai/config/sections/` 의 user / language / statusline 섹션

저장 시 터미널 마법사와 동일한 유효성 검사를 거치므로, 두 경로 중 어느 쪽을 사용해도 결과가 일관됩니다.

---

관련: [프로필 관리](/cli-reference/profile) · [CLI 개요](/getting-started/cli)
