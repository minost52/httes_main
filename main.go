package main

import (
	"httes/ui"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func main() {
	myApp := app.New()
	myApp.Settings().SetTheme(theme.LightTheme())
	window := myApp.NewWindow("Httes")

	// Загружаем иконку
	icon, err := fyne.LoadResourceFromPath("assets/logo2.png")
	if err != nil {
		log.Println("Ошибка загрузки иконки:", err)
	} else {
		window.SetIcon(icon)
	}

	resultOutput := widget.NewTextGrid()
	// Создаем MainPage с пустым username и role (будут заполнены после авторизации)
	mp := ui.NewMainPage(myApp, resultOutput, window, "", "", icon)
	// Устанавливаем экран авторизации как начальный
	window.SetContent(ui.CreateLoginScreen(mp))

	window.CenterOnScreen()
	window.Resize(fyne.NewSize(970, 600))
	window.ShowAndRun()
}
