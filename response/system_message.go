package response

import "time"

type GetSystemMessagesResponse struct {
	Total    int64                   `json:"total"`
	Messages []SystemMessageResponse `json:"messages"`
}

type SystemMessageResponse struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	IsRead    bool      `json:"is_read"`
}
