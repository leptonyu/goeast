package main

import (
	"encoding/csv"
	"flag"
	"github.com/leptonyu/goeast/logic"
	"github.com/leptonyu/wechat"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"os"
)

func main() {
	date := flag.String("file", "conf/user.list", "LogDate")
	flag.Parse()
	f, err := os.Open(*date)
	if err != nil {
		log.Println(err)
		return
	}
	defer f.Close()
	reader := csv.NewReader(f)
	xx, err := reader.ReadAll()
	if err != nil {
		log.Println(err)
		return
	}
	x := wechat.NewLocalMongo("api")
	x.Query(func(db *mgo.Database) error {
		user := db.C("user")
		for _, line := range xx {
			user.Upsert(bson.M{"id": line[0]}, logic.User{
				Id:    line[0],
				Name:  line[1],
				Admin: line[2],
			})
		}
		return nil
	})
}
