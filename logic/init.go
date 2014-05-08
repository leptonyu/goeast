package logic

import (
	"github.com/leptonyu/wechat"
	"github.com/leptonyu/wechat/db"
)

func Init(x *db.MongoStorage) error {
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
	return wc.CreateMenu(menu)
}
