package porting

// TierEntry describes a single engine in a porting tier.
type TierEntry struct {
	Name     string // engine name (lowercase, no underscore)
	BaseType string // which base to use: "xpath", "json_engine", "mediawiki", "custom"
	Priority int    // 1-6, lower = higher priority
	Note     string // why this engine is in this tier
}

// AllTiers returns all tiers sorted by priority (ascending).
func AllTiers() []TierEntry {
	return append(append(append(append(append(
		Tier1(),
		Tier2()...),
		Tier3()...),
		Tier4()...),
		Tier5()...),
		Tier6()...)
}

// TotalEngines returns the total engine count across all tiers.
func TotalEngines() int {
	return len(AllTiers())
}

// Tier1 — Already ported or critical general-purpose engines (~6 engines).
func Tier1() []TierEntry {
	return []TierEntry{
		{Name: "google", BaseType: "custom", Priority: 1, Note: "already ported, reference impl"},
		{Name: "bing", BaseType: "custom", Priority: 1, Note: "already ported"},
		{Name: "brave", BaseType: "custom", Priority: 1, Note: "already ported"},
		{Name: "duckduckgo", BaseType: "custom", Priority: 1, Note: "already ported"},
		{Name: "wikipedia", BaseType: "custom", Priority: 1, Note: "already ported"},
		{Name: "yahoo", BaseType: "custom", Priority: 1, Note: "already ported"},
	}
}

// Tier2 — High-traffic general-purpose engines (~20 engines).
func Tier2() []TierEntry {
	return []TierEntry{
		{Name: "bing_images", BaseType: "xpath", Priority: 2, Note: "Bing image search"},
		{Name: "bing_videos", BaseType: "xpath", Priority: 2, Note: "Bing video search"},
		{Name: "google_images", BaseType: "xpath", Priority: 2, Note: "Google image search"},
		{Name: "google_videos", BaseType: "xpath", Priority: 2, Note: "Google video search"},
		{Name: "google_news", BaseType: "xpath", Priority: 2, Note: "Google news search"},
		{Name: "bing_news", BaseType: "xpath", Priority: 2, Note: "Bing news search"},
		{Name: "duckduckgo_images", BaseType: "json_engine", Priority: 2, Note: "DuckDuckGo image API"},
		{Name: "qwant", BaseType: "json_engine", Priority: 2, Note: "Qwant API"},
		{Name: "startpage", BaseType: "xpath", Priority: 2, Note: "Startpage HTML"},
		{Name: "mojeek", BaseType: "xpath", Priority: 2, Note: "Mojeek search"},
		{Name: "searx_engine", BaseType: "json_engine", Priority: 2, Note: "SearXNG instances"},
		{Name: "wiby", BaseType: "xpath", Priority: 2, Note: "Wiby retro search"},
		{Name: "mwmbl", BaseType: "json_engine", Priority: 2, Note: "Mwmbl API"},
		{Name: "stract", BaseType: "json_engine", Priority: 2, Note: "Stract API"},
		{Name: "brave_news", BaseType: "json_engine", Priority: 2, Note: "Brave news search"},
		{Name: "brave_videos", BaseType: "json_engine", Priority: 2, Note: "Brave video search"},
		{Name: "presearch", BaseType: "xpath", Priority: 2, Note: "Presearch"},
		{Name: "yep", BaseType: "xpath", Priority: 2, Note: "Yep.com"},
		{Name: "crowdview", BaseType: "xpath", Priority: 2, Note: "CrowdView"},
		{Name: "curlie", BaseType: "xpath", Priority: 2, Note: "Curlie directory"},
	}
}

