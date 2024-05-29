package main

import (
	"bufio"
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var Fullmap = make(map[string]bool)
var Addmap = make(map[int][]string)

type FileId struct {
	gorm.Model
	//Делаем ссылку уникальной, иначе gorm clause работать не будет
	Link string `gorm:"unique"`
	Hash string
}

func init() {
}

func addlines(siteID int, link []string, lockCh chan<- string) {
	//Получим название файла
	site := SitemapID[siteID]
	//Пишем в базу только то, что нужно
	if siteID != 0 {
		//Создаем папку, если ее нет
		CheckDir("result/" + site)
		db, err := gorm.Open(sqlite.Open("result/"+site+"/database.db"), &gorm.Config{
			//Меняем уровень логера, тк при создании 200+ строк выскакивают предупреждения о скорости
			Logger: logger.Default.LogMode(logger.Error),
		})
		if err != nil {
			panic("failed to connect database")
		}
		db.AutoMigrate(&FileId{})
		var files = []FileId{}
		//Создаем записи в дб для новых строк
		for _, k := range link {
			files = append(files, FileId{Link: k, Hash: ""})
		}
		db.Model(&FileId{}).Clauses(clause.OnConflict{
			DoNothing: true,
		}).Create(&files)

        sqlDB, _ := db.DB()
        sqlDB.Close()
	}

	//Откроем-создадим файл для дозаписи новых строк
	fi, err := os.OpenFile("result/"+site+".txt", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0777)
	if err != nil {
		panic(err)
	}
	defer fi.Close()
	for _, k := range link {
		_, err = fi.WriteString(k + "\n")
		if err != nil {
			panic(err)
		}
	}
	lockCh <- site
}

func main() {
	fillmap("result")
	fillmap("compare")
	var gorocounter int
	lockCh := make(chan string)
	StartTime := time.Now()
	for siteID, link := range Addmap {
		gorocounter++
		go addlines(siteID, link, lockCh)
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

func fillmap(fname string) {
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
				_, ok := Fullmap[text]
				if ok {
					continue
				}
				if adf {
					var siteID int
					if !strings.HasPrefix(text, "_") {
						url, err := url.Parse(text)
						if err != nil {
							log.Fatal(err, " - ", text)
						}
						surl := strings.Split(url.Hostname(), ".")
						site := surl[max(len(surl)-2, 0)]
						siteID, ok = Sitemap[site]
						//Если сайт в списке нужных, то увеличиваем счетчик
						if ok {
							counter++
						}
					}
					Addmap[siteID] = append(Addmap[siteID], text)
				}
				Fullmap[text] = false
			}
			if err := scanner.Err(); err != nil {
				log.Fatal(err)
			}
		}
	}
	if adf {
		fmt.Println("Added new lines - ", counter)
	}
}
