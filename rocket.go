package cachet

// GODEBUG=netdns=cgo go run rocket.go

import (
	"bytes"
	"crypto/tls"
	"encoding/json"

	// "errors"

	"net/http"
	// "net/url"
	// "os"
	// "os/signal"
	// "strings"
	// "sync"
	"time"

	"github.com/Sirupsen/logrus"
)

type LoginForm struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

type RocketResp struct {
	Status string                 `json:"status"`
	Data   map[string]interface{} `json:"data"`
}

type RocketMessage struct {
	RoomID  string `json:"roomId"`
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

func sendMsgRocket(message string, token string, userID string, cfg *CachetMonitor) {

	msg := RocketMessage{}
	msg.RoomID = cfg.API.RocketRoomID
	msg.Channel = cfg.API.RocketRoomName
	msg.Text = message
	jsonBytes, _ := json.Marshal(msg)

	req, err := http.NewRequest("POST", cfg.API.RocketURL+"/api/v1/chat.postMessage", bytes.NewBuffer(jsonBytes))
	defer req.Body.Close()

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Auth-Token", token)
	req.Header.Set("X-User-Id", userID)

	transport := http.DefaultTransport.(*http.Transport)
	transport.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"http/1.1"},
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   5 * time.Second,
	}

	_, err = client.Do(req)
	if err != nil {
		print("Failed to send message")
		return
	}
	/*
		body, err := ioutil.ReadAll(res.Body)
		bodyString := string(body)
		print("##DEBUG " + bodyString)
	*/
}

func loginRocket(cfg *CachetMonitor) (token string, userID string) {
	loginForm := LoginForm{}
	loginForm.User = cfg.API.RocketUser
	loginForm.Password = cfg.API.RocketPassword
	jsonBytes, _ := json.Marshal(loginForm)

	req, err := http.NewRequest("POST", cfg.API.RocketURL+"/api/v1/login", bytes.NewBuffer(jsonBytes))
	defer req.Body.Close()

	req.Header.Set("Content-Type", "application/json")

	transport := http.DefaultTransport.(*http.Transport)
	transport.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"http/1.1"},
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   5 * time.Second,
	}

	res, err := client.Do(req)
	if err != nil {
		//panic(err)
		print("failed to connect to " + cfg.API.RocketURL + "/api/v1/login")
		return "", ""
	}

	body := RocketResp{}
	err = json.NewDecoder(res.Body).Decode(&body)
	token = body.Data["authToken"].(string)
	userID = body.Data["userId"].(string)
	return token, userID
}

func PostIncidentRocket(message string, cfg *CachetMonitor) {
	l := logrus.WithFields(logrus.Fields{
		"monitor": "rocket",
		"time":    time.Now().Format(cfg.DateFormat),
	})
	l.Printf("Send rocket notification: %s", message)
	token, userID := loginRocket(cfg)
	if token != "" {
		sendMsgRocket(message, token, userID, cfg)
	}
}
