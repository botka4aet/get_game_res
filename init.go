package main

import (
	"bufio"
	"log"
	"os"
	"strings"
)

var WASite = make(map[string]string)
var Sitemap = make(map[string]int)
var SitemapID = make(map[int]string)

func init() {
	var err error
	//Папка сравнения
	_, err = os.Stat("compare")
	if os.IsNotExist(err) {
		if err := os.Mkdir("compare", 0777); err != nil {
			log.Fatal(err.Error())
		}
	}
	//Папка результата
	_, err = os.Stat("result")
	if os.IsNotExist(err) {
		if err := os.Mkdir("result", 0777); err != nil {
			log.Fatal(err.Error())
		}
	}
	//Папка настройки
	_, err = os.Stat("config")
	if os.IsNotExist(err) {
		if err := os.Mkdir("config", 0777); err != nil {
			log.Fatal(err.Error())
		}
	}
	//Белый список и синонимы
	_, err2 := os.Stat("config/sites.txt")
	file, err := os.OpenFile("config/sites.txt", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0777)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer file.Close()

	if os.IsNotExist(err2) {
		_, err = file.WriteString("#Site[:Alias[:Alias[:Alias]]]\n#yandex:ya\n")
		if err != nil {
			panic(err)
		}
	} else {
		SitemapID[0] = "_leftover"
		//Храним нумерацию сайтов, где 0 - все лишнее
		var sitecounter int
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			text := scanner.Text()
			if strings.HasPrefix(text, "#") {
				continue
			}
			res := strings.Split(text, ":")
			id, ok := Sitemap[res[0]]
			if !ok {
				sitecounter++
				id = sitecounter
				SitemapID[id] = res[0]
			}
			for _, link := range res {
				Sitemap[link] = id
			}
		}
	}
}
