# セキュリティ上級ガイド

MoAI-ADKプロジェクトのセキュリティを強化する包括的なガイドです。

## セキュリティ原則

1. **デフォルトで安全**: 安全な設定をデフォルトに
2. **入力検証**: すべての外部入力を検証
3. **最小権限**: 必要な権限のみを付与
4. **多層防御**: 複数の層の防御手段

## OWASP Top 10対応

### 1. SQLインジェクション

```python
# ❌ 危険
user = db.execute(f"SELECT * FROM users WHERE email = '{email}'")

# ✅ 安全
user = db.execute(
    "SELECT * FROM users WHERE email = ?",
    [email]
)
```

### 2. 認証とセッション管理

```python
# ✅ JWTトークン
token = jwt.encode(
    {"user_id": user.id, "exp": datetime.utcnow() + timedelta(hours=1)},
    SECRET_KEY,
    algorithm="HS256"
)

# ✅ パスワードハッシュ化
hashed = bcrypt.hashpw(password.encode(), bcrypt.gensalt())
```

### 3. クロスサイトスクリプティング (XSS)

```python
# ✅ 自動エスケープ (Jinja2)
<p>{{ user_input }}</p>  <!-- 自動的にエスケープされる -->

# ✅ 手動エスケープ
from markupsafe import escape
safe_html = escape(user_input)
```

### 4. クロスサイトリクエストフォージェリ (CSRF)

```python
# ✅ CSRFトークン
<form method="post">
    <input type="hidden" name="csrf_token" value="{{ csrf_token() }}">
</form>
```

## セキュリティチェックリスト

### 開発段階

- [ ] 入力検証を実装
- [ ] 機密データのログ記録を禁止
- [ ] エラーメッセージを最小化（情報漏洩防止）
- [ ] 認証/認可を実装
- [ ] パスワードハッシュ化（bcrypt/scrypt）

### デプロイ段階

- [ ] HTTPSを有効化
- [ ] CORSポリシーを設定
- [ ] セキュリティヘッダーを追加
- [ ] 依存関係のセキュリティチェック
- [ ] データベースバックアップの暗号化

### 運用段階

- [ ] アクセスログのモニタリング
- [ ] 依存関係の更新確認
- [ ] 侵入検知システム
- [ ] 定期的なセキュリティ監査
- [ ] インシデント対応計画

## 暗号化

### データ暗号化

```python
from cryptography.fernet import Fernet

# 対称暗号化
cipher = Fernet(key)
encrypted = cipher.encrypt(b"sensitive data")
decrypted = cipher.decrypt(encrypted)
```

### 通信暗号化

```bash
# HTTPS証明書生成
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem

# Nginx設定
server {
    listen 443 ssl;
    ssl_certificate cert.pem;
    ssl_certificate_key key.pem;
}
```

## セキュリティ脆弱性スキャン

### 自動スキャンツール

```bash
# Python依存関係のセキュリティチェック
safety check

# Snykスキャン
snyk test

# Bandit（静的解析）
bandit -r src/
```

### スキャン結果の処理

```
vulnerability_db: 1.0.2 → 1.0.3 (更新)
unpatched_dep: 2.0.0 (セキュリティ更新が必要)

CRITICAL: SQLインジェクションの可能性 (src/api.py:45)
→ 即座に修正が必要
```

## セキュリティポリシーの確立

### パスワードポリシー

```
- 最小12文字
- 大文字、小文字、数字、特殊文字を含む
- 90日ごとに変更
- 過去5つのパスワードの再利用を禁止
```

### アクセス制御

```
RBAC（ロールベースアクセス制御）:
- admin: すべての権限
- manager: データの読み書き
- user: 自分のデータのみ
```

______________________________________________________________________

**次**: [拡張とカスタマイズ](extensions.md) または [パフォーマンス最適化](performance.md)



