package ai


type OpenAIClient struct {
	apiKey string
}

type Message struct {
	Role 	 	string `json:"role"`
	Content string `json:"content"`
}

type OpenAIRequest struct {
	Model 	 string `json:"model"`
	Messages []Message `json:"messages"`
}
