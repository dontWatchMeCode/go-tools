package main

import (
	"fmt"

	"github.com/dontWatchMeCode/go-tools/crawler"
	"github.com/pterm/pterm"
)

func main() {
	options := []struct {
		text     string
		function func()
	}{
		{"Crawl page", crawler.Start},
	}

	handlerFuncs := make(map[string]func())
	for _, option := range options {
		handlerFuncs[option.text] = option.function
	}

	textOptions := make([]string, len(options))
	for i, option := range options {
		textOptions[i] = option.text
	}

	result, _ := pterm.DefaultInteractiveSelect.WithOptions(textOptions).Show("Please select a function")
	if handlerFunc, exists := handlerFuncs[result]; exists {
		handlerFunc()
	} else {
		panic(fmt.Sprintf("Handler function for %s does not exist", result))
	}
}
