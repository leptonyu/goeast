package util

import (
	"fmt"
	"github.com/leptonyu/goeast/db"
	"github.com/leptonyu/goeast/wechat"
	"log"
	"time"
)

// Register Dispatch handler
func DispatchFunc(config *db.DBConfig, wc *wechat.WeChat) {
	//Text request
	wc.HandleFunc(wechat.MsgTypeText, func(w wechat.ResponseWriter, r *wechat.Request) error {
		txt := r.Content
		//sig := strings.ToLower(txt)
		log.Println(r)
		w.ReplyText(txt)
		return nil
	})
}

func DeamonTask(config *db.DBConfig) {
	go func() {
		for {
			_, err := config.WC.UpdateAccessToken()
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
