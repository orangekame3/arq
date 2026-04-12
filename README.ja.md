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

arq list [--tsv|--json]
arq path <query>
arq open <query>
arq show <query>
arq select
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

## 設定

root ディレクトリの優先順位:

1. `$ARQ_ROOT`
2. `~/.config/arq/config.toml`
3. `~/papers`（デフォルト）

```sh
arq root              # 現在の root を表示
arq root ~/papers     # root を設定
```

```toml
# ~/.config/arq/config.toml
root = "/Users/you/papers"
```

## Acknowledgements

arq は [@motemen](https://github.com/motemen) 氏の [ghq](https://github.com/x-motemen/ghq) にインスパイアされています。パス中心の設計、ホストベースのディレクトリ構造、fzf を前提としたワークフローは ghq の設計思想に大きく影響を受けています。

## ライセンス

[MIT](LICENSE)