// Tier3 — Specialized engines: images, videos, news, files, science (~50 engines).
func Tier3() []TierEntry {
	return []TierEntry{
		{Name: "deviantart", BaseType: "xpath", Priority: 3, Note: "DeviantArt"},
		{Name: "flickr", BaseType: "json_engine", Priority: 3, Note: "Flickr API"},
		{Name: "unsplash", BaseType: "json_engine", Priority: 3, Note: "Unsplash API"},
		{Name: "wallhaven", BaseType: "json_engine", Priority: 3, Note: "Wallhaven API"},
		{Name: "artic", BaseType: "json_engine", Priority: 3, Note: "Art Institute Chicago"},
		{Name: "openverse", BaseType: "json_engine", Priority: 3, Note: "Openverse"},
		{Name: "library_of_congress", BaseType: "json_engine", Priority: 3, Note: "LoC"},
		{Name: "youtube", BaseType: "json_engine", Priority: 3, Note: "YouTube API"},
		{Name: "vimeo", BaseType: "json_engine", Priority: 3, Note: "Vimeo API"},
		{Name: "dailymotion", BaseType: "json_engine", Priority: 3, Note: "Dailymotion"},
		{Name: "odysee", BaseType: "json_engine", Priority: 3, Note: "Odysee"},
		{Name: "invidious", BaseType: "json_engine", Priority: 3, Note: "Invidious"},
		{Name: "bilibili", BaseType: "json_engine", Priority: 3, Note: "Bilibili"},
		{Name: "arxiv", BaseType: "json_engine", Priority: 3, Note: "arXiv API"},
		{Name: "pubmed", BaseType: "json_engine", Priority: 3, Note: "PubMed"},
		{Name: "crossref", BaseType: "json_engine", Priority: 3, Note: "Crossref"},
		{Name: "openaire", BaseType: "json_engine", Priority: 3, Note: "OpenAIRE"},
		{Name: "semantic_scholar", BaseType: "json_engine", Priority: 3, Note: "Semantic Scholar"},
		{Name: "github", BaseType: "json_engine", Priority: 3, Note: "GitHub code search"},
		{Name: "gitlab", BaseType: "json_engine", Priority: 3, Note: "GitLab"},
		{Name: "sourcehut", BaseType: "xpath", Priority: 3, Note: "SourceHut"},
		{Name: "libgen", BaseType: "xpath", Priority: 3, Note: "Library Genesis"},
		{Name: "annas_archive", BaseType: "json_engine", Priority: 3, Note: "Anna's Archive"},
		{Name: "openstreetmap", BaseType: "json_engine", Priority: 3, Note: "OpenStreetMap"},
		{Name: "apple_maps", BaseType: "json_engine", Priority: 3, Note: "Apple Maps (deferred; needs token)"},
		{Name: "photon", BaseType: "json_engine", Priority: 3, Note: "Photon (Komoot)"},
		{Name: "google_maps", BaseType: "json_engine", Priority: 3, Note: "Google Maps (deferred; needs token)"},
		{Name: "1337x", BaseType: "xpath", Priority: 3, Note: "1337x torrents"},
		{Name: "piratebay", BaseType: "xpath", Priority: 3, Note: "The Pirate Bay"},
		{Name: "nyaa", BaseType: "json_engine", Priority: 3, Note: "Nyaa torrents"},
		{Name: "tokyotoshokan", BaseType: "xpath", Priority: 3, Note: "Tokyo Toshokan"},
		{Name: "solidtorrents", BaseType: "json_engine", Priority: 3, Note: "Solid Torrents"},
		{Name: "btdigg", BaseType: "xpath", Priority: 3, Note: "BTDigg"},
		{Name: "wikicommons_images", BaseType: "mediawiki", Priority: 3, Note: "Wikimedia Commons images"},
		{Name: "wikicommons_videos", BaseType: "mediawiki", Priority: 3, Note: "Wikimedia Commons videos"},
		{Name: "wikicommons_files", BaseType: "mediawiki", Priority: 3, Note: "Wikimedia Commons files"},
		{Name: "wikibooks", BaseType: "mediawiki", Priority: 3, Note: "Wikibooks"},
		{Name: "wikinews", BaseType: "mediawiki", Priority: 3, Note: "Wikinews"},
		{Name: "wikiquote", BaseType: "mediawiki", Priority: 3, Note: "Wikiquote"},
		{Name: "wikisource", BaseType: "mediawiki", Priority: 3, Note: "Wikisource"},
		{Name: "wikiversity", BaseType: "mediawiki", Priority: 3, Note: "Wikiversity"},
		{Name: "wikivoyage", BaseType: "mediawiki", Priority: 3, Note: "Wikivoyage"},
		{Name: "dictzone", BaseType: "xpath", Priority: 3, Note: "DictZone"},
		{Name: "lingva", BaseType: "json_engine", Priority: 3, Note: "Lingva translate"},
		{Name: "mymemory_translated", BaseType: "json_engine", Priority: 3, Note: "MyMemory"},
		{Name: "sepiasearch", BaseType: "json_engine", Priority: 3, Note: "SepiaSearch (PeerTube)"},
		{Name: "rumble", BaseType: "json_engine", Priority: 3, Note: "Rumble"},
		{Name: "apple_app_store", BaseType: "json_engine", Priority: 3, Note: "Apple App Store"},
		{Name: "fdroid", BaseType: "json_engine", Priority: 3, Note: "F-Droid"},
		{Name: "google_play", BaseType: "xpath", Priority: 3, Note: "Google Play"},
	}
}

