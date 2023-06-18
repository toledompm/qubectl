package prompt

import (
	"fmt"

	"github.com/manifoldco/promptui"
)

type Item struct {
	ID         string
	Index      int
	IsSelected bool
}

func Select(label string, allItems []*Item) (*Item, error) {
	prompt := promptui.Select{
		Label: label,
		Items: allItems,
	}

	index, _, err := prompt.Run()
	if err != nil {
		fmt.Printf("Port selection prompt failed %v\n", err)
		return nil, err
	}

	return allItems[index], nil
}

func MultiSelect(label string, allItems []*Item, selectedPos int) ([]*Item, error) {
	const doneID = "Done"
	if len(allItems) > 0 && allItems[0].ID != doneID {
		var items = []*Item{
			{
				ID: doneID,
			},
		}
		allItems = append(items, allItems...)
	}

	templates := &promptui.SelectTemplates{
		Active:   "→ {{if .IsSelected}}✔ {{end}}{{ .ID }}",
		Inactive: "{{if .IsSelected}}✔ {{end}}{{ .ID }}",
	}

	prompt := promptui.Select{
		Label:        label,
		Items:        allItems,
		Templates:    templates,
		CursorPos:    selectedPos,
		HideSelected: true,
	}

	selectionIdx, _, err := prompt.Run()
	if err != nil {
		return nil, fmt.Errorf("prompt failed: %w", err)
	}

	chosenItem := allItems[selectionIdx]

	if chosenItem.ID != doneID {
		chosenItem.IsSelected = !chosenItem.IsSelected
		return MultiSelect(label, allItems, selectionIdx)
	}

	var selectedItems []*Item
	for _, i := range allItems {
		if i.IsSelected {
			selectedItems = append(selectedItems, i)
		}
	}
	return selectedItems, nil
}
