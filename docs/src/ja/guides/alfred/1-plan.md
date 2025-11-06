---
title: /alfred:1-plan コマンド
description: SPEC作成と要件定義のための完全ガイド
lang: ja
---

# /alfred:1-plan - SPEC作成コマンド

`/alfred:1-plan`はMoAI-ADKの計画段階コマンドで、EARS形式のSPEC（仕様書）を作成し、明確な要件を定義します。

## 概要

**目的**: 明確な要件定義と実装計画の作成
**実行時間**: 約2分
**主要成果**: SPEC文書、プランボード、ブランチ

## 基本使用法

```bash
/alfred:1-plan "機能説明"
```

### 例

```bash
# APIエンドポイント作成
/alfred:1-plan "GET /helloエンドポイント - クエリパラメータnameを受け取って挨拶を返す"

# ユーザー認証機能
/alfred:1-plan "ユーザーログイン機能 - メールとパスワードで認証してJWTトークン発行"

# データ管理機能
/alfred:1-plan "TODO CRUD機能 - 作成、照会、修正、削除"
```

## 実行プロセス

### ステップ1: 要件分析

Alfredが以下を自動実行します：

1. **入力分析**: ユーザーの要求を解析
2. **ドメイン検出**: 関連する専門分野を特定
3. **複雑度評価**: 実装難易度を判断
4. **専門家推薦**: 必要な専門家エージェントを推薦

### ステップ2: SPEC作成

#### EARS形式の自動生成

```yaml
---
id: HELLO-001
version: 0.1.0
status: draft
priority: high
created: 2025-01-06
updated: 2025-01-06
author: @user
---

# @SPEC:EX-HELLO-001: Hello World API

## Ubiquitous Requirements
- システムはHTTP GET /helloエンドポイントを提供すべきである

## Event-driven Requirements
- WHEN クエリパラメータnameが提供されたら、"Hello, {name}!"を返すべきである
- WHEN nameがない場合、"Hello, World!"を返すべきである

## State-driven Requirements
- WHILE システムが実行中である時、エンドポイントは応答可能であるべきである

## Optional Features
- WHERE nameパラメータがある場合、パーソナライズされた挨拶を提供できる

## Unwanted Behaviors
- nameが50文字を超える場合、400エラーを返すべきである
- 無効な文字が含まれる場合、400エラーを返すべきである
```

#### プランボード作成

```markdown
## Plan Board

### 実装アイデア
1. FastAPIを使用したREST API実装
2. Pydanticで入力検証
3. エラーハンドリング統合

### 技術スタック
- **フレームワーク**: FastAPI
- **検証**: Pydantic
- **テスト**: pytest
- **ドキュメント**: OpenAPI自動生成

### リスク要因
1. 入力検証の網羅性
2. エラーメッセージの一貫性
3. パフォーマンス要件

### 解決戦略
1. まず基本機能実装
2. 次に入力検証追加
3. 最後にエラーハンドリング強化
```

### ステップ3: 専門家コンサルテーション

#### 自動専門家活性化

特定のキーワードを検出して専門家を自動的に活性化：

| キーワード | 活性化される専門家 | 提供内容 |
|-----------|------------------|----------|
| 'api', 'backend', 'server' | backend-expert | アーキテクチャ設計 |
| 'database', 'storage' | database-expert | データモデリング |
| 'security', 'auth' | security-expert | セキュリティ分析 |
| 'frontend', 'ui' | frontend-expert | UI/UX設計 |
| 'performance', 'scale' | devops-expert | 性能最適化 |

#### 専門家アドバイス例

```
⚙️ backend-expertのアドバイス:
- FastAPIは良い選択です。自動APIドキュメンテーション機能があります
- エンドポイントのバージョニングを検討してください
- レート制限の実装を推奨します

🔒 security-expertのアドバイス:
- 入力検証は重要です。すべてのユーザー入力を検証してください
- エラーメッセージは情報漏洩しないように一般的な内容にしてください
- ログ記録を通じてセキュリティイベントを監視してください
```

## 生成される成果物

### 1. SPEC文書

**場所**: `.moai/specs/SPEC-{ID}/spec.md`
**内容**: EARS形式の要件定義
**TAG**: `@SPEC:EX-{ID}`

### 2. プランボード

**場所**: `.moai/specs/SPEC-{ID}/plan.md`
**内容**: 実装計画、リスク分析、解決戦略

### 3. 受諾基準

**場所**: `.moai/specs/SPEC-{ID}/acceptance.md`
**内容**: 検証基準、テストケース

### 4. Gitブランチ（チームモード）

**名前**: `feature/SPEC-{ID}`
**用途**: 機能開発用分離ブランチ

## 高度な機能

### カスタムテンプレート

特定のパターンでSPECを作成：

