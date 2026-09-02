---
title: "Codex デュアルハーネス — AGENTS.md・エージェント二重公開・フックアダプター"
weight: 31
draft: false
added_in: "v3.1.3"
description: "codex-cli が MoAI-ADK を読めるようにする 4 つの成果物 — ルート AGENTS.md standing contract、エージェント TOML の二重公開、.agents/skills スキルミラー、internal/codexadapter フックアダプターライブラリ。"
---

MoAI-ADK の第 1 ハーネス(エージェントを実際に駆動する実行環境)は Claude Code ですが、v3.1.3 から **codex-cli でも同じ契約を読める二重の表面**を備えました。どれひとつとして Claude Code 側の動作を変えません — すでに存在していたルールとエージェント定義を、codex が探す場所と形式でもう一度公開しただけです。この文書はその 4 つの成果物が何で、それぞれがどんな問題を解くのかを扱います。

## ルート AGENTS.md — ハーネス共通の standing contract

リポジトリルートの `AGENTS.md` は Claude Code 専用ではなく、**どのエージェントハーネスがターンを駆動しても束ねる standing contract**(常時契約)です。1 ファイルで存在する理由は codex の読み方にあります: codex はプロジェクト指示文をバイト上限の中で読み、あふれた後ろの部分を **警告なく、終了コード 0 で静かに捨てます**。上限を超えた契約は、あたかも完全なものであるかのように報告されます。上限に収まること自体が要件であり、ビルドガード(ビルド時にこのファイルが上限以内かを検査する仕組み)がそれを守ります。

場所を作るため、常時ロードされていた文書 11 個は、8 個のレイジーコンパニオン(lazy companion、必要なときだけ読む詳細文書)を指すスタブ(短い要約)へと格下げされました。**移動したのは義務ではなく、その義務を説明する文章**です — 信頼できる源は引き続き `.claude/rules/moai/**` と `CLAUDE.md` であり、`AGENTS.md` はそれをハーネス中立な形で運びます。

{{< callout type="info" >}}
個人用の `~/.codex/AGENTS.md` は同じマージチェーンでこのファイルの**前に**消費され、プロジェクト契約が運べる幅を狭めます。あふれは後ろから静かに捨てられるため、このファイルの条項は最も重要なものから前に並んでいます。
{{< /callout >}}

## エージェント二重公開 — 11 個の TOML

維持される 11 個のエージェントが 2 つの形で公開されます。Claude Code 用の `.claude/agents/moai/*.md`(原本)と、codex が読む `.codex/agents/moai/*.toml`(派生)です。TOML は手書きされません — `internal/template/agentemit` がマークダウン原本から**決定論的に**(同じ入力には常に同じ出力)生成し、生成ファイルの先頭には "regenerate, do not edit"(再生成せよ、直接編集するな)と釘が刺さっています。

原本と派生がずれるのを 3 層のガードが防ぎます: ゴールデンファイル比較(期待出力との照合)、埋め込み検証(バイナリに組み込まれたテンプレートとの照合)、配備検証(ユーザーリポジトリに届く結果との照合)。マークダウンを直せば TOML が追従し、TOML だけを直せばガードが捕まえます。

## `.agents/skills` — スキルミラー

codex-cli は Claude Code の `.claude/skills/` を読まないため、スキルを `.agents/skills` 配下に**ミラー**(鏡の複製)として配備します。ミラーのリストは手管理ではなく、配備実行時点の実際のスキル集合から導出されるため、スキルが増減してもリストはずれません。このディレクトリは**ユーザーリポジトリの外**に向かう配備成果物という扱いで git には記録されず、シンボリックリンクを優先しつつリンクを作れない環境ではコピーで代用配備されます(`moai init`・`moai update` の完了要約がその事実を知らせます — 詳しくは [moai update](/ja/cli-reference/update/) の文書を参照)。

## `internal/codexadapter` — フックアダプターライブラリ

2 つのハーネスのフック表面はほぼ同じですが、完全には同じではありません。実測(codex-cli 0.147.0 基準)で分かれた地点はちょうど 2 つ: ハーネスが渡す**イベント名**と、codex が宣言はするが実際には反応しない**出力キー 3 つ**(`systemMessage`・`continue`・`stopReason`)です。それ以外はすべて同一に測定されたため、`internal/codexadapter` はディスパッチャーの**前に**座る薄い翻訳層で、`internal/hook` には触れません。

### 11 イベント表

| Codex イベント | MoAI ディスパッチャー引数 | このマイルストンで適応? |
|---|---|---|
| PreToolUse | `pre-tool` | はい |
| PostToolUse | `post-tool` | はい |
| SessionStart | `session-start` | はい |
| SessionEnd | `session-end` | はい |
| Stop | `stop` | はい |
| UserPromptSubmit | `user-prompt-submit` | はい |
| PreCompact | `compact` | いいえ — 未計測 |
| PostCompact | `post-compact` | いいえ — 未計測 |
| PermissionRequest | `permission-request` | いいえ — 未計測 |
| SubagentStart | `subagent-start` | いいえ — 未計測 |
| SubagentStop | `subagent-stop` | いいえ — 実測で発火しない |

11 個すべてにディスパッチャーの対応物が存在します。適応から除外するのは計測カバー範囲に関するスコープ決定であって、対応物の欠如ではありません。SubagentStop が特別な理由: 実測で**一度も発火しませんでした** — codex では委任はツール名が "collaboration" で始まる PostToolUse として現れるため、これをつなぐと決して流れない経路に線を敷くことになります。

未適応イベントは黙殺されず**拒否**されます。未知のイベント(タイプミス)と、認識はされるが今回は扱わないイベント(スコープ決定)が異なるエラーで区別されるため、運用者はミスと決定を見分けられます。設定バリデータは不明キー違反を最初の 1 件で止まらず**全部収集して**一度に示します。

### まだ呼び出し元がない

{{< icon warning warn >}} このパッケージはライブラリとして出荷された状態で、**まだ何もこれを呼び出していません**。生成された codex 設定にアダプターを配線する後続カードである `--agent` 設定ジェネレーターが接続を完成させます。現時点でこの文書は配線が入る場所の地図であって、スイッチの入った機能の取扱説明書ではありません。

## 次のステップ

- [マルチモデル監査収束](/ja/advanced/multi-model-audit/) — codex バックエンドが今日すでに監査に参加している経路
- [moai update](/ja/cli-reference/update/) — スキルミラーの symlink・コピー配備とその通知
- [エージェントガイド](/ja/advanced/agent-guide/) — 二重公開される 11 個のエージェントの役割
