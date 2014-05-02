package util

import (
	"fmt"
	"github.com/go-martini/martini"
	"github.com/leptonyu/goeast/db"
	"github.com/leptonyu/goeast/wechat"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func Template(key string, m interface{}) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		t := template.New(key)
		t, _ = t.ParseFiles("templates/" + key)
		t.Execute(w, m)
	}
}

func StartWeb(port int, config *db.DBConfig) {
	m := martini.Classic()
	m.NotFound(Template("404.tpl", m))
	m.Use(martini.Static("static", martini.StaticOptions{Prefix: "static"}))
	m.Get("/", Template("index.tpl", m))
	wc, err := config.CreateWeChat()
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}
	if port == 8080 {
		go func() {
			for {
				_, err := wc.UpdateAccessToken()
				if err != nil {
					log.Fatalln(err)
				}
				time.Sleep(10 * time.Second)
			}
		}()
		f := func(wait time.Duration, keys ...string) {
			if len(keys) > 0 {
				m := map[string]db.Msg{}
				for _, v := range keys {
					r, err := config.QueryMsg(v)
					if err == nil {
						m[v] = *r
					}
				}
				for {
					time.Sleep(10 * time.Second)
					for _, v := range keys {
						msg, ok := m[v]
						if ok && time.Since(msg.CreateTime.Add(wait)).Seconds() >= 0 {
							fmt.Println(v)
							config.UpdateMsg(v)
						}
					}
				}
			}
		}
		go f(30*time.Minute, db.Blog, db.Events)
		go f(24*time.Hour,
			db.Home,
			db.Campus,
			db.Contact,
			db.Galleries,
			db.One2one,
			db.Online,
			db.Onsite,
			db.Teachers,
			db.Testimonials)
	}
	wc.HandleFunc(wechat.MsgTypeText, func(w wechat.ResponseWriter, r *wechat.Request) error {
		txt := r.Content
		//sig := strings.ToLower(txt)
		log.Println(r)
		w.ReplyText(txt)
		return nil
	})
	//Create api route
	ff := wc.CreateHandlerFunc()
	m.Get("/"+config.DBName, ff)
	m.Post("/"+config.DBName, ff)
	err = http.ListenAndServe(":"+strconv.Itoa(port), m)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("stop!")
	os.Exit(0)
}
