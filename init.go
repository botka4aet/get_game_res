package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
)

var Sitemap = make(map[string]int)
var Filemap = make(map[string]int)
var SitemapID = make(map[int]Site)

type Sites struct {
	Sites []Site `json:"sites"`
}

type Site struct {
	Name    string   `json:"name"`
	Domain  []string `json:"domain"`
	Aliases []Alias  `json:"alias"`
}
type Alias struct {
	Baseurl   string   `json:"baseurl"`
	Backupurl []string `json:"backupurl"`
}

func init() {
	var err error
	//Папка сравнения
	CheckDir("compare")
	//Папка результата
	CheckDir("result")
	//Папка настройки
	CheckDir("config")

	//json time
	_, errf := os.Stat("config/siteconfig.json")
	file, err := os.OpenFile("config/siteconfig.json", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0777)
	if err != nil {
		log.Fatal(err.Error())
	}
	if os.IsNotExist(errf) {
		_, err = file.WriteString(`{"sites":[{"name":"","domain":[""],"alias":[{"baseurl":"","backupurl":[""]}]}]}`)
		if err != nil {
			log.Fatal(err.Error())
		}
	}
	defer file.Close()

	byteValue, _ := io.ReadAll(file)
	var sites Sites
	json.Unmarshal(byteValue, &sites)

	//Храним нумерацию сайтов, где 0 - все лишнее
	SitemapID[0] = Site{Name: "_leftover"}
	var sitecounter int
	for i := 0; i < len(sites.Sites); i++ {
		siteinfo := sites.Sites[i]
		id, ok := Filemap[siteinfo.Name]
		if !ok && siteinfo.Name != "" {
			sitecounter++
			id = sitecounter
			SitemapID[id] = siteinfo
			Filemap[siteinfo.Name] = id
		}
		for _, link := range siteinfo.Domain {
			Sitemap[link] = id
		}
	}
}
