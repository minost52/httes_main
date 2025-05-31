package ui

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
