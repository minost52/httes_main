package loadtest

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"sync"
	"time"

	"httes/core"
	"httes/core/report"
	"httes/core/types"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// LoadTestUI управляет состоянием UI и тестирования.
type LoadTestUI struct {
	isRunning     bool
	testCtx       context.Context
	testCancel    context.CancelFunc
	reportService report.ReportService
	uiUpdateChan  chan uiUpdate
	startBtn      *widget.Button
	stopBtn       *widget.Button
	app           fyne.App
	resultOutput  *widget.TextGrid
	progressBar   *widget.ProgressBar
	progressText  *widget.Label
}

// uiUpdate представляет обновление для UI.
type uiUpdate struct {
	errMsg       string
	outputText   string
	startEnabled bool
	stopEnabled  bool
}

var uiUpdaterOnce sync.Once

// NewLoadTestUI создаёт новый экземпляр LoadTestUI.
func NewLoadTestUI(app fyne.App, resultOutput *widget.TextGrid, progressBar *widget.ProgressBar, progressText *widget.Label) *LoadTestUI {
	ui := &LoadTestUI{
		app:          app,
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

// initUIUpdater инициализирует канал и горутину для обновления UI.
func (ui *LoadTestUI) initUIUpdater() {
	uiUpdaterOnce.Do(func() {
		ui.uiUpdateChan = make(chan uiUpdate, 100)
		go func() {
			defer func() {
				log.Println("UI update goroutine exiting, closing channel")
				close(ui.uiUpdateChan)
				if ui.startBtn != nil {
					ui.startBtn.Enable()
				}
				if ui.stopBtn != nil {
					ui.stopBtn.Disable()
				}
			}()
			for update := range ui.uiUpdateChan {
				log.Printf("UI Update received: %+v", update)
				if update.errMsg != "" && ui.app != nil {
					dialog := widget.NewLabel(update.errMsg)
					w := ui.app.NewWindow("Error")
					w.SetContent(container.NewVBox(dialog))
					w.Show()
				}
				if update.outputText != "" && ui.resultOutput != nil {
					ui.resultOutput.SetText(update.outputText)
					ui.resultOutput.Refresh()
				}
				if update.startEnabled && ui.startBtn != nil {
					ui.startBtn.Enable()
					log.Println("Enabling Start button")
				} else if ui.startBtn != nil {
					ui.startBtn.Disable()
					log.Println("Disabling Start button")
				}
				if update.stopEnabled && ui.stopBtn != nil {
					ui.stopBtn.Enable()
					log.Println("Enabling Stop button")
				} else if ui.stopBtn != nil {
					ui.stopBtn.Disable()
					log.Println("Disabling Stop button")
				}
			}
			log.Println("UI update channel closed, goroutine exited")
		}()
	})
}

// safeUpdateUI безопасно отправляет обновление UI.
func (ui *LoadTestUI) safeUpdateUI(update uiUpdate) {
	select {
	case ui.uiUpdateChan <- update:
		log.Println("Sent UI update")
	default:
		log.Println("UI update channel is full or closed")
	}
}

// showErrorDialog отображает диалоговое окно с ошибкой.
func (ui *LoadTestUI) showErrorDialog(msg string) {
	log.Println("Error:", msg)
	ui.safeUpdateUI(uiUpdate{errMsg: msg})
}

// resetOnError сбрасывает состояние при ошибке.
func (ui *LoadTestUI) resetOnError(err error) {
	ui.isRunning = false
	ui.testCtx = nil
	ui.testCancel = nil
	ui.reportService = nil
	if ui.startBtn != nil {
		ui.startBtn.Enable()
	}
	if ui.stopBtn != nil {
		ui.stopBtn.Disable()
	}
	if err != nil {
		ui.showErrorDialog(err.Error())
	}
}

func (ui *LoadTestUI) runLoadTest(ctx context.Context, cancel context.CancelFunc, h types.Heart, rs report.ReportService) error {
	fmt.Println("RunLoadTest started")
	e, err := core.NewEngine(ctx, h, rs)
	if err != nil {
		return err
	}
	if err = e.Init(); err != nil {
		return err
	}

	resultChan := e.GetResultChan()
	if resultChan == nil {
		return fmt.Errorf("resultChan is nil")
	}

	// Start report service first
	reportDone := make(chan struct{})
	go func() {
		defer close(reportDone)
		rs.Start(resultChan)
	}()

	// Start engine
	engineDone := make(chan struct{})
	go func() {
		defer close(engineDone)
		e.Start()
	}()

	// Wait for completion
	select {
	case <-engineDone:
		fmt.Println("Engine completed normally")
	case <-ctx.Done():
		fmt.Println("Context cancelled, waiting for engine...")
		<-engineDone // Wait for engine to shutdown
	case <-time.After(time.Duration(h.TestDuration+10) * time.Second):
		fmt.Println("Test timeout, cancelling...")
		cancel()
		<-engineDone // Wait for engine to shutdown
	}

	// Wait for report service to finish
	select {
	case <-reportDone:
		fmt.Println("ReportService completed normally")
	case <-time.After(5 * time.Second):
		fmt.Println("ReportService shutdown timeout")
	}

	// Update UI
	ui.safeUpdateUI(uiUpdate{
		outputText:   ui.resultOutput.Text() + "\n✅ Тест завершен!",
		startEnabled: true,
		stopEnabled:  false,
	})

	return nil
}

// setupStartButton настраивает обработчик для кнопки "Start".
func (ui *LoadTestUI) setupStartButton(verboseCheck *widget.Check) {
	ui.startBtn.OnTapped = func() {
		if ui.isRunning {
			ui.showErrorDialog("Тест уже запущен!")
			return
		}
		ui.isRunning = true
		ui.testCtx, ui.testCancel = context.WithCancel(context.Background())

		reqCount, err := parseInt(reqCount.Text)
		if err != nil {
			ui.resetOnError(fmt.Errorf("invalid request count: %v", err))
			return
		}
		duration, err := parseInt(duration.Text)
		if err != nil {
			ui.resetOnError(fmt.Errorf("invalid duration: %v", err))
			return
		}

		step, err := createScenarioStep(
			protocolSelect.Selected,
			urlEntry.Text,
			methodSelect.Selected,
			usernameEntry.Text,
			passwordEntry.Text,
			certPathEntry.Text,
			certKeyPathEntry.Text,
		)
		if err != nil {
			ui.resetOnError(err)
			return
		}
		scenario := types.Scenario{Steps: []types.ScenarioStep{step}}

		var proxyAddr *url.URL
		if proxyEntry.Text != "" {
			proxyAddr, err = url.Parse(proxyEntry.Text)
			if err != nil {
				ui.resetOnError(fmt.Errorf("invalid proxy URL: %v", err))
				return
			}
		}

		if ui.resultOutput == nil {
			ui.resetOnError(fmt.Errorf("resultOutput not initialized"))
			return
		}

		h := createHeart(scenario, proxyAddr, reqCount, duration, loadType.Selected, "gui", verboseCheck.Checked)
		ui.safeUpdateUI(uiUpdate{
			outputText:   "🚀 Тест запущен...",
			startEnabled: false,
			stopEnabled:  true,
		})

		// Запуск теста в фоне
		go func() {
			defer func() {
				ui.isRunning = false
				ui.testCtx = nil
				ui.testCancel = nil
				ui.reportService = nil
				ui.safeUpdateUI(uiUpdate{
					startEnabled: true,
					stopEnabled:  false,
				})
				log.Println("Test finished, ensuring Start button is enabled via defer")
			}()

			var err error
			ui.reportService, err = report.NewReportService(h.ReportDestination, ui.resultOutput, ui.progressBar, ui.progressText, reqCount)
			if err != nil {
				ui.resetOnError(fmt.Errorf("failed to create report service: %v", err))
				return
			}

			err = ui.runLoadTest(ui.testCtx, ui.testCancel, h, ui.reportService)
			if err != nil {
				log.Println("RunLoadTest failed:", err)
			}
		}()
	}
}

// setupStopButton настраивает обработчик для кнопки "Stop".
func (ui *LoadTestUI) setupStopButton() {
	ui.stopBtn.OnTapped = func() {
		if ui.testCancel != nil {
			log.Println("Cancelling test...")
			ui.testCancel()
		}
		if ui.reportService != nil {
			log.Println("Stopping reportService...")
			ui.reportService.Stop()
		}

		currentOutput := ""
		if ui.resultOutput != nil {
			currentOutput = ui.resultOutput.Text()
		}
		ui.safeUpdateUI(uiUpdate{
			outputText:   currentOutput + "\n🛑 Тест остановлен пользователем.",
			startEnabled: true,
			stopEnabled:  false,
		})

		ui.isRunning = false
		ui.testCtx = nil
		ui.testCancel = nil
		ui.reportService = nil
	}
}

// CreateButtons создаёт контейнер с кнопками и чекбоксом.
func (ui *LoadTestUI) CreateButtons() *fyne.Container {
	debugCheck := widget.NewCheck("Debug Mode", nil)
	debugCheck.Checked = false

	ui.setupStartButton(debugCheck)
	ui.setupStopButton()

	return container.NewHBox(ui.startBtn, ui.stopBtn, debugCheck)
}

// parseInt парсит строку в целое число.
func parseInt(s string) (int, error) {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid number %s: %v", s, err)
	}
	return i, nil
}
