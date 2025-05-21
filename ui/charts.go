package ui

import (
	"bytes"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

func CreateLoadTestCharts() fyne.CanvasObject {
	// Уменьшенные размеры графиков
	chartWidth := 280
	chartHeight := 150

	// Данные для графиков
	x := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	yRPS := []float64{100, 110, 120, 130, 140, 150, 155, 160, 158, 162}
	yDuration := []float64{0.2, 0.21, 0.22, 0.23, 0.22, 0.21, 0.2, 0.19, 0.2, 0.18}
	yErrors := []float64{0, 0, 1, 0, 0, 1, 0, 0, 0, 0}

	// Оптимизированный рендеринг графиков
	rpsChart := renderChart("RPS", x, yRPS, chart.GetDefaultColor(0), chartWidth, chartHeight)
	durationChart := renderChart("Duration (s)", x, yDuration, chart.GetDefaultColor(1), chartWidth, chartHeight)
	errorChart := renderChart("Errors", x, yErrors, chart.GetDefaultColor(2), chartWidth, chartHeight)

	// Компактное расположение
	return container.NewVBox(
		container.NewVBox(
			newChartTitle("График RPS"),
			rpsChart,
		),
		widgetSeparator(),
		container.NewVBox(
			newChartTitle("График времени отклика"),
			durationChart,
		),
		widgetSeparator(),
		container.NewVBox(
			newChartTitle("График ошибок"),
			errorChart,
		),
	)
}

func newChartTitle(text string) *canvas.Text {
	title := canvas.NewText(text, color.Black)
	title.TextSize = 12
	return title
}

func renderChart(title string, x, y []float64, strokeColor color.Color, width, height int) fyne.CanvasObject {
	graph := chart.Chart{
		Width:  width,
		Height: height,
		Background: chart.Style{
			Padding: chart.Box{
				Top:    10,
				Left:   10,
				Right:  10,
				Bottom: 10,
			},
		},
		XAxis: chart.XAxis{
			Style: chart.Style{
				FontSize: 8,
			},
		},
		YAxis: chart.YAxis{
			Style: chart.Style{
				FontSize: 8,
			},
			Name: title,
		},
		Series: []chart.Series{
			chart.ContinuousSeries{
				XValues: x,
				YValues: y,
				Style: chart.Style{
					StrokeColor: toDrawingColor(strokeColor),
					StrokeWidth: 1.5,
					DotWidth:    2,
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := graph.Render(chart.PNG, &buf); err != nil {
		return canvas.NewText("Ошибка рендера", color.RGBA{R: 255, A: 255})
	}

	img := canvas.NewImageFromReader(&buf, "")
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(float32(width), float32(height)))

	return img
}

// Остальные вспомогательные функции остаются без изменений
func toDrawingColor(c color.Color) drawing.Color {
	r, g, b, a := c.RGBA()
	return drawing.Color{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: uint8(a >> 8),
	}
}

func widgetSeparator() *canvas.Line {
	sep := canvas.NewLine(color.Gray{Y: 180})
	sep.StrokeWidth = 1
	return sep
}
