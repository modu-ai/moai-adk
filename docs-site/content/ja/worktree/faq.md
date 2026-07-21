---
title: Git Worktree よくある質問
weight: 40
draft: false
---

Git Worktree を使っていて遭遇する質問と問題を 1 か所に整理しました。

## 目次

1. [基本概念](#基本概念)
2. [使用関連](#使用関連)
3. [問題解決](#問題解決)
4. [性能および最適化](#性能および最適化)
5. [チーム協業](#チーム協業)

---

## 基本概念

### Q: Git Worktree と一般的なブランチの違いは何ですか?

**A**: Git Worktree は **物理的に分離されたディレクトリ** で作業できるように
してくれます:

```mermaid
graph TB
    subgraph Traditional["一般的なブランチ方式"]
        T1[単一のディレクトリ]
        T2[git checkout で<br/>ブランチ切替]
        T3[コンテキスト切替のコストが発生]
    end

    subgraph Worktree["Worktree 方式"]
        W1[ディレクトリ 1<br/>feature/A]
        W2[ディレクトリ 2<br/>feature/B]
        W3[ディレクトリ 3<br/>main]
        W4[同時に複数のブランチで作業可能]
    end

    Traditional -.->|非効率| Worktree
```

**主な違い**:

| 特徴          | 一般的なブランチ         | Git Worktree    |
| ------------- | ------------------- | --------------- |
| 作業ディレクトリ | 1 個を共有            | N 個が独立        |
| ブランチ切替   | `git checkout` が必要 | ディレクトリ移動のみ |
| 同時作業     | 不可能              | 可能            |
| LLM 設定      | 共有される              | 独立            |
| 衝突の可能性   | 高い                | 低い            |

---

### Q: なぜ Worktree を使うべきですか?

**A**: 並列開発とトークノミクス、2 つの理由が核心です:

1. **LLM 設定の独立性** — SPEC ごとに異なる LLM を割り当てられます
   - Plan ステップ: Opus (高品質な推論)
   - Implement ステップ: GLM (低コスト)
   - Document ステップ: Sonnet (中程度)

2. **並列開発** — 複数の SPEC を同時に進められます
3. **衝突防止** — 独立した作業空間が衝突を最小化します
4. **コスト削減** — 実装ステップに GLM を使うと約 70% のコストが減ります

```mermaid
graph TB
    A[Worktree 未使用] --> B[すべてのセッションに<br/>同じ LLM を適用]
    B --> C[高いコスト<br/>Opus のみ使用]

    D[Worktree 使用] --> E[各 Worktree に<br/>独立した LLM]
    E --> F[70% のコスト削減<br/>GLM 使用可能]
```

---

### Q: MoAI-ADK で Worktree は必須ですか?

**A**: いいえ、必須ではありませんが **強く推奨** します:

- **単一 SPEC 開発**: Worktree なしでも可能
- **多重 SPEC 開発**: Worktree が事実上必須
- **チーム協業**: Worktree で衝突を防止
- **コスト最適化**: Worktree で LLM を分離

---

## 使用関連

### Q: Worktree を生成する方法は?

**A**: 2 つの方法があります:

**方法 1: 自動生成 (推奨)**

```bash
# SPEC 計画ステップで自動生成
> /moai plan "機能の説明" --worktree

# 自動的に:
# 1. SPEC ドキュメント生成
# 2. Worktree 生成
# 3. Feature ブランチ生成
```

**方法 2: 手動生成**

```bash
# Worktree 手動生成 (デフォルト: origin/main 基準)
moai worktree new SPEC-AUTH-001

# ローカル main 基準で生成
moai worktree new SPEC-AUTH-001 --base main
```

---

### Q: Worktree にはどう進入しますか?

**A**: `moai worktree go` は Worktree パスを出力します。シェルの `cd` と
組み合わせて移動します (シェルセッションを直接開始はしません):

```bash
# パス出力後に移動
cd "$(moai worktree go SPEC-AUTH-001)"
```

**進入後の作業の流れ**:

```mermaid
flowchart TD
    A[moai worktree go SPEC-ID] --> B[パスを stdout に出力]
    B --> C["cd \"$(...)\" で移動"]
    C --> D{LLM 変更?}
    D -->|はい| E[moai glm]
    D -->|いいえ| F[Claude 開始]
    E --> F
    F --> G["/moai run SPEC-ID"]
```

---

### Q: 複数の Worktree を同時に使えますか?

**A**: はい、無制限に可能です:

```bash
# Terminal 1
cd "$(moai worktree go SPEC-AUTH-001)"
(SPEC-AUTH-001) $ moai glm

# Terminal 2
cd "$(moai worktree go SPEC-LOG-002)"
(SPEC-LOG-002) $ moai glm

# Terminal 3
cd "$(moai worktree go SPEC-API-003)"
(SPEC-API-003) $ moai glm

# すべて同時に作業可能
```

**並列作業の可視化**:

```mermaid
graph TB
    subgraph Time["時間経過"]
        T1[09:00]
        T2[10:00]
        T3[11:00]
        T4[12:00]
    end

    subgraph Worktree1["SPEC-AUTH-001"]
        W1A[Plan]
        W1B[Implement]
        W1C[Done]
    end

    subgraph Worktree2["SPEC-LOG-002"]
        W2A[Plan]
        W2B[Implement]
    end

    subgraph Worktree3["SPEC-API-003"]
        W3A[Plan]
    end

    T1 --> W1A
    T1 --> W2A
    T1 --> W3A

    T2 --> W1B
    T2 --> W2B

    T3 --> W1C
    T3 --> W2B
```

---

### Q: Worktree を完了する方法は?

**A**: `moai worktree done` は Worktree を削除しオプションでブランチを
削除します。**マージ・プッシュはしません** — base マージは `git merge` や
PR で先に処理してください:

```bash
# Worktree 削除のみ
moai worktree done SPEC-AUTH-001

# Worktree 削除 + ブランチ削除
moai worktree done SPEC-AUTH-001 --delete-branch

# 自動化用の無出力モード (PR マージ後の整理)
moai worktree done SPEC-AUTH-001 --auto
```

**完了プロセス**:

```mermaid
flowchart TD
    A[git merge または PR で base マージ] --> B[moai worktree done SPEC-ID]
    B --> C[Worktree 削除]
    C --> D{--delete-branch?}
    D -->|はい| E[ブランチ削除]
    D -->|いいえ| F[ブランチ維持]
    E --> G[完了]
    F --> G[完了]
```

---

## 問題解決

### Q: Worktree の衝突が発生しました

**A**: 次のステップで解決してください:

マージ衝突は `git merge` や PR のステップで発生します。Worktree CLI はマージに
関与しません。

```mermaid
flowchart TD
    A[git merge 衝突が発生] --> B[衝突ファイルの確認]
    B --> C[衝突ファイルを開く]
    C --> D[衝突マーカーを探す &lt;&lt;&lt;&lt;&lt;&lt;&lt;]
    D --> E[手動マージ]
    E --> F[git add]
    F --> G[git commit]
    G --> H[moai worktree done で整理]
```

**実際の例**:

```bash
git checkout main
git merge feature/SPEC-AUTH-001
✗ マージ衝突が発生!

# 1. 衝突ファイルの確認
git status
# 衝突ファイル: src/auth/jwt.ts

# 2. 衝突の解決
code src/auth/jwt.ts

# 3. 衝突マーカーの確認および修正
<<<<<<< HEAD
const secret = process.env.JWT_SECRET;
=======
const secret = config.jwt.secret;
>>>>>>> feature/SPEC-AUTH-001

# 4. マージ
const secret = process.env.JWT_SECRET || config.jwt.secret;

# 5. コミット
git add src/auth/jwt.ts
git commit -m "fix: resolve merge conflict"
git push origin main

# 6. マージ後に Worktree 整理
moai worktree done SPEC-AUTH-001 --delete-branch
✓ 完了!
```

---

### Q: Worktree が破損しました

**A**: 次のステップで復旧してください:

```bash
# 1. 診断
moai worktree status
✗ Worktree ディレクトリが存在しません

# 2. 既存の Worktree を削除
moai worktree remove .moai/worktrees/SPEC-AUTH-001 --force

# 3. Worktree の再生成
moai worktree new SPEC-AUTH-001

# 4. 復旧の確認
moai worktree status
✓ Worktree 正常
```

---

### Q: ディスク容量が足りません

**A**: マージが終わった Worktree を整理してください:

```bash
# 1. ディスク使用量の確認
$ du -sh .moai/worktrees/*
2.5G    .moai/worktrees/SPEC-AUTH-001
1.8G    .moai/worktrees/SPEC-LOG-002
3.2G    .moai/worktrees/SPEC-API-003

# 2. base にマージされた Worktree のみ整理
$ moai worktree clean --merged-only

✓ マージ済みの Worktree の整理完了
✓ ディスク容量の確保
```

**整理戦略**:

```mermaid
graph TD
    A[Worktree 整理が必要] --> B{base にマージ完了?}
    B -->|はい| C[moai worktree clean --merged-only]
    B -->|いいえ| D[作業状態の確認]
    D --> E{不要?}
    E -->|はい| F[moai worktree remove PATH]
    E -->|いいえ| G[維持]
    C --> H[整理完了]
    F --> H
    G --> H
```

---

### Q: LLM が期待どおりに動作しません

**A**: Worktree 別の LLM 設定を確認してください:

```bash
# 現在の LLM を確認
moai config
現在の LLM: GLM 5

# Worktree で LLM を変更
cd "$(moai worktree go SPEC-AUTH-001)"
(SPEC-AUTH-001) $ moai cc
→ Claude Opus に変更された

# 他の Worktree は影響なし
(SPEC-AUTH-001) $ exit
cd "$(moai worktree go SPEC-LOG-002)"
(SPEC-LOG-002) $ moai config
現在の LLM: GLM 5 (変更なし)
```

---

### Q: Git コマンドが動作しません

**A**: 正しいディレクトリにいるか確認してください:

```bash
# Worktree ディレクトリの確認
pwd
/path/to/your-project/.moai/worktrees/SPEC-AUTH-001

# Git 状態の確認
git status
On branch feature/SPEC-AUTH-001
nothing to commit, working tree clean

# もし Git エラーが発生したら
git fetch --all
git rebase origin/feature/SPEC-AUTH-001
```

---

## 性能および最適化

### Q: Worktree は性能に影響を与えますか?

**A**: 微々たる影響のみです:

**利点**:

- 各 Worktree が独立しているのでキャッシュ効率的
- Git 作業が速い (ローカルブランチ)
- ファイルシステムキャッシュの活用

**欠点**:

- ディスク容量の使用 (各 Worktree ごとに重複)
- 初期 Worktree 生成時に時間がかかる

**最適化のヒント**:

```bash
# 1. 不要な Worktree の削除
moai worktree clean --merged-only

# 2. Git ガベージコレクション
git gc --aggressive --prune=now

# 3. Worktree の圧縮
git worktree prune
```

---

### Q: いくつの Worktree を生成できますか?

**A**: 理論的には無制限ですが、実際には次の要因が個数を制限します:

**制限要因**:

1. **ディスク容量**: 各 Worktree は約 100MB-1GB を使用
2. **メモリ**: 各 Worktree で開いたセッション
3. **ファイルシステム**: 同時に開けるファイル数

**推奨事項**:

- **小型プロジェクト**: 5-10 個の Worktree
- **中型プロジェクト**: 3-5 個の Worktree
- **大型プロジェクト**: 2-3 個の Worktree

```mermaid
graph TD
    A[Worktree 個数の決定] --> B{プロジェクトの規模?}
    B -->|小型| C[5-10 個]
    B -->|中型| D[3-5 個]
    B -->|大型| E[2-3 個]

    C --> F[ディスク: 500MB-1GB]
    D --> G[ディスク: 1.5GB-2.5GB]
    E --> H[ディスク: 2GB-3GB]
```

---

### Q: Worktree を自動的に整理できますか?

**A**: はい、定期的な整理スクリプトを使えます:

```bash
#!/bin/bash
# clean-worktrees.sh

# base にマージされた Worktree を整理
moai worktree clean --merged-only

# Git ガベージコレクション
cd /path/to/project
git gc --aggressive --prune=now

echo "Worktree 整理完了"
```

**cron ジョブの設定**:

```bash
# 毎週日曜日の深夜 2 時に実行
0 2 * * 0 /path/to/clean-worktrees.sh >> /var/log/worktree-cleanup.log 2>&1
```

---

## チーム協業

### Q: チームで Worktree をどう使いますか?

**A**: 次のようなワークフローを推奨します:

```mermaid
graph TB
    subgraph DevA["開発者 A"]
        A1[Worktree 生成]
        A2[開発]
        A3[完了および PR]
    end

    subgraph DevB["開発者 B"]
        B1[Worktree 生成]
        B2[開発]
        B3[完了および PR]
    end

    subgraph Remote["リモートリポジトリ"]
        R[main ブランチ]
    end

    A1 --> A2 --> A3 --> R
    B1 --> B2 --> B3 --> R
```

**チーム協業ガイド**:

1. **Worktree 命名規則**: `SPEC-{カテゴリ}-{番号}`
2. **定期的な同期**: `git pull origin main`
3. **PR レビュー前に**: ローカルでテストを完了
4. **衝突防止**: 頻繁に `main` と同期

---

### Q: Worktree とリモートリポジトリを同期する方法は?

**A**: 定期的に `git pull` を実行してください:

```bash
# 各 Worktree で同期
cd "$(moai worktree go SPEC-AUTH-001)"
(SPEC-AUTH-001) $ git pull origin main

# またはすべての Worktree を同期
for spec in SPEC-AUTH-001 SPEC-LOG-002 SPEC-API-003; do
    cd "$(moai worktree go $spec)"
    echo "Syncing $spec..."
    git pull origin main
done
```

---

### Q: PR レビュー中に Worktree をどう管理しますか?

**A**: 次の戦略を使ってください:

```bash
# PR 生成前
moai worktree status
# 状態の確認

git log main..feature/SPEC-AUTH-001
# 変更事項の確認

# PR レビュー中
# Worktree を維持 (マージ待ち)

# PR 承認およびマージ後に Worktree 整理
moai worktree done SPEC-AUTH-001 --delete-branch

# PR 拒否後
cd "$(moai worktree go SPEC-AUTH-001)"
# 修正作業を継続
```

---

## 追加の質問

### Q: Worktree を使わずに MoAI-ADK を使えますか?

**A**: はい、可能ですが推奨しません:

```bash
# Worktree なしで使用
> /moai plan "機能の説明"
# Worktree 生成ステップをスキップ

# ただし次の問題が発生:
# 1. すべてのセッションに同じ LLM を適用
# 2. 並列開発が不可
# 3. コンテキスト切替のコスト
```

---

### Q: Worktree をバックアップすべきですか?

**A**: Worktree は Git で管理されるので別途のバックアップは不要です:

```bash
# Worktree は Git の一部
# リモートリポジトリに push すれば自動バックアップ

# 定期的にリモートに push
git push origin feature/SPEC-AUTH-001

# Worktree 消失時の復旧
git fetch origin
git worktree add SPEC-AUTH-001 origin/feature/SPEC-AUTH-001
```

---

## 関連ドキュメント

- [Git Worktree 概要](/ja/worktree/)
- [完全ガイド](/ja/worktree/guide)
- [実際の使用例](/ja/worktree/examples)

## さらにヘルプが必要ですか?

- [GitHub Issues](https://github.com/modu-ai/moai-adk/issues) — バグレポート、機能リクエスト
- [Discord コミュニティ](https://discord.gg/Z7E7Mdc5aN) — リアルタイムのやり取り、ヒント共有
