# VHS Tape: tickli Interactive Task Creation Demo
# Run: vhs demo/task-create.tape

Output demo/task-create.gif

Set FontFamily "JetBrains Mono"
Set FontSize 14
Set Width 900
Set Height 500
Set Theme "Catppuccin Mocha"
Set WindowBar Colorful
Set BorderRadius 8
Set Padding 20
Set TypingSpeed 50ms

# Create a task interactively
Type "tickli task create -i"
Enter
Sleep 1s

# Fill in title
Type "Review quarterly OKR progress"
Enter
Sleep 500ms

# Fill in content
Type "Check alignment with team goals and update dashboard"
Enter
Sleep 500ms

# Select priority (Medium)
Down
Down
Enter
Sleep 500ms

# Fill in date
Type "next friday at 2pm"
Enter
Sleep 500ms

# Fill in tags
Type "work, review"
Enter
Sleep 500ms

# Submit
Enter
Sleep 2s
