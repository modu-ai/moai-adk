# Alfredワークフローコマンドガイド

MoAI-ADKのコア4ステップワークフローを制御するAlfredコマンドです。

> **重要**: Alfredコマンドは**Claude Code環境内でのみ**使用可能です。

## ワークフロー概要

```
/alfred:0-project（初期化）
    ↓
/alfred:1-plan（計画: SPEC作成）
    ↓
/alfred:2-run（実行: TDD開発）
    ↓
/alfred:3-sync（同期: ドキュメント/検証）
    ↓
完了およびPR作成
```

______________________________________________________________________

## 1. /alfred:0-project

**プロジェクト設定および初期化**

### 構文

```
/alfred:0-project [option]
```

### オプション

```
setting     現在の設定を表示
update      プロジェクト設定を修正
```

### 主要機能

- :bullseye: プロジェクトメタデータ設定（名前、説明、言語）
- 🌐 会話言語選択（韓国語、英語、日本語、中国語）
- 🔧 開発モード選択（solo/team/org）
- 📋 SPEC-First TDDチェックリスト初期化
- 🏷️ TAGシステムアクティベーション
- 📊 テストカバレッジ目標設定（デフォルト85%）

### インタラクション例

```
/alfred:0-project

> Alfred: プロジェクト名を入力してください
User: Hello World API

> Alfred: プロジェクト説明は？
User: シンプルなREST APIチュートリアル

> Alfred: 主に使用する言語は？
User: [1] Python  [2] TypeScript  [3] Go
選択: 1

> Alfred: 会話言語は？
User: [1] 韓国語  [2] 英語
選択: 1

> Alfred: 開発モードは？
User: [1] ソロ  [2] チーム  [3] 組織
選択: 1

✅ プロジェクト初期化完了！
```

### 生成される設定

`.moai/config.json`:

```json
{
  "project": {
    "name": "Hello World API",
    "description": "シンプルなREST APIチュートリアル",
    "language": "python"
  },
  "language": {
    "conversation_language": "ko",
    "conversation_language_name": "韓国語"
  },
  "constitution": {
    "test_coverage_minimum": 85
  }
}
```

______________________________________________________________________

## 2. /alfred:1-plan

**SPEC作成および計画策定**

### 構文

```
/alfred:1-plan "タイトル1" "タイトル2" ... | SPEC-ID 修正事項
```

### 使用ケース

#### 新規SPEC作成

```
/alfred:1-plan "ユーザー認証システム"
```

または複数のSPEC:

```
/alfred:1-plan "ログイン機能" "会員登録" "パスワードリセット"
```

#### 既存SPEC修正

```
/alfred:1-plan SPEC-001 "ログイン機能（修正: OAuth 2.0追加）"
```

### Alfredの計画策定プロセス

1. **意図の把握**: リクエスト分析および明確化

   - 曖昧な場合はAskUserQuestionで追加情報収集

2. **計画策定**: 構造化された実行戦略

   - タスク分解（Decomposition）
   - 依存関係分析（Dependency Analysis）
   - 並列化機会の識別（Parallelization）
   - 影響ファイルの明示（File List）
   - 予想時間の推定（Time Estimation）

3. **ユーザー承認**: 計画提示および承認要求

   ```
   Alfred: 以下のように計画しました。進めますか？

   📋 計画サマリー:
   - SPEC-001: ログイン機能
   - SPEC-002: 会員登録
   - 影響ファイル: 5個
   - 予想時間: 30分

   [進める] [修正] [キャンセル]
   ```

4. **TodoWrite初期化**: すべてのタスク項目の追跡開始

### 生成されるファイル

```
.moai/specs/SPEC-001/
├── spec.md              # SPECドキュメント（EARS形式）
├── requirements.md      # 要件詳細
└── tests.md            # テスト計画
```

### SPECドキュメント構造

