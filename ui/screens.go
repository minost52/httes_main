package ui

import (
	"fmt"
	"httes/store"
	"image/color"
	"sort"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (mp *MainPage) createHomeScreen(
	onTestRun func(),
	onScenarios func(),
	onHistory func(),
) fyne.CanvasObject {
	var logoImg *canvas.Image
	if logo, err := fyne.LoadResourceFromPath("ui/logo.png"); err == nil {
		logoImg = canvas.NewImageFromResource(logo)
	} else {
		logoImg = canvas.NewImageFromResource(theme.FyneLogo())
	}
	logoImg.FillMode = canvas.ImageFillContain
	logoImg.SetMinSize(fyne.NewSize(300, 200))

	title := canvas.NewText("HTTES: Модуль нагрузочного тестирования", theme.PrimaryColor())
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.TextSize = 24
	title.Alignment = fyne.TextAlignCenter

	separator := widget.NewSeparator()

	description := widget.NewLabel(
		"Комплексное решение для тестирования производительности веб-сервисов и API.\n\n" +
			"• Создание и управление тестовыми сценариями\n" +
			"• Настройка различных профилей нагрузки\n" +
			"• Детальный мониторинг и анализ результатов\n" +
			"• История выполненных тестов")
	description.Wrapping = fyne.TextWrapWord
	description.Alignment = fyne.TextAlignCenter

	testRunBtn := widget.NewButtonWithIcon("Запуск теста", theme.MediaPlayIcon(), onTestRun)
	scenariosBtn := widget.NewButtonWithIcon("Сценарии", theme.DocumentCreateIcon(), onScenarios)
	historyBtn := widget.NewButtonWithIcon("История", theme.HistoryIcon(), onHistory)

	testRunBtn.Importance = widget.HighImportance
	scenariosBtn.Importance = widget.HighImportance
	historyBtn.Importance = widget.HighImportance

	// 🔘 Кнопка переключения темы
	var themeButton *widget.Button
	themeButton = widget.NewButton("☀️ Светлая тема", func() {
		if mp.isDarkMode {
			fyne.CurrentApp().Settings().SetTheme(theme.LightTheme())
			themeButton.SetText("🌙 Темная тема")
		} else {
			fyne.CurrentApp().Settings().SetTheme(theme.DarkTheme())
			themeButton.SetText("☀️ Светлая тема")
		}
		mp.isDarkMode = !mp.isDarkMode
	})
	themeButton.Importance = widget.LowImportance

	content := container.NewVBox(
		container.NewPadded(logoImg),
		container.NewCenter(title),
		container.NewVBox(layout.NewSpacer(), separator, layout.NewSpacer()),
		container.NewVBox(layout.NewSpacer(), widget.NewLabel(" ")),
		container.NewPadded(description),
		container.NewVBox(layout.NewSpacer(), widget.NewLabel(" ")),
		container.NewGridWithColumns(3,
			container.NewPadded(testRunBtn),
			container.NewPadded(scenariosBtn),
			container.NewPadded(historyBtn),
		),
		container.NewVBox(layout.NewSpacer(), widget.NewLabel(" ")),
		container.NewCenter(themeButton), // 🎯 Добавлено: кнопка смены темы
	)

	return container.NewPadded(
		container.NewVBox(
			container.NewCenter(content),
			layout.NewSpacer(),
		),
	)
}

func (mp *MainPage) createScenariosScreen(window fyne.Window, tabs *container.AppTabs) fyne.CanvasObject {
	// 1. Модель данных сценариев
	scenarios := []struct {
		Name        string
		CreatedAt   time.Time
		Description string
	}{
		{
			Name:        "API User Flow",
			CreatedAt:   time.Date(2023, 5, 15, 14, 30, 0, 0, time.UTC),
			Description: "Полный цикл работы с пользователями (CRUD)",
		},
		{
			Name:        "Auth Stress Test",
			CreatedAt:   time.Date(2023, 5, 20, 9, 15, 0, 0, time.UTC),
			Description: "Тестирование нагрузки на эндпоинты аутентификации",
		},
		{
			Name:        "Payment Validation",
			CreatedAt:   time.Date(2023, 6, 1, 16, 45, 0, 0, time.UTC),
			Description: "Проверка валидации платежных данных",
		},
	}

	// 2. Состояние сортировки
	sortNewestFirst := true
	refreshList := func() {}

	// 3. Элементы управления
	searchEntry := widget.NewEntry()
	searchEntry.PlaceHolder = "Поиск по названию..."

	// Кнопка сортировки
	sortBtn := widget.NewButtonWithIcon("Новее", theme.MenuDropDownIcon(), nil)
	sortMenu := fyne.NewMenu("",
		fyne.NewMenuItem("Новее", func() {
			sortNewestFirst = true
			sortBtn.SetText("Новее")
			refreshList()
		}),
		fyne.NewMenuItem("Старее", func() {
			sortNewestFirst = false
			sortBtn.SetText("Старее")
			refreshList()
		}),
	)

	sortBtn.OnTapped = func() {
		popUp := widget.NewPopUpMenu(sortMenu, window.Canvas())
		pos := fyne.CurrentApp().Driver().AbsolutePositionForObject(sortBtn)
		popUp.ShowAtPosition(pos.Add(fyne.NewPos(0, sortBtn.Size().Height)))
	}

	// Кнопка создания нового сценария
	newScenarioBtn := widget.NewButtonWithIcon("Новый сценарий", theme.ContentAddIcon(), func() {
		scenarioWindow := mp.app.NewWindow("Создать новый сценарий")
		scenarioWindow.SetContent(mp.createScenarioEditorContent(scenarioWindow))
		scenarioWindow.Resize(fyne.NewSize(800, 600))
		scenarioWindow.Show()
	})

	// 4. Создаем список сценариев
	list := widget.NewList(
		func() int { return len(scenarios) },
		func() fyne.CanvasObject {
			return container.NewBorder(
				nil,
				nil,
				nil,
				container.NewHBox(
					widget.NewButtonWithIcon("", theme.MediaPlayIcon(), nil),
					widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), nil),
					widget.NewButtonWithIcon("", theme.DeleteIcon(), nil),
				),
				container.NewVBox(
					widget.NewLabel("Template"),
					widget.NewLabel(""),
				),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			// Сортируем перед отображением
			if sortNewestFirst {
				sort.Slice(scenarios, func(i, j int) bool {
					return scenarios[i].CreatedAt.After(scenarios[j].CreatedAt)
				})
			} else {
				sort.Slice(scenarios, func(i, j int) bool {
					return scenarios[i].CreatedAt.Before(scenarios[j].CreatedAt)
				})
			}

			scenario := scenarios[id]
			container := item.(*fyne.Container)

			// Заполняем информацию о сценарии
			infoContainer := container.Objects[0].(*fyne.Container)
			infoContainer.Objects[0].(*widget.Label).SetText(scenario.Name)
			infoContainer.Objects[1].(*widget.Label).SetText(
				fmt.Sprintf("Создан: %s | %s",
					scenario.CreatedAt.Format("02.01.2006 15:04"),
					scenario.Description),
			)

			// Настраиваем кнопки
			buttons := container.Objects[1].(*fyne.Container)
			// Кнопка редактирования
			buttons.Objects[1].(*widget.Button).OnTapped = func() {
				tabs.SelectIndex(3)
			}

			// Кнопка удаления
			buttons.Objects[2].(*widget.Button).OnTapped = func() {
				dialog.ShowConfirm("Удаление", fmt.Sprintf("Удалить '%s'?", scenario.Name),
					func(ok bool) {
						if ok {
							// Логика удаления
						}
					}, window)
			}
		},
	)

	// Функция обновления списка
	refreshList = func() {
		list.Refresh()
	}

	// 5. Собираем интерфейс
	header := container.NewHBox(
		widget.NewButtonWithIcon("Назад", theme.NavigateBackIcon(), func() {
			tabs.SelectIndex(1)
		}),
		widget.NewLabelWithStyle("Выберите сценарий", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		layout.NewSpacer(),
		widget.NewLabel(fmt.Sprintf("Всего: %d", len(scenarios))),
	)

	searchPanel := container.NewBorder(
		nil,
		nil,
		nil,
		container.NewHBox(
			sortBtn,
			newScenarioBtn,
		),
		searchEntry,
	)

	content := container.NewBorder(
		container.NewVBox(
			header,
			widget.NewSeparator(),
			searchPanel,
		),
		nil,
		nil,
		nil,
		list,
	)

	return container.NewPadded(
		container.NewGridWrap(
			fyne.NewSize(970, 600),
			content,
		),
	)
}

func (mp *MainPage) createTestRunScreen(window fyne.Window, tabs *container.AppTabs) fyne.CanvasObject {
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
	controls := container.NewVBox(
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

	// Результаты
	results := container.NewVScroll(ui.resultOutput)
	results.SetMinSize(fyne.NewSize(300, 150))

	// Создаем заголовок с текстом и кнопкой истории
	resultHeader := container.NewBorder(
		nil,
		nil,
		widget.NewLabel("Результаты теста"), // Текст слева
		widget.NewButtonWithIcon("История тестов", theme.HistoryIcon(), func() {
			tabs.SelectTabIndex(3) // Предполагая, что вкладка истории имеет индекс 2
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

	right := container.NewVBox(
		container.NewVBox(
			widget.NewLabelWithStyle("Метрики теста", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewSeparator(),
			CreateLoadTestCharts(),
		),
	)

	layout := container.NewHBox(
		left,
		widget.NewSeparator(),
		right,
	)

	return layout
}

func (mp *MainPage) createScenariosSection(tabs *container.AppTabs) fyne.CanvasObject {
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

// func (mp *MainPage) createLoadProfilesScreen(window fyne.Window) fyne.CanvasObject {
// 	// Список профилей
// 	list := widget.NewList(
// 		func() int { return store.LoadProfileCount() },
// 		func() fyne.CanvasObject {
// 			return container.NewHBox(
// 				widget.NewLabel("Template"),
// 				widget.NewButton("Редактировать", nil),
// 			)
// 		},
// 		func(id widget.ListItemID, item fyne.CanvasObject) {
// 			profile := store.GetLoadProfile(id)
// 			container := item.(*fyne.Container)
// 			container.Objects[0].(*widget.Label).SetText(profile.Name)
// 			container.Objects[1].(*widget.Button).OnTapped = func() {
// 				dialog.ShowInformation("Редактирование",
// 					fmt.Sprintf("Редактирование профиля %s", profile.Name),
// 					window)
// 			}
// 		},
// 	)

// 	// Форма редактирования
// 	nameEntry := widget.NewEntry()
// 	nameEntry.PlaceHolder = "Название профиля"

// 	typeSelect := widget.NewSelect([]string{"Constant", "Incremental"}, nil)
// 	typeSelect.PlaceHolder = "Тип профиля"

// 	configEntry := widget.NewMultiLineEntry()
// 	configEntry.PlaceHolder = "JSON конфигурация"
// 	configEntry.SetMinRowsVisible(3)

// 	saveBtn := widget.NewButtonWithIcon("Сохранить", theme.DocumentSaveIcon(), func() {
// 		if nameEntry.Text == "" {
// 			dialog.ShowInformation("Ошибка", "Введите название профиля", window)
// 			return
// 		}

// 		store.AddLoadProfile(store.LoadProfile{
// 			ID:   store.LoadProfileCount() + 1,
// 			Name: nameEntry.Text,
// 			Type: typeSelect.Selected,
// 		})
// 		dialog.ShowInformation("Сохранено", "Профиль сохранён", window)

// 		// Сброс формы
// 		nameEntry.SetText("")
// 		typeSelect.ClearSelected()
// 		configEntry.SetText("")
// 	})

// 	// Основной layout
// 	return container.NewPadded(
// 		container.NewVBox(
// 			container.NewVScroll(list),
// 			widget.NewSeparator(),
// 			widget.NewLabel("Новый профиль:"),
// 			nameEntry,
// 			typeSelect,
// 			configEntry,
// 			container.NewCenter(saveBtn),
// 		),
// 	)
// }

func (mp *MainPage) createScenarioEditorContent(window fyne.Window) fyne.CanvasObject {
	// Поля сценария
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Тестирование эндпоинтов /users")
	nameEntry.SetText("")

	descEntry := widget.NewEntry()
	descEntry.SetPlaceHolder("Проверка работы GET, POST, PUT, DELETE, PATCH")
	descEntry.SetText("")

	profileSelect := widget.NewSelect([]string{"Linear", "incremental", "Waved"}, nil)
	profileSelect.SetSelected("Linear")

	jsonEditor := widget.NewMultiLineEntry()
	jsonEditor.SetPlaceHolder(`Например: {
  "steps": [
    {
      "name": "Create user",
      "request": {
        "method": "POST",
        "url": "/users"
      }
    }
  ]
}`)
	jsonEditor.SetMinRowsVisible(6)

	// Список конечных точек
	var endpoints []store.Endpoint

	endpointURL := widget.NewEntry()
	endpointURL.SetPlaceHolder(`https://api.example.com/users`)
	endpointURL.SetMinRowsVisible(3)

	endpointHeaders := widget.NewMultiLineEntry()
	endpointHeaders.SetPlaceHolder(`Например: {
  "Content-Type": "application/json",
  "Authorization": "Bearer token"
}`)
	endpointHeaders.SetMinRowsVisible(3)

	// Сертификаты
	certPath := widget.NewEntry()
	certPath.SetPlaceHolder("Путь к сертификату (например: /path/to/cert.pem)")
	certPath.Disable()

	certKeyPath := widget.NewEntry()
	certKeyPath.SetPlaceHolder("Путь к ключу (например: /path/to/key.pem)")
	certKeyPath.Disable()

	certLabel := widget.NewLabel("Не выбрано")
	keyLabel := widget.NewLabel("Не выбрано")

	var selectedCert *store.Cert

	// Фильтры файлов
	certFilter := storage.NewExtensionFileFilter([]string{".crt", ".pem"})
	keyFilter := storage.NewExtensionFileFilter([]string{".key"})

	// Кнопка выбора сертификата
	selectCertBtn := widget.NewButton("Выбрать сертификат", func() {
		fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				certPath.SetText(reader.URI().Path())
				certLabel.SetText(reader.URI().Name())
				reader.Close()
			}
		}, window)
		fileDialog.SetFilter(certFilter)
		fileDialog.Show()
	})

	// Кнопка выбора ключа
	selectKeyBtn := widget.NewButton("Выбрать ключ", func() {
		fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				certKeyPath.SetText(reader.URI().Path())
				keyLabel.SetText(reader.URI().Name())
				reader.Close()
			}
		}, window)
		fileDialog.SetFilter(keyFilter)
		fileDialog.Show()
	})

	// Кнопка "Сохранить"
	saveBtn := widget.NewButton("Сохранить сценарий", func() {
		if strings.TrimSpace(nameEntry.Text) == "" {
			dialog.NewInformation("Ошибка", "Название сценария обязательно", window).Show()
			return
		}

		// Добавление конечной точки, если заполнены поля
		if endpointURL.Text != "" {
			endpoints = append(endpoints, store.Endpoint{
				URL:     strings.TrimSpace(endpointURL.Text),
				Headers: endpointHeaders.Text,
			})
		}

		// Обработка сертификата
		if certPath.Text != "" && certKeyPath.Text != "" {
			selectedCert = &store.Cert{
				Name: certLabel.Text,
				Path: certPath.Text,
				Key:  certKeyPath.Text,
			}
		}

		store.AddScenario(store.Scenario{
			ID:          store.ScenarioCount() + 1,
			Name:        nameEntry.Text,
			Description: descEntry.Text,
			Cert:        selectedCert,
		})

		dialog.NewInformation("Сценарий", "Сценарий сохранён", window).Show()

		// Очистка формы
		nameEntry.SetText("")
		descEntry.SetText("")
		jsonEditor.SetText("")
		endpointURL.SetText("")
		endpointHeaders.SetText("")
		certPath.SetText("")
		certKeyPath.SetText("")
		certLabel.SetText("Не выбрано")
		keyLabel.SetText("Не выбрано")
	})
	// Левая колонка
	leftPanel := container.NewVBox(
		widget.NewLabelWithStyle("Создание сценария", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Название:"),
		nameEntry,
		widget.NewLabel("Описание:"),
		descEntry,
		widget.NewLabel("Профиль нагрузки:"),
		profileSelect,
		widget.NewLabelWithStyle("JSON сценария:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		jsonEditor,
	)

	// Правая колонка
	rightPanel := container.NewVBox(
		widget.NewLabelWithStyle("Конечные точки", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("URL:"),
		endpointURL,
		widget.NewLabel("Заголовки (JSON):"),
		endpointHeaders,
		widget.NewLabelWithStyle("Сертификаты", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewHBox(selectCertBtn, certLabel),
		container.NewHBox(selectKeyBtn, keyLabel),
	)

	// Контейнер 2 колонки
	content := container.NewGridWithColumns(2, leftPanel, rightPanel)

	// Создаем основной контейнер с контентом и кнопкой сохранения
	mainContent := container.NewVBox(
		content,
		widget.NewSeparator(),
		container.NewCenter(saveBtn),
	)

	// Обёртка с фиксированным размером
	return container.NewCenter(container.NewGridWrap(fyne.NewSize(970, 550), mainContent))
}

func (mp *MainPage) createHistoryScreen(window fyne.Window, tabs *container.AppTabs) fyne.CanvasObject {
	// 1. Модель данных истории тестов
	type TestHistory struct {
		ID          string
		Name        string
		CreatedAt   time.Time
		Status      string // "success", "failed", "running"
		Description string
		APIDetails  string
		Request     string
		Response    string
	}

	// Инициализация истории тестов с примерами
	history := []TestHistory{
		{
			ID:          "1",
			Name:        "API User Flow Test",
			CreatedAt:   time.Now().Add(-time.Hour * 2),
			Status:      "success",
			Description: "Тестирование CRUD операций пользователей",
			APIDetails:  "POST /users, GET /users/{id}, PUT /users/{id}, DELETE /users/{id}",
			Request:     "POST /users {\"name\":\"test\",\"email\":\"test@example.com\"}",
			Response:    "201 Created {\"id\":123,\"name\":\"test\",\"email\":\"test@example.com\"}",
		},
		{
			ID:          "2",
			Name:        "Auth Stress Test",
			CreatedAt:   time.Now().Add(-time.Hour * 5),
			Status:      "failed",
			Description: "Нагрузочное тестирование аутентификации",
			APIDetails:  "POST /auth/login 100 запросов/сек",
			Request:     "POST /auth/login {\"username\":\"test\",\"password\":\"test123\"}",
			Response:    "429 Too Many Requests",
		},
		{
			ID:          "3",
			Name:        "Payment Validation",
			CreatedAt:   time.Now().Add(-time.Hour * 24),
			Status:      "success",
			Description: "Проверка валидации платежных данных",
			APIDetails:  "POST /payments с различными входными данными",
			Request:     "POST /payments {\"card\":\"4111111111111111\",\"expiry\":\"12/25\",\"cvv\":\"123\"}",
			Response:    "200 OK {\"status\":\"processed\"}",
		},
	}

	// 2. Функция для отображения деталей теста
	showTestDetails := func(test TestHistory) {
		detailWindow := mp.app.NewWindow("Детали теста: " + test.Name)
		detailWindow.Resize(fyne.NewSize(800, 600))

		// Создаем содержимое окна с деталями
		nameLabel := widget.NewLabelWithStyle(test.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

		// Правильный способ создания цветного текста
		statusText := canvas.NewText("Статус: "+test.Status, nil)
		if test.Status == "success" {
			statusText.Color = color.NRGBA{R: 0, G: 180, B: 0, A: 255}
		} else if test.Status == "failed" {
			statusText.Color = color.NRGBA{R: 180, G: 0, B: 0, A: 255}
		}
		statusText.TextStyle.Bold = true

		timeLabel := widget.NewLabel("Время выполнения: " + test.CreatedAt.Format("02.01.2006 15:04:05"))
		descLabel := widget.NewLabel(test.Description)
		descLabel.Wrapping = fyne.TextWrapWord

		apiLabel := widget.NewLabelWithStyle("API детали:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		apiDetails := widget.NewLabel(test.APIDetails)
		apiDetails.Wrapping = fyne.TextWrapWord

		requestLabel := widget.NewLabelWithStyle("Запрос:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		requestText := widget.NewLabel(test.Request)
		requestText.Wrapping = fyne.TextWrapWord

		responseLabel := widget.NewLabelWithStyle("Ответ:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		responseText := widget.NewLabel(test.Response)
		responseText.Wrapping = fyne.TextWrapWord

		content := container.NewVScroll(container.NewVBox(
			nameLabel,
			container.NewHBox(widget.NewLabel("Статус: "), statusText),
			timeLabel,
			widget.NewSeparator(),
			descLabel,
			widget.NewSeparator(),
			apiLabel,
			apiDetails,
			widget.NewSeparator(),
			requestLabel,
			requestText,
			widget.NewSeparator(),
			responseLabel,
			responseText,
		))

		detailWindow.SetContent(content)
		detailWindow.Show()
	}

	// 3. Создаем список истории тестов
	list := widget.NewList(
		func() int { return len(history) },
		func() fyne.CanvasObject {
			return container.NewBorder(
				nil,
				nil,
				nil,
				container.NewHBox(), // Контейнер для цветного статуса
				container.NewVBox(
					widget.NewLabel("Название теста"),
					widget.NewLabel("Время выполнения"),
				),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			test := history[id]
			container := item.(*fyne.Container)

			// Основная информация
			infoContainer := container.Objects[0].(*fyne.Container)
			infoContainer.Objects[0].(*widget.Label).SetText(test.Name)
			infoContainer.Objects[1].(*widget.Label).SetText(test.CreatedAt.Format("02.01.2006 15:04:05"))

			// Создаем цветной текст для статуса
			statusText := canvas.NewText(test.Status, nil)
			if test.Status == "success" {
				statusText.Color = color.NRGBA{R: 0, G: 180, B: 0, A: 255}
			} else if test.Status == "failed" {
				statusText.Color = color.NRGBA{R: 180, G: 0, B: 0, A: 255}
			}
			statusText.TextStyle.Bold = true

			// Очищаем и добавляем новый статус
			statusContainer := container.Objects[1].(*fyne.Container)
			statusContainer.Objects = []fyne.CanvasObject{statusText}
		},
	)

	// Обработчик нажатия на элемент списка
	list.OnSelected = func(id widget.ListItemID) {
		showTestDetails(history[id])
		list.Unselect(id) // Снимаем выделение после выбора
	}

	// 4. Кнопка очистки истории
	clearHistoryBtn := widget.NewButtonWithIcon("Очистить историю", theme.DeleteIcon(), func() {
		dialog.ShowConfirm("Очистка истории", "Вы уверены, что хотите очистить всю историю тестов?", func(ok bool) {
			if ok {
				// Очищаем историю
				history = []TestHistory{}
				list.Refresh()
			}
		}, window)
	})

	// 5. Собираем интерфейс
	header := container.NewHBox(
		widget.NewButtonWithIcon("Назад", theme.NavigateBackIcon(), func() {
			tabs.SelectIndex(1)
		}),
		widget.NewLabelWithStyle("История тестов", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		layout.NewSpacer(),
		widget.NewLabel(fmt.Sprintf("Всего: %d", len(history))),
	)

	content := container.NewBorder(
		container.NewVBox(
			header,
			widget.NewSeparator(),
			container.NewHBox(
				layout.NewSpacer(),
				clearHistoryBtn,
			),
		),
		nil,
		nil,
		nil,
		list,
	)

	return container.NewPadded(
		container.NewGridWrap(
			fyne.NewSize(970, 600),
			content,
		),
	)
}
