package main

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/martinlevesque/wwtt/internal/storage"
	"github.com/rivo/tview"
	"log"
	"strings"
)

type App struct {
	EntriesStorage *storage.StorageFile
}

func (app *App) searchListInputCapture(itemsName string, list *tview.List, event *tcell.EventKey, search *string, getItems func() []string) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyRune:
		keyStr := string(event.Rune())
		*search += keyStr
		list.SetTitle(fmt.Sprintf("%s (searching %s)", itemsName, *search))
		app.findItemsList(list, getItems(), *search)
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(*search) > 0 {
			*search = (*search)[:len(*search)-1]
		}

		list.SetTitle(fmt.Sprintf("%s (searching %s)", itemsName, *search))
		app.findItemsList(list, getItems(), *search)
	}

	return event
}

func (app *App) searchableList(itemsName string, list *tview.List) {
	search := ""

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return app.searchListInputCapture(itemsName, list, event, &search, app.retrieveTagNames)
	})
}

func (app *App) searchableListByTextfield(itemsName string, searchTextField *tview.InputField, list *tview.List) {
	search := ""

	searchTextField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return app.searchListInputCapture(itemsName, list, event, &search, app.retrieveItems)
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

func (app *App) retrieveTagNames() []string {
	return []string{"all", "tag1", "tag2", "abc", "ab", "abcde"}
}

func (app *App) retrieveItems() []string {
	items := []string{}

	for _, item := range app.EntriesStorage.Notes {
		items = append(items, item.Name)
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

func (app *App) makeListTags() *tview.List {
	fixedItems := app.retrieveTagNames()

	listTags := tview.NewList()

	for _, tagName := range fixedItems {
		listTags.AddItem(tagName, "", 0, nil)
	}

	app.searchableList("Tags", listTags)

	return listTags
}

func main() {
	entriesStorage, err := storage.Init("wwtt.json")
	app := &App{EntriesStorage: entriesStorage}
	uiApp := tview.NewApplication()

	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// List for tags with border and title

	listTags := app.makeListTags()

	listTags.SetBorder(true)
	listTags.SetTitle("Tags")

	// Main list to show items
	listNotes := tview.NewList()
	app.findItemsList(listNotes, app.retrieveItems(), "")

	listNotes.SetBorder(true)
	listNotes.SetTitle("Notes")

	// Input field for search
	searchField := tview.NewInputField().
		SetLabel("Search/Create: ")

	searchField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return event
	})

	app.searchableListByTextfield("Notes", searchField, listNotes)

	textContent := tview.NewTextArea().
		SetWrap(false).
		SetPlaceholder("Enter text here...").
		SetText("yooo", true)
	textContent.SetText("truasdf", true)
	textContent.SetTitle("Content")
	textContent.SetBorder(true)

	listNotes.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		note, noteFound := app.findItem(mainText, "all")

		if noteFound {
			textContent.SetText(note.Content, true)
		}
	})

	// Layout for the left side with fixed height for listTags
	leftSide := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(listTags, 6, 1, true).     // Fixed height for listTags
		AddItem(searchField, 2, 0, false). // Fixed height for input
		AddItem(listNotes, 0, 2, false)    // Expandable list

	// Main layout with left and right sides
	mainFlex := tview.NewFlex().
		AddItem(leftSide, 0, 1, true).
		AddItem(textContent, 0, 2, false)

	// Input capture to handle focus navigation
	mainFlex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTAB: // Tab key to cycle through focus
			if uiApp.GetFocus() == listTags {
				uiApp.SetFocus(searchField)
			} else if uiApp.GetFocus() == searchField {
				uiApp.SetFocus(listNotes)
			} else if uiApp.GetFocus() == listNotes {
				uiApp.SetFocus(textContent)
			} else if uiApp.GetFocus() == textContent {
				uiApp.SetFocus(listTags)
			}
		}

		return event
	})

	// Run the application with initial focus on listTags
	if err := uiApp.SetRoot(mainFlex, true).SetFocus(searchField).Run(); err != nil {
		panic(err)
	}
}
