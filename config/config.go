package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

func GetConfig(args []string, conf interface{}) {
	if len(args) < 2 {
		fmt.Println("no config provided")
		os.Exit(1)
	}

	confBytes, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		fmt.Println("cant read config file", err)
		os.Exit(1)
	}

	err = json.Unmarshal(confBytes, &conf)
	if err != nil {
		fmt.Println("cant unmarshal config file", err)
		os.Exit(1)
	}
}
