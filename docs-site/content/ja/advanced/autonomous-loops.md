---
title: 自律連続ループ
weight: 6
draft: false
---

エージェント的ループの核心の問いは「いつ止まり、いつ続くか」です。MoAI-ADKは`/moai goal`と`/moai loop`の2つの連続ループプリミティブを提供し、Claude Code自身がネイティブのgoalコマンドを提供します。このページはこの3つを区別し、それぞれの所有権、実装状態、安全ガードレールを説明します。

## いつ止まり、いつ続くか

単一ターンで終わる作業もありますが、数十ターンにわたり収束が必要な作業もあります — 例えば「全テストがPASSするまで」や「診断ツールのイシューキューが空になるまで」。毎ターンユーザーがプロンプトを入力しなければならないなら、自律性の利点が失われます。

連続ループプリミティブはこの問題を解決します。完了条件を宣言すると、条件が満たされるかターン限度に達するまでセッションが自ら作業を続けます。

## 3つの連続ループプリミティブ

連続ループプリミティブは3つあり — 2つはMoAI-ADK、残る1つはClaude Code自身が所有します — それぞれトリガーセマンティクスが異なります。

| プリミティブ | 所有権 | トリガー | 適切な場合 |
|-------------|--------|---------|-----------|
| `/goal` | ユーザーTUI (HUMAN-ONLY) | モデルが条件を評価 | 「この条件が真になるまで続行」 |
| `/moai goal` | オーケストレータ (PROGRAMMATIC) | stop-goal Stop-hook評価 | MoAIパイプライン内の自律連続 |
| `/moai loop` | Ralph Engine (診断駆動) | 診断ツールのイシューキュー | 「ツールが見つけたイシューをすべて修正」 |

```mermaid
flowchart TD
    G["/goal — native Claude Code<br/>HUMAN-ONLY TUI command"]
    M["/moai goal — MoAI PROGRAMMATIC<br/>orchestrator-owned (Axis B)"]
    L["/moai loop — Ralph Engine<br/>diagnostic-driven preset"]

    G -->|同じセマンティクス、異なる所有権| M
    M -->|goalエンジン上のプリセット| L
```

### `/goal` — native Claude Code (HUMAN-ONLY)

{{< icon arrow-right >}} `/goal`はClaude CodeのネイティブTUIコマンドです。ユーザーが入力するコマンドであり、モデルがユーザーの代わりに呼び出すことはできません。これが**HUMAN-ONLY**制約です。

完了条件を宣言すると、各ターン終了後に小さな高速モデル(Haikuデフォルト)が条件が満たされたかを評価します。満たされていなければ別のターンを開始し、満たされていればループが終了します。

```text
/goal go test ./... exits 0 && lint is clean, or stop after 20 turns
```

条件は最大4,000文字まで可能で、ターン/時間限度を含めてループをバウンドできます。裸の`/goal`でステータスを確認し、`/goal clear`で早期終了できます。

### `/moai goal` — MoAI PROGRAMMATIC (Axis B)

{{< icon arrow-right >}} `/moai goal`はMoAIが所有するプログラミング的再実装です。ネイティブ`/goal`がHUMAN-ONLYであるため、オーケストレータがパイプライン内で自律連続ループを登録・武装(arm)できる唯一の経路です。

4つの動詞を提供します:

```bash
moai goal arm "<completion-condition>"  # 条件登録 + 武装
moai goal status                        # 現在の条件 + ターン/トークン消費確認
moai goal clear                         # 条件削除 (ループ終了)
```

セッション開始時に`PruneOrphans`が孤立goalをクリーンアップします。このメカニズムはSPEC-GOAL-ENGINE-001 (CLOSED)で実装されました。

### `/moai loop` — Ralph Engine (診断駆動プリセット)

{{< icon arrow-right >}} `/moai loop`は診断ツールが見つけたイシューキューをスキャンし、各イシューを修正し、キューが空になるか診断がクリーンになるまで繰り返す決定論的ループです。これはgoalエンジン上のプリセットです。

`/moai loop`は`/moai run --mode loop`のエイリアスではありません。`/moai run --mode loop`はランタイムモードディスパッチ値であり、`/moai loop`は独立したサブコマンドです。両者は同じgoalエンジンを使用しますが、エントリ経路とプリセット動作が異なります。

## ネイティブ /goal 詳細

`/goal <condition>`は完了条件を設定し、条件が真になるまでClaudeがプロンプトなしで作業を続けます。各ターン後、小さな高速モデルが条件を評価します。

効果的な条件の書き方:

- **1つの測定可能な終了状態** — テスト結果、ビルドexit code、ファイル数、空のキュー
- **明示された検証方法** — Claudeがどう証明すべきか(「`go test ./... exits 0`」)
- **重要な制約** — 経路で変更されないもの(「他のテストファイルは変更禁止」)

ターン限度を含めてループをバウンドしてください(「`or stop after 20 turns`」)。`/clear`を実行するとアクティブなgoalも削除されます。`--resume` / `--continue`でセッションを再開するとgoalが復元されます。

## 実装 vs ロードマップ

{{< icon warning warn >}} **REQ-DA-062正直性区別**: 3つのプリミティブの実装状態を明確に区別します。

