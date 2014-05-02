package db

import (
	"github.com/leptonyu/goeast/wechat"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"testing"
	"time"
)

type dd struct {
	Name  string
	Value string
}

// Test Basic Query of MongoDB
func TestA(t *testing.T) {
	abc := NewDBConfig("test")
	d, err := abc.Query(func(dbm *mgo.Database) (data interface{}, err error) {
		c := dbm.C("userinfo")
		err = c.Insert(&dd{Name: "Maria", Value: "Hello"})
		if err != nil {
			return
		}
		m := bson.M{"name": "Maria"}
		data = &dd{}
		err = c.Find(m).One(data)
		if err != nil {
			return
		}
		c.RemoveAll(m)
		return
	})
	if err != nil {
		t.Error(err)
	} else {
		t.Log(d)
	}
}

func TestB(t *testing.T) {
	abc := NewDBConfig("test")
	x := &wechat.AccessToken{
		Token:      "this is token",
		ExpireTime: time.Now(),
	}
	err := abc.Write(x)
	if err != nil {
		t.Error(err)
	}
	y, err := abc.Read()
	if err != nil {
		t.Error(err)
	}
	if x.Token != y.Token {
		t.Error("Token is not same!")
	}
	t.Log(x)
	t.Log(y)
}
