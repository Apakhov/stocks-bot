package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/Apakhov/stocks-bot/chartgen"
	"github.com/Apakhov/stocks-bot/config"
	"github.com/Apakhov/stocks-bot/stockapi"
	"github.com/Apakhov/stocks-bot/tcpproto"

	"github.com/fasthttp/router"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

// HTTPError represents http api error
type HTTPError struct {
	Message string `json:"message"`
}

// StockServerMetrics metrics
type StockServerMetrics struct {
	ChartRequests *prometheus.CounterVec
}

// StockServer server for stocks
type StockServer struct {
	stockAPI       stockapi.StockClient
	chartGenerator *chartgen.ChartGenerator

	metrics *StockServerMetrics
	logger  *zap.Logger
}

// NewStockServer creates new stock server
func NewStockServer(tinkoffToken string) (*StockServer, error) {
	stockAPIClient, err := stockapi.NewTinkoffStockClient(tinkoffToken)
	if err != nil {
		return nil, errors.Wrap(err, "can not initialize stock client")
	}
	generator := &chartgen.ChartGenerator{}
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, errors.Wrap(err, "can not initialize logger")
	}

	logger.Info("server created")
	return &StockServer{
		stockAPI:       stockAPIClient,
		chartGenerator: generator,
		logger:         logger,
		metrics: &StockServerMetrics{
			ChartRequests: prometheus.NewCounterVec(prometheus.CounterOpts{Name: "chart_req"}, []string{"ticker"}),
		},
	}, nil
}

func (s *StockServer) handleRequest(ticker, fromStr, toStr, intervalStr string) ([]byte, error) {
	fmt.Println("handling: ", ticker, fromStr, toStr, intervalStr)

	from, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		return nil, fmt.Errorf("can not parse 'from' path part: %w", err)
	}

	to, err := time.Parse(time.RFC3339, toStr)
	if err != nil {
		return nil, fmt.Errorf("can not parse 'to' path part: %w", err)
	}

	interval, err := stockapi.ParseCandlestickInterval(intervalStr)
	if err != nil {
		return nil, fmt.Errorf("can not parse 'interval' path part: %w", err)
	}

	candlesticksData, err := s.stockAPI.GetCandlesticks(context.Background(), from, to, interval, ticker)
	if err != nil {
		return nil, fmt.Errorf("can not fetch stock api data: %w", err)
	}

	imageBytes, err := s.chartGenerator.GenerateChart(candlesticksData)
	if err != nil {
		return nil, fmt.Errorf("can not generate chart image: %w", err)
	}

	return imageBytes, nil

}

// CandlestickChartHandler handler
func (s *StockServer) CandlestickChartHttpHandler(ctx *fasthttp.RequestCtx) {
	s.metrics.ChartRequests.WithLabelValues("ALL").Inc()
	s.logger.Info("got request", zap.String("uri", ctx.URI().String()))

	ticker := ctx.UserValue("ticker").(string)
	s.metrics.ChartRequests.WithLabelValues(ticker).Inc()

	imageBytes, err := s.handleRequest(
		ctx.UserValue("ticker").(string),
		ctx.UserValue("from").(string),
		ctx.UserValue("to").(string),
		ctx.UserValue("interval").(string),
	)

	if err != nil {
		fmt.Println("err handling", err)
		s.WriteBadRequest(ctx, err.Error())
		return
	}

	if err := s.WriteJPG(ctx, imageBytes); err != nil {
		s.WriteInternalServerError(ctx, "can not write chart image")
		return
	}
}

// WriteBadRequest writes bad request with message
func (s *StockServer) WriteBadRequest(ctx *fasthttp.RequestCtx, message string) {
	if err := s.WriteJSON(ctx, http.StatusBadRequest, &HTTPError{Message: message}); err != nil {
		s.logger.Error("can not send bad request", zap.String("error", err.Error()))
	}
}

// WriteInternalServerError writes internal server error with message
func (s *StockServer) WriteInternalServerError(ctx *fasthttp.RequestCtx, message string) {
	if err := s.WriteJSON(ctx, http.StatusInternalServerError, &HTTPError{Message: message}); err != nil {
		s.logger.Error("can not send internal server error", zap.String("error", err.Error()))
	}
}

// WriteJSON writes json answer
func (s *StockServer) WriteJSON(ctx *fasthttp.RequestCtx, code int, data interface{}) error {
	ctx.Response.Header.Set("Content-Type", "application/json")
	ctx.Response.SetStatusCode(code)
	return json.NewEncoder(ctx).Encode(data)
}

// WriteJPG writes image answer
func (s *StockServer) WriteJPG(ctx *fasthttp.RequestCtx, imageData []byte) error {
	ctx.Response.Header.Set("Content-Type", "image/jpeg")
	_, err := ctx.Write(imageData)
	return err
}

// CandlestickChartHandler handler
func (s *StockServer) CandlestickChartTcpHandler(conn net.Conn) {
	defer conn.Close()

	var ticker, dayAgoStr, nowStr, interval string
	err := tcpproto.ReadMsg(conn, func(buf []byte) error {
		_, err := tcpproto.ParseStrings(buf, &ticker, &dayAgoStr, &nowStr, &interval)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		fmt.Println("Error reading:", err.Error())
	}

	imageBytes, err := s.handleRequest(
		ticker,
		dayAgoStr,
		nowStr,
		interval,
	)
	if err != nil {
		return
	}

	fmt.Println("imageBytes sending size", len(imageBytes))

	tcpproto.WriteMsg(conn, tcpproto.PrepareBytes(nil, imageBytes))
}

func tcpStockServer(stockServer *StockServer, addr string) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()
	fmt.Println("Listening on " + "stockserver:1467")
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			// os.Exit(1)
		}
		// Handle connections in a new goroutine.
		stockServer.CandlestickChartTcpHandler(conn)
	}
}

type Config struct {
	StocksHost   string `json:"StocksHost"`
	StockTCPHost string `json:"StockTCPHost"`
	TinkoffToken string `json:"TinkoffToken"`
}

func main() {
	var conf Config
	config.GetConfig(os.Args, &conf)

	stockServer, err := NewStockServer(conf.TinkoffToken)
	if err != nil {
		panic(err)
	}

	go tcpStockServer(stockServer, conf.StockTCPHost)

	r := router.New()
	r.GET("/candlesticks/{ticker}/{from}/{to}/{interval}/chart.jpg", stockServer.CandlestickChartHttpHandler)

	if err := fasthttp.ListenAndServe(conf.StocksHost, r.Handler); err != nil {
		panic(err)
	}
}
