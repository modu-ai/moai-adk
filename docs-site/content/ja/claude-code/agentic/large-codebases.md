---
title: 大規模コードベース
weight: 80
draft: false
description: "数百万行の単一ツリーや複数パッケージのモノレポにおいて、Claude Code を効率的に使う戦略をまとめます。"
---

大規模コードベース (数百万行の単一リポジトリ、または複数パッケージのモノレポ) で Claude Code は正常に動作します。ただし、基本設定は小規模プロジェクトを想定しているため、**各作業が触れる部分に限定してコンテキストを絞る戦略** が必須です。

{{< callout type="info" >}}
**核心**: 大規模コードベースの問題は「全ファイルを読むこと」ではなく、現在の作業と **無関係な指示とファイルがコンテキストを占有すること** です。
{{< /callout >}}

## 1. 起動場所を決める

`claude` をどこで実行するかが、すべてを決めます。

| 起動位置 | ファイルアクセス範囲 | ロードされる CLAUDE.md | 適したケース |
|---------|-----------|---------------|---------|
| **リポジトリルート** | 全体 | ルートのみ (下位はオンデマンド) | 複数パッケージ・サブシステムにまたがる作業 |
| **下位ディレクトリ** | そのサブツリーのみ | そのディレクトリ + すべての上位ディレクトリ | 1 パッケージ・サブシステムに限定された作業 |

**ヒント**: 1 つのパッケージ (例: `packages/api/`) に集中するなら、そのディレクトリで `claude` を実行してください。すると自動的に `packages/web/` の指示はロードされません。

## 2. CLAUDE.md をディレクトリごとに分割

ルートにすべてのルールを入れると:
- 長すぎて可読性が落ちる
- あまりに一般的で役に立たない
- 作業と無関係な指示もロード

**解決**: ルートにリポジトリ全域ルールを入れ、各下位ディレクトリにその領域のルールを入れてください。

```markdown
# ./CLAUDE.md (ルート、すべてのセッションでロード)
This is a monorepo with three packages:
- packages/api: Node.js REST API with Express, TypeScript, PostgreSQL
- packages/web: React frontend with Vite, TypeScript, TailwindCSS
- packages/shared: shared TypeScript utilities

Run commands from the package directory.
```

```markdown
# ./packages/api/CLAUDE.md (このディレクトリ作業時のみロード)
This package is the REST API server.

- Run tests: `npm test` (uses Vitest)
- Run dev server: `npm run dev` (port 3001)
- Database migrations: `npm run migrate`

API routes are in src/routes/. Never write raw SQL in handlers.
```

Claude が `packages/api/` から起動すると:
- ルート + packages/api/ CLAUDE.md 両方ロード
- packages/web/ 指示は **ロードされない**

## 3. 無関係な CLAUDE.md を除外する

別チームのパッケージやレガシーコードは `claudeMdExcludes` でスキップ:

```json
{
  "claudeMdExcludes": [
    "**/packages/admin-dashboard/**",
    "**/packages/legacy-*/**"
  ]
}
```

ルート CLAUDE.md は相変わらずロードされ、除外されたパッケージは触れられません。

## 4. 生成コード・ベンダーコードを遮断

`.gitignore` に既にある経路 (node_modules、dist、build) は自動的に検索結果から除外されます。

コミットされた生成コードやベンダー SDK は権限ルールで遮断:

```json
{
  "permissions": {
    "deny": [
      "Read(./**/dist/**)",
      "Read(./**/build/**)",
      "Read(./**/*.generated.*)",
      "Read(./vendor/**)"
    ]
  }
}
```

## 5. コードインテリジェンス (LSP) プラグイン

ファイルを 1 行ずつ読んでシンボル定義を探すのは非効率です。言語サーバープラグインをインストールすると:

```bash
/plugin install typescript-lsp@claude-plugins-official
```

Claude が `go to definition`、`find references`、型エラーを直接照会できます。

- TypeScript、Python、Go、Rust など主要言語をサポート
- LSP バイナリが必要 (ガイド参照)

これでファイル読み込みを大幅に削減できます。

## 6. Worktree で必要なディレクトリだけチェックアウト

