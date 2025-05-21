package core

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"sync"
	"time"

	"httes/core/proxy"
	"httes/core/report"
	"httes/core/scenario"
	"httes/core/types"
)

const (
	// интервал в миллисекундах
	tickerInterval = 100
)

type engine struct {
	heart types.Heart // настройки теста

	proxyService    proxy.ProxyService        // сервис для работы с прокси
	scenarioService *scenario.ScenarioService // сервис для выполнения сценариев
	reportService   report.ReportService      // сервис для генерации отчетов

	tickCounter int            // счетчик тиков
	reqCountArr []int          // массив количества запросов на каждый тик
	wg          sync.WaitGroup // группа ожидания для синхронизации горутин

	resultChan chan *types.ScenarioResult // канал для передачи результатов сценариев

	ctx context.Context // контекст для управления жизненным циклом движка
}

// NewEngine - конструктор для создания нового движка.
func NewEngine(ctx context.Context, h types.Heart, rs report.ReportService) (e *engine, err error) {
	// Валидация настроек Heart
	err = h.Validate()
	if err != nil {
		fmt.Println("Heart validation failed:", err)
		return
	}

	// Инициализация сервиса прокси
	ps, err := proxy.NewProxyService(h.Proxy.Strategy)
	if err != nil {
		fmt.Println("NewProxyService failed:", err)
		return
	}

	// Инициализация сервиса сценариев
	ss := scenario.NewScenarioService()

	// Создание экземпляра движка
	e = &engine{
		heart:           h,
		ctx:             ctx,
		proxyService:    ps,
		scenarioService: ss,
		reportService:   rs,
	}

	return
}

// Init - инициализация движка и его сервисов.
func (e *engine) Init() (err error) {
	// Инициализация сервиса прокси
	if err = e.proxyService.Init(e.heart.Proxy); err != nil {
		fmt.Println("ProxyService Init failed:", err)
		return
	}

	// Инициализация сервиса сценариев
	if err = e.scenarioService.Init(e.ctx, e.heart.Scenario, e.proxyService.GetAll(), e.heart.Debug); err != nil {
		fmt.Println("ScenarioService Init failed:", err)
		return
	}

	// Инициализация сервиса отчетов
	if err = e.reportService.Init(e.heart.Debug); err != nil {
		fmt.Println("ReportService Init failed:", err)
		return
	}

	// Инициализация канала результатов
	e.resultChan = make(chan *types.ScenarioResult, e.heart.IterationCount*2) // Увеличим буфер

	// Инициализация массива количества запросов
	e.initReqCountArr()
	return
}

// Start - запуск движка.
func (e *engine) Start() {
	ticker := time.NewTicker(time.Duration(tickerInterval) * time.Millisecond)
	timeout := time.After(time.Duration(e.heart.TestDuration) * time.Second)

	defer func() {
		ticker.Stop()
		e.stop()
	}()

	e.tickCounter = 0
	e.wg = sync.WaitGroup{}
	var mutex = &sync.Mutex{}
	for {
		select {
		case <-e.ctx.Done():
			fmt.Println("Context cancelled, stopping engine")
			return
		case <-timeout:
			fmt.Println("Timeout reached, stopping engine")
			return
		case <-ticker.C:
			if e.tickCounter >= len(e.reqCountArr) {
				fmt.Println("All ticks completed, stopping engine")
				return
			}
			mutex.Lock()
			reqCount := e.reqCountArr[e.tickCounter]
			if reqCount > 0 {
				e.wg.Add(reqCount)
				go e.runWorkers(e.tickCounter)
			}
			e.tickCounter++
			mutex.Unlock()
		}
	}
}

// runWorkers запускает воркеров для выполнения запросов на текущем тике.
func (e *engine) runWorkers(c int) {
	for i := 1; i <= e.reqCountArr[c]; i++ {
		scenarioStartTime := time.Now()
		go func(t time.Time, workerID int) {
			defer e.wg.Done()
			e.runWorker(t)
		}(scenarioStartTime, i)
	}
}

// runWorker выполняет один запрос сценария с обработкой ошибок и прокси.
func (e *engine) runWorker(scenarioStartTime time.Time) {
	select {
	case <-e.ctx.Done():
		fmt.Println("Worker stopped due to context cancellation")
		return
	default:
	}

	var res *types.ScenarioResult
	var err *types.RequestError

	p := e.proxyService.GetProxy()
	retryCount := 3
	for i := 1; i <= retryCount; i++ {
		select {
		case <-e.ctx.Done():
			return
		default:
		}
		res, err = e.scenarioService.Do(p, scenarioStartTime)
		if err != nil {
			fmt.Println("scenarioService.Do returned error:", err)
			if err.Type == types.ErrorProxy {
				fmt.Println("Proxy error, retrying:", err.Reason)
				p = e.proxyService.ReportProxy(p, err.Reason)
				continue
			}
			if err.Type == types.ErrorIntented {
				fmt.Println("Intended error, stopping worker:", err)
				return
			}
			break
		}
		break
	}

	if err != nil {
		fmt.Println("Worker failed:", err)
		return
	}

	res.Others = make(map[string]interface{})
	res.Others["heartOthers"] = e.heart.Others
	res.Others["proxyCountry"] = e.proxyService.GetProxyCountry(p)

	// отправка результата в канал
	select {
	case e.resultChan <- res:
	case <-e.ctx.Done():
		fmt.Println("Result not sent, context cancelled")
	}
}

