package aibase

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

// 获取签名
func (a *AIPPT) getSignature(ts int64) string {
	// 对 app_id 和时间戳进行 MD5 加密
	auth := a.md5(a.APPId + strconv.FormatInt(ts, 10))
	// 使用 HMAC-SHA1 算法对加密后的字符串进行加密
	return a.hmacSha1Encrypt(auth, a.APISecret)
}

func (a *AIPPT) hmacSha1Encrypt(encryptText, encryptKey string) string {
	key := []byte(encryptKey)
	h := hmac.New(sha1.New, key)
	h.Write([]byte(encryptText))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func (a *AIPPT) md5(text string) string {
	h := md5.Sum([]byte(text))
	return fmt.Sprintf("%x", h)
}

func (a *AIPPT) CreatePPT(text string) (string, error) {
	a.Text = text
	url := "https://zwapi.xfyun.cn/api/aippt/create"
	timestamp := time.Now().Unix()
	signature := a.getSignature(timestamp)
	body := a.getBody()

	// 构建请求头
	headers := make(http.Header)
	headers.Set("appId", a.APPId)
	headers.Set("timestamp", strconv.FormatInt(timestamp, 10))
	headers.Set("signature", signature)
	headers.Set("Content-Type", "application/json; charset=utf-8")
	a.header = headers

	// 序列化请求体
	jsonBody, err := json.Marshal(body)
	fmt.Println(string(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON body: %w", err)
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Println("CreatePPT failed to create HTTP request:", err)
		return "", fmt.Errorf("CreatePPT failed to create HTTP request: %w", err)
	}
	req.Header = headers

	// 发送 HTTP 请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("CreatePPT failed to send HTTP request:", err)
		return "", fmt.Errorf("CreatePPT failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// 解析响应 JSON
	var respData map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &respData); err != nil {
		return "", fmt.Errorf("failed to unmarshal response JSON: %w", err)
	}

	// 检查响应码
	code, ok := respData["code"].(float64)
	if !ok {
		return "", fmt.Errorf("invalid response code type")
	}
	if code != 0 {
		return "", fmt.Errorf("failed to create PPT task: code %v", code)
	}

	// 提取 sid
	data, ok := respData["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid response data type")
	}
	sid, ok := data["sid"].(string)
	if !ok {
		return "", fmt.Errorf("sid not found in response data")
	}

	return sid, nil
}
