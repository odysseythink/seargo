package results

import (
	"net/url"
	"path"
	"strings"
	"time"
)

// Result is the interface that all typed search results implement.
type Result interface {
	Kind() string
	GetTitle() string
	GetURL() string
	GetContent() string
	GetEngine() string
	GetTemplate() string
	GetCategory() string
	Base() *BaseResult
	Normalize()
	DedupKey() string
}

// BaseResult holds common fields shared by all result types.
type BaseResult struct {
	Title        string            `json:"title"`
	URL          string            `json:"url"`
	Content      string            `json:"content,omitempty"`
	Engine       string            `json:"engine"`
	Engines      []string          `json:"engines,omitempty"`
	Template     string            `json:"template"`
	Category     string            `json:"category,omitempty"`
	Positions    []int             `json:"positions,omitempty"`
	Score        float64           `json:"score,omitempty"`
	PublishedAt  *time.Time        `json:"published_at,omitempty"`
	ThumbnailURL string            `json:"thumbnail_url,omitempty"`
	Domain       string            `json:"domain,omitempty"`
	Favicon      string            `json:"favicon,omitempty"`
	EngineData   map[string]any    `json:"engine_data,omitempty"`
	ParsedURL    []string          `json:"parsed_url,omitempty"`
	IsOnion      bool              `json:"is_onion,omitempty"`
}

func (b BaseResult) GetTitle() string    { return b.Title }
func (b BaseResult) GetURL() string      { return b.URL }
func (b BaseResult) GetContent() string  { return b.Content }
func (b BaseResult) GetEngine() string   { return b.Engine }
func (b BaseResult) GetTemplate() string { return b.Template }
func (b BaseResult) GetCategory() string { return b.Category }

// ---------------------------------------------------------------------------
// ImageRef — placeholder for alternative image formats.
// ---------------------------------------------------------------------------

type ImageRef struct {
	URL    string `json:"url"`
	Format string `json:"format"`
	Label  string `json:"label"`
}

// ---------------------------------------------------------------------------
// CodeLine — a single code line.
// ---------------------------------------------------------------------------

type CodeLine struct {
	Line int    `json:"line"`
	Text string `json:"text"`
}

// ---------------------------------------------------------------------------
// InfoboxAttribute — a key-value attribute in an infobox.
// ---------------------------------------------------------------------------

type InfoboxAttribute struct {
	Value string `json:"value"`
	Label string `json:"label"`
	URL   string `json:"url,omitempty"`
}

// InfoboxURL — a URL entry in an infobox.
type InfoboxURL struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

// ---------------------------------------------------------------------------
// MainResult — general web result.
// ---------------------------------------------------------------------------

type MainResult struct {
	BaseResult
}

func (r *MainResult) Kind() string     { return "main" }
func (r *MainResult) Base() *BaseResult { return &r.BaseResult }
func (r *MainResult) Normalize()        { r.BaseResult.Normalize(); r.PostNormalize() }
func (r *MainResult) PostNormalize()    {}
func (r *MainResult) DedupKey() string  { return dedupKeyFromBase(&r.BaseResult) }

// ---------------------------------------------------------------------------
// ImageResult — image search result.
// ---------------------------------------------------------------------------

type ImageResult struct {
	BaseResult
	ThumbnailSrc string     `json:"thumbnail_src,omitempty"`
	ImgSrc       string     `json:"img_src,omitempty"`
	ImgFormat    string     `json:"img_format,omitempty"`
	Resolution   string     `json:"resolution,omitempty"`
	ImgAlt       string     `json:"img_alt,omitempty"`
	Source       string     `json:"source,omitempty"`
	Width        int        `json:"width,omitempty"`
	Height       int        `json:"height,omitempty"`
	FileSize     string     `json:"file_size,omitempty"`
	Formats      []ImageRef `json:"formats,omitempty"`
}

