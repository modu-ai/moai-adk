---
title: /moai goal
weight: 25
draft: false
---

完了条件を宣言すると、セッションがその条件を満たすまで自ら働く **条件宣言型の自律ループ** コマンドです。`/moai goal "<条件>"` で完了条件を arm すると、毎ターン終了時に `stop-goal` Stop フックが条件充足の可否を評価し、満たされるまで次のターンを自動的に開始します。

{{< callout type="info" >}}
**一行要約**: `/moai goal` は「終わりの状態を宣言する汎用ループ」です。`/moai loop` が「診断ツールが見つけた問題を全部なくすまで」という条件があらかじめ決まっているプリセットだとすれば、`/moai goal` は完了条件を **直接宣言する** 汎用エンジンです。
{{< /callout >}}

{{< callout type="info" >}}
**プログラマティックコマンド**: ネイティブの Claude Code `/goal` はユーザーだけが入力できる (HUMAN-ONLY) TUI コマンドです。`/moai goal` は同じ意味を **パイプラインからプログラマティックに** 実装した MoAI 所有コマンドで、`moai` スキルルーティングと `moai goal` CLI を通じて進入します。
{{< /callout >}}

## 概要

エージェントに「この条件が満たされるまで任せて働き続けて」と指示したいときに使います。条件は 2 種類を混ぜて使えます。

- **機械的条件 (mechanical)**: シェルコマンドで検証される条件。例: `go test ./... exits 0`。コマンドを実行して終了コードを観察します。
- **モデル評価条件 (model-evaluated)**: トランスクリプトに対する判断で検証される条件。例: `すべての AC 行が PASS として記録される`。セッションがこれまでに残した内容を根拠に評価します。

このループが v3 の 2 つ目の柱、**エージェンティックループエンジニアリング** の汎用エンジンです。goal 状態は `.moai/state/goal/<session-id>.json` にセッションごとに保存され (共有ファイルではない)、**ターン上限 (デフォルト 30)** がループを有界にします。上限に達すると評価器は 5 セクション判定 (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk) を出し、ブロッキングを止めます。`--max-turns 0` を指定すると、コンパクション境界を越えて持続する無限 goal が回り、ターン数の代わりに `--max-duration` (実時間) と停滞ガードが実際の上限になる。実上限なしに `--max-turns 0` を arm すると arm 時に拒否される (fail-closed)。

## 動詞 (verbs)

### `/moai goal "<条件>"` — 登録 + arm

条件テキストを登録し、アクティブセッションに goal を arm します。条件は `conditions[]` 配列としてパースされ、純粋なシェルコマンド文字列は機械的条件、トランスクリプトを参照する主張はモデル条件となります。arm すると `.moai/state/goal/<session-id>.json` がアトミックに (temp+rename) 記録され、`stop-goal` Stop フックが次のターン終了時にこれを拾って評価を開始します。

```bash
> /moai goal "go test ./... exits 0; すべての AC が PASS として記録、または 30 ターン後に中断"
```

### `/moai goal status [--all]`

アクティブセッションの goal (または `--all` ですべてのセッションの goal) を出力します。条件テキスト、conditions 配列、使用したターン数と上限、進行ログ、ライフサイクル状態 (`armed` / `satisfied` / `ceiling-exit` / `cleared`) を表示します。

### `/moai goal clear`

アクティブセッションの goal を解除します (状態ファイル削除)。Stop フックは arm された goal がないことを見てブロッキングを止めます。オーケストレーターがモデル条件を充足と判定した後にループを終える方法です。

{{< callout type="info" >}}
**`resume` 動詞は提供されません。** かつて議論されていた `resume` (解除された goal をアーカイブから復元する) 動詞は現在の CLI にはありません。`moai goal --help` は `arm` / `status` / `clear` のみを列挙します。`clear` が状態ファイルを **削除** するため (アーカイブに tombstone しない)、復元する原本が残りません。
{{< /callout >}}

## `--auto` ミッションモード

```bash
moai goal --auto --session <session-id> "承認済みの範囲で機能を実装し検証する"
```

`--auto` は goal 条件や `progression_mode=autonomous` の別名ではありません。ミッション文をシェルとして実行せず、条件式としても解析せず、`mission_mode=auto`、`state=draft` の別ファイルへ保存します。作成メッセージに `approval required` と出るとおり、この操作は **ミッション草案の作成** であり、自律実行の承認ではありません。

```bash
moai goal approve --scope <パス> --action publish --action commit --completion-evidence <根拠> --max-operations 20
moai goal run --action publish --target <gtd-id> --recommend
moai goal run --supervise --card-worktree <WT-パス> --develop-worktree <develop-パス> --governor-receipt <判断.json> --audit-receipt <監査.json> --completion-receipt <完了.json>
moai goal status
moai goal revoke
moai goal resume
```

`approve` は目標・範囲・許可行為・完了根拠・資源上限を一度だけ封印します。その後の workflow loop は、範囲内の各操作で最新 snapshot と receipt を再確認し、同じ承認を繰り返し尋ねません。`status` は保存状態を読み、`revoke` は新しい効果を止めつつ進行中の調整状態を残します。`resume` は保存済みの **承認済み・ポリシーで blocked** のミッションだけを同じ契約で再開し、新しい範囲は承認しません。

