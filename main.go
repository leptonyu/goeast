package main

import (
	"flag"
	"github.com/go-martini/martini"
	"github.com/kylelemons/go-gypsy/yaml"
	"github.com/leptonyu/goeast/logic"
	"github.com/leptonyu/wechat/db"
	"html/template"
	"log"
	"net/http"
	"strconv"
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
		if err := logic.Init(mongo); err != nil {
			log.Fatal(err)
		}
	} else {
		if *port == 8080 {
			go logic.Spy(mongo)
		}
		if err := logic.Dispatch(mongo); err != nil {
			log.Fatal(err)
		}
		m := martini.Classic()
		m.NotFound(Template("404.tpl", m))
		m.Use(martini.Static("static", martini.StaticOptions{Prefix: "static"}))
		m.Get("/", Template("index.tpl", m))
		wc, _ := mongo.GetWeChat()
		m.Get("/"+api, wc.ServeHTTP)
		m.Post("/"+api, wc.ServeHTTP)
		http.ListenAndServe(":"+strconv.Itoa(*port), m)
	}

}

//Create html from templates files.
func Template(key string, m interface{}) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		t := template.New(key)
		t, _ = t.ParseFiles("templates/" + key)
		t.Execute(w, m)
	}
}
