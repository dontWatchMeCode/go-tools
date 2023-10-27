package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pterm/pterm"
)

func Exit() {
	pterm.Warning.Println("Exiting ...")
	os.Exit(0)
}

func CreateTempFile(identifier string) (*os.File, error) {
	fileName := fmt.Sprintf("dontWatchMeCode-go-tools-%s-%d.csv", identifier, time.Now().Unix())
	tempFilePath := filepath.Join("/tmp", fileName)

	file, err := os.OpenFile(tempFilePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	return file, nil
}
