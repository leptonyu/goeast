package logic

import (
	"errors"
	"fmt"
	json "github.com/bitly/go-simplejson"
	"github.com/leptonyu/wechat"
	"io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	MaxArticles = 3
	Url         = "http://www.goeastmandarin.com"
	Events      = "/events"
	Blog        = "/blog"
)

var BaseTime = time.Date(1970, 1, 1, 0, 0, 0, 0, time.Local).Add(8 * time.Hour)

// this struct is used for caching the GoEast site.
// Then we can speed up the responds of WeChat requests.
// There will be some goroutines used for update the cache in period time.
type Msg struct {
	Name       string    // Key of msg, list at the const in this package.
	Content    string    // Content of msg, this content is formated as json.
	ExpireTime time.Time // create time of the content.
}

func queryMsg(key string, x *wechat.MongoStorage) (*Msg, error) {
	msg := &Msg{}
	err := x.Query(func(m *mgo.Database) error {
		er := m.C("web").Find(bson.M{"name": key}).One(&msg)
		log.Println(time.Since(msg.ExpireTime).Seconds())
		if er != nil || time.Since(msg.ExpireTime).Seconds() >= 0 {
			c := make(chan bool)
			close(c)
			go func() {
				defer func() {
					if x := recover(); x != nil {
						//				log.Println(x)
					}
				}()
				res, err := http.Get(Url + key + "?format=json-pretty")
				if err != nil {
					return
				}
				defer res.Body.Close()
				body, err := ioutil.ReadAll(res.Body)
				if err != nil {
					return
				}
				m := &Msg{}
				m.Name = key
				m.Content = string(body)
				m.ExpireTime = time.Now().Add(30 * time.Minute)
				msg = m
				go x.Query(func(m *mgo.Database) error {

					m.C("web").Upsert(bson.M{"name": key}, msg)
					log.Println("update compelete")
					return nil
				})
				c <- true
			}()
			select {
			case <-time.After(1 * time.Second):
				log.Println("Time out!")
			case <-c:
			}
		}
		return er
	})
	return msg, err
}
func Dispatch(x *wechat.MongoStorage) error {
	wc, err := x.GetWeChat()
	if err != nil {
		return err
	}
	wc.RegisterHandler(textDispatcher(x), wechat.MsgTypeText, wechat.MsgTypeEventClick)
	wc.RegisterHandler(func(w wechat.RespondWriter, r *wechat.Request) error {
		return help(x, w, "Help")
	}, wechat.MsgTypeEventSubscribe)

	wc.RegisterHandler(func(w wechat.RespondWriter, r *wechat.Request) error {
		return help(x, w, "Help")
	}, wechat.MsgTypeEventScan)
	return nil
}

type TextFunc func(*wechat.MongoStorage, wechat.RespondWriter, string) error

type TextRoute struct {
	regx   *regexp.Regexp
	handle TextFunc
}

type TextRoutes []*TextRoute

func (t *TextRoutes) Register(f TextFunc, pattern ...string) {
	for _, p := range pattern {
		r, err := regexp.Compile(p)
		if err != nil {
			panic(err)
		}
		*t = append(*t, &TextRoute{
			regx:   r,
			handle: f,
		})
	}
}

var (
	drs     = &TextRoutes{}
	nameReg = `(?i)^\s*(name\s|我叫|我是|名字\s|my name is\s|this is\s|it\'s\s|I am\s|I\'m\s)\s*(.+)\.?\s*$`
	dateReg = `^\s*([0-9]{8})\s*$`
)