// stop завершает работу движка, ожидая завершения всех горутин и закрывая ресурсы.
func (e *engine) stop() {
	wgTimeout := time.After(5 * time.Second)
	done := make(chan struct{})
	go func() {
		e.wg.Wait()
		close(e.resultChan)
		close(done)
	}()
	select {
	case <-wgTimeout:
		fmt.Println("Warning: Some workers timed out")
	case <-done:
	}
	fmt.Println("Waiting for reportService to finish...")
	select {
	case <-e.reportService.DoneChan():
	case <-time.After(5 * time.Second):
		fmt.Println("Warning: ReportService timeout")
	}
	fmt.Println("Cleaning up...")
	e.proxyService.Done()
	e.scenarioService.Done()
	fmt.Println("Engine stopped, test completed")
}

// GetResultChan возвращает канал результатов
func (e *engine) GetResultChan() chan *types.ScenarioResult {
	if e.resultChan == nil {
		fmt.Println("GetResultChan: WARNING: resultChan is nil")
	}
	return e.resultChan
}

// initReqCountArr инициализирует массив количества запросов
func (e *engine) initReqCountArr() {
	if e.heart.Debug {
		e.reqCountArr = make([]int, e.heart.IterationCount)
		for i := range e.reqCountArr {
			e.reqCountArr[i] = 1
		}
		return
	}
	length := int(e.heart.TestDuration * int(time.Second/(tickerInterval*time.Millisecond)))
	e.reqCountArr = make([]int, length)

	if e.heart.TimeRunCountMap != nil {
		e.createManualReqCountArr()
	} else {
		switch e.heart.LoadType {
		case types.LoadTypeLinear:
			e.createLinearReqCountArr()
		case types.LoadTypeIncremental:
			e.createIncrementalReqCountArr()
		case types.LoadTypeWaved:
			e.createWavedReqCountArr()
		}
	}
}

// createLinearReqCountArr распределяет запросы равномерно
// func (e *engine) createLinearReqCountArr() {
// 	totalRequests := e.heart.IterationCount
// 	length := len(e.reqCountArr)
// 	if length == 0 {
// 		e.reqCountArr = []int{totalRequests}
// 		return
// 	}
// 	base := totalRequests / length
// 	remainder := totalRequests % length
// 	for i := 0; i < length; i++ {
// 		e.reqCountArr[i] = base
// 		if i < remainder {
// 			e.reqCountArr[i]++
// 		}
// 	}
// }

// createManualReqCountArr создает массив запросов на основе пользовательских настроек.
func (e *engine) createManualReqCountArr() {
	tickPerSecond := int(time.Second / (tickerInterval * time.Millisecond))
	stepStartIndex := 0
	for _, t := range e.heart.TimeRunCountMap {
		// Создание линейного распределения для текущего шага
		steps := make([]int, t.Duration)
		createLinearDistArr(t.Count, steps)

		for i := range steps {
			tickArrStartIndex := (i * tickPerSecond) + stepStartIndex
			tickArrEndIndex := tickArrStartIndex + tickPerSecond
			segment := e.reqCountArr[tickArrStartIndex:tickArrEndIndex]
			createLinearDistArr(steps[i], segment)
		}
		stepStartIndex += len(steps) * tickPerSecond
	}
}

// createLinearReqCountArr создает массив запросов с линейным увеличением нагрузки.
// Нагрузка равномерно распределена по всем тикам.
// Если IterationCount = 100, TestDuration = 10, то в секунду — 10 запросов, и в каждом
// тике будет одинаковое распределение (например, 1 запрос в каждом из 10 тиков).
func (e *engine) createLinearReqCountArr() {
	steps := make([]int, e.heart.TestDuration)
	createLinearDistArr(e.heart.IterationCount, steps)
	tickPerSecond := int(time.Second / (tickerInterval * time.Millisecond))
	for i := range steps {
		tickArrStartIndex := i * tickPerSecond
		tickArrEndIndex := tickArrStartIndex + tickPerSecond
		segment := e.reqCountArr[tickArrStartIndex:tickArrEndIndex]
		createLinearDistArr(steps[i], segment)
	}
}

