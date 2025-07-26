package config

import (
	"log"
	"os"
)

func SetupStorage() {
	_, err := os.Stat("./storage")
	if err != nil {
		if os.IsNotExist(err) {
			err := os.Mkdir("storage", 0755)
			if err != nil {
				log.Fatal(err.Error())
			}
		}
		log.Fatal(err.Error())
	}
}
