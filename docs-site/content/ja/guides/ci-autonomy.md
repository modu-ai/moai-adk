---
title: 自律 CI/CD ガイド
weight: 10
draft: false
---

MoAI-ADKの自律CI/CDシステムでプルリクエストの品質を自動的に管理します。

## 概要

SPEC-V3R3-CI-AUTONOMY-001で導入された自律CI/CDシステムは、8つのTierで構成された
品質自動化インフラです。pre-push hookからauto-fixループまで、開発者が手動で
品質を検証する必要なく、CIが自動的に品質を保証します。

## 8-Tierアーキテクチャ

| Tier | 名前 | 優先度 | 説明 |
|------|------|----------|------|
| T1 | Pre-push Hook | P0 | push前の自動品質検証 |
| T2 | Branch Protection | P0 | mainブランチの保護ルール |
| T3 | Auto-fix Loop | P1 | CI失敗時の自動修正 |
| T4 | Auxiliary Workflows | P2 | 補助ワークフローの整理 |
| T5 | Worktree State Guard | P1 | ワークツリー状態の整合性保証 |
| T6 | i18n Validator | P2 | 4言語ドキュメントの一貫性検証 |
| T7 | BODP | P0 | ブランチ起点決定プロトコル |
| T8 | Release Workflow | P1 | リリース自動化 |

## Pre-push Hook (T1)

push前にローカルで自動的に品質検証を実行します。

```bash
# 自動インストール済み (moai init / moai update 実行時)
.git/hooks/pre-push → moai hook pre-push
```

実行される検証:

- `go vet` / `golangci-lint` (プロジェクトの言語に応じて自動検出)
- `go test ./...` (テストスイート)
- MXタグの整合性チェック

## Auto-fix Loop (T3)

CI失敗時に`/moai loop`を自動的に呼び出してエラーを修正します。

```yaml
# .github/workflows/ci.yml (自動生成)
- name: Auto-fix on failure
  if: failure()
  run: |
    claude -p "/moai loop --max-iterations 3"
```

## BODP — Branch Origin Decision Protocol (T7)

新しいブランチ/ワークツリーを作成する際、base branchを自動的に決定します。

### 3-Signal評価

| シグナル | 出所 | 意味 |
|--------|------|------|
| Signal A | SPEC `depends_on` + diff path overlap | コード依存関係 |
| Signal B | `git status`で`.moai/specs/<NewSpecID>/`が一致 | 作業ツリーの同一位置 |
| Signal C | `gh pr list --head <branch> --state open` ≥ 1 | 現在のブランチのPR |

### 決定マトリクス

| シグナル | 決定 |
|--------|------|
| Aのみ存在 | `stacked` — 現在のブランチを基準 |
| Bが存在 | `continue` — 現在のコンテキストで継続 |
| Cのみ存在 | `stacked` — 現在のブランチを基準 |
| 何もない | `main` — origin/mainを基準 |

### 監査証跡

すべてのBODPの決定は`.moai/branches/decisions/<branch-name>.md`に記録されます。

## i18n Validator (T6)

4言語ドキュメントの一貫性を自動検証します。

```bash
scripts/docs-i18n-check.sh
```

検証項目:

- 4locale間のファイル数/パスの一致
- front matter `title`の存在
- H1 headingの存在
- MoAI用語集の遵守

## Worktree State Guard (T5)

ワークツリーの状態整合性を保証します:

- 未コミットの変更を検出
- ワークツリーとmainブランチの同期状態を確認
- `moai status`で状態を表示

## 関連ドキュメント

- [ワークツリーガイド](/ja/worktree/guide) — Git Worktree完全ガイド
- [/moai loop](/ja/utility-commands/moai-loop) — 反復修正ループ
- [/moai fix](/ja/utility-commands/moai-fix) — 自動エラー修正
- [マルチLLM CI](/ja/guides/multi-llm-ci) — Multi-LLM CI統合
