# CLIコマンドリファレンス

`moai-adk` CLIは、プロジェクト管理とテンプレート同期を担当するClickベースのコマンドラインインターフェースです。Alfredコマンド（alfred:*）とは別に、ローカル環境設定とメンテナンスに使用されます。

## コアコマンド

| コマンド                 | 説明                                              | 使用時                          |
| ----------------------- | ------------------------------------------------- | ------------------------------- |
| `moai-adk init [path]` | 新規プロジェクト作成または既存プロジェクトにテンプレート注入 | Alfredを初めて導入する時 |
| `moai-adk doctor`      | 環境チェック（Python、uv、Git、ディレクトリ構造） | インストール後、問題発生時 |
| `moai-adk status`      | TAGサマリー、チェックポイント、テンプレートバージョン照会 | 作業前、レビュー前 |
| `moai-adk backup`      | `.moai/`、`.claude/`、CLAUDE.mdのバックアップ作成 | テンプレート更新前、大規模変更前 |
| `moai-adk update`      | パッケージ&テンプレート同期（最も重要なコマンド） | 新バージョンリリース後、定期チェック |

## コマンド詳細

### `moai-adk init`

**目的**: 新規プロジェクト初期化および基本構造作成

**使用方法**:

```bash
# 新規プロジェクト作成
moai-adk init my-project

# 現在のディレクトリ初期化
moai-adk init .

# 既存プロジェクトにMoAI-ADK注入
moai-adk init .
```

**生成される構造**:

```
my-project/
├── .moai/        # プロジェクトメタデータ
├── .claude/      # Alfredリソース
└── CLAUDE.md     # プロジェクトガイドライン
```

**初期化プロセス**:

1. Python環境確認
2. Gitリポジトリ初期化（存在しない場合）
3. `.moai/`ディレクトリ構造作成
4. `.claude/`リソーステンプレートコピー
5. デフォルト設定ファイル作成

### `moai-adk doctor`

**目的**: システム環境診断およびトラブルシューティング

**使用方法**:

```bash
moai-adk doctor
```

**診断項目**:

- ✅ Pythonバージョン（3.13+）
- ✅ uvパッケージマネージャー
- ✅ Gitリポジトリ状態
- ✅ `.moai/`ディレクトリ構造
- ✅ `.claude/`リソース整合性
- ✅ Claude Codeアクセシビリティ

**予想出力**:

```
🩺 MoAI-ADK System Check
✅ Python 3.13.0
✅ uv 0.5.1
✅ Gitリポジトリ初期化済み
✅ .moai/ディレクトリ構造正常
✅ .claude/リソース74個ロード済み
✅ Claude Codeアクセス可能

システムは正常です。Alfredを開始する準備ができました！
```

### `moai-adk status`

**目的**: プロジェクト状態サマリーおよび状態理解

**使用方法**:

```bash
moai-adk status
```

**表示情報**:

- SPEC進行状況（完了/進行中/待機）
- TAG統計（@SPEC/@TEST/@CODE/@DOC）
- 最近のチェックポイント
- テンプレートバージョン情報
- Gitワークフロー状態

**予想出力**:

```
📊 MoAI-ADK Project Status
:bullseye: Project: MyProject
📅 Last sync: 2025-01-15 14:30

📋 SPEC Progress
- ✅ Completed: 12
- 🔄 In Progress: 3
- ⏳ Pending: 5

🏷️ TAG Statistics
- @SPEC: 20 tags
- @TEST: 18 tags
- @CODE: 17 tags
- @DOC: 16 tags
- 🚨 Orphan tags: 2

📝 Version Info
- Template: v0.15.2
- Last update: 2025-01-10
- Backup available: .moai-backups/20250110/

🔄 Git Status
- Current branch: feature/auth-system
- Ahead of main: 12 commits
- Draft PR: #23
```

### `moai-adk backup`

**目的**: プロジェクトリソースバックアップ作成

**使用方法**:

```bash
moai-adk backup
```

**バックアップ対象**:

- `.moai/`全体ディレクトリ
- `.claude/`リソーステンプレート
- `CLAUDE.md`プロジェクトガイドライン
- Git状態情報

**バックアップ場所**:

```
.moai-backups/
└── 20250115_143000/
    ├── .moai/
    ├── .claude/
    ├── CLAUDE.md
    └── backup-info.json
```

### `moai-adk update`

**目的**: パッケージおよびテンプレート同期（最も重要なコマンド）

**使用方法**:

```bash
moai-adk update
```

**更新ステージ**:

1. **ステージ1**: パッケージバージョン確認
2. **ステージ2**: テンプレートバージョン比較
3. **ステージ3**: バックアップ作成およびマージ

**自動処理**:

- PyPIから最新バージョン確認
- `.moai-backups/`に現在のリソースバックアップ
- 新しいテンプレートと既存設定をマージ
- 競合発生時にガイダンスメッセージ

**出力例**:

```
🔄 MoAI-ADK Update Started
:package: Current version: v0.15.1
:package: Latest version: v0.15.2

📁 Creating backup...
✅ Backup created: .moai-backups/20250115_143000/

🔄 Updating templates...
🔧 Merging .moai/config.json
🔧 Updating Alfred agents
🔧 Syncing Skills (74 → 77)

✅ Update completed successfully!
📝 Changelog: Added moai-domain-ml Skill
⚠️  Please review .claude/settings.json changes
```

## 内部動作

### CLIアーキテクチャ

```
moai-adk
├── __main__.py           # Clickエントリーポイント
├── cli/
│   ├── commands/
│   │   ├── init.py      # プロジェクト初期化
│   │   ├── doctor.py    # 環境診断
│   │   ├── status.py    # 状態照会
│   │   ├── backup.py    # バックアップ作成
│   │   └── update.py    # テンプレート同期
│   └── utils.py          # 共通ユーティリティ
├── core/
│   ├── template.py      # テンプレート管理
│   ├── backup.py        # バックアップ/復元
│   └── filesystem.py    # ファイルシステム操作
└── templates/           # デフォルトテンプレートソース
```

### Richコンソール出力

- **色分け**: 成功（緑）、警告（黄）、エラー（赤）
- **進行バー**: 長時間タスクの進行率表示
- **テーブル形式**: 状態情報を整理して表示
- **ASCIIアート**: ロゴおよび区切り文字

### エラー処理

- **明確なメッセージ**: ユーザーフレンドリーなエラー説明
- **解決提案**: 問題解決のための具体的な方法
- **エラーコード**: 自動化スクリプトのための終了コード
- **ログ記録**: 問題追跡のための詳細ログ

## ベストプラクティス

### 定期的なメンテナンス

```bash
# 月間定期チェック
moai-adk doctor
moai-adk status
moai-adk backup
moai-adk update
```

### 大規模変更前

```bash
# 安全な変更手順
moai-adk backup  # 1. バックアップ作成
# 変更作業実行...
moai-adk status  # 2. 状態確認
moai-adk doctor  # 3. 環境チェック
```

### 新しいチームメンバーオンボーディング

```bash
# 標準オンボーディング手順
git clone <project>
cd <project>
moai-adk doctor  # 環境確認
moai-adk status  # プロジェクト理解
claude           # Alfred開始
/alfred:0-project  # プロジェクト初期化
```

## 関連リンク

- **[プロジェクト構造](project-structure)** - `.moai/`と`.claude/`ディレクトリ詳細
- **[Alfredコマンド](alfred/commands)** - alfred:*ワークフローコマンド
- **[ワークフロー](workflow)** - CLIとAlfredの連携方法



