Issue: String keys for popups are error-prone
Impacted area: chrome/popupmgr.go, screen/home.go
Solution: Use typed PopupID constants (already done) and consider a registry that validates. Also could use iota if many popups.
Code example (If applicable): 
```go
type PopupID string
const (
  PopupHelp PopupID = "help"
  PopupCollections PopupID = "collections"
)
```

Issue: Manager stores updated models but doesn't persist complex state if replaced
Impacted area: chrome/popupmgr.go (Update)
Solution: After `updated, cmd := m.Update(msg)`, write back to map. Already implemented: `pm.popups[id] = updated`. Also need to ensure models return same identity when not replaced (pointer receivers) which we do.
Code example (If applicable): 
```go
updated, cmd := m.Update(msg)
pm.popups[id] = updated
return cmd
```

Issue: Size propagation doesn't reach all registered popups if screen forgets to call SetSize
Impacted area: chrome/popupmgr.go, screen/home.go
Solution: Either call SetSize on all registered on WindowSizeMsg (HomeScreen does via pm.SetSize) or have manager track desired size and apply on Open/Register. Also add SetSizeAll helper.
Code example (If applicable):
```go
func (pm *PopupManager) SetSize(w,h int){ for _,m:=range pm.popups{ if s,ok:=m.(interface{SetSize(int,int)});ok{s.SetSize(w,h)} } }
```

Issue: Single-active model prevents stacked/modals; also opening a popup closes others (mutually exclusive) which is fine for modals but not general overlays
Impacted area: PopupManager design
Solution: Keep current mutually exclusive for this app. If stacked needed later, track active set with z-order (slice) and render topmost last/first. Not required now.

Issue: Key precedence - when popup active, global keys in HomeScreen must be blocked; popup handles its own keys
Impacted area: screen/home.go, popups
Solution: When pm.IsOpen(), don't process global bindings except maybe Esc to close. Currently implemented: if open, only Quit/Keybind/Collections close; others fall through to popup.Update. Popup.Update handles viewport keys (arrows/j/k/mouse). OK.

Issue: WindowSizeMsg should reach popups even if not active? Probably not necessary since SetSize called explicitly
Impacted area: screen/home.go
Solution: Call pm.SetSize on every WindowSizeMsg (already done via HomeScreen.SetSize called from Root). Also popups store their own viewport sizes; fine.

Issue: Popup chrome renders border/padding but outer centering is done by screen; sizes must match
Impacted area: screen/home.go, chrome/*popup*.go
Solution: Screen centers active popup with full width/height; popup View() returns bordered content. Sizes computed as popupWidth/popupHeight (inner-ish) passed to popup via SetSize. Viewport sizes align. OK.

Issue: Command chaining - when popup active, only popup cmd returned; when not, global cmd returned. Manager.Update returns only popup cmd. HomeScreen returns it correctly.
Impacted area: screen/home.go
Solution: Keep as is. If multiple sources of cmds later, use tea.Batch.

Issue: String-based popup IDs could drift; use typed constants (done). Also consider exported vs unexported.
Impacted area: chrome/popupmgr.go
Solution: Keep PopupID typed consts. Add doc comments.

Issue: Models stored as tea.Model in map; SetSize uses type assertion - fine for internal API.
Impacted area: PopupManager
Solution: Define a small interface (Sizer) to avoid repeated assertions: `type Sizer interface{SetSize(int,int)}`. Cleaner.

Issue: collections popup creates content once in New; if dynamic later, rebuild on update. OK for now.
Impacted area: chrome/collectionspopup.go
Solution: If list becomes dynamic, move content generation to View or update viewport.SetContent on changes.

Issue: HomeScreen stores homeScreen reference; after refactor root also stores homeScreen - consistent. PopupManager holds all popups by value in map (models are pointers in practice) so updates persist.
Impacted area: state ownership
Solution: Models are pointer receivers (*HelpPopup etc), so storing in map as tea.Model is fine and updated model replaces the value. OK.

Issue: Closing on same key as toggle - in current code, pressing ? or c when open closes active (treats as toggle-off). That’s reasonable for modals.
Impacted area: screen/home.go
Solution: Acceptable UX. Alternatively track last opened key.

Issue: No CloseAll or Open replacement semantics - Open replaces active (overwrites). CloseActive clears. OK.

Issue: Potential import cycles - ui moved keybinds to chrome; chrome imports keybind/ui where needed. No cycle (chrome -> keybind, chrome -> ui; ui doesn’t import chrome). OK.

Issue: If popup becomes very large, viewport clamps; mousewheel enabled. Scroll keys handled by viewport defaults. OK.
