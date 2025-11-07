# フェーズ9: フィードバック - 課題管理とチームコミュニケーション

`/alfred:9-feedback` コマンドは、開発ワークフローを中断することなく、バグ、機能、改善、チームディスカッションのためのGitHub issueを簡潔に作成する方法を提供します。

## 概要

**目的**: 開発環境から構造化されたGitHub issueを即座に作成します。

**コマンド形式**:

```bash
/alfred:9-feedback
```

**所要時間**: 30〜60秒  
**出力**: 適切なラベル、優先度、チーム通知付きのGitHub issue

## クイック課題作成が重要な理由

### 開発ワークフローの中断

**従来の課題作成**:

1. コーディングを中断
2. ブラウザを開く
3. GitHubにナビゲート
4. issueフォームを記入
5. ラベルと優先度を追加
6. issueを提出
7. コーディングに戻る
8. **失われた時間**: issueあたり3〜5分

**Alfredクイック作成**:

1. 単一コマンドを入力
2. 4つの簡単な質問に答える
3. issueが自動的に作成される
4. コーディングを継続
5. **節約された時間**: コンテキスト切り替えが80%削減

### 即座の課題作成の利点

- **バグキャプチャ**: 記憶が新しいうちに問題を報告
- **アイデアの保存**: 機能アイデアを失う前にキャプチャ
- **改善の追跡**: 最適化の機会を即座に文書化
- **チームの可視性**: 重要な問題をチームに即座に通知
- **コンテキストの保持**: 問題を追跡しながら開発フローを維持

## インタラクティブな課題作成プロセス

### ステップ1: 課題タイプの選択

Alfredは明確で構造化されたメニューを提示します:

```
どのタイプの課題を作成しますか?

[ ] バグレポート - 期待通りに動作していない
[ ] 機能リクエスト - 新機能または機能拡張
[ ] 改善 - 既存機能の最適化
[ ] 質問/ディスカッション - チームコラボレーションが必要
```

**選択のガイダンス**:

**バグレポート** - 次の場合に選択:

- コードの動作が期待と一致しない
- エラーまたはクラッシュが発生
- パフォーマンスの問題が観察される
- セキュリティの脆弱性が発見される

**機能リクエスト** - 次の場合に選択:

- 新機能が価値を追加する
- ユーザーエクスペリエンスを改善できる
- 新しい統合が必要
- 製品の機能拡張が構想される

**改善** - 次の場合に選択:

- 既存のコードを最適化できる
- パフォーマンスを向上できる
- コード品質を改善できる
- 技術的負債に対処する必要がある

**質問/ディスカッション** - 次の場合に選択:

- アーキテクチャの決定にチームの意見が必要
- 技術的アプローチについて議論が必要
- 要件の明確化が必要
- ベストプラクティスについて議論が必要

### ステップ2: 課題タイトルの入力

Alfredは明確で説明的なタイトルを求めます:

```
簡潔な課題タイトルを入力してください (最大100文字):

例:
"無効なメール形式でログインAPIが500エラーを返す"
"二要素認証サポートを追加"
"ユーザープロファイル読み込みのデータベースクエリを最適化"
"APIレスポンスにどのキャッシング戦略を使用すべきか?"

タイトル:
```

**タイトルのベストプラクティス**:

- **具体的に**: 「ログイン失敗」→「無効なメール形式でログインAPIが500エラーを返す」
- **コンテキストを含める**: 影響を受けるコンポーネントまたは機能に言及
- **実行可能に保つ**: タイトルは何をすべきかを示唆すべき
- **キーワードを使用**: 検索とフィルタリングに役立つ用語を含める

### ステップ3: 詳細な説明 (オプション)

Alfredはコンテキストを追加する機会を提供します:

```
詳細な説明を追加 (オプション - Enterキーでスキップ):

含めるべき内容:
- 再現手順 (バグの場合)
- 期待される動作と実際の動作
- 環境の詳細
- スクリーンショットまたはエラーメッセージ
- 影響評価

説明:
```

**効果的な説明テンプレート**:

**バグレポートテンプレート**:

