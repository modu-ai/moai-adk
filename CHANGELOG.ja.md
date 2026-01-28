# v1.9.0 - Memory MCP、SVGスキル、ルール移行 (2026-01-26)

## 概要

セッション間での永続的なメモリ、包括的なSVGスキル、標準準拠のルールシステム移行を導入するマイナーリリース。

**主な機能**:
- **Memory MCP統合**: ユーザー設定とプロジェクトコンテキストの永続的な保存
- **SVGスキル**: SVGO最適化パターンとベストプラクティスを含む包括的なスキル
- **ルール移行**: `.moai/rules/*.yaml`から`.claude/rules/*.md`への移行（Claude Code公式標準）
- **バグ修正**: Rank batch sync表示問題（#300）

**影響**:
- Memory MCPによるエージェント間コンテキスト共有が有効に
- プロフェッショナルなSVG作成と最適化サポート
- よりクリーンで標準準拠のプロジェクト構造
- 正確なバッチ同期統計の表示

## Breaking Changes

なし。すべての変更は下位互換性があります。

## 追加

### Memory MCP統合

- **feat**: Memory MCP Server統合の追加 (99ab5273)
  - Claude Codeセッション間での永続的メモリ
  - ユーザー設定、プロジェクトコンテキスト、学習されたパターンの保存
  - ワークフロー中のエージェント間コンテキスト共有
  - 設定: `.mcp.json`, `.mcp.windows.json`
  - 新しいスキル: `moai-foundation-memory` (420行)

### SVG作成と最適化スキル

- **feat**: `moai-tool-svg`スキルの追加 (54c12a85)
  - W3C SVG 2.0仕様とSVGOドキュメントに基づく
  - 包括的なモジュール: 基本、スタイリング、最適化、アニメーション
  - 12個の動作するコード例
  - SVGO設定パターンとベストプラクティス
  - 合計3,698行（SKILL.md: 410、modules: 2,288、examples: 500、reference: 500）

### 言語ルールの強化

- **feat**: 強化されたツール情報で言語ルールを更新 (54c12a85)
  - Ruff設定パターン（flake8+isort+pyupgradeの置き換え）
  - Mypy strict modeガイドライン
  - テストフレームワークの推奨
  - 16個の言語ファイルを更新

## 変更

### CLAUDE.mdの最適化

- **refactor**: v1.9.0向けの大規模な整理とモジュール化 (4134e60d)
  - CLAUDE.mdを~60kから~30k文字に削減（40k制限準拠）
  - 詳細なコンテンツを`.claude/rules/`に移動して構成を改善
  - クロスプラットフォーム互換性のための`shell_validator.py`ユーティリティを追加
  - CLIコマンドの強化（doctor、init、update）
  - `moai-workflow-thinking`スキルを追加
  - bug-report.ymlイシューテンプレートを追加
  - 影響: 可読性、保守性、Claude Code互換性の改善

### ルールシステム移行

- **feat**: `.moai/rules/*.yaml`から`.claude/rules/*.md`への移行 (99ab5273)
  - 削除: 6,959行のYAMLルール
  - 追加: Claude Code公式Markdownルール
  - 構造: `.claude/rules/{core,development,workflow,languages}/`
  - 影響: 標準準拠、よりクリーンな構成

## 修正

### Rankコマンド

- **fix(rank)**: batch syncのネストされたAPIレスポンスを正しくパース (#300) (31b504ed)
  - 問題: `moai-adk rank sync`が常に"Submitted: 0"を表示
  - 根本原因: ネストされた`data`フィールド抽出の欠落
  - 修正: フィールドアクセス前に`data = response.get("data", {})`を追加
  - 影響: 正確な送信統計の表示

## インストールと更新

```bash
# 最新バージョンに更新
uv tool update moai-adk

# プロジェクトフォルダのテンプレートを更新
moai update

# バージョンを確認
moai --version
```

---

# v1.8.13 - Statusline Context Window修正 (2026-01-26)

## 概要

Statusline context window計算精度を改善するパッチリリース。

**主な修正**:
- Statusline context windowパーセンテージがClaude Codeの事前計算値を使用するように修正

**影響**:
- Context window表示がauto-compactと出力トークン予約を考慮
- より正確な残りトークン情報の提供

## 修正

### Statusline Context Window計算

- **fix(statusline)**: Claude Codeの事前計算されたcontext percentageを使用 (2dacecb7)
  - 優先度1: Claude Codeの`used_percentage`/`remaining_percentage`を使用（最も正確）
  - 優先度2: `current_usage`トークンから計算（フォールバック）
  - 優先度3: データがない時に0%を返す（セッション開始）
  - Auto-compact有効時または出力トークン予約時の精度を保証
  - ファイル: `src/moai_adk/statusline/main.py`

## インストールと更新

```bash
# 最新バージョンに更新
uv tool update moai-adk

# プロジェクトテンプレートを更新
moai update

# バージョンを確認
moai --version
```

---

# v1.8.12 - Hook Format Update & Login Command (2026-01-26)

## 概要

Claude Code hook format互換性修正とUX改善を含むパッチリリース。

**主な変更**:
- Claude Code settings.json hook formatの修正（新しいmatcher-based構造）
- `moai rank register`を`moai rank login`に改名（より直感的）
- settings.jsonが更新時に常に上書きされるようになりました；カスタマイズにはsettings.local.jsonを使用

**影響**:
- MoAI Rank hooksが最新のClaude Codeで動作
- `moai rank login`が新しい主コマンド（registerはまだエイリアスとして動作）
- ユーザーカスタマイズはsettings.local.jsonに保存

## Breaking Changes

なし。`moai rank register`は隠れたエイリアスとしてまだ動作します。
