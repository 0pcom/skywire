// Package tui cmd/skywire/tui/tui.go
//
// An interactive browser over the command tree: the subcommands on the left,
// the selected command's help on the right, and the code rain falling behind
// both. `skywire --tui`, or `--tui` alongside any `--help`.
//
// It shows what is already there rather than describing it again. The tree is
// walked from the live *cobra.Command, and the help in the right-hand pane is
// cobra's own help for that command, captured as text — the same bytes
// `--help` would have printed, colors and all. There is no second copy of
// either to fall out of step with the first.
//
// Read-only. It is a way around the help of a CLI with several hundred
// commands in it, not a launcher: nothing here runs anything.
package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/0magnet/termanim/matrix/backdrop"

	"github.com/skycoin/skywire/pkg/flags"
)

// frameRate is how often the rain is advanced. The simulation is tuned at 30
// steps a second and is driven from elapsed time, so a slower tick costs
// smoothness and not speed — the rain falls at the same rate either way.
const frameRate = 50 * time.Millisecond

// treeWidth is the left pane. Wide enough for the deepest command name at the
// indent it sits at, narrow enough to leave the help the bulk of the screen.
const treeWidth = 30

type tickMsg time.Time

// node is one row of the flattened tree: a command, and how deep it sits.
type node struct {
	cmd   *cobra.Command
	depth int
}

type model struct {
	root   *cobra.Command
	nodes  []node
	cursor int
	// open records which groups are expanded, by command path. Paths rather
	// than pointers so the set survives the tree being flattened again.
	open map[string]bool

	help    viewport.Model
	painter *backdrop.Painter
	w, h    int

	// helpFor is the command the help pane currently holds, so the help is
	// re-captured when the selection moves and not on every frame.
	helpFor *cobra.Command

	// dt is elapsed seconds banked by the ticks and spent by the next frame.
	dt       float64
	lastTick time.Time

	showKeys bool
}

// Run opens the browser on cmd and blocks until the user quits.
func Run(root, focus *cobra.Command) error {
	m := newModel(root, focus)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func newModel(root, focus *cobra.Command) *model {
	m := &model{
		root: root,
		open: map[string]bool{},
		painter: backdrop.New(backdrop.Options{
			// The screen is composed here, so the backdrop is asked for no
			// padding of its own and told where the layout's empty space is.
			Pad:    -1,
			GapMin: 4,
			// Behind a full screen of panes rather than behind a paragraph:
			// darker than the help screen's, since there is much more of it.
			Dim:   40,
			Force: true,
		}),
	}

	// Open the path down to the focused command, so it is on screen and its
	// ancestors are expanded around it rather than the user having to find it.
	for c := focus; c != nil && c != root; c = c.Parent() {
		if c.Parent() != nil {
			m.open[c.Parent().CommandPath()] = true
		}
	}
	m.flatten()
	for i, n := range m.nodes {
		if n.cmd == focus {
			m.cursor = i
		}
	}
	return m
}

func (m *model) Init() tea.Cmd { return tick() }

func tick() tea.Cmd {
	return tea.Tick(frameRate, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// flatten walks the tree into the rows currently on screen, following open.
func (m *model) flatten() {
	m.nodes = m.nodes[:0]
	var walk func(c *cobra.Command, depth int)
	walk = func(c *cobra.Command, depth int) {
		m.nodes = append(m.nodes, node{cmd: c, depth: depth})
		if !m.open[c.CommandPath()] {
			return
		}
		for _, sub := range c.Commands() {
			if !sub.IsAvailableCommand() {
				continue
			}
			walk(sub, depth+1)
		}
	}
	m.open[m.root.CommandPath()] = true
	walk(m.root, 0)
	if m.cursor >= len(m.nodes) {
		m.cursor = len(m.nodes) - 1
	}
}

func (m *model) selected() *cobra.Command {
	if m.cursor < 0 || m.cursor >= len(m.nodes) {
		return m.root
	}
	return m.nodes[m.cursor].cmd
}

// hasChildren reports whether a command is a group worth expanding.
func hasChildren(c *cobra.Command) bool {
	for _, sub := range c.Commands() {
		if sub.IsAvailableCommand() {
			return true
		}
	}
	return false
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		if t := time.Time(msg); !m.lastTick.IsZero() {
			m.dt += t.Sub(m.lastTick).Seconds()
		}
		m.lastTick = time.Time(msg)
		return m, tick()

	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		// The framework is the authority on the size, not the terminal.
		m.painter.SetWidth(msg.Width)
		m.layoutPanes()
		m.helpFor = nil // re-wrap the help at the new width
		return m, nil

	case tea.KeyMsg:
		return m.key(msg)
	}
	return m, nil
}

func (m *model) key(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "ctrl+c":
		return m, tea.Quit

	case "?":
		m.showKeys = !m.showKeys
		m.layoutPanes()

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.nodes)-1 {
			m.cursor++
		}
	case "home", "g":
		m.cursor = 0
	case "end", "G":
		m.cursor = len(m.nodes) - 1

	case "right", "l", "enter":
		// Open a group; on a leaf, do nothing rather than something arbitrary.
		if c := m.selected(); hasChildren(c) && !m.open[c.CommandPath()] {
			m.open[c.CommandPath()] = true
			m.flatten()
		}
	case "left", "h":
		// Close an open group, or step out to the parent of a closed one,
		// which is what makes left feel like "back".
		c := m.selected()
		if m.open[c.CommandPath()] && hasChildren(c) {
			delete(m.open, c.CommandPath())
			m.flatten()
			break
		}
		if p := c.Parent(); p != nil {
			for i, n := range m.nodes {
				if n.cmd == p {
					m.cursor = i
					break
				}
			}
		}

	// The help pane scrolls independently: some of these commands have more
	// help than fits any screen.
	case "pgdown", "ctrl+f", " ":
		m.help.PageDown()
	case "pgup", "ctrl+b":
		m.help.PageUp()
	case "J", "ctrl+n":
		m.help.ScrollDown(1)
	case "K", "ctrl+p":
		m.help.ScrollUp(1)
	}
	return m, nil
}

