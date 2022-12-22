package chartgen

import (
	"image/color"
	"math"

	"github.com/Apakhov/stocks-bot/ohlc"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

// volumePlotterOptions options for VbarPlotter
type volumePlotterOptions struct {
	GrowColor    color.Color
	FallColor    color.Color
	DefaultColor color.Color
}

// NewVolumePlotterOptions returns VolumePlotterOptions
// with default config
func newVolumePlotterOptions() *volumePlotterOptions {
	return &volumePlotterOptions{
		GrowColor:    color.RGBA{R: 0, G: 198, B: 107, A: 255},
		FallColor:    color.RGBA{R: 255, G: 98, B: 103, A: 255},
		DefaultColor: color.Black,
	}
}

type volumePlotter struct {
	tohlcvs []ohlc.TOHLCV
	options *volumePlotterOptions

	minX float64
	minY float64
	maxX float64
	maxY float64
}

func newVolumePlotter(data []ohlc.TOHLCV, opt *volumePlotterOptions) *volumePlotter {
	minX, maxX := float64(math.MaxInt64), float64(math.MinInt64)
	if len(data) > 0 {
		minX = float64(data[0].Timestamp)
		maxX = float64(data[len(data)-1].Timestamp)
	}

	maxY := float64(math.MinInt64)
	for _, tohlc := range data {
		if maxY < tohlc.Volume {
			maxY = tohlc.Volume
		}
	}

	return &volumePlotter{
		tohlcvs: data,
		options: opt,
		minX:    minX,
		minY:    0.,
		maxX:    maxX,
		maxY:    maxY,
	}
}

// Plot implements the Plot method of the plot.Plotter interface.
func (p *volumePlotter) Plot(c draw.Canvas, plt *plot.Plot) {
	trX, trY := plt.Transforms(&c)

	for _, tohlcv := range p.tohlcvs {
		tsX := trX(float64(tohlcv.Timestamp))
		barStartY := trY(0)
		barEndY := trY(tohlcv.Volume)

		if tohlcv.Open < tohlcv.Close {
			c.SetColor(p.options.GrowColor)
		} else if tohlcv.Open > tohlcv.Close {
			c.SetColor(p.options.FallColor)
		} else {
			c.SetColor(p.options.DefaultColor)
		}

		var bar vg.Path
		bar.Move(vg.Point{X: tsX, Y: barStartY})
		bar.Line(vg.Point{X: tsX, Y: barEndY})
		bar.Close()
		c.Stroke(bar)
	}
}

// DataRange implements the DataRange method of the plot.DataRanger interface.
func (p *volumePlotter) DataRange() (xmin, xmax, ymin, ymax float64) {
	return p.minX, p.maxX, p.minY, p.maxY
}
