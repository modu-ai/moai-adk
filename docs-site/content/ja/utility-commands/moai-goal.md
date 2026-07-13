---
title: /moai goal
weight: 25
draft: false
---

完了条件を宣言すると、セッションがその条件を満たすまで自律的に動き続ける **条件宣言型自律ループ** コマンドです。`/moai goal "<条件>"` で完了条件を arm し、各ターン終了時に `stop-goal` Stop フックが条件充足を評価し、満たされるまで次のターンを自動開始します。`/moai goal status [--all]` で進捗を確認し、`/moai goal clear` でループを解除し、`/moai goal resume` で中断地点から再開します。専用スラッシュコマンドファイルはなく、`moai` スキルルーティングと `moai goal` CLI から入るプログラマティック命令面です。
