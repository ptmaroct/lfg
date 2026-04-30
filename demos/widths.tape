# vhs tape — render the welcome screen at every width in the test matrix.
# Useful for eyeballing layout sanity at a glance.
# render: vhs demos/widths.tape  (produces demos/widths.gif)

Output demos/widths.gif
Set Shell zsh
Set FontSize 13
Set Width 1600
Set Height 600

Type "./lfg --theme=lfg"
Enter
Sleep 1500ms

# Resize through each width target. VHS supports `Set Width` mid-tape
# but actual TUI resize requires SIGWINCH; instead we just take frames at
# render-time width. For each target, kill + relaunch via screenshot stage.
# (Simpler approach: use `tools/snap-widths.sh` script for static stills.)

Sleep 3s
