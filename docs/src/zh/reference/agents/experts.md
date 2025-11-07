# 专家 Agents 详细指南

Alfred 的 6 位领域专家的完整参考。

## 概览

| #   | 专家            | 领域                | 激活关键词                          | 技能数 |
| --- | --------------- | ------------------- | ----------------------------------- | ------ |
| 1   | backend-expert  | API, 服务器, DB     | server, api, database, microservice | 12个   |
| 2   | frontend-expert | UI, 状态管理, 性能  | frontend, ui, component, state      | 10个   |
| 3   | devops-expert   | 部署, CI/CD, 基础设施 | deploy, docker, kubernetes, ci/cd   | 14个   |
| 4   | ui-ux-expert    | 设计系统, 可访问性  | design, ux, accessibility, figma    | 8个    |
| 5   | security-expert | 安全, 认证          | security, auth, encryption, owasp   | 11个   |
| 6   | database-expert | DB 设计, 优化       | database, schema, query, index      | 9个    |

______________________________________________________________________

## 1. backend-expert

**领域**: API, 服务器, 数据库架构

### 激活条件

当 SPEC 包含以下关键词时自动激活:

- `server`, `api`, `endpoint`, `microservice`
- `authentication`, `authorization`
- `database`, `ORM`

### 专业领域

| 领域               | 技术栈                  | 职责                     |
| ------------------ | ----------------------- | ------------------------ |
| **API 设计**       | REST, GraphQL           | 编写 OpenAPI 3.1 规范    |
| **框架**           | FastAPI, Flask, Django  | 框架选择和结构设计       |
| **认证**           | JWT, OAuth 2.0, Session | 实现安全认证系统         |
| **微服务**         | Celery, RabbitMQ        | 异步任务处理             |
| **缓存**           | Redis, Memcached        | 性能优化                 |

### 主要职责

1. **API 设计**

   - 遵守 RESTful 原则
   - 设计端点结构
   - 定义请求/响应架构
   - 错误处理策略

2. **数据建模**

   - Entity-Relationship 图
   - ORM 模型设计
   - 关系设置 (1:1, 1:N, N:N)
   - 索引策略

3. **性能优化**

   - 查询优化
   - 数据库索引
   - 缓存策略
   - 负载均衡

### 示例: REST API 设计

```python
# @CODE:SPEC-002:backend-design
from fastapi import FastAPI, HTTPException, Depends
from sqlalchemy.orm import Session

app = FastAPI(title="Todo API v1.0")

# 端点设计
@app.post("/api/v1/todos", status_code=201)
async def create_todo(
    title: str,
    description: str,
    db: Session = Depends(get_db)
):
    """创建待办事项"""
    todo = Todo(title=title, description=description)
    db.add(todo)
    db.commit()
    return todo

@app.get("/api/v1/todos/{todo_id}")
async def get_todo(todo_id: int, db: Session = Depends(get_db)):
    """查询待办事项"""
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
    """修改待办事项"""
    todo = db.query(Todo).filter(Todo.id == todo_id).first()
    if not todo:
        raise HTTPException(status_code=404)
    todo.title = title
    todo.description = description
    db.commit()
    return todo

@app.delete("/api/v1/todos/{todo_id}")
async def delete_todo(todo_id: int, db: Session = Depends(get_db)):
    """删除待办事项"""
    todo = db.query(Todo).filter(Todo.id == todo_id).first()
    if not todo:
        raise HTTPException(status_code=404)
    db.delete(todo)
    db.commit()
    return {"status": "deleted"}
```

### 生成的成果

- OpenAPI 3.1 规范
- API 端点列表
- 请求/响应架构
- 错误代码文档
- 认证流程

______________________________________________________________________

## 2. frontend-expert

**领域**: UI 组件, 状态管理, 性能优化

### 激活条件

当 SPEC 包含以下关键词时自动激活:

- `frontend`, `ui`, `component`, `page`
- `state`, `store`, `context`
- `performance`, `optimization`

### 专业领域

| 领域           | 技术栈                    | 职责                   |
| -------------- | ------------------------- | ---------------------- |
| **框架**       | React 19, Vue 3.5, Angular 19 | 框架选择和结构     |
| **状态管理**   | Redux, Zustand, Pinia     | 全局状态设计           |
| **组件**       | Composition, Hooks        | 可重用组件             |
| **性能**       | 打包优化, 延迟加载        | 渲染性能改进           |
| **可访问性**   | WCAG 2.2, ARIA            | 支持所有用户           |

### 主要职责

1. **组件设计**

   - 可重用组件结构
   - Props 接口定义
   - 样式策略 (CSS-in-JS, Tailwind)

2. **状态管理**

   - 全局状态结构
   - 状态更新逻辑
   - 性能优化 (记忆化)

3. **性能优化**

   - 最小化打包大小
   - 渲染优化
   - 图片优化
   - 缓存策略

