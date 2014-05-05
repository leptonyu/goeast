package wechat

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

const (
	wechatGroupList = `https://api.weixin.qq.com/cgi-bin/groups/get?access_token=`
)

type Group struct {
	Id    int
	Name  string
	Count int
}

type Groups struct {
	Groups []Group
}

func (wc *WeChat) GetGroups() (Groups, error) {
	gs := Groups{}
	token, err := wc.getAccessToken()
	if err != nil {
		return gs, err
	}
	resp, err := http.Get(wechatGroupList + token)
	if err != nil {
		return gs, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return gs, err
	}
	if err := json.Unmarshal([]byte(body), &gs); err != nil {
		return gs, err
	}
	return gs, nil
}