```
環境:
- OS: macOS 14.2
- ブラウザ: Chrome 120
- MoAI-ADKバージョン: 0.17.0

再現手順:
1. /auth/loginエンドポイントに移動
2. 無効な形式のメールを送信: "test@"
3. サーバーレスポンスを観察

期待される動作:
明確なエラーメッセージと共に400 Bad Requestを返すべき

実際の動作:
500 Internal Server Errorを返す

エラーメッセージ:
TypeError: Cannot read property 'validate' of undefined

影響:
ユーザーが認証プロセスを完了できない
```

**機能リクエストテンプレート**:

```
問題:
現在、ユーザーは15分ごとに再認証する必要があり、長時間のセッションでは中断が発生します。

提案ソリューション:
拡張トークン有効期限 (7日間) を持つ「ログイン状態を保持」機能を実装。

ユーザーへの利点:
信頼できるデバイスでの摩擦の減少、ユーザーエクスペリエンスの向上。

技術的考慮事項:
- 安全なリフレッシュトークンメカニズムが必要
- ユーザーの同意を得てオプトインにすべき
- セキュリティ基準を維持する必要がある
```

### ステップ4: 優先度の選択

Alfredは課題の優先順位付けを支援します:

```
優先度レベルを選択:

[ ] クリティカル - システムダウン、データ損失、セキュリティ侵害
[ ] 高 - 主要機能が壊れている、重大な影響
[✓] 中 - 重要だがブロッキングではない課題
[ ] 低 - あると良い、軽微な改善
```

**優先度ガイドライン**:

**クリティカル**:

- 本番システムがダウン
- データの破損または損失
- セキュリティの脆弱性
- 法的コンプライアンスの問題
- **アクション**: 即座の対応が必要

**高**:

- コア機能が壊れている
- ユーザーへの重大な影響
- パフォーマンスの低下
- **アクション**: 現在のスプリントで対処

**中**:

- 重要だが致命的ではない問題
- ユーザーエクスペリエンスの改善
- パフォーマンスの最適化
- **アクション**: 次のスプリントで対処

**低**:

- 軽微な改善
- あると良い機能
- ドキュメント更新
- **アクション**: 時間が許すときに対処

## 自動課題生成

情報を提供すると、Alfredは自動的に:

### 1. 課題をフォーマット

````markdown
# [BUG] 無効なメール形式でログインAPIが500エラーを返す

## 優先度
高

## 環境
- **MoAI-ADKバージョン**: 0.17.0
- **オペレーティングシステム**: macOS 14.2
- **ブラウザ**: Chrome 120.0.6099.129
- **Node.jsバージョン**: 20.10.0
- **報告者**: @developer

## 説明

### 再現手順
1. 認証エンドポイントに移動
2. 無効なメール形式でログインリクエストを送信: `test@`
3. サーバーレスポンスを観察

### 期待される動作
APIは明確なバリデーションエラーメッセージと共に400 Bad Requestを返すべき:
```json
{
  "error": "validation_error",
  "message": "Invalid email format"
}
```
````

### 実際の動作

APIは500 Internal Server Errorを返す:

```json
{
  "error": "internal_server_error",
  "message": "An unexpected error occurred"
}
```

### エラー詳細

```
TypeError: Cannot read property 'validate' of undefined
    at EmailValidator.validate (/src/auth/validators.js:45:15)
    at AuthController.login (/src/auth/controller.js:23:28)
    at Layer.handle [as handle_request] (/node_modules/express/lib/router/layer.js:95:5)
```

### 影響

ユーザーはタイプミスのあるメールアドレスを入力した場合、認証プロセスを完了できません。現在の分析によると、これはログイン試行の約15%に影響します。

## 追加のコンテキスト

- 問題は無効なメール形式で一貫して発生
- 問題はv0.17.0のデプロイ後に開始
- コミットa1b2c3dの最近のメールバリデーション変更に関連

## ラベル

bug, authentication, high-priority, backend, v0.17.0

______________________________________________________________________

