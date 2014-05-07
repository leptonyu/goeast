package main

import (
	"flag"
	"github.com/kylelemons/go-gypsy/yaml"
	"github.com/leptonyu/goeast/logic"
	"github.com/leptonyu/wechat/db"
	"log"
	"net/http"
)

func main() {
	file := flag.String("config", "conf/conf.yaml", "Configuration of yaml")
	init := flag.Bool("init", false, "initialize database")
	port := flag.Int("port", 8080, "Http Port")
	flag.Parse()

	config, err := yaml.ReadFile(*file)
	if err != nil {
		log.Fatalf(`Configuration File %q: %s`, *file, err)
	}
	username, _ := config.Get("username")
	password, _ := config.Get("password")
	host, _ := config.Get("host")
	api, _ := config.Get("api")
	if api == "" {
		log.Fatal("api must be set!")
		return
	}
	mongo := db.New(username, password, host, api)
	if *init {
		appid, _ := config.Get("appid")
		secret, _ := config.Get("secret")
		token, _ := config.Get("token")
		if appid == "" || secret == "" || token == "" {
			log.Fatal("appid,secret,token must be set all!")
		}
		if err := mongo.Init(appid, secret, token); err != nil {
			log.Fatal(err)
		}
		logic.Init(mongo)
	} else {
		if *port == 8080 {
			go logic.Spy(mongo)
		}
		logic.Dispatch(mongo)
		http.HandleFunc("/"+api, mongo.GetWeChat())
		http.ListenAndServe(":"+*port, nil)
	}

}
