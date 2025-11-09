# エキスパートエージェント詳細ガイド

Alfredの6人のドメインエキスパートの完全なリファレンスです。

## 概要

| #   | エキスパート        | ドメイン                | アクティベーションキーワード                  | スキル数 |
| --- | ------------------- | ----------------------- | --------------------------------------------- | -------- |
| 1   | backend-expert       | API、サーバー、DB       | server, api, database, microservice          | 12個     |
| 2   | frontend-expert      | UI、状態管理、パフォーマンス | frontend, ui, component, state            | 10個     |
| 3   | devops-expert        | デプロイ、CI/CD、インフラ | deploy, docker, kubernetes, ci/cd          | 14個     |
| 4   | ui-ux-expert         | デザインシステム、アクセシビリティ | design, ux, accessibility, figma        | 8個      |
| 5   | security-expert      | セキュリティ、認証      | security, auth, encryption, owasp           | 11個     |
| 6   | database-expert      | DB設計、最適化          | database, schema, query, index               | 9個      |

______________________________________________________________________

## 1. backend-expert

**ドメイン**: API、サーバー、データベースアーキテクチャ

### アクティベーション条件

SPECに次のキーワードが含まれている場合、自動アクティベーション:

- `server`, `api`, `endpoint`, `microservice`
- `authentication`, `authorization`
- `database`, `ORM`

### 専門分野

| 領域               | 技術スタック              | 責任                         |
| ------------------ | ------------------------- | ---------------------------- |
| **API設計**        | REST, GraphQL             | OpenAPI 3.1仕様作成          |
| **フレームワーク** | FastAPI, Flask, Django    | フレームワーク選択と構造設計 |
| **認証**           | JWT, OAuth 2.0, Session   | 安全な認証システム実装       |
| **マイクロサービス** | Celery, RabbitMQ        | 非同期タスク処理             |
| **キャッシング**   | Redis, Memcached          | パフォーマンス最適化         |

### 主要な責任

1. **API設計**

   - RESTful原則の遵守
   - エンドポイント構造の設計
   - リクエスト/レスポンススキーマ定義
   - エラー処理戦略

2. **データモデリング**

   - Entity-Relationshipダイアグラム
   - ORMモデル設計
   - 関係設定（1:1、1:N、N:N）
   - インデックス戦略

3. **パフォーマンス最適化**

   - クエリ最適化
   - データベースインデックス
   - キャッシング戦略
   - ロードバランシング

### 例: REST API設計

```python
# @CODE:SPEC-002:backend-design
from fastapi import FastAPI, HTTPException, Depends
from sqlalchemy.orm import Session

app = FastAPI(title="Todo API v1.0")

# エンドポイント設計
@app.post("/api/v1/todos", status_code=201)
async def create_todo(
    title: str,
    description: str,
    db: Session = Depends(get_db)
):
    """Todo作成"""
    todo = Todo(title=title, description=description)
    db.add(todo)
    db.commit()
    return todo

@app.get("/api/v1/todos/{todo_id}")
async def get_todo(todo_id: int, db: Session = Depends(get_db)):
    """Todo取得"""
    todo = db.query(Todo).filter(Todo.id == todo_id).first()
    if not todo:
        raise HTTPException(status_code=404)
    return todo

@app.put("/api/v1/todos/{todo_id}")
async def update_todo(
    todo_id: int,
    title: str,
    description: str,
    db: Session = Depends(get_db)
):
    """Todo更新"""
    todo = db.query(Todo).filter(Todo.id == todo_id).first()
    if not todo:
        raise HTTPException(status_code=404)
    todo.title = title
    todo.description = description
    db.commit()
    return todo

@app.delete("/api/v1/todos/{todo_id}")
async def delete_todo(todo_id: int, db: Session = Depends(get_db)):
    """Todo削除"""
    todo = db.query(Todo).filter(Todo.id == todo_id).first()
    if not todo:
        raise HTTPException(status_code=404)
    db.delete(todo)
    db.commit()
    return {"status": "deleted"}
```

### 生成成果物

- OpenAPI 3.1仕様
- APIエンドポイントリスト
- リクエスト/レスポンススキーマ
- エラーコードドキュメント
- 認証フロー

______________________________________________________________________

## 2. frontend-expert

**ドメイン**: UIコンポーネント、状態管理、パフォーマンス最適化

### アクティベーション条件