### 示例: React 组件设计

```typescript
// @CODE:SPEC-003:frontend-component
import React, { useState, useCallback } from 'react';
import { useTodoStore } from './store';

// 状态管理 (Zustand)
const useTodoStore = create((set) => ({
  todos: [],
  addTodo: (todo) => set((state) => ({
    todos: [...state.todos, todo]
  })),
  removeTodo: (id) => set((state) => ({
    todos: state.todos.filter(t => t.id !== id)
  }))
}));

// 可重用组件
const TodoItem = React.memo(({ todo, onRemove }) => (
  <div className="todo-item">
    <h3>{todo.title}</h3>
    <p>{todo.description}</p>
    <button onClick={() => onRemove(todo.id)}>
      删除
    </button>
  </div>
));

// 主组件
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
        placeholder="输入新待办事项"
      />
      <button onClick={handleAdd}>添加</button>

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

### 生成的成果

- 组件树图
- Props 接口定义
- 状态管理图
- 性能优化报告
- 可访问性检查结果

______________________________________________________________________

## 3. devops-expert

**领域**: 部署, CI/CD, 云基础设施

### 激活条件

当 SPEC 包含以下关键词时自动激活:

- `deploy`, `deployment`, `ci/cd`
- `docker`, `kubernetes`
- `infrastructure`, `cloud`

### 专业领域

| 领域               | 技术栈                | 职责                     |
| ------------------ | --------------------- | ------------------------ |
| **容器**           | Docker, Docker Compose | Dockerfile 和镜像管理    |
| **编排**           | Kubernetes, Helm      | 部署和扩展               |
| **CI/CD**          | GitHub Actions, GitLab CI | 自动化流水线         |
| **云**             | AWS, GCP, Azure       | 编写基础设施代码         |
| **监控**           | Prometheus, Grafana   | 性能监控                 |

### 主要职责

1. **部署流水线设计**

   - 测试 → 构建 → 部署自动化
   - 金丝雀部署和回滚策略
   - 零停机部署

2. **基础设施配置**

   - 生产环境设置
   - 负载均衡
   - 数据库备份/恢复

3. **监控和日志**

   - 应用性能监控
   - 日志收集和分析
   - 警报设置

### 示例: GitHub Actions CI/CD

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

      # 测试
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

      # 部署
      - name: Build Docker image
        run: docker build -t app:latest .

      - name: Deploy to production
        run: |
          docker tag app:latest app:${{ github.sha }}
          # 执行部署脚本
          ./scripts/deploy.sh
```

### 生成的成果

- Dockerfile
- docker-compose.yml
- Kubernetes manifests
- CI/CD 流水线
- 部署指南
- 监控配置

______________________________________________________________________

## 4. ui-ux-expert

**领域**: 设计系统, 可访问性, 用户体验

### 激活条件

当 SPEC 包含以下关键词时自动激活:

- `design`, `ui`, `ux`
- `accessibility`, `a11y`
- `figma`, `design-system`

### 专业领域

| 领域             | 技术栈           | 职责               |
| ---------------- | ---------------- | ------------------ |
| **设计系统**     | Figma, Storybook | 组件库             |
| **可访问性**     | WCAG 2.2, ARIA   | 包容所有用户       |
| **用户研究**     | 用户测试, 分析   | UX 改进            |
| **性能**         | 加载时间, 响应性 | 用户满意度         |

### 主要职责

1. **构建设计系统**

   - 定义颜色、字体、间距
   - 组件库
   - 设计令牌

2. **保证可访问性**

   - 屏幕阅读器支持
   - 键盘导航
   - 颜色对比度

3. **改进用户体验**

   - 用户测试
   - 收集反馈
   - 持续改进

### 示例: 可访问性检查清单

```markdown
# WCAG 2.2 可访问性检查

## 可感知性 (Perceivable)
- [ ] 为图片提供替代文本
- [ ] 不仅仅依靠颜色传递信息
- [ ] 对比度 4.5:1 以上

## 可操作性 (Operable)
- [ ] 所有功能都可以用键盘操作
- [ ] 焦点顺序合理
- [ ] 无闪烁内容

## 可理解性 (Understandable)
- [ ] 文本可读性高
- [ ] 可预测的导航
- [ ] 清晰的错误消息

## 健壮性 (Robust)
- [ ] 有效的 HTML/CSS
- [ ] 正确使用 ARIA
- [ ] 兼容性测试通过
```

### 生成的成果

- 设计系统指南
- Figma 组件库
- Storybook 文档
- 可访问性审计报告
- 用户测试结果

______________________________________________________________________

## 5. security-expert

**领域**: 安全, 认证, 加密

### 激活条件

当 SPEC 包含以下关键词时自动激活:

- `security`, `auth`, `encryption`
- `vulnerability`, `owasp`
- `compliance`, `privacy`

