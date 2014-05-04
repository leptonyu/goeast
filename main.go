package main

import (
	"encoding/csv"
	"flag"
	"github.com/leptonyu/goeast/db"
	"github.com/leptonyu/goeast/util"
	"github.com/leptonyu/goeast/wechat"
	"log"
	"os"
)

func main() {
	port := flag.Int("port", 8080, "Web service port")
	api := flag.String("api", "api", "http://localhost/$api,\n	Also use as database name with prefix wechat_")
	appid := flag.String("appid", "", "App id")
	secret := flag.String("secret", "", "App secret")
	token := flag.String("token", "", "Token")
	init := flag.Bool("init", false, "Init ")
	help := flag.Bool("h", false, "Help")
	admin := flag.Bool("admin", false, "Add administrator")
	flag.Parse()
	if *help {
		flag.PrintDefaults()
		os.Exit(0)
	}
	config, err := db.NewDBConfig(*api)
	if *init {
		err := config.Init(*appid, *secret, *token)
		if err != nil {
			log.Panicln(err)
			os.Exit(1)
		}
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
						Name: `          Teachers 》`,
						Type: "view",
						Url:  db.Url + db.Teachers,
					},
					wechat.MenuButton{
						Name: ` 1 On 1 Tutoring 》`,
						Type: "view",
						Url:  db.Url + db.One2one,
					},
					wechat.MenuButton{
						Name: ` Online Courses 》`,
						Type: "view",
						Url:  db.Url + db.Online,
					},
					wechat.MenuButton{
						Name: `On Site Courses 》`,
						Type: "view",
						Url:  db.Url + db.Onsite,
					},
				},
			},
			wechat.MenuButton{
				Name: "Event",
				SubButtons: []wechat.MenuButton{
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
						Name: "Contect Us",
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
						Name: `Free Trial 》`,
						Type: "view",
						Url:  db.Url + "/free-trials",
					},
				},
			},
		}
		config.WC.CreateMenu(menu)
		func(keys ...string) {
			for _, key := range keys {
				log.Println("Fetching web " + key)
				config.QueryMsg(key)
			}
		}(
			db.Blog,
			db.Events,
			db.Home,
			db.Campus,
			db.Contact,
			db.Galleries,
			db.One2one,
			db.Online,
			db.Onsite,
			db.Teachers,
			db.Testimonials,
		)
		os.Exit(0)
	}
	if *admin {
		file, err := os.Open("tool/admin.list")
		if err != nil {
			log.Panic(err)
			os.Exit(1)
		}
		defer file.Close()
		all, err := csv.NewReader(file).ReadAll()
		if err != nil {
			log.Panic(err)
			os.Exit(1)
		}
		for _, value := range all {
			log.Println("Add administrator", value[1])
			config.UpsertWithUser(value[0], value[1], true)
		}
		os.Exit(0)
	}

	if err != nil {
		log.Panic(err)
		os.Exit(1)
	}

	util.StartWeb(*port, config)
}
