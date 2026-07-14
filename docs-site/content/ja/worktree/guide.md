---
title: Git Worktree 完全ガイド
weight: 20
draft: false
---

Git Worktree を使った MoAI-ADK 並列開発のすべて — 基礎概念からコマンド
リファレンス、ワークフロー、ベストプラクティスまでこのドキュメント 1 編で整理します。

## 目次

1. [Worktree の基礎](#worktree-の基礎)
2. [コマンド詳細参照](#コマンド詳細参照)
3. [ワークフローガイド](#ワークフローガイド)
4. [高度な機能](#高度な機能)
5. [ベストプラクティス](#ベストプラクティス)

---

## Worktree の基礎

### Git Worktree とは何ですか?

Git Worktree は **1 つの Git リポジトリを複数のディレクトリで同時に作業** できるように
してくれる Git 内蔵機能です。ブランチを行き来するたびに `git checkout` でコンテキストを
入れ替える代わりに、ブランチごとにディレクトリを 1 つずつ開いておきます。

```mermaid
graph TB
    subgraph Traditional["従来の方式"]
        T1[単一の作業ディレクトリ]
        T2[ブランチ切替が必要]
        T3[コンテキストスイッチングのコスト]
    end

    subgraph Worktree["Worktree 方式"]
        W1[Worktree 1<br/>feature/A]
        W2[Worktree 2<br/>feature/B]
        W3[Worktree 3<br/>main]
    end

    Traditional -.->|不便| Worktree
```

### MoAI-ADK での Worktree

MoAI-ADK はこの機能の上に SPEC 単位の隔離環境を載せます。各 SPEC が完全に
独立した環境を持つため、エージェントが並列で働いても互いの作業を踏み
ません:

- **独立した Git 状態** — 各 Worktree は自身のブランチとコミット履歴を維持します
- **分離された LLM 設定** — Worktree ごとに異なる LLM 実行モードを使えます。
  計画には Claude、実装には GLM を割り当てるトークノミクス運用がここから出てきます
- **隔離された作業空間** — ファイルシステムレベルで完全に分離されます

---

## コマンド詳細参照

### moai worktree new

新しい Worktree を生成します。

#### 文法

```bash
moai worktree new SPEC-ID [options]
```

#### パラメータ

- **SPEC-ID** (必須): 生成する SPEC の ID (例: `SPEC-AUTH-001`)

#### オプション

- `--path PATH`: Worktree パスの直接指定 (デフォルト値: SPEC ID なら `.moai/worktrees/<SPEC-ID>`、それ以外は `../<branch-name>`)
- `--base BRANCH`: 基準ブランチ (デフォルト値: `origin/main`、自動 fetch)。ローカル専用のコミットには `--base main` を使用
- `--from-current`: 現在の HEAD を基準として使用 (`git fetch origin main` をスキップ、`--base` と相互排他)
- `--tmux`: Worktree 生成後に tmux セッションを生成
- `--team`: 新しい Worktree で Claude/GLM セッションを自動実行

#### 使用例

```bash
# 基本的な使い方 (origin/main 基準)
moai worktree new SPEC-AUTH-001

# ローカル main 基準で生成
moai worktree new SPEC-AUTH-001 --base main

# 現在の HEAD 基準で生成
moai worktree new SPEC-AUTH-001 --from-current

# tmux セッションと一緒に生成
moai worktree new SPEC-AUTH-001 --tmux
```

#### 動作プロセス

```mermaid
sequenceDiagram
    participant User as ユーザー
    participant CLI as moai worktree
    participant Git as Git
    participant FS as ファイルシステム

    User->>CLI: moai worktree new SPEC-AUTH-001
    CLI->>Git: git worktree add
    Git->>Git: feature/SPEC-AUTH-001 ブランチ生成
    Git->>FS: ~/.moai/worktrees/{ProjectName}/SPEC-AUTH-001/ ディレクトリ生成
    Git->>Git: ブランチのチェックアウト
    CLI->>CLI: .moai/config 設定のコピー
    CLI->>User: Worktree 生成完了

    Note over User,FS: SPEC-AUTH-001 のための<br/>完全に独立した環境の生成
```

---

### moai worktree go

Worktree パスを出力します。シェルナビゲーションに使うようパス文字列だけを
標準出力に送り出し、シェルセッションを直接開始はしません。シェルの `cd` と
組み合わせて使います。

#### 文法

```bash
moai worktree go SPEC-ID
```

#### パラメータ

- **SPEC-ID** (必須): パスを出力する Worktree の ID

#### 使用例

```bash
# パスのみ出力
moai worktree go SPEC-AUTH-001

# 出力されたパスへ移動
cd "$(moai worktree go SPEC-AUTH-001)"

# 移動後に開発開始
moai glm
claude
> /moai run SPEC-AUTH-001
```

#### 動作プロセス

```mermaid
flowchart TD
    A[moai worktree go SPEC-ID] --> B{Worktree 存在?}
    B -->|いいえ| C[エラーメッセージ]
    B -->|はい| D[Worktree パスを stdout に出力]
    D --> E["cd \"$(...)\" などシェルで活用"]
```

---

### moai worktree list

すべての Worktree の一覧を表示します。

#### 文法

```bash
moai worktree list [options]
```

#### オプション

- `-v, --verbose`: 各 Worktree の詳細情報を含む

#### 使用例

```bash
# 基本一覧
moai worktree list

# 詳細情報
moai worktree list --verbose

# 出力例
SPEC-AUTH-001  feature/SPEC-AUTH-001  /path/to/worktree/SPEC-AUTH-001  [active]
SPEC-AUTH-002  feature/SPEC-AUTH-002  /path/to/worktree/SPEC-AUTH-002
SPEC-AUTH-003  feature/SPEC-AUTH-003  /path/to/worktree/SPEC-AUTH-003
```

---

### moai worktree done

Worktree を削除しオプションでブランチを削除します。**マージ・プッシュは
行いません** — base ブランチへのマージは `git merge` や PR で別途
進めます。

#### 文法

```bash
moai worktree done SPEC-ID [options]
```

#### パラメータ

- **SPEC-ID** (必須): 完了する Worktree の ID

#### オプション

- `--force`: コミットされていない変更があっても強制削除
- `--delete-branch`: Worktree 削除後にブランチも削除
- `--auto`: 自動化用の無出力モード (例: PR マージ後の整理)

#### 使用例

```bash
# Worktree 削除
moai worktree done SPEC-AUTH-001

# Worktree 削除 + ブランチ削除
moai worktree done SPEC-AUTH-001 --delete-branch

# PR マージ後の自動整理 (無出力)
moai worktree done SPEC-AUTH-001 --auto
```

#### 動作プロセス

```mermaid
flowchart TD
    A[moai worktree done SPEC-ID] --> B{Worktree 存在?}
    B -->|いいえ| C[エラーメッセージ]
    B -->|はい| D[Worktree 削除]
    D --> E{--delete-branch?}
    E -->|はい| F[ブランチ削除]
    E -->|いいえ| G[ブランチ維持]
    F --> H[完了]
    G --> H[完了]
```

---

### moai worktree remove

Worktree を削除します (マージなし)。ブランチは維持されます。

#### 文法

```bash
moai worktree remove PATH [options]
```

#### パラメータ

- **PATH** (必須): 削除する Worktree のパス

#### オプション

- `--force`: コミットされていない変更があっても強制削除

#### 使用例

```bash
# 基本削除
moai worktree remove .moai/worktrees/SPEC-AUTH-001

# 強制削除
moai worktree remove .moai/worktrees/SPEC-AUTH-001 --force
```

---

### moai worktree status

Worktree の状態を確認します。

#### 文法

```bash
moai worktree status [options]
```

#### オプション

- `--all`: 全コミットハッシュを含むすべての詳細情報を表示

#### 使用例

```bash
# Worktree の状態
moai worktree status

# 全詳細情報
moai worktree status --all

# 出力例
Worktree: SPEC-AUTH-001
Branch: feature/SPEC-AUTH-001
Path: /path/to/worktree/SPEC-AUTH-001
Status: Clean (2 commits ahead of main)
LLM: GLM 5
```

---

### moai worktree clean

マージまたは完了した Worktree を整理します。

#### 文法

```bash
moai worktree clean [options]
```

#### オプション

- `--merged-only`: ブランチが base にマージされた Worktree のみ削除
- `--base BRANCH`: `--merged-only` 判定に使う base ブランチ (デフォルト値: `main`)

#### 使用例

```bash
# マージ済みの Worktree を整理 (base=main)
moai worktree clean --merged-only

# 別の base ブランチ基準で整理
moai worktree clean --merged-only --base develop
```

---

### moai worktree config

Worktree 設定を確認または修正します。

#### 文法

```bash
moai worktree config [key] [value]
```

#### パラメータ

- **key** (オプション): 設定キー
- **value** (オプション): 設定値

#### 使用例

```bash
# すべての設定を表示
moai worktree config

# 特定の設定を確認
moai worktree config root

# 設定変更
moai worktree config root /new/path/to/worktrees
```

---

## ワークフローガイド

### 完全な開発サイクル

```mermaid
flowchart TD
    Start(( )) -->|"Plan with Worktree"| Plan["Plan"]
    Plan -->|"Worktree 生成"| Implement["Implement"]
    Implement -->|"DDD 実装"| Implement
    Implement -->|"ドキュメント同期"| Document["Document"]
    Document -->|"コードレビュー"| Review["Review"]
    Review -->|"承認"| Merge["Merge"]
    Review -->|"修正が必要"| Implement
    Merge -->|"moai worktree done"| Done["Done"]
```

### ステップ 1: SPEC 計画 (Phase 1)

```bash
# Terminal 1 で
> /moai plan "ユーザー認証システムの実装" --worktree
```

**出力**:

```
✓ SPEC ドキュメント生成: .moai/specs/SPEC-AUTH-001/spec.md
✓ Worktree 生成: ~/.moai/worktrees/{ProjectName}/SPEC-AUTH-001
✓ ブランチ生成: feature/SPEC-AUTH-001
✓ ブランチ切替完了

次のステップ:
1. 新しいターミナルで実行: moai worktree go SPEC-AUTH-001
2. LLM 変更: moai glm
3. 開発開始: claude
```

### ステップ 2: 実装 (Phase 2)

```bash
# Terminal 2 で
moai worktree go SPEC-AUTH-001

# Worktree に進入するとプロンプトが変わる
(SPEC-AUTH-001) $ moai glm
→ GLM 5 に設定されました。

(SPEC-AUTH-001) $ claude
> /moai run SPEC-AUTH-001
```

**作業の流れ**:

```mermaid
sequenceDiagram
    participant T1 as Terminal 1<br/>Plan
    participant T2 as Terminal 2<br/>Implement
    participant Git as Git Repository

    T1->>Git: feature/SPEC-AUTH-001 生成
    T1->>T2: Worktree 生成完了の通知

    T2->>T2: moai worktree go SPEC-AUTH-001
    T2->>T2: moai glm
    T2->>Git: DDD 実装コミット群
    Note over T2: ANALYZE → PRESERVE → IMPROVE

    T2->>Git: さらに多くの実装コミット
    T2->>T2: /moai sync SPEC-AUTH-001
    T2->>Git: ドキュメント化コミット
```

### ステップ 3: 完了およびマージ (Phase 3)

```bash
# Terminal 2 で作業完了後 (push は別途 git/PR で進行)
exit

# base ブランチのマージは git merge または PR で処理した後、
# Terminal 1 で Worktree 整理
moai worktree done SPEC-AUTH-001 --delete-branch
```

**プロセス**:

```mermaid
flowchart TD
    A[作業完了] --> B[git merge または PR で base マージ]
    B --> C[moai worktree done SPEC-ID]
    C --> D[Worktree 削除]
    D --> E{--delete-branch?}
    E -->|はい| F[ブランチ削除]
    E -->|いいえ| G[ブランチ維持]
    F --> H[完了]
    G --> H[完了]
```

---

## 高度な機能

### 並列作業戦略

#### 戦略 1: Plan と Implement の分離

トークノミクスの基本戦略です。計画ステップは高推論モデル (Opus) にまとめて処理し、
実装ステップは低コストモデル (GLM) で並列分散します:

```mermaid
graph TB
    subgraph Planning["Planning Phase (Opus)"]
        P1[/moai plan<br/>SPEC-001/]
        P2[/moai plan<br/>SPEC-002/]
        P3[/moai plan<br/>SPEC-003/]
    end

    subgraph Implementation["Implementation Phase (GLM)"]
        I1[moai worktree go<br/>SPEC-001]
        I2[moai worktree go<br/>SPEC-002]
        I3[moai worktree go<br/>SPEC-003]
    end

    Planning --> Implementation
```

#### 戦略 2: 同時開発

```bash
# Terminal 1: SPEC-001 Plan
> /moai plan "認証" --worktree

# Terminal 2: SPEC-002 Plan (完了後)
> /moai plan "ログ" --worktree

# Terminal 3, 4, 5: 並列実装
moai worktree go SPEC-001 && moai glm  # Terminal 3
moai worktree go SPEC-002 && moai glm  # Terminal 4
moai worktree go SPEC-003 && moai glm  # Terminal 5
```

### Worktree 間の切替

```bash
# 現在の Worktree を確認
moai worktree status

# 別の Worktree へ切替
moai worktree go SPEC-AUTH-002

# または直接移動
cd ~/.moai/worktrees/SPEC-AUTH-002
```

### 衝突の解決

```mermaid
flowchart TD
    A[マージ試行] --> B{衝突?}
    B -->|いいえ| C[マージ完了]
    B -->|はい| D[衝突ファイルの表示]
    D --> E[手動解決]
    E --> F[git add]
    F --> G[git commit]
    G --> H[マージ完了]
```

---

## ベストプラクティス

### 1. Worktree 命名規則

```bash
# 良い例
moai worktree new SPEC-AUTH-001      # 明確な SPEC ID
moai worktree new SPEC-FRONTEND-007  # カテゴリを含む

# 避けるべき例
moai worktree new feature-branch     # SPEC ID を使わない
moai worktree new temp               # 曖昧な名前
```

### 2. 定期的な整理

```bash
# マージ済みの Worktree を定期整理
moai worktree clean --merged-only
```

### 3. LLM 選択ガイド

作業ステップ別にモデルを分けて割り当てるのが Worktree トークノミクスの核心です:

```mermaid
graph TD
    A[作業タイプ] --> B[Plan<br/>/moai plan]
    A --> C[Implement<br/>/moai run]
    A --> D[Document<br/>/moai sync]

    B --> E[Claude Opus<br/>高コスト/高品質]
    C --> F[GLM 5<br/>低コスト]
    D --> G[Claude Sonnet<br/>中コスト]
```

### 4. コミットメッセージ規則

```bash
# Worktree でコミットするとき
git commit -m "feat(SPEC-AUTH-001): JWT ベースの認証を実装

- JWT トークン生成/検証ロジックを追加
- リフレッシュトークンローテーションを実装
- ログアウト時のトークン無効化

Co-Authored-By: Claude <noreply@anthropic.com>"
```

### 5. ターミナル管理

```bash
# 各 Worktree に別々のターミナルを使用
# iTerm2、VS Code、または tmux の使用を推奨

# tmux の例
tmux new-session -d -s spec-001 'moai worktree go SPEC-001'
tmux new-session -d -s spec-002 'moai worktree go SPEC-002'

# セッションの切替
tmux attach-session -t spec-001
```

### 6. 進行状況の追跡

```bash
# すべての Worktree の状態を確認
moai worktree status --all

# Git ログの確認
cd ~/.moai/worktrees/{ProjectName}/SPEC-AUTH-001
git log --oneline --graph --all

# 変更事項の確認
git diff main
```

## tmux 統合と自動マージ

### moai worktree new --tmux フラグ

tmux セッションを自動生成してワークツリー環境で隔離された開発が可能です。

```bash
moai worktree new SPEC-AUTH-001 --tmux
```

**動作の流れ:**
1. Worktree 生成 (従来の動作)
2. tmux セッションの自動生成 (名前: `moai-{ProjectName}-{SPEC-ID}`)
3. LLM モードに応じた環境変数の注入 (GLM/CG モード)
4. Worktree へ cd 後 `/moai run {SPEC-ID}` を実行

```bash
# tmux セッションにアタッチ
tmux attach-session -t moai-my-project-SPEC-AUTH-001
```

{{< callout type="info" >}}
tmux がインストールされていない場合は graceful degradation: 手動 cd の案内メッセージが表示されます。
{{< /callout >}}

### 実行モード選択ゲート (Decision Point 3.5)

`/moai plan` 完了後 Run 開始前に、実行モードを自動検出しユーザーに選択を要請します。

**tmux 使用可能時 (3 つのオプション):**
- Worktree + \{現在のモード\} (Recommended): ワークツリー + tmux セッションの生成
- Team Mode: Agent Teams の並列実行
- Sub-agent Mode: 順次実行

**tmux 使用不可時 (2 つのオプション):**
- Sub-agent Mode (Recommended)
- Team Mode (in-process)

### Auto-merge のデフォルト動作

ワークツリーコンテキストで `/moai sync` を実行すると auto-merge がデフォルト動作です。

| フラグ | 動作 |
|--------|------|
| (なし) | ワークツリーコンテキストで自動マージ |
| `--no-merge` | 自動マージのスキップ |
| `--merge` | Deprecated (警告を表示) |

### ポストマージ自動クリーンアップ

PR マージ成功時に自動整理:
- ワークツリーディレクトリの削除
- フィーチャーブランチの削除 (`--delete-branch`)
- レジストリの更新

{{< callout type="warning" >}}
クリーンアップの失敗はマージ結果に影響を与えません。失敗時の手動整理: `moai worktree done SPEC-{ID}`
{{< /callout >}}

### エラーハンドリング (errors.go)

構造化されたエラータイプと復旧コマンドを提供します。

| エラータイプ | 説明 | 復旧コマンド |
|-----------|------|-----------|
| `WorktreeCreateError` | Worktree 生成失敗 | `moai worktree new {SPEC-ID}` |
| `TmuxNotAvailableError` | tmux 使用不可 | `cd {path} && /moai run {SPEC-ID}` |
| `AutoMergeBlockedError` | 自動マージ遮断 | `/moai sync {SPEC-ID}` |
| `CleanupFailedError` | 整理失敗 | `moai worktree done {SPEC-ID}` |

## 関連ドキュメント

- [Git Worktree 概要](/ja/worktree/)
- [実際の使用例](/ja/worktree/examples)
- [よくある質問](/ja/worktree/faq)