```markdown
# SPEC-001: ログイン機能

## 要件（EARS形式）

### 基本要件
- GIVEN: ユーザーがログインページを訪問
  WHEN: 有効なメールとパスワードを入力
  THEN: セッション作成およびダッシュボードリダイレクト

### エラー処理
- GIVEN: ログインページ
  WHEN: 間違ったパスワードを入力
  THEN: 「パスワードが一致しません」メッセージ

## テスト計画
- [ ] 有効な認証情報でログイン成功
- [ ] 無効な認証情報でログイン失敗
- [ ] 新規ユーザー登録後のログイン
```

______________________________________________________________________

## 3. /alfred:2-run

**TDD実装実行**

### 構文

```
/alfred:2-run [SPEC-ID | "all"]
```

### 使用ケース

#### 特定SPEC開発

```
/alfred:2-run SPEC-001
```

#### すべてのSPEC開発

```
/alfred:2-run all
```

### 実行ワークフロー

AlfredはTDDの3フェーズを厳密に従います:

#### フェーズ1: RED - 失敗するテスト作成

```
Alfred: REDフェーズ開始
- テストファイル作成: tests/test_login.py
- テスト作成（SPECベース）
- 実行 → すべて失敗 :x:

✅ REDフェーズ完了
すべてのテストが失敗する状態です。

[GREENフェーズに進む]
```

**サンプルテスト**:

```python
# tests/test_login.py @TEST:SPEC-001:*
import pytest
from app import login

def test_valid_login():
    """GIVEN: ログインページ
       WHEN: 有効な認証情報
       THEN: セッション作成"""
    result = login("user@example.com", "password123")
    assert result["status"] == "success"
    assert result["session"] is not None

def test_invalid_password():
    """GIVEN: ログインページ
       WHEN: 間違ったパスワード
       THEN: エラーメッセージ"""
    result = login("user@example.com", "wrong")
    assert result["status"] == "error"
    assert "パスワード" in result["message"]
```

#### フェーズ2: GREEN - 最小実装でテスト通過

```
Alfred: GREENフェーズ開始
- 最小限の実装追加: app.py
- テスト実行 → すべて通過 ✅

✅ GREENフェーズ完了
すべてのテストが通過します。

[REFACTORフェーズに進む]
```

**サンプル実装**:

```python
# app.py @CODE:SPEC-001:*
def login(email, password):
    """ログイン処理"""
    if password == "password123":
        return {
            "status": "success",
            "session": "session_123"
        }
    else:
        return {
            "status": "error",
            "message": "パスワードが一致しません"
        }
```

#### フェーズ3: REFACTOR - コード品質改善

```
Alfred: REFACTORフェーズ開始
- エラー処理改善
- データ検証追加
- コード整理

✅ REFACTORフェーズ完了
すべてのテスト通過、コード品質改善。

[すべてのタスク完了]
```

**改善された実装**:

```python
# app.py（改善後）
from flask import session
from werkzeug.security import check_password_hash
from models import User

def login(email, password):
    """ログイン処理（改善版）"""
    # 入力検証
    if not email or not password:
        raise ValueError("メールとパスワードは必須です")

    # ユーザー検索
    user = User.query.filter_by(email=email).first()
    if not user:
        return {
            "status": "error",
            "message": "登録されていないユーザーです"
        }

    # パスワード検証
    if not check_password_hash(user.password_hash, password):
        return {
            "status": "error",
            "message": "パスワードが一致しません"
        }

    # セッション作成
    session['user_id'] = user.id
    return {
        "status": "success",
        "session": session.sid,
        "user": {
            "id": user.id,
            "email": user.email,
            "name": user.name
        }
    }
```

### TodoWrite追跡

Alfredは各フェーズを自動的に追跡します:

```
[in_progress] RED: SPEC-001テスト作成
[completed]   RED: SPEC-001テスト作成
[in_progress] GREEN: SPEC-001最小実装
[completed]   GREEN: SPEC-001最小実装
[in_progress] REFACTOR: SPEC-001コード改善
[completed]   REFACTOR: SPEC-001コード改善
```

______________________________________________________________________

## 4. /alfred:3-sync

**ドキュメント同期および検証**

### 構文

```
/alfred:3-sync [Mode] [Target] [Path]
```

### モード

