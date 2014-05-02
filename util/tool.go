package util

import (
	"fmt"
	"github.com/go-martini/martini"
	"github.com/leptonyu/goeast/db"
	"github.com/leptonyu/goeast/wechat"
	"html/template"
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

func StartWeb(port int, dbname string, api string, config *db.DBConfig) {
	m := martini.Classic()
	m.NotFound(Template("404.tpl", m))
	m.Use(martini.Static("static", martini.StaticOptions{Prefix: "static"}))
	m.Get("/", Template("index.tpl", m))
	wc, err := config.CreateWeChat(dbname, api)
	if err != nil {
		panic(err)
	}
	if port == 8080 {
		go func() {
			for {
				a, err := wc.UpdateAccessToken()
				if err != nil {
					panic(err)
				}
				time.Sleep((-1 * time.Duration(time.Since(a.ExpireTime).Seconds())))
			}
		}()
		f := func(wait time.Duration, keys ...string) {
			if len(keys) > 0 {
				for {
					time.Sleep(wait)
					for _, v := range keys {
						config.Update(v)
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
		w.ReplyText(txt)
		return nil
	})
	//Create api route
	ff := wc.CreateHandlerFunc()
	m.Get("/"+api, ff)
	m.Post("/"+api, ff)
	err = http.ListenAndServe(":"+strconv.Itoa(port), m)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("stop!")
	os.Exit(0)
}
