package chartgen

import (
	"bytes"
	"time"

	"gonum.org/v1/plot"

	"github.com/Apakhov/stocks-bot/ohlc"

	"github.com/pkg/errors"
)

const (
	// GraphImageFormat graph image format
	graphImageFormat = "jpg"
	timeTicksFormat  = "15:04"
)

var (
	defaultTimezone, _ = time.LoadLocation("Europe/Moscow")
)

// ChartGenerator generates image with graph
type ChartGenerator struct {
}

// GenerateChart creates graph from CandlesticksData
func (g *ChartGenerator) GenerateChart(data *ohlc.CandlesticksData) ([]byte, error) {
	candlesticksPlot := plot.New()
	candlesticksPlot.Title.Text = data.Name + " (" + data.Ticker + " : " + data.Interval + ") "
	candlesticksPlot.X.Tick.Marker = plot.TimeTicks{
		Ticker: &TimeTicker{Delta: 3600, BetweenCount: 3},
		Format: timeTicksFormat,
		Time:   plot.UnixTimeIn(defaultTimezone),
	}
	candlesticksPlot.Y.Label.Text = data.Currency
	candlesticksPlot.Y.Tick.Marker = &CandlesticksTicker{WantLables: 15}

	candlesticksPlotter := newCandlesticksPlotter(data.TOHLCs, newCandlesticksPlotterOptions())
	candlesticksPlot.Add(candlesticksPlotter)

	gridPlotter := newGrid()
	candlesticksPlot.Add(gridPlotter)

	writerTo, err := candlesticksPlot.WriterTo(720, 480, graphImageFormat)
	if err != nil {
		return nil, errors.Wrap(err, "can not generate graph image")
	}

	var buf bytes.Buffer
	if _, err := writerTo.WriteTo(&buf); err != nil {
		return nil, errors.Wrap(err, "can not save image to bytes")
	}
	return buf.Bytes(), nil
}
