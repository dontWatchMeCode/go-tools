package crawler

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
	"github.com/pterm/pterm"
)

var (
	urlChannel   = make(chan urlWithStatus)
	requestCount = 0
	errorCount   = 0
	isDone       = false
)

type urlWithStatus struct {
	url    string
	status int
}

func Start() {
	initialUrl := getInputURL()
	if initialUrl == "" {
		return
	}

	go runCrawler(initialUrl)
	displayResults()
}

func getInputURL() string {
	result, _ := pterm.DefaultInteractiveTextInput.Show("URL")

	_, err := url.ParseRequestURI(result)
	if err != nil {
		pterm.Error.Println("Error parsing URL:", err)
		return ""
	}

	return result
}

func runCrawler(initialUrl string) {
	transport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	c := colly.NewCollector(colly.AllowedDomains(removeHttpPrefix(initialUrl)))
	c.SetClient(httpClient)

	q, _ := queue.New(10, &queue.InMemoryQueueStorage{MaxSize: 10000})

	var mu sync.Mutex

	c.OnScraped(func(r *colly.Response) {
		mu.Lock()
		defer mu.Unlock()

		requestCount++
		urlChannel <- urlWithStatus{
			url:    r.Request.URL.String(),
			status: r.StatusCode,
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		mu.Lock()
		defer mu.Unlock()

		errorCount++
		urlChannel <- urlWithStatus{
			url:    r.Request.URL.String(),
			status: r.StatusCode,
		}
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		mu.Lock()
		defer mu.Unlock()

		url := e.Request.AbsoluteURL(e.Attr("href"))
		if url == "" {
			return
		}
		if strings.HasPrefix(url, initialUrl) {
			q.AddURL(strings.TrimRight(url, "/"))
		}
	})

	go func() {
		q.AddURL(initialUrl)
		q.Run(c)
		c.Wait()
		isDone = true
	}()
}

func displayResults() {
	limit := 5
	data := []urlWithStatus{}
	area, _ := pterm.DefaultArea.Start()

	for {
		str := "\n"

		msg, ok := <-urlChannel
		if ok {
			if len(data) >= limit {
				data = data[1:]
			}
			data = append(data, msg)
		} else {
			break
		}

		for _, msg := range data {
			str += fmt.Sprintf("%d - %s \n", msg.status, msg.url)
		}

		str += fmt.Sprintf("\n-> requests: %d", requestCount)
		str += fmt.Sprintf("\n-> errors: %d", errorCount)
		str += "\n"

		area.Update(str)

		if isDone {
			area.Stop()
			break
		}
	}

	fmt.Println()
	pterm.Success.Println("Done crawling!")
}
