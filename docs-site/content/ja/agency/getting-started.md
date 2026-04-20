---
title: はじめに
weight: 20
draft: false
---

このガイドでは、AI Agency を初めて使用する場合の流れを説明します。最初のブリーフ作成から、ビルド、フィードバック、進化までの一連のプロセスをご紹介します。

## 前提条件

- Node.js 18+ または Python 3.10+
- Claude API キー（有効な認証情報）
- ターミナル環境（bash / zsh / PowerShell）
- Git バージョン管理（推奨）

## ステップ 1：最初のブリーフを作成

AI Agency はブリーフから始まります。プロジェクトの目標、ターゲットオーディエンス、成功指標を明確に定義してください。

```
プロジェクト: 新規 SaaS プロダクト「TaskFlow」用ランディングページ

目標：
- 見込み客からのメール登録 100 件/月
- プロダクトの複雑さを分かりやすく説明
- リード獲得コスト $5 以下

ターゲット：
- 小規模チーム（5-20 人）のプロジェクト管理者
- テック非ネイティブ層
- 既存ツール（Asana, Monday.com）に不満

トーン：
- プロフェッショナルながらアクセシブル
- 具体的な利益を強調
- 親しみやすいアイコン・イラスト使用
```

{{< callout type="info" >}}
**ブリーフ作成のコツ**
- 3-5 つの主要機能に絞る
- ターゲットペルソナを具体的に描く
- 成功指標は定量的に
{{< /callout >}}

## ステップ 2：ビルド実行

ブリーフを Strategy Agent に渡してビルドプロセスを開始します。

```bash
agency build --brief brief.yaml --output ./taskflow-lp
```

このコマンドは以下を実行します：

1. **Strategy フェーズ** - ブリーフを分析し、コンテンツ構成と配置を決定
2. **Create フェーズ** - Copywriting / Design / Dev エージェントがコンテンツ・ビジュアル・コードを並行生成
3. **Review フェーズ** - 品質チェック・ブランド一貫性確認・UX コンプライアンス验证
4. **Build 完了** - HTML / CSS / JavaScript 成果物を出力

## ステップ 3：ブランドコンテキスト設定

AI Agency はプロジェクト固有のブランド設定を理解します。以下の要素を定義してください：

### カラーパレット

```yaml
primary: "#2563EB"      # TaskFlow ブルー
secondary: "#F59E0B"    # アクセント色
neutral: "#6B7280"      # テキスト色
success: "#10B981"      # 成功インジケータ
```

### タイポグラフィ

```yaml
heading_font: "Inter"     # 見出し
body_font: "Inter"        # 本文
code_font: "Fira Code"    # コード
```

### ボイス & トーン

```yaml
tone: "approachable, confident, practical"
avoid: "overly technical, corporate jargon"
emojis: true             # 使用する
formality: "semi-formal"
```

### ロゴ & アセット

```
logo: ./assets/taskflow-logo.svg
favicon: ./assets/favicon.ico
hero_image: ./assets/hero-mockup.png
```

{{< callout type="warning" >}}
**ブランドコンテキスト確認**
設定後、必ず Strategy Agent が正しく理解したか確認してください。異なる解釈があるとすべての生成物に影響します。
{{< /callout >}}

## ステップ 4：フィードバック提供

生成されたコンテンツを確認し、改善フィードバックを与えます。

```bash
agency feedback --project ./taskflow-lp --comment feedback.md
```

### フィードバック例

```markdown
# ランディングページフィードバック

## 良かった点
- ヒーロー画像が説得力的
- 価格表示がシンプルで理解しやすい

## 改善点
1. CTAボタン「Get Started」→「Start Free Trial」に変更したい
2. フィーチャーセクションの順序を入れ替え（現：機能重視 → 希望：ユーザー利益優先）
3. ソーシャルプルーフ（クライアントロゴ）をもっと目立たせる

## 成功指標への影響
- メール登録が 30 件→50 件へ増加期待
```

フィードバックは Learning Pipeline に自動送信され、エージェントが原因分析を実行します。

## ステップ 5：進化を観察

フィードバックが蓄積されると、AI エージェントは自動的に改善されます。

```
Knowledge Graduation Protocol
1x: 初期試行（フィードバック 1 件）→ エージェントが記録
3x: 再現性確認（同じフィードバック 3 件）→ ヒューリスティック化
5x: パターン確立（5 件で確信）→ ルール化
10x+: 高信頼度ルール → Upstream Sync で moai-adk-go に PR
```

例：「CTA ボタンのテキストがコンバージョン率に影響」というフィードバックが 5 件集まると、AI Agency は以下を学習します：

```
Rule: CTA Button Text Optimization
Trigger: CTAボタン作成時
Action: ユーザー利益を直接表現するテキストを選択
Example: "Get Started" → "Start Free Trial"
Confidence: HIGH (5x達成)
```

## ディレクトリ構造

AI Agency プロジェクトの標準的なディレクトリ構造：

```
taskflow-lp/
├── .agency/
│   ├── brief.yaml              # 初期ブリーフ
│   ├── brand-context.yaml      # ブランド設定
│   └── learning-history.json   # 進化履歴
├── src/
│   ├── components/             # React / Vue コンポーネント
│   ├── styles/                 # CSS / Tailwind
│   └── content/                # マークダウン・テキスト
├── output/
│   ├── index.html              # 生成 HTML
│   ├── styles.css              # 生成 CSS
│   └── script.js               # 生成 JavaScript
├── feedback/
│   ├── 2026-04-01.md          # フィードバック履歴
│   └── metrics.json            # 成功指標トラッキング
└── README.md
```

## よくある質問

**Q: どのくらいの頻度でフィードバックを提供すべき？**

A: 初期段階では週 2-3 回、安定期では月 1-2 回が目安です。蓄積されたフィードバックが多いほど AI エージェントの進化は加速します。

**Q: ブランドコンテキストは後で変更できる？**

A: はい、変更できます。ただし既存の生成物との矛盾が生じる可能性があるため、重要な変更は新プロジェクトとして開始することをお勧めします。

**Q: 生成されたコードをカスタマイズしてもいい？**

A: 推奨しません。カスタマイズは FROZEN ゾーンでのみ許可され、EVOLVABLE ゾーンでの手動編集は AI エージェントの学習を妨げます。

## 次のステップ

- [エージェント & スキル](agents-and-skills) - 6つのエージェントの詳細動作を理解する
- [コマンドリファレンス](command-reference) - 全11コマンドの完全リファレンス
