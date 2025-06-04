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
			mp.createURLSection(),
			mp.createParamsSection(),
			container.NewHBox(mp.createProxySection(), mp.createCertFields(window)),
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
			mp.createURLSection(),
			mp.createParamsSection(),
			container.NewHBox(mp.createProxySection(), mp.createCertFields(window)),
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

	// Кнопка крестика для удаления "API"
	closeBtn := widget.NewButtonWithIcon("", theme.CancelIcon(), nil)
	closeBtn.Importance = widget.LowImportance // Делаем кнопку менее заметной

	// Контейнер для динамического управления
	scenarioContainer := container.NewHBox(
		selectBtn,
		apiText,
		closeBtn,
		scenarioLabel,
	)

	// Функция для отключения полей
	disableFields := func() {
		if mp.methodSelect != nil {
			mp.methodSelect.Disable()
		}
		if mp.protocolSelect != nil {
			mp.protocolSelect.Disable()
		}
		if mp.urlEntry != nil {
			mp.urlEntry.Disable()
		}
		if mp.proxyEntry != nil {
			mp.proxyEntry.Disable()
		}
		if mp.reqCount != nil {
			mp.reqCount.Disable()
		}
		if mp.duration != nil {
			mp.duration.Disable()
		}
		if mp.loadType != nil {
			mp.loadType.Disable()
		}
		if mp.usernameEntry != nil {
			mp.usernameEntry.Disable()
		}
		if mp.passwordEntry != nil {
			mp.passwordEntry.Disable()
		}
		if mp.certPathEntry != nil {
			mp.certPathEntry.Disable()
		}
		if mp.certKeyPathEntry != nil {
			mp.certKeyPathEntry.Disable()
		}
		if mp.selectCertButton != nil {
			mp.selectCertButton.Disable()
		}
		if mp.selectKeyButton != nil {
			mp.selectKeyButton.Disable()
		}
	}

	// Функция для включения полей
	enableFields := func() {
		if mp.methodSelect != nil {
			mp.methodSelect.Enable()
		}
		if mp.protocolSelect != nil {
			mp.protocolSelect.Enable()
		}
		if mp.urlEntry != nil {
			mp.urlEntry.Enable()
		}
		if mp.proxyEntry != nil {
			mp.proxyEntry.Enable()
		}
		if mp.reqCount != nil {
			mp.reqCount.Enable()
		}
		if mp.duration != nil {
			mp.duration.Enable()
		}
		if mp.loadType != nil {
			mp.loadType.Enable()
		}
		if mp.usernameEntry != nil {
			mp.usernameEntry.Enable()
		}
		if mp.passwordEntry != nil {
			mp.passwordEntry.Enable()
		}
		if mp.certPathEntry != nil {
			if mp.certPathEntry.Disabled() {
				mp.certPathEntry.Enable()
			}
		}
		if mp.certKeyPathEntry != nil {
			if mp.certKeyPathEntry.Disabled() {
				mp.certKeyPathEntry.Enable()
			}
		}
		if mp.selectCertButton != nil {
			mp.selectCertButton.Enable()
		}
		if mp.selectKeyButton != nil {
			mp.selectKeyButton.Enable()
		}
	}

	// Обработчик для крестика
	closeBtn.OnTapped = func() {
		// Удаляем текст "API" и скрываем кнопку
		apiText.SetText("")
		closeBtn.Hide()
		// Включаем поля
		enableFields()
		// Обновляем контейнер
		scenarioContainer.Refresh()
	}

	// Добавляем обработчик для отображения крестика и отключения полей при выборе "API"
	apiText.OnTapped = func() {
		if apiText.Text == "" {
			apiText.SetText("API")
			closeBtn.Show()
			// Отключаем поля
			disableFields()
			scenarioContainer.Refresh()
		}
	}

	// Синхронизация состояния при инициализации
	if apiText.Text != "" {
		closeBtn.Show()
		disableFields()
	} else {
		closeBtn.Hide()
	}

	return scenarioContainer
}

func (mp *ControlPage) createURLSection() *fyne.Container {
	mp.methodSelect = widget.NewSelect([]string{"GET", "POST"}, nil)
	mp.methodSelect.SetSelected("GET")

	mp.protocolSelect = widget.NewSelect([]string{"HTTP", "HTTPS"}, nil)
	mp.protocolSelect.SetSelected("HTTPS")

	mp.urlEntry = widget.NewEntry()
	mp.urlEntry.SetText("example.com")

	return container.NewHBox(
		container.NewGridWrap(fyne.NewSize(80, mp.methodSelect.MinSize().Height), mp.methodSelect),
		container.NewGridWrap(fyne.NewSize(100, mp.protocolSelect.MinSize().Height), mp.protocolSelect),
		container.NewGridWrap(fyne.NewSize(250, mp.urlEntry.MinSize().Height), mp.urlEntry),
	)
}

