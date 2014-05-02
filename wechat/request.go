package wechat

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (

	// Reply format
	replyText    = "<xml>%s<MsgType><![CDATA[text]]></MsgType><Content><![CDATA[%s]]></Content></xml>"
	replyImage   = "<xml>%s<MsgType><![CDATA[image]]></MsgType><Image><MediaId><![CDATA[%s]]></MediaId></Image></xml>"
	replyVoice   = "<xml>%s<MsgType><![CDATA[voice]]></MsgType><Voice><MediaId><![CDATA[%s]]></MediaId></Voice></xml>"
	replyVideo   = "<xml>%s<MsgType><![CDATA[video]]></MsgType><Video><MediaId><![CDATA[%s]]></MediaId><Title><![CDATA[%s]]></Title><Description><![CDATA[%s]]></Description></Video></xml>"
	replyMusic   = "<xml>%s<MsgType><![CDATA[music]]></MsgType><Music><Title><![CDATA[%s]]></Title><Description><![CDATA[%s]]></Description><MusicUrl><![CDATA[%s]]></MusicUrl><HQMusicUrl><![CDATA[%s]]></HQMusicUrl><ThumbMediaId><![CDATA[%s]]></ThumbMediaId></Music></xml>"
	replyNews    = "<xml>%s<MsgType><![CDATA[news]]></MsgType><ArticleCount>%d</ArticleCount><Articles>%s</Articles></xml>"
	replyHeader  = "<ToUserName><![CDATA[%s]]></ToUserName><FromUserName><![CDATA[%s]]></FromUserName><CreateTime>%d</CreateTime>"
	replyArticle = "<item><Title><![CDATA[%s]]></Title> <Description><![CDATA[%s]]></Description><PicUrl><![CDATA[%s]]></PicUrl><Url><![CDATA[%s]]></Url></item>"
	// QR scene request
	requestQRScene      = "{\"expire_seconds\":%d,\"action_name\":\"QR_SCENE\",\"action_info\":{\"scene\":{\"scene_id\":%d}}}"
	requestQRLimitScene = "{\"action_name\":\"QR_LIMIT_SCENE\",\"action_info\":{\"scene\":{\"scene_id\":%d}}}"
)

// Use to output reply
type ResponseWriter interface {
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

type Music struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	MusicUrl     string `json:"musicurl"`
	HQMusicUrl   string `json:"hqmusicurl"`
	ThumbMediaId string `json:"thumb_media_id"`
}

type Article struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	PicUrl      string `json:"picurl"`
	Url         string `json:"url"`
}

type response struct {
	ErrorCode    int    `json:"errcode,omitempty"`
	ErrorMessage string `json:"errmsg,omitempty"`
}

func (w *WeChat) postRequest(reqURL string, data []byte) ([]byte, error) {
	for i := 0; i < w.retry; i++ {
		token, err := w.getAccessToken()
		if err != nil {
			return nil, err
		}
		r, err := http.Post(reqURL+token, "application/json; charset=utf-8", bytes.NewReader(data))
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
				return nil, errors.New(fmt.Sprintf("WeiXin send post request reply[%d]: %s", result.ErrorCode, result.ErrorMessage))
			}
		}
	}
	return nil, errors.New("WeiXin post request too many times:" + reqURL)
}

func (wc *WeChat) postMessage(msg interface{}) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = wc.postRequest(WeChatHost+"/message/custom/send?access_token=", data)
	return err
}

type responseWriterData struct {
	WC           *WeChat             //WeChat structure
	W            http.ResponseWriter // http response writer
	toUserName   string
	fromUserName string
}

// Format reply message header
func (d *responseWriterData) replyHeader() string {
	return fmt.Sprintf(replyHeader, d.toUserName, d.fromUserName, time.Now().Unix())
}
func (d *responseWriterData) ReplyText(text string) {
	d.W.Write([]byte(fmt.Sprintf(replyText, d.replyHeader(), text)))
}
func (d *responseWriterData) ReplyImage(mediaId string) {
	d.W.Write([]byte(fmt.Sprintf(replyImage, d.replyHeader(), mediaId)))
}
func (d *responseWriterData) ReplyVoice(mediaId string) {
	d.W.Write([]byte(fmt.Sprintf(replyVoice, d.replyHeader(), mediaId)))
}
func (d *responseWriterData) ReplyVideo(mediaId string, title string, description string) {
	d.W.Write([]byte(fmt.Sprintf(replyVideo, d.replyHeader(), mediaId, title, description)))
}
func (d *responseWriterData) ReplyMusic(m *Music) {
	d.W.Write([]byte(fmt.Sprintf(replyMusic, d.replyHeader(), m.Title, m.Description, m.MusicUrl, m.HQMusicUrl, m.ThumbMediaId)))
}
func (d *responseWriterData) ReplyNews(articles []Article) {
	var ctx string
	for _, article := range articles {
		ctx += fmt.Sprintf(replyArticle, article.Title, article.Description, article.PicUrl, article.Url)
	}
	d.W.Write([]byte(fmt.Sprintf(replyNews, d.replyHeader(), len(articles), ctx)))
}

