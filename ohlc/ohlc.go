package ohlc

// OHLCV describes Open/High/Low/Close/Volume stock data.
type OHLCV struct {
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
}

// TOHLCV is OHLCV with timestamp
type TOHLCV struct {
	OHLCV
	Timestamp int64
}

// CandlesticksData TOHLCs data with metainformation
type CandlesticksData struct {
	Ticker   string
	Name     string
	Currency string
	Interval string
	TOHLCs   []TOHLCV
}
