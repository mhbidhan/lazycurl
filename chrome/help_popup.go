package chrome

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mhbidhan/lazycurl/keybind"
	"github.com/mhbidhan/lazycurl/ui"
)

// HelpPopup shows a scrollable list of key bindings.
type HelpPopup struct {
	vp       viewport.Model
	keybinds *keybind.HelpPopupKeys
}

// NewHelpPopup builds the keybinds popup with content sized to w by h.
func NewHelpPopup(bindings []keybind.KeyInfo, w, h int) *HelpPopup {
	vp := viewport.New(w, h)
	vp.MouseWheelEnabled = true
	vp.SetContent(RenderKeybinds(bindings))
	return &HelpPopup{
		vp:       vp,
		keybinds: keybind.NewHelpPopupKeys(),
	}
}

func (h *HelpPopup) Init() tea.Cmd {
	return nil
}

func (h *HelpPopup) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(keyMsg, h.keybinds.Escape):
			return h, func() tea.Msg {
				return DismissPopup{}
			}
		}
	}

	var cmd tea.Cmd
	h.vp, cmd = h.vp.Update(msg)
	return h, cmd
}

// SetSize clamps to one cell: the popup manager already subtracted this
// popup's frame, and a container too small for that cannot be made to fit.
func (h *HelpPopup) SetSize(w, height int) {
	h.vp.Width = max(w, 1)
	h.vp.Height = max(height, 1)
}

func (h *HelpPopup) View() string {
	return ui.PopupBox(h.vp.View())
}
