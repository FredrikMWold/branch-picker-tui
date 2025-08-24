package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fredrikmwold/branch-picker-tui/internal/git"
	"github.com/fredrikmwold/branch-picker-tui/internal/theme"
)

type item struct{ title, desc string }

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type state int

const (
	stateList state = iota
)

type model struct {
	state state
	list  list.Model
	input textinput.Model
	cur   string
	frame lipgloss.Style
	// delegate that can toggle inline editing for the add item
	branchDel *branchDelegate
	// deletion confirmation state
	confirmDelete  bool
	forceOnConfirm bool
	// no extra fields; built-in list filtering is used
}

type loadedBranchesMsg struct {
	branches []git.Branch
	err      error
}

func NewProgram() *tea.Program { return tea.NewProgram(initialModel()) }

func initialModel() model {
	// Build input first so delegate can reference it
	in := textinput.New()
	in.Placeholder = "new-branch-name"
	in.Prompt = ""
	in.CharLimit = 64
	in.Cursor.Style = lipgloss.NewStyle().Foreground(theme.Mauve)
	in.TextStyle = lipgloss.NewStyle().Foreground(theme.Text)
	in.PlaceholderStyle = lipgloss.NewStyle().Foreground(theme.Surface2)

	base := list.NewDefaultDelegate()
	base.Styles.SelectedTitle = base.Styles.SelectedTitle.BorderLeftForeground(theme.Mauve).Foreground(theme.Mauve)
	base.Styles.SelectedDesc = base.Styles.SelectedDesc.BorderLeftForeground(theme.Mauve).Foreground(theme.Surface2)
	base.Styles.NormalTitle = base.Styles.NormalTitle.Foreground(theme.Text)
	base.Styles.NormalDesc = base.Styles.NormalDesc.Foreground(theme.Surface2)
	del := &branchDelegate{base: base, input: &in}

	l := list.New(nil, del, 0, 0)
	l.Title = "Git Branches"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(true)
	s := l.Styles
	// Title and titlebar styling
	s.Title = s.Title.Background(theme.Lavender).Foreground(theme.Crust).Bold(true)
	// Make the filter heading (prompt + cursor) use the same fg/bg as the title
	s.FilterPrompt = s.FilterPrompt.Foreground(theme.Crust).Background(theme.Lavender)
	s.FilterCursor = s.FilterCursor.Foreground(theme.Crust).Background(theme.Lavender)
	l.Styles = s
	// Align the filter input styles with the titlebar colors as well
	l.FilterInput.PromptStyle = l.FilterInput.PromptStyle.Foreground(theme.Crust).Background(theme.Lavender).Bold(true).PaddingLeft(1).MarginRight(1)
	l.FilterInput.TextStyle = l.FilterInput.TextStyle.Foreground(theme.Text)
	l.FilterInput.PlaceholderStyle = l.FilterInput.PlaceholderStyle.Foreground(theme.Crust).Faint(true).Background(theme.Lavender)
	l.FilterInput.Cursor.Style = l.FilterInput.Cursor.Style.Foreground(theme.Text).Background(theme.Lavender)
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "checkout / select")),
			key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
			key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new branch")),
			key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete branch")),
			key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel edit")),
			key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
		}
	}

	m := model{state: stateList, list: l, input: in, branchDel: del}
	m.frame = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(theme.Mauve)
	return m
}

func (m model) Init() tea.Cmd { return tea.Batch(loadBranches, tea.EnterAltScreen) }

