package chartgen

import (
	"image/color"
	"math"

	"github.com/Apakhov/stocks-bot/ohlc"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/font"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

// PaddingConfig padding config
type PaddingConfig struct {
	FromMin float64
	FromMax float64
}

// CandlesticksPlotterOptions options for CandlesticksPlotter
type candlesticksPlotterOptions struct {
	XPadding PaddingConfig
	YPadding PaddingConfig

	GrowColor    color.Color
	FallColor    color.Color
	DefaultColor color.Color
}

// NewCandlesticksPlotterOptions returns CandlesticksPlotterOptions
// with default config
func newCandlesticksPlotterOptions() *candlesticksPlotterOptions {
	return &candlesticksPlotterOptions{
		XPadding: PaddingConfig{
			FromMin: 0,
			FromMax: 900,
		},
		YPadding: PaddingConfig{
			FromMin: 0,
			FromMax: 0,
		},
		GrowColor:    color.RGBA{R: 0, G: 198, B: 107, A: 255},
		FallColor:    color.RGBA{R: 255, G: 98, B: 103, A: 255},
		DefaultColor: color.Black,
	}
}

// CandlesticksPlotter helps to draw candlesticks graph
type candlesticksPlotter struct {
	tohlcs  []ohlc.TOHLCV
	options *candlesticksPlotterOptions

	minX float64
	minY float64
	maxX float64
	maxY float64
}

// NewCandlesticksPlotter creates candlesticks plotter.
func newCandlesticksPlotter(data []ohlc.TOHLCV, opt *candlesticksPlotterOptions) *candlesticksPlotter {
	minX, maxX := float64(math.MaxInt64), float64(math.MinInt64)
	if len(data) > 0 {
		minX = float64(data[0].Timestamp)
		maxX = float64(data[len(data)-1].Timestamp)
	}

	minY, maxY := float64(math.MaxInt64), float64(math.MinInt64)
	for _, tohlc := range data {
		if minY > tohlc.Low {
			minY = tohlc.Low
		}
		if maxY < tohlc.High {
			maxY = tohlc.High
		}
	}

	return &candlesticksPlotter{
		tohlcs:  data,
		options: opt,
		minX:    minX,
		minY:    minY,
		maxX:    maxX,
		maxY:    maxY,
	}
}

// Plot implements the Plot method of the plot.Plotter interface.
func (p *candlesticksPlotter) Plot(c draw.Canvas, plt *plot.Plot) {
	trX, trY := plt.Transforms(&c)

	candleBodyWidth := font.Length(float64(c.Size().X) / float64(len(p.tohlcs)))
	for _, tohlc := range p.tohlcs {
		tsX := trX(float64(tohlc.Timestamp))
		openY := trY(tohlc.Open)
		hightY := trY(tohlc.High)
		lowY := trY(tohlc.Low)
		closeY := trY(tohlc.Close)

		if tohlc.Open < tohlc.Close {
			c.SetColor(p.options.GrowColor)
		} else if tohlc.Open > tohlc.Close {
			c.SetColor(p.options.FallColor)
		} else {
			c.SetColor(p.options.DefaultColor)
		}

		candleBody := vg.Rectangle{
			Min: vg.Point{
				X: tsX - candleBodyWidth/2.,
				Y: openY,
			},
			Max: vg.Point{
				X: tsX + candleBodyWidth/2.,
				Y: closeY,
			},
		}
		c.Fill(candleBody.Path())

		var shadow vg.Path
		shadow.Move(vg.Point{X: tsX, Y: lowY})
		shadow.Line(vg.Point{X: tsX, Y: hightY})
		shadow.Close()
		c.Stroke(shadow)
	}
}

// DataRange implements the DataRange method of the plot.DataRanger interface.
func (p *candlesticksPlotter) DataRange() (xmin, xmax, ymin, ymax float64) {
	return p.minX - p.options.XPadding.FromMin,
		p.maxX + p.options.XPadding.FromMax,
		p.minY - p.options.YPadding.FromMin,
		p.maxY + p.options.YPadding.FromMax
}
