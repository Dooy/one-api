package anthropic

import "github.com/songquanpeng/one-api/relay/model"

type Metadata struct {
	UserId string `json:"user_id"`
}

type Request struct {
	Model             string   `json:"model"`
	Prompt            string   `json:"prompt"`
	MaxTokensToSample int      `json:"max_tokens_to_sample"`
	StopSequences     []string `json:"stop_sequences,omitempty"`
	Temperature       float64  `json:"temperature,omitempty"`
	TopP              float64  `json:"top_p,omitempty"`
	TopK              int      `json:"top_k,omitempty"`
	//Metadata    `json:"metadata,omitempty"`
	Stream bool `json:"stream,omitempty"`
}

type ChatRequest struct {
	Model         string          `json:"model"`
	Messages      []model.Message `json:"messages"`
	MaxTokens     int             `json:"max_tokens"`
	StopSequences []string        `json:"stop_sequences,omitempty"`
	Temperature   float64         `json:"temperature,omitempty"`
	TopP          float64         `json:"top_p,omitempty"`
	TopK          int             `json:"top_k,omitempty"`
	//Metadata    `json:"metadata,omitempty"`
	Stream bool `json:"stream,omitempty"`
}

type Error struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type Response struct {
	Completion string `json:"completion"`
	StopReason string `json:"stop_reason"`
	Model      string `json:"model"`
	Error      Error  `json:"error"`
}

// type ChatCompletionsStreamResponseChoice struct {
// 	Index int `json:"index"`
// 	Delta struct {
// 		Content string `json:"content"`
// 		Role    string `json:"role,omitempty"`
// 	} `json:"delta"`
// 	FinishReason *string `json:"finish_reason,omitempty"`
// }

//	type ChatCompletionsStreamResponse struct {
//		Id      string                                `json:"id"`
//		Object  string                                `json:"object"`
//		Created int64                                 `json:"created"`
//		Model   string                                `json:"model"`
//		Choices []ChatCompletionsStreamResponseChoice `json:"choices"`
//		Usage   *model.Usage                          `json:"usage"`
//	}
type ResponseStream struct {
	//{"type": "content_block_delta", "index": 0, "delta": {"type": "text_delta", "text": "Hello"}}
	Type  string `json:"type"`
	Index int    `json:"index"`
	Delta struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"delta"`
}

//	{
//	  "content": [
//	    {
//	      "text": "Hi! My name is Claude.",
//	      "type": "text"
//	    }
//	  ],
//	  "id": "msg_013Zva2CMHLNnXjNJJKqJ2EF",
//	  "model": "claude-3-opus-20240229",
//	  "role": "assistant",
//	  "stop_reason": "end_turn",
//	  "stop_sequence": null,
//	  "type": "message",
//	  "usage": {
//	    "input_tokens": 10,
//	    "output_tokens": 25
//	  }
//	}
type ResponseNoStream struct {
	Id           string `json:"id"`
	Model        string `json:"model"`
	Role         string `json:"role"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence"`
	Type         string `json:"type"`
	Usage        struct {
		InputTokens int `json:"input_tokens"`
		OnputTokens int `json:"output_tokens"`
	} `json:"usage"`
	Content []struct {
		Text string `json:"text"`
		Type string `json:"type"`
	} `json:"content"`
	Error Error `json:"error"`
}
