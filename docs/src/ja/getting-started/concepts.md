---
title: 核心概念
description: MoAI-ADKの5つの核心概念と基本原則の理解
lang: ja
---

# 核心概念の理解

MoAI-ADKは5つの核心概念で構成されています。各概念は相互に連結し、共に機能する時、強力な開発システムを作ります。

---

## 概念1: SPEC-First (要件優先)

### 比喩：建築家なしで建物を建てるようなもの

設計図なしでコーディングしてはいけません。

### 核心

実装する前に**「何を作るか」を明確に定義**します。これは単なるドキュメントではなく、チームとAIが共同で理解できる**実行可能な仕様**です。

### EARS文法の5つのパターン

1. **Ubiquitous** (基本機能): 「システムはJWTベース認証を提供すべきである」
2. **Event-driven** (条件付き): 「**WHEN** 有効な認証情報が提供されたら、システムはトークンを発行すべきである」
3. **State-driven** (状態中心): 「**WHILE** ユーザーが認証された状態である時、システムは保護されたリソースを許可すべきである」
4. **Optional** (選択): 「**WHERE** リフレッシュトークンがある場合、システムは新しいトークンを発行できる」
5. **Constraints** (制約): 「トークン有効期限は15分を超えてはならない」

### どのように？

`/alfred:1-plan`コマンドがEARS形式で専門的なSPECを自動作成します。

### 得られるもの

- ✅ チーム全員が理解する明確な要件
- ✅ SPECベースのテストケース（何をテストするか既に定義済み）
- ✅ 要件変更時`@SPEC:ID` TAGで影響を受けるすべてのコード追跡可能

### 実践例

```yaml
---
id: AUTH-001
version: 0.1.0
status: draft
priority: high
---

# @SPEC:EX-AUTH-001: ユーザー認証

## Ubiquitous Requirements
- システムはJWTベース認証を提供すべきである

## Event-driven Requirements
- WHEN 有効な認証情報が提供されたら、システムはトークンを発行すべきである
- WHEN 無効な認証情報が提供されたら、システムは401エラーを返すべきである

## Constraints
- トークン有効期限は15分を超えてはならない
- パスワードは最低8文字でなければならない
```

---

## 概念2: TDD (テスト駆動開発)

### 比喩：目的地を決めてから道を探すようなもの

テストで目標を定めてコードを書きます。

### 核心

**実装**前に**テスト**を先に書きます。料理前に材料を確認するように、実装前に要件が何か明確にします。

### 3ステップサイクル

#### 🔴 RED: 失敗するテストを先に書く

- SPECの各要件がテストケースになる
- まだ実装がないので必ず失敗
- Gitコミット: `test(AUTH-001): add failing test`

```python
def test_login_with_valid_credentials_should_return_token():
    """WHEN 有効な認証情報が提供されたら、システムはトークンを発行すべきである"""
    response = client.post("/auth/login", json={
        "email": "user@example.com",
        "password": "valid_password"
    })
    assert response.status_code == 200
    assert "token" in response.json()
    assert response.json()["token"] is not None
```

#### 🟢 GREEN: テストを通過させる最小実装

- 最も単純な方法でテスト通過
- 完璧さより通過が先
- Gitコミット: `feat(AUTH-001): implement minimal solution`

```python
@app.post("/auth/login")
def login(credentials: LoginRequest):
    """@CODE:EX-AUTH-001:LOGIN - ログインエンドポイント"""
    # 最小実装 - すべてのリクエストにトークンを返す
    return {"token": "fake_token_for_testing"}
```

#### ♻️ REFACTOR: コードを整理・改善

- TRUST 5原則適用
- 重複排除、可読性向上
- テストは依然として通過すべき
- Gitコミット: `refactor(AUTH-001): improve code quality`

```python
@app.post("/auth/login")
def login(credentials: LoginRequest):
    """@CODE:EX-AUTH-001:LOGIN - 改善されたログインエンドポイント"""
    user = authenticate_user(credentials.email, credentials.password)
    if not user:
        raise HTTPException(status_code=401, detail="Invalid credentials")

    token = create_access_token(data={"sub": user.email})
    return {"token": token}
```

### どのように？

`/alfred:2-run`コマンドがこの3ステップを自動実行します。

### 得られるもの

- ✅ カバレッジ85%以上保証（テストなしのコードなし）
- ✅ リファクタリング自信（いつでもテストで検証可能）
- ✅ 明確なGit履歴（RED → GREEN → REFACTOR過程追跡）

---

## 概念3: @TAGシステム

### 比喩：宅配便の送り状のようなもの

コードの旅を追跡できる必要があります。

### 核心

すべてのSPEC、テスト、コード、ドキュメントに`@TAG:ID`を付けて**一対一対応**を作ります。

### TAGチェーン

```
@SPEC:EX-AUTH-001 (要件)
    ↓
@TEST:EX-AUTH-001 (テスト)
    ↓
@CODE:EX-AUTH-001 (実装)
    ↓
@DOC:EX-AUTH-001 (ドキュメント)
```

### TAG IDルール

`<ドメイン>-<3桁数字>`

- AUTH-001, AUTH-002, AUTH-003...
- USER-001, USER-002...
- 一度割り当てられたら**絶対に変更しません**

### どのように使用？

要件が変更されたら：

```bash
# AUTH-001と関連するすべてを検索
rg '@TAG:AUTH-001' -n

# 結果: SPEC, TEST, CODE, DOCがすべて一度に表示
# → どこを修正すべきか明確
```

### どのように？

