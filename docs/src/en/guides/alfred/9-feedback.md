---
title: /alfred:9-feedback コマンド
description: GitHub Issue自動作成とフィードバック収集のための完全ガイド
lang: ja
---

# /alfred:9-feedback - フィードバックコマンド

`/alfred:9-feedback`はMoAI-ADKのフィードバック収集コマンドで、開発中のバグ報告、機能要求、改善提案をGitHub Issueとして即座に作成します。

## 概要

**目的**: GitHub Issue自動作成とフィードバック収集
**実行時間**: 約30秒
**主要成果**: GitHub Issue、自動ラベル付与、チーム通知

## 基本使用法

```bash
/alfred:9-feedback
```

## 対話型フロー

### ステップ1: Issueタイプ選択

```
Alfred: どのタイプのIssueを作成しますか？

[ ] 🐛 Bug Report - 問題が発生しました
[ ] ✨ Feature Request - 新機能を提案します
[ ] ⚡ Improvement - 既存機能を改善します
[ ] ❓ Question/Discussion - チームに質問します
```

### ステップ2: Issueタイトル入力

```
Alfred: Issueタイトルを入力してください（簡潔明瞭に）:

例: moai-adk update実行中にテンプレート同期エラーが発生
```

### ステップ3: 詳細説明入力（オプション）

```
Alfred: 詳細な説明を入力してください（Enterキーで省略可能）:

例:
moai-adk update実行時に以下のエラーが発生：
- 症状: .claude/ディレクトリ権限エラー
- 環境: macOS 14.2, Python 3.13, moai-adk v0.15.0
- 期待動作: テンプレートが正常に同期されるべき
- 実際動作: Permission deniedエラーで中断
```

### ステップ4: 優先順位選択

```
Alfred: 優先順位レベルを選択してください:

[ ] 🔴 Critical - システムダウン、データ損失、セキュリティ問題
[ ] 🟠 High - 主要機能の故障、深刻な影響
[✓] 🟡 Medium - 一般優先順位（デフォルト）
[ ] 🟢 Low - 軽微な問題、あれば良い
```

### ステップ5: 自動Issue作成

Alfredが自動的に実行：
1. タイプと優先順位による適切なラベル決定
2. 絵文字付きタイトルフォーマット
3. GitHub Issue作成
4. Issue番号とURL返却

## 完全な実行例

### コードレビュー中のバグ発見時

```bash
# コードレビュー中に深刻な問題を発見し、即座に報告したい場合
$ /alfred:9-feedback

Alfred: どのタイプのIssueを作成しますか？
> 🐛 Bug Report

Alfred: Issueタイトルを入力してください
> モバイルでログインボタンクラッシュ発生

Alfred: 詳細な説明を入力してください（Enterキーで省略可能）
> iPhone 15でログインボタンを押すと5秒間フリーズした後、アプリが終了します。
> iOS 17.2, macOS 14.2 Chrome 120でテストしました。
> 期待動作: ログインモーダルが表示されるべき
> 実際動作: 応答なくクラッシュ

Alfred: 優先順位レベルを選択してください
> 🟠 High

✅ GitHub Issue #234 作成完了！

📋 タイトル: 🐛 [BUG] モバイルでログインボタンクラッシュ発生
🟠 優先順位: High
🏷️  ラベル: bug, reported, priority-high
🔗 URL: https://github.com/owner/repo/issues/234

💡 次へ: 作業を続行してください - Issueは追跡されています！
```

### 機能要求時

```bash
$ /alfred:9-feedback

Alfred: どのタイプのIssueを作成しますか？
> ✨ Feature Request

Alfred: Issueタイトルを入力してください
> moai-adk updateに--dry-runオプション追加

Alfred: 詳細な説明を入力してください（Enterキーで省略可能）
> --dry-runオプションを追加して、実際に更新せずにどのような変更が
> 適用されるかを確認できるようにしてください。これにより、更新前の
> 検証が可能になります。

Alfred: 優先順位レベルを選択してください
> 🟡 Medium

✅ GitHub Issue #235 作成完了！

📋 タイトル: ✨ [FEATURE] moai-adk updateに--dry-runオプション追加
🟡 優先順位: Medium
🏷️  ラベル: enhancement, feature-request, priority-medium
🔗 URL: https://github.com/owner/repo/issues/235
```

## 主な機能

### 1. ⚡ 即時作成

- **実行時間**: 約30秒でGitHub Issue作成
- **中断なし**: 開発フローを維持
- **自動化**: 手動作業不要

