---
title: TRUST 5 品質フレームワーク
weight: 70
draft: false
---

MoAI-ADK のすべてのコードが通過しなければならない 5 つの品質原則の詳細ガイドです。

{{< callout type="info" >}}
  **要約:** TRUST 5 は「コードがテストされ、読みやすく、統一され、安全で、追跡可能か?」を検証する自動品質ゲートです。
{{< /callout >}}

## TRUST 5 とは?

TRUST 5 は MoAI-ADK がすべてのコードに適用する **5 つの品質原則** です。AI 生成コードと人間が書いたコードの両方がこれらの基準を満たす必要があります。

日常の比喩で説明すると、建物の完工検査のようなものです。構造安全性、電気配線、水道配管、消防設備、建築許可書をすべて確認してから入居できます。コードも同じです。

| 建物検査 | TRUST 5 | 確認内容 |
|---------|---------|---------|
| 構造安全性 | **T** (Tested) | テストでコードが正しく動作することを確認 |
| 電気/配管設計図 | **R** (Readable) | 他の開発者がコードを理解できるか |
| 建築基準準拠 | **U** (Unified) | プロジェクトのコーディング規則に一致 |
| 消防/セキュリティ設備 | **S** (Secured) | セキュリティ脆弱性なし |
| 許可書類 | **T** (Trackable) | 変更履歴が明確に記録されている |

```mermaid
flowchart TD
    Code["コード記述完了"] --> T1["T: Tested<br>テスト検証"]
    T1 --> R["R: Readable<br>読みやすさ検証"]
    R --> U["U: Unified<br>一貫性検証"]
    U --> S["S: Secured<br>セキュリティ検証"]
    S --> T2["T: Trackable<br>追跡可能性検証"]
    T2 --> Deploy["デプロイ準備完了"]

    T1 -.- T1D["85%+ カバレッジ<br>LSP 0 タイプエラー"]
    R -.- RD["明確な名前<br>LSP 0 リントエラー"]
    U -.- UD["一貫したスタイル<br>LSP 警告 10 未満"]
    S -.- SD["OWASP Top 10<br>LSP 0 セキュリティ警告"]
    T2 -.- T2D["Conventional Commits<br>Issue 追跡"]
```

## T - Tested (テスト済み)

**核心:** すべてのコードはテストで検証されなければなりません。

### 何を確認するか

| 確認項目 | 基準 | 説明 |
|---------|------|------|
| テストカバレッジ | 85% 以上 | すべてのコードの 85% 以上がテストで検証される |
| 特性化テスト | 既存コードを保護 | リファクタリング時に既存動作を保持するテスト |
| LSP タイプエラー | 0 個 | タイプチェックエラーなし |
| LSP 診断エラー | 0 個 | 言語サーバー診断エラーなし |

### なぜ 85% なのか?

100% を要求しない理由があります。

| カバレッジ | 現実的な意味 |
|-----------|-------------|
| 60% 未満 | 主要な機能もテストされていない可能性 |
| 60-84% | 基本的な機能はテストされているが、エッジケースが不足している可能性 |
| **85-95%** | **コアロジックとほとんどのエッジケースが検証済み (推奨)** |
| 95-100% | テスト保守コストが効果を上回り始める |

### モベストプラクティス

```python
def calculate_discount(price: float, discount_rate: float) -> float:
    """割引価格を計算します。

    Args:
        price: 元価格 (0 以上)
        discount_rate: 割引率 (0.0 ~ 1.0)

    Returns:
        割引後の価格

    Raises:
        ValueError: 無効な入力値の場合
    """
    if price < 0:
        raise ValueError("価格は 0 以上である必要があります")
    if not 0 <= discount_rate <= 1:
        raise ValueError("割引率は 0.0 から 1.0 の間である必要があります")
    return price * (1 - discount_rate)

# テストは正常時と例外時の両方を検証
def test_calculate_discount_normal():
    assert calculate_discount(10000, 0.1) == 9000
    assert calculate_discount(5000, 0.5) == 2500
    assert calculate_discount(0, 0.5) == 0

def test_calculate_discount_invalid_price():
    with pytest.raises(ValueError, match="価格は 0"):
        calculate_discount(-1000, 0.1)

def test_calculate_discount_invalid_rate():
    with pytest.raises(ValueError, match="割引率は"):
        calculate_discount(10000, 1.5)
```

