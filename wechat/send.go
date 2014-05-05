package wechat

import (
	"encoding/json"
	"errors"
	"fmt"
)

const (
	psText       = `{"touser":[%v],"msgtype":"text","text":{"content":"%v"}}`
	psImage      = `{"touser":[%v],"msgtype":"image","image":{"media_id":"%v"}}`
	psVoice      = `{"touser":[%v],"msgtype":"voice","voice":{"media_id":"%v"}}`
	psVideo      = `{"touser":[%v],"msgtype":"video","video":{"media_id":"%v","title":"%v","description":"%v"}}`
	psGetVideoId = `{"media_id": "%v","title": "%v","description": "%v"}`
	psArticle    = `{"touser":[%v],"msgtype":"mpnews","mpnews":{"media_id":"%v"}}`

	gsText    = `{"filter":{"group_id":"%v"},"msgtype":"text","text":{"content":"%v"}}`
	gsImage   = `{"filter":{"group_id":"%v"},"msgtype":"image","image":{"media_id":"%v"}}`
	gsVoice   = `{"filter":{"group_id":"%v"},"msgtype":"voice","voice":{"media_id":"%v"}}`
	gsArticle = `{"filter":{"group_id":"%v"},"msgtype":"mpnews","mpnews":{"media_id":"%v"}}`
	gsVideo   = `{"filter":{"group_id":"%v"},"msgtype":"mpvideo","mpvideo":{"media_id":"%v"}}`

	wechatURLSendWithOpenId = `https://api.weixin.qq.com/cgi-bin/message/mass/send?access_token=`
	wechatURLGetVideoId     = `https://file.api.weixin.qq.com/cgi-bin/media/uploadvideo?access_token=`
)

//Get new video id from wechat.
type GetVideoId struct {
	Type   string
	Id     string `json:"media_id"`
	Create int    `json:"create_at"`
}

type Users []string

func (us Users) String() string {
	l := len(us)
	if l == 0 {
		return ""
	}
	str := `"` + us[0] + `"`
	for i := 1; i < l; i++ {
		str += `,"` + us[i] + `"`
	}
	return str
}

type GroupMessage struct {
	MsgType     string
	GroupId     int
	Content     string
	MediaId     string
	Title       string
	Description string
}
type Message struct {
	MsgType     string
	Users       Users
	Content     string
	MediaId     string
	Title       string
	Description string
}

func (wc *WeChat) getVideoId(mediaid, title, description string) (GetVideoId, error) {
	value := GetVideoId{}
	msg := fmt.Sprintf(psGetVideoId, mediaid, title, description)
	re, err := wc.postRequest(wechatURLGetVideoId, []byte(msg))
	if err != nil {
		return value, err
	}
	err = json.Unmarshal(re, &value)
	return value, err
}

type returnMessage struct {
	Type    string `json:"type"`
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
	MsgId   int    `json:"msg_id"`
}

type SendMessage interface {
	SendText(users Users, text string) error
	SendGroupText(groupid int, text string) error
	SendImage(users Users, mediaid string) error
	SendGroupImage(groupid int, mediaid string) error
	SendVoice(users Users, mediaid string) error
	SendGroupVoice(groupid int, mediaid string) error
	SendMpnews(users Users, mediaid string) error
	SendGroupMpnews(groupid int, mediaid string) error
	SendVideo(users Users, mediaid, title, description string) error
	SendGroupVideo(groupid int, mediaid, title, description string) error
}