// layoutPanes sizes the help viewport to whatever is left over.
func (m *model) layoutPanes() {
	w := m.w - treeWidth - 3
	if w < 20 {
		w = 20
	}
	h := m.h - chromeRows(m.showKeys)
	if h < 3 {
		h = 3
	}
	m.help.Width, m.help.Height = w, h
}

// chromeRows is everything on screen that is not a pane: the title, the rule
// under it, and the key line.
func chromeRows(showKeys bool) int {
	if showKeys {
		return 5
	}
	return 3
}

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
	pathStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	dimStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	selStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("0")).
			Background(lipgloss.Color("10"))
	groupStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	leafStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
)

// View is the composed screen with the rain painted in behind it.
//
// The two halves are kept apart: screen decides what the program looks like
// and paint decides what is behind it, and only the first is worth a test.
func (m *model) View() string {
	if m.w == 0 {
		// No size yet: bubbletea sends the first WindowSizeMsg right after
		// Init, so this is one frame at most.
		return ""
	}
	s := m.screen()

	// dt is what has actually elapsed, accumulated by the ticks, so the rain
	// falls at its own rate however often the screen happens to be redrawn. A
	// redraw that is not a tick — a keypress — passes zero and does not move
	// it, or the rain would run at the speed the user types.
	out := m.painter.Frame(s, m.dt)
	m.dt = 0
	return out
}

// screen composes the frame as plain text: title, the two panes, the key line.
// Exactly m.h rows, none wider than m.w.
func (m *model) screen() string {
	m.refreshHelp()

	rows := make([]string, 0, m.h)
	rows = append(rows,
		titleStyle.Render("skywire")+dimStyle.Render(" — interactive help   ")+
			pathStyle.Render(m.selected().CommandPath()),
		dimStyle.Render(strings.Repeat("─", m.w)),
	)

	tree := strings.Split(m.treePane(), "\n")
	help := strings.Split(m.help.View(), "\n")
	body := m.h - chromeRows(m.showKeys)
	for i := 0; i < body; i++ {
		l, r := "", ""
		if i < len(tree) {
			l = tree[i]
		}
		if i < len(help) {
			r = help[i]
		}
		rows = append(rows,
			lipgloss.NewStyle().Width(treeWidth).MaxWidth(treeWidth).Render(l)+
				dimStyle.Render(" │ ")+r)
	}

	if m.showKeys {
		rows = append(rows,
			dimStyle.Render(strings.Repeat("─", m.w)),
			dimStyle.Render("  ↑/k ↓/j move    →/l enter open    ←/h back    "+
				"PgUp/PgDn scroll help    g/G top/bottom"))
	}
	rows = append(rows, dimStyle.Render("  ?  keys    q  quit"))

	return strings.Join(rows, "\n")
}

// treePane renders the visible rows of the tree, scrolled to keep the cursor
// in view.
func (m *model) treePane() string {
	body := m.h - chromeRows(m.showKeys)
	if body < 1 {
		body = 1
	}

	// Scroll so the cursor is on screen, keeping it off the very edge where
	// there is room to.
	top := 0
	if m.cursor >= body {
		top = m.cursor - body + 1
	}
	if top > 0 && top+body < len(m.nodes) {
		top++
	}

	var b strings.Builder
	for i := top; i < len(m.nodes) && i-top < body; i++ {
		n := m.nodes[i]
		mark := "  "
		if hasChildren(n.cmd) {
			if m.open[n.cmd.CommandPath()] {
				mark = "▾ "
			} else {
				mark = "▸ "
			}
		}
		line := strings.Repeat("  ", n.depth) + mark + n.cmd.Name()

		// Clip rather than let a deep name push the pane wider than the
		// column the whole layout is built on.
		if len([]rune(line)) > treeWidth-1 {
			line = string([]rune(line)[:treeWidth-2]) + "…"
		}

		switch {
		case i == m.cursor:
			b.WriteString(selStyle.Render(line))
		case hasChildren(n.cmd):
			b.WriteString(groupStyle.Render(line))
		default:
			b.WriteString(leafStyle.Render(line))
		}
		b.WriteByte('\n')
	}
	return strings.TrimRight(b.String(), "\n")
}

// refreshHelp captures cobra's help for the selected command, once per
// selection rather than once per frame.
func (m *model) refreshHelp() {
	c := m.selected()
	if m.helpFor == c {
		return
	}
	m.helpFor = c

	text := helpText(c)
	// lipgloss wraps with the escape sequences accounted for, which matters:
	// the help arrives from coloredcobra already colored.
	wrapped := lipgloss.NewStyle().Width(m.help.Width).Render(text)
	m.help.SetContent(wrapped)
	m.help.GotoTop()
}

// helpText is cobra's own help for c, as text.
//
// Through the command's help function rather than its template, so it is the
// same bytes `--help` prints — and inside WithPlainHelp, so the help function
// does not paint a backdrop into a pane that is about to get one anyway.
func helpText(c *cobra.Command) string {
	var buf strings.Builder
	out := c.OutOrStdout()
	c.SetOut(&buf)
	flags.WithPlainHelp(func() {
		if err := c.Help(); err != nil {
			fmt.Fprintf(&buf, "help for %s: %v\n", c.CommandPath(), err)
		}
	})
	c.SetOut(out)
	return buf.String()
}
