package report

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"httes/core/types"

	"fyne.io/fyne/v2/widget"
)

const OutputTypeGui = "gui"

func init() {
	AvailableOutputServices[OutputTypeGui] = func(resultGrid *widget.TextGrid, progressBar *widget.ProgressBar, progressText *widget.Label, totalRequests int) ReportService {
		return NewGuiReportService(resultGrid, progressBar, progressText, totalRequests)
	}
}

type guiReport struct {
	resultGrid    *widget.TextGrid
	progressBar   *widget.ProgressBar
	doneChan      chan struct{}
	updateChan    chan string // Канал для обновления UI
	result        *Result
	mu            sync.Mutex
	debug         bool
	closed        bool // Флаг для отслеживания состояния doneChan
	progressText  *widget.Label
	totalRequests int
}

type duration struct {
	name     string
	duration float32
	order    int
}

var keyToStr = map[string]duration{
	"dnsDuration":           {name: "DNS", order: 1},
	"connDuration":          {name: "Connection", order: 2},
	"tlsDuration":           {name: "TLS", order: 3},
	"reqDuration":           {name: "Request Write", order: 4},
	"serverProcessDuration": {name: "Server Processing", order: 5},
	"resDuration":           {name: "Response Read", order: 6},
	"duration":              {name: "Total", order: 7},
}

func NewGuiReportService(resultGrid *widget.TextGrid, progressBar *widget.ProgressBar, progressText *widget.Label, totalRequests int) ReportService {
	if resultGrid == nil {
		panic("resultGrid cannot be nil")
	}
	svc := &guiReport{
		resultGrid:  resultGrid,
		progressBar: progressBar,
		doneChan:    make(chan struct{}),
		updateChan:  make(chan string, 100),
		result: &Result{
			StepResults:    make(map[uint16]*ScenarioStepResultSummary),
			Durations:      make(map[string]float32),
			StatusCodeDist: make(map[int]int),
			ProgressPoints: make(map[int]float32),
		},
		progressText:  progressText,
		totalRequests: totalRequests,
	}
	svc.progressBar.Max = 100.0
	go svc.runUIUpdater()
	return svc
}

// runUIUpdater обрабатывает обновления UI из канала
func (r *guiReport) runUIUpdater() {
	for data := range r.updateChan {
		if r.resultGrid != nil {
			r.resultGrid.SetText(data)
			r.resultGrid.Refresh()
		}
	}
}

func (r *guiReport) Init(debug bool) error {
	r.debug = debug
	return nil
}

func (r *guiReport) Start(input chan *types.ScenarioResult) {
	if input == nil {
		return
	}

	if r.debug {
		r.printInDebugMode(input)
		return
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	done := make(chan struct{})

	go func() {
		count := 0
		for scr := range input {
			r.mu.Lock()
			aggregate(r.result, scr)
			r.updateProgressBar()
			r.mu.Unlock()
			count++
		}
		close(done)
	}()

	for {
		select {
		case <-ticker.C:
			r.mu.Lock()
			if r.totalRequests > 0 && (r.result.SuccessCount+r.result.FailedCount) > 0 {
				r.updateProgressBar()
			} else {
				r.progressBar.SetValue(0)
				r.progressText.SetText("Request Avg Duration 0.000s")
			}
			r.mu.Unlock()
		case <-done:
			r.mu.Lock()
			r.printDetails()
			r.mu.Unlock()
			select {
			case <-r.doneChan:
			default:
				r.doneChan <- struct{}{}
			}
			r.resetProgressBar()
			return
		}
	}
}

func (r *guiReport) DoneChan() <-chan struct{} {
	return r.doneChan
}

func (r *guiReport) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.closed {
		r.closed = true
		r.resetProgressBar()
	}
}

// resetProgressBar сбрасывает прогресс-бар и текст
func (r *guiReport) resetProgressBar() {
	if r.progressBar != nil {
		r.progressBar.SetValue(0)
	}
	if r.progressText != nil {
		r.progressText.SetText("Request Avg Duration 0.000s")
	}
}

// updateProgressBar обновляет прогресс-бар и текст прогресса с avgDuration
func (r *guiReport) updateProgressBar() {
	if r.resultGrid == nil || r.result == nil || r.progressBar == nil || r.progressText == nil || r.totalRequests <= 0 {
		r.progressBar.SetValue(0)
		r.progressText.SetText("Request Avg Duration 0.000s")
		return
	}

	totalProcessed := float32(r.result.SuccessCount + r.result.FailedCount)
	if totalProcessed == 0 {
		r.progressBar.SetValue(0)
		r.progressText.SetText("Request Avg Duration 0.000s")
		return
	}

	percent := (totalProcessed / float32(r.totalRequests)) * 100
	if percent > 100 {
		percent = 100
	}
	r.progressBar.SetValue(float64(percent))
	r.progressText.SetText(fmt.Sprintf("Request Avg Duration %.3fs", r.getAvgDuration()))
}

