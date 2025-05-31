package report

import (
	"fmt"

	"httes/core/types"

	"fyne.io/fyne/v2/widget"
)

var AvailableOutputServices = make(map[string]func(*widget.TextGrid, *widget.ProgressBar, *widget.Label, int) ReportService)

type ReportService interface {
	DoneChan() <-chan struct{}
	Init(debug bool) error
	Start(input chan *types.ScenarioResult)
	Stop()
}

// NewReportService создаёт новый сервис отчётов с поддержкой прогресс-бара, текста и totalRequests
func NewReportService(s string, resultGrid *widget.TextGrid, progressBar *widget.ProgressBar, progressText *widget.Label, totalRequests int) (ReportService, error) {
	if constructor, ok := AvailableOutputServices[s]; ok {
		return constructor(resultGrid, progressBar, progressText, totalRequests), nil
	}
	return nil, fmt.Errorf("unsupported output type: %s", s)
}
