package main

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/martinlevesque/wwtt/internal/storage"
	"github.com/rivo/tview"
	"log"
	"strings"
)

func searchListInputCapture(itemsName string, list *tview.List, event *tcell.EventKey, search *string, getItems func() []string) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyRune:
		keyStr := string(event.Rune())
		*search += keyStr
		list.SetTitle(fmt.Sprintf("%s (searching %s)", itemsName, *search))
		findItemsList(list, getItems(), *search)
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(*search) > 0 {
			*search = (*search)[:len(*search)-1]
		}

		list.SetTitle(fmt.Sprintf("%s (searching %s)", itemsName, *search))
		findItemsList(list, getItems(), *search)
	}

	return event
}

func searchableList(itemsName string, list *tview.List) {
	search := ""

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return searchListInputCapture(itemsName, list, event, &search, retrieveTagNames)
	})
}

func searchableListByTextfield(itemsName string, searchTextField *tview.InputField, list *tview.List) {
	search := ""

	searchTextField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return searchListInputCapture(itemsName, list, event, &search, retrieveItems)
	})
}

func findItemsList(tviewList *tview.List, itemNames []string, searchingFor string) {
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

func retrieveTagNames() []string {
	return []string{"all", "tag1", "tag2", "abc", "ab", "abcde"}
}

func retrieveItems() []string {
	return []string{"item 1", "item 2", "item 3"}
}

func makeListTags() *tview.List {
	fixedItems := retrieveTagNames()

	listTags := tview.NewList()

	for _, tagName := range fixedItems {
		listTags.AddItem(tagName, "", 0, nil)
	}

	searchableList("Tags", listTags)

	return listTags
}

func main() {
	app := tview.NewApplication()

	_, err := storage.Init("wwtt.json")

	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// List for tags with border and title

	listTags := makeListTags()

	listTags.SetBorder(true)
	listTags.SetTitle("Tags")

	// Main list to show items
	listNotes := tview.NewList()
	findItemsList(listNotes, retrieveItems(), "")

	listNotes.SetBorder(true)
	listNotes.SetTitle("Notes")

	// Input field for search
	searchField := tview.NewInputField().
		SetLabel("Search/Create: ")

	searchField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return event
	})

	searchableListByTextfield("Notes", searchField, listNotes)

	textContent := tview.NewTextArea().
		SetWrap(false).
		SetPlaceholder("Enter text here...").
		SetText("yooo", true)
	textContent.SetText("truasdf", true)
	textContent.SetTitle("Content")
	textContent.SetBorder(true)

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
			if app.GetFocus() == listTags {
				app.SetFocus(searchField)
			} else if app.GetFocus() == searchField {
				app.SetFocus(listNotes)
			} else if app.GetFocus() == listNotes {
				app.SetFocus(textContent)
			} else if app.GetFocus() == textContent {
				app.SetFocus(listTags)
			}
		}

		return event
	})

	// Run the application with initial focus on listTags
	if err := app.SetRoot(mainFlex, true).SetFocus(searchField).Run(); err != nil {
		panic(err)
	}
}
