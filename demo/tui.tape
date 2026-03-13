# VHS Tape: tickli TUI Dashboard Demo
# Run: vhs demo/tui.tape

Output demo/tui.gif

Set FontFamily "JetBrains Mono"
Set FontSize 14
Set Width 1200
Set Height 700
Set Theme "Catppuccin Mocha"
Set WindowBar Colorful
Set BorderRadius 8
Set Padding 20
Set TypingSpeed 50ms

# Launch TUI
Type "tickli tui"
Enter
Sleep 3s

# Navigate through tasks
Down
Sleep 500ms
Down
Sleep 500ms
Down
Sleep 500ms

# View task details
Enter
Sleep 2s

# Go back
Escape
Sleep 1s

# Switch to Tomorrow view
Type "2"
Sleep 2s

# Switch to Week view
Type "3"
Sleep 2s

# Switch to All view
Type "5"
Sleep 2s

# Open filter
Type "/"
Sleep 500ms
Type "meeting"
Sleep 1s
Escape
Sleep 1s

# Browse projects
Type "p"
Sleep 2s
Down
Sleep 500ms
Enter
Sleep 2s

# Quit
Type "q"
Sleep 500ms
