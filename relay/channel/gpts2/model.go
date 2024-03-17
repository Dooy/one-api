package gpts2

type ImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

type Content struct {
	Type   string       `json:"type"`
	Text   string       `json:"text,omitempty"`
	Source *ImageSource `json:"source,omitempty"`
}

type Message struct {
	Role string `json:"role"`
	//Content []Content `json:"content"`
	Content string `json:"content"`
}

// {
//     "max_tokens": 800,
//     "temperature": 0.7,
//     "frequency_penalty": 0,
//     "presence_penalty": 0,
//     "top_p": 0.95,
//     "stream": true,
//     "model": "gpt-4-all",
//     "messages": [
//         {
//             "role": "user",
//             "content": "你是谁？"
//         }
//     ]
// }

type Request struct {
	MaxTokens        int       `json:"max_tokens,omitempty"`
	Temperature      float64   `json:"temperature,omitempty"`
	FrequencyPenalty float64   `json:"frequency_penalty,omitempty"`
	PresencePenalty  float64   `json:"presence_penalty,omitempty"`
	TopP             float64   `json:"top_p,omitempty"`
	Stream           bool      `json:"stream,omitempty"`
	Model            string    `json:"model"`
	Messages         []Message `json:"messages"`
}
