package crawler

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/gocolly/colly"
	"github.com/gocolly/colly/queue"
	"github.com/pterm/pterm"
)

type urlWithStatus struct {
	url    string
	status int
}

func Start() {
	result, _ := pterm.DefaultInteractiveTextInput.Show("URL")

	_, err := url.ParseRequestURI(result)
	if err != nil {
		pterm.Error.Println("Error parsing URL:", err)
		return
	}

	initialUrl := result

	urlChannel := make(chan urlWithStatus)
	requestCount := 0
	errorCount := 0

	c := colly.NewCollector()
	q, _ := queue.New(10, &queue.InMemoryQueueStorage{MaxSize: 10000})

	c.OnScraped(func(r *colly.Response) {
		requestCount++
		urlChannel <- urlWithStatus{
			url:    r.Request.URL.String(),
			status: r.StatusCode,
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		errorCount++
		urlChannel <- urlWithStatus{
			url:    r.Request.URL.String(),
			status: r.StatusCode,
		}

	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		url := e.Request.AbsoluteURL(e.Attr("href"))

		if url == "" {
			return
		}

		if strings.HasPrefix(url, initialUrl) {
			q.AddURL(url)
		}
	})

	isDone := false

	go func() {
		q.AddURL(initialUrl)
		q.Run(c)
		c.Wait()

		isDone = true
		fmt.Println(isDone)
	}()

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
			data = append(data, urlWithStatus{
				url:    msg.url,
				status: msg.status,
			})
		} else {
			break
		}

		for _, msg := range data {
			str += fmt.Sprint(msg.status) + " - " + msg.url + "\n"
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