`--recommend` は旧呼び出しとの互換構文にすぎず、権限を与えません。実操作には、リポジトリ内の `0600` mission-governor 判断 receipt と、別の独立監査 PASS receipt の両方が必要で、ミッション・契約・snapshot・行為・対象・有効期限・発行者・HEAD・true と判定された typed evidence に結び付きます。`run --supervise` は封印済み計画を `publish → pick → lease 付きディスク配車 → commit → local develop --no-ff merge` の順で有限実行します。監督下の Git 効果には分離した `--card-worktree` と `--develop-worktree` が必要で、従来の `--repo` だけでは効果 0 件のまま拒否されます。完了にはマージ ancestry を含む `0600` 完了 receipt が必要で、行為リストの消化だけでは完了しません。blocked または completed で止まり、完了済みミッションの再実行は効果 0 件です。

承認後も、封印された目標・完了根拠・範囲・許可行為・資源上限・停止条件を決定論的なコードが各操作の前に検査します。`mission-governor` は読み取り専用の提案役です。範囲拡大や新しい権限が必要なら、承認を暗黙に広げず、副作用を止めて `blocked` を記録します。

`super-advisor` の意見は非拘束の助言で、読み取り専用の `mission-governor` が構造化された判断を作ります。決定論的 validator と所有役割 adapter だけが効果を実行します。コミットには現在 HEAD のテスト receipt、ローカルマージには manager-git 役割・基準 SHA・lease が必要です。持続 runtime 能力が未確認なら `active-session-only` です。remote batch push・release branch・release PR・main merge の provider は未構成なので、成功を装わず `provider_unsupported` で停止します。GTD の境界は [`/moai gtd`](/ja/utility-commands/moai-gtd) を参照してください。

## 進行モード (自律 / 半自律)

オーケストレーターが実装着手承認 (plan→run 境界の `AskUserQuestion`) を実行するとき、承認/拒否の決定と **区別される別の軸** として **自律 vs 半自律** の進行モードを選択させます。選択したモードは goal 状態の `progression_mode` フィールドに保存されます (ユーザーが選ばなければデフォルト `autonomous`)。

| モード | 動作 |
|------|------|
| **自律 (autonomous, デフォルト)** | 評価器が条件充足または上限到達まで毎ターンブロッキングし、ターンごとにユーザーに尋ねません。既存の Stop フック動作そのままです。 |
| **半自律 (semi-autonomous)** | `stop-goal` フックが毎ターン境界で **チェックポイント信号** ブロック JSON を出力し、オーケストレーターがこれを読んで `AskUserQuestion` 確認ラウンド (続行 / goal 解除 / 自律へ切替) を回します。フック自体は決して `AskUserQuestion` を呼び出しません (フック・サブエージェント境界 — 構造化 JSON のみ放出)。 |

{{< callout type="warning" >}}
**承認は両モードとも必須です。** 進行モードの軸はゲートが通過された **後** に何をするかだけを選択するものであり、ゲートの迂回でもなければ実装着手承認の緩和でもありません。arm された goal はどのモードでも run-phase 進入を承認したり、PR を作ったり、破壊的な作業を行ったりしません。
{{< /callout >}}

## 安全不変式

1. **実装着手承認は両モードとも必須** — 進行モードは承認後の進行選択であってゲート緩和ではなく、スコアと無関係に維持されます。
2. **arm された goal はゲートを迂回しない** — PR を自動生成せず、破壊的な作業を行いません。評価器はターンを続けるかだけを決定し、取り消せない作業を事前承認しません。
3. **`stop-goal` フックは `AskUserQuestion` を呼び出さない** — 構造化 JSON のみ放出します (フック・サブエージェント境界)。
4. **停滞ガード (stagnation guard)** — N 回連続で無進展の反復が検出されるとループを止め、E1/E3 エスカレーションノートを含む 5 セクション判定を出します。

## goal 条件は速くあるべきです

評価器は毎ターン終了時に実行されます。スイート全体より `go test -run <pattern>` を、時間のかかるコマンドより決定論的なコマンドを選んでください。`stop-goal` の Stop フックタイムアウトは 120 秒ですが、速いコマンドがターンループを緻密に保ちます。

## /moai loop との関係

`/moai loop` は **goal エンジンの上のプリセット** です。`/moai goal` がユーザーが完了条件を直接宣言する汎用ループだとすれば、`/moai loop` は「診断ツールが見つけた課題キューを全部空にするまで」という条件をあらかじめ埋めておいたプリセットです。

| エンジン | 目標 | 完了条件 |
|------|------|----------|
| `/moai goal` | 条件宣言型の汎用ループ | ユーザー定義の条件式充足 |
| `/moai loop` | 診断修正ループ (プリセット) | 課題キュー空 + 診断クリーン (0 エラー / テスト通過 / カバレッジ) |

終わりの状態を条件式で表現できるなら `/moai goal`、「ツールが見つける問題を全部なくして」なら `/moai loop` が適しています。

## 関連ドキュメント

- [/moai loop - 反復修正ループ](/ja/utility-commands/moai-loop)
- [/moai fix - 一回限りの自動修正](/ja/utility-commands/moai-fix)
- [/moai - 完全自律の自動化](/ja/utility-commands/moai)
