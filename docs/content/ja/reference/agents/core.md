# コアサブエージェント詳細ガイド

Alfredの10人のコアエージェントの完全なリファレンスです。

## 概要

| #   | エージェント           | 役割            | スキル数 | 最適サイズ     |
| --- | ---------------------- | --------------- | ------- | -------------- |
| 1   | project-manager        | プロジェクト初期化 | 5個   | 1人チーム      |
| 2   | spec-builder           | SPEC作成        | 8個     | すべてのチーム |
| 3   | implementation-planner | 計画策定        | 6個     | チームプロジェクト |
| 4   | tdd-implementer        | TDD実行         | 12個    | すべてのチーム |
| 5   | doc-syncer             | ドキュメント同期 | 8個     | すべてのチーム |
| 6   | tag-agent              | TAG検証         | 4個     | 中大型プロジェクト |
| 7   | git-manager            | Git自動化       | 10個    | すべてのチーム |
| 8   | trust-checker          | 品質検証        | 7個     | リリース段階   |
| 9   | quality-gate           | リリース準備    | 6個     | プロダクション |
| 10  | debug-helper           | エラー解決      | 9個     | 問題発生時     |

______________________________________________________________________

## 1. project-manager

**役割**: プロジェクト初期化およびメタデータ管理

### アクティベーション条件

```
/alfred:0-project [setting|update]
```

### 主要な責任

- プロジェクトメタデータ設定（名前、説明、チームサイズ）
- 会話言語の選択と適用
- 開発モードの決定（solo/team/org）
- `.moai/config.json`の初期化
- TRUST 5原則のデフォルト設定

### インタラクション形式

```
User: /alfred:0-project

Alfred: プロジェクト名は？
→ project-manager: 入力値の検証と保存

Alfred: 開発モードは？
→ project-manager: チームサイズに応じた設定決定

Alfred: 会話言語は？
→ project-manager: すべての後続コミュニケーションの言語設定

Result: .moai/config.json作成完了
```

### 生成されるファイル構造

```
.moai/
├── config.json           # プロジェクト設定
├── specs/               # SPEC保存ディレクトリ
├── docs/                # 生成ドキュメント
├── reports/             # 分析レポート
└── scripts/             # ユーティリティ
```

### 使用シナリオ

- **新規プロジェクト開始**: 初回Alfred初期化
- **設定変更**: 言語、チームモード、テストカバレッジ目標の修正
- **複数プロジェクト**: プロジェクトごとの独立設定

______________________________________________________________________

## 2. spec-builder

**役割**: EARS形式のSPECドキュメント作成

### アクティベーション条件

```
/alfred:1-plan "タイトル1" "タイトル2" ...
/alfred:1-plan SPEC-ID "修正事項"
```

### 主要な責任

- ユーザー要件をEARS形式で構造化
- SPEC IDの自動生成（SPEC-001、SPEC-002...）
- 要件の明確性検証
- テスト計画の草案作成
- 実装範囲の定義

### EARS形式構造

```
GIVEN:     初期状況の説明
WHEN:      ユーザーアクション
THEN:      期待される結果
```

### 例

**ユーザー入力**:

```
/alfred:1-plan "ユーザー認証システム"
```

**生成されるSPEC**:

```markdown
# SPEC-001: ユーザー認証システム

## 要件

### ログイン機能
- GIVEN: ユーザーがログインページを訪問
  WHEN: 有効なメールとパスワードを入力
  THEN: セッション作成とダッシュボードリダイレクト

### パスワードエラー処理
- GIVEN: ログインページ
  WHEN: 間違ったパスワードを入力
  THEN: 「パスワードエラー」メッセージ表示

## テスト計画
- [ ] 正常ログイン
- [ ] パスワードエラー
- [ ] アカウントロック（5回失敗）
```

### 品質基準

- 明確な要件（5つ以上）
- 曖昧でない表現
- テスト可能な条件
- 実装可能な範囲

______________________________________________________________________

## 3. implementation-planner

**役割**: アーキテクチャおよび実行計画策定

### アクティベーション条件

```
/alfred:2-run SPEC-ID（開始時）
```

### 主要な責任

- SPECを実装ステップに分解
- ファイルおよびディレクトリ構造の設計
- タスク依存関係の分析
- 並列実行機会の識別
- 予想時間および難易度の推定

### 計画策定プロセス

```
SPEC分析
    ↓
タスク分解（5-10ステップ）
    ↓
依存関係マッピング
    ↓
並列化機会の識別
    ↓
影響ファイルのリスト化
    ↓
時間推定
    ↓
ユーザー承認要求
```

### 計画ドキュメント例