作成日: 2025-01-15 14:30:00 UTC  
Alfred SuperAgentで生成  
関連SPEC: SPEC:AUTH-001

````

### 2. インテリジェントラベルを適用

Alfredは以下に基づいて関連ラベルを自動的に割り当てます:

**コンテンツ分析**:
```bash
# タイトル/説明で検出されたキーワード
"authentication" → auth, security
"API" → api, backend
"500 error" → bug, server-error
"performance" → performance, optimization
````

**優先度マッピング**:

```bash
クリティカル → priority-critical, urgent
高 → priority-high, needs-attention
中 → priority-medium
低 → priority-low, nice-to-have
```

**コンポーネント検出**:

```bash
"login" → authentication, user-management
"database" → database, backend
"UI" → frontend, user-interface
"API" → api, backend
```

### 3. 課題メタデータを設定

```yaml
# 自動的に適用されるGitHub issueメタデータ
labels:
  - bug
  - authentication
  - high-priority
  - backend
  - v0.17.0

assignees:
  - @backend-team-lead

milestone:
  "Sprint 23 - Q1 2025"

projects:
  - "Authentication System"
  - "Backend Development"
```

### 4. GitHub Issueを作成

AlfredはGitHub CLIを使用してissueを作成します:

```bash
# 同等のAlfred操作
gh issue create \
  --title "[BUG] 無効なメール形式でログインAPIが500エラーを返す" \
  --body "$(cat issue-template.md)" \
  --label bug,authentication,high-priority,backend,v0.17.0 \
  --assignee @backend-team-lead \
  --project "Authentication System"
```

## 開発ワークフローとの統合

### 開発中

**シナリオ1: コーディング中のバグ発見**

```bash
# 機能を実装中にバグを発見
/alfred:9-feedback
→ バグレポート
→ "特殊文字を含むトークンでJWTトークンバリデーションが失敗"
→ [問題の詳細な説明]
→ 高優先度

# Issue #123が即座に作成される
# コンテキストを失うことなくコーディングを継続
```

**シナリオ2: 実装中の機能アイデア**

```bash
# 認証を実装中にアイデアを得る
/alfred:9-feedback
→ 機能リクエスト
→ "セキュリティ強化のためにデバイスフィンガープリンティングを追加"
→ [機能の詳細な説明]
→ 中優先度

# Issue #124が将来の検討のために作成される
# 現在のタスクを継続
```

### コードレビュー中

**シナリオ3: コードレビューの提案**

```bash
# PRレビュー中に改善の機会に気付く
/alfred:9-feedback
→ 改善
→ "ユーザープロファイル読み込みのデータベースクエリを最適化"
→ [具体的な最適化の提案]
→ 中優先度

# Issue #125が作成され、PRにリンクされる
# 中断することなくレビューを継続
```

### テスト中

**シナリオ4: テスト失敗**

```bash
# テストを実行すると予期しない動作が明らかになる
/alfred:9-feedback
→ バグレポート
→ "同時ユーザーセッションの統合テストが失敗"
→ [テスト出力と再現手順]
→ 高優先度

# Issue #126がテスト証拠と共に作成される
# 体系的にデバッグを継続できる
```

## 高度な機能

### 課題テンプレートとカスタマイズ

Alfredはカスタム課題テンプレートをサポートします:

```yaml
# .moai/templates/issue-templates.yml
bug_report:
  title_prefix: "[BUG]"
  required_fields:
    - environment
    - steps_to_reproduce
    - expected_behavior
    - actual_behavior
  optional_fields:
    - screenshots
    - logs
    - additional_context

feature_request:
  title_prefix: "[FEATURE]"
  required_fields:
    - problem_statement
    - proposed_solution
    - user_benefit
  optional_fields:
    - technical_considerations
    - alternatives_considered

improvement:
  title_prefix: "[IMPROVEMENT]"
  required_fields:
    - current_limitation
    - proposed_improvement
    - expected_impact
  optional_fields:
    - implementation_complexity
    - breaking_changes
