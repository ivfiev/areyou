package api

type PostMessage struct {
	Keywords []string `json:"keywords"`
	Message  string   `json:"message"`
}
