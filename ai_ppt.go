package aibase

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

type AIPPT struct {
	APPId       string
	APISecret   string
	Text        string
	header      http.Header
	AiPPTReq    AiPPTReq
	Query       string `json:"query"`
	CreateModel string `json:"create_model"`
	Theme       string `json:"theme"`
	BusinessId  string `json:"business_id"`
	Author      string `json:"author"`
	IsCardNote  bool   `json:"is_card_note"`
	IsCoverImg  bool   `json:"is_cover_img"`
	Language    string `json:"language"`
	IsFigure    bool   `json:"is_figure"`
}

// 创建 PPT 生成任务
type AiPPTReq struct {
	Query       string `json:"query"`
	CreateModel string `json:"create_model"`
	Theme       string `json:"theme"`
	BusinessId  string `json:"business_id"`
	Author      string `json:"author"`
	IsCardNote  bool   `json:"is_card_note"`
	IsCoverImg  bool   `json:"is_cover_img"`
	Language    string `json:"language"`
	IsFigure    bool   `json:"is_figure"`
}

type PPTListResponse struct {
	Flag  bool   `json:"flag"`
	Code  int    `json:"code"`
	Desc  string `json:"desc"`
	Count *int   `json:"count"`
	Data  []Data `json:"data"`
}

type Data struct {
	Key       string `json:"key"`
	Name      string `json:"name"`
	Thumbnail string `json:"thumbnail"`
}

func (a *AIPPT) SetAuthor(s string) *AIPPT {
	a.Author = s
	return a
}
func (a *AIPPT) SetIsFigure(s bool) *AIPPT {
	a.IsFigure = s
	return a
}
func (a *AIPPT) SetTheme(s string) *AIPPT {
	a.Theme = s
	return a
}

// 构建请求 body 体
func (a *AIPPT) getBody() map[string]interface{} {
	body := make(map[string]interface{})
	if a.Text != "" {
		body["query"] = a.Text
	}
	if a.Author == "" {
		body["author"] = "介子云"
	} else {
		body["author"] = a.Author
	}
	if a.Theme != "" {
		body["theme"] = a.Theme
	}
	if a.IsFigure == true {
		body["is_figure"] = true
	}

	return body

}

// 轮询任务进度，返回完整响应信息
func (a *AIPPT) getProcess(sid string) (string, error) {
	fmt.Println("sid:" + sid)
	if sid != "" {
		url := "https://zwapi.xfyun.cn/api/aippt/progress?sid=" + sid
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			logrus.Error("轮询任务,创建 HTTP 请求失败:", err)
			return "", fmt.Errorf("轮询任务 创建 HTTP 请求失败: %w", err)
		}
		req.Header = a.header

		client := &http.Client{}
		resp, _ := client.Do(req)
		defer resp.Body.Close()

		bodyBytes, _ := io.ReadAll(resp.Body)
		return string(bodyBytes), nil
	} else {
		return "", nil
	}
}

type GetResult struct {
	Flag  bool        `json:"flag"`
	Code  int         `json:"code"`
	Desc  string      `json:"desc"`
	Count interface{} `json:"count"`
	Data  struct {
		Process int         `json:"process"`
		PptUrl  interface{} `json:"pptUrl"`
		ErrMsg  interface{} `json:"errMsg"`
	} `json:"data"`
}

// 获取 PPT，以下载连接形式返回
func (a *AIPPT) getResult() (string, error) {
	//创建 PPT 生成任务
	taskId, err := a.CreatePPT(a.Text)
	if err != nil {
		logrus.Error("创建 PPT 任务失败:", err)
		return "", err
	}
	time.Sleep(25 * time.Second)
	for {
		response, err2 := a.getProcess(taskId)
		if err2 != nil {
			logrus.Error("轮询任务失败:", err2)
			return "", err2
		}
		fmt.Println("response:" + response)
		var respData map[string]interface{}
		err := json.Unmarshal([]byte(response), &respData)
		if err != nil {
			logrus.Error("解析响应失败:", err)
			return "", fmt.Errorf("轮询任务失败解析响应失败: %w", err)
		}

		data, _ := respData["data"].(map[string]interface{})
		process, _ := data["process"].(float64)
		Description, _ := respData["desc"].(string)
		fmt.Println("Description:" + fmt.Sprintf("%f", Description))
		if process == 100 {
			pptUrl, _ := data["pptUrl"].(string)
			return pptUrl, nil
		}
		if Description != "成功" {
			logrus.Error("任务失败:" + Description)
			return "", fmt.Errorf("任务失败: %w", Description)
		}
		time.Sleep(5 * time.Second)

	}
}

// 获取 PPT，以下载连接形式返回
func (a *AIPPT) GeneratePPT(text string) (string, error) {
	// 控制台获取

	a.Text = text
	result, err := a.getResult()
	if err != nil {
		logrus.Error("获取 PPT 失败:", err)
		return "", fmt.Errorf("获取 PPT 失败: %w", err)
	}
	fmt.Println("生成的 PPT 请从此地址获取：\n" + result)
	return result, nil
}

func (a *AIPPT) GetThemeList() (*PPTListResponse, error) {
	var respData PPTListResponse
	url := "https://zwapi.xfyun.cn/api/aippt/themeList"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logrus.Error("获取ppt样式列表,创建 HTTP 请求失败:", err)
		return &respData, fmt.Errorf("获取ppt样式列表 创建 HTTP 请求失败: %w", err)
	}

	req.Header = a.header

	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	err = json.Unmarshal(bodyBytes, &respData)
	if err != nil {
		logrus.Error("解析响应失败:", err)
		return &respData, fmt.Errorf("解析响应失败: %w", err)
	}
	fmt.Println("response:", respData)

	return &respData, nil

}
func NewAiPPT() *AIPPT {

	ApiId := viper.GetString("ai_ppt.api_id")
	APISecret := viper.GetString("ai_ppt.api_secret")
	timestamp := time.Now().Unix()
	var a AIPPT
	a.APPId = ApiId
	a.APISecret = APISecret
	signature := a.getSignature(timestamp)

	headers := make(http.Header)
	headers.Set("appId", a.APPId)
	headers.Set("timestamp", strconv.FormatInt(timestamp, 10))
	headers.Set("signature", signature)
	headers.Set("Content-Type", "application/json; charset=utf-8")
	a.header = headers

	return &a

	// 需纠错文本
}
