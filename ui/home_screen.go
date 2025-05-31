package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (mp *ControlPage) createHomeScreen(
	onTestRun func(),
	onScenarios func(),
	onHistory func(),
) fyne.CanvasObject {
	var logoImg *canvas.Image
	if logo, err := fyne.LoadResourceFromPath("assets/logo.png"); err == nil {
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

	// Создаем элементы для верхней панели
	profileIcon := widget.NewIcon(theme.AccountIcon())
	usernameLabel := widget.NewLabel(mp.username)
	logoutButton := widget.NewButtonWithIcon("Выйти", theme.LogoutIcon(), func() {
		// Обнуляем графики перед выходом
		mp.ResetCharts()
		mp.username = ""
		mp.role = ""
		mp.window.SetContent(CreateLoginScreen(mp))
	})

	// Создаем контейнер для верхней панели (иконка, имя, кнопка выхода)
	topBar := container.NewHBox(
		layout.NewSpacer(),
		profileIcon,
		usernameLabel,
		logoutButton,
	)

	// Кнопка переключения темы
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
		container.NewPadded(description),
		container.NewVBox(layout.NewSpacer(), widget.NewLabel(" ")),
		container.NewGridWithColumns(3,
			container.NewPadded(testRunBtn),
			container.NewPadded(scenariosBtn),
			container.NewPadded(historyBtn),
		),
		container.NewVBox(layout.NewSpacer(), widget.NewLabel(" ")),
		container.NewCenter(themeButton),
	)

	// Основной контейнер с верхней панелью и основным содержимым
	return container.NewBorder(
		topBar, // Верхняя панель
		nil,    // Нижняя панель (нет)
		nil,    // Левая панель (нет)
		nil,    // Правая панель (нет)
		container.NewPadded(
			container.NewVBox(
				container.NewCenter(content),
				layout.NewSpacer(),
			),
		),
	)
}
