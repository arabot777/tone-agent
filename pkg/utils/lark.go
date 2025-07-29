package utils

import (
	"bytes"
	"encoding/json"
	"tone/agent/pkg/common/env"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

const (
	demoTitle = "告警"
)

func SendMachineStartSlowMsg(robotUrl string, name string, text string) {
	SendPostMsgToFeishu(robotUrl, fmt.Sprintf("[%s] %s", env.Environment(), demoTitle), []string{
		fmt.Sprintf("name: %s", name),
		fmt.Sprintf("text: %s", text),
	})
}

// ===========
const (
	defaultHttpTimeout = 3
	JsonContentType    = "application/json"
)

var (
	httpClient = &http.Client{Timeout: defaultHttpTimeout * time.Second}
)

type (
	feiShuMsg struct {
		MsgType string                 `json:"msg_type"`
		Content map[string]interface{} `json:"content"`
	}
)

type (
	singleLineMsg struct {
		Tag  string `json:"tag"`
		Text string `json:"text"`
		Href string `json:"href,omitempty"`
	}
)

// 发送飞书富文本消息
func SendPostMsgToFeishu(robotUrl, title string, lines []string) {
	if robotUrl == "" {
		return
	}
	if len(lines) == 0 {
		return
	}
	msgList := make([][]*singleLineMsg, 0, 2*len(lines)-1)
	for i := 0; i < len(lines); i++ {
		msgList = append(msgList, []*singleLineMsg{{Tag: "text", Text: lines[i]}})
		if i != len(lines)-1 {
			msgList = append(msgList, []*singleLineMsg{{Tag: "text", Text: ""}})
		}
	}
	msg := &feiShuMsg{
		MsgType: "post",
		Content: map[string]interface{}{
			"post": map[string]interface{}{
				"zh_cn": map[string]interface{}{
					"title":   title,
					"content": msgList,
				},
			},
		},
	}
	_, _ = postJson(robotUrl, msg)
}

func postJson(url string, data interface{}) ([]byte, error) {
	content, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	buffer := bytes.NewBuffer(content)
	resp, err := httpClient.Post(url, JsonContentType, buffer)
	if err != nil {
		return nil, errors.Wrapf(err, "http request error: url: %s, param: %+v", url, data)
	}
	return parseResp(url, resp)
}

func parseResp(url string, resp *http.Response) ([]byte, error) {
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("http status code: %d,url: %s", resp.StatusCode, url)
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "http response 解析出错。url:%s", url)
	}
	return data, nil
}
