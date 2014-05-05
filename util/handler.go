package util

import (
	"errors"
	"fmt"
	json "github.com/bitly/go-simplejson"
	"github.com/leptonyu/goeast/db"
	"github.com/leptonyu/goeast/wechat"
	"log"
	"os"
	"os/signal"
	"regexp"
	"sort"
	"strings"
	"time"
)

var BaseTime = time.Date(1970, 1, 1, 0, 0, 0, 0, time.Local).Add(8 * time.Hour)

type dispatchRoute struct {
	regx    *regexp.Regexp
	handler RouteFunc
}

type RouteFunc func(*db.DBConfig, wechat.ResponseWriter, *wechat.Request) error

type routes struct {
	rs []*dispatchRoute
}

// Register Dispatch handler
func DispatchFunc(config *db.DBConfig, wc *wechat.WeChat) {
	//Text request
	drs := &routes{}
	drs.register(`^(admin|adm|管理员)(:(help|account))?$`, db.Admin())
	drs.register(`^\s*(contact(\-us)?|maria|emily|jane)\s*$`, db.Teacher())
	drs.register(`^\s*(help|h|帮助)\s*$`, help)
	drs.register(`^\s*(events?|事件|活动)\s*$`, event)
	drs.register(`^\s*(blogs?|日志|博客)\s*$`, blog)
	drs.register(`^\s*(today|今日|当日)\s*$`, today)
	drs.register(`^\s*(week|本周|这周)\s*$`, week)
	drs.register(`^\s*(home|主页|首页)\s*$`, home)
	wc.HandleFunc(wechat.MsgTypeText, func(w wechat.ResponseWriter, r *wechat.Request) error {
		txt := r.Content
		sig := strings.ToLower(txt)
		for _, v := range drs.rs {
			if v.regx.MatchString(sig) {
				err := v.handler(config, w, r)
				if err == nil {
					return nil
				}
			}
		}
		w.ReplyText("You said: " + txt + `
Reply Help to get helps.`)
		return nil
	})
	wc.HandleFunc(wechat.MsgTypeEventSubscribe, func(w wechat.ResponseWriter, r *wechat.Request) error {
		return help(config, w, r)
	})
	wc.HandleFunc(wechat.MsgTypeEventScan, func(w wechat.ResponseWriter, r *wechat.Request) error {
		w.ReplyText("Aha! What are you doing?")
		return nil
	})
	wc.HandleFunc(wechat.MsgTypeEventClick, func(w wechat.ResponseWriter, r *wechat.Request) error {
		txt := r.EventKey
		sig := strings.ToLower(txt)
		for _, v := range drs.rs {
			if v.regx.MatchString(sig) {
				err := v.handler(config, w, r)
				if err == nil {
					return nil
				}
			}
		}
		w.ReplyText("Oops! Wrong key: " + txt)
		return nil
	})
}

func home(c *db.DBConfig, w wechat.ResponseWriter, r *wechat.Request) error {
	w.ReplyNews([]wechat.Article{
		wechat.Article{
			Title:       "GoEast Language Centers",
			PicUrl:      "http://static.squarespace.com/static/52141adee4b0476aaa6af594/t/52e83ec9e4b0dfc630da5953/1390952403577/XJL_8779-small.jpg",
			Url:         db.Url,
			Description: `GoEast Language centers is one of Shanghai's premiere Mandarin language schools. We offer tutoring services, online classes, and on-site classes at our beautiful Shanghai campus. Our mission is to teach Mandarin and Chinese culture in a fun, personal, and effective environment. At GoEast we focus on helping you achieve your learning goals with as much personal attention as possible and no self help gimmicks! Learning language is best done with others, and we hope you'll join us at GoEast! `,
		},
	})
	return nil
}

func help(c *db.DBConfig, w wechat.ResponseWriter, r *wechat.Request) error {
	adm := ``
	ad, err := c.GetAdmin(r.FromUserName)
	if err == nil {
		adm = fmt.Sprintf(`
Hello %v, you can use following command:
adm:help
	Get Admin Helps
adm:test
	This is test
`, ad.Username)
	}
	w.ReplyText(`Welcome to <a href='` + db.Url + `'>GoEast Language Centers</a>, use the following keywords to get information:
Home
	Get Homepage
Help
	Get Helps
Event
	Get Events
Blog
	Get Blogs
Contact
	Get Contact Info
` + adm)
	return nil
}

type dua []time.Duration

func (d dua) Less(i, j int) bool {
	return d[i] < d[j]
}

func (d dua) Len() int {
	return len(d)
}
func (d dua) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

