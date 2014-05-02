package wechat

import (
	"errors"
	"testing"
	"time"
)

type xx struct {
}

//Test create WeChat object
func TestA(t *testing.T) {
	appid, secret, token := `aaaa`, `xxxxx`, `dddd`
	wc, err := New(appid, secret, token, &xx{})
	if err != nil {
		t.Error(err)
	} else {
		//This part of token is just test for once, because this resource is limited.
		//token, err := wc.getAccessToken()
		//if err != nil {
		//	t.Error(err)
		//} else {
		//	t.Log(token)
		//}
		t.Log(wc)
		t.Log(at)
	}

}

var at = &AccessToken{
	Token:      "xxxxxxxxxxxx",
	ExpireTime: time.Now(),
}

func (x *xx) Read() (*AccessToken, error) {
	if at != nil {
		return at, nil
	} else {
		return nil, errors.New("xxx")
	}
}
func (x *xx) Write(t *AccessToken) error {
	at = t
	return nil
}
