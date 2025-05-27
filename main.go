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
	myApp.Settings().SetTheme(theme.DarkTheme())
	window := myApp.NewWindow("Httes")
	// Загружаем иконку
	icon, err := fyne.LoadResourceFromPath("ui/logo2.png")
	if err != nil {
		log.Println("Ошибка загрузки иконки:", err)
	} else {
		window.SetIcon(icon)
	}

	resultOutput := widget.NewTextGrid()
	// Передаем иконку в NewMainPage
	mp := ui.NewMainPage(myApp, resultOutput, window, icon)
	content := mp.CreateUI(window)

	window.SetContent(content)
	window.CenterOnScreen()
	window.Resize(fyne.NewSize(970, 600))
	window.ShowAndRun()
}
