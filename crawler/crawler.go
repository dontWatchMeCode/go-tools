package crawler

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dontWatchMeCode/go-tools/utils"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/pterm/pterm"
)

var done = false

var statusMap = cmap.New[int]()
var sourceMap = cmap.New[string]()

var logFile *os.File

func Start() {
	if file, err := createTempLogFile(); err == nil {
		logFile = file
		defer logFile.Close()
		defer os.Remove(logFile.Name())
	}

	initialUrl := getInputURL()
	if initialUrl == "" {
		return
	}

	go renderInfoDisplay(initialUrl)
	runCrawler(initialUrl)
}

func createTempLogFile() (*os.File, error) {
	fileName := fmt.Sprintf("dontWatchMeCode-go-tools-crawl-%d.csv", time.Now().Unix())
	tempFilePath := filepath.Join("/tmp", fileName)

	file, err := os.OpenFile(tempFilePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func getInputURL() string {
	result, _ := pterm.DefaultInteractiveTextInput.WithOnInterruptFunc(utils.Exit).Show("URL")

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

	c := colly.NewCollector(colly.AllowedDomains(utils.RemoveHttpPrefix(initialUrl)))
	c.SetClient(httpClient)

	q, _ := queue.New(100, &queue.InMemoryQueueStorage{})

	c.OnScraped(func(r *colly.Response) {
		logUrl(r.Request.URL.String(), r.StatusCode)
		statusMap.Set(r.Request.URL.String(), r.StatusCode)
	})

	c.OnError(func(r *colly.Response, err error) {
		logUrl(r.Request.URL.String(), r.StatusCode)
		statusMap.Set(r.Request.URL.String(), r.StatusCode)
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		url := e.Request.AbsoluteURL(e.Attr("href"))
		url = strings.TrimRight(url, "/")

		if url != "" {
			sourceMap.Set(url, e.Request.URL.String())
			q.AddURL(url)
		}

		go func() {
			if !strings.HasPrefix(url, initialUrl) {
				if statusMap.Has(url) {
					return
				}

				statusMap.Set(url, -1)

				getRequest, err := httpClient.Get(url)
				if err != nil {
					return
				}

				logUrl(url, getRequest.StatusCode)
				statusMap.Set(url, getRequest.StatusCode)
			}
		}()
	})

	q.AddURL(initialUrl)
	q.Run(c)
	c.Wait()

	done = true
}

func logUrl(url string, status int) {
	source, ok := sourceMap.Get(url)
	if !ok {
		source = ""
	}

	logFile.WriteString(fmt.Sprintf("%d|%s|%s\n", status, url, source))
}

func renderInfoDisplay(initialUrl string) {
	area, _ := pterm.DefaultArea.Start()

	for {
		str := fmt.Sprintf("crawling: %s\n", initialUrl)
		str += fmt.Sprintf("logfile: %s\n", logFile.Name())

		count := 0
		errors := 0

		for status := range statusMap.Keys() {
			count++
			if status >= 200 && status <= 399 {
				errors++
			}
		}

		str += pterm.DefaultBox.Sprintf("error: %d\ncrawled: %d", errors, count)

		str += "\n"

		area.Update(str)

		time.Sleep(250 * time.Millisecond)

		if done {
			break
		}
	}
	area.Stop()
}
