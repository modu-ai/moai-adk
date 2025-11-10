# TAGシステム完全リファレンス

MoAI-ADKの追跡可能性システムのコアであるTAGシステムです。

## 目的

CODE-FIRST原則でSPEC、TEST、CODE、DOCをすべて接続し、**完全な追跡可能性**を保証します。

```
SPEC-001（要件）
    ↓
@TEST:SPEC-001（テスト）
    ↓
@CODE:SPEC-001（実装）
    ↓
@DOC:SPEC-001（ドキュメント）
    ↓
相互参照（完全な追跡可能性）
```

## TAG種類

| TAG         | 場所         | 用途        | 例              |
| ----------- | ------------ | ----------- | --------------- |
| **SPEC-ID** | .moai/specs/ | 要件        | SPEC-001        |
| **@TEST**   | tests/       | テストコード | @TEST:SPEC-001:\* |
| **@CODE**   | src/         | 実装コード   | @CODE:SPEC-001:\* |
| **@DOC**    | docs/        | ドキュメント | @DOC:SPEC-001:\* |

## TAG作成規則

### SPEC TAG

```
SPEC-001: 最初の仕様
SPEC-002: 2番目の仕様
SPEC-N: N番目の仕様
```

### @TEST TAG

```python
# @TEST:SPEC-001:login_success
def test_login_success():
    pass

# @TEST:SPEC-001:login_failure
def test_login_failure():
    pass
```

### @CODE TAG

```python
# @CODE:SPEC-001:register_user
def register_user(email, password):
    pass

# @CODE:SPEC-001:validate_email
def validate_email(email):
    pass
```

### @DOC TAG

```markdown
# APIドキュメント @DOC:SPEC-001:api

これはSPEC-001のAPIドキュメントです。
```

## TAG検証規則

| 規則       | 説明                              | 違反時 |
| ---------- | --------------------------------- | ------ |
| **一意性** | 同じTAGが重複してはいけない       | エラー |
| **完全性** | SPEC→TEST→CODE→DOCすべて存在する必要がある | 警告 |
| **一貫性** | TAG形式の一貫性                   | エラー |
| **追跡可能性** | 相互参照可能                      | 警告 |

## TAGスキャンおよび検証

```bash
# TAG状態照会
moai-adk status

# 特定SPEC TAG詳細照会
moai-adk status --spec SPEC-001

# TAG検証実行
/alfred:3-sync auto SPEC-001

# TAG重複削除
/alfred:tag-dedup --dry-run
/alfred:tag-dedup --apply --backup
```

## <span class="material-icons">library_books</span> 詳細ガイド

- **[TAGタイプ](types.md)** - 各TAGタイプの詳細説明
- **[追跡可能性システム](traceability.md)** - TAGチェーンおよび完全性検証

______________________________________________________________________

**次**: [TAGタイプ](types.md)または[追跡可能性システム](traceability.md)



