# TDD GREENフェーズガイド：最小実装でテストを通過させる

## 目次

1. [GREENフェーズの目標と原則](#greenフェーズの目標と原則)
2. [最小実装戦略（YAGNI原則）](#最小実装戦略yagni原則)
3. [テスト通過のための迅速な解決策](#テスト通過のための迅速な解決策)
4. [パフォーマンスvs機能実装のバランス](#パフォーマンスvs機能実装のバランス)
5. [実践コード例](#実践コード例)
6. [GREENフェーズで避けるべきこと](#greenフェーズで避けるべきこと)
7. [Gitコミット戦略（GREENフェーズ）](#gitコミット戦略greenフェーズ)
8. [GREENフェーズチェックリスト](#greenフェーズチェックリスト)

______________________________________________________________________

## GREENフェーズの目標と原則

### GREENフェーズの核心目標

GREENフェーズの唯一で明確な目標は、**「失敗するすべてのテストを通過させること」**です。このフェーズでは次のことを覚えておく必要があります：

\`\`\`mermaid
graph TD
    A[REDフェーズ<br/>失敗するテスト] --> B[GREENフェーズ<br/>最小実装]
    B --> C[すべてのテスト通過]
    C --> D[REFACTORフェーズ<br/>コード改善]

    style A fill:#ffeb3b
    style B fill:#4caf50
    style C fill:#81c784
    style D fill:#9c27b0
\`\`\`

### 1. 通過が最優先（Passing First）

GREENフェーズの哲学はシンプルです：

- **完璧さより通過**：美しいコードより通過するコードが先
- **シンプルさが美徳**：最もシンプルな解決策を見つける
- **迅速なフィードバック**：テストが早く通過すれば次のステップに進める

### 2. YAGNI原則（You Aren't Gonna Need It）

> "今必要でない機能は実装するな"

\`\`\`python
# 悪い例：過剰エンジニアリング
class UserService:
    def __init__(self):
        self.cache = RedisCache()           # まだ不要
        self.logger = StructuredLogger()    # まだ不要
        self.metrics = PrometheusMetrics()  # まだ不要
        self.validator = ComplexValidator() # まだ不要

    def create_user(self, user_data):
        # 複雑なロジック...
        pass

# 良い例：最小実装
class UserService:
    def create_user(self, user_data):
        # テスト通過に必要な最小限のロジックのみ
        if not user_data.get("email"):
            raise ValueError("Email is required")

        return User(
            email=user_data["email"],
            name=user_data.get("name", "")
        )
\`\`\`

### 3. KISS原則（Keep It Simple, Stupid）

シンプルな解決策が常に最善です：

- **複雑さの回避**：複雑なアルゴリズム、デザインパターンを避ける
- **直感的なコード**：他の開発者が理解しやすいコード
- **最小依存性**：不要な外部ライブラリやサービスを避ける

______________________________________________________________________

## 最小実装戦略（YAGNI原則）

### 1. ハードコーディング戦略

時にはハードコーディングが最良の選択です：

\`\`\`python
# REDフェーズテスト
def test_get_current_temperature_should_return_25():
    """現在の温度を照会すると25度を返すべき"""
    response = temperature_service.get_current_temperature()
    assert response == 25

# GREENフェーズ最小実装
class TemperatureService:
    def get_current_temperature(self):
        # 実際のセンサーの代わりにハードコードされた値を返す
        return 25  # テスト通過のための最小実装
\`\`\`

**ハードコーディングが適切な場合：**

- テストが単一の固定値を期待する時
- 外部依存性（センサー、API、データベース）が複雑な時
- 最初の通過を素早く得たい時

### 2. フェイク実装（Fake Implementation）

シンプルなメモリベースの実装から開始：

\`\`\`python
# REDフェーズテスト
def test_user_creation_should_return_user_with_id():
    """ユーザー作成時にIDが割り当てられたユーザーオブジェクトを返すべき"""
    user_data = {"name": "John", "email": "john@example.com"}
    user = user_service.create_user(user_data)

    assert user.id is not None
    assert user.name == "John"
    assert user.email == "john@example.com"

# GREENフェーズ最小実装
class UserService:
    def __init__(self):
        self._users = {}  # シンプルなメモリストレージ
        self._next_id = 1

    def create_user(self, user_data):
        # 最小限の検証ロジック
        if not user_data.get("email"):
            raise ValueError("Email is required")

        # 最もシンプルなID生成
        user_id = f"user_{self._next_id}"
        self._next_id += 1

        # 最小限のユーザーオブジェクト作成
        user = User(
            id=user_id,
            email=user_data["email"],
            name=user_data.get("name", "")
        )

        self._users[user_id] = user
        return user
\`\`\`

### 3. 条件付き最小実装

必要な条件のみ実装：

\`\`\`python
# REDフェーズテスト
def test_admin_can_access_admin_panel():
    """管理者は管理パネルにアクセスできるべき"""
    admin = User(role="admin")
    assert auth_service.can_access_admin_panel(admin) is True

def test_regular_user_cannot_access_admin_panel():
    """一般ユーザーは管理パネルにアクセスできないべき"""
    user = User(role="user")
    assert auth_service.can_access_admin_panel(user) is False

# GREENフェーズ最小実装
class AuthService:
    def can_access_admin_panel(self, user):
        # テストに必要な最小限の条件のみ実装
        return user.role == "admin"
\`\`\`

### 4. 返り値固定戦略

\`\`\`python
# REDフェーズテスト
def test_calculate_tax_should_return_10_percent():
    """所得税計算時に10%を返すべき"""
    tax = tax_calculator.calculate_tax(1000)
    assert tax == 100

# GREENフェーズ最小実装
class TaxCalculator:
    def calculate_tax(self, income):
        # すべての所得に対して10%固定（テストに必要な最小実装）
        return income * 0.10
\`\`\`

______________________________________________________________________

## テスト通過のための迅速な解決策

### 1. 段階的アプローチ

複雑なテストを小さく分けて解決：

\`\`\`python
# 複雑なテスト
def test_user_registration_complete_flow():
    """完全なユーザー登録フロー テスト"""
    # 1. 有効なデータで会員登録
    # 2. メール認証トークン送信確認
    # 3. トークンでメール認証
    # 4. 認証されたユーザーログイン
    # 5. JWTトークン受信確認

# GREENフェーズ：一つずつ実装
class UserService:
    def register_user(self, user_data):
        # ステップ1：最小限のユーザー作成のみ実装
        if not user_data.get("email"):
            raise ValueError("Email required")

        user = User(
            id=self._generate_id(),
            email=user_data["email"],
            is_verified=False  # まだ認証ロジック未実装
        )

        return user

    def send_verification_email(self, user):
        # ステップ2：フェイクメール送信
        return True  # 常に成功を返す

    def verify_email(self, token):
        # ステップ3：フェイクトークン検証
        return True  # 常に成功を返す

    def login_user(self, email, password):
        # ステップ4：シンプルなログイン
        return {"token": "fake_jwt_token"}
\`\`\`

### 2. Mock/Stubを活用した依存性除去

\`\`\`python
# REDフェーズテスト
def test_order_processing_should_send_email():
    """注文処理時に確認メールを送信すべき"""
    order = Order(id="123", customer_email="customer@example.com")

    # Mock注入
    mock_email_service = Mock()
    order_service = OrderService(email_service=mock_email_service)

    # When
    order_service.process_order(order)

    # Then
    mock_email_service.send_order_confirmation.assert_called_once_with(order)

# GREENフェーズ最小実装
class OrderService:
    def __init__(self, email_service):
        self.email_service = email_service

    def process_order(self, order):
        # 最小限の注文処理ロジック
        order.status = "processed"
        order.processed_at = datetime.now()

        # メール送信（実際のロジックなしで委譲のみ）
        self.email_service.send_order_confirmation(order)

        return order
\`\`\`

______________________________________________________________________

## パフォーマンスvs機能実装のバランス

### 1. パフォーマンス最適化の延期

GREENフェーズではパフォーマンスを考慮しません：

\`\`\`python
# 悪い例：GREENフェーズでパフォーマンス最適化を試行
class UserService:
    def __init__(self):
        self.user_cache = LRUCache(maxsize=1000)  # 不要な複雑さ
        self.db_pool = ConnectionPool(max_connections=20)  # 過剰エンジニアリング

# 良い例：シンプルな実装
class UserService:
    def __init__(self):
        self.users = {}  # シンプルなメモリストレージ

    def get_user(self, user_id):
        return self.users.get(user_id)  # 最小限の実装
\`\`\`

______________________________________________________________________

## 実践コード例

### Python例：ユーザー認証サービス

#### REDフェーズテスト（前フェーズで作成）

\`\`\`python
# tests/test_auth.py
def test_login_with_valid_credentials_should_return_jwt_token():
    """有効な認証情報でログイン時にJWTトークンを返すべき"""
    login_data = {"email": "test@example.com", "password": "correct_password"}
    response = client.post("/auth/login", json=login_data)

    assert response.status_code == 200
    assert "access_token" in response.json()
    assert response.json()["token_type"] == "bearer"
\`\`\`

#### GREENフェーズ最小実装

\`\`\`python
# src/auth_service.py
import jwt
from datetime import datetime, timedelta
from typing import Dict, Any

class AuthService:
    def __init__(self):
        # フェイクユーザーデータベース
        self.users = {
            "test@example.com": {
                "password": "correct_password",
                "user_id": "user_123"
            }
        }
        self.secret_key = "fake_secret_key_for_testing"

    def authenticate(self, email: str, password: str) -> Dict[str, Any]:
        """最小限の認証ロジック"""
        # ユーザー確認
        if email not in self.users:
            raise AuthenticationError("Invalid credentials")

        # パスワード確認（単純な文字列比較）
        if self.users[email]["password"] != password:
            raise AuthenticationError("Invalid credentials")

        # JWTトークン生成（最小限のクレームのみ）
        token_payload = {
            "sub": self.users[email]["user_id"],
            "email": email,
            "exp": datetime.utcnow() + timedelta(hours=24)
        }

        access_token = jwt.encode(token_payload, self.secret_key, algorithm="HS256")

        return {
            "access_token": access_token,
            "token_type": "bearer"
        }
\`\`\`

**実行結果:**

\`\`\`bash
$ pytest tests/test_auth.py -v
============================ test session starts ============================
tests/test_auth.py::test_login_with_valid_credentials_should_return_jwt_token PASSED
tests/test_auth.py::test_login_with_invalid_credentials_should_return_401 PASSED

============================= 2 passed in 0.12s ==============================
\`\`\`

______________________________________________________________________

## GREENフェーズで避けるべきこと

### 1. 過度な設計（Over-Engineering）

**避けるべきこと：**

- 複雑なデザインパターンの適用
- 不要な抽象化レイヤー
- 将来の要件予測
- パフォーマンス最適化の試行

\`\`\`python
# 悪い例：GREENフェーズでの過剰エンジニアリング
class UserFactory(AbstractFactory):
    def create_user(self, user_type: UserType) -> User:
        if user_type == UserType.ADMIN:
            return AdminUserBuilder().build()
        elif user_type == UserType.CUSTOMER:
            return CustomerUserBuilder().build()
        # ... 複雑なファクトリーパターン

# 良い例：シンプルな実装
class UserService:
    def create_user(self, user_data):
        if not user_data.get("email"):
            raise ValueError("Email required")

        return User(
            id=f"user_{uuid.uuid4().hex[:8]}",
            email=user_data["email"],
            name=user_data.get("name", "")
        )
\`\`\`

### 2. 外部依存性の導入

**GREENフェーズで避けるべき外部依存性：**

- データベース接続
- 外部API呼び出し
- メッセージキュー
- ファイルシステムアクセス

\`\`\`python
# 悪い例：不要な外部依存性
class UserService:
    def __init__(self):
        self.db = PostgreSQLDatabase("connection_string")  # 不要
        self.redis = RedisClient()                        # 不要
        self.email_api = SendGridAPI()                    # 不要

# 良い例：依存性のない実装
class UserService:
    def __init__(self):
        self.users = {}  # シンプルなメモリストレージ

    def create_user(self, user_data):
        user = User(
            id=f"user_{len(self.users) + 1}",
            email=user_data["email"],
            name=user_data.get("name", "")
        )
        self.users[user.id] = user
        return user
\`\`\`

______________________________________________________________________

## Gitコミット戦略（GREENフェーズ）

### 1. コミットメッセージ規則

GREENフェーズのコミットは実装完了を示すべきです：

\`\`\`bash
# 良いコミットメッセージ例
git commit -m "🟢 feat(AUTH-001): implement user authentication service

- Add AuthService with basic email/password validation
- Add JWT token generation functionality
- Add /auth/login endpoint with proper error handling
- Implement in-memory user storage for testing

All authentication tests now passing. Next: REFACTOR phase."

# 簡潔版
git commit -m "🟢 feat(AUTH-001): implement basic auth functionality"
\`\`\`

### 2. コミット単位と内容

**一つのGREENコミットに含まれるべき内容：**

- REDフェーズで失敗していたすべてのテストを通過させる最小実装
- 関連するドメインロジック
- 基本的なエラーハンドリング

______________________________________________________________________

## GREENフェーズチェックリスト

### 実装品質チェックリスト

- [ ] **すべてのテスト通過**：REDフェーズで作成したすべてのテストが通過するか？
- [ ] **最小実装**：YAGNI原則に従っているか？
- [ ] **シンプルさ**：コードがシンプルで理解しやすいか？
- [ ] **機能中心**：パフォーマンス最適化より機能実装に集中したか？

### 機能正確性チェックリスト

- [ ] **要件充足**：テストが検証するすべての要件を実装したか？
- [ ] **エッジケース**：境界値と例外ケースを処理するか？
- [ ] **エラーハンドリング**：適切なエラーを返すか？
- [ ] **データ有効性**：入力データ検証を実装したか？

### 技術的決定チェックリスト

- [ ] **依存性最小化**：不要な外部依存性を避けたか？
- [ ] **メモリベース**：データベースの代わりにメモリストレージを使用したか？
- [ ] **ハードコーディング許可**：シンプルなハードコーディングを適切に使用したか？
- [ ] **Mock/Stub使用**：外部サービスの代わりにMock/Stubを使用したか？

### Gitワークフローチェックリスト

- [ ] **コミットメッセージ**：GREENフェーズ完了を明確に表示したか？
- [ ] **タグ接続**：@TAG:IDで関連SPECと接続したか？
- [ ] **ファイル管理**：不要なファイルをコミットしなかったか？
- [ ] **ブランチ整理**：適切なブランチで作業したか？

______________________________________________________________________

## 結論

GREENフェーズは、TDDサイクルで**実際の機能を実装する最初のステップ**です。このフェーズの成功は：

1. **迅速なフィードバックループ**を通じて開発速度を高め
2. **シンプルな実装**を通じて複雑さを管理し
3. **テスト通過**を通じて進捗状況を明確に示し
4. **REFACTORフェーズの準備**を通じてコード品質改善の基盤を作ります

GREENフェーズで最も重要なことは、**「完璧な実装ではなく通過する実装」**という事実を覚えておくことです。

**GREENフェーズの成功はREFACTORフェーズの成功を保証します！** 🚀

______________________________________________________________________

## 次のステップ

GREENフェーズを完了したら、次のステップに進んでください：

- [**REFACTORフェーズガイド**](./refactor.md) - コード品質改善とリファクタリング
