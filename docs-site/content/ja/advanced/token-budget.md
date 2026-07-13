---
title: トークン予算管理と graceful stop
weight: 2
draft: false
---

トークノミクス4層構造のD層である予算防御(Budget defense)を深掘りします。エージェントがコンテキストウィンドウの限界に達したとき、セッションが停止せず、進行状況を保全しながら次のセッションが引き継げるようにするgraceful abortメカニズムを説明します。

## 予算防御の必要性

Anthropic SSEストリームはコンテキストウィンドウの天井に近づくと`stream_idle_partial`状態で間欠的なストールを起こします。これは確率的ですが、しきい値以上で予測可能です。ストールが発生するとエージェント呼び出しがストリーム途中で失敗し、進行状況を失う可能性があります。

予算防御はこの問題を事前に解決します。コンテキスト使用量がしきい値に達する前にシステムがgraceful abortを実行し、セッションが損失なく次のステップへ移行するように保証します。

## モデルごとのコンテキストしきい値

運用しきい値はモデルごとに異なります。大きなウィンドウはより高いパーセンテージ利用を許容し、小さなウィンドウはパーセンテージでは後になりますが絶対ヘッドルームが少ないです。

| モデルクラス | ウィンドウ | ハンドオフしきい値 | 絶対天井 |
|-------------|-----------|-------------------|---------|
| Opus 4.8 (1M) | 1,000,000 トークン | 50% | ~500,000 トークン |
| GLM-5.2 (1M) | 1,000,000 トークン | 50% | ~500,000 トークン |
| Opus / Fable (256K) | 256,000 トークン | 90% | ~230,000 トークン |
| Sonnet / Opus 標準 (200K) | 200,000 トークン | 90% | ~180,000 トークン |
| Haiku (200K) | 200,000 トークン | 90% | ~180,000 トークン |

GLM-5.2(`moai glm` / `moai cg` GLMパネル)は1Mコンテキストモデルなので50%しきい値で運用します。Claude Codeが報告する`context_window_size`はClaudeスロット基準(Opus=1M, Sonnet/Haiku=200K)なので、GLMセッションで生のtelemetryが~180Kを示してもMoAIが1Mに補正します。statuslineのCW%ゲージを信頼してください。

## 2段階ハンドオフマーカー

statuslineはコンテキストバーに`/clear`ヒントを2段階で表示します。

- {{< icon warning warn >}} **ソフトマーカー** — statuslineのコンテキストバーに`/clear`ヒントが警告色とともに表示されます。バンドのソフトしきい値で現れ、ユーザーが判断して`/clear`を実行できる勧告シグナルです。
- {{< icon warning danger >}} **ハードマーカー** — statuslineのコンテキストバーに`/clear`ヒントが強い警告色とともに表示されます。auto-compact認識天井で現れ、次のアクションは必ず`/clear`でなければなりません。

ハード天井はauto-compactしきい値近くに設定されるため、ランタイムのauto-compactが先に発動してハードマーカーは実際には稀にしかトリガーされません。これはauto-compact認識公式の意図されたトレードオフです。

## graceful abort手順

SPEC-TOKEN-BUDGET-STOP-001で実装されたgraceful abortメカニズムは次の手順で動作します。

1. **検出** — `Tracker.IsAtHardLimit(agentName)`がtrueを返す (累積使用量 ≥ hard_clear_threshold、デフォルト0.90)
2. **状態保存** — 進行中の作業状態を`progress.md`に永続化
3. **ハンドオフ発行** — ペースト可能なresumeメッセージを生成 (6ブロック構造)
4. **ターン終了勧告** — ユーザーに`/clear`を勧告 (HARD: 自動`/clear`は決して実行しない)
5. **証拠永続化** — 検証証拠を`.moai/state/verify/`配下に永続化

`/clear`は決して自動的に実行されません。システムはユーザーが`/clear`を実行するよう勧告のみを行い、ユーザーが判断して実行します。

## paste-ready resume 6ブロック構造

セッションハンドオフメッセージは次の6ブロック構造に従います。各ブロックは次のセッションが最小限の情報で作業を続けられるように設計されています。

```text
✂──── ここからコピー ────✂

ultrathink. <SPEC-ID> <phase> entering.
applied lessons: <memory-file-1>, <memory-file-2>

Preconditions:
1) <検証可能な前提 1>
2) <検証可能な前提 2>

Run: <コマンドまたはアクション>

After merge: <次のアクションまたはSPEC>

✂──── ここまでコピー ────✂
```

各ブロックの役割:

- **ブロック 1** — `ultrathink.`オープナーがeffort:xhighを設定し、進入するphaseとSPEC-IDを宣言
- **ブロック 2** — `applied lessons:` 前のセッションで学習したmemoryファイル参照 (最大4件)
- **ブロック 3** — `Preconditions:` 次のセッションが開始前に確認すべき検証可能な前提 (最大4件、各200字以内)
- **ブロック 4** — 個別の前提項目
- **ブロック 5** — `Run:` 単一の主要アクション (典型的には`/moai <subcommand>`)
- **ブロック 6** — `After merge:` 次のアクションまたはSPEC ID

## 検証ダイエット (verify-diet)

検証コマンドの長文出力をディスクにリダイレクトしコンテキストには要約のみ残すファイルリダイレクト契約(file-redirect contract)です。

ルール: 検証コマンドのverbatim出力が**bounded-tail ceiling**(デフォルト50行または2KBのいずれか小さい方)を超えると、出力をファイルにリダイレクトしコンテキストにはexit code + bounded tailのみ表示します。

```bash
go test ./... > /tmp/moai-verify/1-go-test.log 2>&1; echo "exit=$?"; tail -50 /tmp/moai-verify/1-go-test.log
```

この契約は「ディスクにverbatim証拠を残し、コンテキストにはexit code + bounded tailのみ運搬」します。証拠を捨てるのではなく、インライン出力 + バーナー再引用の二重消費(double-burn)を削除します。

## 検証証拠永続化の義務

ファイルリダイレクト契約が`/tmp`に書き込んだ証拠はOSによって定期的に削除されます(macOSリブート、Linux tmpfsリマウント、systemd-tmpfiles)。引用されたパスがもはやファイルとして存在しない場合、監査時に証拠に到達できません。

永続化の義務はこの問題を解決します。検証証拠は`.moai/state/verify/<session>/`配下に永続化されなければなりません。このディレクトリは`context-usage.json`や`active-sessions.json`と同じgitignoredランタイム状態領域です。

正確な永続化メカニズム(直接書き込みまたは`/tmp`書き込み後コピー)は実装の詳細です。契約は義務を述べます: 証拠は`/tmp`クリア後も監査時に到達可能な引用可能なパスに残なければなりません。

## 次のステップ

- [トークノミクス概論](/ja/advanced/tokenomics-overview/) — 4層構造の全体概観
- [3層エージェントアーキテクチャ](/ja/advanced/no-haiku-3tier/) — Layer Bルーティングのモデルポリシー基礎
