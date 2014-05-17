package main

import (
	"flag"
	"github.com/kylelemons/go-gypsy/yaml"
	"github.com/leptonyu/goeast/logic"
	"github.com/leptonyu/wechat"
	"labix.org/v2/mgo"
	"log"
	"regexp"
	"time"
)

func main() {
	date := flag.String("date", "", "LogDate")
	mail := flag.String("mail", "", "Email address")
	flag.Parse()
	reg, err := regexp.Compile("^\\d{8}$")
	if err != nil {
		log.Println("Parameter invalid!")
		return
	}
	config, err := yaml.ReadFile("conf/conf.yaml")
	if err != nil {
		log.Fatalf(`Configuration File: %s`, err)
	}
	if reg.MatchString(*date) {
		d, err := time.Parse("20060102", *date)
		if err != nil {
			log.Println(err)
			return
		}
		d = d.Add(24 * time.Hour)
		x := wechat.NewLocalMongo(config.Require("api"))
		x.Query(func(db *mgo.Database) error {
			logic.SendMail(d.Year(), d.Month(), d.Day(), db, *mail)
			return nil
		})
	}

}
