package report

import (
	"math"
	"sync"
	"time"

	"httes/core/types"
)

// Процентные точки прогресса
var progressPercentages = []float32{0.10, 0.25, 0.40, 0.50, 0.75, 0.90, 1.00}

func aggregate(r *Result, scr *types.ScenarioResult) {
	r.mu.Lock()
	defer r.mu.Unlock()

	isSuccess := true
	totalDuration := float32(0)

	// Инициализация полей, если ещё не созданы
	if r.Durations == nil {
		r.Durations = make(map[string]float32)
	}
	if r.StatusCodeDist == nil {
		r.StatusCodeDist = make(map[int]int)
	}
	if r.ProgressPoints == nil {
		r.ProgressPoints = make(map[int]float32)
	}

	for _, sr := range scr.StepResults {
		// Подсчёт параметров (ключей в Custom)
		paramCount := len(sr.Custom)
		r.TotalParamCount += paramCount
		r.TotalRequests++

		// Инициализация StepResults, если ещё не создана
		if _, ok := r.StepResults[sr.StepID]; !ok {
			r.StepResults[sr.StepID] = &ScenarioStepResultSummary{
				Name:           sr.StepName,
				Durations:      make(map[string]float32),
				StatusCodeDist: make(map[int]int),
				ErrorDist:      map[string]int{},
			}
		}

		// Обновление статистики шага
		stepResult := r.StepResults[sr.StepID]
		if sr.Err.Type != "" {
			isSuccess = false
			stepResult.FailedCount++
			stepResult.ErrorDist[sr.Err.Reason]++
		} else {
			stepResult.SuccessCount++
		}
		stepResult.StatusCodeDist[sr.StatusCode]++

		// Обновление длительностей для шага
		for k, v := range sr.Custom {
			var dur float32
			switch val := v.(type) {
			case float32:
				dur = val
			case time.Duration:
				dur = float32(val) / float32(time.Second)
			default:
				continue
			}
			if _, exists := stepResult.Durations[k]; !exists {
				stepResult.Durations[k] = 0
			}
			count := stepResult.SuccessCount + stepResult.FailedCount
			stepResult.Durations[k] = (stepResult.Durations[k]*float32(count-1) + dur) / float32(count)
		}

		// Записываем общую длительность шага (duration)
		totalStepDuration := float32(sr.Duration) / float32(time.Second)
		if _, exists := stepResult.Durations["duration"]; !exists {
			stepResult.Durations["duration"] = 0
		}
		count := stepResult.SuccessCount + stepResult.FailedCount
		stepResult.Durations["duration"] = (stepResult.Durations["duration"]*float32(count-1) + totalStepDuration) / float32(count)

		// Обновление общих длительностей
		for k, v := range sr.Custom {
			var dur float32
			switch val := v.(type) {
			case float32:
				dur = val
			case time.Duration:
				dur = float32(val) / float32(time.Second)
			default:
				continue
			}
			if _, exists := r.Durations[k]; !exists {
				r.Durations[k] = 0
			}
			count := r.SuccessCount + r.FailedCount + 1 // Текущий запрос
			r.Durations[k] = (r.Durations[k]*float32(count-1) + dur) / float32(count)
		}

		// Записываем общую длительность (duration)
		if _, exists := r.Durations["duration"]; !exists {
			r.Durations["duration"] = 0
		}
		count = int64(r.SuccessCount) + int64(r.FailedCount) + 1
		r.Durations["duration"] = (r.Durations["duration"]*float32(count-1) + totalStepDuration) / float32(count)

		r.StatusCodeDist[sr.StatusCode]++
		totalDuration += totalStepDuration
	}

	// Обновление общей статистики
	if isSuccess {
		r.SuccessCount++

		// Динамическая генерация точек прогресса
		maxSuccessCount := r.SuccessCount
		for _, percent := range progressPercentages {
			milestone := int(math.Ceil(float64(maxSuccessCount) * float64(percent)))
			if r.SuccessCount == milestone {
				r.ProgressPoints[milestone] = totalDuration
			}
		}
	} else {
		r.FailedCount++
	}

	// Обновление общей средней длительности
	if r.SuccessCount+r.FailedCount > 0 {
		r.AvgDuration = (r.AvgDuration*float32(r.SuccessCount+r.FailedCount-1) + totalDuration) / float32(r.SuccessCount+r.FailedCount)
	}
}

type Result struct {
	SuccessCount    int
	FailedCount     int
	AvgDuration     float32
	StepResults     map[uint16]*ScenarioStepResultSummary
	TotalParamCount int                // Общее количество ключей в Custom
	TotalRequests   int                // Общее количество запросов
	ProgressPoints  map[int]float32    // Средняя длительность на точках прогресса (ключ: SuccessCount, значение: AvgDuration)
	Durations       map[string]float32 // Средние длительности по всем шагам
	StatusCodeDist  map[int]int        // Распределение статус-кодов по всем шагам
	mu              sync.Mutex
}

func (r *Result) successPercentage() int {
	if r.SuccessCount+r.FailedCount == 0 {
		return 0
	}
	t := float32(r.SuccessCount) / float32(r.SuccessCount+r.FailedCount)
	return int(t * 100)
}

func (r *Result) failedPercentage() int {
	if r.SuccessCount+r.FailedCount == 0 {
		return 0
	}
	return 100 - r.successPercentage()
}

type ScenarioStepResultSummary struct {
	Name           string             `json:"name"`
	StatusCodeDist map[int]int        `json:"status_code_dist"`
	ErrorDist      map[string]int     `json:"error_dist"`
	Durations      map[string]float32 `json:"durations"`
	SuccessCount   int64              `json:"success_count"`
	FailedCount    int64              `json:"fail_count"`
}

// func (s *ScenarioStepResultSummary) successPercentage() int {
// 	if s.SuccessCount+s.FailedCount == 0 {
// 		return 0
// 	}
// 	t := float32(s.SuccessCount) / float32(s.SuccessCount+s.FailedCount)
// 	return int(t * 100)
// }

// func (s *ScenarioStepResultSummary) failedPercentage() int {
// 	if s.SuccessCount+s.FailedCount == 0 {
// 		return 0
// 	}
// 	return 100 - s.successPercentage()
// }