// getAvgDuration возвращает среднюю длительность для текущего прогресса
func (r *guiReport) getAvgDuration() float32 {
	if r.result.SuccessCount == 0 {
		return 0.0
	}
	progressKeys := make([]int, 0, len(r.result.ProgressPoints))
	for k := range r.result.ProgressPoints {
		progressKeys = append(progressKeys, k)
	}
	if len(progressKeys) == 0 {
		return 0.0
	}
	sort.Ints(progressKeys)
	lastMilestone := progressKeys[len(progressKeys)-1]
	if lastMilestone > r.result.SuccessCount {
		lastMilestone = r.result.SuccessCount
	}
	return r.result.ProgressPoints[lastMilestone]
}

// // В guiReport.go
// func (r *guiReport) GetProgressData() (float64, float64) {
// 	totalProcessed := float32(r.result.SuccessCount + r.result.FailedCount)
// 	percent := float32(0)
// 	if r.totalRequests > 0 {
// 		percent = (totalProcessed / float32(r.totalRequests)) * 100
// 		if percent > 100 {
// 			percent = 100
// 		}
// 	}
// 	fmt.Println("Progress Data:", totalProcessed, percent) // Отладка
// 	return float64(totalProcessed), float64(percent)
// }

// func (r *guiReport) GetDurationData() ([]string, []float32) {
// 	var durationList []duration
// 	for d, s := range r.result.Durations {
// 		dur, ok := keyToStr[d]
// 		if !ok {
// 			dur = duration{name: d, order: 999}
// 		}
// 		dur.duration = s
// 		durationList = append(durationList, dur)
// 	}
// 	for _, dur := range keyToStr {
// 		found := false
// 		for _, d := range durationList {
// 			if d.name == dur.name {
// 				found = true
// 				break
// 			}
// 		}
// 		if !found {
// 			durationList = append(durationList, duration{name: dur.name, duration: 0, order: dur.order})
// 		}
// 	}
// 	sort.Slice(durationList, func(i, j int) bool {
// 		return durationList[i].order < durationList[j].order
// 	})

// 	x := make([]string, len(durationList))
// 	y := make([]float32, len(durationList))
// 	for i, v := range durationList {
// 		x[i] = v.name
// 		y[i] = v.duration
// 	}
// 	fmt.Println("Duration Data:", x, y) // Отладка
// 	return x, y
// }

// printDetails выводит итоговый результат
func (r *guiReport) printDetails() {
	if r.resultGrid == nil || r.result == nil {
		return
	}

	currentOutput := r.resultGrid.Text()

	bGui := strings.Builder{}
	bGui.WriteString(currentOutput + "\n\nRESULT\n")
	bGui.WriteString("-------------------------------------\n")
	bGui.WriteString(fmt.Sprintf("Success Count:    %-6d (%d%%)\n", r.result.SuccessCount, r.result.successPercentage()))
	bGui.WriteString(fmt.Sprintf("Failed Count:     %-6d (%d%%)\n", r.result.FailedCount, r.result.failedPercentage()))

	bGui.WriteString("\nDurations (Avg):\n")
	var durationList = make([]duration, 0)
	for d, s := range r.result.Durations {
		dur, ok := keyToStr[d]
		if !ok {
			dur = duration{name: d, order: 999}
		}
		dur.duration = s
		durationList = append(durationList, dur)
	}
	for _, dur := range keyToStr {
		found := false
		for _, d := range durationList {
			if d.name == dur.name {
				found = true
				break
			}
		}
		if !found {
			durationList = append(durationList, duration{name: dur.name, duration: 0, order: dur.order})
		}
	}
	sort.Slice(durationList, func(i, j int) bool {
		return durationList[i].order < durationList[j].order
	})
	for _, v := range durationList {
		bGui.WriteString(fmt.Sprintf("  %-20s:%.4fs\n", v.name, v.duration))
	}

	if len(r.result.StatusCodeDist) > 0 {
		bGui.WriteString("\nStatus Code (Message) :Count\n")
		keys := make([]int, 0, len(r.result.StatusCodeDist))
		for k := range r.result.StatusCodeDist {
			keys = append(keys, k)
		}
		sort.Ints(keys)
		for _, s := range keys {
			c := r.result.StatusCodeDist[s]
			bGui.WriteString(fmt.Sprintf("  %-20s:%d\n", fmt.Sprintf("%d (%s)", s, http.StatusText(s)), c))
		}
	}

	avgParamCount := float32(0)
	if r.result.TotalRequests > 0 {
		avgParamCount = float32(r.result.TotalParamCount) / float32(r.result.TotalRequests)
	}
	bGui.WriteString(fmt.Sprintf("\nAvg. Parameter Count: %.2f\n", avgParamCount))

	if r.debug {
		fmt.Println("Updating TextGrid (details) with content:", bGui.String())
	}
	r.updateChan <- bGui.String()
}

