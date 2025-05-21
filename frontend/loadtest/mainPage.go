package loadtest

import (
	"httes/core/report"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// MainPage управляет основным UI и состоянием приложения.
type MainPage struct {
	app           fyne.App
	resultOutput  *widget.TextGrid
	progressBar   *widget.ProgressBar
	progressText  *widget.Label
	reportService report.ReportService
	ui            *LoadTestUI
}

// NewMainPage создаёт новый экземпляр MainPage.
func NewMainPage(app fyne.App) *MainPage {
	mp := &MainPage{app: app}
	mp.init()
	return mp
}

// init инициализирует сервисы и UI-компоненты.
func (mp *MainPage) init() {
	mp.reportService = report.NewGuiReportService(mp.resultOutput, mp.progressBar, mp.progressText, 0)
	mp.ui = NewLoadTestUI(mp.app, mp.resultOutput, mp.progressBar, mp.progressText)
}

// CreateLoadTestContent создаёт содержимое UI для нагрузочного тестирования.
func (mp *MainPage) CreateLoadTestContent(window fyne.Window) fyne.CanvasObject {
	// Создание секций UI
	urlSection := createURLSection()
	proxySection := createProxySection()
	paramsSection := createParamsSection()
	certSection := createCertFields(window)
	buttonsSection := mp.ui.CreateButtons()

	// Прогресс-бар и метка прогресса
	progressBarContainer := container.NewHBox(
		container.NewGridWrap(fyne.NewSize(400, 30), mp.progressBar),
		mp.progressText,
	)

	// Оборачиваем resultOutput в ScrollContainer с ограничением размера
	scrollContainer := container.NewScroll(mp.resultOutput)
	scrollContainer.SetMinSize(fyne.NewSize(300, 300))

	title := widget.NewLabel("Httes")
	title.TextStyle = fyne.TextStyle{Bold: true}
	description := widget.NewLabel("Модуль нагрузочного тестирования для определения таймингов HTTP, HTTPS запросов.")

	// Контейнер для URL, параметров и сертификатов (верхняя часть)
	upContent := container.NewVBox(
		title,
		description,
		urlSection,
		paramsSection,
		container.NewHBox(proxySection, certSection),
	)
	upContentContainer := container.NewMax(upContent)

	// Контейнер для кнопок и прогресс-бара (нижняя часть)
	downContent := container.NewVBox(
		buttonsSection,
		progressBarContainer,
	)
	downContentContainer := container.NewMax(downContent)

	// Вертикальный сплит для разделения кнопок/прогресса (снизу) и остальных секций (сверху)
	inputContent := container.NewVSplit(
		upContentContainer,
		downContentContainer,
	)
	inputContent.SetOffset(0.8) // 80% высоты для upContent, 20% для downContent

	// Контейнер для статистики
	progressContent := container.NewVBox(
		widget.NewLabel("Статистика теста"),
		scrollContainer,
	)
	progressContentContainer := container.NewMax(progressContent)

	// Плейсхолдер для графиков
	chartPlaceholder := container.NewVBox(
		widget.NewLabel("Графики (поддержка в разработке)"),
		widget.NewLabel("Здесь будут визуализации нагрузки и производительности"),
	)
	chartPlaceholderContainer := container.NewMax(chartPlaceholder)

	// Вертикальный сплит для статистики и графиков
	mainContent := container.NewVSplit(
		progressContentContainer,
		chartPlaceholderContainer,
	)
	mainContent.SetOffset(0.3) // 70% для статистики, 30% для графиков

	// Горизонтальный сплит для объединения inputContent и mainContent
	finalContent := container.NewHSplit(
		inputContent,
		mainContent,
	)
	finalContent.SetOffset(0.4) // 40% для inputContent, 60% для статистики/графиков

	return container.NewMax(finalContent)
}