`/alfred:3-sync`コマンドがTAGチェーンを検証し、orphan TAG（対応されないTAG）を検出します。

### 得られるもの

- ✅ すべてのコードの意図が明確（SPECを読めばなぜこのコードがあるか理解）
- ✅ リファクタリング時影響を受けるすべてのコードを即座把握
- ✅ 3ヶ月後でもコード理解可能（TAG → SPEC追跡）

---

## 概念4: TRUST 5原則

### 比喩：健康な体のようなもの

良いコードは5つの要素をすべて満たす必要があります。

### 核心

すべてのコードは以下の5つの原則を必ず守る必要があります。`/alfred:3-sync`がこれを自動検証します。

#### 1. 🧪 Test First (テスト優先)

- テストカバレッジ ≥ 85%
- すべてのコードがテストで保護される
- 機能追加 = テスト追加

#### 2. 📖 Readable (読みやすいコード)

- 関数 ≤ 50行、ファイル ≤ 300行
- 変数名が意図を表す
- リンター(ESLint/ruff/clippy)通過

#### 3. 🎯 Unified (一貫した構造)

- SPECベースアーキテクチャ維持
- 同じパターンが繰り返される（学習曲線減少）
- タイプ安全性またはランタイム検証

#### 4. 🔒 Secured (セキュリティ)

- 入力検証（XSS, SQLインジェクション防御）
- パスワードハッシュ（bcrypt, Argon2）
- 機密情報保護（環境変数）

#### 5. 🔗 Trackable (追跡可能)

- @TAGシステム使用
- GitコミットにTAG含む
- すべての意思決定が文書化される

### どのように？

`/alfred:3-sync`コマンドがTRUST検証を自動実行します。

### 得られるもの

- ✅ プロダクション品質のコード保証
- ✅ チーム全体が同じ基準で開発
- ✅ バグ減少、セキュリティ脆弱性事前防止

---

## 概念5: Alfred SuperAgent

### 比喩：個人秘書のようなもの

Alfredがすべての複雑な作業を処理します。

### 核心

AIエージェントたちが協力して開発過程全体を自動化します。

### エージェント構成

- **Alfred SuperAgent**: 全体オーケストレーション
- **Core Sub-agent**: SPEC作成、TDD実装、ドキュメント同期など専門業務
- **Domain Specialist**: バックエンド、フロントエンド、セキュリティなど
- **Built-in Agent**: 一般質問、コードベース探索

### Claude Skills

- **Foundation**: TRUST/TAG/SPEC/Git/EARS原則
- **Essentials**: デバッグ、性能、リファクタリング、コードレビュー
- **Alfred**: ワークフロー自動化
- **Domain**: バックエンド、フロントエンド、セキュリティなど
- **Language**: Python, JavaScript, Go, Rustなど
- **Ops**: Claude Codeセッション管理

### どのように？

`/alfred:*`コマンドが必要な専門家チームを自動活性化します。

### 得られるもの

- ✅ プロンプト作成不要（標準化されたコマンド使用）
- ✅ プロジェクトコンテキスト自動記憶（同じ質問繰り返さず）
- ✅ 最適の専門家チーム自動構成（状況に合ったサブエージェント活性化）

---

## 🔄 5つの概念の連携

これら5つの概念は相互に連携して完全な開発システムを形成します：

```mermaid
%%{init: {'theme':'neutral'}}%%
graph TD
    SPEC[SPEC-First<br/>明確な要件] --> TDD[TDD<br/>テスト駆動実装]
    TDD --> TAG[@TAGシステム<br/>追跡可能性]
    TAG --> TRUST[TRUST 5原則<br/>品質保証]
    TRUST --> Alfred[Alfred SuperAgent<br/>自動化]
    Alfred --> SPEC
```

### 実際のワークフロー

1. **SPEC作成** → 明確な要件定義
2. **TDD実行** → テスト駆動で高品質なコード作成
3. **@TAG付与** → すべての要素に追跡タグ付与
4. **TRUST検証** → 品質基準満たすか確認
5. **Alfred自動化** → 全過程をAIが支援

---

## 🎯 学習パス

### 初心者向け

1. **SPEC-First**から始める - 明確な要件がすべての基本
2. **TDD**体験 - RED → GREEN → REFACTORサイクルを実際に体験
3. **@TAG**活用 - 追跡可能性の価値を理解

### 中級者向け

1. **TRUST 5原則**適用 - 品質基準をコードに適用
2. **Alfred**活用 - 効率的なコマンド使用法習得
3. **概念連携**理解 - 5つの概念がどのように連携するか理解

### 上級者向け

1. **カスタマイズ** - プロジェクトに合わせた設定
2. **拡張** - 新しいスキルやエージェント追加
3. **最適化** - チームワークフロー最適化

---

## 💡 実践的ヒント

### 日常開発での適用

- **朝**: `/alfred:1-plan`で今日の機能SPEC作成
- **昼**: `/alfred:2-run`でTDD実装
- **夕**: `/alfred:3-sync`でドキュメント同期
- **定期**: `moai-adk doctor`でシステム健康診断

### チーム協業

- **SPECレビュー**: チーム全員でSPEC確認
- **TAG標準**: チーム内TAG命名規則統一
- **TRUST基準**: 品質基準チーム共通認識
- **Alfred活用**: チームメンバー全員がAlfredコマンド習得

---

**🎓 これで5つの核心概念を理解しました！** 次は[Alfredコマンドガイド](../guides/alfred/index.md)で実際の使用方法を学び、[TDDガイド](../guides/tdd/index.md)で実践的なテクニックを習得しましょう。