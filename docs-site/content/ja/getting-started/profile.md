---
title: プロファイル管理
weight: 80
draft: false
---
# プロファイル管理


MoAI-ADKのプロファイルシステムで、複数のClaude Code設定を分離して管理します。

## プロファイルとは？

プロファイルは **分離されたClaude Code設定ディレクトリ**(`CLAUDE_CONFIG_DIR`)です。プロファイルごとに独立した設定、モデル選択、言語環境を維持できます。

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
moai profile setup          # デフォルトプロファイルの設定
moai profile setup work     # "work"プロファイルの設定
```

**ウィザード設定項目:**
- **Identity**: ユーザー名、役割
- **Languages**: 対話言語、コードコメント言語
- **Model Settings**: デフォルトモデル、1Mコンテキストモデルの選択
- **Display**: 出力スタイル、ステータスライン設定

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

## プロファイルでClaude Codeを実行

`-p` (または`--profile`) フラグでプロファイルを指定します。

```bash
moai cc -p work          # workプロファイルでClaudeを実行
moai glm -p cost-save    # cost-saveプロファイルでGLMを実行
moai cg -p team          # teamプロファイルでCGモードを実行
```

{{< callout type="info" >}}
プロファイルを指定しない場合はデフォルトプロファイルが使用されます。初回実行時には自動的に設定ウィザードが起動します。
{{< /callout >}}

## 1Mコンテキストモデルの選択

プロファイル設定時に、1Mコンテキストウィンドウをサポートするモデルを選択できます。

**サポートされているモデル:**
- `claude-opus-4-8[1m]` - Opus 4.8 (1M context)
- `claude-sonnet-4-6[1m]` - Sonnet 4.6 (1M context)

設定ウィザードの「Model Settings」ステップで選択するか、プロファイル設定ファイルを直接編集します。

## プロファイル切り替え時の動作

| 切り替え | 動作 |
|------|------|
| `moai cc` → `moai glm` | GLM環境変数が自動的に注入される |
| `moai glm` → `moai cc` | GLM環境変数が自動的に削除される |
| `moai cc` → `moai cg` | GLM envはtmuxセッションにのみ注入され、Leaderは引き続きClaudeを使用 |

## 関連ドキュメント

- [CLIレファレンス](/getting-started/cli) - CLIコマンドの全リファレンス
- [クイックスタート](/getting-started/quickstart) - はじめての利用開始
- [初期設定](/getting-started/init-wizard) - プロジェクトの初期化
