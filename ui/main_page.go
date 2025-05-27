package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// MainPage управляет основным UI и состоянием приложения.
type MainPage struct {
	app          fyne.App
	resultOutput *widget.TextGrid
	progressBar  *widget.ProgressBar
	progressText *widget.Label
	isDarkMode   bool
	window       fyne.Window // Добавляем ссылку на окно
	icon         fyne.Resource
}

// NewMainPage создаёт новый экземпляр MainPage.
func NewMainPage(app fyne.App, resultOutput *widget.TextGrid, window fyne.Window, icon fyne.Resource) *MainPage {
	mp := &MainPage{
		app:          app,
		resultOutput: resultOutput,
		progressBar:  widget.NewProgressBar(),
		progressText: widget.NewLabel("Request Avg Duration 0.000s"),
		window:       window, // Сохраняем ссылку на окно
		icon:         icon,
	}
	mp.progressBar.Max = 100.0
	return mp
}

// init инициализирует UI-компоненты.
func (mp *MainPage) init() {
	// Заглушка для инициализации LoadTestUI
}

func (mp *MainPage) CreateUI(window fyne.Window) fyne.CanvasObject {
	// Сначала создаем вкладки
	tabs := container.NewAppTabs()

	// Главная вкладка (добавляем первой)
	tabs.Append(container.NewTabItem("Главная", mp.createHomeScreen(
		func() { tabs.SelectIndex(1) }, // Обработчик для "Запуск теста"
		func() { tabs.SelectIndex(2) }, // Обработчик для "Сценарии"
		func() { tabs.SelectIndex(4) }, // Обработчик для "История"
	)))

	// Создаем остальные экраны, передавая tabs туда, где нужно
	testRunScreen := container.NewVScroll(mp.createTestRunScreen(window, tabs))
	scenariosScreen := container.NewVScroll(mp.createScenariosScreen(window, tabs))
	historyScreen := container.NewVScroll(mp.createHistoryScreen(window, tabs))

	// Добавляем остальные вкладки
	tabs.Append(container.NewTabItem("Запуск теста", testRunScreen))
	tabs.Append(container.NewTabItem("Сценарии", scenariosScreen))
	tabs.Append(container.NewTabItem("История", historyScreen))

	return container.NewMax(tabs)
}
