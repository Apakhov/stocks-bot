package main

import (
	"math/rand"
	"os"
	"time"

	"github.com/Apakhov/stocks-bot/config"
)

type Config struct {
	StocksHost    string `json:"StocksHost"`
	StockTCPHost  string `json:"StockTCPHost"`
	TelegramToken string `json:"TelegramToken"`
	TinkoffToken  string `json:"TinkoffToken"`
}

func main() {
	var conf Config
	config.GetConfig(os.Args, &conf)
	rand.Seed(time.Now().UnixNano())

	cfg := &VkRocketBotConfig{
		StocksHost:    conf.StocksHost,
		StocksTCPHost: conf.StockTCPHost,
		TelegramToken: conf.TelegramToken,
		TinkoffToken:  conf.TinkoffToken,
		CommandStocks: []*StockCommand{
			{
				Command: "vkco",
				Ticker:  "VKCO",
			},
			{
				Command: "sber",
				Ticker:  "SBER",
			},
			{
				Command: "sberp",
				Ticker:  "SBERP",
			},
			{
				Command: "yndx",
				Ticker:  "YNDX",
			},
			{
				Command: "gazp",
				Ticker:  "GAZP",
			},
			{
				Command: "gazp",
				Ticker:  "GAZP",
			},
			{
				Command: "vtbr",
				Ticker:  "VTBR",
			},
			{
				Command: "fixp",
				Ticker:  "FIXP",
			},
			{
				Command: "moex",
				Ticker:  "MOEX",
			},
			{
				Command: "ozon",
				Ticker:  "OZON",
			},
			{
				Command: "rasp",
				Ticker:  "RASP",
			},
			{
				Command: "poly",
				Ticker:  "POLY",
			},
			{
				Command: "aapl",
				Ticker:  "AAPL",
			},
			{
				Command: "tal",
				Ticker:  "TAL",
			},
			{
				Command: "msft",
				Ticker:  "MSFT",
			},
			{
				Command: "spce",
				Ticker:  "SPCE",
			},
			{
				Command: "pfe",
				Ticker:  "PFE",
			},
			{
				Command: "mrna",
				Ticker:  "MRNA",
			},
			{
				Command: "baba",
				Ticker:  "BABA",
			},
			{
				Command: "usd",
				Ticker:  "USDRUB",
			},
		},
	}

	bot, err := NewVkRocketBot(cfg)
	if err != nil {
		panic(err)
	}

	bot.Run()
}
