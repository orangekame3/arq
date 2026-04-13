Output docs/demo.gif

Set Shell zsh
Set FontFamily "BlexMono Nerd Font"
Set FontSize 16
Set Width 1400
Set Height 750
Set Theme "Catppuccin Mocha"
Set Padding 30
Set TypingSpeed 40ms
Set CursorBlink false

# Fetch a paper (with auto-summarize enabled)
Type "arq get 2107.13175"
Enter
Sleep 26s

# List papers in table format
Type "arq list"
Enter
Sleep 3.5s

# Show paper details in viewport pager
Type "arq show 2107.13175"
Enter
Sleep 2s

Type "j"
Sleep 0.2s
Type "j"
Sleep 0.2s
Type "j"
Sleep 0.2s
Type "j"
Sleep 0.2s
Type "j"
Sleep 0.2s
Type "j"
Sleep 0.2s
Type "j"
Sleep 0.2s
Type "j"
Sleep 0.2s
Type "j"
Sleep 0.2s
Type "j"
Sleep 0.2s
Type "j"
Sleep 0.2s
Type "j"
Sleep 0.2s
Type "j"
Sleep 0.2s
Type "j"
Sleep 0.2s
Type "j"
Sleep 0.2s
Type "j"
Sleep 0.2s
Type "j"
Sleep 0.2s
Type "j"
Sleep 0.2s
Type "j"
Sleep 0.2s
Type "j"
Sleep 0.2s
Type "j"
Sleep 0.2s
Type "j"
Sleep 0.2s
Type "j"
Sleep 0.2s
Type "j"
Sleep 0.2s
Type "j"
Sleep 0.2s
Type "j"
Sleep 0.2s
Type "j"
Sleep 0.2s
Type "j"
Sleep 0.2s
Type "j"
Sleep 0.2s
Type "j"
Sleep 1s

# Scroll back up with k
Type "k"
Sleep 0.2s
Type "k"
Sleep 0.2s
Type "k"
Sleep 0.2s
Type "k"
Sleep 0.2s
Type "k"
Sleep 0.2s
Type "k"
Sleep 0.2s
Type "k"
Sleep 0.2s
Type "k"
Sleep 0.2s
Type "k"
Sleep 0.2s
Type "k"
Sleep 1s

# Quit pager
Type "q"
Sleep 0.5s

# fzf with preview
Type "arq list --tsv | fzf --with-nth=2.. --preview 'CLICOLOR_FORCE=1 arq show {1}' --preview-window=right:55%:wrap --ansi"
Enter
Sleep 2s

# Quick item switching
Down
Sleep 0.4s
Down
Sleep 0.4s
Down
Sleep 0.4s
Down
Sleep 0.4s
Down
Sleep 0.4s
Down
Sleep 0.4s
Up
Sleep 0.4s
Up
Sleep 0.4s
Up
Sleep 0.4s
Up
Sleep 0.4s
Up
Sleep 1s

# Fuzzy search
Type "coupler"
Sleep 2s
Backspace 7
Sleep 0.5s
Type "fault"
Sleep 2s

Enter
Sleep 2s

# Remove a paper
Type "arq rm 2107.13175"
Enter
Sleep 2s

Sleep 1s
