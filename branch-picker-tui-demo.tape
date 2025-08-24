Output demo.gif

# Terminal appearance and environment
Set FontSize 24
Set FontFamily "MesloLGS NF"
Set Theme "Catppuccin Mocha"
Set Width 1800
Set Height 900
Set Padding 10
Set Margin 30
Set MarginFill "#674EFF"
Set BorderRadius 12
Set Shell zsh

# Build the app and prepare a temporary Git repo with sample branches
Hide
Type "go build -o branch-picker-tui ./cmd/branch-picker-tui"
Enter
Type `PROJ="$PWD"`
Enter
Type `TMP="$(mktemp -d -t branch-picker-tui-demo)"`
Enter
Type `cd "$TMP"`
Enter
Type "git init -q && git config user.name Demo && git config user.email demo@example.com"
Enter
Type "git branch feature/alpha && git branch feature/beta"
Enter
Type `alias branch-picker-tui="$PROJ/branch-picker-tui" && clear`
Enter
Show

# Launch the TUI in the demo repo
Type "branch-picker-tui"
Enter
Sleep 2200ms

# Navigate the branch list
Down
Sleep 700ms
Down
Sleep 700ms
Up
Sleep 700ms

# Delete a non-current branch (feature/alpha). First Enter confirms; if not fully merged, second Enter forces.
Down
Sleep 700ms
Type "d"
Sleep 1s
Enter
Sleep 1700ms

# Create a brand-new branch via inline input and auto-checkout
Type "n"
Sleep 700ms
Type "demo-branch"
Sleep 700ms
Enter
Sleep 1700ms

# Done
Sleep 600ms