```
SPEC-001: ユーザー認証システム

📋 タスク分解:
1. データモデル設計（User、Session）
2. データベーススキーマ作成
3. パスワードハッシュ関数実装
4. ログインエンドポイント実装
5. セッション管理ミドルウェア作成
6. ログアウトエンドポイント
7. パスワードリセット
8. アカウントロックメカニズム

🔄 依存関係:
1 → 2 → 3 → 4
     ↓
     5 → 6, 7 → 8

⚡ 並列化:
- 4と5は並列可能
- 6, 7, 8は並列可能

📁 影響ファイル:
- models/user.py (NEW)
- models/session.py (NEW)
- api/auth.py (NEW)
- middleware/session.py (NEW)
- tests/test_auth.py (NEW)
- docs/auth.md (NEW)

⏱️ 予想時間: 2時間（3フェーズ: RED/GREEN/REFACTOR）
```

______________________________________________________________________

## 4. tdd-implementer

**役割**: RED-GREEN-REFACTORサイクル実行

### アクティベーション条件

```
/alfred:2-run SPEC-ID（実行中）
```

### 主要な責任

- REDフェーズ: 失敗テストの作成
- GREENフェーズ: 最小実装
- REFACTORフェーズ: コード品質の改善
- 各フェーズ完了後のTodoWrite更新
- テスト状態の追跡

### TDD 3フェーズ実装

#### フェーズ1: RED

```python
# 失敗するテストのみ作成
def test_user_registration():
    user = register_user("user@example.com", "password123")
    assert user.email == "user@example.com"
    assert user.is_verified == False

# 実行 → FAIL :x:
```

#### フェーズ2: GREEN

```python
# 最小実装
def register_user(email, password):
    user = User(email=email)
    user.set_password(password)
    db.session.add(user)
    db.session.commit()
    return user

# 実行 → PASS ✅
```

#### フェーズ3: REFACTOR

```python
# コード品質改善（テストは維持）
def register_user(email, password):
    """ユーザー登録"""
    # 入力検証
    if not is_valid_email(email):
        raise ValueError("Invalid email")
    if len(password) < 8:
        raise ValueError("Password too short")

    # 重複確認
    if User.query.filter_by(email=email).first():
        raise ValueError("User already exists")

    # ユーザー作成
    user = User(email=email)
    user.set_password(password)
    db.session.add(user)
    db.session.commit()

    return user
```

### TodoWrite追跡

```
[in_progress] RED: SPEC-001テスト作成
[completed]   RED: SPEC-001テスト作成
[in_progress] GREEN: SPEC-001最小実装
[completed]   GREEN: SPEC-001最小実装
[in_progress] REFACTOR: SPEC-001コード改善
[completed]   REFACTOR: SPEC-001コード改善
```

______________________________________________________________________

## 5. doc-syncer

**役割**: ドキュメント自動生成および同期

### アクティベーション条件

```
/alfred:3-sync auto [SPEC-ID]
```

### 主要な責任

- APIドキュメント自動生成（OpenAPI/Swagger）
- アーキテクチャダイアグラム生成
- デプロイガイド作成
- 変更サマリードキュメント生成
- ドキュメントリンク検証

### 生成ドキュメント種類

| ドキュメント | 内容                | 形式        |
| ------------ | ------------------- | ----------- |
| API Spec     | RESTfulエンドポイント | OpenAPI 3.1 |
| Architecture | システムダイアグラム | Mermaid     |
| Deployment   | デプロイ手順        | Markdown    |
| Changelog    | 変更内容            | Markdown    |
| Migration    | データマイグレーション | SQL + 説明  |

### 生成場所

```
docs/
├── api/
│   └── SPEC-001.md          # APIドキュメント
├── architecture/
│   └── SPEC-001.md          # アーキテクチャ
├── deployment/
│   └── SPEC-001.md          # デプロイガイド
├── migrations/
│   └── 001_create_users.sql # マイグレーション
└── changelog/
    └── v1.0.0.md            # 変更内容
```

______________________________________________________________________

## 6. tag-agent

**役割**: TAG検証および追跡可能性管理

### アクティベーション条件

```
/alfred:3-sync auto [SPEC-ID]
```

### 主要な責任

- SPEC → TEST → CODE → DOC TAGチェーンの検証
- 孤立TAGの検出と削除
- TAG命名規則の検証
- 追跡可能性の整合性確認

### TAGチェーン

```
SPEC-001（要件）
    ↓
@TEST:SPEC-001:*（テスト）
    ↓
@CODE:SPEC-001:*（実装）
    ↓
@DOC:SPEC-001:*（ドキュメント）
    ↓
相互参照（完全な追跡可能性）
```

### 例

