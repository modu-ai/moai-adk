# TAGシステム: 完全トレーサビリティガイド

仕様から実装、テスト、ドキュメントまで完全なトレーサビリティを作成する@TAGシステムをマスターしましょう。このガイドでは、TAGの作成、管理、検証、ベストプラクティスを網羅します。

**最終更新**: 2025-11-06
**オンラインドキュメント**: [TAGシステムガイド](https://adk.mo.ai.kr/guides/specs/tags)
**関連SPEC**: [SPEC-PORTAL-LINK-001](https://adk.mo.ai.kr/specs/PORTAL-LINK-001) - オンラインドキュメントポータル統合

---

## 🌐 オンラインドキュメント統合

このTAGシステムガイドは、https://adk.mo.ai.kr のオンラインドキュメントポータルとシームレスに統合されています。ポータルは以下を提供します:

- **インタラクティブナビゲーション**: TAGタイプ間のクロスリファレンスとリアルタイム検索
- **ライブサンプル**: ライブテスト付きの実働コードサンプル
- **ビジュアルトレーサビリティ**: インタラクティブなTAGチェーン図
- **自動更新**: GitHubリポジトリの変更と同期

### ポータル機能

- **リアルタイムTAG検証**: TAGチェーンの即座のフィードバック
- **影響分析**: TAG関係のビジュアルマッピング
- **カバレッジメトリクス**: ライブ完成統計
- **検索とナビゲーション**: 高度なフィルタリングとリンク

### クイックリンク

- [TAGシステム概要](https://adk.mo.ai.kr/guides/specs/tags#overview)
- [TAGポリシー](https://adk.mo.ai.kr/reference/tags/policy)
- [オンラインサンプル](https://adk.mo.ai.kr/examples/tags)
- [インタラクティブマトリックス](https://adk.mo.ai.kr/matrix/tag-coverage)

## @TAGシステムとは?

@TAGシステムは、一意の識別子を通じてすべてのプロジェクト成果物をリンクするMoAI-ADKのトレーサビリティメカニズムです。開発ライフサイクル全体を通じて、要件、テスト、コード、ドキュメントが接続されたままであることを保証します。

### @TAGが重要な理由

**従来の開発の問題**:

- 「なぜこのコードが書かれたのか?」→ コンテキストの喪失、忘れられた要件
- 「この機能をカバーするテストは?」→ 不完全なテストカバレッジの発見
- 「これのドキュメントはどこ?」→ 散在または古いドキュメント
- 「この要件更新にどのコードを変更する必要があるか?」→ 手動の影響分析

**@TAGシステムの解決策**:

- **完全なトレーサビリティ**: すべての成果物がそのソースにリンクされている
- **影響分析**: 影響を受けるコードの即座の識別
- **生きたドキュメント**: ドキュメントがコードと同期したまま
- **品質保証**: 孤児コードや欠落したテストがない

### @TAGチェーンの概念

```
@SPEC:DOMAIN-001 (要件)
    ↓ 何を構築するかを定義
@TEST:DOMAIN-001 (テスト)
    ↓ 実装を検証
@CODE:DOMAIN-001:SUBTYPE (実装)
    ↓ ソリューションを作成
@DOC:DOMAIN-001 (ドキュメント)
    ↓ ソリューションを説明
```

## @TAG形式と構造

### 基本形式

**標準形式**: `@TYPE:DOMAIN-ID[:SUBTYPE]`

**コンポーネント**:

- **`@`**: TAG インジケーター (必須)
- **`TYPE`**: 成果物タイプ (SPEC, TEST, CODE, DOC)
- **`DOMAIN`**: 機能領域 (AUTH, USER, API等)
- **`ID`**: 連番 (001, 002, 003...)
- **`SUBTYPE`**: オプショナルなサブ分類 (MODEL, SERVICE, API等)

### タイプ定義

| タイプ | 目的 | 例 |
|--------|------|-----|
| **SPEC** | 要件と仕様 | `@SPEC:AUTH-001` |
| **TEST** | テストケースとテストスイート | `@TEST:AUTH-001` |
| **CODE** | 実装コード | `@CODE:AUTH-001:SERVICE` |
| **DOC** | ドキュメントとガイド | `@DOC:AUTH-001` |

### ドメインカテゴリ

| ドメイン | 説明 | 例 |
|---------|------|-----|
| **AUTH** | 認証と承認 | `@SPEC:AUTH-001` |
| **USER** | ユーザー管理とプロフィール | `@SPEC:USER-001` |
| **API** | REST APIとインターフェース | `@SPEC:API-001` |
| **DB** | データベースと永続化 | `@SPEC:DB-001` |
| **UI** | ユーザーインターフェースとコンポーネント | `@SPEC:UI-001` |
| **SEC** | セキュリティとコンプライアンス | `@SPEC:SEC-001` |
| **PERF** | パフォーマンスと最適化 | `@SPEC:PERF-001` |
| **INT** | 統合と外部システム | `@SPEC:INT-001` |
| **CONFIG** | 設定と設定 | `@SPEC:CONFIG-001` |

### サブタイプ分類

#### CODEサブタイプ

| サブタイプ | 使用するタイミング | 例 |
|-----------|-------------------|-----|
| **MODEL** | データモデル、スキーマ、クラス | `@CODE:USER-001:MODEL` |
| **SERVICE** | ビジネスロジック、サービス | `@CODE:AUTH-001:SERVICE` |
| **API** | HTTPエンドポイント、コントローラー | `@CODE:API-001:ENDPOINT` |
| **REPO** | リポジトリパターン、データアクセス | `@CODE:DB-001:REPO` |
| **UTILS** | ユーティリティ関数、ヘルパー | `@CODE:AUTH-001:UTILS` |
| **CONFIG** | 設定クラス | `@CODE:CONFIG-001:SETTINGS` |
| **MIDDLEWARE** | ミドルウェア、インターセプター | `@CODE:API-001:MIDDLEWARE` |
| **VALIDATOR** | 検証ロジック | `@CODE:USER-001:VALIDATOR` |

#### TESTサブタイプ

| サブタイプ | 使用するタイミング | 例 |
|-----------|-------------------|-----|
| **UNIT** | ユニットテスト | `@TEST:AUTH-001:UNIT` |
| **INTEGRATION** | 統合テスト | `@TEST:API-001:INTEGRATION` |
| **E2E** | エンドツーエンドテスト | `@TEST:USER-001:E2E` |
| **PERFORMANCE** | パフォーマンステスト | `@TEST:API-001:PERF` |
| **SECURITY** | セキュリティテスト | `@TEST:AUTH-001:SECURITY` |

#### DOCサブタイプ

| サブタイプ | 使用するタイミング | 例 |
|-----------|-------------------|-----|
| **API** | APIドキュメント | `@DOC:API-001:API` |
| **GUIDE** | ユーザーガイド、チュートリアル | `@DOC:USER-001:GUIDE` |
| **REFERENCE** | 技術リファレンス | `@DOC:AUTH-001:REFERENCE` |
| **DEPLOYMENT** | デプロイメントガイド | `@DOC:INT-001:DEPLOYMENT` |

## @TAGの作成と割り当て

### 自動TAG割り当て

Alfredは開発ワークフロー中に自動的にTAGを割り当てます:

```bash
# フェーズ1: SPEC作成
/alfred:1-plan "ユーザー認証システム"
# Alfredが割り当て: @SPEC:AUTH-001

# フェーズ2: TDD実装
/alfred:2-run AUTH-001
# Alfredが割り当て: @TEST:AUTH-001, @CODE:AUTH-001:*

# フェーズ3: ドキュメント同期
/alfred:3-sync
# Alfredが割り当て: @DOC:AUTH-001
```

### 手動TAG割り当て

ファイルを手動で作成する場合:

```python
# src/auth/service.py
# @CODE:AUTH-001:SERVICE | SPEC: SPEC-AUTH-001.md | TEST: tests/test_auth.py

class AuthService:
    """認証サービス実装"""
    pass
```

```python
# tests/test_auth.py
# @TEST:AUTH-001 | SPEC: SPEC-AUTH-001.md

def test_login_with_valid_credentials():
    """有効な認証情報でのユーザー認証をテスト"""
    pass
```

### TAG割り当てのベストプラクティス

#### 1. 一貫性

**✅ 良い**:

```python
# すべての関連コードが同じDOMAIN-IDを使用
@CODE:AUTH-001:MODEL
@CODE:AUTH-001:SERVICE
@CODE:AUTH-001:API
@CODE:AUTH-001:UTILS
```

**❌ 悪い**:

```python
# 一貫性のないDOMAIN-ID使用
@CODE:AUTH-001:MODEL
@CODE:AUTH-002:SERVICE  # 間違ったID
@CODE:USER-001:API     # 間違ったドメイン
```

#### 2. 具体性

**✅ 良い**:

```python
# 明確な構成のための具体的なサブタイプ
@CODE:AUTH-001:SERVICE
@CODE:AUTH-001:MODEL
@CODE:AUTH-001:VALIDATOR
```

**❌ 悪い**:

```python
# あまりにも一般的 - ファイルの目的を示さない
@CODE:AUTH-001
@CODE:AUTH-001
@CODE:AUTH-001
```

#### 3. トレーサビリティリンク

**✅ 良い**:

```python
# 関連する成果物へのリンクを含む
# @CODE:AUTH-001:SERVICE | SPEC: SPEC-AUTH-001.md | TEST: tests/test_auth.py
```

**❌ 悪い**:

```python
# トレーサビリティ情報の欠落
# @CODE:AUTH-001:SERVICE
```

## @TAGチェーン管理

### 完全なチェーンの例

```markdown
# SPECドキュメント
# .moai/specs/SPEC-AUTH-001/spec.md
# @SPEC:EX-AUTH-001: ユーザー認証システム

## 要件
- システムはユーザー認証を提供しなければならない...
```

```python
# テストファイル
# tests/test_auth.py
# @TEST:EX-AUTH-001 | SPEC: SPEC-AUTH-001.md

def test_login_with_valid_credentials():
    """有効な認証情報での認証をテスト"""
    pass
```

```python
# 実装ファイル
# src/auth/models.py
# @CODE:EX-AUTH-001:MODEL | SPEC: SPEC-AUTH-001.md | TEST: tests/test_auth.py

class User:
    """ユーザーデータモデル"""
    pass

# src/auth/service.py
# @CODE:EX-AUTH-001:SERVICE | SPEC: SPEC-AUTH-001.md | TEST: tests/test_auth.py

class AuthService:
    """認証ビジネスロジック"""
    pass

# src/auth/api.py
# @CODE:EX-AUTH-001:API | SPEC: SPEC-AUTH-001.md | TEST: tests/test_auth.py

@app.post("/auth/login")
def login():
    """ログインエンドポイント"""
    pass
```

```markdown
# ドキュメント
# docs/api/auth.md
# @DOC:EX-AUTH-001 | SPEC: SPEC-AUTH-001.md

# 認証APIドキュメント
...
```

### チェーン検証

AlfredはGPT-5 Pro分析を使用してTAGチェーンを自動的に検証します:

```bash
/alfred:3-sync --validation-mode=gpt5-pro
```

**拡張出力例**:

```
🔍 TAGチェーン検証レポート (GPT-5 Pro強化)

✅ 完全なチェーン: AUTH-001
   @SPEC:AUTH-001 → .moai/specs/SPEC-AUTH-001/spec.md
   @TEST:AUTH-001 → tests/test_auth.py (5個のテスト関数)
   @CODE:AUTH-001:MODEL → src/auth/models.py (2個のクラス)
   @CODE:AUTH-001:SERVICE → src/auth/service.py (1個のクラス)
   @CODE:AUTH-001:API → src/auth/api.py (1個のエンドポイント)
   @DOC:AUTH-001 → docs/api/auth.md (完全なAPIドキュメント)

📊 チェーン整合性: 100%
🔗 孤児TAG: 0
⚠️ 欠落リファレンス: 0
🎯 品質スコア: 95/100
💡 AI推奨: 3個の最適化提案

📈 オンラインポータル統合: https://adk.mo.ai.kr/matrix/tag-coverage
```

### 孤児TAG検出 (強化版)

AlfredはAIによる提案を使用して孤児TAGを識別し修正を支援します:

```bash
⚠️ 強化された孤児TAGが検出されました:

src/auth/validators.py に孤児 @CODE:AUTH-001:VALIDATOR
   @TEST:AUTH-001:VALIDATOR が欠落
   AI推奨: エッジケースカバレッジを持つユニットテストを作成
   影響: 中 - コード品質メトリクスに影響
   推定作業量: 2-3時間

tests/test_auth_advanced.py に孤児 @TEST:AUTH-002
   @SPEC:AUTH-002 が欠落
   AI推奨: 受け入れ基準を持つ仕様を作成
   影響: 高 - 要件トレーサビリティギャップ
   推定作業量: 4-6時間

🛠️ AIによる自動修正オプション:
[1] @CODE:AUTH-001:VALIDATOR の完全なテストスイートを生成
[2] GPT-5拡張テンプレートで仕様を作成
[3] AI提案付きの手動レビュー
[4] 警告を抑制 (推奨されません)
```

## @TAG検索とナビゲーション

### 関連成果物の検索

#### 基本検索

```bash
# AUTH-001のすべての成果物を検索
rg '@(SPEC|TEST|CODE|DOC):AUTH-001' -n

# すべてのSPECを検索
rg '@SPEC:' -n

# すべてのCODEファイルを検索
rg '@CODE:' -n
```

#### 高度な検索パターン

```bash
# SPECと関連テストを検索
rg '@SPEC:AUTH-001' -A 5 -B 5

# 孤児CODE (マッチするTESTなし) を検索
rg '@CODE:AUTH-001' --files-with-matches | while read file; do
  if ! rg -q "@TEST:AUTH-001" "$(dirname $file)/test*"; then
    echo "孤児CODE: $file"
  fi
done

# ドメインのすべてのチェーンを検索
rg '@(SPEC|TEST|CODE|DOC):AUTH-\d+' -n
```

#### 影響分析

```bash
# SPEC変更によって影響を受けるすべてを検索
rg '@SPEC:AUTH-001' -n
# → 表示: spec, tests, code, documentation

# 機能のテストカバレッジを検索
rg '@TEST:AUTH-001' -n
# → AUTH-001をカバーするすべてのテストファイルを表示

# 実装ステータスを検索
rg '@CODE:AUTH-001' -n
# → すべての実装ファイルを表示
```

## @TAGベストプラクティス (GPT-5 Pro強化)

### 1. 一貫したフォーマット

**標準形式を使用**:

```python
# ✅ 正しい形式
@CODE:AUTH-001:SERVICE | SPEC: SPEC-AUTH-001.md | TEST: tests/test_auth.py

# ❌ 不正確な形式
@code:auth-001:service  # 間違った大文字小文字
@CODE:auth-1:SERVICE    # 間違った形式
@CODE:AUTH-001         # サブタイプの欠落
```

### 2. 完全なトレーサビリティ (強化版)

**すべての関連成果物をリンク**:

```python
# ✅ AI強化による完全なトレーサビリティ
# @CODE:AUTH-001:SERVICE | SPEC: SPEC-AUTH-001.md | TEST: tests/test_auth.py
# AI-MONITORING: 品質スコア 95/100 | 最終検証: 2025-11-06
# ONLINE-PORTAL: https://adk.mo.ai.kr/trace/AUTH-001

# ❌ リンクの欠落
# @CODE:AUTH-001:SERVICE
```

### 3. 論理的な構成 (AI最適化)

**AI提案による関連コードのグループ化**:

```python
# ✅ AI推奨の論理的グループ化
src/
├── auth/
│   ├── models.py      # @CODE:AUTH-001:MODEL | AI: 最適な構造検出
│   ├── service.py     # @CODE:AUTH-001:SERVICE | AI: ビジネスロジック分離
│   ├── api.py         # @CODE:AUTH-001:API | AI: RESTful設計パターン
│   └── utils.py       # @CODE:AUTH-001:UTILS | AI: 再利用可能ユーティリティ

# ❌ ランダムな構成 (AI検出問題)
src/
├── models.py         # @CODE:AUTH-001:MODEL | AI: 散在コンポーネント検出
├── auth_service.py   # @CODE:AUTH-001:SERVICE | AI: 混合責任
├── login_api.py      # @CODE:AUTH-001:API | AI: 一貫性のない命名
└── helpers.py        # @CODE:AUTH-001:UTILS | AI: 未分類ユーティリティ
```

### 4. 適切な粒度 (AI支援)

**AI分析による適切なサイズのコンポーネント**:

```python
# ✅ AI検証された適切な粒度
@CODE:AUTH-001:MODEL     # User, Sessionモデル | AI: 単一責任
@CODE:AUTH-001:SERVICE    # AuthServiceクラス | AI: ビジネスロジック封入
@CODE:AUTH-001:API        # ログインエンドポイント | AI: RESTful原則

# ⚠️ AI検出された過剰な粒度化
@CODE:AUTH-001:MODEL:USER     # Userモデルのみ | AI: 統合を検討
@CODE:AUTH-001:MODEL:SESSION   # Sessionモデルのみ | AI: 冗長な抽象化

# ⚠️ AI検出された過剰な広範化
@CODE:AUTH-001               # すべてが1ファイル | AI: SRP違反
```

### 5. 定期的なメンテナンス (AIによる)

**AI支援によるチェーンの更新保持**:

```bash
# AI強化検証
/alfred:3-sync --ai-mode --auto-suggestions

# AIによる手動チェック
rg '@(SPEC|TEST|CODE|DOC):' -n | sort | uniq -c | ai-validate-tags

# AI支援孤児検出
moai-adk find-orphans --ai-analysis --recommendations

# AI最適化TAGクリーンアップ
moai-adk optimize-tags --gpt5-enhanced --quality-metrics
```

### 6. オンラインポータル統合

**ポータル同期の維持**:

```bash
# ポータル同期検証
/alfred:3-sync --portal-sync --validate-links

# ポータル互換レポートの生成
moai-adk portal-report --format=web --interactive-matrix

# ポータル用AI最適化TAG更新
moai-adk update-portal-tags --ai-enhanced --real-time
```

### 7. AI強化ベストプラクティス

**TAG最適化にGPT-5 Proを活用**:

```python
# AIによるTAG生成提案
# @CODE:USER-001:PROFILE | AI: USER-001:PROFILE_MODEL, USER-001:PROFILE_CONTROLLERを検討
# AI-RISK-ASSESSMENT: 低複雑性、高再利用可能性ポテンシャル
# AI-RECOMMENDATION: MODELとCONTROLLERサブタイプに分割

# AIによるテストカバレッジ最適化
# @TEST:USER-001 | AI: 現在のカバレッジ75%、追加エッジケースを推奨
# AI-SUGGESTED-TESTS: [negative_cases, boundary_conditions, integration_scenarios]
```

### 8. 品質メトリクス (AI追跡)

**AIによる品質メトリクスの維持**:

```bash
# 包括的品質レポートの生成
moai-adk tag-quality --ai-analysis --trend-tracking --portal-integration

# AI最適化品質しきい値
# - チェーン完全性: >90% (AI推奨)
# - 孤児検出: 0 (AI強制)
# - 品質スコア: >85/100 (AI計算)
# - ポータル同期: 100% (AI検証)

# AIによる品質改善提案
moai-adk quality-insights --actionable-recommendations --priority-scoring
```

## 異なるファイルタイプの@TAG

### Pythonファイル

```python
# src/auth/service.py
# @CODE:AUTH-001:SERVICE | SPEC: SPEC-AUTH-001.md | TEST: tests/test_auth.py

"""
認証サービス実装。

このファイルには、パスワード検証、トークン生成、
セッション管理を含むユーザー認証のビジネスロジックが含まれています。

関連ファイル:
- モデル: @CODE:AUTH-001:MODEL
- APIエンドポイント: @CODE:AUTH-001:API
- テスト: @TEST:AUTH-001
"""

class AuthService:
    """@CODE:AUTH-001:SERVICE - メイン認証サービス"""

    def authenticate(self, credentials):
        """ユーザー認証情報を認証"""
        # 実装はここ
        pass
```

### JavaScript/TypeScriptファイル

```javascript
// src/auth/service.js
// @CODE:AUTH-001:SERVICE | SPEC: SPEC-AUTH-001.md | TEST: tests/auth.test.js

/**
 * 認証サービス
 *
 * ユーザー認証、トークン管理、セッション処理を扱います。
 *
 * 関連ファイル:
 * - モデル: @CODE:AUTH-001:MODEL
 * - APIルート: @CODE:AUTH-001:API
 * - テスト: @TEST:AUTH-001
 */

class AuthService {
  /**
   * ユーザー認証情報を認証
   * @CODE:AUTH-001:SERVICE:METHOD
   */
  async authenticate(credentials) {
    // 実装
  }
}
```

### テストファイル

```python
# tests/test_auth.py
# @TEST:AUTH-001 | SPEC: SPEC-AUTH-001.md

"""
認証システムテスト。

テストのカバー範囲:
- 有効な認証情報でのログイン
- 無効な認証情報でのログイン
- トークン生成と検証
- セッション管理

関連ファイル:
- 仕様: @SPEC:AUTH-001
- 実装: @CODE:AUTH-001:*
- ドキュメント: @DOC:AUTH-001
"""

class TestAuthService:
    """@CODE:AUTH-001:SERVICE のテストケース"""

    def test_login_with_valid_credentials(self):
        """テスト: @SPEC:AUTH-001 - 有効な認証情報は認証すべき"""
        # テスト実装
        pass
```

### ドキュメントファイル

```markdown
# docs/api/auth.md
# @DOC:AUTH-001 | SPEC: SPEC-AUTH-001.md

# 認証APIドキュメント

このドキュメントは、リクエスト/レスポンス形式、認証方法、
セキュリティ考慮事項を含む認証APIエンドポイントを説明します。

## 関連成果物
- 仕様: @SPEC:AUTH-001
- テスト: @TEST:AUTH-001
- 実装: @CODE:AUTH-001:*
```

## @TAG自動化とツール (強化版)

### Gitフック (AI強化)

AIによる分析を使用したGitフックでの自動TAG検証:

```bash
#!/bin/sh
# .git/hooks/pre-commit

echo "🔍 AI強化によるTAGチェーン検証中..."

# TAGの欠落をチェック
missing_tags=$(rg -L '@(SPEC|TEST|CODE|DOC):' --files-with-matching src/ tests/ docs/)

if [ -n "$missing_tags" ]; then
    echo "❌ TAGが欠落しているファイル:"
    echo "$missing_tags"
    echo "🤖 AI提案: /alfred:3-sync --auto-add-tags を実行"
    exit 1
fi

# 強化されたAIによる孤児検出
echo "🔍 AI強化孤児検出を実行中..."
orphans=$(moai-adk find-orphans --ai-analysis --impact-assessment)

if [ -n "$orphans" ]; then
    echo "⚠️ AI強化孤児TAGが検出されました:"
    echo "$orphans"
    echo "💡 AI推奨:"
    echo "   - 高影響: 欠落している仕様の作成を検討"
    echo "   - 中影響: AIでテストテンプレートを生成"
    echo "   - 低影響: /alfred:3-sync --auto-fix で自動修正を使用"
fi

# AIによる品質検証
echo "🤖 AI品質評価を実行中..."
quality_score=$(moai-adk tag-quality --quick-scan --ai-powered)

if [ "$quality_score" -lt 85 ]; then
    echo "⚠️ 品質スコアがしきい値未満: $quality_score/100"
    echo "💡 AI提案: moai-adk quality-improve --ai-mode を実行"
fi

echo "✅ AI強化TAG検証に合格"
echo "📊 ポータル同期ステータス: https://adk.mo.ai.kr/sync/status"
```

### CI/CD統合 (ポータル強化)

ポータル統合を使用したGitHub Actionsワークフロー:

```yaml
# .github/workflows/tag-validation.yml
name: TAGチェーン検証 & ポータル同期

on: [push, pull_request]

jobs:
  validate-tags:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3

    - name: Pythonセットアップ
      uses: actions/setup-python@v4
      with:
        python-version: '3.13'

    - name: MoAI-ADKインストール
      run: pip install moai-adk

    - name: 強化TAG検証
      run: |
        moai-adk validate-tags --ai-mode --gpt5-enhanced
        moai-adk check-orphans --impact-analysis --ai-suggestions

    - name: ポータル同期検証
      run: |
        moai-adk portal-sync --validate --real-time-check
        moai-adk generate-portal-report --format=web

    - name: 包括的レポート生成
      run: |
        moai-adk tag-report --format=html --ai-analysis --portal-integration > tag-report.html

    - name: TAGレポートのアップロード
      uses: actions/upload-artifact@v3
      with:
        name: tag-report-portal
        path: tag-report.html
        retention-days: 30

    - name: ポータルステータス更新
      run: |
        moai-adk portal-status --update --commit-hash=${{ github.sha }}
        echo "ポータル更新: https://adk.mo.ai.kr/commits/${{ github.sha }}"
```

## トラブルシューティング@TAG問題

### よくある問題

#### 1. TAGの欠落

**症状**:

- 検索結果にファイルが表示されない
- 不完全なトレーサビリティチェーン

**解決策**:

```bash
# TAGのないファイルを検索
find src/ tests/ docs/ -type f \( -name "*.py" -o -name "*.js" -o -name "*.md" \) -exec grep -L '@(SPEC|TEST|CODE|DOC):' {} \;

# 欠落したTAGを手動で追加
# またはAlfredを使用して自動的に修正
/alfred:3-sync --fix-tags
```

#### 2. 不正確なTAG形式

**症状**:

- TAGが認識されない
- 検証エラー

**よくある形式エラー**:

```python
# ❌ 間違った大文字小文字
@code:auth-001:service

# ❌ 間違った形式
@CODE-AUTH-001-SERVICE

# ❌ 欠落部分
@CODE:AUTH-001
```

**解決策**:

```bash
# 形式問題を検索して修正
rg '@[a-zA-Z]+:[a-zA-Z]+-\d+' --files-with-matching

# 自動修正を使用
moai-adk fix-tag-format
```

## 高度な@TAG技術

### 1. 階層型TAG

複雑な機能には階層関係を使用:

```python
# コア認証
@SPEC:AUTH-001
@CODE:AUTH-001:SERVICE

# パスワードリセット (認証に関連)
@SPEC:AUTH-002
@CODE:AUTH-002:SERVICE

# 関連機能をリンク
# AUTH-002 SPECで:
## 依存関係
- @SPEC:AUTH-001 (コア認証システム)
```

### 2. クロスドメインTAG

複数のドメインにまたがる機能の場合:

```python
# 認証を持つユーザープロフィール
@CODE:USER-001:PROFILE | SPEC: SPEC-USER-001.md, SPEC-AUTH-001.md

# セキュリティ懸念を持つAPI
@CODE:API-001:ENDPOINT | SPEC: SPEC-API-001.md, SPEC-SEC-001.md
```

## まとめ (GPT-5 Pro統合強化)

@TAGシステムは、GPT-5 Proインテリジェンスとオンラインポータル統合によって強化された、MoAI-ADKのトレーサビリティと品質保証のバックボーンです。仕様から実装、テスト、ドキュメントまで完全なチェーンを維持することで、以下のような開発環境を作成します:

### コアメリット

- **🎯 何も失われない** - AIによる検証ですべてのコードが要件にトレース可能
- **⚡ 影響分析が即座** - AI影響評価で要件が進化したときに正確に変更内容を把握
- **🛡️ 品質が保証される** - AI品質監視により孤児コードや欠落したテストがない
- **📚 ドキュメントが最新** - AIメンテナンスによる自動同期でドリフトを防止

### GPT-5 Pro強化機能

- **🤖 AIによる検証** - インテリジェントな提案を使用したリアルタイムTAG検証
- **📊 品質メトリクス** - 実行可能な推奨事項を持つAI計算品質スコア
- **🌐 ポータル統合** - オンラインドキュメントポータルとのシームレスな同期
- **🔍 高度な分析** - 継続的改善のためのAIによるインサイト

### オンラインポータル統合のメリット

- **🌐 インタラクティブナビゲーション** - クロスリファレンスとリアルタイム検索機能
- **📈 ライブカバレッジメトリクス** - 動的TAGチェーン完成統計
- **🎨 ビジュアルトレーサビリティ** - インタラクティブなTAGチェーン図とマッピング
- **🔄 自動更新** - GitHubリポジトリの変更と同期

### 始め方

1. **オンラインガイドを読む**: [TAGシステム概要](https://adk.mo.ai.kr/guides/specs/tags)
2. **インタラクティブマトリックスを探索**: [TAGカバレッジマトリックス](https://adk.mo.ai.kr/matrix/tag-coverage)
3. **AIによる検証を試す**: `/alfred:3-sync --ai-mode --auto-suggestions`
4. **ライブサンプルにアクセス**: [オンラインTAGサンプル](https://adk.mo.ai.kr/examples/tags)

### 品質しきい値 (AI推奨)

- **チェーン完全性**: >90% (AI強制)
- **品質スコア**: >85/100 (AI計算)
- **ポータル同期**: 100% (AI検証)
- **孤児検出**: 0 (AI監視)

### 継続的改善

AI支援で@TAGシステムをマスターすれば、ソフトウェア開発プロセスで新しいレベルの自信とコントロールを体験できます! システムはGPT-5 Pro統合により継続的に学習し改善され、開発ワークフローがAIによるソフトウェアエンジニアリングの最先端に留まることを保証します。

**今日からAI強化TAGの旅を始めましょう** - [チュートリアルを開始](https://adk.mo.ai.kr/tutorials/tag-system) 🚀

---

## 追加リソース

### オンラインドキュメント

- [TAGシステムガイド](https://adk.mo.ai.kr/guides/specs/tags) - ライブサンプル付きインタラクティブガイド
- [TAGポリシーリファレンス](https://adk.mo.ai.kr/reference/tags/policy) - 詳細ポリシードキュメント
- [TAGカバレッジマトリックス](https://adk.mo.ai.kr/matrix/tag-coverage) - ライブカバレッジ統計
- [ポータルステータスダッシュボード](https://adk.mo.ai.kr/dashboard/status) - リアルタイムシステムステータス

### AI強化ツール

- **TAG AIアシスタント**: `/alfred:tag-ai --help`
- **品質アナライザー**: `moai-adk quality --ai-mode`
- **ポータル同期ツール**: `moai-adk portal-sync --ai-enhanced`
- **レポートジェネレーター**: `moai-adk report --ai-analysis --portal`

### コミュニティとサポート

- **GitHub Issues**: [TAGシステムバグ](https://github.com/modu-ai/moai-adk/issues)
- **ディスカッション**: [TAGシステムコミュニティ](https://github.com/modu-ai/moai-adk/discussions)
- **Discord**: [MoAIコミュニティ](https://discord.gg/moai)
- **ポータルフィードバック**: [オンラインフィードバック](https://adk.mo.ai.kr/feedback)

**最終更新**: 2025-11-06
**バージョン**: v0.17.0
**AIモデル**: GPT-5 Pro統合
**ポータル**: https://adk.mo.ai.kr ⭐
