package wechat

import (
	"crypto/sha1"
	"encoding/json"
	"encoding/xml"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
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
	// Weixin host URL
	weixinHost        = "https://api.weixin.qq.com/cgi-bin"
	weixinQRScene     = "https://api.weixin.qq.com/cgi-bin/qrcode"
	weixinShowQRScene = "https://mp.weixin.qq.com/cgi-bin/showqrcode"
	weixinFileURL     = "http://file.api.weixin.qq.com/cgi-bin/media"
	// Max retry count
	retryMaxN = 3
	// Reply format
	replyText    = `<xml>%s<MsgType><![CDATA[text]]></MsgType><Content><![CDATA[%s]]></Content></xml>`
	replyImage   = `<xml>%s<MsgType><![CDATA[image]]></MsgType><Image><MediaId><![CDATA[%s]]></MediaId></Image></xml>`
	replyVoice   = `<xml>%s<MsgType><![CDATA[voice]]></MsgType><Voice><MediaId><![CDATA[%s]]></MediaId></Voice></xml>`
	replyVideo   = `<xml>%s<MsgType><![CDATA[video]]></MsgType><Video><MediaId><![CDATA[%s]]></MediaId><Title><![CDATA[%s]]></Title><Description><![CDATA[%s]]></Description></Video></xml>`
	replyMusic   = `<xml>%s<MsgType><![CDATA[music]]></MsgType><Music><Title><![CDATA[%s]]></Title><Description><![CDATA[%s]]></Description><MusicUrl><![CDATA[%s]]></MusicUrl><HQMusicUrl><![CDATA[%s]]></HQMusicUrl><ThumbMediaId><![CDATA[%s]]></ThumbMediaId></Music></xml>`
	replyNews    = `<xml>%s<MsgType><![CDATA[news]]></MsgType><ArticleCount>%d</ArticleCount><Articles>%s</Articles></xml>`
	replyHeader  = `<ToUserName><![CDATA[%s]]></ToUserName><FromUserName><![CDATA[%s]]></FromUserName><CreateTime>%d</CreateTime>`
	replyArticle = `<item><Title><![CDATA[%s]]></Title> <Description><![CDATA[%s]]></Description><PicUrl><![CDATA[%s]]></PicUrl><Url><![CDATA[%s]]></Url></item>`
	// QR scene request
	requestQRScene      = `{"expire_seconds":%d,"action_name":"QR_SCENE","action_info":{"scene":{"scene_id":%d}}}`
	requestQRLimitScene = `{"action_name":"QR_LIMIT_SCENE","action_info":{"scene":{"scene_id":%d}}}`
)

//WeChat Request
type Request struct {
	ToUserName   string
	FromUserName string
	CreateTime   time.Time
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

//Access Token struct
type AccessToken struct {
	accesstoken string    //WeChat access token
	expire      time.Time //expire time
}

//Method to store and create access token
type AccessTokenReaderWriter interface {
	//Read access token
	Read() (*AccessToken, error)
	//Write access token
	Write(*AccessToken) error
}

//WeChat Struct
type WeChat struct {
	appid  string
	secret string
	token  string
	atrw   AccessTokenReaderWriter
	routes []*route
}

// Use to output reply
type ResponseWriter interface {
	// Get weixin
	GetWeChat() *WeChat
	GetUserData() interface{}
	// Reply message
	ReplyText(text string)
	ReplyImage(mediaId string)
	ReplyVoice(mediaId string)
	ReplyVideo(mediaId string, title string, description string)
	ReplyMusic(music *Music)
	ReplyNews(articles []Article)
	// Post message
	PostText(text string) error
	PostImage(mediaId string) error
	PostVoice(mediaId string) error
	PostVideo(mediaId string, title string, description string) error
	PostMusic(music *Music) error
	PostNews(articles []Article) error
	// Media operator
	UploadMediaFromFile(mediaType string, filepath string) (string, error)
	DownloadMediaToFile(mediaId string, filepath string) error
	UploadMedia(mediaType string, filename string, reader io.Reader) (string, error)
	DownloadMedia(mediaId string, writer io.Writer) error
}

// Post text message
func (wc *WeChat) PostText(touser string, text string) error {
	var msg struct {
		ToUser  string `json:"touser"`
		MsgType string `json:"msgtype"`
		Text    struct {
			Content string `json:"content"`
		} `json:"text"`
	}
	msg.ToUser = touser
	msg.MsgType = "text"
	msg.Text.Content = text
	return postMessage(wx.tokenChan, &msg)
}

// Callback function
type HandlerFunc func(ResponseWriter, *Request)

type route struct {
	regex   *regexp.Regexp
	handler HandlerFunc
}

//Create WeChat Object
func New(appid, secret, token string, atrw AccessTokenReaderWriter) (wc *WeChat, err error) {
	wc = &WeChat{}
	wc.appid = appid
	wc.secret = secret
	wc.token = token
	wc.atrw = atrw
	_, err = wc.getAccessToken()
	if err == nil {
		log.Logger.Println("Get token")
	}
	return
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

func (wc *WeChat) getAccessToken() (token string, err error) {
	for i := 1; i <= retryMaxN; i++ {
		t, err := wc.atrw.Read()
		if err == nil {
			if time.Since(t.expire).Seconds() >= 0 {
				t, err = authAccessToken(wc.appid, wc.secret)
				if err != nil {
					continue
				} else {
					err = wc.atrw.Write(t)
					if err != nil {
						log.Fatal(err)
					}
				}
			}
			token = t.accesstoken
			return
		}
	}
	err = errors.New("Cannot get Access Token")
	return
}

//Get accesstoken from wechat.
func authAccessToken(appid string, secret string) (token *AccessToken, err error) {
	resp, err := http.Get(weixinHost + "/token?grant_type=client_credential&appid=" + appid + "&secret=" + secret)
	if err != nil {
		log.Println("Get access token failed: ", err)
		return
	} else {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("Read access token failed: ", err)
			return
		} else {
			var res struct {
				AccessToken string `json:"access_token"`
				ExpiresIn   int64  `json:"expires_in"`
			}
			if err := json.Unmarshal(body, &res); err != nil {
				log.Println("Parse access token failed: ", err)
				return
			} else {
				//log.Printf("AuthAccessToken token=%s expires_in=%d", res.AccessToken, res.ExpiresIn)
				token = &AccessToken{
					accesstoken: res.AccessToken,
					expire:      time.Now().Add(time.Duration((res.ExpiresIn - 100) * 1000 * 1000 * 1000)),
				}
				return
			}
		}
	}
}

//Http Respond, Main method
func (wc *WeChat) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
				writer := responseWriter{}
				writer.wc = wc
				writer.writer = w
				writer.toUserName = msg.FromUserName
				writer.fromUserName = msg.ToUserName
				route.handler(writer, &msg)
				return
			}
			http.Error(w, "", http.StatusNotFound)
		}
	}
}

type responseWriter struct {
	wc           *WeChat
	writer       http.ResponseWriter
	toUserName   string
	fromUserName string
}

//Check signature of wechat server
// t
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

func postMessage(c chan accessToken, msg interface{}) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = postRequest(weixinHost+"/message/custom/send?access_token=", c, data)
	return err
}