func (r *ImageResult) Kind() string     { return "image" }
func (r *ImageResult) Base() *BaseResult { return &r.BaseResult }
func (r *ImageResult) Normalize()        { r.BaseResult.Normalize(); r.PostNormalize() }
func (r *ImageResult) PostNormalize() {
	if r.ThumbnailSrc == "" && r.ImgSrc != "" {
		r.ThumbnailSrc = r.ImgSrc
	}
	if r.Title == "" && r.ImgSrc != "" {
		r.Title = path.Base(r.ImgSrc)
	}
}
func (r *ImageResult) DedupKey() string { return dedupKeyFromBase(&r.BaseResult) }

// IsBase64 returns true if the image source is a base64 data URI.
func (i *ImageResult) IsBase64() bool {
	return strings.HasPrefix(i.ImgSrc, "data:image/")
}

// ---------------------------------------------------------------------------
// VideoResult — video search result.
// ---------------------------------------------------------------------------

type VideoResult struct {
	BaseResult
	Thumbnail  string `json:"thumbnail,omitempty"`
	IFrameSrc  string `json:"iframe_src,omitempty"`
	Length     string `json:"length,omitempty"`
	Duration   string `json:"duration,omitempty"`
	Author     string `json:"author,omitempty"`
	UploadDate string `json:"upload_date,omitempty"`
	ViewCount  int64  `json:"view_count,omitempty"`
}

func (r *VideoResult) Kind() string     { return "video" }
func (r *VideoResult) Base() *BaseResult { return &r.BaseResult }
func (r *VideoResult) Normalize()        { r.BaseResult.Normalize(); r.PostNormalize() }
func (r *VideoResult) PostNormalize() {
	if r.ThumbnailURL == "" && r.Thumbnail != "" {
		r.ThumbnailURL = r.Thumbnail
	}
}
func (r *VideoResult) DedupKey() string { return dedupKeyFromBase(&r.BaseResult) }

// ---------------------------------------------------------------------------
// NewsResult — news/article result.
// ---------------------------------------------------------------------------

type NewsResult struct {
	BaseResult
}

func (r *NewsResult) Kind() string     { return "news" }
func (r *NewsResult) Base() *BaseResult { return &r.BaseResult }
func (r *NewsResult) Normalize()        { r.BaseResult.Normalize(); r.PostNormalize() }
func (r *NewsResult) PostNormalize()    {}
func (r *NewsResult) DedupKey() string  { return dedupKeyFromBase(&r.BaseResult) }

// ---------------------------------------------------------------------------
// PaperResult — academic paper result.
// ---------------------------------------------------------------------------

type PaperResult struct {
	BaseResult
	DOI           string   `json:"doi,omitempty"`
	Journal       string   `json:"journal,omitempty"`
	Authors       []string `json:"authors,omitempty"`
	Publisher     string   `json:"publisher,omitempty"`
	Type          string   `json:"type,omitempty"`
	PublishedDate string   `json:"published_date,omitempty"`
	Editors       []string `json:"editors,omitempty"`
	PDFURL        string   `json:"pdf_url,omitempty"`
	HTMLURL       string   `json:"html_url,omitempty"`
	Comments      string   `json:"comments,omitempty"`
	Tags          []string `json:"tags,omitempty"`
	Pages         string   `json:"pages,omitempty"`
	ISSN          []string `json:"issn,omitempty"`
	ISBN          []string `json:"isbn,omitempty"`
}

func (r *PaperResult) Kind() string     { return "paper" }
func (r *PaperResult) Base() *BaseResult { return &r.BaseResult }
func (r *PaperResult) Normalize()        { r.BaseResult.Normalize(); r.PostNormalize() }
func (r *PaperResult) PostNormalize()    {}
func (r *PaperResult) DedupKey() string  { return dedupKeyFromBase(&r.BaseResult) }

// ---------------------------------------------------------------------------
// CodeResult — code search result.
// ---------------------------------------------------------------------------

