---
title: プロファイル管理
weight: 80
draft: false
---

MoAI-ADK のプロファイルシステムで複数の Claude Code 設定を隔離して管理します。業務用と個人用、高品質セッションとコスト削減セッションをプロファイル 1 つずつに分けておけば、モデル・言語・表示設定を毎回変える必要がありません。

## プロファイルとは?

プロファイルは **隔離された Claude Code 設定ディレクトリ** (`CLAUDE_CONFIG_DIR`) です。プロファイルごとに独立した設定、モデル選択、言語環境を維持できます。

```
~/.moai/claude-profiles/
├── default/           # デフォルトプロファイル
│   ├── settings.json
│   └── settings.local.json
├── work/              # 業務用プロファイル
│   ├── settings.json
│   └── settings.local.json
└── personal/          # 個人用プロファイル
    └── ...
```

## コマンドリファレンス

### moai profile list

利用可能なすべてのプロファイルを表示します。

```bash
moai profile list
```

### moai profile setup [name]

インタラクティブな設定ウィザードを実行します。

```bash
moai profile setup          # デフォルトプロファイル設定
moai profile setup work     # "work" プロファイル設定
```

**ウィザードの設定項目:**
- **Identity**: ユーザー名、役割
- **Languages**: 会話言語、コードコメント言語
- **Model Settings**: デフォルトモデル、1M コンテキストモデル選択
- **Display**: 出力スタイル、ステータスバー設定

### moai profile current

現在アクティブなプロファイル名を表示します。

```bash
moai profile current
```

### moai profile delete [name]

プロファイルを削除します。

```bash
moai profile delete old-profile
```

## プロファイルで Claude Code を実行

`-p` (または `--profile`) フラグでプロファイルを指定します。

```bash
moai cc -p work          # work プロファイルで Claude 実行
moai glm -p cost-save    # cost-save プロファイルで GLM 実行
moai cg -p team          # team プロファイルで CG モード実行
```

{{< callout type="info" >}}
プロファイル未指定時はデフォルトプロファイルが使われます。初回実行時に自動的に設定ウィザードが開始します。
{{< /callout >}}

## 1M コンテキストモデル選択

プロファイル設定時に 1M コンテキストウィンドウをサポートするモデルを選択できます。`[1m]` 接尾辞は別のモデルではなく、Claude Code のネイティブなコンテキストウィンドウ修飾子です。

**選択可能なモデルエイリアス:**
- `opus` / `opus[1m]`
- `sonnet` / `sonnet[1m]`
- `fable` / `fable[1m]`

設定ウィザードの "Model Settings" ステップで選択するか、プロファイル設定ファイルを直接修正します。1M コンテキストモデルは大容量コードベースの分析や長いドキュメント作業に適しています。

## プロファイル切替時の動作

| 切替 | 動作 |
|------|------|
| `moai cc` → `moai glm` | GLM 環境変数を自動注入 |
| `moai glm` → `moai cc` | GLM 環境変数を自動除去 |
| `moai cc` → `moai cg` | GLM env を tmux セッションのみに注入、Leader は Claude を維持 |

## 関連ドキュメント

- [CLI リファレンス](/getting-started/cli) - すべての CLI コマンド
- [クイックスタート](/getting-started/quickstart) - はじめて始める
- [初期設定](/getting-started/init-wizard) - プロジェクト初期化
