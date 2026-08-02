---
title: プロファイル管理
weight: 40
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

この値はグローバル記録を基準とするため、プロジェクトごとに記憶するプロファイルが異なる場合は [プロファイルの自動選択](#プロファイルの自動選択) の制約と併せて参照してください。

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
`-p` で指定したプロファイルは常に優先されます。指定しなかった場合にどのプロファイルが使われるかは、下の [プロファイルの自動選択](#プロファイルの自動選択) を参照してください。プロファイルを初めて使うときは設定ウィザードが自動的に開始します。
{{< /callout >}}

## プロファイルの自動選択

`-p` なしで `moai cc` を実行すると、`~/.moai/claude-profiles/launch.yaml` の記録を参照してプロファイルを選びます。この記録は `-p` で名前付きプロファイルを実行するたびに更新されます。

{{< callout type="note" >}}
以下で説明するプロジェクト別の記憶は次のリリースに含まれます。現在配布されているバージョンはグローバル記録 1 つ (`last_profile`) だけを残すため、プロジェクト B で `-p` によりプロファイルを指定すると、プロジェクト A が記憶していた値が上書きされます。
{{< /callout >}}

`launch.yaml` はグローバル記録とあわせて、プロジェクトの絶対パスをキーとする `projects:` の一覧を保持します。`-p` なしの実行がプロファイルを決める順序は次のとおりです。

1. 現在のプロジェクトが記憶するプロファイル (`projects:` の項目)
2. グローバル記録 (`last_profile`)
3. デフォルトプロファイル

記録されたプロファイルでも、ディレクトリがすでに削除されていれば読み飛ばして次の順序に進みます。`-p` で指定した名前はこの順序全体より優先され、`-p default` でデフォルトプロファイルを明示することもできます。

両方の参照をまとめて無効にするには、環境変数を指定します。

```bash
MOAI_NO_PROFILE_FALLBACK=1 moai cc    # 記録を無視してデフォルトプロファイルで実行
```

プロジェクト別の記録は `-p` で実行したときに残り、[Web コンソール](/ja/cli-reference/web) でプロファイルを切り替えたときにも併せて更新されます。デフォルトプロファイル (`default`) は記録の対象外です。

**知っておきたい制約**

- プロジェクトディレクトリを移動したり名前を変えたりすると、既存の項目はどのパスとも一致しなくなります。この項目は静かに読み飛ばされるため、実行に問題を起こすことはありません。
- `projects:` の一覧はプロジェクトが増えるほど一緒に増えていき、整理するコマンドはまだありません。
- `moai profile current` はグローバル記録をそのまま表示します。そのため、記憶されたプロファイルがグローバル記録と異なるプロジェクトでは、`moai profile current` が知らせる名前と、`-p` なしの `moai cc` が実際に起動するプロファイルが食い違うことがあります。

## 新しいプロファイルの初回実行

新しく作ったプロファイルディレクトリには、Claude Code のアカウント状態を格納する `.claude.json` がまだありません。アカウント状態はどのプラットフォームでも設定ディレクトリごとに個別管理されるため、使っていたセッションが問題なくても、新しいプロファイルで初めて実行するとログイン・オンボーディング画面が表示されます。

{{< callout type="note" >}}
以下の案内メッセージは次のリリースに含まれます。現在配布されているバージョンは何の予告もなくログイン画面に切り替わります。
{{< /callout >}}

ランチャーは Claude Code を起動する前に、標準エラーへ次を知らせます。

```
Notice: profile "work" has no Claude Code configuration yet.
  Claude Code will show the login / onboarding screen on this launch.
  Account state is not inherited between profiles; sign in once and it
  persists for this profile.
```

認証情報を新しいプロファイルへコピーしたり移したりする動作は行いません。アカウント状態を格納する場所はプラットフォームごとに異なり、一方に合わせたコピーはもう一方では食い違うためです。一度ログインすればその状態が当該プロファイルに残るので、次回以降の実行でこの画面は表示されません。

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

- [moai web コンソール](/ja/cli-reference/web) - ブラウザでプロファイルを切り替え・編集
- [CLI リファレンス](/getting-started/cli) - すべての CLI コマンド
- [クイックスタート](/getting-started/quickstart) - はじめて始める
- [初期設定](/getting-started/init-wizard) - プロジェクト初期化
