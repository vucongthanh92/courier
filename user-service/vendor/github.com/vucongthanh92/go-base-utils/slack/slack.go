package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/vucongthanh92/go-base-utils/logger"

	"go.uber.org/zap"
)

type SlackConfig struct {
	Channel         string `json:"channel"`
	Username        string `json:"username"`
	UrlSlackWebHook string `json:"urlSlackWebhook"`
}

func SendSlackMessage(s SlackConfig, errMsg string) {
	// limit the message to 1000 characters
	if len(errMsg) > 1000 {
		errMsg = errMsg[:1000]
	}

	payload := `{
		"channel": "#` + s.Channel + `",
		"username": "` + s.Username + `",
		"text": "` + s.formatMessage(errMsg) + `",
		"icon_emoji": ":ghost:",
	}`

	req, err := http.NewRequest(http.MethodPost, s.UrlSlackWebHook, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		logger.Warn("Error creating request:", zap.Error(err))
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_, err = getHttpRequest[string](http.DefaultClient, req)
	if err != nil {
		return
	}
	logger.Info("Slack message sent to #" + s.Channel)
}

func (s SlackConfig) formatMessage(msg string) string {
	serviceMessage := "*" + s.Username + " :" + "*" + "\n"

	return serviceMessage + "```" + msg + "```"
}

func getHttpRequest[T any](client *http.Client, req *http.Request) (response T, err error) {
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("DefaultClient Error", zap.Error(err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logger.Warn("StatusCode Error", zap.Int("StatusCode", resp.StatusCode))
		return response, fmt.Errorf(resp.Status)
	}

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("ReadAll Error", zap.Error(err))
		return
	}
	err = json.Unmarshal(responseData, &response)
	if err != nil {
		logger.Error("Unmarshal Error", zap.Error(err))
		return
	}

	return response, nil
}
