package chrome

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mhbidhan/lazycurl/keybind"
	"github.com/mhbidhan/lazycurl/ui"
)

// NamePromptEvent is what NamePromptPopup reports back to whoever opened it.
// Cancel wins over Value.
type NamePromptEvent struct {
	Value  string
	Cancel bool
}

// NamePromptPopup asks for one name. Creating and renaming differ only in the
// prompt and the value it starts with, so they share a popup and a set of
// bindings.
type NamePromptPopup struct {
	width, height int

	input    textinput.Model
	prompt   string
	initial  string
	keybinds *keybind.NamePromptKeys
}

// NewNamePromptPopup builds a single-field prompt starting at initial.
func NewNamePromptPopup(prompt, initial string) *NamePromptPopup {
	input := textinput.New()
	input.CharLimit = 50
	input.Width = 30
	input.Prompt = prompt

	return &NamePromptPopup{
		input:    input,
		prompt:   prompt,
		initial:  initial,
		keybinds: keybind.NewNamePromptKeys(),
	}
}

func (np *NamePromptPopup) Init() tea.Cmd {
	return textinput.Blink
}

// OnOpen resets the field to the value it was built with, so reopening never
// shows the previous answer.
func (np *NamePromptPopup) OnOpen() {
	np.input.SetValue(np.initial)
	np.input.CursorEnd()
	np.input.Focus()
}

func (np *NamePromptPopup) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return np, nil
	}

	switch {
	case key.Matches(keyMsg, np.keybinds.Cancel):
		return np, func() tea.Msg {
			return NamePromptEvent{Cancel: true}
		}
	case key.Matches(keyMsg, np.keybinds.Submit):
		return np, func() tea.Msg {
			return NamePromptEvent{Value: np.input.Value()}
		}
	}

	var cmd tea.Cmd
	np.input, cmd = np.input.Update(msg)
	return np, cmd
}

// SetSize fits the field to whatever is left of the width once the prompt has
// had its share.
func (np *NamePromptPopup) SetSize(w, h int) {
	np.width = max(w, 1)
	np.height = max(h, 1)
	np.fitInput(w)
}

// fitInput keeps the prompt, the field and the cursor inside w columns. When the
// prompt alone leaves no room for the field, the prompt is dropped rather than
// letting the line run past the frame: a bare field still shows what is being
// edited, whereas an overlong line gets cropped and takes the border with it.
func (np *NamePromptPopup) fitInput(w int) {
	if lipgloss.Width(np.prompt)+2 > w {
		np.input.Prompt = ""
		np.input.Width = max(w-1, 1)
		return
	}

	np.input.Prompt = np.prompt
	np.input.Width = fieldWidth(w, np.prompt)
}

func (np *NamePromptPopup) View() string {
	inputView := lipgloss.NewStyle().Render(np.input.View())

	content := lipgloss.JoinVertical(lipgloss.Center, inputView)

	return ui.PopupBox(content)
}
