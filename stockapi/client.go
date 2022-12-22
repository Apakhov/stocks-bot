package stockapi

import (
	"context"
	"errors"
	"time"

	"github.com/Apakhov/stocks-bot/ohlc"
)

var (
	// ErrBadCandlestickInterval error for bad CandlestickInterval
	ErrBadCandlestickInterval = errors.New("can not parse CandlestickInterval")
)

// CandlestickInterval interval for one candlestick
type CandlestickInterval int

// Some intervals
const (
	CandlestickInterval1Min CandlestickInterval = iota
	CandlestickInterval5Min
	CandlestickInterval15Min
	CandlestickInterval1Hour
	CandlestickInterval1Day
	CandlestickInterval1Week
	CandlestickInterval1Month
	CandlestickIntervalUnknown
)

// ParseCandlestickInterval returns CandlestickInterval if exists
func ParseCandlestickInterval(interval string) (CandlestickInterval, error) {
	switch interval {
	case "1min":
		return CandlestickInterval1Min, nil
	case "5min":
		return CandlestickInterval5Min, nil
	case "15min":
		return CandlestickInterval15Min, nil
	case "1hour":
		return CandlestickInterval1Hour, nil
	case "1day":
		return CandlestickInterval1Day, nil
	case "1week":
		return CandlestickInterval1Week, nil
	case "1mon":
		return CandlestickInterval1Month, nil
	default:
		return CandlestickIntervalUnknown, ErrBadCandlestickInterval
	}
}

func (i CandlestickInterval) String() string {
	switch i {
	case CandlestickInterval1Min:
		return "1 Min"
	case CandlestickInterval5Min:
		return "5 Min"
	case CandlestickInterval15Min:
		return "15 Min"
	case CandlestickInterval1Hour:
		return "1 Hour"
	case CandlestickInterval1Day:
		return "1 Day"
	case CandlestickInterval1Week:
		return "1 Week"
	case CandlestickInterval1Month:
		return "1 Month"
	default:
		return ""
	}
}

// StockClient client for getting stocks
type StockClient interface {
	GetCandlesticks(ctx context.Context, from, to time.Time, interval CandlestickInterval, ticker string) (*ohlc.CandlesticksData, error)
}
