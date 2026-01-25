package utils

import "github.com/gin-gonic/gin"

const (
	EventFileParseStart      = "file_parse_start"
	EventFileParseDone       = "file_parse_done"
	EventKBRetrievalStart    = "kb_retrieval_start"
	EventKBRetrievalDone     = "kb_retrieval_done"
	EventKBRetrievalChunkNum = "kb_retrieval_chunk_num"
	EventIntermediateSteps   = "intermediate_steps"
	EventFinalAnswer         = "final_answer"
	EventToolCallResult      = "tool_call_results"
	EventError               = "error"
	EventDone                = "done"
)

type Message struct {
	Content any `json:"content"`
}

func SetSSEHeaders(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
}

func SendSSEMessage(c *gin.Context, event string, data any) {
	// 使用 content 字段存储数据，便于前端解析
	c.SSEvent(event, Message{Content: data})
	c.Writer.Flush()
}
