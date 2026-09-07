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

## Hook Delivery 診断 {{< new-badge v3.1.4 >}}

`moai update` は、テンプレートが新しく追加したフック項目を既存プロジェクトの `.claude/settings.json` に入れられないまま、黙って通り過ぎることがあります。**Hook Delivery** 検査はこうした欠落を見つけます。プロジェクトがすでに持っているフックイベントキーの範囲で、配布テンプレートにはあるのにプロジェクト設定にはない項目を比較して報告します。

| 報告項目 | 内容 |
|----------|------|
| 欠落項目 | `hooks.PreToolUse missing handle-pre-tool.sh (matcher AskUserQuestion)` の形で、イベントキーとマッチャーまで伝えます |
| 修正方法 | 名前が指定されたイベントキーの下に、使用中の moai バージョンのテンプレート設定から該当ブロックをコピーして追加するよう案内します |
| 更新後の確認 | `moai update` の後に管理ファイルが削除されていないか確認するコマンド(`git status --porcelain \| grep '^ D'`)を提示します — 引っかかったら `git restore -- <パス>` で戻してから項目を再追加します |

この検査は読み取り専用です — `.claude/settings.json` を決して書きません。テンプレートが初めて提供するイベントキーとユーザーが自分で作った項目は欠落として数えず、opt-out された項目も要求しません(テンプレートはプロジェクトの `hook.opt_in.enabled` 設定を反映してレンダリングされます)。すべての項目が一致すれば `ok` を報告します。

## Codex Wiring 診断 {{< new-badge v3.1.4 >}}

`moai doctor` の全体診断には **Codex Wiring** 項目が並びます。プロジェクトの Codex 配線(生成されたフック、MCP 登録、スキルミラー)が壊れていないかを調べる勧告 (advisory) 性格の検査で、fail-open かつ読み取り専用です。見つけた問題は報告するだけで、決して自分で修復しません。

| 検査項目 | 内容 |
|----------|------|
| `.codex/hooks.json` の存在・キーホワイトリスト | 配線ファイルがあるか、キーがホワイトリストに合致するかを見ます。キーが一つでも外れると codex はファイル全体を黙って無視するため、この検査がその沈黙を代わりに観測します |
| サイドカーハッシュ | 配備時に記録しておいたフック内容のハッシュと現在のファイルを比較し、フックを手で編集した後に残るずれを捉えます |
| `moai` バイナリの PATH | 生成されるフックコマンドは `moai hook ...` の形なので、PATH 上で `moai` を見つけられなければ何も発火しません |
| `.codex/config.toml` の `[mcp_servers.moai]` | MCP 登録テーブルの有無と、標準的な登録形式との一致を見ます。このテーブルはユーザー所有なので、doctor は報告だけで修復しません |
| `.agents/skills` スキルミラー | Codex CLI は `.claude/skills` をスキャンしないため、ミラーが無い、またはリンクが切れていると codex はこのプロジェクトの MoAI スキルを見られません |
| ユーザー層の `[[skills.config]]` 登録 | `~/.codex/config.toml`(`$CODEX_HOME` 設定時はそのファイル)のスキル登録について、パスと `enabled` キーの形式を見ます。この項目は codex が関与する場合、つまりプロジェクトが配線済みか codex がインストールされている場合にだけ検査します |

検査が問題を見つけると、コードに組み込まれた修正指示文が併せて表示されます。

| 発見 | 指示文 |
|------|--------|
| 配線が全く無いプロジェクト | `moai init --agent codex` |
| フック変更後のサイドカーずれ | `codex /hooks` で変更されたフックを再信頼 |
| スキルミラー欠如・切れたリンク | `moai update --templates-only --force --yes` |
| 参照先のスキルファイルが消えた登録 | 項目を削除するか、スキルファイルを復元 |

この検査が出す発見のうち、fatal に分類されるのは一つだけです。ユーザー層設定の `[[skills.config]]` 項目に `enabled` キーが無いか、bare TOML ブーリアンでない値が入っていると、codex はすべての呼び出しで exit 1 で終了します。この唯一の fatal 発見に引っかかると doctor の結果が Fail になり、終了コード 1 につながります。ただしこの動作は codex-cli 0.153.4 で観測されたもので、そのリリースで確認されているにすぎず、他のリリースにまで一般化されるものではありません。正しい形式は、すべての項目に `enabled = true` または `enabled = false` を明示することです。この一つを除くすべての発見は勧告型なので、doctor の終了コードを変えません。

codex がインストールされていないマシンで、配線の無い claude-only プロジェクトであれば、この検査は黙ってスキップします — 情報提供としてのスキップであり、警告行を作りません。参考までに、存在しなくなったスキルファイルを指すゴースト (ghost) 登録を一括で回収する機能は、この検査の指示文には含まれず、別の動詞である `moai clean --codex-skills` が担います。

フック信頼モデルとスキルミラーを含む Codex 配線の全体は、[Codex デュアルハーネス](/ja/advanced/codex-dual-harness) のドキュメントで詳しく解説されています。

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
