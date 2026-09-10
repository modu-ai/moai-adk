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

## エージェント二重公開 — 11 個の TOML

維持される 11 個のエージェントが 2 つの形で公開されます。Claude Code 用の `.claude/agents/moai/*.md`(原本)と、codex が読む `.codex/agents/moai/*.toml`(派生)です。TOML は手書きされません — `internal/template/agentemit` がマークダウン原本から**決定論的に**(同じ入力には常に同じ出力)生成し、生成ファイルの先頭には "regenerate, do not edit"(再生成せよ、直接編集するな)と釘が刺さっています。

原本と派生がずれるのを 3 層のガードが防ぎます: ゴールデンファイル比較(期待出力との照合)、埋め込み検証(バイナリに組み込まれたテンプレートとの照合)、配備検証(ユーザーリポジトリに届く結果との照合)。マークダウンを直せば TOML が追従し、TOML だけを直せばガードが捕まえます。

## `.agents/skills` — スキルミラー

codex-cli は Claude Code の `.claude/skills/` を読まないため、スキルを `.agents/skills` 配下に**ミラー**(鏡の複製)として配備します。ミラーのリストは手管理ではなく、配備実行時点の実際のスキル集合から導出されるため、スキルが増減してもリストはずれません。このディレクトリは**ユーザーリポジトリの外**に向かう配備成果物という扱いで git には記録されず、シンボリックリンクを優先しつつリンクを作れない環境ではコピーで代用配備されます(`moai init`・`moai update` の完了要約がその事実を知らせます — 詳しくは [moai update](/ja/cli-reference/update/) の文書を参照)。

## ハーネス別の個人指示

`AGENTS.local.md` は Codex 専用です。ローカルの `moai codex` ランチャーがプロジェクトルートから読み取り、内容をそのまま Codex セッションの `developer_instructions` 上書きとして渡します。共有ファイルから `@` で取り込むことはありません。`CLAUDE.local.md`、`.claude/settings.local.json`、Claude の自動 `MEMORY.md` は Claude 専用のままです。Codex Web セッションはローカルランチャーを通らないため、この注入を受けません。

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

## 次のステップ

- [マルチモデル監査収束](/ja/advanced/multi-model-audit/) — codex バックエンドが今日すでに監査に参加している経路
- [moai update](/ja/cli-reference/update/) — スキルミラーの symlink・コピー配備とその通知
- [エージェントガイド](/ja/advanced/agent-guide/) — 二重公開される 11 個のエージェントの役割