SPECに次のキーワードが含まれている場合、自動アクティベーション:

- `frontend`, `ui`, `component`, `page`
- `state`, `store`, `context`
- `performance`, `optimization`

### 専門分野

| 領域           | 技術スタック                  | 責任                    |
| -------------- | ----------------------------- | ----------------------- |
| **フレームワーク** | React 19, Vue 3.5, Angular 19 | フレームワーク選択と構造 |
| **状態管理**   | Redux, Zustand, Pinia         | グローバル状態設計      |
| **コンポーネント** | Composition, Hooks        | 再利用可能なコンポーネント |
| **パフォーマンス** | バンドル最適化、遅延ローディング | レンダリングパフォーマンス改善 |
| **アクセシビリティ** | WCAG 2.2, ARIA            | すべてのユーザーサポート |

### 主要な責任

1. **コンポーネント設計**

   - 再利用可能なコンポーネント構造
   - Propsインターフェース定義
   - スタイリング戦略（CSS-in-JS、Tailwind）

2. **状態管理**

   - グローバル状態構造
   - 状態更新ロジック
   - パフォーマンス最適化（メモ化）

3. **パフォーマンス最適化**

   - バンドルサイズ最小化
   - レンダリング最適化
   - 画像最適化
   - キャッシング戦略

### 例: Reactコンポーネント設計

```typescript
// @CODE:SPEC-003:frontend-component
import React, { useState, useCallback } from 'react';
import { useTodoStore } from './store';

// 状態管理（Zustand）
const useTodoStore = create((set) => ({
  todos: [],
  addTodo: (todo) => set((state) => ({
    todos: [...state.todos, todo]
  })),
  removeTodo: (id) => set((state) => ({
    todos: state.todos.filter(t => t.id !== id)
  }))
}));

// 再利用可能なコンポーネント
const TodoItem = React.memo(({ todo, onRemove }) => (
  <div className="todo-item">
    <h3>{todo.title}</h3>
    <p>{todo.description}</p>
    <button onClick={() => onRemove(todo.id)}>
      削除
    </button>
  </div>
));

// メインコンポーネント
export const TodoList = () => {
  const [input, setInput] = useState('');
  const { todos, addTodo } = useTodoStore();

  const handleAdd = useCallback(() => {
    if (input.trim()) {
      addTodo({ id: Date.now(), title: input });
      setInput('');
    }
  }, [input]);

  return (
    <div>
      <input
        value={input}
        onChange={(e) => setInput(e.target.value)}
        placeholder="新しいTodoを入力"
      />
      <button onClick={handleAdd}>追加</button>

      <div className="todo-list">
        {todos.map(todo => (
          <TodoItem
            key={todo.id}
            todo={todo}
            onRemove={() => /* remove */}
          />
        ))}
      </div>
    </div>
  );
};
```

### 生成成果物

- コンポーネントツリーダイアグラム
- Propsインターフェース定義書
- 状態管理ダイアグラム
- パフォーマンス最適化レポート
- アクセシビリティ検証結果

______________________________________________________________________

## 3. devops-expert

**ドメイン**: デプロイ、CI/CD、クラウドインフラ

### アクティベーション条件

SPECに次のキーワードが含まれている場合、自動アクティベーション:

- `deploy`, `deployment`, `ci/cd`
- `docker`, `kubernetes`
- `infrastructure`, `cloud`

### 専門分野

| 領域               | 技術スタック                 | 責任                      |
| ------------------ | ---------------------------- | ------------------------- |
| **コンテナ**       | Docker, Docker Compose       | Dockerfileおよびイメージ管理 |
| **オーケストレーション** | Kubernetes, Helm         | デプロイおよびスケーリング |
| **CI/CD**          | GitHub Actions, GitLab CI    | 自動化パイプライン         |
| **クラウド**       | AWS, GCP, Azure              | インフラコード作成         |
| **モニタリング**   | Prometheus, Grafana          | パフォーマンスモニタリング |

### 主要な責任

1. **デプロイパイプライン設計**

   - テスト → ビルド → デプロイ自動化
   - カナリアデプロイおよびロールバック戦略
   - ゼロダウンタイムデプロイ

2. **インフラ構成**

   - プロダクション環境設定
   - ロードバランシング
   - データベースバックアップ/復旧

3. **モニタリング&ロギング**

   - アプリケーションパフォーマンスモニタリング
   - ログ収集および分析
   - アラート設定

