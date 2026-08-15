---
title: moai web 웹 콘솔
weight: 50
draft: false
description: "로컬 운영 콘솔을 띄우는 moai web 커맨드 — 플래그, 라우트, 포트 회수 동작."
---

`moai web` 은 로컬 운영 화면인 **MoAI Web Console** 을 띄웁니다. 프로젝트의 SPEC 카탈로그와 칸반 체인, 세션 · 목표 · 검증 상태를 브라우저에서 보고, 같은 화면에서 프로필 선호도와 프로젝트 설정을 고칠 수 있습니다.

화면 구성과 각 영역이 무엇을 읽는지는 [MoAI Web Console](/ko/advanced/moai-web-console/) 에서 다룹니다. 이 페이지는 커맨드 자체 — 플래그, 라우트, 포트 처리 — 를 정리합니다.

## 개요

```bash
moai web [OPTIONS]
```

콘솔은 **루프백(127.0.0.1)에만 바인딩** 됩니다. 외부 데이터베이스, 인증, 네트워크 노출이 전혀 없습니다.

## 플래그

| 플래그 | 기본값 | 설명 |
|--------|--------|------|
| `--port <N>` | `3041` | 127.0.0.1 에 바인딩할 TCP 포트 |
| `--no-open` | `false` | 브라우저를 자동으로 열지 않음 |
| `--no-reuse` | `false` | 오래된 moai 인스턴스로부터 포트를 회수하지 않고, 포트 충돌 시 실패 |

## 예시

```bash
moai web                 # 127.0.0.1:3041 에 바인딩하고 브라우저를 엶
moai web --port 9000     # 다른 포트에 바인딩
moai web --no-open       # 브라우저를 열지 않고 시작
moai web --no-reuse      # 포트가 사용 중이면 회수 대신 실패
```

## 포트 회수 동작

대상 포트를 이미 오래된 moai 인스턴스가 점유하고 있으면, 기본값에서는 그 인스턴스를 종료하고 다시 바인딩합니다. moai 가 아닌 외부 프로세스는 **절대 종료하지 않으며**, 이때는 오류를 내고 `--port` 로 다른 포트를 쓰도록 안내합니다. `--no-reuse` 를 붙이면 오래된 moai 인스턴스도 회수하지 않고 그대로 실패합니다.

브라우저를 여는 데 실패해도 서버는 계속 켜져 있습니다. 터미널에 찍힌 주소를 수동으로 열면 됩니다.

## 라우트

콘솔이 여는 경로는 다음과 같습니다. 화면 넷은 읽기 전용이라 GET 이외의 메서드를 405 로 거부합니다.

| 경로 | 메서드 | 하는 일 |
|------|--------|---------|
| `/` | GET | 개요 — 통계 타일, 칸반 체인, 진행 중 SPEC, 주의 목록, 세션 |
| `/kanban` | GET | 체인 세션 보드 + SPEC 파이프라인 |
| `/specs` | GET | SPEC 카탈로그. `?q=` 검색, `?status=` 필터, `?id=` 상세 |
| `/monitor` | GET | 세션 · 목표 · 검증 · 에픽 |
| `/settings` | GET | 설정 9개 탭. `?tab=` 으로 탭, `?profile=` 로 편집 대상 프로필 지정 |
| `/events` | GET | SSE 스트림 — 갱신 신호만 흘려보냄 |
| `/save` | POST | 설정 저장 |
| `/profile/create` · `/profile/rename` · `/profile/delete` | POST | 프로필 생애주기 |
| `/glm-key/reveal` | POST | 저장된 GLM API 키 표시 |
| `/__shutdown__` | POST | 서버 종료 |

`/events` 는 연결을 계속 열어 두는 스트림입니다. `curl` 로 직접 열면 응답이 끝나지 않으므로 타임아웃으로 끊기는 것이 정상 동작입니다.

## 종료

터미널에서 `Ctrl+C` 를 누르거나, 레일 아래의 종료 버튼을 누릅니다. 어느 쪽이든 진행 중인 요청을 마무리한 뒤 안전하게 끝납니다.

## 프로필 기록의 범위

콘솔에서 프로필을 전환하면 그 선택이 `~/.moai/claude-profiles/launch.yaml` 에 현재 프로젝트의 기록으로 남습니다. 같은 프로젝트에서 `-p` 없이 `moai cc` 를 실행할 때 이 값이 쓰입니다.

콘솔이 읽는 값과 쓰는 값은 모두 현재 프로젝트를 기준으로 하므로, 화면에 보이는 프로필과 실제로 기록되는 프로필은 언제나 같습니다. 다만 `moai cc -p X` 로 시작한 세션 안에서 콘솔을 열면 `CLAUDE_CONFIG_DIR` 이 이미 정해져 있어, 기록과 무관하게 `X` 를 그대로 표시합니다.

선택 순서와 제약은 [프로필 관리](/ko/cli-reference/profile/) 에서 자세히 다룹니다.

---

관련: [MoAI Web Console](/ko/advanced/moai-web-console/) · [프로필 관리](/ko/cli-reference/profile/) · [CLI 개요](/ko/getting-started/cli/)
