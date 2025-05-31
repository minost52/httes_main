package ui

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var (
	methodSelect     *widget.Select
	protocolSelect   *widget.Select
	urlEntry         *widget.Entry
	proxyEntry       *widget.Entry
	reqCount         *widget.Entry
	duration         *widget.Entry
	loadType         *widget.RadioGroup
	usernameEntry    *widget.Entry
	passwordEntry    *widget.Entry
	certPathEntry    *widget.Entry
	certKeyPathEntry *widget.Entry
)

func (mp *ControlPage) createTestRunScreen(window fyne.Window, tabs *container.AppTabs) fyne.CanvasObject {
	// Инициализация UI-компонентов
	if mp.progressBar == nil {
		mp.progressBar = widget.NewProgressBar()
		mp.progressBar.Min = 0.0
		mp.progressBar.Max = 100.0
	}
	if mp.progressText == nil {
		mp.progressText = widget.NewLabel("Request Avg Duration 0.000s")
	}
	if mp.resultOutput == nil {
		mp.resultOutput = widget.NewTextGrid()
	}

	ui := NewLoadTestUI(mp.app, window)
	ui.resultOutput = mp.resultOutput
	ui.progressBar = mp.progressBar
	ui.progressText = mp.progressText

	// Верхняя панель управления
	var controls *fyne.Container
	if mp.role != "intern" {
		controls = container.NewVBox(
			widget.NewSeparator(),
			createURLSection(),
			createParamsSection(),
			container.NewHBox(createProxySection(), createCertFields(window)),
			mp.createScenariosSection(tabs),
			widget.NewSeparator(),
			ui.CreateButtons(),
			widget.NewSeparator(),
			container.NewHBox(
				container.NewGridWrap(fyne.NewSize(400, 30), ui.progressBar),
				ui.progressText,
			),
		)
	} else {
		controls = container.NewVBox(
			widget.NewSeparator(),
			createURLSection(),
			createParamsSection(),
			container.NewHBox(createProxySection(), createCertFields(window)),
			widget.NewSeparator(),
			ui.CreateButtons(),
			widget.NewSeparator(),
			container.NewHBox(
				container.NewGridWrap(fyne.NewSize(400, 30), ui.progressBar),
				ui.progressText,
			),
		)
	}

	// Результаты
	results := container.NewVScroll(ui.resultOutput)
	results.SetMinSize(fyne.NewSize(300, 150))

	// Создаем заголовок с текстом и кнопкой истории
	resultHeader := container.NewBorder(
		nil,
		nil,
		widget.NewLabel("Результаты теста"),
		widget.NewButtonWithIcon("История тестов", theme.HistoryIcon(), func() {
			tabs.SelectTabIndex(3)
		}),
		nil,
	)

	resultsBox := container.NewVBox(
		resultHeader,
		results,
	)

	left := container.NewVBox(container.NewVBox(
		controls,
		widget.NewSeparator(),
		resultsBox,
	))

	// Правая часть — графики
	chartContainer := CreateLoadTestCharts()
	chartBox := container.NewVBox(
		widget.NewLabelWithStyle("Метрики теста", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		chartContainer,
	)
	chartBoxContainer := container.NewVScroll(chartBox)
	chartBoxContainer.SetMinSize(fyne.NewSize(300, 200))

	// Сохраняем ссылку на контейнер графиков в UI
	ui.chartsContainer = chartBox

	layout := container.NewHBox(
		left,
		widget.NewSeparator(),
		chartBoxContainer,
	)

	return layout
}

func (mp *ControlPage) createScenariosSection(tabs *container.AppTabs) fyne.CanvasObject {
	// Создаем метку для отображения выбранного сценария
	scenarioLabel := widget.NewLabel("Сценарий не выбран")
	scenarioLabel.Wrapping = fyne.TextTruncate

	// Кнопка для выбора сценария
	selectBtn := widget.NewButtonWithIcon("Выбрать сценарий", theme.NavigateNextIcon(), func() {
		tabs.SelectIndex(2)
	})

	// "API" с эффектом гиперссылки (подчеркнуто и синим цветом)
	apiText := widget.NewHyperlink("API", nil)

	return container.NewHBox(
		selectBtn,
		apiText,
		scenarioLabel,
	)
}

func createURLSection() *fyne.Container {
	methodSelect = widget.NewSelect([]string{"GET", "POST"}, nil)
	methodSelect.SetSelected("GET")

	protocolSelect = widget.NewSelect([]string{"HTTP", "HTTPS"}, nil)
	protocolSelect.SetSelected("HTTPS")

	urlEntry = widget.NewEntry()
	urlEntry.SetText("example.com")

	return container.NewHBox(
		container.NewGridWrap(fyne.NewSize(80, methodSelect.MinSize().Height), methodSelect),
		container.NewGridWrap(fyne.NewSize(100, protocolSelect.MinSize().Height), protocolSelect),
		container.NewGridWrap(fyne.NewSize(250, urlEntry.MinSize().Height), urlEntry),
	)
}

func createProxySection() *fyne.Container {
	proxyEntry = widget.NewEntry()
	proxyEntry.SetPlaceHolder("http://127.0.0.1:8080")
	return container.NewVBox(
		widget.NewLabel("Proxy (необязательно)"),
		container.NewGridWrap(fyne.NewSize(250, proxyEntry.MinSize().Height), proxyEntry),
	)
}

func createParamsSection() *fyne.Container {
	reqCount = widget.NewEntry()
	reqCount.SetText("10")

	duration = widget.NewEntry()
	duration.SetText("1")

	loadType = widget.NewRadioGroup([]string{"Linear", "Incremental", "Waved"}, nil)
	loadType.SetSelected("Linear")

	usernameEntry = widget.NewEntry()
	usernameEntry.SetPlaceHolder("Username")

	passwordEntry = widget.NewEntry()
	passwordEntry.SetPlaceHolder("Password")
	passwordEntry.Password = true

	authAccordion := widget.NewAccordion(widget.NewAccordionItem("Basic Auth", container.NewVBox(
		usernameEntry,
		passwordEntry,
	)))

	reqCount.OnChanged = func(s string) {
		if _, err := parseInt(s); err != nil {
			reqCount.SetText("10")
		}
	}
	duration.OnChanged = func(s string) {
		if _, err := parseInt(s); err != nil {
			duration.SetText("1")
		}
	}

	return container.NewHBox(
		container.NewGridWrap(
			fyne.NewSize(120, widget.NewLabel("Request Count*").MinSize().Height+reqCount.MinSize().Height),
			container.NewVBox(
				widget.NewLabel("Request Count*"),
				reqCount,
			),
		),
		container.NewGridWrap(
			fyne.NewSize(120, widget.NewLabel("Duration (s)*").MinSize().Height+duration.MinSize().Height),
			container.NewVBox(
				widget.NewLabel("Duration (s)*"),
				duration,
			),
		),
		container.NewGridWrap(
			fyne.NewSize(120, widget.NewLabel("Load Type*").MinSize().Height+loadType.MinSize().Height),
			container.NewVBox(
				widget.NewLabel("Load Type*"),
				loadType,
			),
		),
		container.NewGridWrap(
			fyne.NewSize(150, authAccordion.MinSize().Height),
			container.NewVBox(
				authAccordion,
			),
		),
	)
}

func createCertFields(window fyne.Window) *fyne.Container {
	certPathEntry = widget.NewEntry()
	certPathEntry.SetPlaceHolder("Путь к сертификату")
	certPathEntry.Disable()

	certKeyPathEntry = widget.NewEntry()
	certKeyPathEntry.SetPlaceHolder("Путь к ключу")
	certKeyPathEntry.Disable()

	certLabel := widget.NewLabel("Не выбрано")
	keyLabel := widget.NewLabel("Не выбрано")

	certFilter := storage.NewExtensionFileFilter([]string{".crt", ".pem"})
	selectCertButton := widget.NewButton("Выбрать сертификат", func() {
		fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				certPathEntry.SetText(reader.URI().Path())
				certLabel.SetText(reader.URI().Name())
				reader.Close()
			}
		}, window)
		fileDialog.SetFilter(certFilter)
		fileDialog.Show()
	})

	keyFilter := storage.NewExtensionFileFilter([]string{".key"})
	selectKeyButton := widget.NewButton("Выбрать ключ", func() {
		fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				certKeyPathEntry.SetText(reader.URI().Path())
				keyLabel.SetText(reader.URI().Name())
				reader.Close()
			}
		}, window)
		fileDialog.SetFilter(keyFilter)
		fileDialog.Show()
	})

	certAccordion := widget.NewAccordion(widget.NewAccordionItem("Сертификаты", container.NewVBox(
		container.NewHBox(selectCertButton, certLabel),
		container.NewHBox(selectKeyButton, keyLabel),
	)))

	return container.NewVBox(certAccordion)
}

func parseInt(s string) (int, error) {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid number %s: %v", s, err)
	}
	return i, nil
}
