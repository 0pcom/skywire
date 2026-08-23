// Package tui cmd/skywire/tui/tui_test.go: the browser driven through its own
// Update/View, which is the whole of a bubbletea program's behavior and needs
// no terminal to exercise.
package tui

import (
	"regexp"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

var ansi = regexp.MustCompile("\x1b\\[[0-9;]*[a-zA-Z]")

func strip(s string) string { return ansi.ReplaceAllString(s, "") }

// tree builds a stand-in for the real command tree:
//
//	root
//	├── cli
//	│   ├── config
//	│   └── visor
//	└── visor
func tree() (root, cliVisor *cobra.Command) {
	root = &cobra.Command{Use: "root", Short: "root command", Run: func(*cobra.Command, []string) {}}
	cli := &cobra.Command{Use: "cli", Short: "command line interface", Run: func(*cobra.Command, []string) {}}
	config := &cobra.Command{Use: "config", Short: "manage the config", Run: func(*cobra.Command, []string) {}}
	cliVisor = &cobra.Command{
		Use: "visor", Short: "query the visor",
		Long: "UNIQUEMARKERSTRING for the help pane",
		Run:  func(*cobra.Command, []string) {},
	}
	visor := &cobra.Command{Use: "visor", Short: "the visor", Run: func(*cobra.Command, []string) {}}

	cli.AddCommand(config, cliVisor)
	root.AddCommand(cli, visor)
	return root, cliVisor
}

// sized returns a model that has been told how big the terminal is, which is
// the first thing bubbletea does and what everything else depends on.
func sized(t *testing.T, root, focus *cobra.Command) *model {
	t.Helper()
	m := newModel(root, focus)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	return updated.(*model)
}

func press(t *testing.T, m *model, keys ...string) *model {
	t.Helper()
	for _, k := range keys {
		var msg tea.KeyMsg
		switch k {
		case "up", "down", "left", "right", "enter":
			msg = tea.KeyMsg{Type: map[string]tea.KeyType{
				"up": tea.KeyUp, "down": tea.KeyDown, "left": tea.KeyLeft,
				"right": tea.KeyRight, "enter": tea.KeyEnter,
			}[k]}
		default:
			msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
		}
		updated, _ := m.Update(msg)
		m = updated.(*model)
	}
	return m
}

// Opening on a command means opening the path down to it. `skywire cli visor
// --help --tui` should land on that command, not at the top of a collapsed
// tree with the user left to find it.
func TestOpensOnTheFocusedCommand(t *testing.T) {
	root, cliVisor := tree()
	m := sized(t, root, cliVisor)

	require.Equal(t, cliVisor, m.selected(), "did not open on the focused command")
	require.True(t, m.open["root cli"], "the focused command's parent was left closed")
}

func TestCursorMoves(t *testing.T) {
	root, _ := tree()
	m := sized(t, root, root)
	require.Equal(t, root, m.selected())

	m = press(t, m, "down")
	require.Equal(t, "cli", m.selected().Name())

	m = press(t, m, "up", "up")
	require.Equal(t, root, m.selected(), "cursor ran off the top")

	m = press(t, m, "G")
	require.Equal(t, len(m.nodes)-1, m.cursor, "G did not go to the last row")
}

// Right opens a group, left closes it; left on a closed one steps out to the
// parent, which is what makes left feel like "back".
func TestExpandAndCollapse(t *testing.T) {
	root, _ := tree()
	m := sized(t, root, root)

	m = press(t, m, "down") // onto "cli", which starts closed
	require.False(t, m.open["root cli"])
	before := len(m.nodes)

	m = press(t, m, "right")
	require.True(t, m.open["root cli"])
	require.Greater(t, len(m.nodes), before, "expanding showed no children")

	m = press(t, m, "left")
	require.False(t, m.open["root cli"], "left did not collapse the group")
	require.Equal(t, before, len(m.nodes))

	m = press(t, m, "left")
	require.Equal(t, root, m.selected(), "left on a closed group did not step out")
}

// The right-hand pane is cobra's own help for the selected command. If this
// ever stops being true the browser starts describing a CLI that is not there.
func TestHelpPaneHoldsTheRealHelp(t *testing.T) {
	root, cliVisor := tree()
	m := sized(t, root, cliVisor)

	out := strip(m.View())
	require.Contains(t, out, "UNIQUEMARKERSTRING",
		"the help pane is not showing the command's own help")
}

// The help follows the selection. Capturing it once and leaving it there is an
// easy mistake, since it is captured lazily.
func TestHelpFollowsTheSelection(t *testing.T) {
	root, cliVisor := tree()
	m := sized(t, root, cliVisor)
	require.Contains(t, strip(m.View()), "UNIQUEMARKERSTRING")

	m = press(t, m, "up") // onto "config"
	require.NotContains(t, strip(m.View()), "UNIQUEMARKERSTRING",
		"the help pane kept the previous command's help")
}

// A frame is exactly as tall as the terminal. One row over and the whole
// screen scrolls every tick; one under and there is a dead line at the bottom.
func TestFrameIsExactlyTheTerminalHeight(t *testing.T) {
	root, _ := tree()
	for _, h := range []int{10, 24, 60} {
		m := newModel(root, root)
		updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: h})
		m = updated.(*model)

		rows := strings.Split(strings.TrimRight(m.View(), "\n"), "\n")
		require.Equal(t, h, len(rows), "at height %d the frame was %d rows", h, len(rows))

		for i, r := range rows {
			require.LessOrEqual(t, len([]rune(strip(r))), 100,
				"row %d ran past the width: %q", i, strip(r))
		}
	}
}

// The rain is behind the panes and has to actually move, or it is just an
// expensive still.
func TestTheRainMoves(t *testing.T) {
	root, _ := tree()
	m := sized(t, root, root)

	first := m.View()

	// Two ticks a frame apart, with nothing else changing: only the rain can
	// differ. Synthetic times rather than real ones, so the test does not
	// depend on how long it took to get here.
	now := time.Now()
	u, _ := m.Update(tickMsg(now))
	m = u.(*model)
	u, _ = m.Update(tickMsg(now.Add(200 * time.Millisecond)))
	m = u.(*model)

	require.NotEqual(t, first, m.View(), "the rain did not advance between frames")
}

// Quitting has to work from the keys people actually press.
func TestQuitKeys(t *testing.T) {
	root, _ := tree()
	for _, k := range []string{"q", "esc", "ctrl+c"} {
		m := sized(t, root, root)
		var msg tea.KeyMsg
		switch k {
		case "esc":
			msg = tea.KeyMsg{Type: tea.KeyEsc}
		case "ctrl+c":
			msg = tea.KeyMsg{Type: tea.KeyCtrlC}
		default:
			msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
		}
		_, cmd := m.Update(msg)
		require.NotNil(t, cmd, "%s did not quit", k)
	}
}