type CodeResult struct {
	BaseResult
	Repository   string     `json:"repository,omitempty"`
	CodeLanguage string     `json:"code_language,omitempty"`
	Filename     string     `json:"filename,omitempty"`
	CodeLines    []CodeLine `json:"code_lines,omitempty"`
	HLLines      []int      `json:"hl_lines,omitempty"`
}

func (r *CodeResult) Kind() string     { return "code" }
func (r *CodeResult) Base() *BaseResult { return &r.BaseResult }
func (r *CodeResult) Normalize()        { r.BaseResult.Normalize(); r.PostNormalize() }
func (r *CodeResult) PostNormalize() {
	if r.Title == "" && r.Filename != "" {
		r.Title = r.Filename
	}
	if r.CodeLanguage == "" {
		r.CodeLanguage = "guess"
	}
}
func (r *CodeResult) DedupKey() string { return dedupKeyFromBase(&r.BaseResult) }

// ---------------------------------------------------------------------------
// FileResult — file / torrent result.
// ---------------------------------------------------------------------------

type FileResult struct {
	BaseResult
	FileType  string `json:"file_type,omitempty"`
	FileSize  int64  `json:"file_size,omitempty"`
	Filename  string `json:"filename,omitempty"`
	MagnetURI string `json:"magnet_uri,omitempty"`
	Seeders   int    `json:"seeders,omitempty"`
	Leechers  int    `json:"leechers,omitempty"`
}

func (r *FileResult) Kind() string     { return "file" }
func (r *FileResult) Base() *BaseResult { return &r.BaseResult }
func (r *FileResult) Normalize()        { r.BaseResult.Normalize(); r.PostNormalize() }
func (r *FileResult) PostNormalize() {
	if r.Title == "" && r.Filename != "" {
		r.Title = r.Filename
	}
	if r.FileType == "" && r.Filename != "" {
		if idx := strings.LastIndex(r.Filename, "."); idx >= 0 && idx < len(r.Filename)-1 {
			r.FileType = r.Filename[idx+1:]
		}
	}
}
func (r *FileResult) DedupKey() string { return dedupKeyFromBase(&r.BaseResult) }

// ---------------------------------------------------------------------------
// MapResult — map/geolocation result.
// ---------------------------------------------------------------------------

type MapResult struct {
	BaseResult
	Latitude    float64   `json:"latitude,omitempty"`
	Longitude   float64   `json:"longitude,omitempty"`
	BoundingBox []float64 `json:"bounding_box,omitempty"`
	Address     string    `json:"address,omitempty"`
	MapURL      string    `json:"map_url,omitempty"`
}

func (r *MapResult) Kind() string     { return "map" }
func (r *MapResult) Base() *BaseResult { return &r.BaseResult }
func (r *MapResult) Normalize()        { r.BaseResult.Normalize(); r.PostNormalize() }
func (r *MapResult) PostNormalize()    {}
func (r *MapResult) DedupKey() string  { return dedupKeyFromBase(&r.BaseResult) }

// ---------------------------------------------------------------------------
// MusicResult — music track/album result.
// ---------------------------------------------------------------------------

type MusicResult struct {
	BaseResult
	Artist   string `json:"artist,omitempty"`
	Album    string `json:"album,omitempty"`
	Duration string `json:"duration,omitempty"`
}

func (r *MusicResult) Kind() string     { return "music" }
func (r *MusicResult) Base() *BaseResult { return &r.BaseResult }
func (r *MusicResult) Normalize()        { r.BaseResult.Normalize(); r.PostNormalize() }
func (r *MusicResult) PostNormalize()    {}
func (r *MusicResult) DedupKey() string  { return dedupKeyFromBase(&r.BaseResult) }

// ---------------------------------------------------------------------------
// AnswerResult — direct answer result.
// ---------------------------------------------------------------------------

