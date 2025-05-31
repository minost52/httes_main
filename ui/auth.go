package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Структуры для авторизации (совместимы с сервером)
type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	User         *User  `json:"user,omitempty"`
	Token        string `json:"token,omitempty"`
	RequestLimit int    `json:"request_limit,omitempty"`
}

type User struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	Password     string `json:"password,omitempty"`
	Role         string `json:"role"`
	RequestCount int    `json:"request_count"`
	LastReset    string `json:"last_reset"`
	CreatedAt    string `json:"created_at"`
	IsActive     bool   `json:"is_active"`
}

func CreateLoginScreen(mp *ControlPage) fyne.CanvasObject {
	usernameEntry := widget.NewEntry()
	usernameEntry.SetText("keklol")
	usernameEntry.SetPlaceHolder("Имя пользователя")
	usernameEntry.Validator = func(s string) error {
		if len(s) < 3 {
			return fmt.Errorf("имя пользователя слишком короткое")
		}
		return nil
	}

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetText("keklol")
	passwordEntry.SetPlaceHolder("Пароль")
	passwordEntry.Validator = func(s string) error {
		if len(s) < 6 {
			return fmt.Errorf("пароль должен содержать минимум 6 символов")
		}
		return nil
	}

	loginButton := widget.NewButtonWithIcon("Войти", theme.LoginIcon(), func() {
		mp.ResetCharts()
		if err := usernameEntry.Validate(); err != nil {
			dialog.ShowError(err, mp.window)
			return
		}
		if err := passwordEntry.Validate(); err != nil {
			dialog.ShowError(err, mp.window)
			return
		}

		loading := dialog.NewCustom("Авторизация", "Отмена",
			widget.NewProgressBarInfinite(), mp.window)
		loading.Show()

		go func() {
			authReq := AuthRequest{
				Username: usernameEntry.Text,
				Password: passwordEntry.Text,
			}

			reqBody, err := json.Marshal(authReq)
			if err != nil {
				loading.Hide()
				dialog.ShowError(fmt.Errorf("ошибка подготовки запроса: %v", err), mp.window)
				return
			}

			resp, err := http.Post("http://localhost:8080/auth", "application/json", bytes.NewBuffer(reqBody))
			if err != nil {
				loading.Hide()
				dialog.ShowError(fmt.Errorf("ошибка соединения с сервером: %v", err), mp.window)
				return
			}
			defer resp.Body.Close()

			var authResp AuthResponse
			if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
				loading.Hide()
				dialog.ShowError(fmt.Errorf("ошибка чтения ответа сервера: %v", err), mp.window)
				return
			}

			loading.Hide()
			if authResp.Success {
				// Успешная авторизация: сохраняем имя пользователя и роль, переходим к главному интерфейсу
				mp.username = authResp.User.Username
				mp.role = authResp.User.Role
				mp.window.SetContent(mp.CreateUI(mp.window))
				usernameEntry.SetText("")
				passwordEntry.SetText("")
			} else {
				dialog.ShowError(fmt.Errorf(authResp.Message), mp.window)
			}
		}()
	})
	loginButton.Importance = widget.HighImportance

	formTitle := widget.NewLabel("Вход в систему")
	formTitle.Alignment = fyne.TextAlignCenter
	formTitle.TextStyle = fyne.TextStyle{Bold: true}

	formWidth := float32(300)

	form := container.NewVBox(
		formTitle,
		usernameEntry,
		passwordEntry,
		widget.NewSeparator(),
		loginButton,
	)

	mp.window.Resize(fyne.NewSize(970, 600))
	mp.window.CenterOnScreen()
	mp.window.SetContent(mp.CreateUI(mp.window))

	// Обёртка с фиксированной шириной (фон + форма)
	fixedWidth := canvas.NewRectangle(nil)
	fixedWidth.SetMinSize(fyne.NewSize(formWidth, 0))

	formWithFixedWidth := container.NewMax(
		fixedWidth,
		form,
	)

	// Центрирование по горизонтали и вертикали
	centered := container.NewVBox(
		layout.NewSpacer(),
		container.NewHBox(
			layout.NewSpacer(),
			formWithFixedWidth,
			layout.NewSpacer(),
		),
		layout.NewSpacer(),
	)

	// Обёртка в max-контейнер для адаптации к окну
	return container.NewMax(centered)
}