---

## R - Readable (読みやすい)

**核心:** コードは明確で理解しやすくなければなりません。

### 何を確認するか

| 確認項目 | 基準 | 説明 |
|---------|------|------|
| 命名規則 | 意図を明示 | 変数、関数、クラス名が明確であること |
| コードコメント | 複雑なロジックを説明 | 「なぜ」を説明するコメント |
| LSP リントエラー | 0 個 | すべてのリンタールールをパス |
| 関数の長さ | 適切なサイズ | 関数が長すぎないこと |

### モベストプラクティス

```python
# 悪い: 名前から何をするかわからない
def calc(d, r):
    return d * (1 - r)

# 良い: 名前を読むだけで役割がわかる
def calculate_discounted_price(original_price: float, discount_rate: float) -> float:
    """元価格から割引率分割引した価格を計算."""
    return original_price * (1 - discount_rate)
```

{{< callout type="info" >}}
  **読みやすさのヒント:** 「6 ヶ月後でも理解できるか?」と自問してください。できない場合は、名前を変更するかコメントを追加してください。
{{< /callout >}}

---

## U - Unified (統一された)

**核心:** プロジェクト全体で一貫したコードスタイルを維持します。

### 何を確認するか

| 確認項目 | 基準 | 説明 |
|---------|------|------|
| コードフォーマット | 自動フォーマッタ適用 | Python: ruff/black、JS: prettier |
| 命名規則 | プロジェクト標準に準拠 | snake_case、camelCase などを混在させない |
| エラーハンドリング | 一貫したパターン | すべての場所で同じエラーハンドリング方式 |
| LSP 警告 | 10 未満 | 言語サーバー警告がしきい値以下 |

### モベストプラクティス

```python
# 統一されたエラーハンドリングパターン
class AppError(Exception):
    """アプリケーション基底エラー"""
    def __init__(self, message: str, code: int = 500):
        self.message = message
        self.code = code

class NotFoundError(AppError):
    """リソースが見つからない"""
    def __init__(self, resource: str, id: str):
        super().__init__(f"{resource} '{id}' が見つかりません", code=404)

class ValidationError(AppError):
    """入力検証失敗"""
    def __init__(self, field: str, reason: str):
        super().__init__(f"'{field}' 検証失敗: {reason}", code=400)

# すべてのサービスで同じパターンを使用
def get_user(user_id: str) -> User:
    user = user_repository.find_by_id(user_id)
    if not user:
        raise NotFoundError("ユーザー", user_id)
    return user
```

---

## S - Secured (安全な)

**核心:** すべてのコードはセキュリティ検証をパスしなければなりません。

### 何を確認するか

| 確認項目 | 基準 | 説明 |
|---------|------|------|
| OWASP Top 10 | 完全準拠 | 最も一般的な Web セキュリティ脆弱性を防止 |
| 依存関係スキャン | 脆弱性のあるパッケージなし | 既知の脆弱性を持つライブラリを使用しない |
| 暗号化ポリシー | 機密データを保護 | パスワード、トークンは暗号化必須 |
| LSP セキュリティ警告 | 0 個 | セキュリティ関連の警告なし |

### 主要なセキュリティチェック

| 脆弱性 | 防止方法 | 例 |
|--------|---------|-----|
| **SQL Injection** | パラメータ化クエリ | `db.execute("SELECT * FROM users WHERE id = %s", (id,))` |
| **XSS** | 出力エスケープ | HTML 出力を自動エスケープ |
| **パスワード露出** | bcrypt ハッシュ化 | `bcrypt.hashpw(password, salt)` |
| **ハードコードされた秘密鍵** | 環境変数 | `os.environ["SECRET_KEY"]` |
| **CSRF** | トークン検証 | 状態変更リクエストに CSRF トークンを含める |

### モベストプラクティス

```python
# 悪い: SQL Injection 脆弱性
def get_user(username: str) -> dict:
    query = f"SELECT * FROM users WHERE username = '{username}'"
    return db.execute(query)

# 良い: パラメータ化クエリで安全
def get_user(username: str) -> dict:
    query = "SELECT * FROM users WHERE username = %s"
    return db.execute(query, (username,))
```

---

## T - Trackable (追跡可能な)

**核心:** すべての変更は明確に追跡可能でなければなりません。

