package wechat

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	// WeChat host URL
	WeChatHost        = "https://api.weixin.qq.com/cgi-bin"
	WeChatQRScene     = "https://api.weixin.qq.com/cgi-bin/qrcode"
	WeChatShowQRScene = "https://mp.weixin.qq.com/cgi-bin/showqrcode"
	WeChatFileURL     = "http://file.api.weixin.qq.com/cgi-bin/media"
)

//WeChat Struct
type WeChat struct {
	appid       string
	secret      string
	token       string
	retry       int
	atrw        AccessTokenReaderWriter
	routes      []*route
	accesstoken AccessToken
}

//Access Token struct
type AccessToken struct {
	Token      string    //WeChat access token
	ExpireTime time.Time //expire time
}

//Method to store and create access token
type AccessTokenReaderWriter interface {
	//Read access token
	Read() (*AccessToken, error)
	//Write access token
	Write(*AccessToken) error
}

//Create WeChat Object
func New(appid, secret, token string, atrw AccessTokenReaderWriter) (*WeChat, error) {
	wc := &WeChat{
		appid:  appid,
		secret: secret,
		token:  token,
		retry:  3,
		atrw:   atrw,
		routes: []*route{},
	}
	return wc, nil
}

func (wc *WeChat) UpdateAccessToken() (*AccessToken, error) {
	at, err := wc.atrw.Read()
	if err == nil && time.Since(at.ExpireTime).Seconds()+100 < 0 {
		return at, nil
	}
	t, err := fetchAccessToken(wc.appid, wc.secret)
	if err != nil {
		return nil, err
	} else {
		err = wc.atrw.Write(t)
		if err != nil {
			return nil, err
		} else {
			return t, nil
		}
	}
}

//Get Access Token retry three times
func (wc *WeChat) getAccessToken() (string, error) {
	if wc.accesstoken.Token != "" && time.Since(wc.accesstoken.ExpireTime).Seconds() < 0 {
		return wc.accesstoken.Token, nil
	}
	for i := 1; i <= wc.retry; i++ {
		t, err := wc.atrw.Read()
		if err == nil {
			if time.Since(t.ExpireTime).Seconds() >= 0 {
				t, err = fetchAccessToken(wc.appid, wc.secret)
				if err != nil {
					continue
				} else {
					err = wc.atrw.Write(t)
					if err != nil {
						return "", err
					}
				}
			}
			wc.accesstoken = *t
			return t.Token, nil
		}
	}
	return "", errors.New("Cannot get Access Token")
}

//Get accesstoken from wechat.
func fetchAccessToken(appid string, secret string) (*AccessToken, error) {
	resp, err := http.Get(WeChatHost + "/token?grant_type=client_credential&appid=" + appid + "&secret=" + secret)
	if err != nil {
		log.Println("Get access token failed: ", err)
		return nil, err
	} else {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("Read access token failed: ", err)
			return nil, err
		} else {
			var res struct {
				AccessToken string `json:"access_token"`
				ExpiresIn   int64  `json:"expires_in"`
			}
			if err := json.Unmarshal(body, &res); err != nil {
				log.Println("Parse access token failed: ", err)
				return nil, err
			} else {
				//log.Printf("AuthAccessToken token=%s expires_in=%d", res.AccessToken, res.ExpiresIn)

				return &AccessToken{
					Token:      res.AccessToken,
					ExpireTime: time.Now().Add(time.Duration((res.ExpiresIn - 100) * 1000 * 1000 * 1000)),
				}, nil
			}
		}
	}
}