```

### 一括課題作成

関連する課題の場合、Alfredは複数の課題を作成できます:

```bash
/alfred:9-feedback --bulk
# Alfredが関連する課題の作成をガイドします
# 機能エピックやバグクラスターに便利
```

### 課題テンプレート統合

AlfredはGitHubの課題テンプレートと統合します:

```bash
# 既存のGitHub課題テンプレートを使用
# チーム基準との一貫性を維持
# カスタムテンプレート選択をサポート
```

## チームコラボレーション機能

### 自動割り当て

Alfredは以下に基づいて課題を自動的に割り当てることができます:

**コード所有権**:

```bash
# src/auth/のファイル変更 → @auth-team
# データベース関連の課題 → @database-team
# フロントエンドの課題 → @frontend-team
```

**専門知識マッチング**:

```bash
# セキュリティ課題 → @security-expert
# パフォーマンス課題 → @performance-team
# UI/UX課題 → @design-team
```

**ラウンドロビン割り当て**:

```bash
# チームメンバー間で課題を均等に配分
# 現在の作業負荷と専門知識を考慮
# 休暇と可用性を尊重
```

### チーム通知

Alfredは自動チーム通知を設定できます:

```yaml
# .moai/config/team-notifications.yml
notifications:
  critical_issues:
    - slack: #dev-alerts
    - email: oncall@company.com

  high_priority:
    - slack: #backend-team
    - mention: @team-lead

  feature_requests:
    - slack: #product-team
    - create_project_card: true
```

### スプリント計画統合

```bash
# 課題を現在のスプリントにリンク
# 複雑さを自動的に推定
# スプリント割り当てを提案
# ベロシティへの影響を追跡
```

## 課題の品質とベストプラクティス

### 効果的な課題の作成

**良い課題の特徴**:

- **明確なタイトル**: すぐに理解できる
- **具体的なコンテキスト**: 環境、バージョン、条件
- **再現可能な手順**: 明確な再現手順
- **期待と実際**: 明確な比較
- **影響評価**: ビジネスまたはユーザーへの影響
- **視覚的証拠**: スクリーンショット、ログ、エラーメッセージ

**課題品質チェックリスト**:

```bash
✅ タイトルが説明的で100文字未満
✅ 優先度レベルが影響に適切
✅ 説明に再現手順が含まれる
✅ 期待される動作が明確に記述されている
✅ 実際の動作が文書化されている
✅ 環境の詳細が含まれる
✅ エラーメッセージまたはログが提供されている
✅ ユーザー/システムへの影響が評価されている
✅ 関連するコンポーネントまたは機能が言及されている
✅ ラベルが関連性があり役立つ
```

### 課題トリアージプロセス

Alfredは自動トリアージをサポートします:

```bash
# 自動トリアージルール
if priority == "critical":
    assign to oncall
    notify in #alerts channel

if contains "security":
    assign to security team
    set milestone to "Security Review"

if contains "performance":
    add performance label
    assign to performance team

if links to SPEC:
    add specification label
    link to related project card
```

## 分析とレポート

### 課題メトリクス

Alfredは課題作成パターンを追跡します:

```bash
# 週次課題作成レポート
課題作成分析 (1月15日〜21日の週)

作成された課題: 12
├── バグレポート: 5 (42%)
├── 機能リクエスト: 4 (33%)
├── 改善: 2 (17%)
└── 質問: 1 (8%)

優先度分布:
├── クリティカル: 1 (8%)
├── 高: 3 (25%)
├── 中: 6 (50%)
└── 低: 2 (17%)

平均作成時間: 45秒
節約されたコンテキスト切り替え時間: 約3.5時間
```

### チーム生産性

```bash
# チーム生産性インサイト
チーム生産性メトリクス

開発者あたりの課題:
- @alice: 4課題 (33%)
- @bob: 3課題 (25%)
- @carol: 5課題 (42%)

応答時間:
- 平均初回応答: 2.3時間
- クリティカル課題: 15分
- 高優先度: 1.2時間

解決率:
- 今週: 85% (10/12解決)
- 先週: 78% (14/18解決)
```

## トラブルシューティング

### 一般的な問題

**GitHub CLIが認証されていない**:

```bash
# GitHub CLIを認証
gh auth login