func (wc *WeChat) send(m Message) error {
	var msg string
	switch m.MsgType {
	case "text":
		msg = fmt.Sprintf(psText, m.Users.String(), m.Content)
	case "image":
		msg = fmt.Sprintf(psImage, m.Users.String(), m.MediaId)
	case "voice":
		msg = fmt.Sprintf(psVoice, m.Users.String(), m.MediaId)
	case "video":
		{
			gvi, err := wc.getVideoId(m.MediaId, m.Title, m.Description)
			if err != nil {
				return err
			}
			msg = fmt.Sprintf(psVideo, m.Users.String(), gvi.Id, m.Title, m.Description)
		}
	case "mpnews":
		msg = fmt.Sprintf(psArticle, m.Users.String(), m.MediaId)
	default:
		return errors.New("Wrong type of message")
	}
	re, err := wc.postRequest(wechatURLSendWithOpenId, []byte(msg))
	if err != nil {
		return err
	}
	rm := returnMessage{}
	if err := json.Unmarshal(re, &rm); err != nil {
		return err
	}
	if rm.ErrCode != 0 {
		return errors.New(rm.ErrMsg)
	}
	return nil

}

func (wc *WeChat) sendGroup(m GroupMessage) error {
	var msg string
	switch m.MsgType {
	case "text":
		msg = fmt.Sprintf(gsText, m.GroupId, m.Content)
	case "image":
		msg = fmt.Sprintf(gsImage, m.GroupId, m.MediaId)
	case "voice":
		msg = fmt.Sprintf(gsVoice, m.GroupId, m.MediaId)
	case "video":
		{
			gvi, err := wc.getVideoId(m.MediaId, m.Title, m.Description)
			if err != nil {
				return err
			}
			msg = fmt.Sprintf(gsVideo, m.GroupId, gvi.Id)
		}
	case "mpnews":
		msg = fmt.Sprintf(gsArticle, m.GroupId, m.MediaId)
	default:
		return errors.New("Wrong type of message")
	}
	re, err := wc.postRequest(wechatURLSendWithOpenId, []byte(msg))
	if err != nil {
		return err
	}
	rm := returnMessage{}
	if err := json.Unmarshal(re, &rm); err != nil {
		return err
	}
	if rm.ErrCode != 0 {
		return errors.New(rm.ErrMsg)
	}
	return nil

}
func (wc *WeChat) SendText(users Users, text string) error {
	return wc.send(Message{
		MsgType: "text",
		Users:   users,
		Content: text,
	})
}

func (wc *WeChat) SendGroupText(groupid int, text string) error {
	return wc.sendGroup(GroupMessage{
		MsgType: "text",
		GroupId: groupid,
		Content: text,
	})
}
func (wc *WeChat) SendImage(users Users, mediaid string) error {
	return wc.send(Message{
		MsgType: "image",
		Users:   users,
		MediaId: mediaid,
	})
}
func (wc *WeChat) SendGroupImage(groupid int, mediaid string) error {
	return wc.sendGroup(GroupMessage{
		MsgType: "image",
		GroupId: groupid,
		MediaId: mediaid,
	})
}
func (wc *WeChat) SendVoice(users Users, mediaid string) error {
	return wc.send(Message{
		MsgType: "voice",
		Users:   users,
		MediaId: mediaid,
	})
}

func (wc *WeChat) SendGroupVoice(groupid int, mediaid string) error {
	return wc.sendGroup(GroupMessage{
		MsgType: "voice",
		GroupId: groupid,
		MediaId: mediaid,
	})
}
func (wc *WeChat) SendMpnews(users Users, mediaid string) error {
	return wc.send(Message{
		MsgType: "mpnews",
		Users:   users,
		MediaId: mediaid,
	})
}

func (wc *WeChat) SendGroupMpnews(groupid int, mediaid string) error {
	return wc.sendGroup(GroupMessage{
		MsgType: "mpnews",
		GroupId: groupid,
		MediaId: mediaid,
	})
}
func (wc *WeChat) SendVideo(users Users, mediaid, title, description string) error {
	return wc.send(Message{
		MsgType:     "video",
		Users:       users,
		MediaId:     mediaid,
		Title:       title,
		Description: description,
	})
}

func (wc *WeChat) SendGroupVideo(groupid int, mediaid, title, description string) error {
	return wc.sendGroup(GroupMessage{
		MsgType:     "video",
		GroupId:     groupid,
		MediaId:     mediaid,
		Title:       title,
		Description: description,
	})
}
