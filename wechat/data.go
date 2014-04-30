package wechat

import (
	"encoding/xml"
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

func authAccessToken(appid string, secret string) (string, time.Duration) {
	resp, err := http.Get(weixinHost + "/token?grant_type=client_credential&appid=" + appid + "&secret=" + secret)
	if err != nil {
		log.Println("Get access token failed: ", err)
	} else {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("Read access token failed: ", err)
		} else {
			var res struct {
				AccessToken string `json:"access_token"`
				ExpiresIn   int64  `json:"expires_in"`
			}
			if err := json.Unmarshal(body, &res); err != nil {
				log.Println("Parse access token failed: ", err)
			} else {
				//log.Printf("AuthAccessToken token=%s expires_in=%d", res.AccessToken, res.ExpiresIn)
				return res.AccessToken, time.Duration(res.ExpiresIn * 1000 * 1000 * 1000)
			}
		}
	}
	return "", 0
}

func createAccessToken(c chan accessToken, appid string, secret string) {
	token := accessToken{"", time.Now()}
	c <- token
	for {
		if time.Since(token.expires).Seconds() >= 0 {
			var expires time.Duration
			token.token, expires = authAccessToken(appid, secret)
			token.expires = time.Now().Add(expires)
		}
		c <- token
	}
}

func checkSignature(t string, w http.ResponseWriter, r *http.Request) bool {
	r.ParseForm()
	var signature string = r.FormValue("signature")
	var timestamp string = r.FormValue("timestamp")
	var nonce string = r.FormValue("nonce")
	strs := sort.StringSlice{t, timestamp, nonce}
	sort.Strings(strs)
	var str string
	for _, s := range strs {
		str += s
	}
	h := sha1.New()
	h.Write([]byte(str))
	return fmt.Sprintf("%x", h.Sum(nil)) == signature
}