### 何を確認するか

| 確認項目 | 基準 | 説明 |
|---------|------|------|
| コミットメッセージ | Conventional Commits | `feat:`、`fix:`、`refactor:` などの標準形式 |
| Issue リンク | GitHub Issues 参照 | コミットに関連 issue 番号を含める |
| CHANGELOG | 変更ログを維持 | ユーザーに表示される変更を記録 |
| LSP 状態追跡 | 診断履歴を記録 | LSP 状態変化を追跡して回帰検出 |

### Conventional Commits 形式

```bash
# 構造: <type>(<scope>): <description>
# 例:

# 新機能を追加
$ git commit -m "feat(auth): JWT ベースのログイン API を追加"

# バグ修正
$ git commit -m "fix(auth): トークン満期時間計算エラーを修正"

# リファクタリング
$ git commit -m "refactor(auth): 認証ロジックを AuthService に分離"

# セキュリティ改善
$ git commit -m "security(db): パラメータ化クエリで SQL Injection を防止"
```

**コミットタイプ:**

| タイプ | 説明 | 例 |
|--------|------|-----|
| `feat` | 新機能 | `feat(api): ユーザーリスト API を追加` |
| `fix` | バグ修正 | `fix(auth): ログインエラーメッセージを修正` |
| `refactor` | コード改善 (動作変更なし) | `refactor(db): クエリを最適化` |
| `security` | セキュリティ改善 | `security(auth): 秘密鍵を環境変数に` |
| `docs` | ドキュメント変更 | `docs(readme): インストールガイドを更新` |
| `test` | テスト追加/変更 | `test(auth): ログインテストケースを追加` |

---

## LSP 品質ゲート

MoAI-ADK は **LSP** (Language Server Protocol) を使用してコード品質をリアルタイムで検証します。LSP は IDE で赤い下線でエラーを表示するシステムです。

### フェーズ別 LSP 閾値

Plan、Run、Sync フェーズで異なる LSP 基準が適用されます。

| フェーズ | エラー許容 | タイプエラー許容 | リントエラー許容 | 警告許容 | 回帰許容 |
|--------|----------|----------|----------|--------|--------|
| **Plan** | ベースライン取得 | ベースライン取得 | ベースライン取得 | - | - |
| **Run** | 0 個 | 0 個 | 0 個 | - | 不可 |
| **Sync** | 0 個 | - | - | 最大 10 個 | 不可 |

**各フェーズの意味:**

- **Plan フェーズ**: 現在のコードの LSP 状態を「ベースライン」として取得。これが基準値になります。
- **Run フェーズ**: 実装完了時に LSP エラーは 0 必須。ベースラインからエラーが増加しない (回帰なし)。
- **Sync フェーズ**: ドキュメント作成と PR 前に LSP はクリーンである必要。警告は 10 までが許可。

```mermaid
flowchart TD
    P["Plan フェーズ<br>LSP ベースラインを取得"] --> R["Run フェーズ<br>0 エラー、0 タイプエラー、0 リントエラー<br>回帰なし"]
    R --> S["Sync フェーズ<br>0 エラー、警告 10 未満<br>クリーンな LSP 状態"]
    S --> Deploy["デプロイ準備完了"]

    R -.- RCheck{"ベースラインから<br>エラー増加?"}
    RCheck -->|"増加"| Block["ブロック: 回帰を検出"]
    RCheck -->|"同等または減少"| Pass["パス"]
```

## Ralph Engine との統合

**Ralph Engine** は MoAI-ADK の自律品質検証ループです。LSP 診断結果に基づいてコードの問題を自動検出して修正します。

```mermaid
flowchart TD
    A["コード変更"] --> B["LSP 診断を実行"]
    B --> C{"TRUST 5<br>すべての項目パス?"}
    C -->|"すべてパス"| D["検証完了<br>デプロイ準備完了"]
    C -->|"失敗項目あり"| E["Ralph Engine<br>自動修正を試行"]
    E --> F["修正されたコード"]
    F --> B
```

**動作方法:**

1. コード変更時に LSP が診断を実行
2. TRUST 5 基準を満たさない項目がある場合、Ralph Engine が自動修正を試みる
3. 修正後に LSP 診断を再実行してパスを確認
4. パスするまで反復 (最大 3 回のリトライ)

**関連コマンド:**

