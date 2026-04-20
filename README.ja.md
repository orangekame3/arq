<p align="center">
  <img src="docs/logo.png" alt="arq logo" width="200">
</p>

# arq

[English](README.md)

> [!WARNING]
> このプロジェクトは実験的です。予告なく破壊的変更が行われる可能性があります。

ターミナルで arXiv 論文を管理 — 取得・LLM 要約・fzf 探索。[ghq](https://github.com/x-motemen/ghq) にインスパイアされたツール。

![demo](docs/demo.gif)

## インストール

```bash
curl -fsSL https://raw.githubusercontent.com/orangekame3/arq/main/scripts/install.sh | sh
```

### Homebrew

```bash
brew install orangekame3/tap/arq
```

### Shell installer

```bash
curl -fsSL https://raw.githubusercontent.com/orangekame3/arq/main/scripts/install.sh | sh
```

デフォルトでは `~/.local/bin` にインストールされます。`ARQ_INSTALL_DIR` で変更可能です:

```bash
ARQ_INSTALL_DIR=/usr/local/bin curl -fsSL https://raw.githubusercontent.com/orangekame3/arq/main/scripts/install.sh | sh
```

スクリプトを確認してから実行したい場合:

```bash
curl -fsSL -o install-arq.sh https://raw.githubusercontent.com/orangekame3/arq/main/scripts/install.sh
sh install-arq.sh
```

### Go

```bash
go install github.com/orangekame3/arq@latest
```

## 使い方

```bash
arq get 2303.12345
arq get https://arxiv.org/abs/2303.12345
arq get 2303.12345 2401.67890       # 複数ID指定
cat ids.txt | arq get -             # stdin から一括取得
arq get --force 2303.12345          # 再取得して上書き
arq get --open 2303.12345           # ダウンロード後に PDF を開く
arq get --translate 2303.12345      # タイトルとアブストラクトを翻訳
arq get --summarize 2303.12345      # 図表付き要約を生成

arq list [--tsv|--json|--id]
arq show <query> [--json|--summary]
arq summarize <query> [--force]    # 要約を生成/再生成 (alias: sum)
arq summarize --all [--force]      # 全論文を一括要約
arq search <keyword> [keyword...]  # ローカル論文を検索
arq view <query> [-t group]        # サマリーを mo（ブラウザ）で開く
arq path <query>
arq open <query>
arq has <id> [...]                 # 1つ以上のIDを存在チェック
arq has -                          # stdin からIDを読み込んでチェック
arq select
arq remove <query>                 # 論文を削除 (alias: rm)
arq thumbnail set <query> <image>  # サムネイル設定
arq thumbnail path <query>         # サムネイルパス取得
arq config
arq version
```

### fzf 連携

`arq select` は内部で fzf を起動し、選択した論文の ID を出力する。

```bash
arq open "$(arq select)"
arq show "$(arq select)"
```

`arq list --tsv` をパイプして自由にカスタマイズもできる:

```bash
arq list --tsv | fzf --with-nth=2.. | cut -f1
```

#### シェル関数

`.zshrc` / `.bashrc` に追加:

```bash
# fzf プレビュー付きで論文を開く
arqo() {
  local id
  id=$(arq list --tsv | fzf --with-nth=2.. --preview 'arq show {1}' | cut -f1)
  [ -n "$id" ] && arq open "$id"
}

# 論文ディレクトリに cd
arqd() {
  local path
  path=$(arq list --tsv | fzf --with-nth=2.. --preview 'arq show {1}' | cut -f1)
  [ -n "$path" ] && cd "$(dirname "$(arq path "$path")")"
}

# fzf プレビューで要約を表示
arqs() {
  local id
  id=$(arq list --tsv | fzf --with-nth=2.. --preview 'arq show --summary {1}' | cut -f1)
  [ -n "$id" ] && arq show --summary "$id"
}
```

### クエリ

`<query>` は完全一致ID・部分一致ID・タイトル部分一致を許容する。複数ヒット時はエラー。

## ディレクトリ構造

ghq がリポジトリを `<root>/<host>/<owner>/<repo>` で管理するように、arq は論文を `<root>/<host>/<category>/<id>` で管理する。カテゴリは arXiv の primary_category から自動判定。

```bash
# ghq
~/src/github.com/orangekame3/arq/

# arq
~/papers/arxiv.org/quant-ph/2303.12345/paper.pdf
```

```bash
$ARQ_ROOT/
  arxiv.org/
    quant-ph/
      2303.12345/
        paper.pdf
        meta.json
        summary.md        # arq summarize で生成
        assets/            # ar5iv からダウンロードした図
          x1.png
          x2.png
    cs.AI/
      2401.67890/
        paper.pdf
        meta.json
```

## 要約

LLM を使って論文のマークダウン要約を生成する。[ar5iv](https://ar5iv.labs.arxiv.org/) から全文 HTML を取得し、図表やセクション構造を含むリッチなコンテンツ分析を行う。ar5iv が利用できない場合はアブストラクトベースの要約にフォールバック。

```bash
arq summarize 2303.12345              # 要約を生成
arq summarize --force 2303.12345      # 再生成
arq summarize --all                   # 要約がない全論文を一括要約
arq summarize --all --force           # 全論文の要約を再生成
arq show 2303.12345                   # glamour で要約をレンダリング表示
arq show --summary 2303.12345         # 要約のみ表示
```

`arq get` 時に自動要約:

```bash
arq get --summarize 2303.12345         # 今回だけ
arq config set summarize.enabled true  # 常に自動要約
arq get --no-summarize 2303.12345      # 今回だけスキップ
```

生成される `summary.md` には以下が含まれる:
- 論文メタデータ（タイトル、著者、arXiv リンク）
- 構造化された要約（概要、背景、手法、結果、意義）
- ar5iv からダウンロードした図の埋め込み（`![caption](assets/x1.png)`）

### カスタムプロンプト

config でデフォルトの要約プロンプトを上書きできる:

```toml
[summarize]
prompt = """You are a quantum computing expert. Analyze the following paper in {{lang}}.
..."""
```

`{{lang}}` プレースホルダは実行時に設定言語に置換される。

## ブラウザ表示（mo 連携）

論文のサマリーを [mo](https://github.com/k1LoW/mo) でブラウザ表示する。KaTeX 数式レンダリング、シンタックスハイライト、画像表示に対応。

```bash
arq view 2303.12345                   # サマリーを mo で開く（デフォルトグループ: "arq"）
arq view 2303.12345 --target reads    # 名前付きグループに整理
arq view "$(arq select)"             # fzf で選択 → 表示
```

mo サーバーが既に起動している場合、既存セッションにファイルが追加される。[mo](https://github.com/k1LoW/mo)（`brew install k1LoW/tap/mo`）が必要。mo なしでターミナル表示する場合は `arq show --summary` を使う。

## 検索

ローカルに保存された論文のタイトル・アブストラクト・要約を横断検索する。

```bash
arq search "surface code"                     # 全フィールドを検索
arq search "quantum" "calibration"            # AND 検索（全キーワードが一致する必要あり）
arq search --field summary "decoder"          # 要約のみ検索
arq search --field title "transmon"           # タイトルのみ検索
arq search --json "surface code"              # JSON 出力
arq search --id "surface code"               # ID のみ出力（パイプ用）
```

他のコマンドと組み合わせ:

```bash
arq search --id "calibration" | arq get -     # 一致する論文を再取得
arq show "$(arq search --id "decoder" | head -1)"
```

## 一括存在チェック

複数の論文を一度にチェックする。バッチモードでは、見つかった ID が stdout に出力される。

```bash
arq has 2303.12345 2401.67890          # 複数ID を一度にチェック
cat ids.txt | arq has -                # stdin から読み込んでチェック
arq has 2303.12345                     # 単一ID: exit 0/1、出力なし
```

全 ID が見つかれば exit 0、1つでも見つからなければ exit 1。

## 翻訳

`arq get` 時に LLM でタイトルとアブストラクトを翻訳する。OpenAI、Anthropic、OpenRouter に対応。

```bash
arq config set translate.enabled true
arq config set translate.api_key sk-xxx
arq get 2303.12345                      # 自動翻訳
arq get --no-translate 2303.12345       # 今回だけスキップ
arq show 2303.12345                    # 英語・日本語の両方を表示
```

サムネイルが設定されている場合、`arq show` は Kitty graphics protocol 対応ターミナル（Ghostty, Kitty 等）で画像を表示する。

## 設定

```bash
arq config                              # 全設定を表示
arq config set <key> <value>            # 値を設定
arq config get <key>                    # 値を取得
arq config setup                        # 対話式セットアップウィザード
```

`arq config setup` は [huh](https://github.com/charmbracelet/huh) ベースの TUI ウィザードで、プロバイダ選択・モデルピッカー・プロンプトエディタを提供する。

設定キー:

| キー | 説明 |
|---|---|
| `root` | 論文ストレージのルートディレクトリ |
| `translate.enabled` | get 時に自動翻訳（`true` / `false`） |
| `translate.provider` | `openai`、`anthropic`、`openrouter`（デフォルト: 自動検出） |
| `translate.model` | モデル名（デフォルト: `gpt-5.4-mini`） |
| `translate.lang` | 翻訳先の言語（デフォルト: `Japanese`） |
| `translate.api_key` | API キー（環境変数も可） |
| `summarize.enabled` | get 時に自動要約（`true` / `false`） |
| `summarize.provider` | プロバイダ（未設定時は translate から継承） |
| `summarize.model` | モデル（未設定時は translate から継承） |
| `summarize.lang` | 言語（未設定時は translate から継承） |
| `summarize.api_key` | API キー（未設定時は translate から継承） |
| `summarize.prompt` | カスタム指示プロンプト（`{{lang}}` プレースホルダ） |

root ディレクトリの優先順位:

1. `$ARQ_ROOT`
2. `~/.config/arq/config.toml`
3. `~/papers`（デフォルト）

```toml
# ~/.config/arq/config.toml
root = "/Users/you/papers"

[translate]
enabled = true
provider = "openai"
model = "gpt-5.4-mini"
lang = "Japanese"

[summarize]
enabled = true
# provider, model, lang は未設定時 [translate] から継承
```

### 対応プロバイダ

| プロバイダ | 環境変数 | モデル |
|---|---|---|
| OpenAI | `OPENAI_API_KEY` | GPT-5.4, GPT-4.1, o4-mini, o3, ... |
| Anthropic | `ANTHROPIC_API_KEY` | Claude Opus 4.6, Sonnet 4.6, Haiku 4.5, ... |
| OpenRouter | `OPENROUTER_API_KEY` | 上記すべて + Gemini, Llama, ... |

LLM 連携には [fantasy](https://github.com/charmbracelet/fantasy) を使用。

## Acknowledgements

arq は [@motemen](https://github.com/motemen) 氏の [ghq](https://github.com/x-motemen/ghq) にインスパイアされています。パス中心の設計、ホストベースのディレクトリ構造、fzf を前提としたワークフローは ghq の設計思想に大きく影響を受けています。

[charmbracelet](https://github.com/charmbracelet) ツール群で構築: [lipgloss](https://github.com/charmbracelet/lipgloss)、[glamour](https://github.com/charmbracelet/glamour)、[huh](https://github.com/charmbracelet/huh)、[fantasy](https://github.com/charmbracelet/fantasy)。

## ライセンス

[MIT](LICENSE)
