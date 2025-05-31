package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// MainPage управляет основным UI и состоянием приложения.
type ControlPage struct {
	app             fyne.App
	resultOutput    *widget.TextGrid
	progressBar     *widget.ProgressBar
	progressText    *widget.Label
	isDarkMode      bool
	window          fyne.Window
	username        string // Поле для имени пользователя
	role            string // Поле для роли пользователя (developer или intern)
	icon            fyne.Resource
	chartsContainer fyne.CanvasObject // Контейнер для графиков
}

// NewMainPage создаёт новый экземпляр MainPage.
func NewMainPage(app fyne.App, resultOutput *widget.TextGrid, window fyne.Window, username string, role string, icon fyne.Resource) *ControlPage {
	mp := &ControlPage{
		app:          app,
		resultOutput: resultOutput,
		progressBar:  widget.NewProgressBar(),
		progressText: widget.NewLabel("Request Avg Duration 0.000s"),
		window:       window,
		username:     username, // Сохраняем имя пользователя
		role:         role,     // Сохраняем роль пользователя
		icon:         icon,
	}
	mp.progressBar.Max = 100.0

	// Инициализируем графики один раз
	mp.chartsContainer = CreateLoadTestCharts()

	return mp
}

// ResetCharts обнуляет данные графиков
func (mp *ControlPage) ResetCharts() {
	GlobalMetrics.mu.Lock()
	defer GlobalMetrics.mu.Unlock()
	GlobalMetrics.Times = []float64{0, 1} // Начальные значения
	GlobalMetrics.RPS = []float64{0, 0}
	GlobalMetrics.RespTimes = []float64{0, 0}
	GlobalMetrics.Errors = []float64{0, 0}
	// Отправляем сигнал для обновления графиков
	select {
	case GlobalMetrics.updateChan <- struct{}{}:
	default:
	}
}

// init инициализирует UI-компоненты.
func (mp *ControlPage) init() {
	// Заглушка для инициализации LoadTestUI
}

func (mp *ControlPage) CreateUI(window fyne.Window) fyne.CanvasObject {
	// Сначала создаем вкладки
	tabs := container.NewAppTabs()

	// Главная вкладка (добавляем первой)
	tabs.Append(container.NewTabItem("Главная", mp.createHomeScreen(
		func() { tabs.SelectIndex(1) }, // Обработчик для "Запуск теста"
		func() {
			// Для роли intern вкладка "Сценарии" недоступна
			if mp.role != "intern" {
				tabs.SelectIndex(2)
			}
		},
		func() { tabs.SelectIndex(3) }, // Обработчик для "История"
	)))

	// Создаем остальные экраны, передавая tabs туда, где нужно
	testRunScreen := container.NewVScroll(mp.createTestRunScreen(window, tabs))
	historyScreen := container.NewVScroll(mp.createHistoryScreen(window, tabs))

	// Добавляем вкладки
	tabs.Append(container.NewTabItem("Запуск теста", testRunScreen))
	// Вкладка "Сценарии" добавляется только для роли developer
	if mp.role != "intern" {
		scenariosScreen := container.NewVScroll(mp.createScenariosScreen(window, tabs))
		tabs.Append(container.NewTabItem("Сценарии", scenariosScreen))
	}
	tabs.Append(container.NewTabItem("История", historyScreen))

	return container.NewMax(tabs)
}
