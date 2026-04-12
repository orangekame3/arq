# arq

[日本語](README.ja.md)

> [!WARNING]
> This project is experimental. Breaking changes may occur without notice.

[ghq](https://github.com/x-motemen/ghq) for arXiv papers — manage local paper copies and explore them with fzf.

![demo](docs/demo.gif)

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
arq get --force 2303.12345          # re-fetch and overwrite
arq get --open 2303.12345           # open PDF after download

arq list [--tsv|--json|--id]
arq show <query> [--json]
arq path <query>
arq open <query>
arq has <id>
arq select
arq remove <query>                 # remove a paper (alias: rm)
arq thumbnail set <query> <image>  # set thumbnail
arq thumbnail path <query>         # get thumbnail path
arq config
arq version
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

## Translation

Translate title and abstract using an LLM on `arq get`. Supports OpenAI and Anthropic.

```bash
arq config set translate.enabled true
arq config set translate.api_key sk-xxx
arq get 2303.12345                      # auto-translates
arq get --no-translate 2303.12345       # skip for this one
arq show 2303.12345                    # shows both en and ja
```

When a thumbnail is set, `arq show` displays it in terminals that support the Kitty graphics protocol (e.g. Ghostty, Kitty).

## Configuration

```bash
arq config                              # show all
arq config set <key> <value>            # set a value
arq config get <key>                    # get a value
arq config setup                        # interactive setup
```

Available keys:

| Key | Description |
|---|---|
| `root` | Paper storage root directory |
| `translate.enabled` | Auto-translate on get (`true` / `false`) |
| `translate.provider` | `openai` or `anthropic` (default: auto-detect from API key) |
| `translate.model` | Model name (default: `gpt-4o-mini` / `claude-haiku-4-5-20251001`) |
| `translate.lang` | Target language (default: `Japanese`) |
| `translate.api_key` | API key (or use `OPENAI_API_KEY` / `ANTHROPIC_API_KEY` env var) |

Root directory priority:

1. `$ARQ_ROOT`
2. `~/.config/arq/config.toml`
3. `~/papers` (default)

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

arq is inspired by [ghq](https://github.com/x-motemen/ghq) by [@motemen](https://github.com/motemen). The path-centric, host-based directory structure and the fzf-first workflow are directly influenced by ghq's design philosophy.

## License

[MIT](LICENSE)