```bash
# REST APIテンプレート
/alfred:1-plan "REST API: CRUD operations for user management"

# データベーススキーマ
/alfred:1-plan "Database: User authentication schema with roles"

# UIコンポーネント
/alfred:1-plan "UI Component: Modal dialog with form validation"
```

### 依存関係管理

```bash
# 依存関係のあるSPEC
/alfred:1-plan "User profile management - depends on AUTH-001"

# 生成される依存関係セクション
## Dependencies
- @SPEC:EX-AUTH-001: User Authentication (required)
- @SPEC:EX-USER-002: User Database (optional)
```

### バージョン管理

```yaml
---
id: USER-001
version: 0.2.0  # バージョン自動増加
status: draft    # 状態管理
priority: high   # 優先順位
---

## 変更履歴
- v0.2.0: パスワード検証ルール強化
- v0.1.0: 初期バージョン
```

## 状態管理

### 状態遷移

```
planning → draft → in_progress → testing → completed → deprecated
```

### 状態変更

```bash
# 状態確認
grep "status:" .moai/specs/SPEC-HELLO-001/spec.md

# 手動状態変更（推奨しない）
/alfred:3-sync  # 状態自動同期
```

## 品質基準

### SPEC品質チェックリスト

- ✅ **明確性**: すべての要件が明確で曖昧さがない
- ✅ **完全性**: 必要なすべての機能が含まれている
- ✅ **一貫性**: 用語と構造が一貫している
- ✅ **検証可能性**: 各要件がテスト可能
- ✅ **追跡可能性**: @TAGで追跡可能

### EARSパターン検証

```bash
# EARSパターン確認
grep -E "(WHEN|WHILE|WHERE|システムは.*すべきである)" .moai/specs/SPEC-HELLO-001/spec.md
```

## チーム協業

### SPECレビュープロセス

1. **作成**: `/alfred:1-plan`でSPEC作成
2. **レビュー**: チームメンバーがSPECレビュー
3. **フィードバック**: コメントや改善提案
4. **承認**: 全員が承認後実装開始

### プルリクエスト統合

```bash
# 自動PR作成（チームモード）
/alfred:1-plan "新機能"
→ feature/SPEC-XXXブランチ作成
→ Draft PR自動作成
→ レビュアー自動割り当て
```

## ベストプラクティス

### 1. 良い要件定義

**悪い例**:
```
- ユーザー機能
- ログイン作成
- データ保存
```

**良い例**:
```
- WHEN 有効なメールとパスワードが提供されたら、システムはJWTトークンを発行すべきである
- WHILE ユーザーが認証されている時、保護されたリソースへのアクセスを許可すべきである
- WHERE リフレッシュトークンが有効な場合、新しいアクセストークンを発行できる
```

### 2. 適切な粒度

- **一つのSPEC**: 一つの機能または密接に関連する機能群
- **原子性**: 分割できない最小単位
- **独立性**: 他のSPECに依存しすぎない

### 3. 明確な受け入れ基準

```yaml
## Acceptance Criteria
### 機能要件
- [ ] 有効な認証情報でログインできる
- [ ] 無効な認証情報で401エラーが返る
- [ ] トークン有効期限が15分である

### 非機能要件
- [ ] レスポンスタイムが200ms以内
- [ ] 並列リクエスト1000個処理可能
- [ ] セキュリティ基準準拠
```

## トラブルシューティング

### よくある問題

**SPECが作成されない**:
```bash
# プロジェクト状態確認
moai-adk doctor

# Claude Code再起動
exit && claude

# 依存関係確認
ls .claude/agents/ .claude/skills/
```

**内容が不十分**:
```bash
# より詳細な説明で再実行
/alfred:1-plan "GET /users/{id} - パスパラメータでユーザーIDを受け取り、ユーザー情報をJSON形式で返す。存在しない場合は404エラーを返す"
```

**専門家が活性化されない**:
```bash
# キーワードを明確に含める
/alfred:1-plan "API endpoint for user authentication with JWT tokens and database integration"
```

## 統合と連携

### /alfred:2-runとの連携

```bash
# SPEC作成後すぐ実装
/alfred:1-plan "ユーザー認証機能"
/alfred:2-run AUTH-001  # 作成されたSPEC-IDを使用
```

### /alfred:3-syncとの連携

```bash
# 実装完了後ドキュメント同期
/alfred:3-sync  # SPEC状態をcompletedに更新
```

### GitHub連携

```bash
# 自動Issue作成
/alfred:1-plan "機能"
→ GitHub Issue自動作成
→ PRと連携
→ ラベル自動付与
```

---

**📚 次のステップ**:
- [/alfred:2-run](2-run.md)でTDD実装
- [TDDガイド](../tdd/index.md)でテスト駆動開発
- [SPECガイド](../specs/basics.md)で仕様書作成技術