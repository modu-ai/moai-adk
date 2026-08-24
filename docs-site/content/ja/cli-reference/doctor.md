---
title: moai doctor 診断
weight: 60
draft: false
---

`moai doctor` は包括的なシステム診断を実行します。Claude Code の設定、依存関係、プロジェクト構造、言語別の開発ツール、環境を検査し、検出された問題に対する修正案を提示できます。

## 概要

```bash
moai doctor [OPTIONS]
```

## フラグ

| フラグ | 説明 |
|--------|------|
| `-v, --verbose` | 詳細な診断情報 (ツールのバージョン、言語検出) を表示 |
| `--fix` | 検出された問題の修正案を提示 |
| `--export` | 診断結果を JSON ファイルにエクスポート |
| `--check <tool>` | 特定のチェックのみ実行 (例: git, go, config) |

## サブコマンド

`moai doctor` は特定領域を掘り下げるサブコマンドを提供します。

| コマンド | 説明 |
|----------|------|
| `moai doctor config` | 設定診断 — マージされた設定を provenance 付きで検査 |
| `moai doctor hook` | 27 イベントのフックカバレッジ表を表示 |
| `moai doctor permission` | 権限解決の診断 |
| `moai doctor sandbox` | サンドボックスバックエンドの可用性診断 |

`moai doctor config` はさらに `dump` (マージ設定のダンプ) と `diff <tier-a> <tier-b>` (2 つの設定ティアの比較) を提供します。

## Home Disk Usage 診断 {{< new-badge v3.1.1 >}}

`moai doctor` の全体診断には **Home Disk Usage** 項目が並びます。`~/.moai` ホームディレクトリがどれだけ埋まっているかを報告する**勧告 (advisory)** 性格の検査なので、閾値を超えても他のコマンドを止めません。

| 報告項目 | 内容 |
|----------|------|
| 全体サイズ | `~/.moai` の総容量と上位 3 項目 |
| プロファイル別内訳 | `claude-profiles/<プロファイル>` それぞれのサイズとカテゴリ分解 |
| リリース数 | `releases/` に残るバイナリ数と現行バージョン |
| 整理可能量 | `moai clean --home` が実際に削除できる推定バイト数 |
| `~/.claude` | サイズのみ報告 — どの経路でも整理対象ではない |

整理可能量が閾値(コンパイル既定値 500 MB)を超えると状態が WARN に変わり、`moai clean --home`(既定は dry-run)を勧めます。それ未満なら OK のままです。`~/.moai` がそもそも無ければ「報告するものなし」として OK になります。

この推定値は `moai clean --home` が使うのと**同じスキャナ**を呼ぶので、doctor が言う数字と clean が実際に削除する一覧がずれることはありません。詳細は [ホームディレクトリ衛生](/ja/advanced/home-hygiene) にあります。

## 終了コード

スクリプトや CI ラッパーが `moai doctor` を呼ぶとき読むのは、要約の行ではなく終了コードです。

| 終了コード | 意味 |
|------------|------|
| `0` | Fail 項目なし。Warn は勧告なので終了コードを変えません |
| `1` | 一件以上が Fail。要約の `Fail N` がそのまま反映されます |

Constitution Registry の項目は、レジストリが解析できるかを見るだけでなく、`moai constitution validate` と**同じドリフト検証**を実行します。したがって同じチェックアウトで doctor が ok と言い、validate が失敗する、ということは起こりません。`MOAI_CONSTITUTION_SKIP_VALIDATE=1` で迂回すると、doctor は構造検査の判定に戻ります。

## 例

```bash
moai doctor                            # 全体診断
moai doctor --verbose                  # 詳細診断
moai doctor --export diagnostics.json  # 結果をエクスポート
moai doctor hook                       # フックカバレッジ表
```

---

関連: [プロジェクト状態](/ja/cli-reference/status) · [CLI 概要](/ja/getting-started/cli)
