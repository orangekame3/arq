# arq

[日本語](README.ja.md)

[ghq](https://github.com/x-motemen/ghq) for arXiv papers — manage local paper copies and explore them with fzf.

## Install

### Homebrew

```bash
brew install orangekame3/tap/arq
```

### Go

```bash
go install github.com/orangekame3/arq@latest
```

## Usage

```bash
arq get 2303.12345
arq get https://arxiv.org/abs/2303.12345
arq get 2303.12345 2401.67890       # multiple IDs
cat ids.txt | arq get -             # batch from stdin

arq list [--tsv|--json|--id]
arq show <query> [--json]
arq path <query>
arq open <query>
arq has <id>
arq select
```

### fzf integration

`arq select` runs fzf internally and outputs the selected paper ID.

```bash
arq open "$(arq select)"
arq show "$(arq select)"
```

For custom fzf usage, pipe `arq list --tsv`:

```bash
arq list --tsv | fzf --with-nth=2.. | cut -f1
```

#### Shell functions

Add to `.zshrc` / `.bashrc`:

```bash
# Open a paper with fzf preview
arqo() {
  local id
  id=$(arq list --tsv | fzf --with-nth=2.. --preview 'arq show {1}' | cut -f1)
  [ -n "$id" ] && arq open "$id"
}

# cd to a paper directory
arqd() {
  local path
  path=$(arq list --tsv | fzf --with-nth=2.. --preview 'arq show {1}' | cut -f1)
  [ -n "$path" ] && cd "$(dirname "$(arq path "$path")")"
}
```

### Query

`<query>` accepts exact ID, partial ID, or title substring match. Errors on multiple hits.

## Directory structure

Like ghq organizes repos under `<root>/<host>/<owner>/<repo>`, arq organizes papers under `<root>/<host>/<category>/<id>`. The category is automatically determined from the arXiv primary_category.

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

## Configuration

Root directory priority:

1. `$ARQ_ROOT`
2. `~/.config/arq/config.toml`
3. `~/papers` (default)

```sh
arq root              # print current root
arq root ~/papers     # set root
```

```toml
# ~/.config/arq/config.toml
root = "/Users/you/papers"
```

## Acknowledgements

arq is inspired by [ghq](https://github.com/x-motemen/ghq) by [@motemen](https://github.com/motemen). The path-centric, host-based directory structure and the fzf-first workflow are directly influenced by ghq's design philosophy.

## License

[MIT](LICENSE)