### 2. 🏷️ 自動ラベリング

| Issueタイプ | 付与されるラベル |
|-----------|----------------|
| 🐛 Bug Report | `bug`, `reported` |
| ✨ Feature Request | `enhancement`, `feature-request` |
| ⚡ Improvement | `enhancement`, `improvement` |
| ❓ Question/Discussion | `question`, `discussion` |

| 優先順位 | 付与されるラベル |
|---------|----------------|
| 🔴 Critical | `priority-critical` |
| 🟠 High | `priority-high` |
| 🟡 Medium | `priority-medium` |
| 🟢 Low | `priority-low` |

### 3. 🎯 優先順位選択

#### Critical (🔴)

- **システムダウン**: サービス完全停止
- **データ損失**: 重要データの喪失
- **セキュリティ問題**: 脆弱性発見

**例**:
```
🔴 [CRITICAL] 本番環境でデータベース接続がすべて失われました
```

#### High (🟠)

- **主要機能故障**: コア機能が動作しない
- **深刻な影響**: 多数のユーザーに影響
- **代替案なし**: 回避策がない

**例**:
```
🟠 [HIGH] ユーザーログイン機能が完全に動作しません
```

#### Medium (🟡)

- **一般優先順位**: 通常の機能要求やバグ
- **影響限定**: 一部のユーザーに影響
- **代替案あり**: 一時的な回避策可能

**例**:
```
🟡 [MEDIUM] APIレスポンス時間が遅い（約2秒）
```

#### Low (🟢)

- **軽微な問題**: 些細な不便さ
- **改善提案**: 機能強化アイデア
- **あれば良い**: なくても問題ない

**例**:
```
🟢 [LOW] ダッシュボードの配色変更提案
```

### 4. 🔗 チーム可視性

- **即時通知**: Issue作成後すぐにチーム全員に通知
- **追跡可能**: GitHubネイティブな追跡機能
- **協業**: コメント、ラベル、マイルストーン管理

## Issueテンプレート

### Bug Reportテンプレート

```markdown
## 🐛 Bug Report

### 現象
<!-- 問題の簡潔な説明 -->

### 再現手順
1. `...` に移動
2. `...` をクリック
3. エラーが発生

### 期待動作
<!-- 期待される動作の簡潔な説明 -->

### 実際動作
<!-- 実際に起こったことの簡潔な説明 -->

### 環境情報
- OS: [e.g. macOS 14.2]
- Pythonバージョン: [e.g. 3.13.0]
- MoAI-ADKバージョン: [e.g. 0.15.0]

### 追加コンテキスト
<!-- 問題に関する追加情報 -->
```

### Feature Requestテンプレート

```markdown
## ✨ Feature Request

### 機能説明
<!-- 機能の簡潔な説明 -->

### 問題解決
<!-- この機能が解決する問題 -->

### 提案ソリューション
<!-- 望ましい解決策の説明 -->

### 代替案
<!-- 考慮した代替ソリューション -->

### 追加コンテキスト
<!-- 機能に関する追加情報 -->
```

## MoAI-ADKワークフローとの統合

### 1. 開発中

```bash
# コーディング中に問題発見
/alfred:9-feedback
→ Issue即時作成
→ 開発継続
```

### 2. コードレビュー

```bash
# レビュー中に改善提案
/alfred:9-feedback
→ Issueとして追跡
→ 後で対応
```

### 3. 計画段階

```bash
# 計画中に質問発生
/alfred:9-feedback
→ チームディスカッション作成
→ 意思決定支援
```

### 4. 同期段階

```bash
# 同期中に発見された問題
/alfred:3-sync
→ 問題検出
/alfred:9-feedback
→ Issue作成
```

## 高度な機能

### カスタムラベル

```bash
# プロジェクト固有のラベル設定
/alfred:9-feedback --custom-labels

# 設定例:
custom_labels:
  bug: ["bug", "needs-triage", "module-auth"]
  feature: ["enhancement", "backlog", "module-ui"]
```

### テンプレート統合

```bash
# 特定テンプレート使用
/alfred:9-feedback --template security-bug

# テンプレート選択
[ ] 🐛 General Bug
[ ] 🔒 Security Issue
[ ] 🚀 Performance Issue
[ ] <span class="material-icons">menu_book</span> Documentation Issue
```

### バッチ作成

