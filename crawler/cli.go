package crawler

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	purpleText = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff79c6"))
	grayText   = lipgloss.NewStyle().Foreground(lipgloss.Color("#44475a"))
)

type crawlStatus struct {
	spinner  spinner.Model
	results  []resultUrls
	quitting bool
	done     bool
}

type resultUrls struct {
	duration time.Duration
	path     string
	status   int
}

func (r resultUrls) String() string {
	if r.duration == 0 {
		return grayText.Render(strings.Repeat(".", 30))
	}
	return fmt.Sprintf(
		"(%d) %s %s",
		r.status,
		r.path,
		grayText.Render(r.duration.String()),
	)
}

func newCrawlStatus() crawlStatus {
	const numLastResults = 5
	s := spinner.New()
	s.Style = purpleText
	return crawlStatus{
		spinner: s,
		results: make([]resultUrls, numLastResults),
	}
}

func (m crawlStatus) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m crawlStatus) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
		return m, nil
	case resultUrls:
		m.results = append(m.results[1:], msg)
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	default:
		return m, nil
	}
}

func (m crawlStatus) View() string {
	var s string

	if m.done {
		m.quitting = true

		s += "Crawler finished"
		s += "\n\n"
	} else {
		if m.quitting {
			s += "Crawler stopped"
		} else {
			s += m.spinner.View() + " Crawling pages ..."
		}
		s += "\n\n"
	}

	if !m.done {
		for _, res := range m.results {
			s += res.String() + "\n"
		}

		if !m.quitting {
			s += "\n" + grayText.Render("Press Ctrl+C to cancel") + "\n"
		}

		if m.quitting {
			s += "\n"
		}
	}

	return lipgloss.NewStyle().Margin(1, 2, 0, 2).Render(s)
}
