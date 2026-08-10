---
title: ナビゲータ結合トークン
weight: 25
draft: false
---
# ナビゲータ結合トークン

コードとドキュメントが互いを指し示すようにすると、エージェントが一方を修正したときに他方の文脈を直接引き出せるようになります。**ナビゲータ結合トークン** (Navigator Binding Tokens) は、設計決定・コードシンボル・SPECを1つのアドレス可能なグラフに繋ぐ作成用トークン3つです。これらのトークンが集まって`.moai/project/navigator/nav-graph.json`という単一のアーティファクトになります。

## 3つのトークン

ナビゲータ統合レイヤーは、3つの結合トークンファミリーを1つのグラフに統合します。

| ファミリー | トークン形式 | 作成箇所 | 役割 |
|------|----------|-----------|------|
| `NAV:DEC` | `@NAV:DEC-<id>` | 設計ドキュメント (`.moai/project/*.md`, `.moai/docs/**/*.md`) | 設計決定をSPECやシンボルに接続 |
| `NAV:SYM` | `@NAV:SYM:<symbol>` | コードコメント + 設計ドキュメント | ドキュメント箇所をコードの名前付きシンボルに接続 |
| `MX:SPEC` | `@MX:SPEC:<SPEC-ID>` | コードコメント（`@MX:`タグの下位行） | コード箇所をSPECに接続 |

`MX:SPEC`はすでに[MXタグシステム](/ja/advanced/mx-tags/)が扱っています。ナビゲータ統合レイヤーはMXスキャナーの`SpecAssociator`出力を**消費**するだけで再スキャンしません。したがって、このトークンを新規に作成せず、既存のMXタグルールに従ってください。

## トークンをいつ作成するか

### `@NAV:DEC-<id>` を作成するとき

- `.moai/project/tech.md`、`structure.md`、`product.md`や`.moai/docs/`以下の設計ドキュメントの決定が特定のSPECやコードシンボルに対応するとき。
- 後でコードを修正するときにその決定の文脈が再び浮かび上がることを期待するとき。

### `@NAV:SYM:<symbol>` を作成するとき

- ドキュメント箇所やコードコメントが名前付きコードシンボルに結びついている必要があり、グラフを読む人がドキュメントからコードへ（またはシンボルからシンボルへ）移動できるようにしたいとき。

`@MX:SPEC:`はここでは作成しません。すでにmx-scannerの表面です。再作成は不要です。

## トークン文法

両方のトークンに空値を入れてはなりません。スキャナーは空値に出会うと診断警告を`.moai/logs/navigator-sync.log`に残し、その項目をスキップしますが、グラフビルド全体を停止はしません（fail-open）。

### `@NAV:DEC-<id>`

`<id>`は`[A-Z][A-Z0-9-]*`でなければなりません。大文字のASCIIと数字、そして内部ハイフンのみ許可されます。SPEC-IDドメイントークンとの整合的な規則です。`@NAV:DEC-`接頭辞が明確な識別子なので、id自体は接頭辞なしで登場しません。

### `@NAV:SYM:<symbol>`

`<symbol>`は`[A-Za-z_][A-Za-z0-9_.]*`でなければなりません。識別子の形状であればよく、言語中立です。パッケージ修飾形（`pkg.ParseHeader`）が規約であり、短縮形（`ParseHeader`）も受け入れられ、既存シンボル集合に対する接尾辞一致として解決されます。

## スキャンルート

ナビゲータ統合レイヤーは以下の表面をスキャンします。

- **設計ドキュメント** — `.moai/project/{product,structure,tech}.md`と`.moai/docs/**/*.md`。
- **コード**（`@NAV:SYM`のみ） — `*_test.go`と`vendor/`を除外したGo `*.go`ファイル。設計ドキュメント表面も同時に。

以下はスキャンし**ません**。

- `.moai/specs/` — すでにmx-scannerが本体ベースの連携でカバーしています。
- `.moai/reports/`、`.moai/state/` — 一時的または実行時点の状態。
- 既存の3つのナビゲータチェインのソースコード（消費のみ）。

## アーティファクト — `nav-graph.json`

`.moai/project/navigator/nav-graph.json`の1つのファイルになります。形状は以下の通りです。

```json
{
  "provenance": { "extract_commit_sha": "...", "captured_at": "..." },
  "nodes": [
    { "entity_type": "decision", "identifier": "...", "display_name": "..." }
  ],
  "edges": [
    { "edge_type": "dec-edge", "source_node": "...", "target_node": "...", "source_path": "...", "line_number": 0 }
  ]
}
```

`entity_type`は`decision | spec | symbol`の3つのうち1つ、`edge_type`は`dec-edge | spec-edge | sym-edge`の3つのうち1つです。

このアーティファクトは**バイト安定**です。同じgit HEADで2回実行するとバイト単位で同じ結果が出ます。壁時計タイムスタンプを刻まないため、誰がいつ実行したかとは無関係に結果が同じになります。監査と再現可能性がこの性質の上に成り立ちます。

{{< callout type="info" >}}
**fail-open** — グラフビルドは常に終了コード0を出します。誤ったトークンがあっても停止せず、診断警告だけを残して健全な部分のグラフを作ります。
{{< /callout >}}

## 作成例

設計ドキュメントで決定とシンボルを指し示し、コードコメントで同じ決定とシンボルを受け取る最も単純な形です。

設計ドキュメント (`tech.md`):

```markdown
# Tech

セッションレイヤーは委譲アプローチのためにOAuth2を採用する。

決定 @NAV:DEC-AUTH-STRATEGY: クライアント資格証明(client-credentials)方式のOAuth2。

ヘッダーパーサー (see @NAV:SYM:pkg.ParseHeader) がBearerトークンを抽出する。
```

コード (`auth/auth.go`):

```go
package auth

// @NAV:DEC-AUTH-STRATEGY: OAuth2 client-credentialsフローを実装する。
// @NAV:SYM:auth.ParseBearer がAuthorizationヘッダーからBearerトークンを抽出する。
func ParseBearer(h string) string { ... }
```

この2つのファイルからグラフは3つのノード（決定`AUTH-STRATEGY`、シンボル`pkg.ParseHeader`、シンボル`auth.ParseBearer`）とその間のエッジを作ります。グラフを読む人は設計ドキュメントからコードへ、コードから設計根拠へ自由に行き来できます。

## 前方互換性

トークン文法、結合レコードの5フィールド形状、グラフスキーマはすべて前方向互換（追加のみ）です。後続するマイルストーンはフィールドを追加できますが、既存フィールドの名前と形状は変更しません。一度作ったトークンは長期的に有効です。

## 関連ドキュメント

- [MXタグシステム](/ja/advanced/mx-tags/) — `@MX:SPEC`トークンの原則ルール。ナビゲータ統合レイヤーはこの出力を消費します。
- [SPECベース開発](/ja/core-concepts/spec-based-dev/) — SPECライフサイクルと`@MX:SPEC`の上位文脈。
- [エージェントガイド](/ja/advanced/agent-guide/) — エージェントがコードコメントと設計ドキュメントをどう行き来するか。
