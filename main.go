package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func searchableList(itemsName string, list *tview.List) {
	search := ""

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			keyStr := string(event.Rune())
			search += keyStr
			list.SetTitle(fmt.Sprintf("%s (searching %s)", itemsName, search))
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			if len(search) > 0 {
				search = search[:len(search)-1]
			}

			list.SetTitle(fmt.Sprintf("%s (searching %s)", itemsName, search))
		}

		return event
	})

}

func makeListTags() *tview.List {
	listTags := tview.NewList().
		AddItem("all", "", 0, nil).
		AddItem("tag 2", "", 0, nil).
		AddItem("tag 3", "", 0, nil).
		AddItem("tag 4", "", 0, nil).
		AddItem("tag 5", "", 0, nil)

	searchableList("Tags", listTags)

	return listTags
}

func main() {
	app := tview.NewApplication()

	// List for tags with border and title

	listTags := makeListTags()
	// listTags.SetTitle(fmt.Sprintf("selected index %d", listTags.GetCurrentItem()))

	// Input field for search
	inputField := tview.NewInputField().
		SetLabel("Search: ")

	listTags.SetBorder(true)
	listTags.SetTitle("Tags")

	// Main list to show items
	list := tview.NewList().
		AddItem("Item 1", "Description 1", '1', nil).
		AddItem("Item 2", "Description 2", '2', nil).
		AddItem("Item 3", "Description 3", '3', nil)

	// Right-side text/code view
	textView := tview.NewTextView().
		SetText("Code or text will be displayed here...").
		SetDynamicColors(true).
		SetWordWrap(true).
		SetBorder(true).
		SetTitle("Content")

	// Layout for the left side with fixed height for listTags
	leftSide := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(listTags, 6, 1, true).    // Fixed height for listTags
		AddItem(inputField, 2, 0, false). // Fixed height for input
		AddItem(list, 0, 2, false)        // Expandable list

	// Main layout with left and right sides
	mainFlex := tview.NewFlex().
		AddItem(leftSide, 0, 1, true).
		AddItem(textView, 0, 2, false)

	// Input capture to handle focus navigation
	mainFlex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTAB: // Tab key to cycle through focus
			if app.GetFocus() == listTags {
				app.SetFocus(inputField)
			} else if app.GetFocus() == inputField {
				app.SetFocus(list)
			} else if app.GetFocus() == list {
				app.SetFocus(listTags)
			}
		}

		return event
	})

	// Run the application with initial focus on listTags
	if err := app.SetRoot(mainFlex, true).SetFocus(listTags).Run(); err != nil {
		panic(err)
	}
}
