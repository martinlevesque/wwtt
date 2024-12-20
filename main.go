package main

import (
	"fmt"
	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
	"github.com/martinlevesque/wwtt/internal/storage"
	"github.com/rivo/tview"
	"log"
	"strings"
)

type App struct {
	EntriesStorage   *storage.StorageFile
	ListNotes        *tview.List
	ListTags         *tview.List
	Legend           *tview.TextView
	SearchField      *tview.InputField
	TextContent      *tview.TextArea
	CurrentTag       string
	CurrentSearchTag string
	CurrentSearch    string
	CurrentNoteName  string
}

// generic ui tools

func (app *App) searchListInputCapture(itemsName string, list *tview.List, event *tcell.EventKey, search *string, tag string, getItems func(t string) []string) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyRune:
		keyStr := string(event.Rune())
		*search += keyStr
		list.SetTitle(fmt.Sprintf("%s (searching %s)", itemsName, *search))
		app.findItemsList(list, getItems(tag), *search)
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(*search) > 0 {
			*search = (*search)[:len(*search)-1]
		}

		list.SetTitle(fmt.Sprintf("%s (searching %s)", itemsName, *search))
		app.findItemsList(list, getItems(tag), *search)
	}

	return event
}

func (app *App) searchableList(search *string, tag string, itemsName string, list *tview.List, updateSelected *string) {
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {

		switch event.Key() {
		case tcell.KeyEnter:
			if list.GetItemCount() == 0 {
				*updateSelected = *search
				app.search()
			}
		}

		return app.searchListInputCapture(itemsName, list, event, search, tag, app.retrieveTagNames)
	})

	list.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		*updateSelected = mainText
	})
}

func (app *App) searchableListByTextfield(search *string, itemsName string, searchTextField *tview.InputField, list *tview.List) {
	searchTextField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return app.searchListInputCapture(itemsName, list, event, search, app.CurrentTag, app.retrieveItems)
	})
}

func (app *App) findItemsList(tviewList *tview.List, itemNames []string, searchingFor string) {
	currentIndex := tviewList.GetCurrentItem()
	var currentLabel string

	// If an item is selected, get its label
	if currentIndex >= 0 && currentIndex < tviewList.GetItemCount() {
		currentLabel, _ = tviewList.GetItemText(currentIndex)
	}

	tviewList.Clear()

	for index, itemName := range itemNames {
		strippedItemName := strings.TrimSpace(strings.ToLower(itemName))
		strippedSearchingFor := strings.TrimSpace(strings.ToLower(searchingFor))

		if strings.Contains(strippedItemName, strippedSearchingFor) || strippedSearchingFor == "" {
			tviewList.AddItem(itemName, "", 0, nil)

			if currentLabel == itemName {
				tviewList.SetCurrentItem(index)
			}
		}
	}
}

// tags

func (app *App) retrieveTagNames(tag string) []string {
	items := []string{}
	seen := make(map[string]struct{})
	seen["all"] = struct{}{}
	items = append(items, "all")

	for _, item := range app.EntriesStorage.Notes {
		if _, exists := seen[item.Tag.Name]; !exists {
			seen[item.Tag.Name] = struct{}{} // Mark as added
			items = append(items, item.Tag.Name)
		}
	}

	return items
}

func (app *App) makeListTags() *tview.List {
	fixedItems := app.retrieveTagNames("")

	listTags := tview.NewList()

	for _, tagName := range fixedItems {
		listTags.AddItem(tagName, "", 0, nil)
	}

	app.searchableList(&app.CurrentSearchTag, app.CurrentTag, "Tags", listTags, &app.CurrentTag)

	return listTags
}

// Notes

func (app *App) retrieveItems(tag string) []string {
	items := []string{}

	for _, item := range app.EntriesStorage.Notes {
		if tag != "" && (tag == item.Tag.Name || tag == "all") {
			items = append(items, item.Name)
		}
	}

	return items
}

func (app *App) findItem(itemName string, itemTag string) (storage.Note, bool) {
	for _, item := range app.EntriesStorage.Notes {
		if item.Name == itemName && (itemTag == "all" || item.Tag.Name == itemTag) {
			return item, true
		}
	}

	return storage.Note{}, false
}

func (app *App) search() {
	// with the current search and current tag
	app.findItemsList(app.ListNotes, app.retrieveItems(app.CurrentTag), app.CurrentSearch)
}

func (app *App) copyNoteClipboard() {
	noteContent := app.TextContent.GetText()

	if err := clipboard.WriteAll(noteContent); err != nil {
		app.TextContent.SetTitle("error copying to clipboard")
	}
}

func (app *App) loadNote(noteName string) {
	note, noteFound := app.findItem(noteName, app.CurrentTag)

	if noteFound {
		app.TextContent.SetText(note.Content, true)
		app.TextContent.SetTitle(fmt.Sprintf("%s - tag %s", app.CurrentSearch, app.CurrentTag))
		app.CurrentNoteName = noteName
	} else {
		app.TextContent.SetTitle("Content")
	}
}
func (app *App) deleteCurrentNode() {
	if app.ListNotes.GetItemCount() > 0 {
		app.EntriesStorage.DeleteNote(app.CurrentNoteName, app.CurrentTag)
		app.EntriesStorage.Save()
		app.search()

		if app.ListNotes.GetItemCount() == 0 {
			app.TextContent.SetText("", true)
		}
	}
}