// Tier4 — Regional / language-specific engines (~50 engines).
func Tier4() []TierEntry {
	return []TierEntry{
		{Name: "baidu", BaseType: "xpath", Priority: 4, Note: "Baidu (Chinese)"},
		{Name: "sogou", BaseType: "xpath", Priority: 4, Note: "Sogou"},
		{Name: "sputnik", BaseType: "xpath", Priority: 4, Note: "Sputnik"},
		{Name: "yandex", BaseType: "xpath", Priority: 4, Note: "Yandex"},
		{Name: "naver", BaseType: "xpath", Priority: 4, Note: "Naver (Korean)"},
		{Name: "daum", BaseType: "xpath", Priority: 4, Note: "Daum"},
		{Name: "goo", BaseType: "xpath", Priority: 4, Note: "goo (Japanese)"},
		{Name: "yahoo_jp", BaseType: "xpath", Priority: 4, Note: "Yahoo Japan"},
		{Name: "seznam", BaseType: "xpath", Priority: 4, Note: "Seznam (Czech)"},
		{Name: "qwant_lite", BaseType: "xpath", Priority: 4, Note: "Qwant Lite"},
		{Name: "duden", BaseType: "xpath", Priority: 4, Note: "Duden (German)"},
		{Name: "leo", BaseType: "xpath", Priority: 4, Note: "LEO dictionary"},
		{Name: "linguee", BaseType: "xpath", Priority: 4, Note: "Linguee"},
		{Name: "ecosia", BaseType: "xpath", Priority: 4, Note: "Ecosia"},
		{Name: "metager", BaseType: "xpath", Priority: 4, Note: "MetaGer"},
		{Name: "swisscows", BaseType: "xpath", Priority: 4, Note: "Swisscows"},
		{Name: "kagi", BaseType: "json_engine", Priority: 4, Note: "Kagi (needs API key)"},
		{Name: "marginalia", BaseType: "json_engine", Priority: 4, Note: "Marginalia"},
		{Name: "alexandria", BaseType: "json_engine", Priority: 4, Note: "Alexandria"},
		{Name: "rightdao", BaseType: "json_engine", Priority: 4, Note: "Right Dao"},
		{Name: "seekr", BaseType: "xpath", Priority: 4, Note: "Seekr"},
		{Name: "andisearch", BaseType: "json_engine", Priority: 4, Note: "AndiSearch"},
		{Name: "searchmysite", BaseType: "json_engine", Priority: 4, Note: "SearchMySite"},
		{Name: "filmweb", BaseType: "xpath", Priority: 4, Note: "Filmweb"},
		{Name: "imdb", BaseType: "xpath", Priority: 4, Note: "IMDb"},
		{Name: "tmdb", BaseType: "json_engine", Priority: 4, Note: "TMDB API"},
		{Name: "genius", BaseType: "json_engine", Priority: 4, Note: "Genius lyrics"},
		{Name: "bandcamp", BaseType: "json_engine", Priority: 4, Note: "Bandcamp"},
		{Name: "soundcloud", BaseType: "json_engine", Priority: 4, Note: "SoundCloud"},
		{Name: "invidious_music", BaseType: "json_engine", Priority: 4, Note: "Invidious music"},
		{Name: "mixcloud", BaseType: "xpath", Priority: 4, Note: "Mixcloud"},
		{Name: "discogs", BaseType: "json_engine", Priority: 4, Note: "Discogs API"},
		{Name: "reddit", BaseType: "json_engine", Priority: 4, Note: "Reddit search"},
		{Name: "hackernews", BaseType: "json_engine", Priority: 4, Note: "Hacker News (Algolia)"},
		{Name: "lobsters", BaseType: "json_engine", Priority: 4, Note: "Lobsters"},
		{Name: "stackoverflow", BaseType: "json_engine", Priority: 4, Note: "Stack Exchange API"},
		{Name: "askubuntu", BaseType: "json_engine", Priority: 4, Note: "Ask Ubuntu"},
		{Name: "superuser", BaseType: "json_engine", Priority: 4, Note: "Super User"},
		{Name: "docker_hub", BaseType: "json_engine", Priority: 4, Note: "Docker Hub"},
		{Name: "pypi", BaseType: "json_engine", Priority: 4, Note: "PyPI"},
		{Name: "npm", BaseType: "json_engine", Priority: 4, Note: "npm registry"},
		{Name: "crates_io", BaseType: "json_engine", Priority: 4, Note: "crates.io"},
		{Name: "packagist", BaseType: "json_engine", Priority: 4, Note: "Packagist"},
		{Name: "hoogle", BaseType: "json_engine", Priority: 4, Note: "Hoogle (Haskell)"},
		{Name: "chefkoch", BaseType: "xpath", Priority: 4, Note: "Chefkoch"},
		{Name: "wolframalpha", BaseType: "json_engine", Priority: 4, Note: "Wolfram Alpha (needs API key)"},
		{Name: "wikipedia_mini", BaseType: "mediawiki", Priority: 4, Note: "Various small Wikipedias"},
		{Name: "apple_music", BaseType: "json_engine", Priority: 4, Note: "Apple Music API"},
		{Name: "spotify", BaseType: "json_engine", Priority: 4, Note: "Spotify (needs API key)"},
		{Name: "deezer", BaseType: "json_engine", Priority: 4, Note: "Deezer"},
		{Name: "lastfm", BaseType: "json_engine", Priority: 4, Note: "Last.fm API"},
	}
}