```bash
# 複数Issueを一度に作成
/alfred:9-feedback --batch

# 対話型バッチ作成
Issue 1/3: タイトルを入力 > API認証エラー
Issue 2/3: タイトルを入力 > データベース接続タイムアウト
Issue 3/3: タイトルを入力 > フロントエンド表示崩れ

✅ 3個のIssueを作成しました: #456, #457, #458
```

## 設定とカスタマイズ

### GitHub連携設定

```bash
# GitHub CLI認証確認
gh auth status

# リポジトリ設定確認
gh repo view
```

### デフォルト値設定

```json
// .moai/config.json
{
  "feedback": {
    "default_priority": "medium",
    "auto_assign_labels": true,
    "include_environment": true,
    "custom_template": "internal"
  }
}
```

### 通知設定

```bash
# Slack連携
/alfred:9-feedback --notify slack

# メール通知
/alfred:9-feedback --notify email

# チームメンション
/alfred:9-feedback --mention @team-leads
```

## ベストプラクティス

### 1. 良いIssueタイトル

**悪い例**:
```
- 問題です
- 動きません
- 要望
```

**良い例**:
```
- 🐛 [BUG] モバイルでログインボタンクラッシュ発生
- ✨ [FEATURE] moai-adk updateに--dry-runオプション追加
- ⚡ [IMPROVEMENT] APIレスポンスタイム最適化
```

### 2. 詳細な説明

**最小限の情報**:
- 環境情報 (OS, バージョン)
- 再現手順
- 期待動作 vs 実際動作
- エラーメッセージ（ある場合）

**追加情報**:
- スクリーンショット
- ログファイル
- コードスニペット
- 関連Issueへのリンク

### 3. 適切な優先順位選択

```bash
# 優先順位判断基準
🔴 Critical: 製品が使えない、データ損失、セキュリティ
🟠 High: 主要機能が動かない、多くのユーザーに影響
🟡 Medium: 一般的なバグ、機能改善
🟢 Low: 軽微な問題、将来の改善案
```

### 4. 重複回避

```bash
# Issue作成前に重複検索
/alfred:9-feedback
→ Alfredが類似Issueを自動検索
→ 重複していれば既存Issueを表示
→ 必要なら既存Issueにコメント追加
```

## トラブルシューティング

### よくある問題

**GitHub認証エラー**:
```bash
# GitHub CLI再認証
gh auth login

# 権限確認
gh auth status
```

**リポジトリアクセス権限なし**:
```bash
# リポジトリ権限確認
gh repo view

# 権限がない場合、リポジトリ管理者に連絡
```

**ラベルが付与されない**:
```bash
# リポジトリラベル確認
gh label list

# 必要なラベル作成
gh label create priority-high --color "d73a4a"
```

**Issue作成失敗**:
```bash
# ネットワーク接続確認
ping github.com

# GitHubステータス確認
curl https://www.githubstatus.com/api/v2/status.json
```

## 統合と連携

### CI/CD連携

```yaml
# .github/workflows/feedback.yml
name: Process Feedback
on:
  issues:
    types: [opened, labeled]

jobs:
  process-feedback:
    runs-on: ubuntu-latest
    if: contains(github.event.label.name, 'priority-critical')
    steps:
      - name: Notify team
        run: |
          # 緊急Issueチーム通知
          # Slack/Discord/メール通知
```

### プロジェクト管理ツール連携

```bash
# Jira連携
/alfred:9-feedback --jira-integration

# Trello連携
/alfred:9-feedback --trello-integration

# Asana連携
/alfred:9-feedback --asana-integration
```

## 分析とレポート

### フィードバック分析

```bash
# フィードバック統計
/alfred:9-feedback --analytics

# 出力例:
📊 フィードバック分析 (過去30日間)
- 総Issue数: 15個
- バグ報告: 8個 (53%)
- 機能要求: 4個 (27%)
- 改善提案: 3個 (20%)
- 平均解決時間: 2.3日
```

### チーム生産性

```bash
# チーム別フィードバック傾向
/alfred:9-feedback --team-analytics

# 出力例:
👥 チームフィードバック分析
- 最も多くIssueを作成: @開発者A (5個)
- 最も早く対応: @開発者B (平均0.5日)
- 最も複雑なIssue: 認証システム再設計 (15コメント)
```

---

**📚 次のステップ**:
- [プロジェクトガイド](../project/index.md)でプロジェクト管理
- [品質ガイド](../project/config.md)で品質管理
- [デプロイガイド](../project/deploy.md)で本番環境展開