func (app *App) onCreateNote(noteName string) {
	app.EntriesStorage.CreateNote(noteName, app.CurrentTag)
	app.EntriesStorage.Save()
	app.CurrentNoteName = app.CurrentSearch
	app.search()
	app.loadNote(app.CurrentNoteName)
}

func (app *App) updateLegend(focusedComponent tview.Primitive) {
	switch focusedComponent {
	case app.ListTags:
		app.Legend.SetText(`[yellow]TAB:[white] Next Focus [yellow]Enter:[white] Select Tag`)
	case app.SearchField:
		app.Legend.SetText(`[yellow]TAB:[white] Next Focus [yellow]Enter:[white] Search/Create Note`)
	case app.ListNotes:
		app.Legend.SetText(`[yellow]TAB:[white] Next Focus [yellow]Ctrl+X:[white] Copy Note [yellow]Delete:[white] Delete Note`)
	case app.TextContent:
		app.Legend.SetText(`[yellow]Ctrl+S:[white] Save [yellow]Ctrl+X:[white] Copy [yellow]Esc:[white] Go Back`)
	default:
		app.Legend.SetText("")
	}
}

func main() {
	entriesStorage, err := storage.Init("wwtt.json")

	app := &App{
		ListNotes:        nil,
		ListTags:         nil,
		TextContent:      nil,
		EntriesStorage:   entriesStorage,
		CurrentTag:       "all",
		CurrentSearchTag: "",
		CurrentSearch:    "",
	}

	uiApp := tview.NewApplication()

	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// List for tags with border and title

	app.ListTags = app.makeListTags()

	app.ListTags.SetBorder(true)
	app.ListTags.SetTitle("Tags")

	app.ListTags.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		app.CurrentTag = mainText
		app.search()
	})

	// Main list to show items
	app.ListNotes = tview.NewList()
	app.findItemsList(app.ListNotes, app.retrieveItems(app.CurrentTag), "")

	app.ListNotes.SetBorder(true)
	app.ListNotes.SetTitle("Notes")

	app.ListNotes.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		app.loadNote(mainText)
	})

	app.ListNotes.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlX:
			app.copyNoteClipboard()
		}

		return event
	})

	// Input field for search
	app.SearchField = tview.NewInputField().
		SetLabel("Search/Create: ")

	app.searchableListByTextfield(&app.CurrentSearch, "Notes", app.SearchField, app.ListNotes)

	app.TextContent = tview.NewTextArea().
		SetWrap(false).
		SetPlaceholder("Enter text here...")
	app.TextContent.SetTitle("Content")
	app.TextContent.SetBorder(true)

	app.TextContent.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlS:
			errRecording := entriesStorage.RecordNote(app.CurrentNoteName, app.TextContent.GetText())

			if errRecording != nil {
				app.TextContent.SetTitle("error recording")
			}

			errSaving := entriesStorage.Save()

			if errSaving != nil {
				app.TextContent.SetTitle("error saving ")
			}
		case tcell.KeyCtrlX:
			app.copyNoteClipboard()
		}
		return event
	})

	// Layout for the left side with fixed height for listTags
	leftSide := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(app.ListTags, 6, 1, true).     // Fixed height for listTags
		AddItem(app.SearchField, 2, 0, false). // Fixed height for input
		AddItem(app.ListNotes, 0, 2, false)    // Expandable list

	app.Legend = tview.NewTextView()
	app.Legend.
		SetDynamicColors(true).
		SetText("Legend will appear here").
		SetTextAlign(tview.AlignCenter).
		SetBorder(true).
		SetTitle("Legend")

	// Main layout with left and right sides
	mainFlex := tview.NewFlex().
		AddItem(leftSide, 0, 1, true).
		AddItem(app.TextContent, 0, 2, false)

	// Input capture to handle focus navigation
	mainFlex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTAB: // Tab key to cycle through focus
			if uiApp.GetFocus() == app.ListTags {
				uiApp.SetFocus(app.SearchField)
			} else if uiApp.GetFocus() == app.SearchField {
				uiApp.SetFocus(app.ListNotes)
			} else if uiApp.GetFocus() == app.ListNotes {
				uiApp.SetFocus(app.ListTags)
			}
		case tcell.KeyEnter: // Tab key to cycle through focus
			if uiApp.GetFocus() == app.SearchField {
				uiApp.SetFocus(app.ListNotes)

				_, noteFound := app.findItem(app.CurrentSearch, app.CurrentTag)

				if !noteFound {
					app.onCreateNote(app.CurrentSearch)
				}
			} else if uiApp.GetFocus() == app.ListNotes {
				uiApp.SetFocus(app.TextContent)
				app.loadNote(app.CurrentNoteName)
			}
		case tcell.KeyDelete:
			if uiApp.GetFocus() == app.ListNotes {
				app.deleteCurrentNode()
			}
		case tcell.KeyEscape:
			if uiApp.GetFocus() == app.TextContent {
				uiApp.SetFocus(app.ListTags)
			}
		}

		app.updateLegend(uiApp.GetFocus())

		return event
	})

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(mainFlex, 0, 1, true).   // Main application layout
		AddItem(app.Legend, 3, 0, false) // Fixed height for the legend

	// Run the application with initial focus on listTags
	if err := uiApp.SetRoot(layout, true).SetFocus(app.ListTags).Run(); err != nil {
		panic(err)
	}
}