# 認証を確認
gh auth status

# 課題作成を再試行
/alfred:9-feedback
```

**リポジトリ権限**:

```bash
# リポジトリアクセスを確認
gh repo view

# 書き込み権限を確認
gh api repos/:owner/:repo/collaborators/:username

# 必要に応じてアクセスをリクエスト
# リポジトリメンテナーに連絡
```

**ネットワーク接続**:

```bash
# GitHub接続を確認
ping github.com

# APIアクセスを確認
gh api user

# レート制限を確認
gh api rate_limit
```

### エラー処理

**課題作成の失敗**:

```bash
# Alfredは詳細なエラーメッセージを提供します
課題作成失敗: バリデーションエラー

詳細:
- タイトルが長すぎます (125文字、最大100)
- 必須フィールドが不足: steps_to_reproduce
- 無効な優先度: "urgent" (使用: critical, high, medium, low)

問題を修正して再試行するか、--forceフラグを使用して上書き
```

**テンプレートエラー**:

```bash
# カスタムテンプレートを確認
cat .moai/templates/issue-templates.yml

# テンプレート構文を検証
moai-adk validate-templates

# 必要に応じてデフォルトテンプレートにリセット
/alfred:9-feedback --reset-templates
```

## 他のツールとの統合

### プロジェクト管理

**Jira統合**:

```bash
# 対応するJiraチケットを作成
/alfred:9-feedback --jira-integration

# GitHubとJira間で課題ステータスを同期
# 一貫したラベル付けと優先度を維持
```

**Trello統合**:

```bash
# 課題用のTrelloカードを作成
# 適切なボードとリストに追加
# チームメンバーと期限を割り当て
```

**Asana統合**:

```bash
# Asanaタスクを作成
# プロジェクトとセクションに割り当て
# カスタムフィールドと依存関係を設定
```

### コミュニケーションツール

**Slack統合**:

```bash
# Slackチャンネルに課題通知を投稿
# 関連するチームメンバーに@mention
# 課題プレビューとアクションアイテムを含める
```

**Microsoft Teams統合**:

```bash
# Teams会話を作成
# 関連するチャンネルに通知
# 課題の詳細と優先度を含める
```

### モニタリングとアラート

**PagerDuty統合**:

```bash
# クリティカル課題のPagerDutyインシデントを作成
# オンコールエンジニアに通知
# 解決時間を追跡
```

**Datadog統合**:

```bash
# 課題をモニタリングアラートにリンク
# パフォーマンスメトリクスと相関
# システムへの課題の影響を追跡
```

## ベストプラクティスの要約

### 個人向け

1. **即座に問題を報告**: 待たずに、新鮮なうちに課題をキャプチャ
2. **明確なコンテキストを提供**: 環境、手順、期待される動作を含める
3. **適切な優先度を使用**: ユーザーとシステムへの影響を考慮
4. **関連項目をリンク**: SPECs、他の課題、またはPRに接続
5. **フォローアップ**: 課題の進捗を監視し、追加情報を提供

### チーム向け

1. **テンプレートを確立**: 一貫した課題テンプレートを作成
2. **ワークフローを定義**: トリアージと割り当てプロセスを設定
3. **メトリクスを監視**: 課題作成と解決パターンを追跡
4. **品質をレビュー**: 課題の品質と完全性を定期的に評価
5. **継続的改善**: チームフィードバックに基づいてプロセスを改善

### 組織向け

1. **プロセスを標準化**: すべてのリポジトリでAlfredを使用
2. **ツールを統合**: プロジェクト管理とモニタリングツールに接続
3. **分析を追跡**: 課題パターンとチーム生産性を監視
4. **トレーニングを提供**: チームメンバーが効果的な課題作成を理解できるようにする
5. **反復と改善**: 課題管理プロセスを継続的に改善

`/alfred:9-feedback` コマンドは、課題作成を中断的なタスクから開発ワークフローのシームレスな一部に変換し、コーディングフローを維持しながら何も失われないようにします!