func loadBranches() tea.Msg {
	brs, err := git.ListBranches()
	return loadedBranchesMsg{branches: brs, err: err}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		w := msg.Width - 4
		h := msg.Height - 2
		if w < 0 {
			w = 0
		}
		if h < 0 {
			h = 0
		}
		m.frame = m.frame.Width(msg.Width - 2)
		m.list.SetSize(w, h)
		// Size input for inline editing width; actual rendering uses list title updated
		iw := w - 6
		if iw < 10 {
			iw = 10
		}
		m.input.Width = iw
		return m, nil
	case loadedBranchesMsg:
		if msg.err != nil {
			return m, m.list.NewStatusMessage(fmt.Sprintf("Error: %v", msg.err))
		}
		items := make([]list.Item, 0, len(msg.branches)+1)
		items = append(items, item{title: "[+] Create new branch", desc: "Type a new branch name"})
		var cur string
		for _, b := range msg.branches {
			t := b.Name
			if b.Current {
				cur = b.Name
			}
			desc := "local branch"
			if b.Current {
				desc = lipgloss.NewStyle().Foreground(theme.Green).Render("Active")
			}
			items = append(items, item{title: t, desc: desc})
		}
		m.cur = cur
		m.list.SetItems(items)
		return m, nil
	case tea.KeyMsg:
		k := msg.String()
		if k == "ctrl+c" {
			return m, tea.Quit
		}
		switch m.state {
		case stateList:
			// While editing inline, route keys to input and update the first item
			if m.branchDel != nil && m.branchDel.editing {
				switch k {
				case "esc":
					m.branchDel.editing = false
					m.input.Blur()
					m.resetAddItemTitle()
					return m, nil
				case "enter":
					name := strings.TrimSpace(m.input.Value())
					if name == "" {
						return m, nil
					}
					if err := git.CreateBranch(name); err != nil {
						return m, m.list.NewStatusMessage(fmt.Sprintf("Error: %v", err))
					}
					if err := git.Checkout(name); err != nil {
						return m, m.list.NewStatusMessage(fmt.Sprintf("Error: %v", err))
					}
					return m, tea.Quit
				}
				var cmd tea.Cmd
				m.input, cmd = m.input.Update(msg)
				m.updateAddItemTitle(m.input.Value())
				// stay on the add item
				m.list.Select(0)
				return m, cmd
			}

			// Always let the list process the key first to avoid swallowing filter input
			var listCmd tea.Cmd
			m.list, listCmd = m.list.Update(msg)
			// If list is filtering, don't intercept shortcuts; only handle Enter selection
			if m.list.FilterState() == list.Filtering {
				if k == "enter" {
					if it, ok := m.list.SelectedItem().(item); ok {
						if strings.HasPrefix(it.title, "[+]") {
							if m.branchDel != nil {
								m.branchDel.editing = true
								m.input.SetValue("")
								m.input.Focus()
								m.list.Select(0)
								m.updateAddItemTitle("")
							}
							return m, nil
						}
						if err := git.Checkout(it.title); err != nil {
							return m, m.list.NewStatusMessage(fmt.Sprintf("Error: %v", err))
						}
						return m, tea.Quit
					}
				}
				return m, listCmd
			}

			// If in delete confirmation mode, handle only enter/esc
			if m.confirmDelete {
				switch k {
				case "esc":
					m.confirmDelete = false
					m.forceOnConfirm = false
					m.restoreSelectedItemDesc()
					return m, nil
				case "enter":
					if it, ok := m.list.SelectedItem().(item); ok {
						name := it.title
						if strings.HasPrefix(name, "[+]") {
							// should not happen; just cancel
							m.confirmDelete = false
							m.forceOnConfirm = false
							m.restoreSelectedItemDesc()
							return m, nil
						}
						// Don't allow deleting current branch
						if name == m.cur {
							m.confirmDelete = false
							m.forceOnConfirm = false
							m.restoreSelectedItemDesc()
							return m, m.list.NewStatusMessage("Cannot delete the current branch")
						}
						// If already asked to force, do it now
						if m.forceOnConfirm {
							if err := git.DeleteBranch(name, true); err != nil {
								m.confirmDelete = false
								m.forceOnConfirm = false
								m.restoreSelectedItemDesc()
								return m, m.list.NewStatusMessage(fmt.Sprintf("Force delete failed: %v", err))
							}
							m.confirmDelete = false
							m.forceOnConfirm = false
							return m, tea.Batch(m.list.NewStatusMessage(fmt.Sprintf("Deleted %s", name)), loadBranches)
						}
						// Try normal delete; if it fails, suggest force if not fully merged
						if err := git.DeleteBranch(name, false); err != nil {
							es := strings.ToLower(err.Error())
							if strings.Contains(es, "not fully merged") || strings.Contains(es, "fully merged") {
								m.forceOnConfirm = true
								m.replaceSelectedItemDesc(lipgloss.NewStyle().Foreground(theme.Red).Render("Not fully merged. Enter to FORCE delete, Esc to cancel"))
								return m, nil
							}
							// Other errors: cancel confirm and show message
							m.confirmDelete = false
							m.forceOnConfirm = false
							m.restoreSelectedItemDesc()
							return m, m.list.NewStatusMessage(fmt.Sprintf("Delete failed: %v", err))
						}
						m.confirmDelete = false
						m.forceOnConfirm = false
						// reload branches to reflect deletion
						return m, tea.Batch(m.list.NewStatusMessage(fmt.Sprintf("Deleted %s", name)), loadBranches)
					}
				}
				// swallow other keys while confirming
				return m, nil
			}
			switch k {
			case "q":
				return m, tea.Quit
			case "r":
				return m, loadBranches
			case "n":
				if m.branchDel != nil {
					m.branchDel.editing = true
					m.input.SetValue("")
					m.input.Focus()
					m.list.Select(0)
					m.updateAddItemTitle("")
				}
				return m, nil
			case "d":
				if it, ok := m.list.SelectedItem().(item); ok {
					// can't delete the synthetic add item
					if strings.HasPrefix(it.title, "[+]") {
						return m, nil
					}
					// Replace desc with confirmation prompt
					m.confirmDelete = true
					m.forceOnConfirm = false
					m.replaceSelectedItemDesc(lipgloss.NewStyle().Foreground(theme.Surface2).Render("Enter: Yes   Esc: No"))
					return m, nil
				}
			case "enter":
				if it, ok := m.list.SelectedItem().(item); ok {
					if strings.HasPrefix(it.title, "[+]") {
						if m.branchDel != nil {
							m.branchDel.editing = true
							m.input.SetValue("")
							m.input.Focus()
							m.list.Select(0)
							m.updateAddItemTitle("")
						}
						return m, nil
					}
					// checkout branch
					if err := git.Checkout(it.title); err != nil {
						return m, m.list.NewStatusMessage(fmt.Sprintf("Error: %v", err))
					}
					return m, tea.Quit
				}
			}
			// Return the list's previously produced command
			return m, listCmd
		}
	}
	// Forward any other messages (e.g., list.FilterMatchesMsg) to the list
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string { return m.frame.Render(m.list.View()) }

