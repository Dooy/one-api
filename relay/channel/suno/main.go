package suno

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/relay/channel/openai"
	"github.com/songquanpeng/one-api/relay/model"
	"github.com/songquanpeng/one-api/relay/util"
	"github.com/tidwall/gjson"
)

func ConvertRequest(textRequest model.GeneralOpenAIRequest, a *Adaptor) *Request {
	rq := Request{
		Mv: "chirp-v3-0",
	}

	prompt := "[Verse]\n漫长的夜会找到她\n等待阳光照在脸上\n我们心跳齐奏的舞台\n用真实的话去演出 (ooh-oh)\n\n[Verse 2]\n我告诉你心中的秘密\n用音乐将情感流泪\n我们一起唱出真心话\n让每个人都能感受到 (ooh-yeah)\n\n[Chorus]\n真心话 真心话\n流淌在每个音符间\n不需要掩饰的真心话\n让我们的爱无限延续 (ooh-oh-oh)"

	for _, message := range textRequest.Messages {

		rq.Title = message.StringContent()
	}
	rq.Prompt = prompt
	rq.Tags = ""
	//fmt.Printf("go new : %+v\n", a.meta)
	fp, err := a.GetLyrics(rq.Title)

	for i := 0; i < 1; i++ {
		//fmt.Printf("my god\n\n")
		if err != nil || fp.Title == "" {
			a.fetchBearer()
			fp, err = a.GetLyrics(rq.Title)
		} else {
			break
		}
	}
	rq.Prompt = ""
	if err == nil && fp.Title != "" {
		rq.Prompt = fp.Lyrics
		rq.Tags = fp.Tag
		rq.Title = fp.Title
	}
	//fmt.Printf("ConvertRequest : %+v\n", rq)
	return &rq
}

func StreamHandler(c *gin.Context, resp *http.Response, meta *util.RelayMeta, request *model.GeneralOpenAIRequest, a *Adaptor) (*model.ErrorWithStatusCode, *model.Usage) {
	logger.Debugf(c.Request.Context(), "StreamHandler %v , %v ", resp.StatusCode, request.Stream)
	defer resp.Body.Close()
	myMsg := ""
	toSteam := func(c *gin.Context, Content string) {
		var choice openai.ChatCompletionsStreamResponseChoice
		choice.Delta.Content = Content
		response := openai.ChatCompletionsStreamResponse{
			Id:      fmt.Sprintf("chatcmpl-%s", helper.GetUUID()),
			Object:  "chat.completion.chunk",
			Created: helper.GetTimestamp(),
			Model:   "suno-v3",
			Choices: []openai.ChatCompletionsStreamResponseChoice{choice},
		}
		jsonResponse, err := json.Marshal(response)
		if err != nil {
			logger.SysError("error marshalling stream response: " + err.Error())
			return
		}
		if request.Stream {
			c.Render(-1, common.CustomEvent{Data: "data: " + string(jsonResponse)})
			c.Writer.Flush()
		}

		myMsg = myMsg + Content

	}
	dazi := func(c *gin.Context, text string, sec int) {
		//toSteam(c, "\n\n---\n\n")
		for _, char := range text {
			//result = append(result, string(char))
			toSteam(c, string(char))
			time.Sleep(time.Duration(sec) * time.Millisecond)
		}
		//toSteam(c, "\n\n---\n\n")
	}

	//defer resp.Body.Close()
	//var responseData map[string]interface{}
	//err := json.NewDecoder(resp.Body).Decode(&responseData)
	json2, err := io.ReadAll(resp.Body)
	if err != nil {
		return openai.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	jsonstr := string(json2)
	logger.Infof(c.Request.Context(), "StreamHandler, id= %v, %v  ", gjson.Get(jsonstr, "id"), gjson.Get(jsonstr, "clips.#.id"))
	ids := gjson.Get(jsonstr, "clips.#.id").Array()
	//idkey:= ids[0].String()+','+ids[1].String()
	var idsStr []string
	for _, id := range ids {
		idsStr = append(idsStr, id.String())
	}
	//先执行不然白费
	go a.FetchResut(idsStr)

	//toSteam(c, "歌名："+a.MyRequest.Title+"\n\n")
	dazi(c, "#### 🎵 "+a.MyRequest.Title+"\n\n", 10)
	dazi(c, a.MyRequest.Tags+"\n\n", 10)

	var usage model.Usage
	toSteam(c, "\n\n---\n\n")
	dazi(c, a.MyRequest.Prompt, 50)
	toSteam(c, "\n\n---\n\n")
	dazi(c, ">ID\n>"+gjson.Get(jsonstr, "id").String(), 2)

	toSteam(c, "\n\n生成中:")
	isCompt := false
	for i := 0; i < 500; i++ {
		if isCompt {
			time.Sleep(5 * time.Second)
		} else {
			time.Sleep(9 * time.Second)
			toSteam(c, "🎵")
		}

		rz, err := a.FetchResut(idsStr)

		if err == nil && rz != nil && len(rz) > 0 {

			for _, item := range rz {
				logger.Debugf(c.Request.Context(), "FetchResut [%s] %s , is:%v", item.Id, item.Status, a.isStop)
				if item.Status == "complete" {
					toSteam(c, "\n\n\n"+item.Title+"\n\n")
					toSteam(c, "![image]("+item.ImageUrl+")\n")
					toSteam(c, "音频🎧：[点击播放]("+item.AudioUrl+")\n")
					toSteam(c, "视频🖥：[点击播放]("+item.VideoUrl+")\n")
					toSteam(c, "\n")

					idsStr = filterSlice(idsStr, func(s string) bool {
						return s != item.Id
					})
					isCompt = true
				}
			}

		} else {
			go a.fetchBearer()
		}
		if len(idsStr) == 0 {
			break
		}
	}

	//}
	usage.PromptTokens = openai.CountTokenMessages(a.request.Messages, "gpt-3.5")
	usage.CompletionTokens = openai.CountTokenText(myMsg, "gpt-3.5")
	usage.TotalTokens = usage.CompletionTokens + usage.PromptTokens
	if !request.Stream {
		choice := openai.TextResponseChoice{
			Index: 0,
			Message: model.Message{
				Role:    "assistant",
				Content: myMsg,
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
	} else {
		c.Render(-1, common.CustomEvent{Data: "data: [DONE]"})
		c.Writer.Flush()
	}
	return nil, &usage
}

func filterSlice(arr []string, filter func(string) bool) []string {
	var filtered []string
	for _, v := range arr {
		if filter(v) {
			filtered = append(filtered, v)
		}
	}
	return filtered
}
