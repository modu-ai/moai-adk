# 安全高级指南

增强MoAI-ADK项目安全性的综合指南。

## 🎯 安全原则

1. **默认安全**: 将安全设置设为默认值
2. **输入验证**: 验证所有外部输入
3. **最小权限**: 仅授予必要权限
4. **纵深防御**: 多层防御手段

## 🛡️ OWASP Top 10对策

### 1. SQL注入

```python
# ❌ 危险
user = db.execute(f"SELECT * FROM users WHERE email = '{email}'")

# ✅ 安全
user = db.execute(
    "SELECT * FROM users WHERE email = ?",
    [email]
)
```

### 2. 认证与会话管理

```python
# ✅ JWT令牌
token = jwt.encode(
    {"user_id": user.id, "exp": datetime.utcnow() + timedelta(hours=1)},
    SECRET_KEY,
    algorithm="HS256"
)

# ✅ 密码哈希
hashed = bcrypt.hashpw(password.encode(), bcrypt.gensalt())
```

### 3. 跨站脚本攻击 (XSS)

```python
# ✅ 自动转义 (Jinja2)
<p>{{ user_input }}</p>  <!-- 自动转义 -->

# ✅ 手动转义
from markupsafe import escape
safe_html = escape(user_input)
```

### 4. 跨站请求伪造 (CSRF)

```python
# ✅ CSRF令牌
<form method="post">
    <input type="hidden" name="csrf_token" value="{{ csrf_token() }}">
</form>
```

## 🔐 安全检查清单

### 开发阶段

- [ ] 实现输入验证
- [ ] 禁止记录敏感数据
- [ ] 最小化错误消息 (防止信息泄露)
- [ ] 实现认证/授权
- [ ] 密码哈希 (bcrypt/scrypt)

### 部署阶段

- [ ] 启用HTTPS
- [ ] 设置CORS策略
- [ ] 添加安全头
- [ ] 依赖项安全检查
- [ ] 数据库备份加密

### 运维阶段

- [ ] 访问日志监控
- [ ] 依赖项更新检查
- [ ] 入侵检测系统
- [ ] 定期安全审计
- [ ] 故障响应计划

## 🔑 加密

### 数据加密

```python
from cryptography.fernet import Fernet

# 对称加密
cipher = Fernet(key)
encrypted = cipher.encrypt(b"sensitive data")
decrypted = cipher.decrypt(encrypted)
```

### 通信加密

```bash
# HTTPS证书生成
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem

# Nginx配置
server {
    listen 443 ssl;
    ssl_certificate cert.pem;
    ssl_certificate_key key.pem;
}
```

## 🚨 安全漏洞扫描

### 自动扫描工具

```bash
# Python依赖项安全检查
safety check

# Snyk扫描
snyk test

# Bandit (静态分析)
bandit -r src/
```

### 扫描结果处理

```
vulnerability_db: 1.0.2 → 1.0.3 (update)
unpatched_dep: 2.0.0 (需要安全更新)

CRITICAL: 可能存在SQL注入 (src/api.py:45)
→ 需要立即修复
```

## 📋 制定安全策略

### 密码策略

```
- 最少12个字符
- 包含大写、小写、数字、特殊字符
- 每90天更改一次
- 禁止重复使用前5个密码
```

### 访问控制

```
RBAC (基于角色的访问控制):
- admin: 所有权限
- manager: 数据读写
- user: 仅自己的数据
```

______________________________________________________________________

**下一步**: [扩展与自定义](extensions.md) 或 [性能优化](performance.md)
