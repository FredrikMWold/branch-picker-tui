# branch-picker-tui

[![Go Reference](https://pkg.go.dev/badge/github.com/fredrikmwold/branch-picker-tui.svg)](https://pkg.go.dev/github.com/fredrikmwold/branch-picker-tui)
[![Release](https://img.shields.io/github/v/release/FredrikMWold/branch-picker-tui?sort=semver)](https://github.com/FredrikMWold/branch-picker-tui/releases)

A minimal, keyboard-first TUI for Git branches built with [Bubble Tea](https://github.com/charmbracelet/bubbletea). List, create, checkout, and delete branches.

![Demo](./demo.gif)


<details>
	<summary><strong>Quick keys</strong></summary>

| Context | Key | Action |
|---|---|---|
| Branch picker | `↑`/`↓` or `j`/`k` | Move selection |
| Branch picker | `/` | Filter branches |
| Branch picker | `n` | Create new branch (inline input at top) |
| Branch picker | `d` | Delete selected branch (inline confirm; force if not merged) |
| Branch picker | `Enter` | Checkout selected branch / confirm delete |
| Branch picker | `Esc` | Cancel create/delete |
| Branch picker | `r` | Refresh branches |
| Branch picker | `q` or `Ctrl+C` | Quit |

> Tip: The help footer updates based on what you can do at the moment.

</details>

## Features

- 📋 List local branches with an “Active” indicator for the current branch
- 🌱 Create a brand‑new branch inline from the list
- 🔀 Checkout the selected branch with Enter

## Install

Install with Go:

```sh
go install github.com/fredrikmwold/branch-picker-tui/cmd/branch-picker-tui@latest
```

Or download a prebuilt binary from the Releases page and place it on your PATH:

- https://github.com/FredrikMWold/git-worktree-tui/releases
