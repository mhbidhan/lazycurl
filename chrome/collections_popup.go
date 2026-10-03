package chrome

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mhbidhan/lazycurl/keybind"
	"github.com/mhbidhan/lazycurl/ui"
)

type CollectionsPopup struct {
	keybind *keybind.CollectionKeys
	vp      viewport.Model
}

func NewCollectionsPopup(w, h int) *CollectionsPopup {
	kb := keybind.NewCollectionKeys()
	vp := viewport.New(w, h)
	vp.MouseWheelEnabled = true

	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color("9")).
		Align(lipgloss.Center, lipgloss.Center).
		Render("Collections")

	cl := []string{}
	for i := range 10 {
		cl = append(cl, fmt.Sprintf("Collection %d", i))
	}

	body := lipgloss.
		NewStyle().
		Align(lipgloss.Left).
		Render(strings.Join(cl, "\n"))

	coll := lipgloss.JoinVertical(
		lipgloss.Top,
		header,
		ui.GapVertical(1),
		body,
	)

	vp.SetContent(coll)

	return &CollectionsPopup{
		keybind: kb,
		vp:      vp,
	}
}

func (c *CollectionsPopup) Init() tea.Cmd {
	return nil
}

func (c *CollectionsPopup) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(keyMsg, c.keybind.Quit):
			return c, nil
		}
	}

	var cmd tea.Cmd
	c.vp, cmd = c.vp.Update(msg)
	return c, cmd
}

func (c *CollectionsPopup) SetSize(w, h int) {
	c.vp.Width = w
	c.vp.Height = h
}

func (c *CollectionsPopup) View() string {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("212")).
		Padding(1, 2).
		Render(c.vp.View())
}
