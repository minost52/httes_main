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
	icon, err := fyne.LoadResourceFromPath("ui/logo.png")
	if err != nil {
		log.Println("Ошибка загрузки иконки:", err)
	} else {
		window.SetIcon(icon)
	}

	resultOutput := widget.NewTextGrid()
	mp := ui.NewMainPage(myApp, resultOutput, window) // Добавляем window как третий аргумент
	content := mp.CreateUI(window)

	// Устанавливаем начальный размер окна
	window.SetContent(content)
	window.CenterOnScreen()
	window.Resize(fyne.NewSize(970, 600))
	window.ShowAndRun()
}
