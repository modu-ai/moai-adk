# クイックスタートガイド

**@DOC:QUICK-START-001** | **最終更新**: 2025-11-05 | **所要時間**: 5分

______________________________________________________________________

## 🎯 目標

このガイドを通じて、MoAI-ADKオンライン文書システムを完璧に設定し実行する方法を学びます。

______________________________________________________________________

## 🚀 1段階: システム要件

### 必須要件

- **Python**: 3.13以上
- **UV**: 最新バージョン推奨
- **Git**: 最新バージョン
- **ブラウザ**: Chrome 90+, Firefox 88+, Safari 14+, Edge 90+

### 選択的要件

- **Vercel CLI**: 自動デプロイのための選択的ツール
- **Node.js**: v18+ (一部ビルドツールに必要)

______________________________________________________________________

## ⚡ 2段階: プロジェクト設定 (30秒)

### UVインストール

```bash
# macOS/Linux
curl -LsSf https://astral.sh/uv/install.sh | sh

# Windows
powershell -c "irm https://astral.sh/uv/install.ps1 | iex"

# インストール確認
uv --version
```

### プロジェクトのクローンと設定

```bash
# 1. プロジェクトのクローン
git clone https://github.com/moai-adk/MoAI-ADK.git
cd MoAI-ADK

# 2. 依存関係のインストール (自動)
uv sync

# 3. 開発サーバーの実行
uv run dev
```

### クイック確認

```bash
# サーバー状態の確認
curl http://127.0.0.1:8080

# ビルド状態の確認
uv run build
```

______________________________________________________________________

## 🎨 3段階: ドキュメントシステムの構築

### MkDocs設定

```bash
# MkDocs基本設定の確認
uv run mkdocs --help

# プロジェクト構造の生成
mkdir -p docs/{getting-started,alfred,commands,development,advanced,api,contributing}

# テーマ設定
uv run mkdocs new .
```

### 多言語設定の追加

```yaml
# mkdocs.yml
site_name: MoAI-ADK Documentation
nav:
  - ホーム: index.md
  - クイックスタート: getting-started/
  - Alfred: alfred/
  - コマンド: commands/
  - 開発: development/
  - 高度な機能: advanced/
  - API: api/
  - 貢献: contributing/

theme:
  name: material
  language: ja
  palette:
    - media: "(prefers-color-scheme: light)"
      scheme: default
      primary: blue
      accent: indigo
    - media: "(prefers-color-scheme: dark)"
      scheme: slate
      primary: blue
      accent: indigo
```

______________________________________________________________________

## <span class="material-icons">search</span> 4段階: 検索とナビゲーション

### 検索システムの有効化

```bash
# 依存関係の追加
uv add mkdocs-search

# 設定の更新
uv run mkdocs build --strict
```

### リアルタイム検索のテスト

1. 開発サーバーの実行: `uv run dev`
2. ブラウザでhttp://127.0.0.1:8080にアクセス
3. 検索窓に「MoAI」を入力
4. リアルタイム検索結果を確認

______________________________________________________________________

## 🌍 5段階: 多言語設定

### 言語ファイルの作成

```bash
# 英語
echo "# English Documentation" > docs/getting-started/index-en.md

# 日本語
echo "# 日本語ドキュメント" > docs/getting-started/index-ja.md

# 中国語
echo "# 中文文档" > docs/getting-started/index-zh.md
```

### 言語切り替えのテスト

```bash
# 多言語ドキュメントのビルド
uv run build

# 結果の確認
ls -la site/getting-started/
```

______________________________________________________________________

## 📊 6段階: デプロイの準備

### ローカルビルド

```bash
# 静的サイトのビルド
uv run build

# 結果の確認
ls -la site/

# ファイルサイズの確認
du -sh site/
```

### Vercelデプロイ

```bash
# 1. Vercel CLIのインストール
npm i -g vercel

# 2. ログイン
vercel login

# 3. プロジェクトのデプロイ
vercel --prod

# 4. デプロイの確認
vercel ls
```

______________________________________________________________________

## 🧪 7段階: テストと検証

### 自動化テスト

```bash
# 1. ドキュメントの妥当性検証
uv run validate

# 2. リンクチェック
uv run check-links

# 3. ビルドテスト
uv run build --strict
```

### 手動テストチェックリスト

- [ ] モッキングページが正常に表示される
- [ ] ダーク/ライトモード切り替えが動作する
- [ ] 検索機能が正常に動作する
- [ ] モバイルレスポンシブデザインを確認
- [ ] 多言語ドキュメントのアクセシビリティを確認

______________________________________________________________________

## 📋 完成確認チェックリスト

### システム状態

- [ ] UVのインストール完了
- [ ] プロジェクトのクローン成功
- [ ] 依存関係のインストール成功
- [ ] 開発サーバーの実行成功
- [ ] ドキュメントのビルド成功

### 機能確認

- [ ] モッキングページの正常表示
- [ ] 検索機能の正常動作
- [ ] 多言語サポートの確認
- [ ] レスポンシブデザインの確認
- [ ] ダーク/ライトモード切り替えの確認

### デプロイ確認

- [ ] ローカルビルドの成功
- [ ] Vercelデプロイの成功
- [ ] ドメインアクセスの確認
- [ ] SSL証明書の確認
- [ ] CDNパフォーマンスの確認

______________________________________________________________________

## 🚀 次のステップ

### 1. カスタマイズ

- デザインシステムの修正
- 新しい言語の追加
- カスタムコンポーネントの開発

### 2. コンテンツの追加

- APIドキュメントの生成
- チュートリアルの作成
- 高度なガイドの追加

### 3. プロダクションデプロイ

- 自動デプロイの設定
- モニタリングツールの接続
- パフォーマンスの最適化

______________________________________________________________________

## 🐛 トラブルシューティング

### 一般的な問題

#### UVインストールエラー

```bash
# キャッシュのクリア
uv cache clean

# 再インストール
pip install --upgrade uv
```

#### ビルドエラー

```bash
# キャッシュのクリア
rm -rf site/ .doit_db/

# 依存関係の再インストール
uv sync --force

# 再ビルド
uv run build
```

#### サーバー起動エラー

```bash
# ポートの変更
uv run dev --port 3000

# ログの確認
uv run dev --verbose
```

______________________________________________________________________

## 📞 サポート

### 公式ドキュメント

- **アドレス**: https://adk.mo.ai.kr
- **ステータス**: 24/7運営
- **更新**: リアルタイム同期

### 開発サポート

- **GitHub Issues**: [技術問題の報告](https://github.com/moai-adk/MoAI-ADK/issues)
- **GitHub Discussions**: [Q&A](https://github.com/moai-adk/MoAI-ADK/discussions)
- **メール**: support@mo.ai.kr

______________________________________________________________________

*最終更新: 2025-11-05 | バージョン: v0.17.0 | ステータス: 100% 完了*