```json
{
  "worktree": {
    "sparsePaths": [
      ".claude",
      "packages/api",
      "packages/shared"
    ]
  }
}
```

`--worktree` で生成したワークツリーは全体ではなく、**指定したディレクトリだけ** チェックアウトします。

- 高速な生成 (全複製 vs 必要な部分のみ)
- ディスク容量を節約
- 複数ワークツリーの node_modules 重複排除:

```json
{
  "worktree": {
    "sparsePaths": ["packages/api", "packages/shared"],
    "symlinkDirectories": ["node_modules"]
  }
}
```

## 7. 他パッケージ・リポジトリへのアクセス権限

1 つのパッケージから始めたが兄弟パッケージの修正が必要なら:

```json
{
  "permissions": {
    "additionalDirectories": [
      "../shared",
      "../web"
    ]
  }
}
```

または実行時に:

```bash
claude --add-dir ../shared --add-dir ../web
```

## 8. パッケージ別 Skills を追加

各パッケージはその領域だけの自動化コマンド (Skills) を持つことができます。

```bash
mkdir -p packages/api/.claude/skills/api-testing
```

```markdown
# packages/api/.claude/skills/api-testing/SKILL.md
---
name: api-testing
description: API パッケージのテストパターン
---

## テスト構造

テストは `src/__tests__/` にあり `src/` の構造をミラーリングします。

## テスト実行

- すべて: `npm test`
- 単一ファイル: `npm test -- src/__tests__/routes/users.test.ts`

## テストユーティリティ

- `src/__tests__/helpers/db.ts`: setupTestDb()、teardownTestDb()
- `src/__tests__/helpers/auth.ts`: createTestUser()、getAuthToken()
```

packages/api から作業すると api-testing スキルが自動ロード。packages/web ではロードされません。

## 9. パッケージ間の作業調整

同じ変更が複数パッケージに触れるとき (例: 共有型アップデート + すべての呼び出し元修正):

**1 つのセッションで全変更処理**: すべてのファイルを一度に読み込んで決定の一貫性を維持します。

**事前に計画をファイルに保存**: 計画をマークダウンファイルに保存します。長いセッションはコンテキストが圧縮されますが、保存された計画は消えません。

## 10. 具体的な設定例: モノレポ

以下は完全な設定例です。

**ルート** (`.moai/config/sections/workflow.yaml` のような他の設定もルートに):

```json
// .claude/settings.json
{
  "permissions": {
    "deny": [
      "Read(./**/dist/**)",
      "Read(./**/build/**)"
    ]
  }
}
```

**packages/api** (`.claude/settings.json`):

```json
{
  "worktree": {
    "sparsePaths": [
      ".claude",
      "packages/api",
      "packages/shared"
    ],
    "symlinkDirectories": ["node_modules"]
  },
  "permissions": {
    "additionalDirectories": ["../shared"],
    "deny": [
      "Read(./**/dist/**)",
      "Read(./**/build/**)"
    ]
  }
}
```

この設定により:
- `.claude/`、`packages/api/`、`packages/shared/` だけチェックアウト (worktree)
- shared パッケージアクセス可能
- 生成・ベンダーファイルアクセス遮断

## 11. 大規模コードベースのヒント

### 範囲別検索

大きな変更をするときは、影響範囲を事前に把握してください:

```bash
grep -r "FunctionName" packages/api/  # api のみ検索
grep -r "FunctionName" packages/      # すべてのパッケージ
```

### レイヤー別分析

複数レイヤー (DB、API、UI) に触れる変更は、各レイヤーをそれぞれ理解して、1 つのセッションでは 1 つの変更に集中します。

### ドキュメント指示

大きな変更後もドキュメント維持が続くよう、変更計画に「docs 修正」項目を入れてください。

## 参考

このガイドは Anthropic の公式 [Set up Claude Code in a monorepo or large codebase](https://code.claude.com/docs/en/large-codebases) ドキュメントに基づいています。

追加の戦略は Anthropic の [Best practices for Claude Code](https://code.claude.com/docs/en/best-practices) ドキュメントも参照してください。