// Tier5 — Niche / special-purpose engines (~50 engines).
func Tier5() []TierEntry {
	return []TierEntry{
		{Name: "bt4g", BaseType: "xpath", Priority: 5, Note: "BT4G"},
		{Name: "acgsou", BaseType: "xpath", Priority: 5, Note: "ACGSou"},
		{Name: "tokyotoshokan_images", BaseType: "xpath", Priority: 5, Note: "Tokyo Toshokan"},
		{Name: "kickass", BaseType: "xpath", Priority: 5, Note: "Kickass Torrents"},
		{Name: "limetorrents", BaseType: "xpath", Priority: 5, Note: "LimeTorrents"},
		{Name: "torlock", BaseType: "xpath", Priority: 5, Note: "TorLock"},
		{Name: "zoonomaly", BaseType: "xpath", Priority: 5, Note: "Zoonomaly"},
		{Name: "ahmia", BaseType: "xpath", Priority: 5, Note: "Ahmia (dark web)"},
		{Name: "abbreviations", BaseType: "xpath", Priority: 5, Note: "Abbreviations.com"},
		{Name: "alpinelinux", BaseType: "xpath", Priority: 5, Note: "Alpine Linux packages"},
		{Name: "archlinux", BaseType: "xpath", Priority: 5, Note: "Arch Linux packages"},
		{Name: "ask", BaseType: "xpath", Priority: 5, Note: "Ask.com"},
		{Name: "deepl", BaseType: "json_engine", Priority: 5, Note: "DeepL (proprietary)"},
		{Name: "etsy", BaseType: "xpath", Priority: 5, Note: "Etsy"},
		{Name: "ebay", BaseType: "xpath", Priority: 5, Note: "eBay"},
		{Name: "google_scholar", BaseType: "xpath", Priority: 5, Note: "Google Scholar"},
		{Name: "habrahabr", BaseType: "xpath", Priority: 5, Note: "Habrahabr"},
		{Name: "internet_archive", BaseType: "json_engine", Priority: 5, Note: "Internet Archive"},
		{Name: "jisho", BaseType: "json_engine", Priority: 5, Note: "Jisho (Japanese dict)"},
		{Name: "library_thing", BaseType: "xpath", Priority: 5, Note: "LibraryThing"},
		{Name: "mdn", BaseType: "xpath", Priority: 5, Note: "MDN Web Docs"},
		{Name: "openlibrary", BaseType: "json_engine", Priority: 5, Note: "OpenLibrary"},
		{Name: "pdbe", BaseType: "json_engine", Priority: 5, Note: "PDBe"},
		{Name: "peertube", BaseType: "json_engine", Priority: 5, Note: "PeerTube instances"},
		{Name: "piped", BaseType: "json_engine", Priority: 5, Note: "Piped (YouTube proxy)"},
		{Name: "pornhub", BaseType: "xpath", Priority: 5, Note: "Pornhub"},
		{Name: "redtube", BaseType: "xpath", Priority: 5, Note: "RedTube"},
		{Name: "xvideos", BaseType: "xpath", Priority: 5, Note: "XVideos"},
		{Name: "youporn", BaseType: "xpath", Priority: 5, Note: "YouPorn"},
		{Name: "rumble_videos", BaseType: "json_engine", Priority: 5, Note: "Rumble videos"},
		{Name: "presearch_videos", BaseType: "json_engine", Priority: 5, Note: "Presearch videos"},
		{Name: "bpb", BaseType: "xpath", Priority: 5, Note: "BPB"},
		{Name: "gpodder", BaseType: "json_engine", Priority: 5, Note: "gPodder"},
		{Name: "mediathekviewweb", BaseType: "json_engine", Priority: 5, Note: "MediathekViewWeb"},
		{Name: "radio_browser", BaseType: "json_engine", Priority: 5, Note: "Radio Browser"},
		{Name: "rumble_channel", BaseType: "json_engine", Priority: 5, Note: "Rumble channels"},
		{Name: "tineye", BaseType: "json_engine", Priority: 5, Note: "TinEye reverse image"},
		{Name: "wordnik", BaseType: "json_engine", Priority: 5, Note: "Wordnik"},
		{Name: "z_library", BaseType: "xpath", Priority: 5, Note: "Z-Library"},
		{Name: "curlie_images", BaseType: "xpath", Priority: 5, Note: "Curlie images"},
		{Name: "encyclopedia_britannica", BaseType: "xpath", Priority: 5, Note: "Britannica"},
		{Name: "freesound", BaseType: "json_engine", Priority: 5, Note: "Freesound"},
		{Name: "google_docs", BaseType: "xpath", Priority: 5, Note: "Google Docs search"},
		{Name: "google_pdf", BaseType: "xpath", Priority: 5, Note: "Google PDF search"},
		{Name: "material_icons", BaseType: "xpath", Priority: 5, Note: "Material Icons"},
		{Name: "svg_repo", BaseType: "xpath", Priority: 5, Note: "SVG Repo"},
		{Name: "tagesschau", BaseType: "json_engine", Priority: 5, Note: "Tagesschau"},
		{Name: "wikidata", BaseType: "json_engine", Priority: 5, Note: "Wikidata"},
		{Name: "wiktionary", BaseType: "mediawiki", Priority: 5, Note: "Wiktionary"},
		{Name: "metacritic", BaseType: "xpath", Priority: 5, Note: "Metacritic"},
		{Name: "producthunt", BaseType: "xpath", Priority: 5, Note: "Product Hunt"},
		{Name: "alternativeto", BaseType: "xpath", Priority: 5, Note: "AlternativeTo"},
	}
}

