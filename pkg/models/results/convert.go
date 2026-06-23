package results

import (
	"time"

	"github.com/seargo/seargo/pkg/models"
)

// kindTemplate returns the default template string for a given result kind.
func kindTemplate(kind string) string {
	switch kind {
	case "main":
		return "default.html"
	case "image":
		return "images.html"
	case "video":
		return "videos.html"
	case "news":
		return "default.html"
	case "paper":
		return "paper.html"
	case "code":
		return "code.html"
	case "file":
		return "files.html"
	case "map":
		return "map.html"
	case "music":
		return "music.html"
	case "answer":
		return "answer/legacy.html"
	case "keyvalue":
		return "keyvalue.html"
	case "infobox":
		return "infobox.html"
	default:
		return "default.html"
	}
}

// syncThumbnailURL syncs the ThumbnailURL field from type-specific thumbnail fields.
func syncThumbnailURL(r Result, br *BaseResult) {
	switch t := r.(type) {
	case *ImageResult:
		if br.ThumbnailURL == "" && t.ThumbnailSrc != "" {
			br.ThumbnailURL = t.ThumbnailSrc
		}
	case *VideoResult:
		if br.ThumbnailURL == "" && t.Thumbnail != "" {
			br.ThumbnailURL = t.Thumbnail
		}
	}
}

// ToAPIResult converts a slice of typed Result values into the tagged-union
// models.Result format used by the API response pipeline.
func ToAPIResult(typed []Result) []models.Result {
	if len(typed) == 0 {
		return nil
	}

	out := make([]models.Result, 0, len(typed))
	for _, r := range typed {
		kind := r.Kind()
		br := r.Base()

		// Sync thumbnail from type-specific fields before copying.
		syncThumbnailURL(r, br)

		api := models.Result{
			Kind:     kind,
			Template: kindTemplate(kind),
			Title:    r.GetTitle(),
			URL:      r.GetURL(),
			Content:  r.GetContent(),
			Engine:   r.GetEngine(),
			Category: models.Category(r.GetCategory()),
		}

		// Copy BaseResult fields
		api.Engines = br.Engines
		api.Positions = br.Positions
		api.Score = br.Score
		api.ThumbnailURL = br.ThumbnailURL
		api.PublishedAt = br.PublishedAt
		api.Domain = br.Domain
		api.Favicon = br.Favicon
		api.Extra = buildExtra(r, br)

		out = append(out, api)
	}

	return out
}

// buildExtra builds the Extra map for the tagged-union API result.
func buildExtra(r Result, br *BaseResult) map[string]any {
	ed := make(map[string]any)

	switch t := r.(type) {
	case *ImageResult:
		ed["img_src"] = t.ImgSrc
		ed["thumbnail_src"] = t.ThumbnailSrc
		ed["resolution"] = t.Resolution
		ed["img_format"] = t.ImgFormat
		ed["source"] = t.Source
		ed["width"] = t.Width
		ed["height"] = t.Height
		ed["file_size"] = t.FileSize

	case *VideoResult:
		ed["thumbnail"] = t.Thumbnail
		ed["iframe_src"] = t.IFrameSrc
		ed["length"] = t.Length
		ed["duration"] = t.Duration
		ed["author"] = t.Author
		ed["upload_date"] = t.UploadDate
		ed["view_count"] = t.ViewCount

	case *PaperResult:
		ed["doi"] = t.DOI
		ed["journal"] = t.Journal
		ed["authors"] = t.Authors
		ed["publisher"] = t.Publisher
		ed["type"] = t.Type
		ed["pdf_url"] = t.PDFURL
		ed["html_url"] = t.HTMLURL
		ed["issn"] = t.ISSN
		ed["isbn"] = t.ISBN
		ed["pages"] = t.Pages
		ed["tags"] = t.Tags

	case *CodeResult:
		ed["repository"] = t.Repository
		ed["code_language"] = t.CodeLanguage
		ed["filename"] = t.Filename
		if len(t.CodeLines) > 0 {
			ed["code_lines"] = t.CodeLines
		}
		if len(t.HLLines) > 0 {
			ed["hl_lines"] = t.HLLines
		}

	case *FileResult:
		ed["file_type"] = t.FileType
		ed["file_size"] = t.FileSize
		ed["filename"] = t.Filename
		ed["magnet_uri"] = t.MagnetURI
		if t.Seeders > 0 {
			ed["seeders"] = t.Seeders
		}
		if t.Leechers > 0 {
			ed["leechers"] = t.Leechers
		}

	case *MapResult:
		ed["latitude"] = t.Latitude
		ed["longitude"] = t.Longitude
		ed["map_url"] = t.MapURL
		ed["address"] = t.Address
		if len(t.BoundingBox) > 0 {
			ed["bounding_box"] = t.BoundingBox
		}

	case *MusicResult:
		ed["artist"] = t.Artist
		ed["album"] = t.Album
		ed["duration"] = t.Duration

	case *AnswerResult:
		ed["answer"] = t.Answer

	case *KeyValueResult:
		ed["kv_map"] = t.KVMap
		ed["caption"] = t.Caption
		ed["key_title"] = t.KeyTitle
		ed["value_title"] = t.ValueTitle

	case *InfoboxResult:
		ed["infobox_id"] = t.InfoboxID
		if len(t.Attributes) > 0 {
			ed["attributes"] = t.Attributes
		}
		if len(t.URLs) > 0 {
			ed["urls"] = t.URLs
		}
		if len(t.RelatedTopics) > 0 {
			ed["related_topics"] = t.RelatedTopics
		}
		ed["img_src"] = t.ImgSrc
		ed["img_alt"] = t.ImgAlt
	}

	// Copy engine data for types without kind-specific extras (MainResult, NewsResult, etc.)
	if br != nil && br.EngineData != nil && len(ed) == 0 {
		for k, v := range br.EngineData {
			ed[k] = v
		}
	}

	if len(ed) == 0 {
		return nil
	}
	return ed
}

// WrapAPIMainResult converts a flat models.Result (from engines that haven't
// been migrated yet) into a typed *MainResult.
func WrapAPIMainResult(api models.Result) *MainResult {
	tmpl := api.Template
	if tmpl == "" {
		tmpl = "default.html"
	}
	return &MainResult{
		BaseResult: BaseResult{
			Title:        api.Title,
			URL:          api.URL,
			Content:      api.Content,
			Engine:       api.Engine,
			Engines:      api.Engines,
			Template:     tmpl,
			Category:     string(api.Category),
			Positions:    api.Positions,
			Score:        api.Score,
			PublishedAt:  api.PublishedAt,
			ThumbnailURL: api.ThumbnailURL,
			Domain:       api.Domain,
			Favicon:      api.Favicon,
			EngineData:   api.Extra,
		},
	}
}

// Keep the time import alive (used in WrapAPIMainResult for PublishedAt).
var _ = time.Time{}
