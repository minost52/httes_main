package ui

import (
	"fmt"
	"httes/store"
	"math/rand"
	"sort"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Metric представляет метрику теста с именем, длительностью и порядком.
type Metric struct {
	Name     string
	Duration float64 // в миллисекундах
	Order    int
}

type LoadTestUI struct {
	isRunning    bool
	uiUpdateChan chan uiUpdate
	startBtn     *widget.Button
	stopBtn      *widget.Button
	app          fyne.App
	window       fyne.Window
	resultOutput *widget.TextGrid
	progressBar  *widget.ProgressBar
	progressText *widget.Label
}

type uiUpdate struct {
	errMsg       string
	outputText   string
	startEnabled bool
	stopEnabled  bool
	status       string  // "loading", "completed", или "" для сброса
	progress     float64 // Добавлено для прогресса
	progressText string  // текст для Request Avg Duration
}

var uiUpdaterOnce sync.Once

func NewLoadTestUI(app fyne.App, window fyne.Window) *LoadTestUI {
	resultOutput := widget.NewTextGrid()
	progressBar := widget.NewProgressBar()
	progressBar.Min = 0.0
	progressBar.Max = 100.0
	progressText := widget.NewLabel("Request Avg Duration 0.000s")
	loadingIcon := widget.NewProgressBarInfinite()
	loadingIcon.Hide()
	checkIcon := canvas.NewImageFromResource(theme.ConfirmIcon())
	checkIcon.FillMode = canvas.ImageFillContain
	checkIcon.SetMinSize(fyne.NewSize(24, 24))
	checkIcon.Hide()

	ui := &LoadTestUI{
		app:          app,
		window:       window,
		resultOutput: resultOutput,
		progressBar:  progressBar,
		progressText: progressText,
		startBtn:     widget.NewButton("Start Load Test", nil),
		stopBtn:      widget.NewButton("Stop", nil),
	}
	ui.stopBtn.Disable()
	ui.initUIUpdater()
	return ui
}

func (ui *LoadTestUI) initUIUpdater() {
	uiUpdaterOnce.Do(func() {
		ui.uiUpdateChan = make(chan uiUpdate, 1000)
		go func() {
			defer func() {
				close(ui.uiUpdateChan)
				if ui.startBtn != nil {
					ui.startBtn.Enable()
				}
				if ui.stopBtn != nil {
					ui.stopBtn.Disable()
				}
			}()
			for update := range ui.uiUpdateChan {
				if update.errMsg != "" {
					dialog.NewInformation("Error", update.errMsg, ui.window).Show()
				}
				if update.outputText != "" && ui.resultOutput != nil {
					ui.resultOutput.SetText(update.outputText)
					ui.resultOutput.Refresh()
				}
				if update.progressText != "" && ui.progressText != nil {
					ui.progressText.SetText(update.progressText)
					ui.progressText.Refresh()
				}
				if update.startEnabled && ui.startBtn != nil {
					ui.startBtn.Enable()
				} else if ui.startBtn != nil && !update.startEnabled {
					ui.startBtn.Disable()
				}
				if update.stopEnabled && ui.stopBtn != nil {
					ui.stopBtn.Enable()
				} else if ui.stopBtn != nil && !update.stopEnabled {
					ui.stopBtn.Disable()
				}
				if ui.progressBar != nil && update.progress >= 0 {
					ui.progressBar.SetValue(update.progress)
					ui.progressBar.Refresh()
				}
			}
		}()
	})
}

func (ui *LoadTestUI) safeUpdateUI(update uiUpdate) {
	if update.startEnabled && ui.startBtn != nil {
		ui.startBtn.Enable()
	} else if ui.startBtn != nil && !update.startEnabled {
		ui.startBtn.Disable()
	}
	if update.stopEnabled && ui.stopBtn != nil {
		ui.stopBtn.Enable()
	} else if ui.stopBtn != nil && !update.stopEnabled {
		ui.stopBtn.Disable()
	}

	select {
	case ui.uiUpdateChan <- update:
	default:
		if ui.progressBar != nil && update.progress >= 0 {
			ui.progressBar.SetValue(update.progress)
		}
		if update.outputText != "" && ui.resultOutput != nil {
			ui.resultOutput.SetText(update.outputText)
		}
		if update.progressText != "" && ui.progressText != nil {
			ui.progressText.SetText(update.progressText)
		}
	}
}

func (ui *LoadTestUI) showErrorDialog(msg string) {
	ui.safeUpdateUI(uiUpdate{errMsg: msg})
}

func (ui *LoadTestUI) resetOnError(err error) {
	ui.isRunning = false
	if ui.startBtn != nil {
		ui.startBtn.Enable()
	}
	if ui.stopBtn != nil {
		ui.stopBtn.Disable()
	}
	if err != nil {
		ui.showErrorDialog(err.Error())
	}
	if ui.progressBar != nil {
		ui.safeUpdateUI(uiUpdate{progress: 0.0})
	}
	if ui.progressText != nil {
		ui.safeUpdateUI(uiUpdate{progressText: "Request Avg Duration 0.000s"})
	}
}

func (ui *LoadTestUI) setupStartButton() {
	ui.startBtn.OnTapped = func() {
		if ui.isRunning {
			ui.showErrorDialog("Тест уже запущен!")
			return
		}
		ui.isRunning = true

		// Очистка результатов перед началом нового теста
		if ui.resultOutput != nil {
			ui.resultOutput.SetText("")
		}
		if ui.progressBar != nil {
			ui.safeUpdateUI(uiUpdate{progress: 0.0})
		}
		if ui.progressText != nil {
			ui.safeUpdateUI(uiUpdate{progressText: "Request Avg Duration 0.000s"})
		}

		reqCountVal, err := parseInt(reqCount.Text)
		if err != nil {
			ui.resetOnError(fmt.Errorf("invalid request count: %v", err))
			return
		}
		if _, err := parseInt(duration.Text); err != nil {
			ui.resetOnError(fmt.Errorf("invalid duration: %v", err))
			return
		}

		if urlEntry.Text == "" {
			ui.resetOnError(fmt.Errorf("URL is required"))
			return
		}

		// Показать значок загрузки и начать прогресс
		ui.safeUpdateUI(uiUpdate{
			outputText:   "",
			startEnabled: false,
			stopEnabled:  true,
			status:       "loading",
			progress:     0.0,
			progressText: "Request Avg Duration 0.000s",
		})

		go func() {
			defer func() {
				ui.isRunning = false
				ui.safeUpdateUI(uiUpdate{
					startEnabled: true,
					stopEnabled:  false,
					status:       "completed", // Устанавливаем completed в конце
					progress:     100.0,
					progressText: "Request Avg Duration 0.000s",
				})
			}()

			totalDuration := 0.0
			var metrics []Metric

			for i := 0; i < reqCountVal; i++ {
				if !ui.isRunning {
					currentOutput := ""
					if ui.resultOutput != nil {
						currentOutput = ui.resultOutput.Text()
					}
					ui.safeUpdateUI(uiUpdate{
						outputText:   currentOutput + "\n🛑 Тест остановлен пользователем.",
						startEnabled: true,
						stopEnabled:  false,
						status:       "", // Сбрасываем статус при остановке
						progress:     0.0,
						progressText: "Request Avg Duration 0.000s",
					})
					return
				}

				metrics = []Metric{
					{Name: "DNS", Duration: 10.50 + rand.Float64()*5, Order: 1},
					{Name: "Connection", Duration: 20.75 + rand.Float64()*10, Order: 2},
					{Name: "TLS", Duration: 15.30 + rand.Float64()*5, Order: 3},
					{Name: "Request Write", Duration: 5.25 + rand.Float64()*2, Order: 4},
					{Name: "Server Processing", Duration: 50.80 + rand.Float64()*20, Order: 5},
					{Name: "Response Read", Duration: 10.20 + rand.Float64()*5, Order: 6},
				}
				for _, m := range metrics {
					totalDuration += m.Duration
				}
				metrics = append(metrics, Metric{Name: "Total", Duration: totalDuration, Order: 7})

				sort.Slice(metrics, func(i, j int) bool {
					return metrics[i].Order < metrics[j].Order
				})

				progress := float64(i+1) * 100.0 / float64(reqCountVal)
				avgDuration := (totalDuration / 1000.0) / float64(i+1)
				progressText := fmt.Sprintf("Request Avg Duration %.3fs", avgDuration)

				// Сохраняем status: "loading" для промежуточных обновлений
				ui.safeUpdateUI(uiUpdate{
					progress:     progress,
					progressText: progressText,
					startEnabled: false,
					stopEnabled:  true,
					status:       "loading", // Добавляем status
				})

				time.Sleep(time.Duration(50+rand.Intn(50)) * time.Millisecond)
			}

			resultOutput := "Results:\n"
			for _, metric := range metrics {
				resultOutput += fmt.Sprintf("%s: %.2fms\n", metric.Name, metric.Duration)
			}
			avgDuration := totalDuration / 1000.0 / float64(reqCountVal)
			finalProgressText := fmt.Sprintf("Request Avg Duration %.3fs", avgDuration)

			store.AddTestRun(store.TestRun{
				ID:        store.TestRunCount() + 1,
				Name:      fmt.Sprintf("Test Run %d", store.TestRunCount()+1),
				StartTime: time.Now().Format("2006-01-02 15:04"),
				Status:    "Completed",
			})

			ui.safeUpdateUI(uiUpdate{
				outputText:   resultOutput,
				progress:     100.0,
				progressText: finalProgressText,
				status:       "completed",
				startEnabled: true,
				stopEnabled:  false,
			})
		}()
	}
}

func (ui *LoadTestUI) setupStopButton() {
	ui.stopBtn.OnTapped = func() {
		if !ui.isRunning {
			return
		}

		ui.isRunning = false

		currentOutput := ""
		if ui.resultOutput != nil {
			currentOutput = ui.resultOutput.Text()
		}
		ui.safeUpdateUI(uiUpdate{
			outputText:   currentOutput + "\n🛑 Тест остановлен пользователем.",
			startEnabled: true,
			stopEnabled:  false,
			status:       "",
			progress:     0.0,
			progressText: "Request Avg Duration 0.000s",
		})
	}
}

func (ui *LoadTestUI) CreateButtons() *fyne.Container {
	debugCheck := widget.NewCheck("Debug Mode", nil)
	ui.setupStartButton()
	ui.setupStopButton()
	return container.NewHBox(ui.startBtn, ui.stopBtn, debugCheck)
}
