# arq

[English](README.md)

arXiv 論文版の [ghq](https://github.com/x-motemen/ghq) — 論文をローカル管理し、fzf で高速に探索する。

## インストール

### Homebrew

```bash
brew install orangekame3/tap/arq
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

arq list [--tsv|--json|--id]
arq show <query> [--json]
arq path <query>
arq open <query>
arq has <id>
arq select
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
    cs.AI/
      2401.67890/
        paper.pdf
        meta.json
```

## 翻訳

`arq get` 時に LLM でタイトルとアブストラクトを翻訳する。OpenAI と Anthropic に対応。

```bash
arq config set translate.enabled true
arq config set translate.api_key sk-xxx
arq get 2303.12345                      # 自動翻訳
arq get --no-translate 2303.12345       # 今回だけスキップ
arq show 2303.12345                    # 英語・日本語の両方を表示
```

fzf プレビューで日本語表示:

```bash
arq list --tsv | fzf --with-nth=2.. --preview 'arq show {1}'
```

## 設定

```bash
arq config                              # 全設定を表示
arq config set <key> <value>            # 値を設定
arq config get <key>                    # 値を取得
arq config setup                        # 対話式セットアップ
```

設定キー:

| キー | 説明 |
|---|---|
| `root` | 論文ストレージのルートディレクトリ |
| `translate.enabled` | get 時に自動翻訳（`true` / `false`） |
| `translate.provider` | `openai` または `anthropic`（デフォルト: API キーから自動検出） |
| `translate.model` | モデル名（デフォルト: `gpt-4o-mini` / `claude-haiku-4-5-20251001`） |
| `translate.lang` | 翻訳先の言語（デフォルト: `Japanese`） |
| `translate.api_key` | API キー（`OPENAI_API_KEY` / `ANTHROPIC_API_KEY` 環境変数も可） |

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
model = "gpt-4o-mini"
lang = "Japanese"
api_key = "sk-xxx"
```

## Acknowledgements

arq は [@motemen](https://github.com/motemen) 氏の [ghq](https://github.com/x-motemen/ghq) にインスパイアされています。パス中心の設計、ホストベースのディレクトリ構造、fzf を前提としたワークフローは ghq の設計思想に大きく影響を受けています。

## ライセンス

[MIT](LICENSE)
