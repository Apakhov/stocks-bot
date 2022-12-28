package main

import (
	"fmt"
	"net/http"
	"os"
	"text/template"

	"github.com/Apakhov/stocks-bot/config"
)

type Config struct {
	HtmlFile   string `json:"HtmlFile"`
	StocksHost string `json:"StocksHost"`
	WebHost    string `json:"WebHost"`
}

type HtmlConf struct {
	StocksHost string
}

func main() {
	var conf Config
	config.GetConfig(os.Args, &conf)

	tmpl := template.Must(template.ParseFiles(conf.HtmlFile))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("make template")
		tmpl.Execute(w, HtmlConf{StocksHost: conf.StocksHost})
	})

	fmt.Println("start web on ", conf.WebHost)
	if err := http.ListenAndServe(conf.WebHost, nil); err != nil {
		fmt.Println(err.Error())
	}
}