// Tier6 — .onion engines + extremely niche (~30 engines).
func Tier6() []TierEntry {
	return []TierEntry{
		{Name: "1337x_onion", BaseType: "xpath", Priority: 6, Note: "1337x .onion"},
		{Name: "nyaa_onion", BaseType: "json_engine", Priority: 6, Note: "Nyaa .onion"},
		{Name: "torlock_onion", BaseType: "xpath", Priority: 6, Note: "TorLock .onion"},
		{Name: "kickass_onion", BaseType: "xpath", Priority: 6, Note: "Kickass .onion"},
		{Name: "piratebay_onion", BaseType: "xpath", Priority: 6, Note: "PirateBay .onion"},
		{Name: "btdigg_onion", BaseType: "xpath", Priority: 6, Note: "BTDigg .onion"},
		{Name: "dicausa", BaseType: "xpath", Priority: 6, Note: "DicaUSA"},
		{Name: "bacon", BaseType: "xpath", Priority: 6, Note: "Bacon"},
		{Name: "tlws", BaseType: "xpath", Priority: 6, Note: "TLWS"},
		{Name: "scanr_structures", BaseType: "json_engine", Priority: 6, Note: "scanR"},
		{Name: "voidlinux", BaseType: "xpath", Priority: 6, Note: "Void Linux packages"},
		{Name: "gentoo", BaseType: "xpath", Priority: 6, Note: "Gentoo packages"},
		{Name: "searchcode_code", BaseType: "json_engine", Priority: 6, Note: "Searchcode"},
		{Name: "searchcode_doc", BaseType: "json_engine", Priority: 6, Note: "Searchcode docs"},
		{Name: "sepiasearch_music", BaseType: "json_engine", Priority: 6, Note: "SepiaSearch music"},
		{Name: "wttr", BaseType: "json_engine", Priority: 6, Note: "wttr.in weather"},
		{Name: "presearch_images", BaseType: "json_engine", Priority: 6, Note: "Presearch images"},
		{Name: "yacy", BaseType: "json_engine", Priority: 6, Note: "YaCy"},
		{Name: "yep_images", BaseType: "json_engine", Priority: 6, Note: "Yep images"},
		{Name: "yep_news", BaseType: "json_engine", Priority: 6, Note: "Yep news"},
		{Name: "zlibrary_onion", BaseType: "xpath", Priority: 6, Note: "Z-Library .onion"},
		{Name: "mankier", BaseType: "xpath", Priority: 6, Note: "Mankier man pages"},
	}
}