func (mp *ControlPage) createProxySection() *fyne.Container {
	mp.proxyEntry = widget.NewEntry()
	mp.proxyEntry.SetPlaceHolder("http://127.0.0.1:8080")
	return container.NewVBox(
		widget.NewLabel("Proxy (необязательно)"),
		container.NewGridWrap(fyne.NewSize(250, mp.proxyEntry.MinSize().Height), mp.proxyEntry),
	)
}

func (mp *ControlPage) createParamsSection() *fyne.Container {
	mp.reqCount = widget.NewEntry()
	mp.reqCount.SetText("10")

	mp.duration = widget.NewEntry()
	mp.duration.SetText("1")

	mp.loadType = widget.NewRadioGroup([]string{"Linear", "Incremental", "Waved"}, nil)
	mp.loadType.SetSelected("Linear")

	mp.usernameEntry = widget.NewEntry()
	mp.usernameEntry.SetPlaceHolder("Username")

	mp.passwordEntry = widget.NewEntry()
	mp.passwordEntry.SetPlaceHolder("Password")
	mp.passwordEntry.Password = true

	authAccordion := widget.NewAccordion(widget.NewAccordionItem("Basic Auth", container.NewVBox(
		mp.usernameEntry,
		mp.passwordEntry,
	)))

	mp.reqCount.OnChanged = func(s string) {
		if _, err := parseInt(s); err != nil {
			mp.reqCount.SetText("10")
		}
	}
	mp.duration.OnChanged = func(s string) {
		if _, err := parseInt(s); err != nil {
			mp.duration.SetText("1")
		}
	}

	return container.NewHBox(
		container.NewGridWrap(
			fyne.NewSize(120, widget.NewLabel("Request Count*").MinSize().Height+mp.reqCount.MinSize().Height),
			container.NewVBox(
				widget.NewLabel("Request Count*"),
				mp.reqCount,
			),
		),
		container.NewGridWrap(
			fyne.NewSize(120, widget.NewLabel("Duration (s)*").MinSize().Height+mp.duration.MinSize().Height),
			container.NewVBox(
				widget.NewLabel("Duration (s)*"),
				mp.duration,
			),
		),
		container.NewGridWrap(
			fyne.NewSize(120, widget.NewLabel("Load Type*").MinSize().Height+mp.loadType.MinSize().Height),
			container.NewVBox(
				widget.NewLabel("Load Type*"),
				mp.loadType,
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

func (mp *ControlPage) createCertFields(window fyne.Window) *fyne.Container {
	mp.certPathEntry = widget.NewEntry()
	mp.certPathEntry.SetPlaceHolder("Путь к сертификату")
	mp.certPathEntry.Disable()

	mp.certKeyPathEntry = widget.NewEntry()
	mp.certKeyPathEntry.SetPlaceHolder("Путь к ключу")
	mp.certKeyPathEntry.Disable()

	certLabel := widget.NewLabel("Не выбрано")
	keyLabel := widget.NewLabel("Не выбрано")

	certFilter := storage.NewExtensionFileFilter([]string{".crt", ".pem"})
	mp.selectCertButton = widget.NewButton("Выбрать сертификат", func() {
		fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				mp.certPathEntry.SetText(reader.URI().Path())
				certLabel.SetText(reader.URI().Name())
				reader.Close()
			}
		}, window)
		fileDialog.SetFilter(certFilter)
		fileDialog.Show()
	})

	keyFilter := storage.NewExtensionFileFilter([]string{".key"})
	mp.selectKeyButton = widget.NewButton("Выбрать ключ", func() {
		fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				mp.certKeyPathEntry.SetText(reader.URI().Path())
				keyLabel.SetText(reader.URI().Name())
				reader.Close()
			}
		}, window)
		fileDialog.SetFilter(keyFilter)
		fileDialog.Show()
	})

	certAccordion := widget.NewAccordion(widget.NewAccordionItem("Сертификаты", container.NewVBox(
		container.NewHBox(mp.selectCertButton, certLabel),
		container.NewHBox(mp.selectKeyButton, keyLabel),
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
