package suno

//	{
//	  "prompt": "[Verse]\n清晨的阳光洒满大地\n梦里的世界都无边际\n岁月的风 都变得温柔\n为你我愿化身为星辰\n\n[Verse 2]\n弯弯的月儿照亮夜空\n思念的绵绵永不停息\n宁静的风 带走了烦恼\n在你的怀抱我找到了家\n\n[Chorus]\n温柔是一种力量\n化解了人间的忧伤\n我的世界因你而美妙\n爱的味道甜蜜又静好",
//	  "tags": "",
//	  "mv": "chirp-v3-0",
//	  "title": "阳光洒满",
//	  "continue_clip_id": null,
//	  "continue_at": null
//	}
type Request struct {
	Prompt string `json:"prompt,omitempty"`
	Tags   string `json:"tags,omitempty"`
	Mv     string `json:"mv,omitempty"`
	Title  string `json:"title,omitempty"`
	// ContinueClipId string `json:"continue_clip_id"`
	// ContinueAt     string `json:"continue_at"`
}

type BearerType struct {
	Object string `json:"object,omitempty"`
	Jwt    string `json:"jwt,omitempty"`
}
type LyricsRep struct {
	Lyrics string
	Title  string
	Tag    string
}
type LyricsFetchRep struct {
	Text   string `json:"text,omitempty"`
	Title  string `json:"title,omitempty"`
	Status string `json:"status,omitempty"`
}

type ResutType struct {
	//A                 string
	Id                string `json:"id,omitempty"`
	VideoUrl          string `json:"video_url,omitempty"`
	AudioUrl          string `json:"audio_url,omitempty"`
	ImageLargeUrl     string `json:"image_large_url,omitempty"`
	ImageUrl          string `json:"image_url,omitempty"`
	MajorModelVersion string `json:"major_model_version,omitempty"`
	ModelName         string `json:"model_name,omitempty"`
	Status            string `json:"status,omitempty"`
	Title             string `json:"title,omitempty"`
	// Id                string `json:"image_url,omitempty"`
	// Id                string `json:"image_url,omitempty"`
}

var ModelList = []string{
	"suno-v3",
}