func textDispatcher(config *wechat.MongoStorage) wechat.HandleFunc {
	drs.Register(contact, `^\s*(contact(\-us)?|联系方式)\s*$`)
	drs.Register(teacher, `^\s*(maria|emily|jane)\s*$`)
	drs.Register(name, nameReg)
	drs.Register(help, `^\s*(help|h|帮助)\s*$`)
	drs.Register(event, `^\s*(events?|事件|活动)\s*$`)
	drs.Register(blog, `^\s*(blogs?|日志|博客)\s*$`)
	drs.Register(next, `^\s*(next|下一个)\s*$`)
	drs.Register(week, `^\s*(week|本周|这周)\s*$`)
	drs.Register(home, `^\s*(home|主页|首页)\s*$`)
	drs.Register(course, `^\s*(courses?|课程)\s*$`)
	drs.Register(free, `^\s*(trial|free|试听)\s*$`)
	return func(w wechat.RespondWriter, r *wechat.Request) error {
		txt := strings.TrimSpace(r.Content + r.EventKey)
		sig := strings.ToLower(txt)
		for _, v := range *drs {
			if v.regx.MatchString(sig) {
				err := v.handle(config, w, txt)
				if err == nil {
					return nil
				}
			}
		}
		wel := `Thank you for your message, it has been received! You also can send an email to coursecenter@goeast.cn . 
		
Reply 'Help' to get helps.`
		name, ad, err := config.GetUserName(r.FromUserName)
		if err == nil {
			if ad == `0` {
				if err := admin(config, w, txt); err == nil {
					return nil
				}
			}
			w.ReplyText(`Hi ` + name + `, 
` + wel)
		} else {
			w.ReplyText(`Hi GoEaster,
` + wel + `

Could you please tell me your name? Just reply 'Name'+Space+'Your Name' to me, Thank you!

e.g.  Name Daniel`)
		}
		return nil
	}
}

func admin(c *wechat.MongoStorage, w wechat.RespondWriter, txt string) error {
	reg, _ := regexp.Compile(dateReg)
	if reg.MatchString(txt) {
		d, err := time.Parse("20060102", strings.TrimSpace(txt))
		if err == nil {
			d = d.Add(24 * time.Hour)
			return c.Query(func(m *mgo.Database) error {
				msg := SendText(d.Year(), d.Month(), d.Day(), m)
				if msg == "" {
					msg = "No message today!"
				}
				w.ReplyText(msg)
				return nil
			})

		}
	}
	return errors.New("")
}
func name(c *wechat.MongoStorage, w wechat.RespondWriter, txt string) error {
	reg, _ := regexp.Compile(nameReg)
	str := reg.FindStringSubmatch(txt)
	log.Println(str)
	if len(str) >= 3 && str[2] != "" {
		go c.SetUserName(w.FromUserId(), str[2], `1`)
		w.ReplyText(`Hello ` + str[2] + `, thank you for your information.`)
		return nil
	} else {
		return errors.New("")
	}
}

func home(c *wechat.MongoStorage, w wechat.RespondWriter, txt string) error {
	w.ReplyNews([]wechat.Article{
		wechat.Article{
			Title:       "GoEast Language Center",
			PicUrl:      "http://static.squarespace.com/static/52141adee4b0476aaa6af594/t/52e83ec9e4b0dfc630da5953/1390952403577/XJL_8779-small.jpg",
			Url:         Url,
			Description: `GoEast Language center is one of Shanghai's premiere Mandarin language schools. We offer tutoring services, online classes, and on-site classes at our beautiful Shanghai campus. Our mission is to teach Mandarin and Chinese culture in a fun, personal, and effective environment. At GoEast we focus on helping you achieve your learning goals with as much personal attention as possible and no self help gimmicks! Learning language is best done with others, and we hope you'll join us at GoEast! `,
		},
	})
	return nil
}
func help(c *wechat.MongoStorage, w wechat.RespondWriter, txt string) error {
	w.ReplyText(`Welcome to <a href='` + Url +
		`'>GoEast Language Center</a>, please use the following keywords to get further information:` +
		"\n\nHome/Blog\n  Get homepage or recent blogs." +
		"\n\nEvent/Next/Week\n  Get information of our recent events" +
		"\n\nCourse/Trial\n  Get information of our courses" +
		"\n\nContact\nEmily/Maria/Jane\n  Get contact infomation of us." +
		"\n\nIf you have any question about us, please send an email to coursecenter@goeast.cn")
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
	s, e := time.Duration(event.Get("startDate").MustInt64()), time.Duration(event.Get("endDate").MustInt64())
	ts, te := BaseTime.Add(s*time.Millisecond), BaseTime.Add(e*time.Millisecond)
	desp := ts.Format(`Monday, Jan _2, 2006
  03:04pm`) + " - " + te.Format(`03:04pm`)
	return wechat.Article{
		Url:         Url + event.Get("fullUrl").MustString(),
		Title:       event.Get("title").MustString(),
		PicUrl:      event.Get("assetUrl").MustString(),
		Description: desp,
	}, s
}

