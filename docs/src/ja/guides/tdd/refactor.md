# TDD REFACTORフェーズガイド：コード品質改善とリファクタリング

## 目次

1. [REFACTORフェーズの目標とタイミング](#refactorフェーズの目標とタイミング)
2. [コード臭と重複の除去](#コード臭と重複の除去)
3. [パフォーマンス最適化技法](#パフォーマンス最適化技法)
4. [可読性向上戦略](#可読性向上戦略)
5. [SOLID原則の適用](#solid原則の適用)
6. [安全なリファクタリング技法](#安全なリファクタリング技法)
7. [実践コード例（前後比較）](#実践コード例前後比較)
8. [Gitコミット戦略（REFACTORフェーズ）](#gitコミット戦略refactorフェーズ)
9. [REFACTORフェーズチェックリスト](#refactorフェーズチェックリスト)

______________________________________________________________________

## REFACTORフェーズの目標とタイミング

### REFACTORフェーズの核心目標

REFACTORフェーズは、**「動作するコードをより良くすること」**です。核心目標は：

\`\`\`mermaid
graph TD
    A[GREENフェーズ<br/>通過するコード] --> B[REFACTORフェーズ<br/>コード改善]
    B --> C[より良い設計]
    B --> D[より良い可読性]
    B --> E[より良いパフォーマンス]
    C --> F[テスト継続通過]
    D --> F
    E --> F

    style A fill:#4caf50
    style B fill:#9c27b0
    style C fill:#2196f3
    style D fill:#ff9800
    style E fill:#f44336
    style F fill:#4caf50
\`\`\`

### 1. リファクタリングの定義

**リファクタリングとは：**

- **外部動作を変更せずに**コードの内部構造を改善するプロセス
- コードの理解度、保守性、パフォーマンスを向上させる活動
- **テスト保護の下で**安全に進められるコード改善

**リファクタリングでないもの：**

- 新機能の追加（それは機能開発）
- バグ修正（それはデバッグ）
- テストなしでのコード変更（それは危険な行為）

### 2. リファクタリングのタイミング

**適切なリファクタリングタイミング：**

\`\`\`mermaid
graph LR
    A[GREENフェーズ完了] --> B[すべてのテスト通過]
    B --> C[リファクタリング開始]
    C --> D[テスト実行]
    D --> E{テスト通過?}
    E -->|はい| F[リファクタリング継続]
    E -->|いいえ| G[変更を元に戻す]
    F --> D
    G --> C

    style A fill:#4caf50
    style B fill:#81c784
    style C fill:#9c27b0
    style D fill:#ffeb3b
    style E fill:#ff9800
    style F fill:#9c27b0
    style G fill:#f44336
\`\`\`

**リファクタリングの黄金律：**

- ✅ すべてのテストが通過した直後
- ✅ コードが動作することが確認された後
- ✅ 新機能追加前（コードベースをクリーンに）
- ❌ テストが失敗している時
- ❌ 新機能開発と同時
- ❌ リリース直前

______________________________________________________________________

## コード臭と重複の除去

### 1. 重複コード（Duplicated Code）

**悪い例（重複が多い）：**

\`\`\`python
class UserService:
    def create_user(self, user_data):
        if not user_data.get("email"):
            raise ValueError("Email is required")
        if not "@" in user_data["email"]:
            raise ValueError("Invalid email format")
        
        user = User(**user_data)
        self.users[user.id] = user
        return user

    def update_user(self, user_id, user_data):
        if not user_data.get("email"):
            raise ValueError("Email is required")
        if not "@" in user_data["email"]:
            raise ValueError("Invalid email format")
        
        user = self.users[user_id]
        user.update(user_data)
        return user
\`\`\`

**良い例（重複除去）：**

\`\`\`python
class UserService:
    def _validate_email(self, email):
        """メール検証ロジックを一箇所に集約"""
        if not email:
            raise ValueError("Email is required")
        if not "@" in email:
            raise ValueError("Invalid email format")

    def create_user(self, user_data):
        self._validate_email(user_data.get("email"))
        user = User(**user_data)
        self.users[user.id] = user
        return user

    def update_user(self, user_id, user_data):
        self._validate_email(user_data.get("email"))
        user = self.users[user_id]
        user.update(user_data)
        return user
\`\`\`

### 2. 長い関数（Long Method）

**悪い例（長すぎる関数）：**

\`\`\`python
def process_order(order_data):
    # 注文検証
    if not order_data.get("items"):
        raise ValueError("Items required")
    if len(order_data["items"]) == 0:
        raise ValueError("At least one item required")
    
    # 在庫確認
    for item in order_data["items"]:
        product = products_db.get(item["product_id"])
        if not product:
            raise ValueError(f"Product {item['product_id']} not found")
        if product.stock < item["quantity"]:
            raise ValueError(f"Insufficient stock for {product.name}")
    
    # 価格計算
    total = 0
    for item in order_data["items"]:
        product = products_db.get(item["product_id"])
        total += product.price * item["quantity"]
    
    # 配送費追加
    if total < 50:
        total += 5
    
    # 税金計算
    tax = total * 0.1
    total_with_tax = total + tax
    
    # 在庫減少
    for item in order_data["items"]:
        product = products_db.get(item["product_id"])
        product.stock -= item["quantity"]
        products_db.save(product)
    
    # 注文作成
    order = Order(
        items=order_data["items"],
        total=total_with_tax,
        status="pending"
    )
    orders_db.save(order)
    
    return order
\`\`\`

**良い例（小さな関数に分割）：**

\`\`\`python
def process_order(order_data):
    """注文処理のメインフロー"""
    validate_order_data(order_data)
    check_inventory(order_data["items"])
    total = calculate_total_with_shipping_and_tax(order_data["items"])
    decrease_inventory(order_data["items"])
    order = create_order(order_data["items"], total)
    return order

def validate_order_data(order_data):
    """注文データ検証"""
    if not order_data.get("items"):
        raise ValueError("Items required")
    if len(order_data["items"]) == 0:
        raise ValueError("At least one item required")

def check_inventory(items):
    """在庫確認"""
    for item in items:
        product = products_db.get(item["product_id"])
        if not product:
            raise ValueError(f"Product {item['product_id']} not found")
        if product.stock < item["quantity"]:
            raise ValueError(f"Insufficient stock for {product.name}")

def calculate_total_with_shipping_and_tax(items):
    """合計金額計算（配送費と税金含む）"""
    subtotal = sum(
        products_db.get(item["product_id"]).price * item["quantity"]
        for item in items
    )
    
    shipping = 5 if subtotal < 50 else 0
    tax = subtotal * 0.1
    
    return subtotal + shipping + tax

def decrease_inventory(items):
    """在庫減少"""
    for item in items:
        product = products_db.get(item["product_id"])
        product.stock -= item["quantity"]
        products_db.save(product)

def create_order(items, total):
    """注文作成"""
    order = Order(items=items, total=total, status="pending")
    orders_db.save(order)
    return order
\`\`\`

### 3. 大きなクラス（Large Class）

**悪い例（責任が多すぎるクラス）：**

\`\`\`python
class UserManager:
    def create_user(self, user_data):
        # ユーザー作成
        pass
    
    def update_user(self, user_id, user_data):
        # ユーザー更新
        pass
    
    def delete_user(self, user_id):
        # ユーザー削除
        pass
    
    def send_welcome_email(self, user):
        # ウェルカムメール送信
        pass
    
    def send_password_reset_email(self, user):
        # パスワードリセットメール送信
        pass
    
    def calculate_user_statistics(self, user):
        # ユーザー統計計算
        pass
    
    def export_user_data_to_csv(self, user):
        # CSVエクスポート
        pass
\`\`\`

**良い例（単一責任原則に従った分割）：**

\`\`\`python
class UserService:
    """ユーザーCRUD操作のみ"""
    def create_user(self, user_data):
        pass
    
    def update_user(self, user_id, user_data):
        pass
    
    def delete_user(self, user_id):
        pass

class UserEmailService:
    """ユーザーメール関連操作のみ"""
    def send_welcome_email(self, user):
        pass
    
    def send_password_reset_email(self, user):
        pass

class UserStatisticsService:
    """ユーザー統計関連操作のみ"""
    def calculate_statistics(self, user):
        pass

class UserExportService:
    """ユーザーデータエクスポート操作のみ"""
    def export_to_csv(self, user):
        pass
\`\`\`

______________________________________________________________________

## パフォーマンス最適化技法

### 1. データベースクエリ最適化

**悪い例（N+1問題）：**

\`\`\`python
def get_users_with_posts():
    """すべてのユーザーと投稿を取得（非効率）"""
    users = User.objects.all()
    result = []
    
    for user in users:
        user_data = {
            "id": user.id,
            "name": user.name,
            "posts": []
        }
        
        # 各ユーザーごとに個別クエリ（N+1問題）
        posts = Post.objects.filter(user_id=user.id)
        for post in posts:
            user_data["posts"].append({
                "title": post.title,
                "content": post.content
            })
        
        result.append(user_data)
    
    return result
\`\`\`

**良い例（Eager Loadingで最適化）：**

\`\`\`python
def get_users_with_posts():
    """すべてのユーザーと投稿を取得（最適化）"""
    # select_relatedまたはprefetch_relatedを使用
    users = User.objects.prefetch_related('posts').all()
    
    result = []
    for user in users:
        user_data = {
            "id": user.id,
            "name": user.name,
            "posts": [
                {"title": post.title, "content": post.content}
                for post in user.posts.all()
            ]
        }
        result.append(user_data)
    
    return result
\`\`\`

### 2. キャッシング戦略

**悪い例（キャッシングなし）：**

\`\`\`python
class ProductService:
    def get_product_details(self, product_id):
        # 毎回データベースに問い合わせ
        product = db.query("SELECT * FROM products WHERE id = ?", product_id)
        
        # 複雑な計算
        ratings = db.query("SELECT * FROM ratings WHERE product_id = ?", product_id)
        average_rating = sum(r.rating for r in ratings) / len(ratings)
        
        return {
            "product": product,
            "average_rating": average_rating
        }
\`\`\`

**良い例（キャッシング適用）：**

\`\`\`python
from functools import lru_cache
from datetime import datetime, timedelta

class ProductService:
    def __init__(self):
        self.cache = {}
        self.cache_ttl = timedelta(minutes=5)
    
    def get_product_details(self, product_id):
        """キャッシング付きで商品詳細を取得"""
        cache_key = f"product_{product_id}"
        
        # キャッシュ確認
        if cache_key in self.cache:
            cached_data, cached_time = self.cache[cache_key]
            if datetime.now() - cached_time < self.cache_ttl:
                return cached_data
        
        # キャッシュミス：データベースから取得
        product = db.query("SELECT * FROM products WHERE id = ?", product_id)
        ratings = db.query("SELECT * FROM ratings WHERE product_id = ?", product_id)
        average_rating = sum(r.rating for r in ratings) / len(ratings) if ratings else 0
        
        result = {
            "product": product,
            "average_rating": average_rating
        }
        
        # キャッシュに保存
        self.cache[cache_key] = (result, datetime.now())
        
        return result
\`\`\`

______________________________________________________________________

## 可読性向上戦略

### 1. 意味のある変数名

**悪い例（不明瞭な変数名）：**

\`\`\`python
def calc(a, b, c):
    x = a * b
    y = x * c
    z = y * 0.1
    return x + y + z
\`\`\`

**良い例（明確な変数名）：**

\`\`\`python
def calculate_total_price_with_tax(quantity, unit_price, discount_rate):
    subtotal = quantity * unit_price
    discounted_price = subtotal * discount_rate
    tax = discounted_price * 0.1
    total = subtotal + discounted_price + tax
    return total
\`\`\`

### 2. マジックナンバーの除去

**悪い例（マジックナンバー）：**

\`\`\`python
def calculate_shipping_cost(weight):
    if weight < 1:
        return 5
    elif weight < 5:
        return 10
    elif weight < 10:
        return 15
    else:
        return 20
\`\`\`

**良い例（定数化）：**

\`\`\`python
# 定数として定義
SHIPPING_COST_LIGHT = 5      # 1kg未満
SHIPPING_COST_MEDIUM = 10    # 1-5kg
SHIPPING_COST_HEAVY = 15     # 5-10kg
SHIPPING_COST_EXTRA_HEAVY = 20  # 10kg以上

WEIGHT_THRESHOLD_LIGHT = 1
WEIGHT_THRESHOLD_MEDIUM = 5
WEIGHT_THRESHOLD_HEAVY = 10

def calculate_shipping_cost(weight):
    """重量に基づいて配送費を計算"""
    if weight < WEIGHT_THRESHOLD_LIGHT:
        return SHIPPING_COST_LIGHT
    elif weight < WEIGHT_THRESHOLD_MEDIUM:
        return SHIPPING_COST_MEDIUM
    elif weight < WEIGHT_THRESHOLD_HEAVY:
        return SHIPPING_COST_HEAVY
    else:
        return SHIPPING_COST_EXTRA_HEAVY
\`\`\`

### 3. 複雑な条件の簡素化

**悪い例（複雑なネスト条件）：**

\`\`\`python
def can_user_access_resource(user, resource):
    if user:
        if user.is_active:
            if user.role == "admin":
                return True
            else:
                if resource.is_public:
                    return True
                else:
                    if resource.owner_id == user.id:
                        return True
                    else:
                        return False
        else:
            return False
    else:
        return False
\`\`\`

**良い例（早期リターンとクリアな条件）：**

\`\`\`python
def can_user_access_resource(user, resource):
    """ユーザーがリソースにアクセスできるか確認"""
    # 早期リターン：無効なケース
    if not user or not user.is_active:
        return False
    
    # 管理者は常にアクセス可能
    if user.role == "admin":
        return True
    
    # 公開リソースは誰でもアクセス可能
    if resource.is_public:
        return True
    
    # 所有者のみアクセス可能
    return resource.owner_id == user.id
\`\`\`

______________________________________________________________________

## SOLID原則の適用

### 1. 単一責任原則（Single Responsibility Principle）

**悪い例（複数の責任）：**

\`\`\`python
class Invoice:
    def __init__(self, items):
        self.items = items
    
    def calculate_total(self):
        """合計計算"""
        return sum(item.price * item.quantity for item in self.items)
    
    def save_to_database(self):
        """データベース保存"""
        db.save(self)
    
    def generate_pdf(self):
        """PDF生成"""
        pdf = PDFGenerator()
        pdf.create(self)
        return pdf
    
    def send_email(self, recipient):
        """メール送信"""
        email = EmailService()
        email.send(recipient, self)
\`\`\`

**良い例（単一責任）：**

\`\`\`python
class Invoice:
    """請求書ビジネスロジックのみ"""
    def __init__(self, items):
        self.items = items
    
    def calculate_total(self):
        return sum(item.price * item.quantity for item in self.items)

class InvoiceRepository:
    """請求書データベース操作のみ"""
    def save(self, invoice):
        db.save(invoice)
    
    def find_by_id(self, invoice_id):
        return db.query(f"SELECT * FROM invoices WHERE id = {invoice_id}")

class InvoicePDFGenerator:
    """請求書PDF生成のみ"""
    def generate(self, invoice):
        pdf = PDFGenerator()
        pdf.create(invoice)
        return pdf

class InvoiceEmailService:
    """請求書メール送信のみ"""
    def send(self, invoice, recipient):
        email = EmailService()
        email.send(recipient, invoice)
\`\`\`

### 2. 開放閉鎖原則（Open/Closed Principle）

**悪い例（拡張のためにコード修正が必要）：**

\`\`\`python
class PaymentProcessor:
    def process_payment(self, payment_method, amount):
        if payment_method == "credit_card":
            # クレジットカード処理
            return self._process_credit_card(amount)
        elif payment_method == "paypal":
            # PayPal処理
            return self._process_paypal(amount)
        elif payment_method == "bank_transfer":
            # 銀行振込処理
            return self._process_bank_transfer(amount)
        # 新しい支払い方法を追加する度にこのメソッドを修正する必要がある
\`\`\`

**良い例（拡張に開放、修正に閉鎖）：**

\`\`\`python
from abc import ABC, abstractmethod

class PaymentMethod(ABC):
    """支払い方法の抽象基底クラス"""
    @abstractmethod
    def process(self, amount):
        pass

class CreditCardPayment(PaymentMethod):
    def process(self, amount):
        # クレジットカード処理ロジック
        return f"Processed {amount} via credit card"

class PayPalPayment(PaymentMethod):
    def process(self, amount):
        # PayPal処理ロジック
        return f"Processed {amount} via PayPal"

class BankTransferPayment(PaymentMethod):
    def process(self, amount):
        # 銀行振込処理ロジック
        return f"Processed {amount} via bank transfer"

class PaymentProcessor:
    def process_payment(self, payment_method: PaymentMethod, amount):
        """既存コードを変更せずに新しい支払い方法を追加可能"""
        return payment_method.process(amount)

# 使用例
processor = PaymentProcessor()
credit_card = CreditCardPayment()
result = processor.process_payment(credit_card, 100)
\`\`\`

______________________________________________________________________

## 安全なリファクタリング技法

### 1. テスト駆動リファクタリング

**リファクタリングプロセス：**

1. **テストが緑色であることを確認**
2. **小さな変更を加える**
3. **すべてのテストを実行**
4. **テストが通過** → 次の変更に進む
5. **テストが失敗** → 変更を元に戻す

\`\`\`python
# ステップ1：既存コード（すべてのテストが通過している状態）
class OrderService:
    def calculate_total(self, items):
        total = 0
        for item in items:
            total += item.price * item.quantity
        return total

# ステップ2：小さな変更（リスト内包表記に変更）
class OrderService:
    def calculate_total(self, items):
        return sum(item.price * item.quantity for item in items)

# ステップ3：テストを実行
$ pytest tests/test_order_service.py
# すべてのテストが通過 → 変更を保持
\`\`\`

### 2. 抽出リファクタリング

**メソッド抽出（Extract Method）：**

\`\`\`python
# リファクタリング前
def process_order(order):
    # 在庫確認
    for item in order.items:
        product = get_product(item.product_id)
        if product.stock < item.quantity:
            raise InsufficientStockError()
    
    # 価格計算
    total = 0
    for item in order.items:
        total += item.price * item.quantity
    
    return create_order(order.items, total)

# リファクタリング後
def process_order(order):
    check_stock_availability(order.items)
    total = calculate_total(order.items)
    return create_order(order.items, total)

def check_stock_availability(items):
    """在庫確認ロジックを抽出"""
    for item in items:
        product = get_product(item.product_id)
        if product.stock < item.quantity:
            raise InsufficientStockError()

def calculate_total(items):
    """価格計算ロジックを抽出"""
    return sum(item.price * item.quantity for item in items)
\`\`\`

______________________________________________________________________

## Gitコミット戦略（REFACTORフェーズ）

### 1. コミットメッセージ規則

REFACTORフェーズのコミットは改善内容を明確に示すべきです：

\`\`\`bash
# 良いコミットメッセージ例
git commit -m "🔵 refactor(AUTH-001): extract email validation logic

- Extract email validation to separate method
- Remove duplicated validation code
- Improve code readability

All tests still passing."

# 簡潔版
git commit -m "🔵 refactor(AUTH-001): extract email validation logic"
\`\`\`

### 2. リファクタリングコミットの原則

**一つのREFACTORコミットに含まれるべき内容：**

- 一つの明確なリファクタリング改善
- すべてのテストが引き続き通過する状態
- 外部動作の変更なし

\`\`\`bash
# コミット前：テスト実行
pytest tests/ -v
# すべてのテスト通過を確認

# リファクタリングコミット
git add src/auth_service.py
git commit -m "🔵 refactor(AUTH-001): extract email validation to separate method"

# コミット後：再度テスト実行
pytest tests/ -v
# すべてのテスト通過を確認
\`\`\`

______________________________________________________________________

## REFACTORフェーズチェックリスト

### コード品質チェックリスト

- [ ] **重複除去**：重複コードを除去したか？
- [ ] **関数サイズ**：関数が適切なサイズ（20行以下推奨）か？
- [ ] **クラス責任**：各クラスが単一責任を持つか？
- [ ] **命名**：変数名、関数名、クラス名が明確か？

### パフォーマンスチェックリスト

- [ ] **クエリ最適化**：N+1問題を解決したか？
- [ ] **キャッシング**：適切なキャッシング戦略を適用したか？
- [ ] **アルゴリズム**：効率的なアルゴリズムを使用しているか？
- [ ] **メモリ使用**：不要なメモリ使用を削減したか？

### 可読性チェックリスト

- [ ] **コメント**：複雑なロジックに適切なコメントがあるか？
- [ ] **一貫性**：コードスタイルが一貫しているか？
- [ ] **構造**：論理的な構造とフローか？
- [ ] **マジックナンバー**：すべての定数が定義されているか？

### SOLID原則チェックリスト

- [ ] **単一責任**：各クラスが一つの責任のみを持つか？
- [ ] **開放閉鎖**：拡張に開放、修正に閉鎖か？
- [ ] **リスコフ置換**：サブクラスが基底クラスと置換可能か？
- [ ] **インターフェース分離**：インターフェースが適切に分離されているか？
- [ ] **依存性逆転**：具象ではなく抽象に依存しているか？

### テストチェックリスト

- [ ] **すべてのテスト通過**：リファクタリング前後でテストが通過するか？
- [ ] **テストカバレッジ**：カバレッジが維持または向上したか？
- [ ] **テスト速度**：テスト実行速度が適切か？
- [ ] **テスト品質**：テストが明確で理解しやすいか？

### Gitワークフローチェックリスト

- [ ] **コミットメッセージ**：REFACTORフェーズであることを明確に表示したか？
- [ ] **小さなコミット**：各コミットが一つの改善に焦点を当てているか？
- [ ] **テスト確認**：各コミット後にテストを実行したか？
- [ ] **レビュー準備**：コードレビューの準備ができているか？

______________________________________________________________________

## 結論

REFACTORフェーズは、TDDサイクルで**コード品質を継続的に改善する**重要なステップです。このフェーズの成功は：

1. **保守性向上**を通じて将来の変更を容易にし
2. **可読性改善**を通じてチーム協業を促進し
3. **パフォーマンス最適化**を通じてユーザー体験を向上させ
4. **技術的負債削減**を通じて長期的な開発速度を維持します

REFACTORフェーズで最も重要なことは、**「テストの保護下で安全に改善する」**という原則を守ることです。

**良いリファクタリングは、将来の開発を加速させます！** 🚀

______________________________________________________________________

## 次のステップ

REFACTORフェーズを完了したら、次のサイクルに進んでください：

- 新しい機能のための**REDフェーズ**に戻る
- または、プロジェクト全体の品質レビューを実施
