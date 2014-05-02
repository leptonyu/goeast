package wechat

import (
	"crypto/sha1"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"sort"
)

const (
	msgEvent = "event"
	// Event Type
	EventSubscribe   = "subscribe"
	EventUnsubscribe = "unsubscribe"
	EventScan        = "scan"
	EventClick       = "CLICK"
	// Message type
	MsgTypeDefault          = ".*"
	MsgTypeText             = "text"
	MsgTypeImage            = "image"
	MsgTypeVoice            = "voice"
	MsgTypeVideo            = "video"
	MsgTypeLocation         = "location"
	MsgTypeLink             = "link"
	MsgTypeEvent            = msgEvent + ".*"
	MsgTypeEventSubscribe   = msgEvent + "\\." + EventSubscribe
	MsgTypeEventUnsubscribe = msgEvent + "\\." + EventUnsubscribe
	MsgTypeEventScan        = msgEvent + "\\." + EventScan
	MsgTypeEventClick       = msgEvent + "\\." + EventClick
	// Media type
	MediaTypeImage = "image"
	MediaTypeVoice = "voice"
	MediaTypeVideo = "video"
	MediaTypeThumb = "thumb"
	// Button type
	MenuButtonTypeKey = "click"
	MenuButtonTypeUrl = "view"
)

type route struct {
	regex   *regexp.Regexp
	handler HandlerFunc
}

type HandlerFunc func(ResponseWriter, *Request) error

//WeChat Request
type Request struct {
	ToUserName   string
	FromUserName string
	CreateTime   int
	MsgType      string
	MsgId        int64
	Content      string  `json:",omitempty"`
	PicUrl       string  `json:",omitempty"`
	MediaId      string  `json:",omitempty"`
	Format       string  `json:",omitempty"`
	ThumbMediaId string  `json:",omitempty"`
	LocationX    float32 `json:"Location_X,omitempty",xml:"Location_X"`
	LocationY    float32 `json:"Location_Y,omitempty",xml:"Location_Y"`
	Scale        float32 `json:",omitempty"`
	Label        string  `json:",omitempty"`
	Title        string  `json:",omitempty"`
	Description  string  `json:",omitempty"`
	Url          string  `json:",omitempty"`
	Event        string  `json:",omitempty"`
	EventKey     string  `json:",omitempty"`
	Ticket       string  `json:",omitempty"`
	Latitude     float32 `json:",omitempty"`
	Longitude    float32 `json:",omitempty"`
	Precision    float32 `json:",omitempty"`
	Recognition  string  `json:",omitempty"`
}

//Http Respond, Main method
func (wc *WeChat) HttpHandle(w http.ResponseWriter, r *http.Request) {
	if !checkSignature(wc.token, w, r) {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}
	// Verify request
	if r.Method == "GET" {
		fmt.Fprintf(w, r.FormValue("echostr"))
		return
	}
	// Process message
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("WeChat receive message failed:", err)
		http.Error(w, "", http.StatusBadRequest)
	} else {
		var msg Request
		if err := xml.Unmarshal(data, &msg); err != nil {
			log.Println("WeChat parse message failed:", err)
			http.Error(w, "", http.StatusBadRequest)
		} else {
			requestPath := msg.MsgType
			if requestPath == msgEvent {
				requestPath += "." + msg.Event
			}
			for _, route := range wc.routes {
				if !route.regex.MatchString(requestPath) {
					continue
				}
				route.handler(&responseWriterData{
					WC:           wc,
					W:            w,
					toUserName:   msg.FromUserName,
					fromUserName: msg.ToUserName,
				}, &msg)
				return
			}
			http.Error(w, "", http.StatusNotFound)
		}
	}
}

// Create handler func
func (wc *WeChat) CreateHandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wc.HttpHandle(w, r)
	}
}

//Check signature of wechat server
// token is given by WeChat
func checkSignature(token string, w http.ResponseWriter, r *http.Request) bool {
	r.ParseForm()
	var signature string = r.FormValue("signature")
	var timestamp string = r.FormValue("timestamp")
	var nonce string = r.FormValue("nonce")
	strs := sort.StringSlice{token, timestamp, nonce}
	sort.Strings(strs)
	var str string
	for _, s := range strs {
		str += s
	}
	h := sha1.New()
	h.Write([]byte(str))
	return fmt.Sprintf("%x", h.Sum(nil)) == signature
}

// Register request callback.
func (wc *WeChat) HandleFunc(pattern string, handler HandlerFunc) {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		panic(err)
		return
	}
	route := &route{regex, handler}
	wc.routes = append(wc.routes, route)
}
