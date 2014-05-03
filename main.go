package main

import (
	"flag"
	"github.com/leptonyu/goeast/db"
	"github.com/leptonyu/goeast/util"
	"log"
	"os"
)

func main() {
	port := flag.Int("port", 8080, "Web service port")
	api := flag.String("api", "api", "http://localhost/$api,\n	Also use as database name with prefix wechat_")
	appid := flag.String("appid", "", "App id")
	secret := flag.String("secret", "", "App secret")
	token := flag.String("token", "", "Token")
	init := flag.Bool("init", false, "Init ")
	help := flag.Bool("h", false, "Help")
	flag.Parse()
	if *help {
		flag.PrintDefaults()
		os.Exit(0)
	}
	config, err := db.NewDBConfig(*api)
	if *init {
		err := config.Init(*appid, *secret, *token)
		if err != nil {
			log.Panicln(err)
			os.Exit(1)
		}
		func(keys ...string) {
			for _, key := range keys {
				log.Println("Fetching web " + key)
				config.UpdateMsg(key)
			}
		}(db.Blog, db.Events,
			db.Home,
			db.Campus,
			db.Contact,
			db.Galleries,
			db.One2one,
			db.Online,
			db.Onsite,
			db.Teachers,
			db.Testimonials)
		os.Exit(0)
	}

	if err != nil {
		log.Panic(err)
		os.Exit(1)
	}

	util.StartWeb(*port, config)
}
