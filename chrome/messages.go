package chrome

import "github.com/mhbidhan/lazycurl/store"

// OpenCollectionMsg asks the root screen to open the explorer for a collection.
// The collections popup emits it rather than switching screens itself, so only
// the root owns screen changes.
type OpenCollectionMsg struct {
	Collection store.Collection
}