// branchDelegate mirrors worktree approach; we reuse default rendering.
type branchDelegate struct {
	base    list.DefaultDelegate
	input   *textinput.Model
	editing bool
}

func (d *branchDelegate) Height() int                               { return d.base.Height() }
func (d *branchDelegate) Spacing() int                              { return d.base.Spacing() }
func (d *branchDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return d.base.Update(msg, m) }
func (d *branchDelegate) Render(w io.Writer, m list.Model, index int, li list.Item) {
	d.base.Render(w, m, index, li)
}

// updateAddItemTitle sets the title for the first synthetic item to val or the default label
func (m *model) updateAddItemTitle(val string) {
	items := m.list.Items()
	if len(items) == 0 {
		return
	}
	it0, ok := items[0].(item)
	if !ok {
		return
	}
	title := strings.TrimSpace(val)
	if title == "" && !(m.branchDel != nil && m.branchDel.editing) {
		title = "[+] Create new branch"
	}
	it0.title = title
	items[0] = it0
	m.list.SetItems(items)
}

func (m *model) resetAddItemTitle() { m.updateAddItemTitle("") }

// replaceSelectedItemDesc sets the description of the currently selected item
func (m *model) replaceSelectedItemDesc(desc string) {
	idx := m.list.Index()
	if idx < 0 {
		return
	}
	items := m.list.Items()
	if idx >= len(items) {
		return
	}
	if it, ok := items[idx].(item); ok {
		it.desc = desc
		items[idx] = it
		m.list.SetItems(items)
	}
}

// restoreSelectedItemDesc restores the description based on branch state
func (m *model) restoreSelectedItemDesc() {
	idx := m.list.Index()
	if idx < 0 {
		return
	}
	items := m.list.Items()
	if idx >= len(items) {
		return
	}
	if it, ok := items[idx].(item); ok {
		// recompute default desc quickly
		if it.title == m.cur {
			it.desc = lipgloss.NewStyle().Foreground(theme.Green).Render("Active")
		} else if strings.HasPrefix(it.title, "[+]") {
			it.desc = "Type a new branch name"
		} else {
			it.desc = "local branch"
		}
		items[idx] = it
		m.list.SetItems(items)
	}
}
