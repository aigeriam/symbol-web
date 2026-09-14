package handlers

type AIHandler struct {
	Client *ai.Client
}

func NewAiHandler(client *ai.Client) *AIHandler {
	return &AIHandler{Client: client}
}
