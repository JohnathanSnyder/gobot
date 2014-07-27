// A framework that allows you to create your own web robots!
package gobot

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"runtime"
)

var (
	tagRe  = regexp.MustCompilePOSIX("<a[^>]*>")
	attrRe = regexp.MustCompilePOSIX("href=\"[^\"]*\"")
	urlRe  = regexp.MustCompilePOSIX("http://[^\"]*")
)

type VisitAction func(*http.Response)

type ErrorAction func(*http.Request)

type VisitDecision func(string) bool

type GoBot struct {
	http.Client
	OnVisit     VisitAction
	OnError     ErrorAction
	ShouldVisit VisitDecision
	visited     map[string]bool
}

func NewGoBot() *GoBot {
	bot := new(GoBot)
	bot.OnVisit = defaultVisitAction
	bot.OnError = defaultErrorAction
	bot.ShouldVisit = defaultVisitDecision
	bot.Jar = NewBotCookieJar()
	bot.visited = make(map[string]bool)
	return bot
}

func (bot *GoBot) StartCrawl(seed string) {
	resp, err := bot.Get(seed)
	ncpu := runtime.NumCPU()
	runtime.GOMAXPROCS(ncpu)
	if err != nil {
		log.Printf("StartCrawl: ERROR\n")
	}
	urls := ExtractLinks(resp)
	for i, u := range urls {
		if i == (ncpu - 1) {
			break
		}
		go bot.Crawl(u)
	}
	bot.Crawl(seed)
}

func (bot *GoBot) Crawl(seed string) {
	queue := make([]string, 0)
	currUrl := seed

	for {
		resp, err := bot.Get(currUrl)
		for err != nil {
			go bot.OnError(resp.Request)
			if len(queue) > 0 {
				currUrl = queue[0]
				queue = queue[1:]
				resp, err = bot.Get(currUrl)
			} else {
				os.Exit(1) // TODO: Should find a better way to exit.
			}
		}

		go bot.OnVisit(resp)

		urls := ExtractLinks(resp)
		for _, url := range urls {
			_, present := bot.visited[url]
			if !present {
				queue = append(queue, url)
				bot.visited[url] = true
			}
		}

		if len(queue) > 0 {
			currUrl = queue[0]
			queue = queue[1:]
		} else {
			break
		}

	}
}

// Extracts all the links from the body of an http response.
func ExtractLinks(resp *http.Response) []string {
	urls := make([]string, 0)
	body := ResponseBodyToString(resp)
	tags := tagRe.FindAllString(body, -1)

	for _, tag := range tags {
		url := urlRe.FindString(attrRe.FindString(tag))
		if url != "" {
			urls = append(urls, url)
		}
	}
	return urls
}

// Returns a string of the body of an http response.
func ResponseBodyToString(resp *http.Response) string {
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	return string(body)
}

func defaultVisitAction(resp *http.Response) {
	request := resp.Request
	log.Printf("%s\n", request.URL.String())
}

func defaultErrorAction(resp *http.Request) {
}

func defaultVisitDecision(url string) bool {
	return true
}

type BotCookieJar struct {
	cookies map[string][]*http.Cookie
}

func NewBotCookieJar() *BotCookieJar {
	jar := new(BotCookieJar)
	jar.cookies = make(map[string][]*http.Cookie)
	return jar
}

func (jar *BotCookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	jar.cookies[u.Host] = cookies
}

func (jar *BotCookieJar) Cookies(u *url.URL) []*http.Cookie {
	return jar.cookies[u.Host]
}
