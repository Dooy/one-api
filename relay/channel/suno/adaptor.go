package suno

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/relay/channel"
	"github.com/songquanpeng/one-api/relay/model"
	"github.com/songquanpeng/one-api/relay/util"
)

type Adaptor struct {
	request       *model.GeneralOpenAIRequest
	sess          string
	cookie        string
	Authorization string
	isStop        bool
	Context       context.Context
	meta          *util.RelayMeta
	MyRequest     *Request
}

func (a *Adaptor) Init(meta *util.RelayMeta) {
	//a.meta = meta
	a.initCookie(meta)
}
func (a *Adaptor) GetRequestURL(meta *util.RelayMeta) (string, error) {
	//return fmt.Sprintf("%s/v1/messages", meta.BaseURL), nil
	a.meta = meta
	return fmt.Sprintf("%s/api/generate/v2/", meta.BaseURL), nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *util.RelayMeta) error {
	channel.SetupCommonRequestHeader(c, req, meta)

	req.Header.Set("Cookie", a.cookie)

	req.Header.Set("Authorization", "Bearer "+a.Authorization)

	return nil
}

func (a *Adaptor) initCookie(meta *util.RelayMeta) error {
	//a.Context = c.Request.Context()
	if meta.APIKey == "" {
		return nil
	}
	parts := strings.Split(meta.APIKey, "||")
	a.sess = parts[0]
	a.cookie = parts[len(parts)-1]
	a.isStop = true
	a.meta = meta
	//defer a.over()
	err := a.fetchBearer()
	if err != nil {
		return err
	}
	return nil
}
func (a *Adaptor) over() {
	logger.SysLog("do over")
	a.isStop = true
}

func (a *Adaptor) reGo() {
	if a.isStop {
		return
	}
	time.Sleep(60 * time.Second)
	a.fetchBearer()
}

func (a *Adaptor) fetchBearer() error {
	go a.reGo()
	fullRequestURL := fmt.Sprintf("https://clerk.suno.ai/v1/client/sessions/%s/tokens?_clerk_js_version=4.71.0", a.sess)
	req, err := http.NewRequest(http.MethodPost, fullRequestURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Cookie", a.cookie)
	resp, err := util.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var responseData BearerType

	if err = json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return err
	}
	if responseData.Jwt != "" {
		a.Authorization = responseData.Jwt
	}
	logger.SysLog("Auth:  " + a.Authorization)
	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	a.request = request
	a.MyRequest = ConvertRequest(*request, a)
	return a.MyRequest, nil
}

func (a *Adaptor) setHeader(req *http.Request) {
	req.Header.Set("Cookie", a.cookie)
	req.Header.Set("Authorization", "Bearer "+a.Authorization)
}

func (a *Adaptor) GetLyrics(prompt string) (*LyricsRep, error) {
	//https://studio-api.suno.ai/api/generate/lyrics/
	lyricsRep := &LyricsRep{Lyrics: ""}
	lyricsRep.Tag = a.RandStyle()

	barUrl := fmt.Sprintf("%s/api/generate/lyrics/", a.meta.BaseURL)
	values := map[string]string{"prompt": prompt}
	jsonData, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, barUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	// req.Header.Set("Cookie", a.cookie)
	// req.Header.Set("Authorization", "Bearer "+a.Authorization)
	a.setHeader(req)

	resp, err := util.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var responseData map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseData)
	if err != nil {
		return nil, err
	}

	id, _ := responseData["id"].(string)
	lk := &LyricsFetchRep{}
	for i := 0; i < 20; i++ {
		lk, err = a.fetchLyrics(id)
		if err == nil && lk.Status == "complete" {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if lk == nil {
		return nil, errors.New("未知道歌词")
	}
	lyricsRep.Lyrics = lk.Text
	lyricsRep.Title = lk.Title
	return lyricsRep, nil
}

func (a *Adaptor) RandStyle() string {
	//rand.Seed(time.Now().UnixNano())
	s := []string{"acoustic", "aggressive", "anthemic", "atmospheric", "bouncy", "chill", "dark", "dreamy", "electronic", "emotional", "epic", "experimental", "futuristic", "groovy", "heartfelt", "infectious", "melodic", "mellow", "powerful", "psychedelic", "romantic", "smooth", "syncopated", "uplifting", ""}
	l := []string{"afrobeat", "anime", "ballad", "bedroom pop", "bluegrass", "blues", "classical", "country", "cumbia", "dance", "dancepop", "delta blues", "electropop", "disco", "dream pop", "drum and bass", "edm", "emo", "folk", "funk", "future bass", "gospel", "grunge", "grime", "hip hop", "house", "indie", "j-pop", "jazz", "k-pop", "kids music", "metal", "new jack swing", "new wave", "opera", "pop", "punk", "raga", "rap", "reggae", "reggaeton", "rock", "rumba", "salsa", "samba", "sertanejo", "soul", "synthpop", "swing", "synthwave", "techno", "trap", "uk garage"}
	randomS := s[rand.Intn(len(s))]
	randomL := l[rand.Intn(len(l))]
	randomS2 := s[rand.Intn(len(s))]
	randomL2 := l[rand.Intn(len(l))]

	// 拼凑新的词组
	return randomS + " " + randomL + "," + randomS2 + " " + randomL2
}

func (a *Adaptor) fetchLyrics(id string) (*LyricsFetchRep, error) {
	//lyricsRep := &LyricsFetchRep{ }
	barUrl := fmt.Sprintf("%s/api/generate/lyrics/%s", a.meta.BaseURL, id)

	req, _ := http.NewRequest(http.MethodGet, barUrl, nil)

	a.setHeader(req)

	resp, err := util.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var lyricsRep LyricsFetchRep
	err = json.NewDecoder(resp.Body).Decode(&lyricsRep)
	if err != nil {
		return nil, err
	}
	//fmt.Printf("lyricsRep : %+v\n", lyricsRep)
	return &lyricsRep, nil
}

// 做req准备 会调用自己的  GetRequestURL-> SetupRequestHeader
func (a *Adaptor) DoRequest(c *gin.Context, meta *util.RelayMeta, requestBody io.Reader) (*http.Response, error) {
	a.Context = c.Request.Context()

	return channel.DoRequestHelper(a, c, meta, requestBody)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *util.RelayMeta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	// if meta.IsStream {
	// 	err, usage = StreamHandler(c, resp, meta, a.request)
	// } else {
	// 	err, usage = Handler(c, resp, meta)
	// }

	err, usage = StreamHandler(c, resp, meta, a.request, a)
	a.over()
	return
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return "suno"
}

// 查询结果
func (a *Adaptor) FetchResut(idsStr []string) ([]ResutType, error) {
	return a.fetchResut(idsStr)
}

func (a *Adaptor) fetchResut(idsStr []string) ([]ResutType, error) {
	id := strings.Join(idsStr, ",")
	baseUrl := fmt.Sprintf("%s/api/feed/?ids=%s", a.meta.BaseURL, id)
	req, _ := http.NewRequest(http.MethodGet, baseUrl, nil)
	//logger.SysLog("start ID :" + id)
	a.setHeader(req)

	util.HTTPClient.Timeout = 30 * time.Second

	resp, err := util.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// var lyricsRep LyricsFetchRep
	// err = json.NewDecoder(resp.Body).Decode(&lyricsRep)
	body, _ := io.ReadAll(resp.Body)
	var rz []ResutType
	err = json.Unmarshal(body, &rz)
	if err != nil {
		return nil, err
	}
	//logger.SysLog("ab:" + rz[0].Status)
	return rz, nil
}
