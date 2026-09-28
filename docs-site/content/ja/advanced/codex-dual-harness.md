---
title: "Codex デュアルハーネス — AGENTS.md・エージェント二重公開・フックアダプター"
weight: 31
draft: false
added_in: "v3.1.3"
description: "codex-cli が Claude Code と並行して MoAI-ADK を使うための共通面と、ハーネス別の個人指示。"
---

MoAI-ADK の第 1 ハーネス(エージェントを実際に駆動する実行環境)は Claude Code ですが、v3.1.3 から **codex-cli でも同じ契約を読める二重の表面**を備えました。共通ルールとエージェント定義は Codex が探す場所と形式でも公開し、個人指示はハーネスごとに分離します。この文書では、各表面がどの問題を解くのかを説明します。

## ルート AGENTS.md — ハーネス共通の standing contract

リポジトリルートの `AGENTS.md` は Claude Code 専用ではなく、**どのエージェントハーネスがターンを駆動しても束ねる standing contract**(常時契約)です。1 ファイルで存在する理由は codex の読み方にあります: codex はプロジェクト指示文をバイト上限の中で読み、あふれた後ろの部分を **警告なく、終了コード 0 で静かに捨てます**。上限を超えた契約は、あたかも完全なものであるかのように報告されます。上限に収まること自体が要件であり、ビルドガード(ビルド時にこのファイルが上限以内かを検査する仕組み)がそれを守ります。

場所を作るため、常時ロードされていた文書 11 個は、8 個のレイジーコンパニオン(lazy companion、必要なときだけ読む詳細文書)を指すスタブ(短い要約)へと格下げされました。**移動したのは義務ではなく、その義務を説明する文章**です — `AGENTS.md` がハーネス共通契約の基準であり、`.claude/rules/moai/**` と `CLAUDE.md` は Claude 専用の仕組みを補足します。

{{< callout type="info" >}}
個人用の `~/.codex/AGENTS.md` は同じマージチェーンでこのファイルの**前に**消費され、プロジェクト契約が運べる幅を狭めます。あふれは後ろから静かに捨てられるため、このファイルの条項は最も重要なものから前に並んでいます。
{{< /callout >}}

## エージェント二重公開 — 12 個の TOML

維持される 12 個の MoAI カスタムエージェントが 2 つの形で公開されます。Claude Code 用の `.claude/agents/moai/*.md`(原本)と、codex が読む `.codex/agents/moai/*.toml`(派生)です。TOML は手書きされません — `internal/template/agentemit` がマークダウン原本から**決定論的に**(同じ入力には常に同じ出力)生成し、生成ファイルの先頭には "regenerate, do not edit"(再生成せよ、直接編集するな)と釘が刺さっています。

原本と派生がずれるのを 3 層のガードが防ぎます: ゴールデンファイル比較(期待出力との照合)、埋め込み検証(バイナリに組み込まれたテンプレートとの照合)、配備検証(ユーザーリポジトリに届く結果との照合)。マークダウンを直せば TOML が追従し、TOML だけを直せばガードが捕まえます。

## `.agents/skills` — スキルミラー

codex-cli は Claude Code の `.claude/skills/` を読まないため、スキルを `.agents/skills` 配下に**ミラー**(鏡の複製)として配備します。ミラーのリストは手管理ではなく、配備実行時点の実際のスキル集合から導出されるため、スキルが増減してもリストはずれません。このディレクトリは**ユーザーリポジトリの外**に向かう配備成果物という扱いで git には記録されず、シンボリックリンクを優先しつつリンクを作れない環境ではコピーで代用配備されます(`moai init`・`moai update` の完了要約がその事実を知らせます — 詳しくは [moai update](/ja/cli-reference/update/) の文書を参照)。

## ハーネス別の個人指示

`AGENTS.local.md` は両方のハーネスが読む個人の指示ファイルで、`CLAUDE.local.md` はその以前の名前として、フォールバックの入力にだけ残っています。ローカルの `moai codex` は、引数なし、`cli`、`app`、`--spawn`、`-w` のどの起動経路でも、プロジェクトルートにある空でない通常ファイルをこの順に読み込みます。各本文の前に `<!-- source: <filename> -->` という出典ヘッダーを置き、結合した内容を 1 つの `developer_instructions` 上書きとして渡します。`-w` でも Codex の実行場所がワークツリーに移るだけで、入力元は元のプロジェクトルートのままです。共有の `AGENTS.md` と `CLAUDE.md` は、これらのローカルファイルを取り込んだりリンクしたりしません。ランチャーはリンクや通常ファイルでない入力を拒否し、検査した同じファイルディスクリプターから読み込みます。また、オペレーターが `developer_instructions` を重ねて指定した場合や direct/spawn 引数が大きすぎる場合は、起動前に失敗します。ほかのハーネス固有設定とメモリは、それぞれのハーネスだけに残ります。Codex Web はローカルランチャーを通らないため、この注入を受けません。

## `CLAUDE.local.md` のフォールバックと移行の案内

`AGENTS.local.md` がなく `CLAUDE.local.md` だけがあるプロジェクトでも、`moai codex` はそのファイルをフォールバックとして読み、`developer_instructions` に入れます。出典ヘッダーには実際に読んだファイル名がそのまま入ります — この場合は `<!-- source: CLAUDE.local.md -->` です。

フォールバックを通ったとき、ランチャーは標準エラーに移行の案内を 1 行出力します。

```text
Advisory: CLAUDE.local.md is a legacy local instruction file; run `moai migrate local-instructions` to move it to AGENTS.local.md.
```

