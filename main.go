package main

import (
	"fmt"

	"github.com/dontWatchMeCode/go-tools/crawler"
	"github.com/dontWatchMeCode/go-tools/utils"
	"github.com/pterm/pterm"
)

type option struct {
	text     string
	function func()
}

var (
	options = []option{
		{"Crawl page", func() { crawler.Start("") }},
	}
)

func main() {
	handlerFuncs := make(map[string]func())
	for _, option := range options {
		handlerFuncs[option.text] = option.function
	}

	textOptions := make([]string, len(options))
	for i, option := range options {
		textOptions[i] = option.text
	}

	result, _ := pterm.DefaultInteractiveSelect.WithOptions(textOptions).WithOnInterruptFunc(utils.Exit).Show("Please select a function")

	if handlerFunc, exists := handlerFuncs[result]; exists {
		handlerFunc()
	} else {
		panic(fmt.Sprintf("Handler function for %s does not exist", result))
	}
}
