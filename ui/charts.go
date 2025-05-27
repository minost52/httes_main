package ui

import (
	"bytes"
	"image"
	"image/color"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

// MetricsData хранит метрики для графиков
type MetricsData struct {
	mu         sync.RWMutex
	Times      []float64 // Временные метки (секунды)
	RPS        []float64 // Requests Per Second
	RespTimes  []float64 // Время отклика (мс)
	Errors     []float64 // Количество ошибок
	updateChan chan struct{}
}

// GlobalMetrics хранит метрики для графиков
var GlobalMetrics = &MetricsData{
	Times:      []float64{0, 1}, // Начальные значения для рендеринга
	RPS:        []float64{0, 0},
	RespTimes:  []float64{0, 0},
	Errors:     []float64{0, 0},
	updateChan: make(chan struct{}, 100), // Буферизованный канал
}

// Добавляем новые данные с блокировкой
func (m *MetricsData) AddData(times, rps, respTimes, errors []float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Times = append(m.Times, times...)
	m.RPS = append(m.RPS, rps...)
	m.RespTimes = append(m.RespTimes, respTimes...)
	m.Errors = append(m.Errors, errors...)

	select {
	case m.updateChan <- struct{}{}:
	default:
	}
}

type chartRenderer struct {
	chart *chart.Chart
	img   *canvas.Image
}

func (r *chartRenderer) render() {
	var buf bytes.Buffer
	if err := r.chart.Render(chart.PNG, &buf); err == nil {
		img, _, err := image.Decode(&buf)
		if err == nil {
			r.img.Image = img
			r.img.Refresh()
		}
	}
}

func toChartColor(c color.Color) drawing.Color {
	r, g, b, a := c.RGBA()
	return drawing.Color{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: uint8(a >> 8),
	}
}

type chartWidget struct {
	renderer *chartRenderer
	mu       sync.Mutex
	title    string
	width    int
	height   int
}

func newChartWidget(title string, lineColor color.Color, width, height int) *chartWidget {
	c := &chartWidget{
		title:  title,
		width:  width,
		height: height,
	}

	// Создаем базовый график
	graph := chart.Chart{
		Width:  width,
		Height: height,
		Background: chart.Style{
			FillColor: drawing.Color{R: 30, G: 30, B: 30, A: 255},
		},
		Canvas: chart.Style{
			FillColor: drawing.Color{R: 30, G: 30, B: 30, A: 255},
		},
		XAxis: chart.XAxis{
			Style: chart.Style{
				FontColor:   drawing.Color{R: 200, G: 200, B: 200, A: 255},
				StrokeColor: drawing.Color{R: 200, G: 200, B: 200, A: 255},
			},
		},
		YAxis: chart.YAxis{
			Name: title,
			Style: chart.Style{
				FontColor:   drawing.Color{R: 200, G: 200, B: 200, A: 255},
				StrokeColor: drawing.Color{R: 200, G: 200, B: 200, A: 255},
			},
		},
	}

	img := canvas.NewImageFromResource(nil)
	img.FillMode = canvas.ImageFillOriginal
	img.SetMinSize(fyne.NewSize(float32(width), float32(height)))

	c.renderer = &chartRenderer{
		chart: &graph,
		img:   img,
	}

	// Первоначальный рендер
	c.update([]float64{0, 1}, []float64{0, 0}, lineColor)
	return c
}

// hasAllZeros проверяет, все ли значения в массиве равны нулю
func hasAllZeros(values []float64) bool {
	for _, v := range values {
		if v != 0 {
			return false
		}
	}
	return true
}

func (c *chartWidget) update(x, y []float64, lineColor color.Color) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Если данные пустые, создаем нулевые значения
	if len(x) == 0 {
		x = []float64{0, 1} // Минимальный временной диапазон
	}
	if len(y) == 0 {
		y = make([]float64, len(x)) // Заполняем нулями
	}

	// Убираем фиксированные диапазоны для автомасштабирования
	c.renderer.chart.XAxis.Range = nil
	c.renderer.chart.YAxis.Range = nil

	// Основная серия данных
	mainSeries := chart.ContinuousSeries{
		XValues: x,
		YValues: y,
		Style: chart.Style{
			StrokeColor: toChartColor(lineColor),
			StrokeWidth: 2,
		},
	}

	// Проверяем, есть ли ненулевые значения в данных
	hasNonZero := false
	for _, val := range y {
		if val != 0 {
			hasNonZero = true
			break
		}
	}

	// Если все значения Y равны нулю, принудительно устанавливаем диапазон Y
	if !hasNonZero {
		// Для всех графиков устанавливаем одинаковый диапазон вокруг нуля
		c.renderer.chart.YAxis.Range = &chart.ContinuousRange{
			Min: -0.1,
			Max: 0.5,
		}

		// Для графика ошибок добавляем точки на линии для лучшей видимости
		if c.title == "Errors" {
			mainSeries.Style.DotWidth = 3
			mainSeries.Style.DotColor = toChartColor(lineColor)
		}
	}

	c.renderer.chart.Series = []chart.Series{mainSeries}
	c.renderer.render()
}

func (c *chartWidget) getImage() *canvas.Image {
	return c.renderer.img
}

func CreateLoadTestCharts() fyne.CanvasObject {
	chartWidth := 300
	chartHeight := 200

	// Создаем виджеты графиков
	rpsChart := newChartWidget("RPS", color.NRGBA{R: 255, A: 255}, chartWidth, chartHeight)
	durationChart := newChartWidget("Duration (ms)", color.NRGBA{G: 255, A: 255}, chartWidth, chartHeight)
	errorChart := newChartWidget("Errors", color.NRGBA{B: 255, A: 255}, chartWidth, chartHeight)

	// Первоначальное обновление
	GlobalMetrics.mu.RLock()
	rpsChart.update(GlobalMetrics.Times, GlobalMetrics.RPS, color.NRGBA{R: 255, A: 255})
	durationChart.update(GlobalMetrics.Times, GlobalMetrics.RespTimes, color.NRGBA{G: 255, A: 255})
	errorChart.update(GlobalMetrics.Times, GlobalMetrics.Errors, color.NRGBA{B: 255, A: 255})
	GlobalMetrics.mu.RUnlock()

	// Запускаем обработчик обновлений
	go func() {
		for range GlobalMetrics.updateChan {
			GlobalMetrics.mu.RLock()
			rpsChart.update(GlobalMetrics.Times, GlobalMetrics.RPS, color.NRGBA{R: 255, A: 255})
			durationChart.update(GlobalMetrics.Times, GlobalMetrics.RespTimes, color.NRGBA{G: 255, A: 255})
			errorChart.update(GlobalMetrics.Times, GlobalMetrics.Errors, color.NRGBA{B: 255, A: 255})
			GlobalMetrics.mu.RUnlock()
		}
	}()

	return container.NewVBox(
		container.NewVBox(
			newChartTitle("График RPS"),
			rpsChart.getImage(),
		),
		widgetSeparator(),
		container.NewVBox(
			newChartTitle("График времени отклика"),
			durationChart.getImage(),
		),
		widgetSeparator(),
		container.NewVBox(
			newChartTitle("График ошибок"),
			errorChart.getImage(),
		),
	)
}

func newChartTitle(text string) *canvas.Text {
	title := canvas.NewText(text, color.White)
	title.TextSize = 12
	return title
}

func widgetSeparator() *canvas.Line {
	sep := canvas.NewLine(color.Gray{Y: 80})
	sep.StrokeWidth = 1
	return sep
}
