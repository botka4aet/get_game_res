package main

import (
	"log"
	"os"
)

func CheckDir(name string) {
	_, err := os.Stat(name)
	if os.IsNotExist(err) {
		if err := os.Mkdir(name, 0777); err != nil {
			log.Fatal(err.Error())
		}
	}
}
