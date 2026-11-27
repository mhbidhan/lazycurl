package chrome

import tea "github.com/charmbracelet/bubbletea"

// OperationError carries a failed store operation to the popup that started it.
// The popup renders it as its status line instead of discarding the error.
type OperationError struct {
	Err error
}

// report clears a stale status on success and turns a failure into a message.
// Returning a command keeps error handling in Update, where View cannot mutate
// state.
func (c *CollectionsPopup) report(err error) tea.Cmd {
	if err == nil {
		c.err = nil
		return nil
	}

	return func() tea.Msg { return OperationError{Err: err} }
}
