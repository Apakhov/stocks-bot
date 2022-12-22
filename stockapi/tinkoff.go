package stockapi

import (
	"context"
	"time"

	"github.com/Apakhov/stocks-bot/ohlc"

	sdk "github.com/TinkoffCreditSystems/invest-openapi-go-sdk"
	"github.com/pkg/errors"
)

var (
	// ErrUnknownTicker unknown ticker
	ErrUnknownTicker = errors.New("ticker is unknown")
)

// TinkoffStockDescription description.
type TinkoffStockDescription struct {
	FIGI     string
	Ticker   string
	Name     string
	Currency string
}

// TinkoffStockClient client for tinkoff api
type TinkoffStockClient struct {
	client *sdk.SandboxRestClient
	stocks map[string]*TinkoffStockDescription
}

// NewTinkoffStockClient creates new TinkoffStockClient
func NewTinkoffStockClient(token string) (StockClient, error) {
	client := sdk.NewSandboxRestClient(token)
	tcsStocks, err := client.Stocks(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "can not initialize list of available stocks")
	}

	stocks := make(map[string]*TinkoffStockDescription, len(tcsStocks))
	for _, stock := range tcsStocks {
		description := &TinkoffStockDescription{
			FIGI:     stock.FIGI,
			Ticker:   stock.Ticker,
			Name:     stock.Name,
			Currency: string(stock.Currency),
		}
		stocks[stock.Ticker] = description
	}

	stocks["USDRUB"] = &TinkoffStockDescription{
		FIGI:     "BBG0013HGFT4",
		Ticker:   "USD000UTSTOM",
		Name:     "USD",
		Currency: "RUB",
	}

	return &TinkoffStockClient{
		client: client,
		stocks: stocks,
	}, nil
}

// GetCandlesticks returns candlesticks for specified period
func (c *TinkoffStockClient) GetCandlesticks(ctx context.Context, from, to time.Time, interval CandlestickInterval, ticker string) (*ohlc.CandlesticksData, error) {
	tcsDescription, ok := c.stocks[ticker]
	if !ok {
		return nil, ErrUnknownTicker
	}

	candles, err := c.client.Candles(ctx, from, to, transformToTinkoffCandleInterval(interval), tcsDescription.FIGI)
	if err != nil {
		return nil, errors.Wrap(err, "can not get candles")
	}

	tohlcs := make([]ohlc.TOHLCV, 0, len(candles))
	for _, candle := range candles {
		tohlcs = append(tohlcs, ohlc.TOHLCV{
			Timestamp: candle.TS.Unix(),
			OHLCV: ohlc.OHLCV{
				Open:   candle.OpenPrice,
				High:   candle.HighPrice,
				Low:    candle.LowPrice,
				Close:  candle.ClosePrice,
				Volume: candle.Volume,
			},
		})
	}
	return &ohlc.CandlesticksData{
		TOHLCs:   tohlcs,
		Name:     tcsDescription.Name,
		Ticker:   tcsDescription.Ticker,
		Currency: tcsDescription.Currency,
		Interval: interval.String(),
	}, nil
}

func transformToTinkoffCandleInterval(interval CandlestickInterval) sdk.CandleInterval {
	switch interval {
	case CandlestickInterval1Min:
		return sdk.CandleInterval1Min
	case CandlestickInterval5Min:
		return sdk.CandleInterval5Min
	case CandlestickInterval15Min:
		return sdk.CandleInterval15Min
	case CandlestickInterval1Hour:
		return sdk.CandleInterval1Hour
	case CandlestickInterval1Day:
		return sdk.CandleInterval1Day
	case CandlestickInterval1Week:
		return sdk.CandleInterval1Week
	case CandlestickInterval1Month:
		return sdk.CandleInterval1Month
	default:
		panic("wrong CandlestickInterval value")
	}
}
