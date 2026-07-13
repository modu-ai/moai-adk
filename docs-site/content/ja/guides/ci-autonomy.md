---
title: 自律 CI/CD ガイド
weight: 10
draft: false
---

MoAI-ADK の自律 CI/CD システムは、プルリクエストの品質を自動で管理します。
ローカルセッションで `/moai loop` が行っていた「診断 → 修正 → 検証」ループを CI まで
延長したもので、開発者が手動で品質を検証しなくても CI が自ら
品質を保証します — エージェンティックループエンジニアリングをリポジトリレベルに適用した
事例です。

## 概要

SPEC-V3R3-CI-AUTONOMY-001 で導入された自律 CI/CD システムは、8 つのティアで構成される
品質自動化インフラです。push 前のローカル検証 (pre-push hook) から CI 失敗時の
自動修正 (auto-fix loop) まで、1 本の防衛線としてつながります。

## 8-Tier アーキテクチャ

| Tier | 名前 | 優先度 | 説明 |
|------|------|----------|------|
| T1 | Pre-push Hook | P0 | push 前の自動品質検証 |
| T2 | Branch Protection | P0 | main ブランチ保護ルール |
| T3 | Auto-fix Loop | P1 | CI 失敗時の自動修正 |
| T4 | Auxiliary Workflows | P2 | 補助ワークフローの整理 |
| T5 | Worktree State Guard | P1 | ワークツリー状態の整合性保証 |
| T6 | i18n Validator | P2 | 4 か国語ドキュメントの一貫性検証 |
| T7 | BODP | P0 | ブランチ起点決定プロトコル |
| T8 | Release Workflow | P1 | リリース自動化 |

## Pre-push Hook (T1)

push 前にローカルで自動的に品質検証を実行します。CI まで行って失敗して
戻ってくる往復コストをローカルで事前に断ち切る、最初の防衛線です。

```bash
# 自動インストールされる (moai init / moai update 時)
.git/hooks/pre-push → moai hook pre-push
```

実行される検証:

- `go vet` / `golangci-lint` (プロジェクトの言語に応じて自動検出)
- `go test ./...` (テストスイート)
- MX タグの整合性チェック

## Auto-fix Loop (T3)

CI 失敗時に `/moai loop` を自動で呼び出してエラーを修正します。ローカルの
診断型自己修正ループが CI ランナー上でそのまま回る構造です。

```yaml
# .github/workflows/ci.yml (自動生成)
- name: Auto-fix on failure
  if: failure()
  run: |
    claude -p "/moai loop --max-iterations 3"
```

## BODP — Branch Origin Decision Protocol (T7)

新しいブランチ/ワークツリーを作成する際、base branch を自動で決定します。

### 3-Signal 評価

| シグナル | 出所 | 意味 |
|--------|------|------|
| Signal A | SPEC `depends_on` + diff path overlap | コード依存性 |
| Signal B | `git status` で `.moai/specs/<NewSpecID>/` にマッチ | 作業ツリーの同位置 |
| Signal C | `gh pr list --head <branch> --state open` ≥ 1 | 現在のブランチの PR |

### 決定マトリクス

| シグナル | 決定 |
|--------|------|
| A のみ | `stacked` — 現在のブランチ基点 |
| B あり | `continue` — 現在のコンテキストで継続 |
| C のみ | `stacked` — 現在のブランチ基点 |
| いずれもなし | `main` — origin/main 基点 |

### 監査証跡

すべての BODP 決定は `.moai/branches/decisions/<branch-name>.md` に記録されます。
決定を推測ではなく記録として残すこと — 証拠に基づく完了判定という MoAI の
原則がブランチ決定にも適用されます。

## i18n Validator (T6)

4 か国語ドキュメントの一貫性を自動検証します。

```bash
scripts/docs-i18n-check.sh
```

検証項目:

- 4 つの locale 間のファイル数/パスの一致
- front matter `title` の存在
- H1 heading の存在
- MoAI 用語集の遵守

## Worktree State Guard (T5)

ワークツリーの状態の整合性を保証します:

- コミットされていない変更の検出
- ワークツリーと main ブランチの同期状態の確認
- `moai status` での状態表示

## 関連ドキュメント

- [ワークツリーガイド](/ja/worktree/guide) — Git Worktree 完全ガイド
- [/moai loop](/ja/utility-commands/moai-loop) — 反復修正ループ
- [/moai fix](/ja/utility-commands/moai-fix) — 自動エラー修正
- [マルチ LLM CI](/ja/guides/multi-llm-ci) — Multi-LLM CI 統合
