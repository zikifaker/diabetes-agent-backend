package voicerecognition

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	modelName  = "paraformer-realtime-v2"
	sampleRate = 16000
)

type Header struct {
	Action       string `json:"action"`
	TaskID       string `json:"task_id"`
	Streaming    string `json:"streaming"`
	Event        string `json:"event"`
	ErrorCode    string `json:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

type Output struct {
	Sentence struct {
		BeginTime   int64  `json:"begin_time"`
		EndTime     *int64 `json:"end_time"`
		Text        string `json:"text"`
		SentenceEnd bool   `json:"sentence_end"`
	} `json:"sentence"`
}

type Payload struct {
	TaskGroup  string `json:"task_group"`
	Task       string `json:"task"`
	Function   string `json:"function"`
	Model      string `json:"model"`
	Parameters Params `json:"parameters"`
	Input      Input  `json:"input"`
	Output     Output `json:"output,omitempty"`
}

type Params struct {
	Format        string   `json:"format"`
	SampleRate    int      `json:"sample_rate"`
	LanguageHints []string `json:"language_hints"`
}

type Input struct {
}

type Event struct {
	Header  Header  `json:"header"`
	Payload Payload `json:"payload"`
}

// Recognize 语音识别服务
func Recognize(audioFile *multipart.FileHeader) (string, error) {
	conn, err := wsConnectionPool.Get()
	if err != nil {
		return "", fmt.Errorf("failed to get WebSocket connection: %v", err)
	}
	defer wsConnectionPool.Put(conn)

	taskStarted := make(chan bool)
	taskDone := make(chan bool)
	var result strings.Builder

	// 异步接收 WebSocket 消息
	go startMessageReceiver(conn, taskStarted, taskDone, &result)

	// 发送 run-task 命令
	taskID, err := sendRunTaskCmd(conn)
	if err != nil {
		return "", fmt.Errorf("failed to send run-task cmd: %v", err)
	}

	// 等待 task-started 事件
	if err := waitForTaskStarted(taskStarted, taskID); err != nil {
		return "", fmt.Errorf("failed to wait for task started: %v", err)
	}

	// 发送音频数据
	if err := sendAudioData(conn, audioFile); err != nil {
		return "", fmt.Errorf("failed to send audio data: %v", err)
	}

	// 发送 finish-task 命令
	if err := sendFinishTaskCmd(conn, taskID); err != nil {
		return "", fmt.Errorf("failed to send finish-task cmd: %v", err)
	}

	<-taskDone

	return result.String(), nil
}

func startMessageReceiver(wsConnection *WSConnection, taskStarted chan<- bool, taskDone chan<- bool, result *strings.Builder) {
	for {
		_, message, err := wsConnection.conn.ReadMessage()
		if err != nil {
			slog.Error("Failed to read messages from server", "err", err)
			return
		}

		var event Event
		if err = json.Unmarshal(message, &event); err != nil {
			slog.Error("Failed to parse event", "err", err)
			continue
		}

		if handleEvent(event, taskStarted, taskDone, result) {
			return
		}
	}
}

func sendRunTaskCmd(wsConnection *WSConnection) (string, error) {
	runTaskCmd := generateRunTaskCmd()
	runTaskCmdJSON, err := json.Marshal(runTaskCmd)
	if err != nil {
		return "", fmt.Errorf("failed to marshal run-task cmd: %v", err)
	}

	err = wsConnection.conn.WriteMessage(websocket.TextMessage, runTaskCmdJSON)
	if err != nil {
		return "", fmt.Errorf("failed to write run-task cmd: %v", err)
	}

	return runTaskCmd.Header.TaskID, nil
}

func waitForTaskStarted(taskStarted chan bool, taskID string) error {
	select {
	case <-taskStarted:
		slog.Info("start task successfully", "task_id", taskID)
	case <-time.After(10 * time.Second):
		return fmt.Errorf("timeout waiting for task-started")
	}
	return nil
}

func sendAudioData(wsConnection *WSConnection, audioFile *multipart.FileHeader) error {
	file, err := audioFile.Open()
	if err != nil {
		return fmt.Errorf("failed to open audio file: %v", err)
	}
	defer file.Close()

	// 假设每 100ms 的音频文件为 1024 字节
	buf := make([]byte, 1024)
	for {
		n, err := file.Read(buf)
		if n > 0 {
			if err := wsConnection.conn.WriteMessage(websocket.BinaryMessage, buf[:n]); err != nil {
				return fmt.Errorf("failed to send audio data: %v", err)
			}
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error reading file: %v", err)
		}
	}
	return nil
}

func sendFinishTaskCmd(wsConnection *WSConnection, taskID string) error {
	finishTaskCmd := generateFinishTaskCmd(taskID)
	finishTaskCmdJSON, err := json.Marshal(finishTaskCmd)
	if err != nil {
		return fmt.Errorf("failed to marshal finish-task cmd: %v", err)
	}

	err = wsConnection.conn.WriteMessage(websocket.TextMessage, finishTaskCmdJSON)
	if err != nil {
		return fmt.Errorf("failed to write finish-task cmd: %v", err)
	}
	return nil
}

func handleEvent(event Event, taskStarted chan<- bool, taskDone chan<- bool, result *strings.Builder) bool {
	switch event.Header.Event {
	case "task-started":
		slog.Info("receive task-started event", "task_id", event.Header.TaskID)
		taskStarted <- true
	case "result-generated":
		// 若语音识别出完整的句子，将结果写入 result
		if event.Payload.Output.Sentence.SentenceEnd {
			result.WriteString(event.Payload.Output.Sentence.Text)
		}
	case "task-finished":
		slog.Info("task finished", "task_id", event.Header.TaskID)
		taskDone <- true
		return true
	case "task-failed":
		handleTaskFailed(event)
		taskDone <- true
		return true
	default:
		slog.Info("unexpected event", "event", event)
	}
	return false
}

func handleTaskFailed(event Event) {
	if event.Header.ErrorMessage != "" {
		slog.Error("task failed",
			"task_id", event.Header.TaskID,
			"err", event.Header.ErrorMessage,
		)
	} else {
		slog.Error("task failed due to unknown reason", "task_id", event.Header.TaskID)
	}
}

func generateRunTaskCmd() *Event {
	return &Event{
		Header: Header{
			Action:    "run-task",
			TaskID:    uuid.New().String(),
			Streaming: "duplex",
		},
		Payload: Payload{
			TaskGroup: "audio",
			Task:      "asr",
			Function:  "recognition",
			Model:     modelName,
			Parameters: Params{
				Format:     "wav",
				SampleRate: sampleRate,
			},
			Input: Input{},
		},
	}
}

func generateFinishTaskCmd(taskID string) *Event {
	return &Event{
		Header: Header{
			Action:    "finish-task",
			TaskID:    taskID,
			Streaming: "duplex",
		},
		Payload: Payload{
			Input: Input{},
		},
	}
}
