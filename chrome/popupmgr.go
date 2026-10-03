package chrome

import (
	tea "github.com/charmbracelet/bubbletea"
)

type PopupManager struct {
	popups map[string]tea.Model
	active string
}

func NewPopupManager() *PopupManager {
	return &PopupManager{
		popups: make(map[string]tea.Model),
		active: "",
	}
}

func (pm *PopupManager) Register(id string, p tea.Model) {
	pm.popups[id] = p
}

func (pm *PopupManager) Open(id string) {
	if _, ok := pm.popups[id]; ok {
		pm.active = id
	}
}

func (pm *PopupManager) Close(id string) {
	if pm.active == id {
		pm.active = ""
	}
}

func (pm *PopupManager) CloseActive() {
	pm.active = ""
}

func (pm *PopupManager) Active() (string, tea.Model) {
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
		var cmd tea.Cmd
		updated, cmd := m.Update(msg)
		pm.popups[id] = updated
		return cmd
	}
	return nil
}

func (pm *PopupManager) SetSize(w, h int) {
	for _, m := range pm.popups {
		if s, ok := m.(interface{ SetSize(int, int) }); ok {
			s.SetSize(w, h)
		}
	}
}
