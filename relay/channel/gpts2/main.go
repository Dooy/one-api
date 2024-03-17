package gpts2

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/launchdarkly/eventsource"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/relay/channel/openai"
	"github.com/songquanpeng/one-api/relay/model"
	"github.com/songquanpeng/one-api/relay/util"
)

func ConvertRequest(textRequest model.GeneralOpenAIRequest) *Request {
	claudeRequest := Request{
		Model:       textRequest.Model,
		MaxTokens:   textRequest.MaxTokens,
		Temperature: textRequest.Temperature,
		TopP:        textRequest.TopP,
		Stream:      true, //textRequest.Stream,
		//Messages:    textRequest.Messages ,
	}
	if claudeRequest.MaxTokens == 0 {
		claudeRequest.MaxTokens = 4096
	}
	for _, message := range textRequest.Messages {
		iMessage := Message{
			Role:    message.Role,
			Content: message.StringContent(),
		}
		claudeRequest.Messages = append(claudeRequest.Messages, iMessage)
	}

	return &claudeRequest
}

func StreamHandler(c *gin.Context, resp *http.Response, meta *util.RelayMeta, request *model.GeneralOpenAIRequest) (*model.ErrorWithStatusCode, *model.Usage) {
	//logger.Debugf(c.Request.Context(), "StreamHandler %v , %v ", resp.StatusCode, request.Stream)
	var usage model.Usage
	decoder := eventsource.NewDecoder(resp.Body)
	defer decoder.Decode()
	defer resp.Body.Close()
	message := ""

	usage.PromptTokens = meta.PromptTokens
	usage.CompletionTokens = 0
	usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens

	err2 := evenDecode(decoder, func(text string) {
		if request.Stream {
			c.Render(-1, common.CustomEvent{Data: "data: " + text})
		}
		if text != "[DONE]" {
			var gptResponse openai.ChatCompletionsStreamResponse
			err := json.Unmarshal([]byte(text), &gptResponse)
			if err == nil {
				message = message + gptResponse.Choices[0].Delta.Content
			}
		}
	})
	if err2 != nil {
		return err2, nil
	}
	usage.CompletionTokens = openai.CountTokenText(message, "gpt-4")
	usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens

	if !request.Stream {
		choice := openai.TextResponseChoice{
			Index: 0,
			Message: model.Message{
				Role:    "assistant",
				Content: message,
				Name:    nil,
			},
			FinishReason: "stop",
		}

		fullTextResponse := openai.TextResponse{
			Id:      fmt.Sprintf("chatcmpl-%s", strings.ToLower(helper.GetRandomString(30))),
			Model:   meta.ActualModelName,
			Object:  "chat.completion",
			Created: helper.GetTimestamp(),
			Choices: []openai.TextResponseChoice{choice},
		}
		fullTextResponse.Usage = usage
		fullTextResponse.Model = meta.ActualModelName
		jsonResponse, err := json.Marshal(fullTextResponse)
		if err != nil {
			return openai.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
		}
		c.Writer.Header().Set("Content-Type", "application/json")
		c.Writer.WriteHeader(resp.StatusCode)
		c.Writer.Write(jsonResponse)
	}

	return nil, &usage
}

func evenDecode(decoder *eventsource.Decoder, myFun func(s string)) *model.ErrorWithStatusCode {
	for {
		event, err := decoder.Decode()
		if err != nil {
			if err == io.EOF {
				break
			}
		}

		text := event.Data()
		//logger.Debugf(c.Request.Context(), "event.Data %v", text)
		if text == "" {
			continue
		}
		if strings.Contains(text, "gptscopilot") || strings.Contains(text, "openai-now") || strings.Contains(text, "limit for conversations") {
			logger.SysLog("gptscopilot error:  " + text)
			return &model.ErrorWithStatusCode{
				Error: model.Error{
					Message: "try again",
					Type:    "gpts_fail",
					Code:    "not enough",
					Param:   "gpts_fail",
				},
				StatusCode: 425, // 这个状态直接死了 401也是直接死
			}
		}
		myFun(text)
		if text == "[DONE]" {
			break
		}

	}

	return nil

}
