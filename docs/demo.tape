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

# Fetch a paper
Type "arq get 2107.13175"
Enter
Sleep 6s
Sleep 2s

# List papers in table format
Type "arq list"
Enter
Sleep 3.5s

# Clear and show paper details
Type "clear"
Enter
Sleep 0.5s
Type "arq show 2107.13175"
Enter
Sleep 4s

# Clear and fzf integration
Type "clear"
Enter
Sleep 0.5s
Type "arq list --tsv | fzf --with-nth=2.. --preview 'CLICOLOR_FORCE=1 arq show {1}' --preview-window=right:50%:wrap --ansi"
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
