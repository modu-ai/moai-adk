#!/bin/sh
# who-sits-where — "지금 누가 어느 트리에 앉아 있는가"에만 답하는 탐침.
#
# card: t1038
#
# 상태 파일을 한 줄도 읽지 않는다. 네 층이 전부 라이브다:
#   1) /tmp/cc-socks/<pid>.sock  — 도달 가능한 세션 집합 (파일명이 곧 pid)
#   2) kill -0 <pid>             — 그 pid 가 살아 있는가 (소켓은 죽어도 남는다)
#   3) ps -o command=            — 세션 이름 (--name 또는 단축 -n)
#   4) lsof -a -d cwd            — 그 프로세스의 현재 cwd (이동을 즉시 따라온다)
#   5) git rev-parse --show-toplevel — cwd 를 트리로 정규화
#
# 5)가 없으면 비교가 성립하지 않는다: 세션은 워크트리 안의 하위 디렉터리에
# 앉아 있을 수 있고, 그 철자는 트리가 아니다.
#
# 출력: <pid> <name> <git-toplevel>
# 같은 toplevel 이 두 줄에 나오면 그 트리에 쓰기 주체가 둘이라는 뜻이다.

for sock in /tmp/cc-socks/*.sock; do
  pid=$(basename "$sock" .sock)
  kill -0 "$pid" 2>/dev/null || continue
  cmd=$(ps -o command= -p "$pid" 2>/dev/null)
  case "$cmd" in *claude*) ;; *) continue ;; esac
  name=$(printf '%s' "$cmd" | sed -nE 's/.*(--name|-n) ([A-Za-z0-9_-]+).*/\2/p')
  [ -z "$name" ] && name="<unnamed>"
  cwd=$(lsof -a -d cwd -p "$pid" -Fn 2>/dev/null | sed -n 's/^n//p')
  if [ -n "$cwd" ]; then
    tree=$(git -C "$cwd" rev-parse --show-toplevel 2>/dev/null)
  fi
  printf '%-7s %-12s %s\n' "$pid" "$name" "${tree:-<not-a-git-tree>}"
  tree=
done
