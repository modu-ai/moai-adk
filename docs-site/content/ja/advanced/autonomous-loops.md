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
moai goal arm "<completion-condition>"  # 条件登録 + 武装 (arm 専用)
moai goal status                        # 現在の条件 + ターン/トークン消費確認
moai goal clear                         # 条件削除 (ループ終了)
moai goal render                        # 現在の goal ダッシュボードを HTML でレンダリング
```

> **arm 専用属性**: `arm` は条件を登録して有効化するだけで、それ自体は作業を開始しません。arm された goal は毎ターン終了時に `stop-goal` Stop-hook 評価者が条件が満たされたか判定し、次のターンを続けるかを決めます。実際の作業開始コマンド(例: `/moai run SPEC-XXX`)と組み合わせて使う必要があります — arm だけ立てて作業コマンドがないとターンだけを消費します。

### 無限 goal とブロックキャップ (SPEC-INFINITE-GOAL-001)

`--max-turns 0` を指定するとターン上限がなくなる**無限 goal** になります。無限 goal は必ず `--max-duration <sec>` (壁時間上限、秒単位) とペアにする必要があります — arm 時点で fail-closed として拒否されます。実際の bound なしに無限に放置すると安全ガードが成立しないためです。

- `--cost-cap <value>` は**記録専用 (recorded-only)** — 呼び出し回数上限として保存されるだけ、現在 enforce ロジックがないため実際の bound として機能しません。そのため `--max-turns 0` に要求される実際の bound 要件を cost-cap 単独では満たせず、拒否されます。
- **ブロックキャップの先取り**: Claude Code ランタイムの連続 block キャップ `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP` (デフォルト 8) がターン上限より先にループを切ります。無限 goal (`--max-turns 0`) を正しく回すにはこのキャップを上げる必要があります(例: `200`)。`moai cc` / `moai cg` ランチャーは無限 goal が arm されたセッションを開始するとき、この値を自動的に注入します。既に起動しているセッションでは arm 前に環境変数を直接設定してください。
- **ターン上限到達時の判定**: ターン上限(またはブロックキャップ)に達すると、評価者は 5 セクションの判定文 (verdict) を出します — `Claim / Evidence / Baseline-attribution / Gaps / Residual-risk`。この判定文は「収束した」というシグナルではなく「上限に達して止まった」という報告です。
- **Progression Mode**: 自律 (autonomous) vs 半自律 (semi-autonomous) の選択は Implementation Kickoff Approval ゲートで行われます。arm 自体がこのゲートを飛び越えることはありません。

セッション開始時に`PruneOrphans`が孤立goalをクリーンアップします。このメカニズムはSPEC-GOAL-ENGINE-001 (CLOSED)で実装されました。

現在のループ状態を静的HTMLダッシュボードとしてレンダリングするには `moai goal render` を使います — 詳細は [/moai goal - ゴールダッシュボード](/ja/utility-commands/moai-goal/) を参照してください。

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

## 次のステップ

- [トークノミクス概論](/ja/advanced/tokenomics-overview/) — 自律ループがトークノミクスと接続するポイント
- [ハーネス自己進化](/ja/advanced/self-evolving/) — `/moai loop` / `/moai goal`収束軌跡がLoop 0観察に統合
