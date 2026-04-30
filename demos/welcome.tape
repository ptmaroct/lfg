# vhs tape — record full lfg flow as gif (welcome → tree → confirm → done)
# install vhs:  brew install vhs
# render:       vhs demos/welcome.tape
# output:       demos/welcome.gif

Output demos/welcome.gif
Set Shell zsh
Set FontSize 14
Set Width 1200
Set Height 720
Set Theme "Dracula"
Set Padding 20

Type "./lfg"
Enter
Sleep 2.5s

# Welcome — let logo gradient sweep for a beat
Sleep 1.5s

# Pick install recommended setup
Enter
Sleep 1s

# Tree picker — expand DEFAULT, scroll a bit, select stuff
Right
Sleep 800ms
Down Down Down
Sleep 400ms
Space
Sleep 200ms
Down Down
Sleep 200ms
Space
Sleep 600ms
Left
Sleep 600ms
Down
Right
Sleep 500ms
Down Down
Sleep 200ms
Space
Sleep 400ms
Enter
Sleep 1.2s

# Confirm — install
Type "y"
Sleep 8s

# Done
Enter
