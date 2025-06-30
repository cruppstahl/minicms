package core

import (
	"sort"
)

type Navigation struct {
	Root NavigationItem
}

type NavigationItem struct {
	Url         string
	Title       string
	Position    int
	Children    []NavigationItem
	IsActive    bool // helper field for templating
	IsDirectory bool
}

func createNavigationItem(context *Context, directory *Directory) (NavigationItem, error) {
	item := NavigationItem{
		Url:         directory.Url,
		Title:       directory.Title,
		Children:    make([]NavigationItem, 0),
		IsDirectory: true,
	}

	usedPositions := make(map[int]bool)

	// Create a NavigationItem for each file
	for _, file := range directory.Files {
		if file.NavHidden {
			continue
		}

		child := NavigationItem{
			Url:         file.Url,
			Title:       file.Title,
			IsDirectory: false,
			Position:    file.NavPosition,
		}

		// remember file.NavPosition to make sure that it is not used again
		if file.NavPosition >= 0 {
			usedPositions[file.NavPosition] = true
		}

		item.Children = append(item.Children, child)
	}

	// Create a NavigationItem for each subdirectory
	for _, dir := range directory.Subdirectories {
		if dir.NavHidden {
			continue
		}

		child, err := createNavigationItem(context, &dir)
		if err != nil {
			return NavigationItem{}, err
		}
		child.Position = dir.NavPosition

		// remember dir.NavPosition to make sure that it is not used again
		if dir.NavPosition >= 0 {
			usedPositions[dir.NavPosition] = true
		}

		item.Children = append(item.Children, child)
	}

	// Now assign a unique position to those items who do not yet have one
	pos := 0
	for i, child := range item.Children {
		// If this child already has a position, we can skip it
		if child.Position >= 0 {
			continue
		}

		// otherwise increment pos until we find a position that is not used
		for ; ; pos++ {
			used, ok := usedPositions[pos]
			if !ok || !used {
				// found a position that is not used
				break
			}
		}

		child.Position = pos
		item.Children[i] = child
		pos++
	}

	// sort all Children based on their position
	sort.Slice(item.Children, func(i, j int) bool {
		return item.Children[i].Position < item.Children[j].Position
	})

	return item, nil
}

func InitializeNavigation(context *Context) (Navigation, error) {
	var navigation Navigation

	// Go through the Filesystem structure and build the navigation tree
	item, err := createNavigationItem(context, &context.Root)
	if err != nil {
		return Navigation{}, err
	}
	navigation.Root = item
	return navigation, nil
}
