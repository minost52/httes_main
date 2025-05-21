package loadtest

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

// Глобальные переменные для UI элементов (уже определены)
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

func createURLSection() *fyne.Container {
	methodSelect = widget.NewSelect([]string{"GET", "POST"}, nil)
	methodSelect.SetSelected("GET")

	protocolSelect = widget.NewSelect([]string{"HTTP", "HTTPS"}, nil)
	protocolSelect.SetSelected("HTTPS")

	urlEntry = widget.NewEntry()
	urlEntry.SetText("example.com")

	urlSection := container.NewHBox(
		container.NewGridWrap(fyne.NewSize(80, methodSelect.MinSize().Height), methodSelect),
		container.NewGridWrap(fyne.NewSize(100, protocolSelect.MinSize().Height), protocolSelect),
		container.NewGridWrap(fyne.NewSize(250, urlEntry.MinSize().Height), urlEntry),
	)

	return urlSection
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

	// Создаём поля аутентификации
	usernameEntry = widget.NewEntry()
	usernameEntry.SetPlaceHolder("Username")

	passwordEntry = widget.NewEntry()
	passwordEntry.SetPlaceHolder("Password")
	passwordEntry.Password = true // Скрываем ввод пароля

	// Аккордеон для аутентификации
	authAccordion := widget.NewAccordion(widget.NewAccordionItem("Basic Auth", container.NewVBox(
		usernameEntry,
		passwordEntry,
	)))

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
		// Добавляем аутентификацию справа от Load Type
		container.NewGridWrap(
			fyne.NewSize(150, authAccordion.MinSize().Height),
			container.NewVBox(
				authAccordion,
			),
		),
	)
}

func createCertFields(window fyne.Window) *fyne.Container {
	// Инициализация глобальных переменных
	certPathEntry = widget.NewEntry()
	certPathEntry.SetPlaceHolder("Путь к сертификату")
	certPathEntry.Disable()

	certKeyPathEntry = widget.NewEntry()
	certKeyPathEntry.SetPlaceHolder("Путь к ключу")
	certKeyPathEntry.Disable()

	// Метки для отображения имён файлов
	certLabel := widget.NewLabel("Не выбрано")
	keyLabel := widget.NewLabel("Не выбрано")

	// Фильтр для файлов сертификатов (.crt, .pem)
	certFilter := storage.NewExtensionFileFilter([]string{".crt", ".pem"})
	selectCertButton := widget.NewButton("Выбрать сертификат", func() {
		fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				certPathEntry.SetText(reader.URI().Path())
				certLabel.SetText(reader.URI().Name()) // Отображаем только имя файла
				reader.Close()
			}
		}, window)
		fileDialog.SetFilter(certFilter)
		fileDialog.Show()
	})

	// Фильтр для файлов ключей (.key)
	keyFilter := storage.NewExtensionFileFilter([]string{".key"})
	selectKeyButton := widget.NewButton("Выбрать ключ", func() {
		fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				certKeyPathEntry.SetText(reader.URI().Path())
				keyLabel.SetText(reader.URI().Name()) // Отображаем только имя файла
				reader.Close()
			}
		}, window)
		fileDialog.SetFilter(keyFilter)
		fileDialog.Show()
	})

	// Аккордеон для сертификатов
	certAccordion := widget.NewAccordion(widget.NewAccordionItem("Сертификаты", container.NewVBox(
		container.NewHBox(selectCertButton, certLabel),
		container.NewHBox(selectKeyButton, keyLabel),
	)))

	return container.NewVBox(certAccordion)
}