### 专业领域

| 领域         | 技术栈               | 职责             |
| ------------ | -------------------- | ---------------- |
| **认证**     | JWT, OAuth 2.0, SAML | 安全认证系统     |
| **加密**     | AES-256, RSA, HTTPS  | 数据保护         |
| **OWASP**    | Top 10, SAST/DAST    | 漏洞防护         |
| **访问控制** | RBAC, ABAC           | 权限管理         |
| **审计**     | 日志, 监控           | 安全事件跟踪     |

### 主要职责

1. **安全设计**

   - 威胁建模
   - 安全架构
   - 入侵防护策略

2. **漏洞防护**

   - 防止 SQL 注入
   - 防止 XSS
   - CSRF 令牌
   - 输入验证

3. **安全监控**

   - 日志分析
   - 入侵检测
   - 事件响应

### 示例: 安全实现

```python
# @CODE:SPEC-005:security-implementation
from flask import Flask, request
from werkzeug.security import generate_password_hash, check_password_hash
from cryptography.fernet import Fernet
import jwt

app = Flask(__name__)
SECRET_KEY = "your-secret-key"

# 密码哈希
def hash_password(password: str) -> str:
    """安全地哈希密码"""
    return generate_password_hash(password, method='pbkdf2:sha256')

def verify_password(hashed, password: str) -> bool:
    """验证密码"""
    return check_password_hash(hashed, password)

# JWT 令牌
def create_token(user_id: int) -> str:
    """生成 JWT 令牌"""
    return jwt.encode(
        {'user_id': user_id},
        SECRET_KEY,
        algorithm='HS256'
    )

# 输入验证
@app.before_request
def validate_input():
    """验证所有输入"""
    if request.method == 'POST':
        # CSRF 令牌验证
        token = request.headers.get('X-CSRF-Token')
        if not verify_csrf_token(token):
            return {'error': 'Invalid CSRF token'}, 403

        # 防止 SQL 注入 (参数化查询)
        # - 使用 SQLAlchemy ORM

        # 防止 XSS (HTML 转义)
        # - Jinja2 自动转义

# 强制 HTTPS
@app.after_request
def secure_headers(response):
    """设置安全头"""
    response.headers['X-Content-Type-Options'] = 'nosniff'
    response.headers['X-Frame-Options'] = 'DENY'
    response.headers['X-XSS-Protection'] = '1; mode=block'
    response.headers['Strict-Transport-Security'] = 'max-age=31536000'
    return response
```

### 生成的成果

- 安全策略文档
- 威胁建模图
- 安全审计检查清单
- 渗透测试报告
- 合规文档 (GDPR, HIPAA)

______________________________________________________________________

## 6. database-expert

**领域**: 数据库设计, 优化, 迁移

### 激活条件

当 SPEC 包含以下关键词时自动激活:

- `database`, `db`, `schema`
- `query`, `index`, `migration`
- `optimization`, `performance`

### 专业领域

| 领域         | 技术栈                     | 职责               |
| ------------ | -------------------------- | ------------------ |
| **设计**     | PostgreSQL, MySQL, MongoDB | 架构设计           |
| **优化**     | 索引, 查询调优             | 性能改进           |
| **迁移**     | Alembic, Flyway            | 版本管理           |
| **扩展性**   | 分区, 分片                 | 大规模数据处理     |
| **备份**     | PITR, 复制                 | 数据安全性         |

### 主要职责

1. **数据库设计**

   - Entity-Relationship 图
   - 规范化 (1NF ~ 3NF)
   - 约束设置

2. **性能优化**

   - 创建适当的索引
   - 查询优化
   - 执行计划分析

3. **迁移管理**

   - 版本控制
   - 回滚策略
   - 零停机迁移

### 示例: 数据库设计

```sql
-- @CODE:SPEC-006:database-schema
-- 用户表
CREATE TABLE users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_email (email)
);

-- 待办事项表
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

-- 关系图
-- users (1) --< (N) todos
```

### 生成的成果

- ERD (Entity-Relationship Diagram)
- DDL (Data Definition Language) 脚本
- 迁移脚本
- 性能调优报告
- 备份/恢复计划

______________________________________________________________________

## 专家激活矩阵

| SPEC 关键词 | backend | frontend | devops | ui-ux | security | database |
| ----------- | ------- | -------- | ------ | ----- | -------- | -------- |
| API         | ✅      |          |        |       |          |          |
| Frontend    |         | ✅       |        | ✅    |          |          |
| Database    | ✅      |          |        |       |          | ✅       |
| Deploy      |         |          | ✅     |       |          |          |
| Security    |         |          |        |       | ✅       |          |
| Performance | ✅      | ✅       |        | ✅    |          | ✅       |

______________________________________________________________________

**下一步**: [核心 Sub-agents](core.md) 或 [Agents 概览](index.md)
