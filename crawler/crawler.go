package crawler

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
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

var logFiles = make(map[string]*os.File)
var logFileMu sync.Mutex
var logPrefix = ""

type returnValues struct {
	statusMap  map[string]int
	sourceMap  map[string]string
	initialUrl string
	success    bool
}

func Start() returnValues {
	initialUrl := getInputURL()
	logPrefix = getFileNamePrefix(initialUrl)

	if initialUrl == "" {
		return returnValues{
			statusMap:  statusMap.Items(),
			sourceMap:  sourceMap.Items(),
			initialUrl: initialUrl,
			success:    false,
		}
	}

	go renderInfoDisplay(initialUrl)
	runCrawler(initialUrl)
	writeToFiles(initialUrl)

	for _, file := range logFiles {
		if strings.Contains(file.Name(), "-temp") {
			os.Remove(file.Name())
		}

		file.Close()
	}

	return returnValues{
		statusMap:  statusMap.Items(),
		sourceMap:  sourceMap.Items(),
		initialUrl: initialUrl,
		success:    true,
	}
}

func getInputURL() string {
	result, _ := pterm.DefaultInteractiveTextInput.WithOnInterruptFunc(utils.Exit).Show("URL")

	_, err := url.ParseRequestURI(result)
	if err != nil {
		pterm.Error.Println("Error parsing URL:", err)
		utils.Exit()
		return ""
	}

	return result
}

func getFileNamePrefix(initialUrl string) string {
	result, _ := pterm.DefaultInteractiveTextInput.WithOnInterruptFunc(utils.Exit).Show("File prefix (leave blank for url + timestamp)")

	if result == "" {
		result = utils.RemoveHttpPrefix(initialUrl) + "_" + fmt.Sprintf("%d", time.Now().Unix())
		result = strings.ReplaceAll(result, "www.", "")
		result = strings.ReplaceAll(result, ".", "_")
	}

	specialCharPattern := regexp.MustCompile("[^a-zA-Z0-9_]+")
	return specialCharPattern.ReplaceAllString(result, "")
}

func runCrawler(initialUrl string) {
	transport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   300 * time.Second,
	}

	c := colly.NewCollector(colly.AllowedDomains(utils.RemoveHttpPrefix(initialUrl)))
	c.SetClient(httpClient)

	q, _ := queue.New(50, &queue.InMemoryQueueStorage{})

	c.OnScraped(func(r *colly.Response) {
		processUrl(httpClient, r.Request.URL.String(), r.StatusCode)
	})

	c.OnError(func(r *colly.Response, err error) {
		processUrl(httpClient, r.Request.URL.String(), r.StatusCode)
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		urlVal := e.Request.AbsoluteURL(e.Attr("href"))

		urlVal = strings.TrimRight(urlVal, "/")
		urlVal = strings.TrimSpace(urlVal)

		if urlVal != "" {
			sourceMap.Set(urlVal, e.Request.URL.String())
			q.AddURL(urlVal)
		}

		go func() {
			if !strings.HasPrefix(urlVal, initialUrl) {
				if statusMap.Has(urlVal) {
					return
				}

				statusMap.Set(urlVal, 0)

				_, err := url.ParseRequestURI(urlVal)
				if err != nil {
					return
				}

				getRequest, err := httpClient.Get(urlVal)
				if err != nil {
					return
				}

				processUrl(httpClient, urlVal, getRequest.StatusCode)
			}
		}()
	})

	q.AddURL(initialUrl)
	q.Run(c)
	c.Wait()

	done = true
}

func writeToFiles(initialUrl string) {
	for url, status := range statusMap.Items() {
		firstStatusNumber := string(strconv.Itoa(status)[0])

		source, ok := sourceMap.Get(url)
		if !ok {
			source = ""
		}

		if url == "" && status == 0 && source == "" {
			continue
		}

		logStringToFileCreateIfNotExists(
			logPrefix+"-full.csv",
			fmt.Sprintf("%d|%s|%s|\n", status, url, source),
		)

		logStringToFileCreateIfNotExists(
			logPrefix+"-"+firstStatusNumber+"xx.csv",
			fmt.Sprintf("%d|%s|%s|\n", status, url, source),
		)

		if !strings.HasPrefix(url, initialUrl) {
			logStringToFileCreateIfNotExists(
				logPrefix+"-external.csv",
				fmt.Sprintf("%d|%s|%s|\n", status, url, source),
			)
		}
	}
}

func processUrl(httpClient *http.Client, url string, status int) {
	if status == 0 {
		r, err := httpClient.Head(url)
		if err != nil {
			panic(r)
		}
		status = r.StatusCode
	}

	logUrl(url, status)
	statusMap.Set(url, status)
}

func logUrl(url string, status int) {
	source, ok := sourceMap.Get(url)
	if !ok {
		source = ""
	}

	logStringToFileCreateIfNotExists(
		logPrefix+"-temp.csv",
		fmt.Sprintf("%d|%s|%s|\n", status, url, source),
	)
}

func logStringToFileCreateIfNotExists(fileName string, str string) {
	logFileMu.Lock()
	defer logFileMu.Unlock()

	if logFiles[fileName] == nil {
		if file, err := os.OpenFile(fileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644); err == nil {
			logFiles[fileName] = file
		}
	}

	logFiles[fileName].WriteString(str)
}

func renderInfoDisplay(initialUrl string) {
	area, _ := pterm.DefaultArea.Start()

	for {
		str := fmt.Sprintf("crawling: %s\n", initialUrl)

		str += pterm.DefaultBox.Sprintf("crawled: %d", statusMap.Count())

		str += "\n"

		area.Update(str)

		time.Sleep(250 * time.Millisecond)

		if done {
			break
		}
	}
	area.Stop()
}
