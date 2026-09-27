---
title: moai contract 自律実行契約
weight: 100
draft: false
---

`moai contract` は SPEC の自律実行契約（`.moai/specs/<SPEC-ID>/contract.yaml`）を検証・表示・署名します。契約には、エージェントが人の確認なしに行ってよい作業（actions）、書き込んでよいパスと決して触れてはならないパス（ownership）、停止して人に引き渡す条件（escalate_on）、予算（budget）を記します。署名はその内容と `acceptance.md` をハッシュで結び付けるため、署名後にどちらかが変わると `verify` がすぐに不一致を報告します。

{{< callout type="info" >}}
現行バージョンでは、署名は**実装着手承認（Implementation Kickoff Approval）の代わりにはなりません。** 契約は記録して検証するための手段であり、人による承認ゲートはそのまま残ります。
{{< /callout >}}

## サブコマンド

| コマンド | 説明 | 終了コード |
|----------|------|------------|
| `moai contract verify <SPEC-ID>` | 契約を検査します。読み取り専用なのでフックから呼んでも安全です | 0 有効 · 1 無効 · 2 使い方/入出力エラー |
| `moai contract show <SPEC-ID>` | 契約のセクション、署名状態、派生集合（実効の書き込み禁止パスなど）を表示します | 0 表示 · 2 使い方/入出力エラー |
| `moai contract sign <SPEC-ID>...` | 契約に署名します | 0 署名 · 1 拒否 · 2 使い方/入出力エラー |

`verify` と `show` は `--json` で機械可読な JSON オブジェクトを出力します。`verify` の無効理由は閉じたコード集合（`unsigned`、`acceptance_hash_mismatch`、`contract_digest_mismatch` など）でのみ報告されます。

## moai contract sign

```bash
moai contract sign SPEC-AUTH-001
moai contract sign SPEC-AUTH-001 --resign
moai contract sign SPEC-AUTH-001 --signer llm \
  --receipt .moai/specs/SPEC-AUTH-001/kickoff-receipt.json
```

| フラグ | 説明 |
|--------|------|
| `--signer <human\|llm\|llm+jev>` | 署名者。既定は人（`human`）の経路です |
| `--receipt <path>` | 着手レシートのパス（プロジェクトルート基準）。`llm`・`llm+jev` の署名に必要です |
| `--resign` | 署名済みの契約を、`acceptance.md` の変更後に署名し直します。以前の署名は `supersedes` として残ります |

**人の経路。** 対話型ターミナルでのみ動作します。署名の要約を表示し、確認トークン（SPEC ID、複数をまとめて署名する場合は `sign N contracts`）を入力したときだけ署名します。エージェント実行を示す環境変数が設定されている場合や、端末でない場合は、プロンプトを出す前に拒否します。複数の SPEC をまとめて署名するには `workflow.autonomy.contract.batch_sign` を有効にする必要があります。

**レシート経路。** `--signer llm` または `--signer llm+jev` と `--receipt` を併せて指定すると、端末での確認なしに着手レシートを根拠として署名します。この経路は `workflow.autonomy.mode` が `contract` のときだけ使え、一度に署名できる SPEC は一つです。

**拒否。** 署名が拒否されるとファイルは一切変更されず、`refused <code> (<SPEC-ID>): <理由>` の一行が出力されます。コードは閉じた集合（`not_tty`、`confirmation_mismatch`、`already_signed`、`plan_audit_not_passing`、`verify_failed` など）です。署名対象のファイルは書き込み前に再検証され、有効な署名と確認できた場合にのみ書き込まれます。作成者のコメントや空行は保持されます。

## 関連ドキュメント

- [config セクションリファレンス — workflow.yaml autonomy](/ja/advanced/config-sections/#workflowyaml--autonomy)
- [moai spec ドキュメント管理](/ja/cli-reference/spec)
- [自律性ティア (MOAI_AUTONOMY_TIER)](/ja/advanced/autonomy-tier) — 名前は似ていますが無関係の設定です
