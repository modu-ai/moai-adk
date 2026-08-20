---
title: moai graph
weight: 16
draft: false
new: true
added_in: "v3.1.1"
---

{{< new-badge v3.1.1 >}}

コードベースの関係をひとつの成果物に集めて**逆方向の問い**に答えるツールです。「このパッケージを直すとどこまで揺れるか」「この SPEC は実際のコードとつながっているか」— こうした問いは grep では答えが出ず、関係が集まっていてはじめて答えられます。

{{< callout type="info" >}}
**一行要約**: `moai graph build` が codemaps·@MX タグ·SPEC·レポートに散らばった関係を `.moai/project/graph/edges.jsonl` の1ファイルに集め、`moai graph query` がそのファイルに逆方向の問合せをします。
{{< /callout >}}

## なぜ必要か

MoAI-ADK は関係情報をすでに複数の層に持っています — codemaps のインポートグラフ、コード内の `@MX:SPEC` タグ、SPEC 文書の依存宣言、レポートのマイルストーン記録。問題は、これらの層がそれぞれ別のファイルに散らばっている点です。「このコードを直すとどの SPEC が影響を受けるか」を問うには、インポート方向(@MX:SPEC タグのあるファイルをインポートするファイル)と SPEC 依存の方向を**同じグラフの中で**逆向きにたどる必要があります。edges.jsonl はそのひとつのグラフです。

## moai graph build

```bash
$ moai graph build
```

インポートエッジ·`@MX:SPEC` 接続·SPEC 間依存を集めて `.moai/project/graph/edges.jsonl` に記録します。同じ git HEAD で2回実行すれば同じ内容が出るよう、決定論的に動作します。問合せは常にこの成果物を読むので、**問合せの前にまず build を回して**おく必要があります。

## moai graph query

1回の呼び出しに与えるセレクタは**ちょうど1つ**だけです。

| セレクタ | 問い | 答 |
|--------|------|-----|
| `--callers <ノード>` | このパッケージ/SPEC を直接依存している対象は? | 逆方向の隣人 — インポートするパッケージ、依存する SPEC、`@MX:SPEC` タグの付いたコードファイル |
| `--blast <ノード>` | ここを直すとどこまで揺れるか? | 逆方向エッジを幅優先で洗った(BFS)影響半径。`@MX:SPEC` エッジは双方向に伝播し、コードファイルが実装する SPEC まで届きます |
| `--fanin [--limit N]` | 最もよく使われるパッケージは? | インポートのファンイン順位 — @MX:DEBT ファンイン問合せの代用品(タグ種別ごとのエッジはまだありません) |
| `--specs-no-code` | コードとつながっていない SPEC は? | edges.jsonl に `@MX:SPEC` エッジが0個の SPEC の一覧 |
| `--milestones-no-card` | カードなしで通過したマイルストーンは? | カード交差検査の行がカードを主張していないか、主張したカードが生存するバックログキューにないマイルストーン |

```bash
$ moai graph query --callers SPEC-FOO-001
$ moai graph query --blast internal/config
$ moai graph query --fanin --limit 20
$ moai graph query --specs-no-code
$ moai graph query --milestones-no-card
```

`--edges <パス>`で別の edges.jsonl を指したり、ルート引数で別のプロジェクトルートを指定したりできます。

## 2つのセレクタへの注意書き

{{< callout type="warning" >}}
{{< icon warning warn >}} **`--specs-no-code`**: 「未接続」は「未実装」ではありません。大部分の SPEC は文書・ルール・ハーネスを出すだけで、コードがなくても完成です。この一覧は欠陥リストではなくカバレッジの地図として読んでください。
{{< /callout >}}

{{< callout type="warning" >}}
{{< icon warning warn >}} **`--milestones-no-card`**: バックログキューはカードが終わると(`done`)行を消します。そのため「キューにない」は**終わったカードと、そもそも発行されなかったカードを一緒に**束ねます。各項目は `git log --oneline --grep 'merge: tNN'` で判定してください — コミットがあれば通過、なければ新しいカードの候補です。grep 0件も「作業をしなかった」という意味ではありません。カードが新しい id で再発行されている可能性があるので、新しいカードを切る前に系譜を確認してください。
{{< /callout >}}

## 関連ドキュメント

- [カンバンモード](/ja/advanced/kanban-mode) — マイルストーンとカードの交差検査が見張るカードの流れ
- [`/moai mx`](/ja/utility-commands/moai-mx) — @MX タグと `@MX:SPEC` 接続の原本
- [Navigator](/ja/core-concepts/navigator) — 設計決定·SPEC·シンボルを束ねるもうひとつのグラフ(nav-graph.json)
