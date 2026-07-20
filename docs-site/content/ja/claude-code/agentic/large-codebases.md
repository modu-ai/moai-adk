---
title: 大規模コードベース
weight: 80
draft: false
description: "数百万行の単一ツリーや複数パッケージのモノレポで Claude Code を効率的に使うコンテキスト縮小戦略を整理します。"
---

# 大規模コードベース

大規模コードベース — 数百万行の単一リポジトリでも、複数パッケージからなるモノレポでも — で Claude Code はきちんと動作します。ただしデフォルト設定は小さなプロジェクトを想定しているため、**各作業が実際に触れる部分だけへコンテキストを絞る戦略** が必須です。

{{< callout type="info" >}}
**ひとことで言うと**: 大規模コードベースの本当の問題は「ファイルが多いこと」ではなく、いまの作業と **無関係な指示文とファイルがコンテキストを埋めること** です。無関係なトークンは品質を下げると同時にコストを上げます — コンテキスト縮小こそがトークノミクスです。
{{< /callout >}}

## 開始位置を決める

`claude` をどこで実行するかが、その後のすべてを決めます。

| 開始位置 | ファイルアクセス範囲 | ロードされる CLAUDE.md | 適した場合 |
|---------|-----------|---------------|---------|
| **リポジトリルート** | 全体 | ルートのみ (下位はオンデマンド) | 複数のパッケージ/サブシステムにまたがる作業 |
| **サブディレクトリ** | そのサブツリーのみ | そのディレクトリ + すべての上位ディレクトリ | 1 つのパッケージ/サブシステムに限定された作業 |

1 つのパッケージ (例: `packages/api/`) だけに集中する作業なら、そのディレクトリで `claude` を実行してください。`packages/web/` の指示文はそもそもロードされないため、ルールを削る努力なしにコンテキストが自然と軽くなります。

## CLAUDE.md をディレクトリごとに分割する

ルートの CLAUDE.md 1 つにすべてのルールを入れると、3 つの問題が生まれます。

- 長くなりすぎて可読性が落ち
- すべてのパッケージに通用させようとして一般的すぎて役に立たなくなり
- 作業と無関係な指示文まで毎セッションロードされます

解決策は階層化です。ルートにはリポジトリ全域のルールだけを置き、各サブディレクトリにその領域のルールを置きます。

```markdown
# ./CLAUDE.md (ルート、すべてのセッションでロード)
This is a monorepo with three packages:
- packages/api: Node.js REST API with Express, TypeScript, PostgreSQL
- packages/web: React frontend with Vite, TypeScript, TailwindCSS
- packages/shared: shared TypeScript utilities

Run commands from the package directory.
```

```markdown
# ./packages/api/CLAUDE.md (このディレクトリの作業時のみロード)
This package is the REST API server.

- Run tests: `npm test` (uses Vitest)
- Run dev server: `npm run dev` (port 3001)
- Database migrations: `npm run migrate`

API routes are in src/routes/. Never write raw SQL in handlers.
```

Claude が `packages/api/` で起動すると、ルートと `packages/api/` の CLAUDE.md はどちらもロードされますが、`packages/web/` の指示文は **ロードされません**。

## 無関係な CLAUDE.md を除外する

他チームのパッケージやレガシーコードの指示文は、`claudeMdExcludes` 設定でスキップします。

```json
{
  "claudeMdExcludes": [
    "**/packages/admin-dashboard/**",
    "**/packages/legacy-*/**"
  ]
}
```

ルートの CLAUDE.md は引き続きロードされ、除外したパッケージの指示文だけがコンテキストから外れます。

## 生成コードとベンダーコードをブロックする

`.gitignore` にすでにあるパス (node_modules、dist、build) は、自動的に検索結果から除外されます。

コミットされた生成コードやベンダー SDK は、権限ルールで読み取り自体をブロックします。生成ファイルは長く反復的で、コンテキストの浪費が特に大きいのです。

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

## コードインテリジェンス (LSP) プラグイン

シンボルの定義を探すためにファイルを 1 行ずつ読むのは、トークンの観点で最も高くつく探索です。言語サーバープラグインをインストールすれば、定義への移動、参照検索、型エラーの直接照会が可能になり、ファイル読み取りそのものを大きく減らせます。

```bash
/plugin install typescript-lsp@claude-plugins-official
```

- TypeScript、Python、Go、Rust など主要言語をサポートします
- 該当言語の LSP バイナリがシステムにインストールされている必要があります ([プラグインのドキュメント](/ja/claude-code/extensibility/plugins) を参照)

## ワークツリーで必要なディレクトリだけチェックアウト

