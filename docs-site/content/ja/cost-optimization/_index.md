---
title: コスト最適化
weight: 50
draft: false
---

MoAI-ADK トークノミクスは 2 つの軸で推論コストを削減します。**コンテキストダイエット** (Context Diet) が常時ロードされるコンテキスト自体を減らす軸だとすれば、**プロンプトキャッシング** (Prompt Caching) は残ったコンテキストを 90% 割引のコストで再利用する軸です。このセクションではキャッシングをいつ有効化するか判断する損益分岐ルールと設定方法を扱います。

{{< mascot bubble >}}

## このセクションのドキュメント

- [プロンプトキャッシング — 損益分岐分析と実装ガイド](/ja/cost-optimization/prompt-caching) — 2 リクエスト損益分岐ルール、`cacheStrategy` 設定、statusline モニタリング
