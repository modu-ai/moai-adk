---
title: moai memory
weight: 18
draft: false
new: true
added_in: "v3.1.1"
---

{{< new-badge v3.1.1 >}}

教訓ストア(Lessons Protocol memory store)の**健康診断とアーカイブ**のツールです。MoAI の教訓は `MEMORY.md` 索引とトピックファイル群で構成されますが、セッション開始時に読むのは索引であってディレクトリではありません。そのため索引行なしでファイルだけが書かれた教訓は、**保存はされたのに二度と想起されない**状態になります。

{{< callout type="info" >}}
**一行要約**: `moai memory doctor` が索引↔ファイルの不一致(孤児ファイル・空リンク)とトピックファイル数の上限を診断し、`moai memory archive` が指名したファイルをアーカイブに畳んで索引から下ろします。
{{< /callout >}}

## moai memory doctor

```bash
$ moai memory doctor            # 人が読む報告
$ moai memory doctor --json     # 構造化された出力
$ moai memory doctor --dir <パス>   # 別のストアを診断
$ moai memory doctor --cap 80   # 上限を変えて検査
```

診断項目:

| 項目 | 意味 |
|------|-----|
| 孤児トピックファイル | 索引行のないトピックファイル — 想起されない教訓 |
| 空の索引リンク | 索引行が指すファイルがない — 読み取り失敗として残る行 |
| トピックファイル数 | プロジェクトごとの上限(既定 50)に対する現在の数 |

`--dir` と `--cap` は検査対象・基準を変えるオプションであり、ファイルを直しません。doctor は診断だけを行います。

## moai memory archive

```bash
$ moai memory archive feedback-old-lesson.md
```

指名したトピックファイルを `memory/_archive/` へ移し、索引行を下ろします。**削除ではありません** — アーカイブは監査記録を保存します。何が古びたかは判断の問題なので、対象は運営者がファイル名で直接指名し、自動選別はありません。上限(既定 50 個)を超えたら超過分をアーカイブに畳んで索引を軽く保つのが、このツールの主な用途です。

## 関連ドキュメント

- [コンテキストとメモリ](/ja/claude-code/context-memory/memory) — Claude Code メモリの動作原理
- [自己進化](/ja/advanced/self-evolving) — 教訓がルール昇格提案へ上がる流れ
- [決定メモリ](/ja/advanced/decision-memory) — ルーティング決定が積もるもうひとつの記憶
