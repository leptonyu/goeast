package logic

import (
	"crypto/tls"
	"github.com/leptonyu/wechat"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"net"
	"net/smtp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Mail struct {
	Name     string
	User     string
	Password string
	Host     string
	To       string
}

func SendText(year int, month time.Month, day int, d *mgo.Database) string {
	t := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
	b, a := int(t.Add(-24*time.Hour).Unix()), int(t.Unix())
	res := []*wechat.Request{}
	err := d.C("request").Find(bson.M{
		"tousername": "gh_cd1fee56bc3a",
		"msgtype":    "text",
		"createtime": bson.M{"$lt": a, "$gte": b},
	}).All(&res)
	uu := d.C("user")
	if err == nil && len(res) > 0 {
		nres := map[string]xxxx{}
		for _, v := range res {
			flag := true
			/*			for _, d := range *drs {
							if d.regx.MatchString(v.Content) {
								flag = false
								break
							}
						}
			*/if flag {
				xx, ok := nres[v.FromUserName]

				if !ok {
					xx = xxxx{}
				}
				xx = append(xx, v)
				nres[v.FromUserName] = xx

			}
		}
		if len(nres) > 0 {
			message := "Messages:\n"
			for k, v := range nres {
				ux := User{}
				if err := uu.Find(bson.M{"id": k}).One(&ux); err == nil {
					k = k + "(" + ux.Name + ")"
				}
				message += "\n  " + k + "\n"
				sort.Sort(v)
				for _, r := range v {
					message += "    " +
						time.Unix(int64(r.CreateTime), 0).Format("15:04:05") +
						" " + r.Content + "\n"
				}
			}
			return message
		}
	}
	return ""
}
func SendMail(year int, month time.Month, day int, dm *mgo.Database, tomail string) {
	message := SendText(year, month, day, dm)
	if message == "" {
		return
	}
	mail := Mail{}
	dm.C("wechat").Find(bson.M{"name": "mail"}).One(&mail)
	date := time.Date(year, month, day, 0, 0, 0, 0, time.Local).Add(-24 * time.Hour).Format("2006-01-02")
	if tomail != "" {
		mail.To = tomail
	}
	mailList := strings.Split(mail.To, ",")
	hh := strings.Split(mail.Host, ":")
	auth := smtp.PlainAuth(
		"",
		mail.User,
		mail.Password,
		hh[0],
	)
	msg := []byte("To: " + mailList[0] +
		"\r\nFrom: GoEast WeChat<" + mail.User + ">\r\nSubject: " + date +
		"\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n" + message)
	log.Println(string(msg))
	err := SendMailUsingTLS(mail.Host, auth, mail.User, mailList, msg)
	//err := smtp.SendMail(mail.Host, auth, mail.User, mailList, msg)
	if err != nil {
		log.Println(err)
	}
}
func Spy(x *wechat.MongoStorage) error {

	go func() {
		for {
			x.Query(func(d *mgo.Database) error {
				y, m, dd := time.Now().Add(-8 * time.Hour).Date()
				date := strconv.Itoa(y) + "-" + m.String() + "-" + strconv.Itoa(dd)
				c, err := d.C("request_date").Find(bson.M{"date": date}).Count()
				if err != nil || c == 0 {
					y, m, dd = time.Now().Date()
					SendMail(y, m, dd, d, "")
					d.C("request_date").Insert(bson.M{"date": date})
				}
				return nil
			})
			time.Sleep(5 * time.Minute)
		}
	}()
	return nil
}

type xxxx []*wechat.Request

func (x xxxx) Len() int           { return len(x) }
func (x xxxx) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
func (x xxxx) Less(i, j int) bool { return x[i].CreateTime < x[j].CreateTime }

//参考net/smtp的func SendMail()
//使用net.Dial连接tls(ssl)端口时,smtp.NewClient()会卡住且不提示err
//len(to)>1时,to[1]开始提示是密送
func SendMailUsingTLS(addr string, auth smtp.Auth, from string,
	to []string, msg []byte) (err error) {

	//create smtp client
	c, err := Dial(addr)
	if err != nil {
		log.Println("Create smpt client error:", err)
		return err
	}
	defer c.Close()

	if auth != nil {
		if ok, _ := c.Extension("AUTH"); ok {
			if err = c.Auth(auth); err != nil {
				log.Println("Error during AUTH", err)
				return err
			}
		}
	}

	if err = c.Mail(from); err != nil {
		return err
	}

	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	_, err = w.Write(msg)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return c.Quit()
}

//return a smtp client
func Dial(addr string) (*smtp.Client, error) {
	conn, err := tls.Dial("tcp", addr, nil)
	if err != nil {
		log.Println("Dialing Error:", err)
		return nil, err
	}
	//分解主机端口字符串
	host, _, _ := net.SplitHostPort(addr)
	return smtp.NewClient(conn, host)
}
