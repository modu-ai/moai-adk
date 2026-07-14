---
title: 自律 CI/CD ガイド
weight: 10
draft: false
---

MoAI-ADK の自律 CI/CD システムはプルリクエストの品質を自動的に管理します。
ローカルセッションで `/moai loop` がやっていた「診断 → 修正 → 検証」ループを CI まで
延長したもので、開発者が手動で品質を検証しなくても CI が自ら
品質を保証します — エージェンティックループエンジニアリングをリポジトリレベルに適用した
事例です。

## 概要

SPEC-V3R3-CI-AUTONOMY-001 で導入された自律 CI/CD システムは、8 つのティアで構成される
品質自動化インフラです。push 前のローカル検証 (pre-push hook) から CI 失敗時の
自動修正 (auto-fix loop) まで、一つの防衛線としてつながっています。

## 8-Tier アーキテクチャ

| Tier | 名前 | 優先順位 | 説明 |
|------|------|----------|------|
| T1 | Pre-push Hook | P0 | push 前の自動品質検証 |
| T2 | Branch Protection | P0 | main ブランチ保護ルール |
| T3 | Auto-fix Loop | P1 | CI 失敗時の自動修正 |
| T4 | Auxiliary Workflows | P2 | 補助ワークフローの整理 |
| T5 | Worktree State Guard | P1 | ワークツリー状態の整合性保証 |
| T6 | i18n Validator | P2 | 4 言語ドキュメントの一貫性検証 |
| T7 | BODP | P0 | ブランチ原点決定プロトコル |
| T8 | Release Workflow | P1 | リリース自動化 |

## Pre-push Hook (T1)

push 前にローカルで自動的に品質検証を実行します。CI まで行って失敗して
戻ってくる往復コストをローカルで先に断ち切る最初の防衛線です。

```bash
# 自動インストールされる (moai init / moai update 時)
.git/hooks/pre-push → moai hook pre-push
```

実行される検証:

- `go vet` / `golangci-lint` (プロジェクト言語に応じて自動検出)
- `go test ./...` (テストスイート)
- MX タグ整合性検査

## Auto-fix Loop (T3)

`/moai sync` が PR を生成した後、CI 監視スクリプトと CI ループスキルが一緒に
「診断 → 修正 → 再検証」ループを回します。ローカルの診断型自己修正ループを
PR パイプラインの上に延長した構造です。

**CI 監視スクリプト (`scripts/ci-watch/run.sh`)**

```bash
sh scripts/ci-watch/run.sh <PR_NUMBER> [BRANCH]
```

- 30 秒間隔で `gh pr checks` をポーリングし、必須 (required) チェックと
  補助 (auxiliary) チェックを分類します
- 終了コード: `0` 全体通過 · `2` 必須チェック失敗 (構造化 JSON ハンドオフを
  stdout に出力) · `3` 30 分ハードタイムアウト · `1` エラー
- required チェックリストは SSoT ファイルから読み、テスト用の環境変数
  オーバーライド (`MOAI_CIWATCH_GH`、`CIWATCH_TIMEOUT_SECONDS` など) をサポートします

**CI ループスキル (`moai-workflow-ci-loop`)**

監視スクリプトが必須の失敗をハンドオフすると、`moai-workflow-ci-loop` スキルが
失敗を分類し、安全な自動パッチを最大 3 回まで試行します。意味レベルの
失敗 (自動修正が危険な場合) はユーザーにエスカレーションします。

## BODP — Branch Origin Decision Protocol (T7)

新しいブランチ/ワークツリーを生成するとき、base branch を自動的に決定します。

### 3-Signal 評価

| シグナル | 出所 | 意味 |
|--------|------|------|
| Signal A | SPEC `depends_on` + diff path overlap | コード依存性 |
| Signal B | `git status` で `.moai/specs/<NewSpecID>/` マッチ | 作業ツリー同位置 |
| Signal C | `gh pr list --head <branch> --state open` ≥ 1 | 現在ブランチの PR |

### 決定マトリックス

| シグナル | 決定 |
|--------|------|
| A のみ | `stacked` — 現在ブランチベース |
| B あり | `continue` — 現在のコンテキストで継続 |
| C のみ | `stacked` — 現在ブランチベース |
| 何もない | `main` — origin/main ベース |

### 監査トレース

すべての BODP 決定は `.moai/branches/decisions/<branch-name>.md` に記録されます。
決定を推測ではなく記録として残すこと — 証拠ベースの完了判定という MoAI
原則がブランチ決定にも適用されます。

## i18n Validator (T6)

4 言語ドキュメントの一貫性を自動検証します。

```bash
scripts/docs-i18n-check.sh
```

検証項目:

- 4 つの locale 間のファイル数/パス一致
- front matter `title` の存在
- H1 heading の存在
- MoAI 用語集の遵守

## Worktree State Guard (T5)

ワークツリーの状態整合性を保証します:

- コミットされていない変更の検出
- ワークツリーとメインブランチの同期状態確認
- `moai status` での状態表示

## 関連ドキュメント

- [ワークツリーガイド](/ja/worktree/guide) — Git Worktree 完全ガイド
- [/moai loop](/ja/utility-commands/moai-loop) — 反復修正ループ
- [/moai fix](/ja/utility-commands/moai-fix) — 自動エラー修正
- [GitHub 連携ガイド](/ja/guides/multi-llm-ci) — Issue パース・SPEC リンク
