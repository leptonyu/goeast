package wechat

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Use to store QR code
type QRScene struct {
	Ticket        string `json:"ticket"`
	ExpireSeconds int    `json:"expire_seconds"`
}

// Custom Menu
type Menu struct {
	Buttons []MenuButton `json:"button,omitempty"`
}

// Menu Button
type MenuButton struct {
	Name       string       `json:"name"`
	Type       string       `json:"type,omitempty"`
	Key        string       `json:"key,omitempty"`
	Url        string       `json:"url,omitempty"`
	SubButtons []MenuButton `json:"sub_button,omitempty"`
}

// Create QR scene
func (wc *WeChat) CreateQRScene(sceneId int, expires int) (*QRScene, error) {
	reply, err := wc.postRequest(WeChatQRScene+"/create?access_token=", []byte(fmt.Sprintf(requestQRScene, expires, sceneId)))
	if err != nil {
		return nil, err
	}
	var qr QRScene
	if err := json.Unmarshal(reply, &qr); err != nil {
		return nil, err
	}
	return &qr, nil
}

// Create  QR limit scene
func (wc *WeChat) CreateQRLimitScene(sceneId int) (*QRScene, error) {
	reply, err := wc.postRequest(WeChatQRScene+"/create?access_token=", []byte(fmt.Sprintf(requestQRLimitScene, sceneId)))
	if err != nil {
		return nil, err
	}
	var qr QRScene
	if err := json.Unmarshal(reply, &qr); err != nil {
		return nil, err
	}
	return &qr, nil
}

// Custom menu
func (wc *WeChat) CreateMenu(menu *Menu) error {
	data, err := json.Marshal(menu)
	if err != nil {
		return err
	} else {
		_, err := wc.postRequest(WeChatHost+"/menu/create?access_token=", data)
		return err
	}
}

func (wc *WeChat) GetMenu() (*Menu, error) {
	reply, err := wc.sendGetRequest(WeChatHost + "/menu/get?access_token=")
	if err != nil {
		return nil, err
	} else {
		var result struct {
			MenuCtx *Menu `json:"menu"`
		}
		if err := json.Unmarshal(reply, &result); err != nil {
			return nil, err
		} else {
			return result.MenuCtx, nil
		}
	}
}

// Delete Menu
func (wc *WeChat) DeleteMenu() error {
	_, err := wc.sendGetRequest(WeChatHost + "/menu/delete?access_token=")
	return err
}

// Send get request to WeChat server.
func (wc *WeChat) sendGetRequest(reqURL string) ([]byte, error) {
	for i := 0; i < wc.retry; i++ {
		token, err := wc.getAccessToken()
		if err != nil {
			return nil, err
		}
		r, err := http.Get(reqURL + token)
		if err != nil {
			return nil, err
		}
		defer r.Body.Close()
		reply, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		var result response
		if err := json.Unmarshal(reply, &result); err != nil {
			return nil, err
		} else {
			switch result.ErrorCode {
			case 0:
				return reply, nil
			case 42001: // access_token timeout and retry
				continue
			default:
				return nil, errors.New(fmt.Sprintf("WeChat send get request reply[%d]: %s", result.ErrorCode, result.ErrorMessage))
			}

		}
	}
	return nil, errors.New("WeChat post request too many times:" + reqURL)
}