- {{< icon check ok >}} `/goal` (native) — Claude Codeランタイムで実装 (v2.1.139+必要)
- {{< icon check ok >}} `/moai goal` (PROGRAMMATIC) — SPEC-GOAL-ENGINE-001 CLOSED、4動詞CLI完全実装
- {{< icon check ok >}} `/moai loop` (Ralph Engine) — 診断駆動ループとして実装完了
- {{< icon clock >}} AGENTIC-CORE Epic — 進行中。SPEC-1 (Analyze-Firstルーティング) CLOSED。SPEC-2 (自律/半自律kickoff REQ)はユーザー要求待機中。

## 安全ガードレール

{{< icon warning danger >}} 全ループプリミティブに対し安全ガードレールは変更されません。

- **Implementation Kickoff Approval** (plan → run HUMAN GATE)はどのループでもバイパス不可です。`/goal`がアクティブでもrun-phase進入前のユーザー承認は必須です。
- **安全境界 unchanged** — ループがアクティブでも「元に戻しにくい / 共有システム作業前の確認」境界は緩和されません。goal評価者は継続の可否のみを決定し、破壊的操作を事前承認しません。
- **auto modeとの組合せ** — Claude Code auto mode(ツールごとの自動承認)と`/moai goal`(ターンごとの連続)を組合せると無人`ac_converge`ループが可能です。auto modeはツールごとの承認プロンプトを削除し、`/moai goal`はターンごとのSTOPプロンプトを削除します。Implementation Kickoff Approvalはrun-phase進入前に依然必須です。

## マルチモデルレビューゲート (オプション)

{{< icon info >}} オプションの Stop フックが自律ループにクロスモデル敵対レビューを追加し、完全自律の `/moai goal` 実行にマルチモデル安全網を与えます (自律性再設計の Path C)。

### `audit_model: multi` 収束

`audit_model: multi` を選ぶと、`audit_multi` MCP ツールがアクティブなバックエンド全体に監査を並列展開します。claude がセッション内アンカー、codex と GLM がセカンダリ (各々 `audit_gate` に従う) となり、評決は 4 ステップ方針で収束します。

- 必須バックエンドのいずれかが `FAIL` → `overall_verdict = FAIL`。
- 必須バックエンドが全て `PASS` → `overall_verdict = PASS`。
- 必須バックエンド間の分裂 → 保守的 `FAIL` に `disagreement_flag` を付加。
- アドバイザリ専用バックエンドとの衝突 → `PASS` に `disagreement_flag` を付加。

不一致は `ConvergenceResult` の `disagreement_flag` と `residual_risk_note` で表面化し、それ自体がフローをハードブロックすることはありません。`disagreement_flag` は `participant_count`(比較可能な `pass`/`fail` 判定を出したバックエンドの数)と併せて3値を取ります。`true` は分かれが観測されたこと(単一参加者自身の合成内での分かれを含む)、`false` は参加者2以上の比較で分かれがなかったこと、`null` は参加者が2未満で一致も不一致も根拠づけられない状態です。独立性は構造的に保証されます。セカンダリ展開ゴルーチンは `(target, focus, model, effort)` のみを受け取り、`claude_verdict` を一切受け取りません。よって codex と GLM は汚染された再サンプルではなく相関のない第二意見を生成します。フェイルオープンは両方向で成り立ちます。欠落または未認証のオプショナルバックエンドは `VerdictInconclusive` を返し claude にフォールバックするだけで、ハードエラーになりません。

### `moai hook multi-review-gate` Stop フック

マルチレビューゲート Stop フックはオプションです (`workflow.multi_review_gate.enabled`、`workflow.codex.review_gate` の兄弟である BranchGuard パターン、デフォルトオフ)。moai デフォルトの 5 秒フックタイムアウトは 900 秒で上書きされます。コード編集ターンごとに直近の `ConvergenceResult` (収束エンジンが `.moai/state/audit-multi/<session>.json` に記録) を読み、標準の ALLOW/BLOCK 契約を出します。

- 必須バックエンドが全て PASS → ALLOW。
- 必須バックエンドのいずれかが FAIL → BLOCK。
- アドバイザリ専用の衝突 → ALLOW (不一致はアドバイザリとしてのみ表面化し、ブロックしない)。
- claude 以外のバックエンドが全て inconclusive → claude 評決にフェイルオープン。

必須セルフゲートは編集のないターンを即座に ALLOW します。ステータス報告、レビュー結果、その他の編集を伴わないターンが誤ってブロックされることはありません。

### どこに当てはまるか

{{< icon arrow-right >}} マルチレビューゲートは Stop フックであり、連続プリミティブではありません。`/moai goal` (ターンごとの連続) と `/moai loop` (診断駆動プリセット) の上に重ねて使います。ループがセッションを前に進め、ゲートがコード編集の境界ごとにクロスモデル収束契約を適用します。いかなる組合せでも run-phase 進入前の Implementation Kickoff Approval は引き続き必須です。

## 次のステップ

- [トークノミクス概論](/ja/advanced/tokenomics-overview/) — 自律ループがトークノミクスと接続するポイント
- [ハーネス自己進化](/ja/advanced/self-evolving/) — `/moai loop` / `/moai goal`収束軌跡がLoop 0観察に統合
