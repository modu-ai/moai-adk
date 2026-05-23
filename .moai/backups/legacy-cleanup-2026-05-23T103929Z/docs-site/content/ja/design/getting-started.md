---
title: はじめに
description: /moai designコマンドでハイブリッド設計ワークフローを開始
weight: 20
draft: false
---

# はじめに

## 前提条件

- MoAI-ADKプロジェクト初期化完了
- `.moai/project/brand/`ディレクトリ準備または新規作成予定
- Claude Codeデスクトップクライアント v2.1.50以上

## ブランドコンテキストインタビュー

`/moai design`実行時に、まず**ブランドコンテキストインタビュー**が進行します。

### 3つのブランドファイル生成

インタビューで`.moai/project/brand/`に以下3つのファイルが生成されます:

1. **brand-voice.md** — ブランドトーン、用語、メッセージングガイドライン
2. **visual-identity.md** — 色、タイポグラフィ、ビジュアル言語
3. **target-audience.md** — 顧客プロファイルと選好

### インタビュープロセス

1. Claude Codeで`/moai design`実行
2. 「ブランドコンテキストが不完全です」メッセージ表示
3. ブランドインタビュー選択
4. `manager-spec`エージェントが質問提示
5. 自由形式で回答
6. 3つのファイル自動生成

質問例:
- 「ブランドトーンは専門的ですか、それとも親しみやすいですか?」
- 「主要なブランドカラー3つを選んでください」
- 「ターゲット顧客の主要な課題は何ですか?」

## パス選択

ブランドコンテキスト設定後、パス選択UIが表示されます:

### オプション1(推奨) — Claude Design利用

**Claude.ai/design**でデザイン生成後、**handoffバンドル**としてエクスポート

**必須:**
- Claude.ai Pro、Max、Team、またはEnterpriseサブスクリプション

**利点:**
- 直感的なUI/UX
- リアルタイムコラボレーション(Teamサブスクリプション)
- 複数の入力形式対応(テキスト、画像、Figma、GitHubリポジトリ)

### オプション2 — コードベース設計

**moai-domain-copywriting**と**moai-domain-brand-design**スキル利用

**必須:**
- 完成した`brand-voice.md`と`visual-identity.md`

**利点:**
- 追加サブスクリプション不要
- 自動設計トークン生成
- バージョン管理便利

## 最初の実行

```bash
# Claude Codeで実行
/moai design
```

実行順序:
1. `.agency/`チェック(移行ガイド表示)
2. ブランドコンテキスト確認
3. 不足ファイルのインタビュー実行
4. パス選択UI表示
5. 選択パスのワークフロー実行

## セットアップ確認

生成されたブランドファイル確認:

```bash
ls -la .moai/project/brand/
# brand-voice.md
# visual-identity.md
# target-audience.md
```

## 次のステップ

- **パスA選択時:** [Claude Designハンドオフ](./claude-design-handoff.md)ガイド参照
- **パスB選択時:** [コードベースパス](./code-based-path.md)ガイド参照

## トラブルシューティング

### ブランドコンテキスト更新

ファイルを直接編集:

```bash
# 任意のエディタで編集
vim .moai/project/brand/brand-voice.md
```

変更は次の`/moai design`実行で自動反映。

### インタビュー再実行

```bash
# 現在のファイルをバックアップして再実行
mv .moai/project/brand .moai/project/brand.backup
/moai design
```