// Post message
func (d *responseWriterData) PostText(text string) error {
	var msg struct {
		ToUser  string `json:"touser"`
		MsgType string `json:"msgtype"`
		Text    struct {
			Content string `json:"content"`
		} `json:"text"`
	}
	msg.ToUser = d.toUserName
	msg.MsgType = "text"
	msg.Text.Content = text
	return d.WC.postMessage(msg)
}
func (d *responseWriterData) PostImage(mediaId string) error {
	var msg struct {
		ToUser  string `json:"touser"`
		MsgType string `json:"msgtype"`
		Image   struct {
			MediaId string `json:"media_id"`
		} `json:"image"`
	}
	msg.ToUser = d.toUserName
	msg.MsgType = "image"
	msg.Image.MediaId = mediaId
	return d.WC.postMessage(msg)
}
func (d *responseWriterData) PostVoice(mediaId string) error {
	var msg struct {
		ToUser  string `json:"touser"`
		MsgType string `json:"msgtype"`
		Voice   struct {
			MediaId string `json:"media_id"`
		} `json:"voice"`
	}
	msg.ToUser = d.toUserName
	msg.MsgType = "voice"
	msg.Voice.MediaId = mediaId
	return d.WC.postMessage(msg)
}
func (d *responseWriterData) PostVideo(mediaId string, title string, description string) error {
	var msg struct {
		ToUser  string `json:"touser"`
		MsgType string `json:"msgtype"`
		Video   struct {
			MediaId     string `json:"media_id"`
			Title       string `json:"title"`
			Description string `json:"description"`
		} `json:"video"`
	}
	msg.ToUser = d.toUserName
	msg.MsgType = "video"
	msg.Video.MediaId = mediaId
	msg.Video.Title = title
	msg.Video.Description = description
	return d.WC.postMessage(msg)
}
func (d *responseWriterData) PostMusic(music *Music) error {
	var msg struct {
		ToUser  string `json:"touser"`
		MsgType string `json:"msgtype"`
		Music   *Music `json:"music"`
	}
	msg.ToUser = d.toUserName
	msg.MsgType = "video"
	msg.Music = music
	return d.WC.postMessage(msg)
}
func (d *responseWriterData) PostNews(articles []Article) error {
	var msg struct {
		ToUser  string `json:"touser"`
		MsgType string `json:"msgtype"`
		News    struct {
			Articles []Article `json:"articles"`
		} `json:"news"`
	}
	msg.ToUser = d.toUserName
	msg.MsgType = "news"
	msg.News.Articles = articles
	return d.WC.postMessage(msg)
}

// Media operator
func (d *responseWriterData) UploadMediaFromFile(mediaType string, fp string) (string, error) {
	file, err := os.Open(fp)
	if err != nil {
		return "", err
	}
	defer file.Close()
	return d.UploadMedia(mediaType, filepath.Base(fp), file)
}
func (d *responseWriterData) DownloadMediaToFile(mediaId string, filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()
	return d.DownloadMedia(mediaId, file)
}
func (d *responseWriterData) UploadMedia(mediaType string, filename string, reader io.Reader) (string, error) {
	reqURL := WeChatFileURL + "/upload?type=" + mediaType + "&access_token="
	for i := 0; i < d.WC.retry; i++ {
		bodyBuf := &bytes.Buffer{}
		bodyWriter := multipart.NewWriter(bodyBuf)
		fileWriter, err := bodyWriter.CreateFormFile("filename", filename)
		if err != nil {
			return "", err
		}
		if _, err = io.Copy(fileWriter, reader); err != nil {
			return "", err
		}
		contentType := bodyWriter.FormDataContentType()
		bodyWriter.Close()
		token, err := d.WC.getAccessToken()
		if err != nil {
			return "", err
		}
		r, err := http.Post(reqURL+token, contentType, bodyBuf)
		if err != nil {
			return "", err
		}
		defer r.Body.Close()
		reply, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return "", err
		}
		var result struct {
			response
			Type      string `json:"type"`
			MediaId   string `json:"media_id"`
			CreatedAt int64  `json:"created_at"`
		}
		err = json.Unmarshal(reply, &result)
		if err != nil {
			return "", err
		} else {
			switch result.ErrorCode {
			case 0:
				return result.MediaId, nil
			case 42001: // access_token timeout and retry
				continue
			default:
				return "", errors.New(fmt.Sprintf("WeiXin upload[%d]: %s", result.ErrorCode, result.ErrorMessage))
			}

		}
	}
	return "", errors.New("WeiXin upload media too many times")
}
func (d *responseWriterData) DownloadMedia(mediaId string, writer io.Writer) error {
	reqURL := WeChatFileURL + "/get?media_id=" + mediaId + "&access_token="
	for i := 0; i < d.WC.retry; i++ {
		token, err := d.WC.getAccessToken()
		if err != nil {
			return err
		}
		r, err := http.Get(reqURL + token)
		if err != nil {
			return err
		}
		defer r.Body.Close()
		if r.Header.Get("Content-Type") != "text/plain" {
			_, err := io.Copy(writer, r.Body)
			return err
		} else {
			reply, err := ioutil.ReadAll(r.Body)
			if err != nil {
				return err
			}
			var result response
			if err := json.Unmarshal(reply, &result); err != nil {
				return err
			} else {
				switch result.ErrorCode {
				case 0:
					return nil
				case 42001: // access_token timeout and retry
					continue
				default:
					return errors.New(fmt.Sprintf("WeiXin download[%d]: %s", result.ErrorCode, result.ErrorMessage))
				}
			}
		}

	}
	return errors.New("WeiXin download media too many times")
}
