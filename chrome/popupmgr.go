package chrome

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// fieldWidth is how wide a text field can be and still fit in w columns beside
// its prompt. The extra column is the cursor: bubbles renders a cursor cell on
// top of the padded value, so reserving the prompt alone leaves the line one
// column too wide, and the popup frame wraps instead of fitting.
func fieldWidth(w int, prompt string) int {
	return max(w-lipgloss.Width(prompt)-1, 1)
}

// PopupID identifies a popup registered with a PopupManager.
type PopupID string

const (
	PopupHelp        PopupID = "help"
	PopupCollections PopupID = "collections"
)

// Sizer is implemented by popups that need to know the area they render into.
type Sizer interface {
	SetSize(w, h int)
}

// OnOpener is implemented by popups that need to reset themselves each time
// they are opened.
type OnOpener interface {
	OnOpen()
}

// DismissPopup asks the owner of a PopupManager to close the popup that is
// currently open. Popups emit it instead of closing themselves, so that only
// the owner decides what "open" means.
type DismissPopup struct{}

type PopupManager struct {
	popups map[PopupID]tea.Model
	active PopupID

	// The area popups are sized against. Remembered so a popup registered after
	// the last resize still fits: the collections popup builds its confirm and
	// rename prompts on demand, long after the window was last measured.
	w, h int
}

func NewPopupManager() *PopupManager {
	return &PopupManager{
		popups: make(map[PopupID]tea.Model),
		active: "",
	}
}

func (pm *PopupManager) Register(id PopupID, p tea.Model) {
	pm.popups[id] = p

	if s, ok := p.(Sizer); ok {
		s.SetSize(pm.w, pm.h)
	}
}

// Open makes the popup registered under id active, if it is registered. An
// unknown id is ignored.
func (pm *PopupManager) Open(id PopupID) {
	m, ok := pm.popups[id]
	if !ok {
		return
	}
	if o, ok := m.(OnOpener); ok {
		o.OnOpen()
	}
	pm.active = id
}

// Close deactivates id if it is the active popup. Closing an inactive popup is
// a no-op.
func (pm *PopupManager) Close(id PopupID) {
	if pm.active == id {
		pm.active = ""
	}
}

// Active returns the open popup and the id it was registered under, or
// ("", nil) when nothing is open.
func (pm *PopupManager) Active() (PopupID, tea.Model) {
	if m, ok := pm.popups[pm.active]; ok {
		return pm.active, m
	}
	return "", nil
}

func (pm *PopupManager) IsOpen() bool {
	return pm.active != ""
}

func (pm *PopupManager) Update(msg tea.Msg) tea.Cmd {
	if id, m := pm.Active(); m != nil {
		updated, cmd := m.Update(msg)
		pm.popups[id] = updated
		return cmd
	}
	return nil
}

// SetSize propagates the available area to every registered popup, not just
// the open one, so nothing renders at a stale size.
func (pm *PopupManager) SetSize(w, h int) {
	pm.w, pm.h = w, h

	for _, m := range pm.popups {
		if s, ok := m.(Sizer); ok {
			s.SetSize(w, h)
		}
	}
}
