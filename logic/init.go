package logic

import (
	"github.com/kylelemons/go-gypsy/yaml"
	"github.com/leptonyu/wechat"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

type User struct {
	Id    string
	Name  string
	Admin string
}

func Init(x *wechat.MongoStorage, y *yaml.File) error {
	menu := &wechat.Menu{}
	menu.Buttons = []wechat.MenuButton{
		wechat.MenuButton{
			Name: "Site",
			SubButtons: []wechat.MenuButton{
				wechat.MenuButton{
					Name: "Home",
					Type: "click",
					Key:  "Home",
				},
				wechat.MenuButton{
					Name: `       Free Trial 》`,
					Type: "view",
					Url:  Url + "/free-trials",
				},
				wechat.MenuButton{
					Name: ` 1 On 1 Tutoring 》`,
					Type: "view",
					Url:  Url + "/1-on-1-tutoring",
				},
				wechat.MenuButton{
					Name: ` Online Courses 》`,
					Type: "view",
					Url:  Url + "/online-courses",
				},
				wechat.MenuButton{
					Name: `On Site Courses 》`,
					Type: "view",
					Url:  Url + "/on-site-courses",
				},
			},
		},
		wechat.MenuButton{
			Name: "Event",
			SubButtons: []wechat.MenuButton{
				wechat.MenuButton{
					Name: "Next Event",
					Type: "click",
					Key:  "Next",
				},
				wechat.MenuButton{
					Name: "Coming Week",
					Type: "click",
					Key:  "Week",
				},
				wechat.MenuButton{
					Name: "Recent Events",
					Type: "click",
					Key:  "Event",
				},
				wechat.MenuButton{
					Name: "Recent Blogs",
					Type: "click",
					Key:  "Blog",
				},
				wechat.MenuButton{
					Name: "Help",
					Type: "click",
					Key:  "Help",
				},
			},
		},

		wechat.MenuButton{
			Name: "Contact",
			SubButtons: []wechat.MenuButton{
				wechat.MenuButton{
					Name: "Contact Us",
					Type: "click",
					Key:  "contact-us",
				},
				wechat.MenuButton{
					Name: "Emily",
					Type: "click",
					Key:  "Emily",
				},
				wechat.MenuButton{
					Name: "Maria",
					Type: "click",
					Key:  "Maria",
				},
				wechat.MenuButton{
					Name: "Jane",
					Type: "click",
					Key:  "Jane",
				},
				wechat.MenuButton{
					Name: `Teachers 》`,
					Type: "view",
					Url:  Url + "/our-teachers",
				},
			},
		},
	}
	wc, err := x.GetWeChat()
	if err != nil {
		return err
	}
	if err := wc.CreateMenu(menu); err != nil {
		return err
	}

	return x.Query(func(d *mgo.Database) error {
		_, err := d.C("wechat").Upsert(bson.M{"name": "mail"}, Mail{
			Name:     "mail",
			User:     y.Require("muser"),
			Password: y.Require("mpassword"),
			Host:     y.Require("mhost"),
			To:       y.Require("mto"),
		})
		return err
	})
}
