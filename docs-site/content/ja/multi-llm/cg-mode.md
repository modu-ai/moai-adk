---
title: CG の廃止と設定の移行
weight: 20
draft: false
description: CG の廃止と設定の移行
---

`moai cg` は廃止されました。Claude や GLM を起動せず、移行案内を表示して終了します。`moai cc` の別名ではありません。`llm.team_mode: cg` が残るプロジェクトでは、セッションを起動する前に移行先を明示的に選ぶ必要があります。

## 変更前のプレビュー

プロジェクトのルートで選択肢を確認します。プレビューでは設定を変更せず、バックアップも作成しません。

```bash
moai migrate cg
moai migrate cg --target claude-only
```

## Claude のみの役割へ移行

GLM チームメイトの自動割り当てをなくす役割変更に同意する場合に限り、次のコマンドで適用します。

```bash
moai migrate cg --target claude-only --apply --accept-role-change
```

`llm.team_mode: claude`、`llm.gateway.teammate_mode: in-process`、`llm.gateway.teammate_provider: inherit` を保存します。従来の混合構成の役割分担を解除する変更です。Claude リーダーと GLM チームメイトのペインを維持する移行ではありません。

## 混合構成の検証条件

`claude-glm` は Claude リーダーと tmux 内の GLM チームメイトを表します。現在は TEAMMATE の統合検証を通過していないため、適用と起動は利用できず、プレビューのみ可能です。tmux のインストールや `verified: true` の設定では、この制限は解除されません。

```bash
moai migrate cg --target claude-glm
```

## 設定の保持とエラー対応

未定義の設定値、コメント、GLM モデル設定、認証情報への参照を保持します。適用前に元のバイト列を `.moai/backups/cg-migration/<source-sha256>.yaml` へ保存します。YAML の書式は変わることがあります。同じ移行先で再実行すると変更なしとなり、別の移行先への再移行は拒否されます。

gateway の値の競合、空でない `llm.mode`、重複キー、未対応の YAML エイリアスは明示的に解消してください。プレビューと事前検査の失敗では元の設定もバックアップディレクトリも変わりません。書き込み途中の失敗ではバックアップが残る場合があるため、エラーと保存された原本を確認してから再実行します。

## 認証情報の取り扱い {#tmux-env-security}

旧 CG の環境変数注入手順は、現在利用できるランチャーの動作ではありません。移行はプロバイダーを起動せず、認証情報も移動しません。バックアップには元の設定が含まれるため、アクセス権を制限してください。

## 次の手順

役割変更に同意して `claude-only` に移行した後は、対応するランチャーを明示的に選びます。`moai cc` と `moai glm` は旧ハイブリッド構成の役割を再現しません。GPT gateway の起動には別の統合検証が必要であり、この移行では有効になりません。

- [CLI](/ja/cli-reference/launchers/)