`--worktree` で作るワークツリーは、`worktree.sparsePaths` 設定で全体ではなく **列挙したディレクトリのみ** をチェックアウトできます。

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

- 作成が速くなり (フルクローンの代わりに必要な部分だけ)
- ディスク容量を節約でき
- `symlinkDirectories` で複数ワークツリーの node_modules の重複も取り除けます。

```json
{
  "worktree": {
    "sparsePaths": ["packages/api", "packages/shared"],
    "symlinkDirectories": ["node_modules"]
  }
}
```

`symlinkDirectories` に列挙したディレクトリは、メインチェックアウトのものをシンボリックリンクで共有します。

## 他のパッケージ/リポジトリへのアクセス権を与える

1 つのパッケージで始めたのに兄弟パッケージの修正が必要になったら、`additionalDirectories` でアクセス範囲を広げます。

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

設定の代わりにランタイムフラグでも可能です。

```bash
claude --add-dir ../shared --add-dir ../web
```

## パッケージ別スキルの追加

各パッケージは、その領域だけのスキルを持てます。スキルは必要なときだけロードされるため、パッケージ専用の知識をコンテキスト負担なしに保管する良い入れ物です。

```bash
mkdir -p packages/api/.claude/skills/api-testing
```

```markdown
# packages/api/.claude/skills/api-testing/SKILL.md
---
name: api-testing
description: API パッケージのテストパターン
---

## Test structure
Tests are in `src/__tests__/` mirroring `src/`.

## Running tests
- All: `npm test`
- Single file: `npm test -- src/__tests__/routes/users.test.ts`

## Test utilities
- `src/__tests__/helpers/db.ts`: setupTestDb(), teardownTestDb()
- `src/__tests__/helpers/auth.ts`: createTestUser(), getAuthToken()
```

`packages/api` で作業すると api-testing スキルが自動的にロードされ、`packages/web` ではロードされません。

## パッケージをまたぐ作業の調律

同じ変更が複数のパッケージに触れるとき (例: 共有型の更新とすべての呼び出し箇所の修正) は、2 つの原則が有効です。

- **1 つのセッションで変更全体を処理**: 関連ファイルを一度にロードし、決定の一貫性を維持します。
- **まず計画をファイルに保存**: 計画をマークダウンファイルに残しておきましょう。セッションが長くなるとコンテキストは圧縮されますが、ディスクに保存された計画は消えません。「重要な状態はファイルに残す」は、エージェンティックループ運用の基本でもあります。

## 具体的な設定例: モノレポ

以下は完全な設定例です。ルートにはリポジトリ全域のブロックルールを、パッケージにはそのパッケージのワークツリー・アクセス設定を置きます (MoAI-ADK プロジェクトなら `.moai/config/sections/workflow.yaml` のようなワークフロー設定もルートに位置します)。

**ルート** (`.claude/settings.json`):

```json
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

この設定の効果は次のとおりです。

- ワークツリーは `.claude/`、`packages/api/`、`packages/shared/` のみチェックアウト
- shared パッケージへアクセス可能
- 生成/ベンダーファイルへのアクセスをブロック

## ヒントとコツ

### 範囲を指定した検索

大きな変更をする前に、影響範囲をまず把握しましょう。検索範囲を狭める習慣が、読むべきファイル数を減らします。

```bash
grep -r "FunctionName" packages/api/  # api のみ検索
grep -r "FunctionName" packages/      # すべてのパッケージ
```

### レイヤーごとの分析

DB・API・UI のように複数レイヤーに触れる変更なら、各レイヤーを別々に理解し、1 つのセッションでは 1 つの変更だけに集中します。

### ドキュメント化の指示

大規模な変更の後もドキュメントが古びないよう、変更計画に「docs の修正」項目を含めてください。

## 関連ドキュメント

- [コンテキストウィンドウ](/ja/claude-code/context-memory/context-window)
- [ワークツリー](/ja/claude-code/agentic/worktrees)
- [ベストプラクティス](/ja/claude-code/agentic/best-practices)

## 参考資料

- [Set up Claude Code in a monorepo or large codebase (公式ドキュメント)](https://code.claude.com/docs/en/large-codebases)
- [Best practices for Claude Code (公式ドキュメント)](https://code.claude.com/docs/en/best-practices)

{{< callout type="tip" >}}
モノレポで最も手軽な第一手は、「1 つのパッケージの作業はそのパッケージのディレクトリで `claude` を実行する」ことです。設定ファイルを 1 つも触らずに無関係な指示文のロードを断ち切る、費用対効果が最も大きい習慣です。
{{< /callout >}}
