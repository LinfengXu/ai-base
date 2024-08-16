package aibase

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io"
	"net/http"
	"time"
)

const (
	API_create_session  = "/api/ai-base/session/create_session" // 创建会话
	API_send_message    = "/api/ai-base/qwen/qwenlong_use"      // 发送编排消息
	API_QwenLong_chat   = "/api/ai-base/qwen/qwenlong_chat"     // 发送session和user问题聊天消息
	API_QwenModel       = "/api/ai-base/qwen/open_model"        // 发送session和user问题聊天消息
	APi_QwenLONG_stream = "/api/ai-base/qwen/qwenlong_stream"   //Qwenlong流式聊天
)

type QwenUse struct {
	SessionId   string `json:"session_id"` // 会话ID
	Messages    []Message
	GroupID     uint   `json:"group_id"`     // 群组ID
	ServiceType string `json:"service_type"` // 聊天服务类型
	c           echo.Context
}

func (m *QwenUse) SetContext(c echo.Context) *QwenUse {
	m.c = c
	return m
}
func (m *QwenUse) GetMessages() []Message {
	return m.Messages
}

func (m *QwenUse) Add(role string, content string) {
	m.Messages = append(m.Messages, Message{role, content})
}

// AddSystemMessage 方法用于向Messages结构体中添加一条系统消息
func (m *QwenUse) SystemAdd(content string) {
	m.Messages = append(m.Messages, Message{"system", content})
}

// AddUserMessage 方法用于向Messages结构体中添加一条用户消息
func (m *QwenUse) UserAdd(content string) {
	m.Messages = append(m.Messages, Message{"user", content})
}

// AddAssistantMessage 方法用于向Messages结构体中添加一条助手消息
func (m *QwenUse) AssistantAdd(content string) {
	m.Messages = append(m.Messages, Message{"assistant", content})
}

func qwenLongModelStream(qwenlongreq QwenLongStreamUseReq, ch chan ResponseToFontend, stopped *bool) {

	fmt.Println("QwenLongModelStream start:")

	url := fmt.Sprintf("%s%s", viper.GetString("ai_base_service_host"), APi_QwenLONG_stream)

	requestBody, err := json.Marshal(qwenlongreq)
	if err != nil {

		logrus.Error("QwenLongStream Error encoding request:", err)
		ch <- ResponseToFontend{
			Problem:    fmt.Sprintf("QwenLongStream Error encoding request: %v", err),
			IsFinished: true,
		}
		return
	}
	logrus.Info("requestBody:", string(requestBody))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		ch <- ResponseToFontend{
			Problem:    fmt.Sprintf("QwenLongStream Error creating request: %v", err),
			IsFinished: true,
		}
		fmt.Println("QwenLongStream Error creating request:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Error sending request:", err)
		ch <- ResponseToFontend{
			Problem:    fmt.Sprintf("Error sending request: %v", err),
			IsFinished: true,
		}
		return
		// 确保通道关闭
	}
	if resp.StatusCode > 399 {
		//logrus.Errorf("model QwenLongStream stream api status error: %s; status: %d-%s",  resp.Body, resp.StatusCode, resp.Status)

		errinfo, _ := io.ReadAll(resp.Body)
		logrus.Errorf(
			"model QwenLongStream stream api status error: %s; status: %d-%s",
			errinfo,
			resp.StatusCode,
			resp.Status,
		)
		ch <- ResponseToFontend{
			Problem:    fmt.Sprintf("model QwenLongStream stream api status error: %s; status: %d-%s", errinfo, resp.StatusCode, resp.Status),
			IsFinished: true,
		}

		return
	}

	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)
	var responseToFontend ResponseToFontend
	for {

		if *stopped {
			logrus.Infof("stream chat stopped")
			req.Close = true
			return
		}

		line, err := reader.ReadBytes('\n')
		if err != nil {
			logrus.Error("Error reading response:", err)
			break
		}

		//logrus.Info("line: ", string(line))
		// Check if the line starts with "data: "
		if bytes.HasPrefix(line, []byte("data: ")) {
			// Extract JSON data after "data: "
			jsonData := line[len("data: "):]

			//logrus.Info("jsonData: ", string(jsonData))

			err := json.Unmarshal(jsonData, &responseToFontend)
			if err != nil {

				logrus.Error("Error decoding response:", err)
				continue
			}

			logrus.Info("Content: ", responseToFontend.Content)

			if responseToFontend.IsFinished == true {
				fmt.Println("responseToFontend: ", responseToFontend.Content)

				logrus.Info("Finish Reason: ", responseToFontend.IsFinished)

				ch <- responseToFontend
				*stopped = true
				break

			} else {

				fmt.Println("responseToFontend: ", responseToFontend.Content)
				ch <- responseToFontend
			}
		}
	}

	fmt.Println("==================[DONE]============================")
	return
}

func (q *QwenUse) SendQwenLongStream() (ResponseToFontend, error) {

	qwenlongreq := QwenLongStreamUseReq{
		Messages:    q.Messages,
		GroupID:     q.GroupID,
		ServiceType: q.ServiceType,
	}
	eventCh := make(chan ResponseToFontend)
	var stopped bool
	defer func() {
		fmt.Println(" defer close:")
		close(eventCh)
	}()

	var res ResponseToFontend

	go qwenLongModelStream(qwenlongreq, eventCh, &stopped)

	for {
		fmt.Println("for start..........")
		select {
		case event, ok := <-eventCh:
			if !ok {
				fmt.Println("event end..........")
				stopped = true
				break
			}
			res = event

			data, _ := json.Marshal(event)

			if event.IsFinished {
				fmt.Fprintf(q.c.Response(), "data: %s\n\n", data)
				q.c.Response().Flush()
				stopped = true
				return event, nil

			} else {

				fmt.Fprintf(q.c.Response(), "data: %s\n\n", data)
				q.c.Response().Flush()
			}

		case <-q.c.Request().Context().Done():
			fmt.Println("client disconnected..........")
			time.Sleep(time.Second * 3)
			stopped = true
			break
			return res, fmt.Errorf("client disconnected")

		}

		if stopped {
			fmt.Println("for end..........")
			break
		}

	}
	fmt.Println("QwenLongStreamUse end..........")

	return res, fmt.Errorf("QwenLongStreamUse end")
}
