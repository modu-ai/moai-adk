______________________________________________________________________

## title: /alfred:3-sync コマンド description: ドキュメント同期と品質検証のための完全ガイド lang: ja

# /alfred:3-sync - ドキュメント同期コマンド

`/alfred:3-sync`はMoAI-ADKの同期段階コマンドで、ドキュメントを自動生成し、システム整合性を検証します。

## 概要

**目的**: ドキュメント自動生成とTAGチェーン検証 **実行時間**: 約1-3分 **主要成果**: APIドキュメント、README/CHANGELOG更新、品質レポート

## 基本使用法

```bash
/alfred:3-sync [options]
```

### オプション

```bash
# 基本実行
/alfred:3-sync

# 自動マージ（チームモード）
/alfred:3-sync --auto-merge

# ドキュメントのみ
/alfred:3-sync --target=docs

# TAGのみ検証
/alfred:3-sync --target=tags
```

## 実行プロセス

### フェーズ1: TAGチェーン検証

**tag-agentが実行**:

1. **TAGスキャン**: 全ファイルの@TAGマーカーを検索
2. **チェーン検証**: SPEC→TEST→CODE→DOC連結確認
3. **孤立TAG検出**: 連結されていないTAGを発見
4. **整合性レポート**: TAGチェーン完全性メトリクス

**出力例**:
```
✅ @SPEC:EX-HELLO-001 → .moai/specs/SPEC-HELLO-001/spec.md
✅ @TEST:EX-HELLO-001 → tests/test_hello.py (3テスト関数)
✅ @CODE:EX-HELLO-001:MODEL → src/hello/models.py (2クラス)
✅ @CODE:EX-HELLO-001:API → src/hello/api.py (1エンドポイント)
✅ @DOC:EX-HELLO-001 → docs/api/hello.md (自動生成)

📊 TAGチェーン要約:
- 発見されたTAG総数: 5
- 完全なチェーン: 1/1 (100%)
- 孤立TAG: 0
- 参照欠落: 0
```

### フェーズ2: ドキュメント同期

**doc-syncerが実行**:

1. **APIドキュメント生成**
   - OpenAPI仕様から自動生成
   - エンドポイント、パラメータ、レスポンス
   - 使用例とエラーコード

2. **README更新**
   - API使用法追加
   - インストール手順
   - クイックスタートガイド

3. **CHANGELOG生成**
   - バージョンベースの変更履歴
   - 機能追加、修正、変更事項
   - 破壊的変更の警告

4. **Living Documentation**
   - アーキテクチャ図
   - データフロー図
   - 追跡性マップ

**生成例**:
````markdown
# Hello API Documentation

## GET /hello

パーソナライズされた挨拶メッセージを返します。

### Parameters
- `name` (query, optional): 挨拶する名前 (デフォルト: "World", 最大50文字)

### Response
- **200**: 成功
  ```json
  {"message": "Hello, Alice!"}
  ```
- **400**: 検証エラー

### Examples

```bash
curl "http://localhost:8000/hello?name=Alice"
# → {"message": "Hello, Alice!"}
```

### Traceability
- @SPEC:EX-HELLO-001 - 要件
- @TEST:EX-HELLO-001 - テスト
- @CODE:EX-HELLO-001 - 実装
````

### フェーズ3: 品質ゲート検証

**trust-checkerが実行**:

#### TRUST 5検証

```
✅ Test First: 100%カバレッジ (15/15テスト合格)
✅ Readable: 平均関数長15行 (目標: <50)
✅ Unified: API一貫性パターン準拠
✅ Secured: 入力検証実装済み
✅ Trackable: 全コードに@TAG付与

🎯 総合品質スコア: 95/100
✅ プロダクションデプロイ準備完了
```

#### セキュリティ検証

```
🔒 セキュリティ検証レポート...

### 認証セキュリティ
✅ パスワードハッシュ: bcrypt 12ラウンド
✅ トークン生成: 暗号学的に安全
✅ セッション管理: 適切な有効期限
✅ レート制限: 実装済み、効果的

### データ保護
✅ SQLインジェクション: パラメータ化クエリ
✅ XSS防止: 出力エンコーディング
✅ CSRF保護: SameSite Cookie
✅ HTTPS強制: プロダクションのみ

🛡️ セキュリティステータス: 安全
重大な問題は発見されませんでした
```

#### パフォーマンス検証

```
⚡ パフォーマンス検証レポート...

### レスポンスタイム
✅ ログインエンドポイント: 平均145ms (目標: <500ms)
✅ トークン更新: 平均89ms (目標: <200ms)
✅ ユーザー検証: 平均23ms (目標: <100ms)

### リソース使用
✅ メモリ使用量: 45MB平均 (目標: <100MB)
✅ CPU使用率: 負荷時15%平均
✅ データベース接続: 効率的なプーリング

🚀 パフォーマンスステータス: 最適化済み
全てのパフォーマンス目標を達成
```

