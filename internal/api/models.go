package api

type PostMessageRequest struct {
	Keywords []string `json:"keywords"`
	Message  string   `json:"message"`
}

type GetMessagesResponse struct {
	Messages []string `json:"messages"`
}