//Handle event request.
func handleEvent(c *wechat.MongoStorage,
	w wechat.RespondWriter,
	txt string,
	check func(time.Time) bool,
	errstr string, max int) error {
	deferr := errors.New(errstr)
	msg, err := queryMsg(Events, c)
	if err != nil {
		log.Println(err)
		w.ReplyText(deferr.Error())
		return nil
	}
	j, err := json.NewJson([]byte(msg.Content))
	if err != nil {
		log.Println(err)
		w.ReplyText(deferr.Error())
		return nil
	}
	mapa := map[time.Duration]wechat.Article{}
	past := j.Get("upcoming")
	for i, _ := range past.MustArray() {
		a, f := getEventArticle(past.GetIndex(i))
		if check(BaseTime.Add(f * time.Millisecond)) {
			mapa[f] = a
		}
	}
	if len(mapa) == 0 {
		log.Println(err)
		w.ReplyText(deferr.Error())
		return nil
	}
	var keys dua
	for k, _ := range mapa {
		keys = append(keys, k)
	}
	sort.Sort(keys)
	res := []wechat.Article{}
	if len(keys) > max {
		keys = keys[:max]
	}
	for _, k := range keys {
		res = append(res, mapa[k])
	}
	w.ReplyNews(res)
	return nil
}

//Upcoming event, only list 3 events.
func event(c *wechat.MongoStorage, w wechat.RespondWriter, txt string) error {
	return handleEvent(c, w, txt, func(t time.Time) bool {
		return true
	}, `Oops! No event recently!`, MaxArticles)
}

//Next event.
func next(c *wechat.MongoStorage, w wechat.RespondWriter, txt string) error {
	return handleEvent(c, w, txt, func(t time.Time) bool {
		return true
	}, `Oops! No next event!`, 1)
}

//This week's events.
func week(c *wechat.MongoStorage, w wechat.RespondWriter, txt string) error {
	return handleEvent(c, w, txt, func(t time.Time) bool {
		return time.Since(t).Hours() >= -7*24
	}, `Oops! No event in next 7 days!`, MaxArticles)
}

