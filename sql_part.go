package main

import (
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

type FileId struct {
	gorm.Model
	//Делаем ссылку уникальной, иначе gorm clause работать не будет
	Link string `gorm:"unique"`
	Hash string
}

func addlines(siteID int, link []string, lockCh chan<- string) {
	//Получим название файла
	site := SitemapID[siteID].Name
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
		}).CreateInBatches(&files,1000)

		//Закрываем соединения, тк позже база переоткрывается
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