func (r *guiReport) printInDebugMode(input chan *types.ScenarioResult) {
	if input == nil {
		fmt.Println("printInDebugMode: ERROR: input channel is nil")
		return
	}

	b := strings.Builder{}
	b.WriteString("🐞 Debug Mode\n")
	b.WriteString("----------------------------------------------------\n")

	count := 0
	for scr := range input {
		fmt.Printf("Debug: Processing ScenarioResult #%d: %+v\n", count, scr)
		r.mu.Lock()
		aggregate(r.result, scr)

		for _, sr := range scr.StepResults {
			fmt.Printf("Debug: StepResult: StepID=%d, StatusCode=%d, Duration=%v, Err=%+v\n",
				sr.StepID, sr.StatusCode, sr.Duration, sr.Err)
			verboseInfo := ScenarioStepResultToVerboseHttpRequestInfo(sr)
			b.WriteString(fmt.Sprintf("\n\nSTEP (%d) %s\n", verboseInfo.StepId, verboseInfo.StepName))
			b.WriteString("------------------------------------\n")
			b.WriteString("- Environment Variables\n")
			for eKey, eVal := range verboseInfo.Envs {
				switch eVal.(type) {
				case map[string]interface{}, []string, []float64, []bool:
					valPretty, _ := json.Marshal(eVal)
					b.WriteString(fmt.Sprintf("  %s: %s\n", eKey, valPretty))
				default:
					b.WriteString(fmt.Sprintf("  %s: %v\n", eKey, eVal))
				}
			}

			if verboseInfo.Error != "" && isVerboseInfoRequestEmpty(verboseInfo.Request) {
				b.WriteString(fmt.Sprintf("\n⚠️ Error: %s\n", verboseInfo.Error))
				continue
			}

			b.WriteString("\n- Request\n")
			b.WriteString(fmt.Sprintf("  Target: %s\n", verboseInfo.Request.Url))
			b.WriteString(fmt.Sprintf("  Method: %s\n", verboseInfo.Request.Method))
			b.WriteString("  Headers:\n")
			for hKey, hVal := range verboseInfo.Request.Headers {
				b.WriteString(fmt.Sprintf("    %s: %s\n", hKey, hVal))
			}

			contentType := sr.DebugInfo["requestHeaders"].(http.Header).Get("content-type")
			b.WriteString("  Body: ")
			if verboseInfo.Request.Body == nil {
				b.WriteString("null\n")
			} else if strings.Contains(contentType, "application/json") {
				valPretty, _ := json.MarshalIndent(verboseInfo.Request.Body, "    ", "  ")
				b.WriteString(fmt.Sprintf("\n    %s\n", valPretty))
			} else {
				b.WriteString(fmt.Sprintf("%v\n", verboseInfo.Request.Body))
			}

			if verboseInfo.Error != "" {
				if len(verboseInfo.FailedCaptures) > 0 {
					b.WriteString("\n- Failed Captures\n")
					for wKey, wVal := range verboseInfo.FailedCaptures {
						b.WriteString(fmt.Sprintf("    %s: %s\n", wKey, wVal))
					}
				}
				b.WriteString(fmt.Sprintf("\n⚠️ Error: %s\n", verboseInfo.Error))
			} else {
				b.WriteString("\n- Response\n")
				b.WriteString(fmt.Sprintf("  StatusCode: %d\n", verboseInfo.Response.StatusCode))
				b.WriteString("  Headers:\n")
				for hKey, hVal := range verboseInfo.Response.Headers {
					b.WriteString(fmt.Sprintf("    %s: %s\n", hKey, hVal))
				}

				contentType = sr.DebugInfo["responseHeaders"].(http.Header).Get("content-type")
				b.WriteString("  Body: ")
				if verboseInfo.Response.Body == nil {
					b.WriteString("null\n")
				} else if strings.Contains(contentType, "application/json") {
					valPretty, _ := json.MarshalIndent(verboseInfo.Response.Body, "    ", "  ")
					b.WriteString(fmt.Sprintf("\n    %s\n", valPretty))
				} else {
					b.WriteString(fmt.Sprintf("%v\n", verboseInfo.Response.Body))
				}

				if len(verboseInfo.FailedCaptures) > 0 {
					b.WriteString("\n- Failed Captures\n")
					for wKey, wVal := range verboseInfo.FailedCaptures {
						b.WriteString(fmt.Sprintf("    %s: %s\n", wKey, wVal))
					}
				}
			}
		}

		if r.debug {
			fmt.Println("Debug: Updating TextGrid with content:", b.String())
		}
		r.updateChan <- b.String()
		r.mu.Unlock()
		count++
	}

	fmt.Println("Debug: Finished processing input channel")
	r.mu.Lock()
	r.printDetails()
	r.mu.Unlock()
	select {
	case <-r.doneChan:
		fmt.Println("doneChan already closed, skipping signal")
	default:
		fmt.Println("Debug: Sending done signal")
		r.doneChan <- struct{}{}
	}
	r.resetProgressBar()
}
