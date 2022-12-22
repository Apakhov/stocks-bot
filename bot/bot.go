package main

import (
	"context"
	"fmt"
	"math"
	"net"
	"time"

	"github.com/Apakhov/stocks-bot/ohlc"
	"github.com/Apakhov/stocks-bot/stockapi"
	"github.com/Apakhov/stocks-bot/tcpproto"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

// StockCommand description for stock command
type StockCommand struct {
	Command string
	Ticker  string
}

// VkRocketBotConfig config for vk rocket bot
type VkRocketBotConfig struct {
	StocksHost    string
	StocksTCPHost string
	TelegramToken string
	TinkoffToken  string
	CommandStocks []*StockCommand
}

// VkRocketBot bot for drawing candlesticks
type VkRocketBot struct {
	botAPI         *tgbotapi.BotAPI
	stockAPIClient stockapi.StockClient

	stocksHost    string
	stocksTCPHost string

	tickerCommands map[string]string
	logger         *zap.Logger
}

// NewVkRocketBot returns new CandlesticksBot
func NewVkRocketBot(cfg *VkRocketBotConfig) (*VkRocketBot, error) {
	bot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		return nil, err
	}
	//bot.Debug = true

	stockAPIClient, err := stockapi.NewTinkoffStockClient(cfg.TinkoffToken)
	if err != nil {
		return nil, err
	}

	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}

	tickerCommands := make(map[string]string, len(cfg.CommandStocks))
	for _, command := range cfg.CommandStocks {
		tickerCommands[command.Command] = command.Ticker
	}

	return &VkRocketBot{
		botAPI:         bot,
		stockAPIClient: stockAPIClient,
		stocksHost:     cfg.StocksHost,
		stocksTCPHost:  cfg.StocksTCPHost,
		tickerCommands: tickerCommands,
		logger:         logger,
	}, nil
}

// GenerateDefaultCaption default generates capition
func (b *VkRocketBot) generateDefaultCaption(ticker string, candles []ohlc.TOHLCV) string {
	openPrice := candles[0].Open
	closePrice := candles[len(candles)-1].Close
	priceDelta := closePrice - openPrice
	percentDelta := priceDelta / openPrice * 100
	percentDeltaAbs := math.Abs(percentDelta)

	grade := "нейтральный"
	if percentDeltaAbs > 1.5 {
		grade = "хороший"
	}
	if percentDeltaAbs > 5 {
		grade = "прекрасный"
	}
	if percentDeltaAbs > 10 {
		grade = "выдающийся"
	}
	if percentDeltaAbs > 15 {
		grade = "фантастический"
	}

	negativeAdj := ""
	if len(grade) > 0 && percentDelta < 0 {
		negativeAdj = " отрицательно"
	}

	return fmt.Sprintf("%s стоит %.2f RUB (%+.2f%% за сутки). Какой%s %s результат!", ticker, closePrice, percentDelta, negativeAdj, grade)
}

func (b *VkRocketBot) requestStock(ticker string, dayAgo time.Time, now time.Time) ([]byte, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", b.stocksTCPHost)
	if err != nil {
		return nil, fmt.Errorf("ResolveTCPAddr failed: %w", err)
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, fmt.Errorf("dial failed: %w", err)
	}
	defer conn.Close()

	err = tcpproto.WriteMsg(conn,
		tcpproto.PrepareStrings(nil,
			ticker,
			dayAgo.Format(time.RFC3339),
			now.Format(time.RFC3339),
			"5min",
		),
	)
	if err != nil {
		return nil, fmt.Errorf("write to server failed: %w", err)
	}

	var imageBytes []byte
	err = tcpproto.ReadMsg(conn, func(buf []byte) error {
		_, err = tcpproto.ParseBytes(buf, &imageBytes)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("read from server failed: %w", err)
	}

	fmt.Printf("read %d img bytes\n", len(imageBytes))
	return imageBytes, nil
}

func (b *VkRocketBot) generalStockHandler(chatID int64, ticker string) {
	now := time.Now()
	dayAgo := now.Add(-24 * time.Hour)

	imgBytes, err := b.requestStock(ticker, dayAgo, now)
	if err != nil {
		b.logger.Info("requesting tcp img: ", zap.Error(err))
	}

	fakeCandle, err := b.stockAPIClient.GetCandlesticks(context.Background(), dayAgo, now, stockapi.CandlestickInterval1Day, ticker)
	if err != nil {
		b.logger.Error("can not fetch tinkoff api: " + err.Error())
		return
	}

	// imageURLRaw := fmt.Sprintf(
	// 	"http://%s/candlesticks/%s/%s/%s/5min/chart.jpg",
	// 	b.stocksHost,
	// 	ticker,
	// 	dayAgo.Format(time.RFC3339),
	// 	now.Format(time.RFC3339),
	// )

	// imageResp, err := http.Get(imageURLRaw)
	// if err != nil {
	// 	b.logger.Info("requesting image", zap.Error(err))
	// 	return
	// }
	// defer imageResp.Body.Close()

	// imageBytes, err := ioutil.ReadAll(imageResp.Body)
	// if err != nil {
	// 	b.logger.Info("reading image", zap.Error(err))
	// 	return
	// }

	resp := tgbotapi.NewPhoto(chatID, tgbotapi.FileBytes{Name: "Some Name", Bytes: imgBytes})
	resp.Caption = b.generateDefaultCaption(ticker, fakeCandle.TOHLCs)
	_, err = b.botAPI.Send(resp)
	if err != nil {
		b.logger.Info(err.Error())
	} else {
		b.logger.Info("ok")

	}

	b.logger.Info(
		"inline query done",
		zap.Time("now", now),
		zap.Int64("chat_id", chatID),
		zap.String("ticker", ticker),
	)
}

// HelpHandler handles help command
func (b *VkRocketBot) HelpHandler(chatID int64) {
	helpMessage := "Available commands:\n"
	for command, ticker := range b.tickerCommands {
		helpMessage += fmt.Sprintf("/%s for %s\n", command, ticker)
	}
	helpMessage += "/start or /help prints this message\n"

	resp := tgbotapi.NewMessage(chatID, helpMessage)
	b.botAPI.Send(resp)
}

// Run start bot and blocks until end
func (b *VkRocketBot) Run() {
	b.logger.Info("bot startup")
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	for update := range b.botAPI.GetUpdatesChan(u) {
		if update.Message == nil {
			continue
		}
		botCommand := update.Message.Command()
		chatID := update.Message.Chat.ID
		if botCommand == "start" || botCommand == "help" {
			b.HelpHandler(chatID)
			continue
		}

		if ticker, ok := b.tickerCommands[botCommand]; ok {
			b.generalStockHandler(chatID, ticker)
		}
	}
}
