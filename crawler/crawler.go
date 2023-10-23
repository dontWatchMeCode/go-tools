package crawler

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func StartCrawler() {
	p := tea.NewProgram(newCrawlStatus())

	go func() {
		for {
			pause := time.Duration(rand.Int63n(899)+100) * time.Millisecond
			time.Sleep(pause)

			p.Send(resultUrls{path: "test", duration: pause, status: 200})
		}
	}()

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
