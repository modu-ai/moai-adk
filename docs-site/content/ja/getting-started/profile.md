---
title: プロファイル管理
weight: 80
draft: false
---
# プロファイル管理


MoAI-ADK のプロファイルシステムで、複数の Claude Code 設定を分離して管理します。仕事用と個人用、高品質セッションとコスト削減セッションをそれぞれ1つのプロファイルに分けておけば、モデル・言語・表示設定を毎回変更する必要がありません。

## プロファイルとは?

プロファイルは **分離された Claude Code 設定ディレクトリ**(`CLAUDE_CONFIG_DIR`)です。プロファイルごとに独立した設定、モデル選択、言語環境を維持できます。

```
~/.moai/claude-profiles/
├── default/           # 既定のプロファイル
│   ├── settings.json
│   └── settings.local.json
├── work/              # 仕事用プロファイル
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

インタラクティブ設定ウィザードを実行します。

```bash
moai profile setup          # 既定のプロファイルを設定
moai profile setup work     # "work" プロファイルを設定
```

**ウィザードの設定項目:**
- **Identity**: ユーザー名、役割
- **Languages**: 会話言語、コードコメント言語
- **Model Settings**: 既定モデル、1M コンテキストモデルの選択
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

## プロファイルで Claude Code を実行

`-p` (または `--profile`) フラグでプロファイルを指定します。

```bash
moai cc -p work          # work プロファイルで Claude を実行
moai glm -p cost-save    # cost-save プロファイルで GLM を実行
moai cg -p team          # team プロファイルで CG モードを実行
```

{{< callout type="info" >}}
プロファイル未指定時は既定のプロファイルが使用されます。初回実行時には設定ウィザードが自動的に起動します。
{{< /callout >}}

## 1M コンテキストモデルの選択

プロファイル設定時に、1M コンテキストウィンドウをサポートするモデルを選択できます。`[1m]` サフィックスは別のモデルではなく、Claude Code のネイティブなコンテキストウィンドウ修飾子です。

**選択可能なモデルエイリアス:**
- `opus` / `opus[1m]`
- `sonnet` / `sonnet[1m]`
- `fable` / `fable[1m]`
- `haiku`, `opusplan`

設定ウィザードの「Model Settings」ステップで選択するか、プロファイル設定ファイルを直接編集します。1M コンテキストモデルは、大規模コードベースの分析や長いドキュメント作業に適しています。

## プロファイル切り替え時の動作

| 切り替え | 動作 |
|------|------|
| `moai cc` → `moai glm` | GLM 環境変数を自動注入 |
| `moai glm` → `moai cc` | GLM 環境変数を自動削除 |
| `moai cc` → `moai cg` | GLM env を tmux セッションにのみ注入、Leader は Claude を維持 |

## 関連ドキュメント

- [CLI リファレンス](/getting-started/cli) - 全 CLI コマンド
- [クイックスタート](/getting-started/quickstart) - 最初のスタート
- [初期設定](/getting-started/init-wizard) - プロジェクトの初期化
