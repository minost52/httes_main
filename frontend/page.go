package frontend

// import (
// 	"httes/frontend/loadtest"

// 	"fyne.io/fyne/v2"
// 	"fyne.io/fyne/v2/app"
// 	"fyne.io/fyne/v2/container"
// 	"fyne.io/fyne/v2/theme"
// 	"fyne.io/fyne/v2/widget"
// )

// type Page struct {
// 	Name    string
// 	Content fyne.CanvasObject
// }

// // initPages инициализирует страницы приложения.
// func initPages(app fyne.App, window fyne.Window) *Page {
// 	// Создаем элементы интерфейса для LoadTest
// 	resultOutput := widget.NewTextGrid()

// 	// Создаем экземпляр MainPage
// 	mainPage := loadtest.NewMainPage(app, resultOutput)

// 	// Создаем контент для страницы LoadTest
// 	loadTestContent := mainPage.CreateLoadTestContent(window)

// 	return &Page{Name: "LoadTest", Content: loadTestContent}
// }

// // createMainMenu создает главное меню приложения.
// func createMainMenu(app fyne.App) *fyne.MainMenu {
// 	// Подменю "Тема" с опциями Светлая и Темная
// 	lightThemeItem := fyne.NewMenuItem("Светлая", func() {
// 		app.Settings().SetTheme(theme.LightTheme())
// 	})
// 	lightThemeItem.Icon = theme.VisibilityOffIcon()

// 	darkThemeItem := fyne.NewMenuItem("Темная", func() {
// 		app.Settings().SetTheme(theme.DarkTheme())
// 	})
// 	darkThemeItem.Icon = theme.VisibilityIcon()

// 	// Создаем подменю "Тема"
// 	themeMenu := fyne.NewMenu("", lightThemeItem, darkThemeItem)

// 	// Создаем пункт меню "Тема" с подменю
// 	themeMenuItem := fyne.NewMenuItem("Тема", nil)
// 	themeMenuItem.ChildMenu = themeMenu

// 	// Создаем главное меню
// 	mainMenu := fyne.NewMainMenu(
// 		fyne.NewMenu("Настройка", themeMenuItem),
// 	)

// 	return mainMenu
// }

// // Run запускает приложение.
// func Run() {
// 	a := app.New()
// 	a.Settings().SetTheme(theme.DarkTheme())
// 	w := a.NewWindow("Httes")
// 	w.Resize(fyne.NewSize(700, 700))

// 	// Инициализируем страницу
// 	page := initPages(a, w)
// 	contentContainer := container.NewMax(page.Content)

// 	// Создаем главное меню
// 	mainMenu := createMainMenu(a)
// 	w.SetMainMenu(mainMenu)

// 	// Устанавливаем содержимое окна
// 	w.SetContent(contentContainer)
// 	w.ShowAndRun()
// }
