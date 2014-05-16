package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/xml"
	"fmt"
	"github.com/kylelemons/go-gypsy/yaml"
	"github.com/leptonyu/wechat"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"
)

func main() {
	config, err := yaml.ReadFile("conf/conf.yaml")
	if err != nil {
		log.Fatalf(`Configuration File: %s`, err)
	}
	abc := wechat.Request{}
	abc.FromUserName = `aaaaa`
	abc.ToUserName = `bbbbb`
	abc.CreateTime = int(time.Now().Unix())
	abc.MsgId = 12333
	abc.MsgType = `text`
	abc.Content = `blog`
	a, _ := xml.Marshal(abc)
	//log.Println(string(a))
	url := getUrl(config.Require("token"))
	//log.Println(url)
	r, err := http.Post(url,
		"application/xml", bytes.NewReader(a))
	if err != nil {
		log.Println(err)
		return
	}
	defer r.Body.Close()
	x, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(string(x))
	res := wechat.Request{}
	err = xml.Unmarshal(x, &res)

	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println(res.Content)
}

func getUrl(token string) string {
	timestamp := strconv.Itoa(int(time.Now().Unix()))
	nonce := `ThisNothing`
	strs := sort.StringSlice{token, timestamp, nonce}
	sort.Strings(strs)
	var str string
	for _, s := range strs {
		str += s
	}
	h := sha1.New()
	h.Write([]byte(str))
	sig := fmt.Sprintf("%x", h.Sum(nil))
	//log.Println(sig)
	return fmt.Sprintf(`http://localhost/api?signature=%s&timestamp=%s&nonce=%s&echostr=123456`, sig, timestamp, nonce)
}