//Recent Blogs.
func blog(c *wechat.MongoStorage, w wechat.RespondWriter, txt string) error {
	deferr := errors.New(`Oops! No blog recently!`)
	msg, err := queryMsg(Blog, c)
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
			Url:         Url + event.Get("fullUrl").MustString(),
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
	if len(keys) > MaxArticles {
		keys = keys[len(keys)-MaxArticles:]
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

var (
	maria = teacherinfo{
		Id:      "maria",
		Name:    "Maria MAO",
		Chinese: "毛瑞",
		Email:   "mariamao@goeast.cn",
		Skype:   "maoruimaria",
	}
	emily = teacherinfo{
		Id:      "emily",
		Name:    "Emily WANG",
		Chinese: "王蓉",
		Email:   "emilywang@goeast.cn",
		Skype:   "rongni_123",
	}
	jane = teacherinfo{
		Id:      "jane",
		Name:    "Jane LUO",
		Chinese: "罗琼",
		Email:   "janeluo@goeast.cn",
		Skype:   "jane.qiongluo",
	}
)

type teacherinfo struct {
	Id      string
	Name    string
	Chinese string
	Phone   string
	Email   string
	Skype   string
}

func teacher(c *wechat.MongoStorage, w wechat.RespondWriter, txt string) error {
	switch strings.ToLower(txt) {
	case maria.Id:
		return t(w, maria)
	case emily.Id:
		return t(w, emily)
	case jane.Id:
		return t(w, jane)
	default:
		return errors.New("Not found!")
	}
}
func t(w wechat.RespondWriter, tt teacherinfo) error {
	w.ReplyText(fmt.Sprintf(`%v (%v)
Teacher & Consultant
GoEast Language Center

Skype:
  %v

Email: 
%v`, tt.Name, tt.Chinese, tt.Skype, tt.Email))
	return nil
}

func contact(c *wechat.MongoStorage, w wechat.RespondWriter, txt string) error {
	w.ReplyText(`GoEast Language Center

No 194-196 Zhengmin Road, Yangpu District, Shanghai, China

Telephone: 
  86-21-31326611  

Email: 
 coursecenter@goeast.cn`)
	return nil
}

var (
	trial = wechat.Article{
		Title:       `Free Trial`,
		Url:         Url + `/free-trials`,
		PicUrl:      `http://static.squarespace.com/static/52141adee4b0476aaa6af594/t/53026a93e4b01cb1fb30d170/1392667284963/XJL_8799hero.jpg?format=1500w`,
		Description: `Feel free to sign up for our free online trial classes or 1 hr free one on one tutoring sessions. Sign up now to reserve your spot, before they run out! NOTE: Free trial courses all take place on Fridays, and free tutoring trials take place on Wednesdays and Saturdays.`,
	}
	o2o = wechat.Article{
		Title:  `1 On 1 Tutoring`,
		Url:    Url + `/1-on-1-tutoring`,
		PicUrl: `http://static.squarespace.com/static/52141adee4b0476aaa6af594/t/52f5334ce4b087ee0881b625/1391801166804/Lindasmall.jpg?format=1500w`,
	}
	online = wechat.Article{
		Title:       `Online Courses`,
		Url:         Url + `/online-courses`,
		PicUrl:      `http://static.squarespace.com/static/52141adee4b0476aaa6af594/t/52e853e6e4b00b97daff0302/1390957544785/XJL_1224small.jpg?format=1500w`,
		Description: `Learning Mandarin online has never been easier than with GoEast. We offer a wide array of Mandarin language and cultural courses to suit everyone's needs and capabilities. GoEast courses are, as always, taught live by a professional instructor with years of experience and a passion for language and learning.`,
	}

	onsite = wechat.Article{
		Title:       `On Site Courses`,
		Url:         Url + `/on-site-courses`,
		PicUrl:      `http://static.squarespace.com/static/52141adee4b0476aaa6af594/t/52f95ea7e4b0116de8c3582e/1392074410688/XJL_8918-small.jpg?format=1500w`,
		Description: `You can't replicate the experience of learning a new language with real people in a collaborative environment. For this reason GoEast regularly runs a series of courses that take place at our Shanghai campus.  Whether you need a crash course in Mandarin for business, obtain HSK4 certification, or simply would like to learn some travel Mandarin, we have a course for you. GoEast also offers cultural events if you'd like to meet some people, or casually learn about Chinese culture.`,
	}
)

func course(c *wechat.MongoStorage, w wechat.RespondWriter, txt string) error {
	w.ReplyNews([]wechat.Article{
		o2o, online, onsite, trial,
	})
	return nil
}

func free(c *wechat.MongoStorage, w wechat.RespondWriter, txt string) error {
	w.ReplyNews([]wechat.Article{
		trial,
	})
	return nil
}
