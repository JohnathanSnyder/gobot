package main

import (
       "fmt"
       "os"
       "log"
       "net/url"
       "net/http"
       "github.com/JohnathanSnyder/gobot"
)


func setCookies(bot *gobot.GoBot) {
    url, _ := url.Parse("http://www.reddit.com")
    cookies := make([]*http.Cookie, 0)
    cookie := http.Cookie{Name: "over18", Value: "1"}
    cookies = append(cookies, &cookie)
    bot.Jar.SetCookies(url, cookies)
}

func main() {
    if len(os.Args) <= 1 {
        fmt.Printf("Please enter url.\n")
        os.Exit(1)
    }
    file, err := os.Create("bot.log")
    if err != nil {
      log.Fatal(err)
    }
    log.SetOutput(file)
    seed := os.Args[1]
    bot := gobot.NewGoBot()
    setCookies(bot)
    bot.StartCrawl(seed)
}
