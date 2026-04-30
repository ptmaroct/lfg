# vhs tape — record welcome → bundles → tools flow as a gif
# install vhs:  brew install vhs
# render:       vhs demos/welcome.tape
# output:       demos/welcome.gif

Output demos/welcome.gif
Set Shell zsh
Set FontSize 14
Set Width 1200
Set Height 700
Set Theme "Dracula"

Type "./lfg"
Enter
Sleep 1500ms

# Welcome screen: pick install
Sleep 500ms
Enter
Sleep 1s

# Bundle picker: toggle ai-clis on
Down
Down
Sleep 200ms
Type "x"
Sleep 400ms
Enter
Sleep 1s

# Tool picker: scroll, filter, toggle
Down Down Down Down
Sleep 200ms
Type "/"
Type "lazy"
Sleep 600ms
Escape
Sleep 400ms
Enter
Sleep 1500ms

# Confirm
Sleep 600ms
Enter
Sleep 6s

# Done screen
Sleep 1s
Enter