type AnswerResult struct {
	BaseResult
	Answer string `json:"answer"`
}

func (r *AnswerResult) Kind() string     { return "answer" }
func (r *AnswerResult) Base() *BaseResult { return &r.BaseResult }
func (r *AnswerResult) Normalize()        { r.BaseResult.Normalize(); r.PostNormalize() }
func (r *AnswerResult) PostNormalize()    {}
func (r *AnswerResult) DedupKey() string  { return dedupKeyFromBase(&r.BaseResult) }

// ---------------------------------------------------------------------------
// KeyValueResult — generic key/value table.
// ---------------------------------------------------------------------------

type KeyValueResult struct {
	BaseResult
	KVMap      map[string]string `json:"kv_map"`
	Caption    string            `json:"caption,omitempty"`
	KeyTitle   string            `json:"key_title,omitempty"`
	ValueTitle string            `json:"value_title,omitempty"`
}

func (r *KeyValueResult) Kind() string     { return "keyvalue" }
func (r *KeyValueResult) Base() *BaseResult { return &r.BaseResult }
func (r *KeyValueResult) Normalize()        { r.BaseResult.Normalize(); r.PostNormalize() }
func (r *KeyValueResult) PostNormalize()    {}
func (r *KeyValueResult) DedupKey() string  { return dedupKeyFromBase(&r.BaseResult) }

// ---------------------------------------------------------------------------
// InfoboxResult — knowledge-panel style result.
// ---------------------------------------------------------------------------

type InfoboxResult struct {
	BaseResult
	InfoboxID     string             `json:"infobox_id,omitempty"`
	Attributes    []InfoboxAttribute `json:"attributes,omitempty"`
	URLs          []InfoboxURL       `json:"urls,omitempty"`
	RelatedTopics []string           `json:"related_topics,omitempty"`
	ImgSrc        string             `json:"img_src,omitempty"`
	ImgAlt        string             `json:"img_alt,omitempty"`
}

func (r *InfoboxResult) Kind() string     { return "infobox" }
func (r *InfoboxResult) Base() *BaseResult { return &r.BaseResult }
func (r *InfoboxResult) Normalize()        { r.BaseResult.Normalize(); r.PostNormalize() }
func (r *InfoboxResult) PostNormalize() {
	if r.InfoboxID == "" && r.URL != "" {
		r.InfoboxID = r.URL
	}
	if r.InfoboxID == "" {
		r.InfoboxID = "infobox:" + r.Title
	}
}
func (r *InfoboxResult) DedupKey() string { return dedupKeyFromBase(&r.BaseResult) }

// ---------------------------------------------------------------------------
// ResultTypes is a convenience type for holding heterogeneous result lists.
// ---------------------------------------------------------------------------

type ResultTypes struct {
	Main      []MainResult      `json:"main,omitempty"`
	Images    []ImageResult     `json:"images,omitempty"`
	Videos    []VideoResult     `json:"videos,omitempty"`
	News      []NewsResult      `json:"news,omitempty"`
	Papers    []PaperResult     `json:"papers,omitempty"`
	Code      []CodeResult      `json:"code,omitempty"`
	Files     []FileResult      `json:"files,omitempty"`
	Maps      []MapResult       `json:"maps,omitempty"`
	Music     []MusicResult     `json:"music,omitempty"`
	Answers   []AnswerResult    `json:"answers,omitempty"`
	Infoboxes []InfoboxResult   `json:"infoboxes,omitempty"`
}

// ---------------------------------------------------------------------------
// dedupKeyFromBase generates a deduplication key from a BaseResult.
// ---------------------------------------------------------------------------

func dedupKeyFromBase(br *BaseResult) string {
	u, err := url.Parse(br.URL)
	if err != nil {
		return br.URL + "|" + br.ThumbnailURL
	}
	return u.Host + "|" + u.Path + "|" + u.RawQuery + "|" + br.ThumbnailURL
}