// createIncrementalReqCountArr создает массив запросов с инкрементальной нагрузкой.
// Нагрузка равномерно нарастает от начала к концу.
// Всего 100 запросов на 10 секунд.
// Распределение запросов по секундам: примерно
// 2, 4, 5, 7, 9, 11, 13, 15, 16, 18
// Внутри каждой секунды: запросы равномерно распределены по тикам
// (например, 1 запрос на каждый тик, где это возможно).
func (e *engine) createIncrementalReqCountArr() {
	steps := createIncrementalDistArr(e.heart.IterationCount, e.heart.TestDuration)
	tickPerSecond := int(time.Second / (tickerInterval * time.Millisecond))
	for i := range steps {
		tickArrStartIndex := i * tickPerSecond
		tickArrEndIndex := tickArrStartIndex + tickPerSecond
		segment := e.reqCountArr[tickArrStartIndex:tickArrEndIndex]
		createLinearDistArr(steps[i], segment)
	}
}

// createWavedReqCountArr создает массив запросов с волнообразной нагрузкой.
// Нагрузка колеблется по синусоиде: рост → спад → рост.
// Всего 100 запросов, делятся на 3 "четверти волны" (по логике log₂(TestDuration)).
// Пример распределения по секундам:
// 6, 11, 16, 16, 11, 6, 6, 11, 17, 0
// Внутри каждой секунды — равномерное распределение по тикам.
func (e *engine) createWavedReqCountArr() {
	tickPerSecond := int(time.Second / (tickerInterval * time.Millisecond))
	quarterWaveCount := int((math.Log2(float64(e.heart.TestDuration)))) // Количество четвертей волны
	if quarterWaveCount == 0 {
		quarterWaveCount = 1
	}
	qWaveDuration := int(e.heart.TestDuration / quarterWaveCount)      // Длительность одной четверти волны
	reqCountPerQWave := int(e.heart.IterationCount / quarterWaveCount) // Количество запросов на одну четверть волны
	tickArrStartIndex := 0

	for i := 0; i < quarterWaveCount; i++ {
		if i == quarterWaveCount-1 {
			// Добавление оставшихся запросов к последней волне
			reqCountPerQWave += e.heart.IterationCount - (reqCountPerQWave * quarterWaveCount)
		}

		// Создание инкрементального распределения для текущей четверти волны
		steps := createIncrementalDistArr(reqCountPerQWave, qWaveDuration)
		if i%2 == 1 {
			// Инвертирование волны для создания волнообразного эффекта
			reverse(steps)
		}

		for j := range steps {
			tickArrEndIndex := tickArrStartIndex + tickPerSecond
			segment := e.reqCountArr[tickArrStartIndex:tickArrEndIndex]
			// Создание линейного распределения для текущего сегмента
			createLinearDistArr(steps[j], segment)
			tickArrStartIndex += tickPerSecond
		}
	}
}

// createLinearDistArr создает линейное распределение запросов в массиве.
func createLinearDistArr(count int, arr []int) {
	arrLen := len(arr)
	minReqCount := int(count / arrLen)      // Минимальное количество запросов на элемент
	remaining := count - minReqCount*arrLen // Оставшиеся запросы
	for i := range arr {
		plusOne := 0
		if i < remaining {
			plusOne = 1 // Добавление одного запроса к первым элементам
		}
		reqCount := minReqCount + plusOne
		arr[i] = reqCount
	}
}

// createIncrementalDistArr создает инкрементальное распределение запросов.
func createIncrementalDistArr(count int, len int) []int {
	steps := make([]int, len)
	sum := (len * (len + 1)) / 2                                   // Сумма арифметической прогрессии
	incrementStep := int(math.Ceil(float64(sum) / float64(count))) // Шаг инкремента
	val := 0
	for i := range steps {
		if i > 0 {
			val = steps[i-1]
		}

		if i%incrementStep == 0 {
			steps[i] = val + 1 // Увеличение значения на шаг
		} else {
			steps[i] = val
		}
	}

	sum = arraySum(steps) // Сумма текущего распределения

	factor := count / sum                     // Коэффициент масштабирования
	remaining := count - (sum * factor)       // Оставшиеся запросы
	plus := remaining / len                   // Дополнительные запросы на элемент
	lastRemaining := remaining - (plus * len) // Оставшиеся запросы после распределения
	for i := range steps {
		steps[i] = steps[i]*factor + plus
		if len-i-1 < lastRemaining {
			steps[i]++ // Добавление оставшихся запросов
		}
	}
	return steps
}

// arraySum возвращает сумму элементов массива.
func arraySum(steps []int) int {
	sum := 0
	for i := range steps {
		sum += steps[i]
	}
	return sum
}

// reverse инвертирует порядок элементов в массиве.
func reverse(s interface{}) {
	n := reflect.ValueOf(s).Len()
	swap := reflect.Swapper(s)
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}
