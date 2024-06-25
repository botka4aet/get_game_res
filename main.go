package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

func main() {
	var addmap = make(map[int][]string)
	filladdmap(addmap)
	//Запомним - какие из сайтов были запущены, а какие - нужно запустить
	var SitemapUsed = make(map[int]bool)
	var gorocounter int
	lockCh := make(chan string)
	StartTime := time.Now()
	for siteID, link := range addmap {
		gorocounter++
		go addlines(siteID, link, lockCh)
		SitemapUsed[siteID] = true
	}
	for gorocounter > 0 {
		fmt.Printf("%v - %v\n", <-lockCh, time.Since(StartTime))
		gorocounter--
	}

	//Очищение папки сравнения
	os.RemoveAll("compare")
	if err := os.Mkdir("compare", 0777); err != nil {
		log.Fatal(err.Error())
	}

	var client = &http.Client{
		Timeout: time.Second * 10,
	}
	res, err := client.Head("https://about.google/assets-main/img/glue-google-color-logo.svg")
	if err != nil {
		if os.IsTimeout(err) {
			// timeout
			fmt.Println("TimeOut!")
		} else {
			panic(err)
		}
	}
	fmt.Println("Status:", res.StatusCode)
	fmt.Println("ContentLength:", res.ContentLength)
}

func filladdmap(addmap map[int][]string) {
	var fullmap = make(map[string]bool)
	fillmap("result", addmap, fullmap)
	fillmap("compare", addmap, fullmap)
}

func fillmap(fname string, addmap map[int][]string, fullmap map[string]bool) {
	adf := false
	if fname != "result" {
		adf = true
	}
	file, err := os.Open(fname)
	if err != nil {
		log.Fatalf("failed opening directory: %s", err)
	}
	defer file.Close()

	//read all files and folders
	list, _ := file.Readdirnames(0)
	var counter int
	for _, name := range list {
		if strings.HasSuffix(name, ".txt") {
			file, err := os.Open(fname + "/" + name)
			if err != nil {
				log.Fatal(fname, "/", name, " - ", err)
			}
			defer file.Close()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				text := scanner.Text()

				//Если уже есть, то пропускаем
				_, ok := fullmap[text]
				if ok || text == "" {
					continue
				}
				if adf {
					domain := extractDomain(text)

					var siteID int
					surl := strings.Split(domain, ".")
					if len(surl) > 2 {
						domain = surl[len(surl)-2] + "." + surl[len(surl)-1]
					}

					siteID, ok = Sitemap[domain]
					//Если сайт в списке нужных, то увеличиваем счетчик, ищем синонимы
					if ok {
						counter++
					}

					addmap[siteID] = append(addmap[siteID], text)
				}
				fullmap[text] = false
			}
			if err := scanner.Err(); err != nil {
				log.Fatal(err)
			}
		}
	}
	//Срабатывает, если в папке есть новые строки. Для result это всегда 0
	if counter > 0 {
		fmt.Println("Added new lines - ", counter)
	}
}

// get domain name via SOuser1660210
func extractDomain(urlLikeString string) string {

	urlLikeString = strings.TrimSpace(urlLikeString)

	if regexp.MustCompile(`^https?`).MatchString(urlLikeString) {
		read, _ := url.Parse(urlLikeString)
		urlLikeString = read.Host
	}

	if regexp.MustCompile(`^www\.`).MatchString(urlLikeString) {
		urlLikeString = regexp.MustCompile(`^www\.`).ReplaceAllString(urlLikeString, "")
	}

	return regexp.MustCompile(`([a-z0-9\-]+\.)+[a-z0-9\-]+`).FindString(urlLikeString)
}
