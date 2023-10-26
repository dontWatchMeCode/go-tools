package utils

import (
	"os"

	"github.com/pterm/pterm"
)

func Exit() {
	pterm.Warning.Println("Exiting...")
	os.Exit(0)
}
