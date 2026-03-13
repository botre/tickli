# VHS Tape: tickli Smart Views Demo
# Run: vhs demo/smart-views.tape

Output demo/smart-views.gif

Set FontFamily "JetBrains Mono"
Set FontSize 14
Set Width 900
Set Height 500
Set Theme "Catppuccin Mocha"
Set WindowBar Colorful
Set BorderRadius 8
Set Padding 20
Set TypingSpeed 50ms

# Show today's tasks
Type "tickli today"
Enter
Sleep 2s

# Show a task detail
Down
Sleep 300ms
Enter
Sleep 2s

# Clear and show week
Ctrl+C
Sleep 500ms
Type "tickli week"
Enter
Sleep 2s

# Clear and show task details in JSON
Ctrl+C
Sleep 500ms
Type "tickli today --json | head -20"
Enter
Sleep 2s
