package core

type Navigation struct {
	Root NavigationItem
}

type NavigationItem struct {
	Url         string
	Title       string
	Children    []NavigationItem
	IsActive    bool // helper field for templating
	IsDirectory bool
}

func createNavigationItem(context *Context, directory Directory) (NavigationItem, error) {
	item := NavigationItem{
		Url:         directory.Url,
		Title:       directory.Title,
		Children:    make([]NavigationItem, 0),
		IsDirectory: true,
	}

	// Create a NavigationItem for each file
	for _, file := range directory.Files {
		child := NavigationItem{
			Url:         file.Url,
			Title:       file.Title,
			IsDirectory: false,
		}
		item.Children = append(item.Children, child)
	}

	// Create a NavigationItem for each subdirectory
	for _, dir := range directory.Subdirectories {
		child, err := createNavigationItem(context, dir)
		if err != nil {
			return NavigationItem{}, err
		}
		item.Children = append(item.Children, child)
	}

	return item, nil
}

func InitializeNavigation(context *Context) (Navigation, error) {
	var navigation Navigation

	// Go through the Filesystem structure and build the navigation tree
	item, err := createNavigationItem(context, context.Root)
	if err != nil {
		return Navigation{}, err
	}
	navigation.Root = item
	return navigation, nil
}