### 例: GitHub Actions CI/CD

```yaml
# @CODE:SPEC-004:devops-pipeline
name: CI/CD Pipeline

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-python@v4
        with:
          python-version: '3.13'

      # テスト
      - name: Install dependencies
        run: |
          python -m pip install --upgrade pip
          pip install -r requirements.txt

      - name: Run tests
        run: pytest --cov=src tests/

      - name: Check coverage
        run: |
          coverage report --fail-under=85

  deploy:
    needs: test
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      # デプロイ
      - name: Build Docker image
        run: docker build -t app:latest .

      - name: Deploy to production
        run: |
          docker tag app:latest app:${{ github.sha }}
          # デプロイスクリプト実行
          ./scripts/deploy.sh
```

### 生成成果物

- Dockerfile
- docker-compose.yml
- Kubernetesマニフェスト
- CI/CDパイプライン
- デプロイガイド
- モニタリング構成

______________________________________________________________________

## 4. ui-ux-expert

**ドメイン**: デザインシステム、アクセシビリティ、ユーザーエクスペリエンス

### アクティベーション条件

SPECに次のキーワードが含まれている場合、自動アクティベーション:

- `design`, `ui`, `ux`
- `accessibility`, `a11y`
- `figma`, `design-system`

### 専門分野

| 領域              | 技術スタック         | 責任                |
| ----------------- | -------------------- | ------------------- |
| **デザインシステム** | Figma, Storybook  | コンポーネントライブラリ |
| **アクセシビリティ** | WCAG 2.2, ARIA    | すべてのユーザー包含   |
| **ユーザー研究**   | ユーザーテスト、分析 | UX改善               |
| **パフォーマンス** | ロード時間、応答性   | ユーザー満足度       |

### 主要な責任

1. **デザインシステム構築**

   - 色、タイポグラフィ、間隔定義
   - コンポーネントライブラリ
   - デザイントークン

2. **アクセシビリティ保証**

   - スクリーンリーダーサポート
   - キーボードナビゲーション
   - 色コントラスト

3. **ユーザーエクスペリエンス改善**

   - ユーザーテスト
   - フィードバック収集
   - 継続的改善

### 例: アクセシビリティチェックリスト

```markdown
# WCAG 2.2アクセシビリティ検証

## 知覚可能（Perceivable）
- [ ] 画像に代替テキスト提供
- [ ] 色だけで情報伝達しない
- [ ] コントラスト比4.5:1以上

## 操作可能（Operable）
- [ ] すべての機能をキーボードで操作
- [ ] フォーカス順序が論理的
- [ ] 点滅コンテンツなし

## 理解可能（Understandable）
- [ ] テキスト読み取り可能性高い
- [ ] 予測可能なナビゲーション
- [ ] エラーメッセージ明確

## 堅牢（Robust）
- [ ] 有効なHTML/CSS
- [ ] ARIA正しい使用
- [ ] 互換性テスト通過
```

### 生成成果物

- デザインシステムガイド
- Figmaコンポーネントライブラリ
- Storybookドキュメント
- アクセシビリティ監査レポート
- ユーザーテスト結果

______________________________________________________________________

## 5. security-expert

**ドメイン**: セキュリティ、認証、暗号化

### アクティベーション条件

SPECに次のキーワードが含まれている場合、自動アクティベーション:

- `security`, `auth`, `encryption`
- `vulnerability`, `owasp`
- `compliance`, `privacy`

### 専門分野

| 領域          | 技術スタック            | 責任             |
| ------------- | ----------------------- | ---------------- |
| **認証**      | JWT, OAuth 2.0, SAML    | セキュア認証システム |
| **暗号化**    | AES-256, RSA, HTTPS     | データ保護       |
| **OWASP**     | Top 10, SAST/DAST        | 脆弱性防止       |
| **アクセス制御** | RBAC, ABAC           | 権限管理         |
| **監査**      | ロギング、モニタリング | セキュリティイベント追跡 |

### 主要な責任

1. **セキュリティ設計**

   - 脅威モデリング
   - セキュリティアーキテクチャ
   - 侵入防止戦略

2. **脆弱性防止**

   - SQLインジェクション防止
   - XSS防止
   - CSRFトークン
   - 入力検証

3. **セキュリティ監視**

   - ログ分析
   - 侵入検出
   - インシデント対応

### 例: セキュリティ実装