```
auto         自動同期（推奨）
force        強制同期
status       現在の状態のみ表示
project      全体プロジェクト検証
```

### ターゲット

```
SPEC-001         特定SPEC同期
all              すべてのSPEC同期
```

### 主要機能

1. **ドキュメント生成**（生成設定に基づく）

   - APIドキュメント自動生成
   - アーキテクチャダイアグラム
   - デプロイガイド

2. **TAG検証**

   - SPEC → TEST → CODE → DOC接続確認
   - 孤立TAG検出および削除
   - 追跡可能性整合性確認

3. **品質検証**

   - テストカバレッジ85%以上確認
   - すべてのテスト通過確認
   - コードスタイルチェック

4. **PR作成**（チームモード）

   - developブランチ対象PR作成
   - 変更サマリー
   - 検証結果を含める

### 同期プロセス

```
/alfred:3-sync auto SPEC-001

➡️ ステップ1: SPEC検証
✅ SPEC-001構造正常
✅ 要件8個確認

➡️ ステップ2: TAG検証
✅ @TEST:SPEC-001タグ12個
✅ @CODE:SPEC-001タグ15個
✅ @DOC:SPEC-001タグ3個
⚠️ 孤立TAG2個削除されました

➡️ ステップ3: 品質検証
✅ テストカバレッジ: 92%
✅ すべてのテスト通過
✅ コードスタイル正常

➡️ ステップ4: ドキュメント生成
✅ APIドキュメント生成: docs/api/SPEC-001.md
✅ アーキテクチャダイアグラム生成

➡️ ステップ5: PR作成
✅ PR #23作成
📝 タイトル: "feat: SPEC-001ログイン機能実装"
```

______________________________________________________________________

## 5. /alfred:9-feedback

**GitHub Issue作成（フィードバック）**

### 構文

```
/alfred:9-feedback
```

### 機能

- 🐛 バグ報告
- 💡 機能提案
- 📝 改善
- ❓ 質問

### インタラクション例

```
/alfred:9-feedback

> Alfred: フィードバックタイプは？
選択: [1] バグ  [2] 機能  [3] 改善  [4] 質問

> 選択: 1

> タイトルは？
入力: "ログイン後セッション維持されない"

> 説明は？
入力: "ログイン後リフレッシュするとログアウトされる"

> 再現ステップは？
入力: "1. ログイン 2. リフレッシュ 3. ダッシュボードアクセス"

> 期待される動作は？
入力: "セッションが維持されるべき"

✅ GitHub Issue #24作成されました
📝 タイトル: "🐛 ログイン後セッション維持されない"
```

______________________________________________________________________

## コマンドクイックリファレンス

### 完全なワークフロー

```bash
# 1. プロジェクト初期化
/alfred:0-project

# 2. SPEC作成
/alfred:1-plan "ログイン機能" "会員登録"

# 3. TDD実装
/alfred:2-run all

# 4. 同期および検証
/alfred:3-sync auto all

# 完了！PRが自動的に作成されました
```

### 部分ワークフロー

```bash
# 特定SPECのみ修正
/alfred:1-plan SPEC-001 "ログイン機能（OAuth追加）"

# そのSPECのみ開発
/alfred:2-run SPEC-001

# そのSPECのみ同期
/alfred:3-sync auto SPEC-001
```

______________________________________________________________________

## エラー処理

### "Alfredコマンドが認識されない"

```bash
# Claude Code再起動
exit

# 新しいセッション開始
claude

# プロジェクト再初期化
/alfred:0-project
```

### "SPECファイルが見つからない"

```bash
# プロジェクト状態確認
moai-adk status

# 再初期化
moai-adk init . --force
/alfred:0-project
```

### "テストカバレッジ不足"

```bash
# 現在のカバレッジ確認
moai-adk status

# 不足しているテスト追加
# tests/ディレクトリにテスト追加

# 再度同期
/alfred:3-sync auto SPEC-001
```

______________________________________________________________________

**次**: [moai-adkコマンドリファレンス](moai-adk.md)または[Alfred概念](../../guides/alfred/index.md)