func getEventArticle(event *json.Json) (wechat.Article, time.Duration) {
	return wechat.Article{
		Url:         db.Url + event.Get("fullUrl").MustString(),
		Title:       event.Get("title").MustString(),
		PicUrl:      event.Get("assetUrl").MustString(),
		Description: event.Get("excerpt").MustString(),
	}, time.Duration(event.Get("startDate").MustInt64())
}

//Handle event request.
func handleEvent(c *db.DBConfig,
	w wechat.ResponseWriter,
	r *wechat.Request,
	check func(*time.Time) bool,
	errstr string) error {
	deferr := errors.New(errstr)
	msg, err := c.QueryMsg(db.Events)
	if err != nil {
		w.ReplyText(deferr.Error())
		return nil
	}
	j, err := json.NewJson([]byte(msg.Content))
	if err != nil {
		w.ReplyText(deferr.Error())
		return nil
	}
	mapa := map[time.Duration]wechat.Article{}
	past := j.Get("upcomming")
	for i, _ := range past.MustArray() {
		a, f := getEventArticle(past.GetIndex(i))
		tf := BaseTime.Add(f * time.Millisecond)
		if check(&tf) {
			mapa[f] = a
		}
	}
	if len(mapa) == 0 {
		w.ReplyText(deferr.Error())
		return nil
	}
	var keys dua
	for k, _ := range mapa {
		keys = append(keys, k)
	}
	sort.Sort(keys)
	res := []wechat.Article{}
	if len(keys) > db.MaxArticles {
		keys = keys[len(keys)-db.MaxArticles:]
	}
	n := len(keys)
	for i := 0; i < n/2; i++ {
		keys[i], keys[n-1-i] = keys[n-1-i], keys[i]
	}
	for _, k := range keys {
		res = append(res, mapa[k])
	}
	w.ReplyNews(res)
	return nil
}

//Upcoming event, only list 3 events.
func event(c *db.DBConfig, w wechat.ResponseWriter, r *wechat.Request) error {
	return handleEvent(c, w, r, func(t *time.Time) bool {
		return true
	}, `Oops! No event recently!`)
}

//Today's events.
func today(c *db.DBConfig, w wechat.ResponseWriter, r *wechat.Request) error {
	return handleEvent(c, w, r, func(t *time.Time) bool {
		return time.Now().Day() == t.Day()
	}, `Oops! No event today!`)
}

//This week's events.
func week(c *db.DBConfig, w wechat.ResponseWriter, r *wechat.Request) error {
	return handleEvent(c, w, r, func(t *time.Time) bool {
		return time.Now().Day() == t.Day()
	}, `Oops! No event this week!`)
}

//Recent Blogs.
func blog(c *db.DBConfig, w wechat.ResponseWriter, r *wechat.Request) error {
	deferr := errors.New(`Oops! No blog recently!`)
	msg, err := c.QueryMsg(db.Blog)
	if err != nil {
		w.ReplyText(deferr.Error())
		return nil
	}
	j, err := json.NewJson([]byte(msg.Content))
	if err != nil {
		w.ReplyText(deferr.Error())
		return nil
	}
	mapa := map[time.Duration]wechat.Article{}
	past := j.Get("items")
	for i, _ := range past.MustArray() {
		event := past.GetIndex(i)
		a := wechat.Article{
			Url:         db.Url + event.Get("fullUrl").MustString(),
			Title:       event.Get("title").MustString(),
			PicUrl:      event.Get("assetUrl").MustString(),
			Description: event.Get("excerpt").MustString(),
		}
		t := time.Duration(event.Get("addedOn").MustInt64())
		mapa[t] = a
	}
	if len(mapa) == 0 {
		w.ReplyText(deferr.Error())
		return nil
	}
	var keys dua
	for k, _ := range mapa {
		keys = append(keys, k)
	}
	sort.Sort(keys)
	res := []wechat.Article{}
	if len(keys) > db.MaxArticles {
		keys = keys[:db.MaxArticles]
	}
	for _, k := range keys {
		res = append(res, mapa[k])
	}
	w.ReplyNews(res)
	return nil
}

func (rs *routes) register(pattern string, handler RouteFunc) error {
	r, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	rs.rs = append(rs.rs, &dispatchRoute{regx: r, handler: handler})
	return nil
}

func DeamonTask(config *db.DBConfig) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	shutdown := false
	go func() {
		for !shutdown {
			_, err := config.WC.UpdateAccessToken()
			if err != nil {
				log.Fatalln(err)
			}
			time.Sleep(10 * time.Second)
		}
	}()
	go func() {
		<-c
		shutdown = true
	}()
}