```python
# @CODE:SPEC-005:security-implementation
from flask import Flask, request
from werkzeug.security import generate_password_hash, check_password_hash
from cryptography.fernet import Fernet
import jwt

app = Flask(__name__)
SECRET_KEY = "your-secret-key"

# パスワードハッシュ化
def hash_password(password: str) -> str:
    """パスワードを安全にハッシュ化"""
    return generate_password_hash(password, method='pbkdf2:sha256')

def verify_password(hashed, password: str) -> bool:
    """パスワード検証"""
    return check_password_hash(hashed, password)

# JWTトークン
def create_token(user_id: int) -> str:
    """JWTトークン作成"""
    return jwt.encode(
        {'user_id': user_id},
        SECRET_KEY,
        algorithm='HS256'
    )

# 入力検証
@app.before_request
def validate_input():
    """すべての入力検証"""
    if request.method == 'POST':
        # CSRFトークン検証
        token = request.headers.get('X-CSRF-Token')
        if not verify_csrf_token(token):
            return {'error': 'Invalid CSRF token'}, 403

        # SQLインジェクション防止（パラメータ化クエリ）
        # - SQLAlchemy ORM使用

        # XSS防止（HTMLエスケープ）
        # - Jinja2自動エスケープ

# HTTPS強制
@app.after_request
def secure_headers(response):
    """セキュリティヘッダー設定"""
    response.headers['X-Content-Type-Options'] = 'nosniff'
    response.headers['X-Frame-Options'] = 'DENY'
    response.headers['X-XSS-Protection'] = '1; mode=block'
    response.headers['Strict-Transport-Security'] = 'max-age=31536000'
    return response
```

### 生成成果物

- セキュリティポリシードキュメント
- 脅威モデリングダイアグラム
- セキュリティ監査チェックリスト
- 侵入テストレポート
- コンプライアンスドキュメント（GDPR、HIPAA）

______________________________________________________________________

## 6. database-expert

**ドメイン**: データベース設計、最適化、マイグレーション

### アクティベーション条件

SPECに次のキーワードが含まれている場合、自動アクティベーション:

- `database`, `db`, `schema`
- `query`, `index`, `migration`
- `optimization`, `performance`

### 専門分野

| 領域             | 技術スタック                  | 責任               |
| ---------------- | ----------------------------- | ------------------ |
| **設計**         | PostgreSQL, MySQL, MongoDB    | スキーマ設計       |
| **最適化**       | インデックス、クエリチューニング | パフォーマンス改善 |
| **マイグレーション** | Alembic, Flyway            | バージョン管理     |
| **スケーラビリティ** | パーティショニング、シャーディング | 大規模データ処理 |
| **バックアップ** | PITR、レプリケーション         | データ安全性       |

### 主要な責任

1. **データベース設計**

   - Entity-Relationshipダイアグラム
   - 正規化（1NF〜3NF）
   - 制約設定

2. **パフォーマンス最適化**

   - 適切なインデックス作成
   - クエリ最適化
   - 実行計画分析

3. **マイグレーション管理**

   - バージョン制御
   - ロールバック戦略
   - ゼロダウンタイムマイグレーション

### 例: データベース設計

```sql
-- @CODE:SPEC-006:database-schema
-- ユーザーテーブル
CREATE TABLE users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_email (email)
);

-- Todoテーブル
CREATE TABLE todos (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    completed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_completed (user_id, completed)
);

-- 関係図
-- users (1) --< (N) todos
```

### 生成成果物

- ERD（Entity-Relationship Diagram）
- DDL（Data Definition Language）スクリプト
- マイグレーションスクリプト
- パフォーマンスチューニングレポート
- バックアップ/復旧計画

______________________________________________________________________

## エキスパートアクティベーションマトリックス

| SPECキーワード | backend | frontend | devops | ui-ux | security | database |
| -------------- | ------- | -------- | ------ | ----- | -------- | -------- |
| API            | ✅      |          |        |       |          |          |
| Frontend       |         | ✅       |        | ✅    |          |          |
| Database       | ✅      |          |        |       |          | ✅       |
| Deploy         |         |          | ✅     |       |          |          |
| Security       |         |          |        |       | ✅       |          |
| Performance    | ✅      | ✅       |        | ✅    |          | ✅       |

______________________________________________________________________

**次**: [コアサブエージェント](core.md)または[エージェント概要](index.md)