### フェーズ4: Gitワークフロー管理

**git-managerが実行**:

#### チームモードブランチ操作

```
🌿 Gitワークフロー管理...

現在のブランチ: feature/SPEC-HELLO-001
ステータス: マージ準備完了

ブランチ検証:
✅ 全テスト合格
✅ ドキュメント同期済み
✅ 品質ゲート通過
✅ マージコンフリクトなし
✅ developブランチと最新

マージオプション:
[1] Draft PR作成 (デフォルト)
[2] developに自動マージ
[3] ブランチで作業継続
[4] リリースブランチ作成

📄 PR情報:
- タイトル: "feat(hello): Hello APIシステム実装"
- 説明: SPEC-HELLO-001から自動生成
- ラベル: feature, api
- レビュアー: コード所有権に基づき自動割り当て
- テスト: 15個合格、100%カバレッジ
- ドキュメント: APIドキュメント更新済み
```

#### コミット履歴最適化

```
📄 コミット履歴分析...

最近のコミット (TDDパターン維持):
a1b2c3d ✅ sync(HELLO-001): ドキュメントと品質チェック更新
d4e5f6c ♻️ refactor(HELLO-001): セキュリティとエラーハンドリング改善
b2c3d4e 🟢 feat(HELLO-001): 認証サービス実装
a3b4c5d 🔴 test(HELLO-001): 失敗する認証テスト追加
e5f6g7h 🌿 feature/SPEC-HELLO-001をdevelopから作成

✅ コミットメッセージ一貫性: 100%
✅ TDDパターン準拠: 100%
✅ コミット内のTAG参照: 100%
✅ サインオフ要件: 満たしている
```

## 生成される成果物

### 1. APIドキュメント
**場所**: `docs/api/{module}.md`
**内容**: エンドポイント、パラメータ、レスポンス、例
**TAG**: `@DOC:EX-{ID}`

### 2. README.md
**更新内容**:
- 新機能説明
- API使用例
- インストール手順

### 3. CHANGELOG.md
**フォーマット**:
```markdown
## [0.2.0] - 2025-01-06

### Added
- ユーザー認証システム (@SPEC:EX-AUTH-001)
  - JWTベース認証 (アクセス・リフレッシュトークン)
  - bcryptを使用した安全なパスワードハッシュ化 (12ラウンド)
  - ブルートフォース攻撃防止のためのレート制限
```

### 4. 品質レポート
**場所**: `.moai/reports/quality-{date}.md`
**内容**: TRUST 5検証結果、セキュリティスキャン、パフォーマンステスト

## 高度な機能

### カスタムドキュメントテンプレート

```yaml
# .moai/templates/api-docs.yml
api_documentation:
  sections:
    - overview
    - authentication
    - endpoints
    - examples
    - security
    - traceability
```

### 多言語ドキュメント

```markdown
# docs/api/auth.ja.md (日本語)
# docs/api/auth.en.md (英語)
# docs/api/auth.ko.md (韓国語)
```

### CI/CD統合

```yaml
# .github/workflows/sync.yml
name: Alfred Sync and Quality Check

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  alfred-sync:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - name: Run Alfred Sync
      run: alfred-sync --ci-mode
```

## トラブルシューティング

### ドキュメントが生成されない

```bash
# ファイル権限確認
ls -la docs/

# 強制再生成
/alfred:3-sync --force --target=docs

# テンプレート確認
cat .moai/templates/api-docs.yml
```

### TAGチェーン破損

```bash
# 破損参照検索
rg '@(SPEC|TEST|CODE|DOC):' -A 2 -B 2

# TAG修正
/alfred:3-sync --fix-tags

# 手動TAG追加
echo "# @TEST:EX-AUTH-001:VALIDATOR" >> tests/test_validators.py
```

### 品質ゲート失敗

```bash
# 詳細品質レポート
/alfred:3-sync --verbose

# 特定問題修正
pytest tests/ --cov=src --cov-report=term-missing

# 再検証
/alfred:3-sync
```

## ベストプラクティス

### 1. 実行前チェックリスト
- ✅ すべてのテスト合格
- ✅ すべての変更コミット済み
- ✅ 適切なTAG付与
- ✅ 依存関係インストール済み

### 2. 同期中
- ⚠️ 警告とエラーを監視
- 📖 生成されたドキュメントをレビュー
- ✅ 品質ゲート通過確認
- 🌿 Gitステータス確認

### 3. 実行後
- 📚 ドキュメントの正確性確認
- 🧪 機能の手動テスト
- 👥 チームへの通知 (該当する場合)
- 📋 次の反復計画

______________________________________________________________________

**📚 次のステップ**:

- [プロジェクト設定](../project/config.md)でカスタマイズ
- [TAGシステム](../../reference/tags/index.md)で追跡性理解
- [品質保証](../quality/index.md)で品質基準学習