この案内はオペレーター向けのメッセージなので、モデルのコンテキストになる `developer_instructions` には入りません。2 つのファイルが両方ある場合はどちらも読み込まれ、案内は手作業でまとめるよう促す文面に変わります。共通の契約は `AGENTS.md`、Claude 専用の層は `CLAUDE.md`、個人の指示は `AGENTS.local.md` という 3 ファイル構成の中で、`CLAUDE.local.md` は移行を待つ以前の名前にすぎません。

## `internal/codexadapter` — フックアダプターライブラリ

2 つのハーネスのフック表面はほぼ同じですが、完全には同じではありません。実測(codex-cli 0.153.4 基準)で分かれた地点は 3 つ: ハーネスが渡す**イベント名**、codex が宣言はするが実際には反応しない**出力キー 3 つ**(`systemMessage`・`continue`・`stopReason`)、そして **PreToolUse 決定契約**です — codex パーサーは `updatedInput` のない `permissionDecision:allow` と `permissionDecision:ask` を拒否します。`internal/codexadapter` はディスパッチャーの**前に**座る薄い翻訳層で(`internal/hook` には触れない)、拒否される決定形(allow・ask・defer)は無意見の `{}` に劣化させて codex 自身の承認フローに委ね、各劣化は discard シンクで通知され、理由のない deny には既定理由が補われます。

### 12 イベント表

| Codex イベント | MoAI ディスパッチャー引数 | このマイルストンで適応? |
|---|---|---|
| PreToolUse | `pre-tool` | はい |
| PostToolUse | `post-tool` | はい |
| SessionStart | `session-start` | はい |
| SessionEnd | `session-end` | はい |
| Stop | `stop` | はい |
| UserPromptSubmit | `user-prompt-submit` | はい |
| PreCompact | `compact` | いいえ — 非対話実行では圧縮が一度も発生せず |
| PostCompact | `post-compact` | いいえ — 非対話実行では圧縮が一度も発生せず |
| PermissionRequest | `permission-request` | いいえ — 非対話実行では承認要求が発生せず |
| SubagentStart | `subagent-start` | はい |
| SubagentStop | `subagent-stop` | はい |
| Interrupt | — (対応物なし) | いいえ — SIGINT で発火。適応には新しいディスパッチャー サブコマンドが必要(後続カード) |

12 個のイベントを扱います: 11 個にはディスパッチャーの対応物が存在し、公式に文書化された 12 番目のイベント `Interrupt` は対応物ができるまで、対応物がないことを伝う別のエラーで認識・拒否されます。codex-cli 0.153.4 の実測で `SubagentStart` と `SubagentStop` は**発火が確認され**(SubagentStop は発火しないという以前の 0.147.0 観測を覆す結果)、現在は適応済みです — `RenderHooks` がユーザーの `.codex/hooks.json` に `moai hook subagent-start --harness codex` と `moai hook subagent-stop --harness codex` の行を書き込みます。compact・permission 系は推定ではなく誠実な根拠で保留されています: 非対話の `codex exec` 実行は圧縮に届かず(実測の入力上限 1,048,576 文字に対し最善で 264,808 入力トークン)、承認要求も引き出せませんでした — 「発火しない」ではなく**トリガー未達成**として記録されます。

未適応イベントは黙殺されず**拒否**されます。未知のイベント(タイプミス)と、認識はされるが扱わないイベント(スコープ決定、または `Interrupt` のような対応物の欠如)が異なるエラーで区別されるため、運用者はミスと決定を見分けられます。設定バリデータは不明キー違反を最初の 1 件で止まらず**全部収集して**一度に示します。

### 現在の呼び出し元

`RenderHooks` は、適応済みの 8 イベントのコマンドをユーザーの `.codex/hooks.json` に書き込みます。`moai init --llm codex|both` がこの配線を作成し、既存プロジェクトでは `moai tool enable codex` で追加または更新します。

## デスクトップアプリと v3.1.3 のリリース条件

Codex アプリのローカルセッションでは、生成された `.codex/hooks.json` と `.codex/config.toml` を含むプロジェクトを開きます。Codex がプロジェクトのフックと設定を読み込むには、プロジェクトの信頼が必要です。Codex CLI の `/hooks` でフックを確認し、定義が変わった場合は再び信頼してください。同じプロジェクトで `moai doctor` を実行すると MoAI の配線状態を確認できます。アプリの実行環境から `moai` コマンドを実行できることも必要です。信頼の条件は [Codex のフック資料](https://developers.openai.com/codex/hooks)を参照してください。

Claude Code デスクトップアプリの **Local** Code セッションは、CLI と同じプロジェクトの `CLAUDE.md`、`.mcp.json`、フック、スキル、設定を読み込みます。初期化済みのプロジェクトを Code タブで開き、そのセッションから `moai` を実行できることを確認してください。SSH セッションは接続先で動くため、接続先にもプロジェクトと `moai` が必要です。詳しくは [Claude Code Desktop の設定](https://code.claude.com/docs/en/desktop#shared-configuration)を参照してください。

ここで示したのは設定手順であり、v3.1.3 のデスクトップ版リリース判定が完了したという意味ではありません。運用者自身が MoAI、Codex、両デスクトップアプリを検証し、結果を宣言するまでリリースは保留です。

## 次のステップ

- [マルチモデル監査収束](/ja/advanced/multi-model-audit/) — codex バックエンドが今日すでに監査に参加している経路
- [moai update](/ja/cli-reference/update/) — スキルミラーの symlink・コピー配備とその通知
- [エージェントガイド](/ja/advanced/agent-guide/) — 二重公開される 12 個のエージェントの役割