```python
# @CODE:SPEC-001:register_user
def register_user(email: str, password: str) -> User:
    """ユーザー登録"""
    # @CODE:SPEC-001:validate_email
    if not is_valid_email(email):
        raise ValueError("Invalid email")

    # @CODE:SPEC-001:hash_password
    hashed = hash_password(password)

    # @CODE:SPEC-001:create_user
    user = User(email=email, password_hash=hashed)
    db.session.add(user)
    db.session.commit()

    return user

# @TEST:SPEC-001:test_register_success
def test_register_success():
    user = register_user("test@example.com", "password123")
    assert user.email == "test@example.com"
```

______________________________________________________________________

## 7. git-manager

**役割**: Gitワークフロー自動化

### アクティベーション条件

すべての段階で自動アクティベーション

### 主要な責任

- 機能ブランチ作成（feature/SPEC-001）
- コミットメッセージ自動生成
- RED/GREEN/REFACTORフェーズ別コミット
- PR作成および管理
- マージ前検証

### Gitワークフロー

```
main
    ↓
develop（ベースブランチ）
    ↓
feature/SPEC-001（機能ブランチ）
    │
    ├── feat: REDフェーズ（コミット）
    ├── feat: GREENフェーズ（コミット）
    ├── refactor: コード品質（コミット）
    │
    ↓
PR #23（develop ← feature/SPEC-001）
    ├── テスト検証
    ├── コードレビュー
    └── マージ
    ↓
develop（マージ完了）
    ↓
main（リリース時）
```

### コミットメッセージ形式

```
<type>: <description>

🤖 Claude Codeで生成

Co-Authored-By: 🎩 Alfred@MoAI
```

**タイプ**:

- `feat`: 新機能
- `fix`: バグ修正
- `refactor`: コード改善
- `test`: テスト追加
- `docs`: ドキュメント更新

______________________________________________________________________

## 8. trust-checker

**役割**: TRUST 5原則検証

### アクティベーション条件

```
/alfred:2-run SPEC-ID（完了後）
```

### TRUST 5原則

| 原則           | 説明             | 検証           |
| -------------- | ---------------- | -------------- |
| **T**est First | テスト駆動開発   | カバレッジ85%+ |
| **R**eadable   | 読みやすいコード | Linting通過    |
| **U**nified    | 一貫した構造     | 命名規則準拠   |
| **S**ecured    | セキュリティ     | セキュリティスキャン通過 |
| **T**rackable  | 追跡可能性       | TAG整合性      |

### 検証結果

```
✅ Test First: 92%カバレッジ（目標: 85%）
✅ Readable: MyPy完了、ruff通過
✅ Unified: 命名規則準拠
✅ Secured: 依存関係セキュリティチェック通過
✅ Trackable: 12個のTAG検証

:bullseye: TRUST 5準拠: PASS ✅
```

______________________________________________________________________

## 9. quality-gate

**役割**: リリース準備状態確認

### アクティベーション条件

```
/alfred:3-sync auto all（最終段階）
```

### 検証項目

- ✅ すべてのSPEC完了
- ✅ テストカバレッジ85%以上
- ✅ すべてのテスト通過
- ✅ セキュリティ脆弱性0個
- ✅ ドキュメント完成度100%
- ✅ TAG整合性

### リリース決定

```
すべての項目通過 → PRマージ → リリース可能

失敗項目存在 → 詳細レポート → 改善必要
```

______________________________________________________________________

## 10. debug-helper

**役割**: エラー分析および自動修正

### アクティベーション条件

```
エラーまたは例外発生時に自動アクティベーション
```

### 主要な責任

- エラースタックトレース分析
- 根本原因の把握
- 解決方法の提示
- 自動修正可能性の判断
- 一時的な回避策の提示

### エラー処理プロセス

```
エラー発生
    ↓
debug-helper: 分析
    ├─ タイプ把握
    ├─ 原因追跡
    ├─ 類似事例検索
    └─ 解決策提示
    ↓
[自動修正可能？]
    ├─ YES → 修正および再実行
    └─ NO → 詳細ガイド提示
```

______________________________________________________________________

## エージェント間協力事例

### 完全なワークフロー例

```
SPEC-001作成（spec-builder）
    ↓
実装計画（implementation-planner）
    ↓
REDフェーズテスト（tdd-implementer）
    ↓
GREENフェーズ実装（tdd-implementer）
    ↓
REFACTORフェーズ（tdd-implementer）
    ↓
TRUST 5検証（trust-checker）
    ↓
Gitコミット（git-manager）
    ↓
ドキュメント生成（doc-syncer）
    ↓
TAG検証（tag-agent）
    ↓
リリース準備（quality-gate）
    ↓
完了！
```

______________________________________________________________________

**次**: [エキスパートエージェント](experts.md)または[エージェント概要](index.md)



