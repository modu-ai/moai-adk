---
title: 貢献する
weight: 110
draft: false
---

MoAI-ADK はオープンソースプロジェクトであり、コントリビューションを歓迎します! MoAI-ADK 自体も SPEC
ベースの 3-phase ワークフローと TRUST 5 品質ゲートで開発されています — 貢献手順の
品質基準 (カバレッジ、リント、Conventional Commits) はその基準をそのまま踏襲します。

{{< mascot coding >}}

## クイックスタート

1. リポジトリを **Fork** します
2. 機能ブランチの作成: `git checkout -b feature/my-feature`
3. テストの作成 (新規コードは TDD、既存コードは特性化テスト)
4. すべてのテストの通過を確認: `make test`
5. リントの通過を確認: `make lint`
6. コードフォーマット: `make fmt`
7. Conventional Commit メッセージでコミット
8. Pull Request の作成

## コード品質要件

TRUST 5 フレームワークの **T**ested / **T**rackable 基準がそのまま適用されます:

| 項目 | 基準 |
|------|------|
| テストカバレッジ | **85%** 以上 |
| リントエラー | **0** 件 |
| 型エラー | **0** 件 |
| コミットメッセージ | Conventional Commits 形式 |

## コミットメッセージ形式

```
<type>(<scope>): <description>

[任意の本文]

[任意のフッター]
```

### タイプ

| タイプ | 説明 |
|------|------|
| `feat` | 新機能 |
| `fix` | バグ修正 |
| `docs` | ドキュメント変更 |
| `style` | コードフォーマット (機能変更なし) |
| `refactor` | リファクタリング (機能変更なし) |
| `perf` | パフォーマンス改善 |
| `test` | テストの追加/修正 |
| `chore` | ビルド/ツールの変更 |
| `revert` | 以前のコミットの取り消し |

### 例

```
feat(template): add SessionEnd hook to settings.json generator
fix(cli): prevent race condition in hook execution
test(settings): add TestEnsureGlobalSettingsEnv test cases
docs(readme): update agent count and statistics
```

## 開発環境の設定

### 必須ツール

- **Go 1.26+** — コア開発言語
- **Git** — バージョン管理
- **make** — ビルドコマンド

### 主なコマンド

```bash
make build        # プロジェクトのビルド
make test         # テストの実行
make test-race    # Race condition 検出テスト
make lint         # リンターの実行
make fmt          # コードフォーマット
make install      # ローカルインストール
make clean        # ビルド成果物のクリーンアップ
```

## Pull Request ガイド

### PR 作成時

- 明確で簡潔なタイトル (70 文字以内)
- 変更内容の要約 (Summary セクション)
- テスト計画 (Test Plan セクション)
- 関連 Issue の参照 (例: `Fixes #123`)

### PR チェックリスト

- [ ] テストの追加/更新
- [ ] すべてのテストの通過 (`make test`)
- [ ] リントの通過 (`make lint`)
- [ ] コミットメッセージが Conventional Commits 形式
- [ ] ドキュメントの更新 (必要な場合)

## コミュニティ

- **Issue トラッカー**: [GitHub Issues](https://github.com/modu-ai/moai-adk/issues) — バグレポート、機能リクエスト。MoAI-ADK を使用中なら `/moai feedback` でセッション内から直接 Issue を作成できます
- **Discord**: [Discord コミュニティ](https://discord.gg/Z7E7Mdc5aN) — リアルタイムの交流、Tips の共有
- **公式ドキュメント**: [adk.mo.ai.kr](https://adk.mo.ai.kr)

## ライセンス

[Apache License 2.0](https://github.com/modu-ai/moai-adk/blob/main/LICENSE) — 自由に使用、修正、配布できます。
