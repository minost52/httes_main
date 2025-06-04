package ui

import (
	"fmt"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (mp *ControlPage) createHistoryScreen(window fyne.Window, tabs *container.AppTabs) fyne.CanvasObject {
	// 1. Модель данных истории тестов
	type TestHistory struct {
		ID          string
		Name        string
		CreatedAt   time.Time
		Status      string // "success", "failed", "running"
		Description string
		Scenario    string // Выбранный сценарий (например, "API")
		APIDetails  string
		Request     string
		Response    string
		// Настройки теста
		Method       string // Например, "GET" или "POST"
		Protocol     string // Например, "HTTP" или "HTTPS"
		URL          string // URL из urlEntry
		Proxy        string // Прокси, если указано
		RequestCount string // Количество запросов
		Duration     string // Длительность
		LoadType     string // Тип нагрузки ("Linear", "Incremental", "Waved")
		Username     string // Имя пользователя для Basic Auth
		Password     string // Пароль для Basic Auth
		CertPath     string // Путь к сертификату
		CertKeyPath  string // Путь к ключу сертификата
	}

	// Инициализация истории тестов с примерами
	history := []TestHistory{
		{
			ID:           "1",
			Name:         "API User Flow Test",
			CreatedAt:    time.Now().Add(-time.Hour * 2),
			Status:       "success",
			Description:  "Тестирование CRUD операций пользователей",
			Scenario:     "API",
			APIDetails:   "POST /users, GET /users/{id}, PUT /users/{id}, DELETE /users/{id}",
			Request:      "POST /users {\"name\":\"test\",\"email\":\"test@example.com\"}",
			Response:     "201 Created {\"id\":123,\"name\":\"test\",\"email\":\"test@example.com\"}",
			Method:       "POST",
			Protocol:     "HTTPS",
			URL:          "example.com/users",
			Proxy:        "", // Прокси не используется
			RequestCount: "10",
			Duration:     "1",
			LoadType:     "Linear",
			Username:     "admin",
			Password:     "admin123",
			CertPath:     "",
			CertKeyPath:  "",
		},
		{
			ID:           "2",
			Name:         "Auth Stress Test",
			CreatedAt:    time.Now().Add(-time.Hour * 5),
			Status:       "failed",
			Description:  "Нагрузочное тестирование аутентификации",
			Scenario:     "API",
			APIDetails:   "POST /auth/login 100 запросов/сек",
			Request:      "POST /auth/login {\"username\":\"test\",\"password\":\"test123\"}",
			Response:     "429 Too Many Requests",
			Method:       "POST",
			Protocol:     "HTTPS",
			URL:          "example.com/auth/login",
			Proxy:        "http://127.0.0.1:8080",
			RequestCount: "100",
			Duration:     "10",
			LoadType:     "Incremental",
			Username:     "test",
			Password:     "test123",
			CertPath:     "",
			CertKeyPath:  "",
		},
		{
			ID:           "3",
			Name:         "Payment Validation",
			CreatedAt:    time.Now().Add(-time.Hour * 24),
			Status:       "success",
			Description:  "Проверка валидации платежных данных",
			Scenario:     "API",
			APIDetails:   "POST /payments с различными входными данными",
			Request:      "POST /payments {\"card\":\"4111111111111111\",\"expiry\":\"12/25\",\"cvv\":\"123\"}",
			Response:     "200 OK {\"status\":\"processed\"}",
			Method:       "POST",
			Protocol:     "HTTPS",
			URL:          "example.com/payments",
			Proxy:        "",
			RequestCount: "5",
			Duration:     "2",
			LoadType:     "Waved",
			Username:     "",
			Password:     "",
			CertPath:     "/path/to/cert.crt",
			CertKeyPath:  "/path/to/cert.key",
		},
	}

	// 2. Функция для отображения деталей теста
	showTestDetails := func(test TestHistory) {
		detailWindow := mp.app.NewWindow("Детали теста: " + test.Name)
		detailWindow.Resize(fyne.NewSize(800, 600))
		// Устанавливаем иконку для нового окна
		if mp.icon != nil {
			detailWindow.SetIcon(mp.icon)
		}

		// Создаем содержимое окна с деталями
		nameLabel := widget.NewLabelWithStyle(test.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

		// Статус с цветом
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

		// Секция: Детали сценария
		scenarioLabel := widget.NewLabelWithStyle("Сценарий:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		scenarioText := widget.NewLabel(test.Scenario)
		scenarioText.Wrapping = fyne.TextWrapWord

		// Секция: Настройки теста
		settingsLabel := widget.NewLabelWithStyle("Настройки теста:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		methodText := widget.NewLabel("Метод: " + test.Method)
		protocolText := widget.NewLabel("Протокол: " + test.Protocol)
		urlText := widget.NewLabel("URL: " + test.URL)

		// Прокси (если указано)
		proxyText := widget.NewLabel("Прокси: " + func() string {
			if test.Proxy == "" {
				return "Не используется"
			}
			return test.Proxy
		}())

		paramsText := widget.NewLabel(fmt.Sprintf("Параметры: %s запросов, %s сек, тип нагрузки: %s",
			test.RequestCount, test.Duration, test.LoadType))

		// Basic Auth (если указано)
		authText := widget.NewLabel("Basic Auth: " + func() string {
			if test.Username == "" && test.Password == "" {
				return "Не используется"
			}
			return fmt.Sprintf("Имя пользователя: %s, Пароль: %s", test.Username, test.Password)
		}())

		// Сертификаты (если указаны)
		certText := widget.NewLabel("Сертификаты: " + func() string {
			if test.CertPath == "" && test.CertKeyPath == "" {
				return "Не используются"
			}
			return fmt.Sprintf("Путь к сертификату: %s, Путь к ключу: %s", test.CertPath, test.CertKeyPath)
		}())

		// Секция: API детали
		apiLabel := widget.NewLabelWithStyle("API детали:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		apiDetails := widget.NewLabel(test.APIDetails)
		apiDetails.Wrapping = fyne.TextWrapWord

		// Секция: Запрос и ответ
		requestLabel := widget.NewLabelWithStyle("Запрос:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		requestText := widget.NewLabel(test.Request)
		requestText.Wrapping = fyne.TextWrapWord

		responseLabel := widget.NewLabelWithStyle("Ответ:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		responseText := widget.NewLabel(test.Response)
		responseText.Wrapping = fyne.TextWrapWord

		// Собираем содержимое
		content := container.NewVScroll(container.NewVBox(
			nameLabel,
			container.NewHBox(widget.NewLabel("Статус: "), statusText),
			timeLabel,
			widget.NewSeparator(),
			descLabel,
			widget.NewSeparator(),
			scenarioLabel,
			scenarioText,
			widget.NewSeparator(),
			settingsLabel,
			methodText,
			protocolText,
			urlText,
			proxyText,
			paramsText,
			authText,
			certText,
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