```bash
# 自動修正を実行
> /moai fix

# 完了まで自動修正を反復
> /moai loop
```

## quality.yaml 設定

`.moai/config/sections/quality.yaml` ファイルで TRUST 5 関連の設定を管理します。

### 主要設定項目

```yaml
constitution:
  # TRUST 5 品質検証を有効化
  enforce_quality: true

  # 目標テストカバレッジ
  test_coverage_target: 85

  # LSP 品質ゲート設定
  lsp_quality_gates:
    enabled: true

    plan:
      require_baseline: true # Plan 開始時にベースラインを取得

    run:
      max_errors: 0 # Run フェーズのエラー許容: 0
      max_type_errors: 0 # タイプエラー許容: 0
      max_lint_errors: 0 # リントエラー許容: 0
      allow_regression: false # ベースラインからの回帰を禁止

    sync:
      max_errors: 0 # Sync フェーズのエラー許容: 0
      max_warnings: 10 # 警告許容: 最大 10
      require_clean_lsp: true # クリーンな LSP 状態を要求

    cache_ttl_seconds: 5 # LSP 診断キャッシュ有効期間
    timeout_seconds: 3 # LSP 診断タイムアウト
```

### 設定のカスタマイズヒント

| 状況 | 調整方法 |
|------|---------|
| プロジェクト初期、テストがほぼない | `test_coverage_target` を 70 に下げて段階的に上げる |
| レガシーコードが多い | `allow_regression` を一時的に true に設定 |
| 厳格なセキュリティが必要 | `max_warnings` を 0 に設定 |

## 実践的な適用: 品質ゲート通過シナリオ

実際の開発で TRUST 5 がどのように適用されるか見ましょう。

### シナリオ: ユーザー検索 API を実装

```bash
# 1. Plan: SPEC を作成 (LSP ベースラインを取得)
> /moai plan "ユーザー検索 API を実装"
```

```bash
# 2. Run: DDD で実装 (TRUST 5 検証)
> /moai run SPEC-SEARCH-001
```

**Run フェーズでの TRUST 5 検証:**

| 項目 | 検証内容 | 結果 |
|------|---------|------|
| **T** (Tested) | テストカバレッジ 85%、タイプエラー 0 個 | 通過 |
| **R** (Readable) | リントエラー 0 個、明確な関数名 | 通過 |
| **U** (Unified) | ruff/black フォーマット適用、LSP 警告 3 個 | 通過 |
| **S** (Secured) | SQL Injection 防止、入力値検証 | 通過 |
| **T** (Trackable) | Conventional Commit 形式、SPEC 参照 | 通過 |

```bash
# 3. Sync: ドキュメント生成と PR (最終 LSP クリーン確認)
> /moai sync SPEC-SEARCH-001
```

**Sync フェーズの最終確認:**

```
LSP 診断結果:
- エラー: 0 個
- タイプエラー: 0 個
- リントエラー: 0 個
- 警告: 3 個 (閾値 10 以下)
- セキュリティ警告: 0 個

TRUST 5 全項目通過: デプロイ準備完了
```

## TRUST 5 一覧

| 原則 | 核心質問 | 自動検証ツール | 基準 |
|------|---------|-------------|------|
| **T** (Tested) | テストで検証されたか? | pytest、LSP タイプ検査 | 85%+ カバレッジ、0 タイプエラー |
| **R** (Readable) | 他の人が読めるか? | ruff、eslint、LSP リント | 0 リントエラー、明確な名前 |
| **U** (Unified) | プロジェクト規則に従うか? | black、prettier、LSP | 一貫したフォーマット、警告 10 未満 |
| **S** (Secured) | セキュリティ脆弱性がないか? | bandit、semgrep、LSP | OWASP 準拠、0 セキュリティ警告 |
| **T** (Trackable) | 変更履歴が追跡可能か? | commitlint、git | Conventional Commits |

## 関連ドキュメント

- [MoAI-ADK とは?](/ja/core-concepts/what-is-moai-adk) — MoAI-ADK の全体構造を理解
- [SPEC ベース開発](/ja/core-concepts/spec-based-dev) — TRUST 5 が適用される Plan フェーズを学ぶ
- [ドメイン駆動開発](/ja/core-concepts/ddd) — TRUST 5 が適用される Run フェーズを学ぶ
