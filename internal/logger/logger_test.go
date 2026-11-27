package logger

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"
)

type loggerMsg struct {
	Time      string `json:"time"`
	Level     string `json:"level"`
	Msg       string `json:"msg"`
	Port      int    `json:"port"`
	Framework string `json:"framework"`
}

func TestLogger(t *testing.T) {

	var recorder bytes.Buffer
	logger := New(&recorder)
	msg := "server started"
	port := 8080
	framework := "gin"
	logger.Info(msg, "port", port, "framework", framework)

	var msgLogged loggerMsg
	if err := json.NewDecoder(&recorder).Decode(&msgLogged); err != nil {
		t.Fatalf("error decoding message: %v", err)
	}
	if _, err := time.Parse(time.RFC3339Nano, msgLogged.Time); err != nil {
		t.Errorf("invalid timestamp: %v", err)
	}
	if msgLogged.Framework != framework {
		t.Fatalf("error expected %v, got: %v", framework, msgLogged.Framework)
	}
	if msgLogged.Port != 8080 {
		t.Fatalf("error expected %d, got: %d", port, msgLogged.Port)
	}
	if msgLogged.Msg != msg {
		t.Fatalf("error expected %q, got: %q", msg, msgLogged.Msg)
	}
	if msgLogged.Level != "INFO" {
		t.Fatalf("error info level, got: %v", msgLogged.Level)
	}

}
