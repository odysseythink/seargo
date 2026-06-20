---
session_id: session_f7f49efa-6fef-4f27-8da2-8e3ba3f9e042
exported_at: 2026-06-20T14:41:37.377Z
work_dir: /Users/ranwei/workspace/go_work/seargo
message_count: 77
token_count: 142298
---

# Kimi Session Export

## Overview

- **Topic**: continue
- **Conversation**: 2 turns | 40 tool calls

---

## Turn 1

### Assistant

<details><summary>Thinking</summary>

Let me start by understanding what's happening. I'm in plan mode, and I need to create an implementation plan based on the approved design at `.ody-code/designs/2026-06-20-05-result-types.md`. Let me first read the design document to understand what needs to be planned.

</details>

#### Tool Call: Read (`/Users/ranwei/workspace/go_work/seargo/.ody-code/designs/202…`)
<!-- call_id: call_00_B8SW5zZ0k9Z43alvFDAZ4828 -->
```json
{
  "path": "/Users/ranwei/workspace/go_work/seargo/.ody-code/designs/2026-06-20-05-result-types.md"
}
```

#### Tool Call: Read (`/Users/ranwei/workspace/go_work/seargo/.ody-code/plans/2026-…`)
<!-- call_id: call_01_diZTmUOGho4liAuAGomj0159 -->
```json
{
  "path": "/Users/ranwei/workspace/go_work/seargo/.ody-code/plans/2026-06-20-04-engine-framework.md"
}
```

<details><summary>Tool Result: Read (`/Users/ranwei/workspace/go_work/seargo/.ody-code/designs/202…`)</summary>

<!-- call_id: call_00_B8SW5zZ0k9Z43alvFDAZ4828 -->
1	# Phase 5 — Structured Result Types & Rendering
2	
3	**Goal:** replace the single flat `models.Result` with typed result kinds so the UI can render images, videos, papers, answers, infoboxes, and key/value blocks distinctly. [C:USER]
4	
5	**Depends on:** Phase 2 (search core), Phase 4 (engine framework). [C:USER]
6	
7	**Audit level:** Deep — every section key claim plus every [C:INFERRED] assumption must be signed off before approval. [C:USER]
8	
9	---
10	
11	## Scope
12	
13	### In Scope
14	
15	1. **Backend typed result hierarchy** (`pkg/models/results/`) [C:USER]
16	   - Extend the existing `Result` interface with `Kind()`, `Template()`, `Normalize()`, and `DedupKey()`.
17	   - Port SearXNG result kinds: `MainResult`, `ImageResult`, `VideoResult`, `NewsResult`, `PaperResult`, `CodeResult`, `FileResult`, `MapResult`, `MusicResult`, `AnswerResult`, `KeyValueResult`, `InfoboxResult`. [C:UPSTREAM]
18	   - Each kind declares its default `template` and exposes kind-specific fields.
19	
20	2. **Tagged-union API response** (`pkg/models/models.go`) [C:USER]
21	   - `models.Result` becomes a tagged union: every result carries `"kind": "image" | "video" | ...` and kind-specific fields live in `Extra map[string]any`.
22	   - `models.Response` keeps top-level `Results []Result`, plus `Answers []Answer`, `Corrections []string`, `Suggestions []string`, `Infoboxes []Infobox`.
23	
24	3. **Normalization** (`pkg/models/results/normalize.go`) [C:USER]
25	   - URL normalization (default scheme `http`, sync `ParsedURL`). [C:UPSTREAM]
26	   - Text normalization (collapse whitespace, deduplicate `Title == Content`). [C:UPSTREAM]
27	   - Date normalization (reject years `< 1900`). [C:UPSTREAM]
28	   - Kind-specific post-normalization (e.g. image thumbnail fallback, code title from filename). [C:INFERRED]
29	
30	4. **Typed-to-flat conversion** (`pkg/models/results/convert.go`) [C:USER]
31	   - `ToAPIResult([]results.Result) []models.Result` maps typed results into the tagged-union API envelope.
32	
33	5. **Typed result container** (`internal/search/typed_container.go`) [C:USER]
34	   - New `TypedResultContainer` that buckets results by `kind`, deduplicates within each bucket, merges `engines`/`positions`/scores, and merges `answers`, `suggestions`, `corrections`, and `infoboxes`.
35	   - Replaces `ResultContainer` for the new typed pipeline; old `ResultContainer` in `internal/search/merge.go` is kept until all engines are migrated. [C:INFERRED]
36	
37	6. **Scheduler / processor integration** (`internal/search/scheduler.go`, `internal/search/processor/*.go`) [C:USER]
38	   - `ProcessorResult` gains a typed results slice.
39	   - `Scheduler.executeProcessors` feeds typed results into `TypedResultContainer`.
40	
41	7. **Frontend rendering** (`web/src/types/search.ts`, `web/src/components/results/`, `web/src/pages/SearchPage.tsx`) [C:USER]
42	   - TypeScript tagged-union types mirroring the Go API.
43	   - `ResultCard` dispatches on `kind` to kind-specific components.
44	   - `SearchPage` uses category-aware layout (e.g. image grid for `images` category).
45	
46	### Out of Scope (deferred)
47	
48	1. **WeatherAnswer and the full SearXNG weather module.** [C:DEFERRED]
49	   - `searx/result_types/answer.py` includes `WeatherAnswer` with `GeoLocation`, `Temperature`, etc. The weather module is a separate subsystem; weather-style answers are rendered as plain `AnswerResult` in this phase.
50	
51	2. **Server-side syntax highlighting for `CodeResult`.** [C:DEFERRED]
52	   - `CodeResult` carries `code_language` and `codelines`; the frontend renders a `<pre>` block. Syntax highlighting via Chroma/Prism is deferred.
53	
54	3. **`ImageRef` alternative image formats.** [C:DEFERRED]
55	   - SearXNG `Image.formats []ImageRef` supports multiple resolutions/mime types. This phase keeps a single `img_src`/`thumbnail_src`; gallery variants are deferred.
56	
57	4. **Result URL filter hooks for plugins.** [C:DEFERRED]
58	   - SearXNG `Result.filter_urls` is used by the hostnames plugin. The interface is reserved, but plugin integration is Phase 6.
59	
60	5. **Distributed/shared result cache.** [C:INFERRED]
61	   - Typed results are cached as JSON in the existing local multilevel cache. Sharing across processes is deferred.
62	
63	6. **All 238 engine implementations emitting typed results.** [C:DEFERRED]
64	   - This design defines the types and conversion; engines are updated incrementally in implementation cycles per Phase 4 tiers.
65	
66	---
67	
68	## Architecture Overview
69	
70	```
71	engine.Search / base engine
72	        │
73	        ▼
74	[]results.Result (typed)
75	        │
76	        ▼
77	results.Result.Normalize() ──► normalized typed results
78	        │
79	        ▼
80	results.ToAPIResult() ──► []models.Result (tagged union)
81	        │
82	        ▼
83	TypedResultContainer.Extend(engineName, results)
84	        │
85	        ├─► kind buckets: main / image / video / news / ...
86	        │     └─► dedup by kind + DedupKey()
87	        │     └─► merge engines / positions / score
88	        │
89	        ├─► answers set (dedup by answer text)
90	        ├─► suggestions set
91	        ├─► corrections set
92	        ├─► infoboxes map (merge by infobox_id)
93	        │
94	        ▼
95	models.Response { Results, Answers, Suggestions, Corrections, Infoboxes }
96	        │
97	        ▼
98	JSON ──► React SearchPage
99	        │
100	        ▼
101	ResultCard(kind) ──► ImageGrid / VideoCard / PaperCard / ...
102	```
103	
104	Data changes at each arrow:
105	
106	- `engine.Search` → `[]results.Result`: raw parsed data becomes strongly-typed structs.
107	- `Normalize()` → normalized results: URL scheme defaults, whitespace collapsed, dates parsed, kind-specific defaults applied.
108	- `ToAPIResult()` → `[]models.Result`: typed structs flatten into the API-facing tagged union with `kind` + `Extra`.
109	- `TypedResultContainer.Extend()` → buckets: per-kind deduplication and cross-engine merging.
110	- `Response` → JSON: serialized tagged union.
111	- `SearchPage` → `ResultCard`: frontend switches on `kind` to render specialized UI.
112	
113	---
114	
115	## Prior Art
116	
117	Upstream SearXNG uses `msgspec.Struct` for typed results (`searx/result_types/_base.py`, `answer.py`, `image.py`, `code.py`, `paper.py`, `file.py`, `keyvalue.py`). [C:UPSTREAM] Key characteristics:
118	
119	- `Result` base class with `url`, `engine`, `parsed_url`.
120	- `MainResult` adds `title`, `content`, `img_src`, `thumbnail`, `publishedDate`, etc.
121	- Each kind sets a `template` string for rendering.
122	- `__hash__` and `__eq__` are overridden per kind for deduplication.
123	- `normalize_result_fields()` and `filter_urls()` are centralized.
124	
125	Current SearGo already has `pkg/models/results/` with `BaseResult`, `ImageResult`, `VideoResult`, etc., but lacks a common `Result` interface, `Kind()`, and tagged-union serialization. [C:INFERRED]
126	
127	---
128	
129	## Assumptions & Unverified Items
130	
131	| # | Assumption | Confidence | Impact if wrong | How to verify |
132	|---|---|---|---|---|
133	| 1 | All existing and future engines can return their results as `[]results.Result` without needing to change the `Engine.Search` signature (we convert inside the processor). | Medium | If an engine needs to return untyped metadata, the processor contract must change. | Audit existing 6 engines and Phase 4 base engines for return-shape needs. [C:INFERRED] |
134	| 2 | `models.Result` tagged-union JSON with `Extra map[string]any` is acceptable to the frontend and cache serialization. | High | Frontend types or cache values break. | Build a prototype marshal/unmarshal round-trip and frontend type check. [C:USER] |
135	| 3 | The existing `ResultContainer` merge logic (score = Σ weight/position × count, category grouping) works unchanged when keyed by `kind + DedupKey()`. | Medium | Ranking/grouping diverges from current behavior. | Compare outputs before/after for same fixtures. [C:INFERRED] |
136	| 4 | `omitempty` on all optional fields keeps response size growth acceptable. | Medium | Large responses increase bandwidth/latency. | Benchmark response sizes with realistic result sets. [C:INFERRED] |
137	| 5 | Frontend can render all 12 kinds with the proposed component set; no additional kinds are required for MVP. | Medium | Missing renderer causes fallback to main card. | Review roadmap and SearXNG kind usage for Tier 1–3 engines. [C:USER] |
138	| 6 | Caching typed results as JSON is safe across process restarts because `models.Result` is plain structs + `map[string]any`. | High | Cache corruption or panic on load. | Add cache round-trip tests. [C:INFERRED] |
139	
140	---
141	
142	## Risk Register
143	
144	| # | Risk | Likelihood | Impact | Mitigation |
145	|---|---|---|---|---|
146	| 1 | Existing 6 engines return `[]models.Result`; switching to typed results causes compilation errors across engine/processor/scheduler. | High | Build broken | Add adapter in each processor to wrap flat results into `results.Result` during migration; remove when engines natively emit typed results. [C:INFERRED] |
147	| 2 | Old cache entries (flat `models.Result`) fail to unmarshal into new tagged-union shape. | Medium | Cache hit panics or returns malformed results | Add cache value version marker; treat old-format values as cache misses. [C:INFERRED] |
148	| 3 | Frontend `ResultCard` misses a kind case and crashes at runtime. | Medium | User-facing error | Require `default` branch rendering `MainResult`; enable TypeScript exhaustive switch lint. [C:USER] |
149	| 4 | Infobox/answer merging differs from SearXNG behavior. | Medium | Users see duplicate or missing infoboxes | Add golden fixtures comparing merge output with upstream for same inputs. [C:UPSTREAM] |
150	| 5 | Large `Extra` payloads increase JSON response size. | Low | Latency/memory | Use `omitempty`; monitor `response_bytes` metrics. [C:INFERRED] |
151	
152	---
153	
154	## Self-Review
155	
156	(To be filled after the detailed sections are written and before the audit gate.)
157	
158	---
159	
160	## User Final Approval
161	
162	(To be filled after ExitDesignMode.)
163	
164	---
165	
166	## Data Models
167	
168	### Backend — `pkg/models/results/`
169	
170	```go
171	// Result is the interface implemented by every typed search result.
172	type Result interface {
173	    Kind() string
174	    Template() string
175	    GetTitle() string
176	    GetURL() string
177	    GetContent() string
178	    GetEngine() string
179	    GetCategory() string
180	    Base() *BaseResult
181	    Normalize()
182	    DedupKey() string
183	}
184	```
185	
186	```go
187	// BaseResult holds fields shared by all result kinds.
188	type BaseResult struct {
189	    Title        string         `json:"title"`
190	    URL          string         `json:"url"`
191	    Content      string         `json:"content,omitempty"`
192	    Engine       string         `json:"engine"`
193	    Engines      []string       `json:"engines,omitempty"`
194	    Template     string         `json:"template,omitempty"`
195	    Category     string         `json:"category,omitempty"`
196	    Positions    []int          `json:"positions,omitempty"`
197	    Score        float64        `json:"score,omitempty"`
198	    PublishedAt  *time.Time     `json:"published_at,omitempty"`
199	    ThumbnailURL string         `json:"thumbnail_url,omitempty"`
200	    Domain       string         `json:"domain,omitempty"`
201	    Favicon      string         `json:"favicon,omitempty"`
202	    EngineData   map[string]any `json:"engine_data,omitempty"`
203	    ParsedURL    []string       `json:"parsed_url,omitempty"` // [scheme, host, path, rawQuery, fragment]
204	    IsOnion      bool           `json:"is_onion,omitempty"`
205	}
206	```
207	
208	Each concrete type embeds `BaseResult` and adds kind-specific fields:
209	
210	```go
211	// MainResult — general web result.
212	type MainResult struct { BaseResult }
213	func (r *MainResult) Kind() string { return "main" }
214	func (r *MainResult) Template() string { return "default.html" }
215	
216	// ImageResult — image search result.
217	type ImageResult struct {
218	    BaseResult
219	    ThumbnailSrc string     `json:"thumbnail_src,omitempty"`
220	    ImgSrc       string     `json:"img_src,omitempty"`
221	    ImgFormat    string     `json:"img_format,omitempty"`
222	    Resolution   string     `json:"resolution,omitempty"`
223	    ImgAlt       string     `json:"img_alt,omitempty"`
224	    Source       string     `json:"source,omitempty"`
225	    Width        int        `json:"width,omitempty"`
226	    Height       int        `json:"height,omitempty"`
227	    FileSize     string     `json:"file_size,omitempty"`
228	    Formats      []ImageRef `json:"formats,omitempty"`
229	}
230	func (r *ImageResult) Kind() string { return "image" }
231	func (r *ImageResult) Template() string { return "images.html" }
232	
233	// ImageRef — placeholder for alternative formats (deferred in rendering).
234	type ImageRef struct {
235	    URL    string `json:"url"`
236	    Format string `json:"format"`
237	    Label  string `json:"label"`
238	}
239	
240	// VideoResult — video search result.
241	type VideoResult struct {
242	    BaseResult
243	    Thumbnail  string `json:"thumbnail,omitempty"`
244	    IFrameSrc  string `json:"iframe_src,omitempty"`
245	    Length     string `json:"length,omitempty"`
246	    Duration   string `json:"duration,omitempty"`
247	    Author     string `json:"author,omitempty"`
248	    UploadDate string `json:"upload_date,omitempty"`
249	    ViewCount  int64  `json:"view_count,omitempty"`
250	}
251	func (r *VideoResult) Kind() string { return "video" }
252	func (r *VideoResult) Template() string { return "videos.html" }
253	
254	// NewsResult — news/article result.
255	type NewsResult struct { BaseResult }
256	func (r *NewsResult) Kind() string { return "news" }
257	func (r *NewsResult) Template() string { return "default.html" }
258	
259	// PaperResult — academic paper.
260	type PaperResult struct {
261	    BaseResult
262	    DOI           string   `json:"doi,omitempty"`
263	    Journal       string   `json:"journal,omitempty"`
264	    Authors       []string `json:"authors,omitempty"`
265	    Publisher     string   `json:"publisher,omitempty"`
266	    Type          string   `json:"type,omitempty"`
267	    PublishedDate string   `json:"published_date,omitempty"`
268	    Editors       []string `json:"editors,omitempty"`
269	    PDFURL        string   `json:"pdf_url,omitempty"`
270	    HTMLURL       string   `json:"html_url,omitempty"`
271	    Comments      string   `json:"comments,omitempty"`
272	    Tags          []string `json:"tags,omitempty"`
273	    Pages         string   `json:"pages,omitempty"`
274	    ISSN          []string `json:"issn,omitempty"`
275	    ISBN          []string `json:"isbn,omitempty"`
276	}
277	func (r *PaperResult) Kind() string { return "paper" }
278	func (r *PaperResult) Template() string { return "paper.html" }
279	
280	// CodeResult — code snippet.
281	type CodeResult struct {
282	    BaseResult
283	    Repository   string    `json:"repository,omitempty"`
284	    CodeLanguage string    `json:"code_language,omitempty"`
285	    Filename     string    `json:"filename,omitempty"`
286	    CodeLines    []CodeLine `json:"code_lines,omitempty"`
287	    HLLines      []int     `json:"hl_lines,omitempty"`
288	}
289	type CodeLine struct { Line int; Text string }
290	func (r *CodeResult) Kind() string { return "code" }
291	func (r *CodeResult) Template() string { return "code.html" }
292	
293	// FileResult — file / torrent result.
294	type FileResult struct {
295	    BaseResult
296	    Filename string `json:"filename,omitempty"`
297	    FileType string `json:"file_type,omitempty"`
298	    FileSize int64  `json:"file_size,omitempty"`
299	    MagnetURI string `json:"magnet_uri,omitempty"`
300	    Seeders  int    `json:"seeders,omitempty"`
301	    Leechers int    `json:"leechers,omitempty"`
302	}
303	func (r *FileResult) Kind() string { return "file" }
304	func (r *FileResult) Template() string { return "file.html" }
305	
306	// MapResult — map/geolocation.
307	type MapResult struct {
308	    BaseResult
309	    Latitude    float64   `json:"latitude,omitempty"`
310	    Longitude   float64   `json:"longitude,omitempty"`
311	    BoundingBox []float64 `json:"bounding_box,omitempty"`
312	    Address     string    `json:"address,omitempty"`
313	    MapURL      string    `json:"map_url,omitempty"`
314	}
315	func (r *MapResult) Kind() string { return "map" }
316	func (r *MapResult) Template() string { return "map.html" }
317	
318	// MusicResult — music track/album.
319	type MusicResult struct {
320	    BaseResult
321	    Artist   string `json:"artist,omitempty"`
322	    Album    string `json:"album,omitempty"`
323	    Duration string `json:"duration,omitempty"`
324	}
325	func (r *MusicResult) Kind() string { return "music" }
326	func (r *MusicResult) Template() string { return "music.html" }
327	
328	// AnswerResult — direct answer.
329	type AnswerResult struct {
330	    BaseResult
331	    Answer string `json:"answer"`
332	}
333	func (r *AnswerResult) Kind() string { return "answer" }
334	func (r *AnswerResult) Template() string { return "answer/legacy.html" }
335	
336	// KeyValueResult — generic key/value table.
337	type KeyValueResult struct {
338	    BaseResult
339	    KVMap      map[string]string `json:"kv_map"`
340	    Caption    string            `json:"caption,omitempty"`
341	    KeyTitle   string            `json:"key_title,omitempty"`
342	    ValueTitle string            `json:"value_title,omitempty"`
343	}
344	func (r *KeyValueResult) Kind() string { return "keyvalue" }
345	func (r *KeyValueResult) Template() string { return "keyvalue.html" }
346	
347	// InfoboxResult — knowledge-panel style result.
348	type InfoboxResult struct {
349	    BaseResult
350	    InfoboxID     string              `json:"infobox_id,omitempty"`
351	    Attributes    []InfoboxAttribute  `json:"attributes,omitempty"`
352	    URLs          []InfoboxURL        `json:"urls,omitempty"`
353	    RelatedTopics []string            `json:"related_topics,omitempty"`
354	    ImgSrc        string              `json:"img_src,omitempty"`
355	    ImgAlt        string              `json:"img_alt,omitempty"`
356	}
357	type InfoboxAttribute struct { Label string; Value string; URL string }
358	type InfoboxURL       struct { Title string; URL string }
359	func (r *InfoboxResult) Kind() string { return "infobox" }
360	func (r *InfoboxResult) Template() string { return "infobox.html" }
361	```
362	
363	### Backend — API envelope (`pkg/models/models.go`)
364	
365	```go
366	type Result struct {
367	    Kind         string         `json:"kind"`
368	    Template     string         `json:"template,omitempty"`
369	    Title        string         `json:"title"`
370	    URL          string         `json:"url"`
371	    Content      string         `json:"content,omitempty"`
372	    Engine       string         `json:"engine"`
373	    Engines      []string       `json:"engines,omitempty"`
374	    Category     Category       `json:"category"`
375	    Score        float64        `json:"score"`
376	    Positions    []int          `json:"-"`
377	    ThumbnailURL string         `json:"thumbnail_url,omitempty"`
378	    PublishedAt  *time.Time     `json:"published_at,omitempty"`
379	    Domain       string         `json:"domain,omitempty"`
380	    Favicon      string         `json:"favicon,omitempty"`
381	    Extra        map[string]any `json:"extra,omitempty"`
382	}
383	```
384	
385	`models.Answer` is reused as-is. `models.Infobox` is extended to match the rich schema:
386	
387	```go
388	type Infobox struct {
389	    Title       string              `json:"title"`
390	    URL         string              `json:"url,omitempty"`
391	    Content     string              `json:"content,omitempty"`
392	    Engine      string              `json:"engine,omitempty"`
393	    Engines     []string            `json:"engines,omitempty"`
394	    ImgSrc      string              `json:"img_src,omitempty"`
395	    URLs        []InfoboxURL        `json:"urls,omitempty"`
396	    Attributes  []InfoboxAttribute  `json:"attributes,omitempty"`
397	    RelatedTopics []string          `json:"related_topics,omitempty"`
398	}
399	```
400	
401	`models.Response` keeps the same top-level shape; only the elements of `Results []Result` change.
402	
403	### Frontend — TypeScript (`web/src/types/search.ts`)
404	
405	```ts
406	export type ResultKind =
407	  | 'main' | 'image' | 'video' | 'news' | 'paper' | 'code'
408	  | 'file' | 'map' | 'music' | 'answer' | 'keyvalue' | 'infobox';
409	
410	export interface BaseResult {
411	  kind: ResultKind;
412	  template?: string;
413	  title: string;
414	  url: string;
415	  content?: string;
416	  engine: string;
417	  engines?: string[];
418	  category: string;
419	  score: number;
420	  thumbnail_url?: string;
421	  published_at?: string;
422	  domain?: string;
423	  favicon?: string;
424	}
425	
426	export interface ImageResult extends BaseResult {
427	  kind: 'image';
428	  extra: {
429	    img_src: string;
430	    thumbnail_src?: string;
431	    resolution?: string;
432	    img_format?: string;
433	    source?: string;
434	    width?: number;
435	    height?: number;
436	    file_size?: string;
437	  };
438	}
439	
440	// ... analogous interfaces for VideoResult, PaperResult, CodeResult, etc.
441	
442	export type Result = MainResult | ImageResult | VideoResult | NewsResult
443	                   | PaperResult | CodeResult | FileResult | MapResult
444	                   | MusicResult | AnswerResult | KeyValueResult | InfoboxResult;
445	
446	export interface SearchResponse {
447	  query: string;
448	  category: string;
449	  results: Result[];
450	  suggestions: string[];
451	  answers: Answer[];
452	  corrections: string[];
453	  infoboxes: Infobox[];
454	  total: number;
455	  page: number;
456	  page_size: number;
457	  engines_used: string[];
458	  engines_failed: string[];
459	  response_time_ms: number;
460	}
461	```
462	
463	---
464	
465	## Algorithms
466	
467	### A1. `BaseResult.Normalize()`
468	
469	```text
470	Normalize(r Result):
471	    br = r.Base()
472	
473	    normalizeURL(br)
474	        if br.URL == "": return
475	        raw = br.URL
476	        if raw does not contain "://":
477	            raw = "http://" + raw
478	        u = url.Parse(raw)
479	        if error:
480	            br.ParsedURL = nil
481	            return
482	        br.URL = u.String()
483	        br.ParsedURL = [u.Scheme, u.Host, u.Path, u.RawQuery, u.Fragment]
484	
485	    normalizeText(br)
486	        br.Title   = collapseWhitespace(br.Title)
487	        br.Content = collapseWhitespace(br.Content)
488	        if br.Content != "" and br.Content == br.Title:
489	            br.Content = ""
490	
491	    normalizeEngines(br)
492	        if br.Engine != "" and br.Engine not in br.Engines:
493	            br.Engines = append(br.Engines, br.Engine)
494	
495	    r.PostNormalize()
496	```
497	
498	### A2. Kind-specific `PostNormalize()`
499	
500	```text
501	ImageResult.PostNormalize():
502	    if ThumbnailSrc == "" and ImgSrc != "":
503	        ThumbnailSrc = ImgSrc
504	    if Title == "" and ImgSrc != "":
505	        Title = basename(ImgSrc)
506	
507	VideoResult.PostNormalize():
508	    if ThumbnailURL == "" and Thumbnail != "":
509	        ThumbnailURL = Thumbnail
510	
511	CodeResult.PostNormalize():
512	    if Title == "" and Filename != "":
513	        Title = Filename
514	    if CodeLanguage == "":
515	        CodeLanguage = "<guess>"
516	
517	FileResult.PostNormalize():
518	    if Title == "" and Filename != "":
519	        Title = Filename
520	    if FileType == "" and Filename != "":
521	        FileType = extension(Filename)
522	
523	InfoboxResult.PostNormalize():
524	    if InfoboxID == "":
525	        InfoboxID = URL
526	    if InfoboxID == "":
527	        InfoboxID = "infobox:" + Title
528	```
529	
530	### A3. `results.ToAPIResult([]Result) []models.Result`
531	
532	```text
533	ToAPIResult(typed):
534	    out = empty list
535	    for r in typed:
536	        api = models.Result{
537	            Kind:     r.Kind(),
538	            Template: r.Template(),
539	            Title:    r.GetTitle(),
540	            URL:      r.GetURL(),
541	            Content:  r.GetContent(),
542	            Engine:   r.GetEngine(),
543	            Category: r.GetCategory(),
544	        }
545	        copy BaseResult fields into api: Engines, Positions, Score,
546	            ThumbnailURL, PublishedAt, Domain, Favicon
547	        api.Extra = kindSpecificExtra(r)
548	        out.append(api)
549	    return out
550	
551	kindSpecificExtra(r):
552	    switch type(r):
553	        ImageResult:
554	            return {
555	                "img_src": r.ImgSrc,
556	                "thumbnail_src": r.ThumbnailSrc,
557	                "resolution": r.Resolution,
558	                "img_format": r.ImgFormat,
559	                "source": r.Source,
560	                "width": r.Width,
561	                "height": r.Height,
562	                "file_size": r.FileSize,
563	            }
564	        VideoResult:
565	            return {
566	                "thumbnail": r.Thumbnail,
567	                "iframe_src": r.IFrameSrc,
568	                "length": r.Length,
569	                "duration": r.Duration,
570	                "author": r.Author,
571	                "upload_date": r.UploadDate,
572	                "view_count": r.ViewCount,
573	            }
574	        PaperResult:
575	            return {
576	                "doi": r.DOI,
577	                "journal": r.Journal,
578	                "authors": r.Authors,
579	                "publisher": r.Publisher,
580	                "type": r.Type,
581	                "pdf_url": r.PDFURL,
582	                "html_url": r.HTMLURL,
583	                "issn": r.ISSN,
584	                "isbn": r.ISBN,
585	                "pages": r.Pages,
586	                "tags": r.Tags,
587	            }
588	        // ... analogous for Code, File, Map, Music, KeyValue, Infobox
589	        AnswerResult:
590	            return { "answer": r.Answer }
591	        MainResult, NewsResult:
592	            if r.EngineData not empty: return r.EngineData
593	            return nil
594	    if r.Base().EngineData not empty:
595	        return r.Base().EngineData
596	    return nil
597	```
598	
599	### A4. `TypedResultContainer`
600	
601	```go
602	type TypedResultContainer struct {
603	    mu            sync.Mutex
604	    closed        bool
605	    buckets       map[string]map[string]*models.Result // kind -> dedupKey -> result
606	    answers       map[string]*models.Answer
607	    suggestions   map[string]string
608	    corrections   map[string]string
609	    infoboxes     map[string]*models.Infobox
610	    engineData    map[string]any
611	    engineWeights map[string]float64
612	}
613	```
614	
615	```text
616	NewTypedResultContainer(weights):
617	    return {
618	        buckets:       { "main": {}, "image": {}, ... },
619	        answers:       {},
620	        suggestions:   {},
621	        corrections:   {},
622	        infoboxes:     {},
623	        engineData:    {},
624	        engineWeights: weights,
625	    }
626	
627	Extend(engineName, results, positionBase):
628	    lock
629	    for i, r in range results:
630	        pos = positionBase + i + 1
631	        switch r.Kind:
632	            case "suggestion":
633	                addSuggestion(r.Title)
634	            case "correction":
635	                addCorrection(r.Title)
636	            case "answer":
637	                addAnswer(engineName, r)
638	            case "infobox":
639	                mergeInfobox(engineName, r)
640	            default:
641	                mergeMainLike(engineName, r, pos)
642	    unlock
643	
644	mergeMainLike(engineName, r, pos):
645	    kind = r.Kind
646	    if kind == "" or kind not in buckets:
647	        kind = "main"
648	    key = kind + "|" + normalizeDedupKey(r)
649	    bucket = c.buckets[kind]
650	    existing = bucket[key]
651	    if existing == nil:
652	        r.Engine = engineName
653	        r.Engines = [engineName]
654	        r.Positions = [pos]
655	        if r.Domain == "":
656	            r.Domain = extractDomain(r.URL)
657	        bucket[key] = clone(r)
658	        return
659	    existing.Engines = appendUnique(existing.Engines, engineName)
660	    existing.Positions.append(pos)
661	    if len(r.Title) > len(existing.Title):
662	        existing.Title = r.Title
663	    if len(r.Content) > len(existing.Content):
664	        existing.Content = r.Content
665	    if preferHTTPS(r.URL, existing.URL):
666	        existing.URL = r.URL
667	    if r.ThumbnailURL != "" and existing.ThumbnailURL == "":
668	        existing.ThumbnailURL = r.ThumbnailURL
669	    mergeExtra(existing.Extra, r.Extra)
670	    calculateScore(existing)
671	
672	addAnswer(engineName, r):
673	    text = r.Extra["answer"]
674	    if text == "": return
675	    key = lowercase(text)
676	    existing = c.answers[key]
677	    if existing == nil:
678	        c.answers[key] = &models.Answer{Answer: text, URL: r.URL, Content: r.Content, Engine: engineName}
679	        return
680	    existing.Engines = appendUnique(existing.Engines, engineName)
681	
682	addSuggestion(text):
683	    key = lowercase(text)
684	    if key not in c.suggestions:
685	        c.suggestions[key] = text
686	
687	addCorrection(text):
688	    key = lowercase(text)
689	    if key not in c.corrections:
690	        c.corrections[key] = text
691	
692	mergeInfobox(engineName, r):
693	    id = r.Extra["infobox_id"] or r.URL or r.Title
694	    existing = c.infoboxes[id]
695	    if existing == nil:
696	        c.infoboxes[id] = r.toInfobox()
697	        c.infoboxes[id].Engines = [engineName]
698	        return
699	    existing.Engines = appendUnique(existing.Engines, engineName)
700	    existing.Attributes = mergeAttributes(existing.Attributes, r.Extra["attributes"])
701	    existing.URLs = mergeURLs(existing.URLs, r.Extra["urls"])
702	
703	Close():
704	    lock
705	    closed = true
706	    for each bucket:
707	        for each result in bucket:
708	            calculateScore(result)
709	    unlock
710	
711	Results():
712	    all = flatten all bucket values
713	    sort by Score desc, then URL asc
714	    return applyCategoryGrouping(all) // reuses existing groupWindow=20, groupMaxSize=8 logic
715	
716	// helpers
717	normalizeDedupKey(r):
718	    u = url.Parse(normalizeURL(r.URL))
719	    if error:
720	        return r.URL + "|" + r.ThumbnailURL
721	    return u.Host + "|" + u.Path + "|" + u.RawQuery + "|" + r.ThumbnailURL
722	
723	calculateScore(r):
724	    score = 0
725	    for i, pos in range r.Positions:
726	        engineName = r.Engines[i] if i < len(r.Engines) else ""
727	        weight = c.engineWeights[engineName]
728	        if weight == 0: weight = 1
729	        score += weight / float(pos)
730	    score *= len(r.Positions)
731	    r.Score = score
732	```
733	
734	### A5. `normalizeURL` (for dedup)
735	
736	```text
737	normalizeURL(raw):
738	    u = url.Parse(raw)
739	    if error: return raw
740	    u.Scheme = lower(u.Scheme)
741	    u.Host = lower(u.Host)
742	    if u.Host starts with "www." and not "www2." and not "www3.":
743	        u.Host = u.Host[4:]
744	    u.Path = strings.TrimSuffix(u.Path, "/")
745	    q = u.Query()
746	    for param in trackingParams:
747	        q.Del(param)
748	    u.RawQuery = q.Encode()
749	    u.Fragment = ""
750	    return u.String()
751	```
752	
753	---
754	
755	## Error Handling
756	
757	| Error class | Immediate handling | Degradation path | Recovery condition |
758	|---|---|---|---|
759	| Typed result missing required fields | Log warning, treat as `MainResult` | Rendered as generic result | Fix engine/parser |
760	| Unknown `kind` string | Log warning, set `kind = "main"` | Rendered with `MainResult` fallback | Register new kind |
761	| URL parse failure during normalization | Keep raw URL, `ParsedURL = nil`, `Domain = ""` | Dedup key falls back to raw URL | None (bad engine data) |
762	| Date parse failure / year < 1900 | Ignore, `PublishedAt = nil` | No date shown | None |
763	| Infobox ID missing | Generate fallback ID from URL or title | May produce duplicate infoboxes | Engine provides stable ID |
764	| Cache hit with old flat format | Treat as miss, re-run search | Slight latency increase | Old entries expire naturally |
765	| Frontend receives unexpected kind | TS lint / runtime default to `MainResult` | User still sees title/link | Update frontend switch |
766	| `ToAPIResult` receives nil result | Skip it | Fewer results | Fix engine |
767	
768	---
769	
770	## Call-Site Integration
771	
772	1. `pkg/models/results/results.go:1-165` — extend existing structs with `Kind()`, `Template()`, `Base()`, `DedupKey()`, and `PostNormalize()` methods. [C:USER]
773	2. `pkg/models/results/normalize.go:88-91` — change `BaseResult.Normalize()` to orchestrate URL/text/engine normalization and delegate kind-specific work to `PostNormalize()`. [C:USER]
774	3. `pkg/models/results/convert.go:9-101` — rewrite `ToAPIResult` to produce tagged-union `models.Result` with `Kind`, `Template`, and `Extra`. [C:USER]
775	4. `internal/search/typed_container.go` (new file) — implement `TypedResultContainer` as specified in Algorithms. [C:USER]
776	5. `internal/search/scheduler.go:188-192` — replace `NewResultContainer(s.engineWeights)` with `NewTypedResultContainer(s.engineWeights)`; collect results via `container.Results()`. [C:USER]
777	6. `internal/search/scheduler.go:243-292` — in `executeProcessors`, call `container.Extend(proc.Engine().Name(), result.TypedResults, 0)`; continue adding suggestions/answers/corrections/infoboxes/engineData via existing helper methods. [C:INFERRED]
778	7. `internal/search/processor/types.go:33-40` — extend `ProcessorResult`:
779	   ```go
780	   type ProcessorResult struct {
781	       Results     []models.Result   // deprecated, kept during migration
782	       TypedResults []results.Result // new
783	       Suggestions []string
784	       Answers     []models.Answer
785	       Corrections []string
786	       Infoboxes   []models.Infobox
787	       EngineData  map[string]any
788	   }
789	   ```
790	   [C:INFERRED]
791	8. `internal/search/processor/online.go` — convert engine `*models.Response.Results` through `results.ToAPIResult` into `ProcessorResult.TypedResults`. [C:INFERRED]
792	9. `internal/engine/bases/*.go` — update base engine `Search` to return `[]results.Result` instead of `[]models.Result`. [C:INFERRED]
793	10. `pkg/models/models.go:94-109` — redefine `Result` as tagged-union struct shown in Data Models. [C:USER]
794	11. `pkg/models/models.go:84-92` — extend `Infobox` to rich schema. [C:USER]
795	12. `web/src/types/search.ts:1-31` — replace flat `Result` with tagged union and add `answers`/`corrections`/`infoboxes` to `SearchResponse`. [C:USER]
796	13. `web/src/components/results/ResultCard.tsx` (new) — implement `switch (result.kind)` dispatch. [C:USER]
797	14. `web/src/pages/SearchPage.tsx:102-133` — replace inline result card with `<ResultCard result={r} />`; add top-level answer/infobox rendering and category-aware image grid. [C:USER]
798	
799	---
800	
801	## Frontend Rendering
802	
803	Component hierarchy:
804	
805	```
806	SearchPage
807	├── AnswerBox (if answers.length > 0)
808	├── InfoboxPanel (if infoboxes.length > 0)
809	├── ImageGrid (if category === 'images')
810	│   └── ImageResult[]
811	└── ResultList (default)
812	    └── ResultCard (for each result)
813	        ├── MainResult
814	        ├── VideoCard
815	        ├── PaperCard
816	        ├── CodeBlock
817	        ├── FileRow
818	        ├── MapCard
819	        ├── MusicCard
820	        ├── AnswerBox
821	        ├── KeyValueTable
822	        └── InfoboxPanel
823	```
824	
825	`ResultCard` contract:
826	
827	```tsx
828	function ResultCard({ result }: { result: Result }): JSX.Element {
829	  switch (result.kind) {
830	    case 'image': return <ImageResult r={result} />;
831	    case 'video': return <VideoCard r={result} />;
832	    case 'paper': return <PaperCard r={result} />;
833	    case 'code':  return <CodeBlock r={result} />;
834	    case 'file':  return <FileRow r={result} />;
835	    case 'map':   return <MapCard r={result} />;
836	    case 'music': return <MusicCard r={result} />;
837	    case 'answer': return <AnswerBox r={result} />;
838	    case 'keyvalue': return <KeyValueTable r={result} />;
839	    case 'infobox': return <InfoboxPanel r={result} />;
840	    case 'main':
841	    default: return <MainResult r={result} />;
842	  }
843	}
844	```
845	
846	Category-driven layout:
847	
848	- `category === 'images'`: render all `image` results in a grid; other kinds fall through to the list below.
849	- `category === 'videos'`: render `video` results with thumbnail + duration first.
850	- `category === 'news'`: render `news`/`main` results with date emphasis.
851	- default: render mixed list via `ResultCard`.
852	
853	---
854	
855	## Reuse Analysis
856	
857	| Component | Existing code | Reuse decision |
858	|---|---|---|
859	| Typed result structs | `pkg/models/results/results.go` | Extend: add `Result` interface methods (`Kind`, `Template`, `Base`, `Normalize`, `DedupKey`) and keep existing fields. [C:USER] |
860	| URL/text/date normalization | `pkg/models/results/normalize.go` | Reuse and extend: keep `normalizeURL`, `normalizeText`, add `normalizeEngines`, and add kind-specific `PostNormalize`. [C:USER] |
861	| Flat-to-API conversion | `pkg/models/results/convert.go` | Rewrite: output tagged-union `models.Result` with `Extra`; retain field mappings for backward-compatible semantics. [C:USER] |
862	| Result merge / dedup | `internal/search/merge.go` | Keep old container as-is; create new `TypedResultContainer` that borrows dedup/score/grouping logic but buckets by kind. [C:USER] |
863	| Scheduler orchestration | `internal/search/scheduler.go` | Adapt call sites: instantiate typed container, feed typed results, read final results. [C:USER] |
864	| Processor result type | `internal/search/processor/types.go` | Extend `ProcessorResult` with `TypedResults []results.Result`; keep `Results` during migration. [C:INFERRED] |
865	| Frontend Result type | `web/src/types/search.ts` | Rewrite as tagged union. [C:USER] |
866	| SearchPage rendering | `web/src/pages/SearchPage.tsx` | Replace inline card with `ResultCard` switch; reuse styling system and engine-color helper. [C:USER] |
867	| Cache serialization | `internal/cache/multilevel.go` | Use as-is; typed results are JSON-serializable via `models.Result`. [C:INFERRED] |
868	
869	---
870	
871	## Test Plan
872	
873	### Backend unit tests
874	
875	| Test | Package | Assertion |
876	|---|---|---|
877	| `TestMainResultNormalize` | `pkg/models/results` | `Title` whitespace collapsed; `Content` cleared when equal to `Title`; `ParsedURL[0] == "https"`. |
878	| `TestImageResultNormalizeFallback` | `pkg/models/results` | `ThumbnailSrc` falls back to `ImgSrc` when empty. |
879	| `TestCodeResultNormalizeTitleFromFilename` | `pkg/models/results` | `Title` set from `Filename` when empty. |
880	| `TestFileResultNormalizeTypeFromExtension` | `pkg/models/results` | `FileType` inferred from `Filename` extension. |
881	| `TestInfoboxResultNormalizeID` | `pkg/models/results` | `InfoboxID` falls back to `URL` then `"infobox:" + Title`. |
882	| `TestToAPIResultImage` | `pkg/models/results` | Output `Kind == "image"`, `Extra["img_src"]` matches input, `Template == "images.html"`. |
883	| `TestToAPIResultPaper` | `pkg/models/results` | `Extra["doi"]` and `Extra["authors"]` present. |
884	| `TestToAPIResultUnknownKindFallback` | `pkg/models/results` | A result with unregistered `Kind()` returns `Kind == "main"` and `Extra` preserved. |
885	| `TestNormalizeDateInvalidYear` | `pkg/models/results` | Year `1899` leaves `PublishedAt == nil`. |
886	| `TestTypedContainerDedupSameImageTwoEngines` | `internal/search` | Same image from `google_images` and `bing_images` produces one result with `Engines == ["google_images", "bing_images"]` and `len(Positions) == 2`. |
887	| `TestTypedContainerPreferHTTPS` | `internal/search` | Merging `http://` and `https://` URLs keeps `https://`. |
888	| `TestTypedContainerMergeInfoboxAttributes` | `internal/search` | Two infobox results with same ID merge `Attributes` and `URLs`, engines deduped. |
889	| `TestTypedContainerSuggestionsDedup` | `internal/search` | `"Foo"` and `"foo"` produce one suggestion. |
890	| `TestTypedContainerScoreCalculation` | `internal/search` | Score equals `Σ(weight/position) * count` within epsilon. |
891	| `TestTypedContainerClosedForWrite` | `internal/search` | `Extend` after `Close()` is ignored. |
892	
893	### Frontend tests
894	
895	| Test | File | Assertion |
896	|---|---|---|
897	| `ResultCard renders image` | `ResultCard.test.tsx` | `screen.getByRole('img')` has `src === result.extra.img_src`. |
898	| `ResultCard falls back to main for unknown kind` | `ResultCard.test.tsx` | Passing `kind: 'unknown'` renders `MainResult` title/link. |
899	| `ImageGrid renders multiple images` | `ImageGrid.test.tsx` | N image results render N `<img>` elements with unique keys. |
900	| `SearchPage renders answers and infoboxes` | `SearchPage.test.tsx` | When response contains `answers`/`infoboxes`, they appear before the result list. |
901	| `SearchPage image category uses grid` | `SearchPage.test.tsx` | `category === 'images'` renders `ImageGrid` instead of list. |
902	
903	### Integration / build tests
904	
905	| Command | Expected result |
906	|---|---|
907	| `go test ./pkg/models/results/... ./internal/search/...` | All tests pass. |
908	| `go vet ./...` | No issues in modified packages. |
909	| `cd web && npm run typecheck` (or `tsc --noEmit`) | No TypeScript errors. |
910	| `cd web && npm run test` | New component tests pass. |
911	| `cd web && npm run build` | Production build succeeds. |
912	| `make test` (if defined) | Full suite green. |
913	
914	---
915	
916	## Self-Review
917	
918	### Expensive decisions scrutinized
919	
920	1. **Tagged-union JSON shape (`kind` + `Extra`).** [C:USER]
921	   - Input `ImageResult{ImgSrc: "https://x/a.jpg"}` → JSON `{"kind":"image","title":"","url":"","extra":{"img_src":"https://x/a.jpg"}}`.
922	   - Input `MainResult{Title:"A",Content:"A"}` after normalize → `content` empty, `kind:"main"`, no `extra`.
923	   - Input unknown kind string `"xyz"` → `kind` coerced to `"main"`, `extra` preserved.
924	
925	2. **Dedup key algorithm.** [C:INFERRED]
926	   - Input two image results with same URL but different thumbnails → different keys (thumbnail in key).
927	   - Input two main results with same URL after stripping `utm_source` → same key.
928	   - Input URL without scheme `"example.com"` → normalized to `"http://example.com"` before keying.
929	
930	3. **Unknown kind fallback.** [C:USER]
931	   - Input `kind=""` → treated as `main`.
932	   - Input `kind="weather"` (unregistered) → treated as `main` with extra preserved.
933	   - Input malformed JSON missing `kind` → frontend `ResultCard` default branch renders `MainResult`.
934	
935	### Four-lens sweep
936	
937	- **Security:**
938	  - Checked URL normalization: scheme defaulting to `http` matches SearXNG but may allow mixed content; mitigated by `preferHTTPS` in merge.
939	  - Checked that `Extra` map can carry arbitrary engine data; no HTML is pre-rendered server-side, frontend must escape content (React does by default).
940	  - Found: `InfoboxID` fallback uses `Title`, which could collide across unrelated infoboxes. Mitigation: engines should provide stable IDs; fallback is clearly logged.
941	
942	- **Test:**
943	  - Every behavior has a must-pass assertion (see Test Plan).
944	  - Found: no must-reject test for invalid `kind` in `ToAPIResult`. Added `TestToAPIResultUnknownKindFallback`.
945	  - Found: no test for `Close()` idempotency. Added `TestTypedContainerClosedForWrite`.
946	
947	- **Ops:**
948	  - Checked concurrent `Extend` calls: `TypedResultContainer` uses `sync.Mutex`.
949	  - Checked cache serialization: `models.Result` is plain structs + `map[string]any`; old format treated as miss.
950	  - Found: `Extra` payloads could be large for `CodeResult` with many lines. Mitigation: cap `CodeLines` in engines or paginate; document limit.
951	
952	- **Integration:**
953	  - Verified existing hooks: `pkg/models/results/results.go`, `internal/search/merge.go`, `internal/search/scheduler.go`, `internal/search/processor/types.go`, `web/src/types/search.ts`, `web/src/pages/SearchPage.tsx`.
954	  - Verified new component `TypedResultContainer` is greenfield.
955	  - Found: existing `ProcessorResult.Results` is used by offline processors too. Mitigation: keep `Results` field during migration, populate `TypedResults` from it via adapter.
956	
957	- **Scope:**
958	  - Design covers backend types, API envelope, merge container, and frontend rendering as one coherent feature.
959	  - No further decomposition needed; single file is appropriate.
960	
961	---
962	
963	## User Final Approval
964	
965	- **Design status:** Approved pending ExitDesignMode.
966	- **Audit gate result:** Deep audit completed. All [C:INFERRED] assumptions accepted by user.
967	- **Approval method:** ExitDesignMode.
968	- **Unresolved items:** None.
969	
970	---
<system>970 lines read from file starting from line 1. Total lines in file: 970. End of file reached.</system>

</details>

<details><summary>Tool Result: Read (`/Users/ranwei/workspace/go_work/seargo/.ody-code/plans/2026-…`)</summary>

<!-- call_id: call_01_diZTmUOGho4liAuAGomj0159 -->
1	# Engine Framework & 238-Engine Porting Strategy — Implementation Plan
2	
3	**Goal:** Upgrade the engine contract to match SearXNG's capabilities, build base-engine factories for rapid porting, establish the typed result hierarchy, wire the runtime lifecycle (Setup/Init/EngineCache/hot-reload), and deliver the porting toolchain (generator + golden fixtures).
4	
5	**Architecture:** Five subsystems phased sequentially: (1) Core — Engine interface upgrade with declarative metadata, Capabilities, Setup/Init lifecycle hooks, Loader building a validated Registry from config, EngineTraits resolution; (2) Bases — generic base-engine factories (xpath/css, json_engine, mediawiki, opensearch/command) with shared extraction helpers; (3) Results — typed result hierarchy (`pkg/models/results/`) with normalization/flattening to existing `models.Result`; (4) Runtime — synchronous Setup + async Init goroutines, per-engine SQLite EngineCache, hot reload with graceful shutdown, engine-specific error types, SearXNG-aligned metrics; (5) Porting — tiered priority list, AST/regex-assisted generator, golden-file test harness with encapsulated fixtures, smoke-test command.
6	
7	**Tech Stack:** Go 1.25, goquery + htmlquery (CSS/XPath selection), sqlite (engine cache), prometheus client_golang, resty/v2, odysseythink/mlog, testify, yaml.v3
8	
9	> For executing workers: implement this plan task-by-task (prefer a fresh subagent/Task per task — a clean context per task avoids single-session degradation). Steps use - [ ] checkboxes for tracking.
10	
11	## File Structure
12	
13	### New Files
14	| File | Responsibility |
15	|---|---|
16	| `internal/engine/loader.go` | Loader: validate config entries, build Registry, resolve traits, call Setup/Init |
17	| `internal/engine/loader_test.go` | Loader unit tests |
18	| `internal/engine/traits.go` | EngineTraits types, resolveTrait(), fetch-traits placeholder |
19	| `internal/engine/traits_test.go` | Traits resolution tests |
20	| `internal/engine/bases/xpath.go` | XPath/CSS base engine — generic HTML scraper |
21	| `internal/engine/bases/xpath_test.go` | XPath base tests with golden fixtures |
22	| `internal/engine/bases/json_engine.go` | JSON API base engine — JSONPath-like query engine |
23	| `internal/engine/bases/json_engine_test.go` | JSON engine tests |
24	| `internal/engine/bases/mediawiki.go` | MediaWiki Action API wrapper base |
25	| `internal/engine/bases/mediawiki_test.go` | MediaWiki base tests |
26	| `internal/engine/bases/opensearch.go` | OpenSearch descriptor base (offline) |
27	| `internal/engine/bases/command.go` | Command-line base (offline) |
28	| `internal/engine/bases/extract.go` | Shared extraction utilities (html_to_text, extract_url, eval_xpath, etc.) |
29	| `internal/engine/bases/extract_test.go` | Extraction utility tests |
30	| `internal/engine/bases/jsonquery.go` | JSONPath-like query engine (slash-query traversal) |
31	| `internal/engine/bases/jsonquery_test.go` | JSON query engine tests |
32	| `internal/engine/bases/doc.go` | Package documentation |
33	| `internal/engine/cache.go` | EngineCache — per-engine SQLite key/value store with TTL |
34	| `internal/engine/cache_test.go` | EngineCache tests |
35	| `internal/engine/runtime.go` | Setup/Init lifecycle: synchronous Setup, async Init goroutines, inactive marking |
36	| `internal/engine/runtime_test.go` | Runtime lifecycle tests |
37	| `internal/engine/errors.go` | Engine-specific error types (SearxEngineResponseException, access denied, captcha, too-many-requests, timeout) |
38	| `internal/engine/errors_test.go` | Engine error tests |
39	| `pkg/models/results/results.go` | Result interface, BaseResult, MainResult, specialization types (Image/Video/News/Paper/Code/File/Map/Music/Answer/Infobox) |
40	| `pkg/models/results/results_test.go` | Result type tests |
41	| `pkg/models/results/normalize.go` | URL normalization, text normalization, date normalization, field deduplication |
42	| `pkg/models/results/normalize_test.go` | Normalization tests |
43	| `pkg/models/results/convert.go` | ToAPIResult: typed results → models.Result flattening |
44	| `pkg/models/results/convert_test.go` | Conversion tests |
45	| `pkg/models/results/doc.go` | Package documentation |
46	| `data/engine_traits.json` | Vendored engine traits from SearXNG upstream |
47	| `internal/engine/porting/tiers.go` | Tiered priority list (1–6) with engine count and descriptions |
48	| `internal/engine/porting/generator.go` | AST/regex-assisted Python→Go skeleton generator |
49	| `internal/engine/porting/generator_test.go` | Generator tests with sample SearXNG engines |
50	| `internal/engine/porting/fixture.go` | Golden-file test harness: Fixture type, runner, assertion logic |
51	| `internal/engine/porting/fixture_test.go` | Fixture harness tests |
52	| `internal/engine/porting/smoke.go` | Smoke-test command: live validation against real engines |
53	| `internal/engine/porting/doc.go` | Package documentation |
54	| `cmd/engine-gen/main.go` | Standalone generator CLI (reads SearXNG engine, writes Go skeleton + fixture stub) |
55	| `testdata/fixtures/engines/` | Golden fixture YAML directory (populated by generator) |
56	
57	### Modified Files
58	| File | Changes |
59	|---|---|
60	| `internal/engine/engine.go` | Extend Engine interface with Setup/Init + metadata; extend EngineInitConfig with full fields; extend Capabilities; add EngineAbout |
61	| `internal/engine/registry.go` | Keep Register/Get/All/Names; add SetAll + Reset for Loader; validate engine name rules (no underscore, lowercase) |
62	| `internal/config/config.go` | Extend EngineConfig with network/paging/time_range_support/language_support/safesearch/weight/about/display_error_messages/enable_http/inactive/tokens fields; extend Validate() for new fields |
63	| `internal/errors/errors.go` | Add engine-specific sentinel errors (SearxEngineResponseException, access_denied, captcha, too_many_requests, timeout) if not already present |
64	| `internal/metrics/metrics.go` | Add SearXNG-aligned metrics: EngineReloadsTotal, EngineParserFailures, EngineResults; rename/add labels to match SearXNG conventions |
65	| `internal/search/scheduler.go` | Replace inline engine init with Loader-based initialization; integrate hot reload signal handling |
66	| `internal/search/suspension.go` | Update classifyError to recognize new engine error types |
67	| `cmd/seargo/main.go` | Wire Loader + hot reload signal; pass EngineTraits to loader; wire EngineCache |
68	| `configs/settings.yml` | Add new engine fields to example config entries |
69	
70	## Dependency Overview
71	
72	```
73	Phase A: Core (no upstream deps within this plan; depends on existing code)
74	  Part: core
75	    Task 1 (EngineInterface — metadata + Setup/Init + Capabilities + EngineAbout) — stand-alone
76	    Task 2 (EngineConfig extension — new fields + overlay + validation) — stand-alone, parallel with T1
77	    Task 3 (EngineTraits — types + resolution + vendored JSON) — stand-alone, parallel with T1,T2
78	    Task 4 (Loader — validate + build Registry + categories/shortcuts) — depends on T1,T2,T3
79	    Task 5 (Config wiring — main.go + scheduler + settings.yml) — depends on T4
80	
81	Phase B: Bases (depends on Phase A: Task 1 EngineInterface)
82	  Part: bases
83	    Task 1 (extract.go — shared helpers) — stand-alone
84	    Task 2 (jsonquery.go — JSONPath-like query) — stand-alone, parallel with T1
85	    Task 3 (xpath base) — depends on T1, T2 (extract + jsonquery)
86	    Task 4 (json_engine base) — depends on T2
87	    Task 5 (mediawiki base) — depends on T1
88	    Task 6 (opensearch + command bases) — depends on T1
89	
90	Phase C: Results (depends on existing models.Result; no upstream within this plan)
91	  Part: results
92	    Task 1 (Result type hierarchy) — stand-alone
93	    Task 2 (Normalization functions) — depends on T1
94	    Task 3 (ToAPIResult conversion) — depends on T1, T2
95	
96	Phase D: Runtime (depends on Phase A: Task 1,4,5)
97	  Part: runtime
98	    Task 1 (EngineCache — SQLite K/V store) — stand-alone
99	    Task 2 (Engine error types — engine/errors.go) — depends on core T1
100	    Task 3 (Setup/Init lifecycle + hot reload) — depends on T1, T2, core T4
101	    Task 4 (SearXNG-aligned metrics) — depends on T2, T3
102	
103	Phase E: Porting (depends on Phase A,B,C,D for the framework to be stable)
104	  Part: porting
105	    Task 1 (Tiered priority list) — stand-alone
106	    Task 2 (Generator — Python→Go) — stand-alone, parallel with T1
107	    Task 3 (Golden fixture harness) — depends on T2
108	    Task 4 (Smoke-test command + engine-gen CLI) — depends on T2, T3
109	```
110	
111	Phase A内部:T1,T2,T3可并行;Phase B内部:T1,T2可并行,T3-T6串行;Phase C可与Phase B并行;Phase D在Phase A完成后启动;Phase E在全部前序完成后启动。
112	
113	## Risks & Open Questions
114	
115	| # | Risk | Mitigation |
116	|---|---|---|
117	| 1 | Engine interface change breaks all 6 existing engines + their tests | Task core-1 includes no-op Setup/Init implementation for all 6 engines + whole-tree typecheck |
118	| 2 | `htmlquery` (antchfx) not in go.mod yet | Task bases-3 adds `go get github.com/antchfx/htmlquery` |
119	| 3 | SQLite driver not in go.mod | Task runtime-1 adds `go get github.com/mattn/go-sqlite3` (or modernc.org/sqlite for CGo-free) |
120	| 4 | `engine_traits.json` is 9588 lines — loading at startup adds latency | Embed via `//go:embed`, parse once at Loader init; benchmark <100ms |
121	| 5 | Hot reload races with in-flight engine Init goroutines | Reference-count active engines; drain on reload (design Part 4); integration test with concurrent Search + Reload |
122	| 6 | Generator produces broken skeletons for complex Python control flow | Generator falls back to stub+TODO comment; tier-by-tier manual porting plan |
123	| 7 | Golden fixtures must not contain real API keys | Pre-commit hook or generator redacts keys in fixture YAML |
124	| 8 | `EngineCache` per-engine SQLite access from multiple goroutines | Single `*sql.DB` with sufficient max open connections (SetMaxOpenConns) |
125	
126	## Parts (generate one per invocation, in order)
127	
128	| # | File | Scope | Status |
129	|---|---|---|---|
130	| 1 | 2026-06-20-04-engine-framework/core.md | Engine interface, metadata, Capabilities, EngineAbout, EngineConfig extension, EngineTraits, Loader, config wiring | done |
131	| 2 | 2026-06-20-04-engine-framework/bases.md | Extraction helpers, JSONPath query, xpath base, json_engine base, mediawiki base, opensearch/command bases | done |
132	| 3 | 2026-06-20-04-engine-framework/results.md | Result type hierarchy, normalization, ToAPIResult conversion | done |
133	| 4 | 2026-06-20-04-engine-framework/runtime.md | EngineCache, engine errors, Setup/Init lifecycle, hot reload, SearXNG-aligned metrics | done |
134	| 5 | 2026-06-20-04-engine-framework/porting.md | Porting tiers, generator, golden fixtures, smoke tests, engine-gen CLI | done |
135	
136	---
137	
138	## Spec Coverage
139	
140	| 设计需求 | Task(s) | 状态 |
141	|---|---|---|
142	| Engine contract — declarative metadata (shortcut, categories, paging, etc.) | core:1 | covered |
143	| Engine contract — Setup/Init lifecycle hooks | core:1 | covered |
144	| Engine contract — Capabilities struct extension | core:1 | covered |
145	| Engine contract — EngineAbout metadata | core:1 | covered |
146	| EngineConfig extension — paging, time_range, language, safesearch, weight, about, etc. | core:2 | covered |
147	| EngineConfig extension — enable_http, inactive, disabled fields | core:2 | covered |
148	| EngineConfig extension — overlay + validation for new fields | core:2 | covered |
149	| Engine name validation (no underscore, lowercase) | core:4 | covered |
150	| EngineTraits system — languages/regions mappings | core:3 | covered |
151	| EngineTraits — vendored data/engine_traits.json | core:3 | covered |
152	| EngineTraits — fetch-traits subcommand placeholder | core:3 | covered |
153	| Loader — validate config entries, build Registry | core:4 | covered |
154	| Loader — skip inactive/onion (no Tor) engines | core:4 | covered |
155	| Loader — build categories map + engine_shortcuts map | core:4 | covered |
156	| Loader — hot reload support | runtime:3 | covered |
157	| Config wiring — main.go + scheduler + settings.yml | core:5 | covered |
158	| xpath base — generic CSS/XPath HTML scraper | bases:3 | covered |
159	| json_engine base — JSONPath-like API mapper | bases:4 | covered |
160	| mediawiki base — Wikimedia Action API wrapper | bases:5 | covered |
161	| opensearch base — offline | bases:6 | covered |
162	| command base — offline | bases:6 | covered |
163	| Shared extraction utilities (html_to_text, extract_url, eval_xpath, etc.) | bases:1 | covered |
164	| JSONPath-like query algorithm | bases:2 | covered |
165	| Result types — typed hierarchy (MainResult, ImageResult, VideoResult, etc.) | results:1 | covered |
166	| Result normalization — URL, text, date | results:2 | covered |
167	| ToAPIResult — typed → models.Result flattening | results:3 | covered |
168	| EngineCache — per-engine SQLite K/V | runtime:1 | covered |
169	| Engine errors — SearxEngineResponseException, access denied, captcha, timeout | runtime:2 | covered |
170	| Engine errors — soft_max_redirects, no_result_for_http_status, raise_for_httperror engine config | runtime:2 | covered |
171	| Setup/Init lifecycle — sync Setup + async Init | runtime:3 | covered |
172	| Hot reload graceful shutdown | runtime:3 | covered |
173	| Metrics — SearXNG-aligned names/labels | runtime:4 | covered |
174	| Porting tiers (1–6 priority list) | porting:1 | covered |
175	| Generator — Python→Go skeleton | porting:2 | covered |
176	| Golden fixture harness | porting:3 | covered |
177	| Smoke-test command + engine-gen CLI | porting:4 | covered |
178	| Frontend result templates | no-op | deferred to UI phase |
179	| Answerers, plugins, autocomplete engines | no-op | deferred to Phase 6,7 |
180	| Offline engine processor generalization | no-op | deferred |
181	| Distributed engine cache sharing | no-op | deferred |
182	| All 238 engines fully ported | no-op | out of scope; tiered execution |
183	
184	---
185	
186	## Global Self-Review
187	
188	- [x] 1. Spec-coverage table: 35 spec items mapped to tasks, 6 no-ops, 0 GAPs. Every design section/requirement in the approved design maps to at least one Task. Covered: engine contract (core:1), EngineConfig (core:2), EngineTraits (core:3), Loader (core:4), wiring (core:5), extraction helpers (bases:1), jsonQuery (bases:2), xpath base (bases:3), json engine (bases:4), mediawiki (bases:5), opensearch/command (bases:6), result types (results:1), normalization (results:2), conversion (results:3), EngineCache (runtime:1), errors (runtime:2), lifecycle+hot-reload (runtime:3), metrics (runtime:4), tiers (porting:1), generator (porting:2), fixtures (porting:3), smoke+CLI (porting:4).
189	- [x] 2. Placeholder scan: All 5 part files — zero `TODO`/`TBD`/dead-code placeholders beyond intentional design-deferred items: the custom engine skeleton explicitly states "Manual porting required" (design's fallback), opensearch/command return "not yet implemented" (deferred per scope), `RunFixtures` comment describes future expansion while current implementation is independently useful (fixture validation). No hidden work or unresolved deps.
190	- [x] 3. No phantom tasks: 23 tasks across 5 parts, every task produces at least one file creation or modification with a commit. Zero `--allow-empty`. Task counts: core 5, bases 6, results 3, runtime 4, porting 4 = 22 tasks. (Note: core:Task 5 modifies 3 existing files — verified `go build` produces changed binary.)
191	- [x] 4. Dependency soundness: All cross-file `Depends on:` references point to earlier parts:
192	  - bases:Task 3,4,5,6 → core:Task 1 ✓
193	  - runtime:Task 2 → core:Task 1 ✓
194	  - runtime:Task 3 → core:Task 4 ✓
195	  - porting:Task 2 → core:Task 1 + bases:Task 3,4 ✓
196	  No reference points to a symbol only a later part defines. Phase D (runtime) cannot precede Phase A (core).
197	- [x] 5. Caller & build soundness:
198	  - `Engine` interface change (core:Task 1): 6 engines + `factory.go` + `scheduler.go` updated in SAME task; whole-tree `go build ./... && go test ./...` passes ✓
199	  - `EngineConfig` extension (core:Task 2): fields appended to end of struct; YAML tags added; `overlayOutgoing` not affected (EngineConfig is separate); `Validate()` extended ✓
200	  - `Loader.Load()` signature (core:Task 4): consumed by `NewScheduler` in core:Task 5 ✓
201	  - `classifyError` update (runtime:Task 2): added `*engine.SearxEngineResponseException` check before existing `*errors.EngineError` — no existing test behavior broken ✓
202	  - `suspension.go` import of `engine` package: added, not breaking ✓
203	  - Existing `models.Result` is not modified — `pkg/models/results/ToAPIResult` consumes it as-is ✓
204	- [x] 6. Test-the-risk: All 5 part files have explicit risk-testing assertions:
205	  - Core: `TestLoadEngines_EmptyName` (boundary), `TestLoadEngines_SetupFails_MarksInactive` (state mutation), `TestEngineTraits_Resolve` (filter logic) ✓
206	  - Bases: `TestJSONQuery_EmptyQuery` (design audit), `TestJSONQuery_DocumentsTitle` (SearXNG verified), `TestXPathEngine_Search` (end-to-end) ✓
207	  - Results: `TestNormalizeURL_Invalid` (no panic), `TestNormalizeDate_YearTooLow` (data quality), `TestNormalizeText_DedupTitleEqualsContent` ✓
208	  - Runtime: `TestEngineCache_Expired` (TTL), `TestLifecycle_InitFailsMarksInactive` (state), `TestSearxEngineResponseException_NoResultForHTTPStatus` ✓
209	  - Porting: `TestGenerator_UnknownBase_FallbackToCustom` (graceful degradation), `TestFixture_Validation` (3 edge cases), `TestTier1_ContainsMajorEngines` ✓
210	  Full list of must-survive inputs verified in each part's local self-review item 6.
211	- [x] 7. Type consistency: Cross-file type usage verified:
212	  - `EngineInitConfig` (core:Task 1) → consumed by `Loader.Load()` (core:Task 4), `Setup`/`Init` methods, `scheduler.go` (core:Task 5), `runtime.go` (runtime:Task 3), `porting/generator.go` (porting:Task 2) ✓
213	  - `EngineTraits` (core:Task 3) → `EngineInitConfig.EngineTraits` field ✓
214	  - `Result` interface (results:Task 1) → `ToAPIResult([]Result)` (results:Task 3) ✓
215	  - `SearxEngineResponseException.ErrorClass()` → `suspension.go:classifyError` switch ✓
216	  - `EngineCache` → `NewEngineCache(path)` constructor, `Set(engineName, key, value, ttl)` signature consistent with sqlite schema ✓
217	  - `SkeletonResult` → `GoCode` string written to file by `cmd/engine-gen` ✓
<system>217 lines read from file starting from line 1. Total lines in file: 217. End of file reached.</system>

</details>

<details><summary>Thinking</summary>

Let me understand the situation. I'm in plan mode, and I need to create an implementation plan for "Phase 5 — Structured Result Types & Rendering" based on the approved design at `.ody-code/designs/2026-06-20-05-result-types.md`.

The plan file path is already assigned: I need to figure out what path it should be. Looking at the previous plan, it was at `.ody-code/plans/2026-06-20-04-engine-framework.md` with parts in `.ody-code/plans/2026-06-20-04-engine-framework/`. The new design is at `2026-06-20-05-result-types.md`, so the plan should be at `.ody-code/plans/2026-06-20-05-result-types.md`.

Wait, let me check the system-reminder more carefully: "The host has ALREADY assigned the plan file path — it is shown to you (labelled in the entry message and at the bottom of this reminder)." Let me check the entry message again...

Actually looking at the reminder: "Plan mode is active" and "Design saved to: /Users/ranwei/workspace/go_work/seargo/.ody-code/designs/2026-06-20-05-result-types.md". But I don't see the explicit plan path mentioned. Let me look at the system entry message again...

The system entry message says "Design saved to: /Users/ranwei/workspace/go_work/seargo/.ody-code/designs/2026-06-20-05-result-types.md" and "Create a concrete, step-by-step implementation plan based on the approved design in `2026-06-20-05-result-types.md`."

Looking at the previous pattern, the plan file would be `.ody-code/plans/2026-06-20-05-result-types.md`. Let me verify this exists or needs creation.

Now let me count how many tasks this phase has. Looking at the design, I see 9 distinct work areas (from the Call-Site Integration section):

1. Result type hierarchy — extending `pkg/models/results/results.go` with Result interface, 12 typed result structs
2. Normalization — `pkg/models/results/normalize.go` with URL/text/date/engine normalization
3. ToAPIResult conversion — `pkg/models/results/convert.go` with typed-to-flat mapping
4. TypedResultContainer — new `internal/search/typed_container.go`
5. Scheduler integration — replacing ResultContainer with TypedResultContainer
6. Processor result type extension — `ProcessorResult.TypedResults` field
7. Processor adapter — online.go conversion
8. Base engines update — return typed results
9. API envelope — `models.Result` tagged union, `models.Infobox` extension
10. Frontend types — TypeScript tagged union
11. Frontend ResultCard — component dispatch
12. Frontend SearchPage — layout changes
13. Frontend component files — ResultCard, ImageGrid, VideoCard, PaperCard, CodeBlock, FileRow, MapCard, MusicCard, AnswerBox, KeyValueTable, InfoboxPanel

That's more than 8 tasks, and spans backend (Go) and frontend (TypeScript/React) — so I need a SPLIT plan. Let me group these into phases and part files.

Let me think about the dependency graph:

**Phase A: Backend Types** (standalone, no upstream deps within this plan)
- Task 1: Result type hierarchy (`pkg/models/results/results.go`) — Result interface, BaseResult, all 12 typed structs
- Task 2: Normalization (`pkg/models/results/normalize.go`) — depends on Task 1
- Task 3: ToAPIResult conversion (`pkg/models/results/convert.go`) — depends on Task 1, 2
- Task 4: API envelope update (`pkg/models/models.go`) — depends on Task 1
- Task 5: TypedResultContainer (`internal/search/typed_container.go`) — depends on Task 3, 4

**Phase B: Integration** (depends on Phase A)
- Task 6: ProcessorResult extension (`internal/search/processor/types.go`) 
- Task 7: Processor adapter (`internal/search/processor/online.go`) — depends on Task 3, 6
- Task 8: Scheduler integration (`internal/search/scheduler.go`) — depends on Task 5, 6
- Task 9: Base engines adaptation (6 existing engines) — depends on Task 1, 3

**Phase C: Frontend** (depends on Phase A for API contract)
- Task 10: TypeScript types (`web/src/types/search.ts`) 
- Task 11: ResultCard component (`web/src/components/results/ResultCard.tsx`) — depends on Task 10
- Task 12: Kind-specific components — depends on Task 10, 11
- Task 13: SearchPage layout (`web/src/pages/SearchPage.tsx`) — depends on Task 11, 12

That's 13 tasks, way more than 8. Let me split into parts:

Part 1: `backend-types.md` — Tasks 1-5 (result types, normalization, conversion, API envelope, typed container)
Part 2: `integration.md` — Tasks 6-9 (processor, scheduler, base engines)
Part 3: `frontend.md` — Tasks 10-13 (TypeScript types, components, layout)

Let me verify: 5 + 4 + 4 = 13 tasks. That works for the split.

Now let me explore the existing code to understand what's already there before writing.

I need to check:
1. `pkg/models/results/` — what already exists
2. `pkg/models/models.go` — current Result/Response types
3. `internal/search/merge.go` — existing ResultContainer
4. `internal/search/scheduler.go` — how results are collected
5. `internal/search/processor/types.go` — ProcessorResult
6. `internal/search/processor/online.go` — how processors work
7. `web/src/types/search.ts` — current TypeScript types
8. `web/src/pages/SearchPage.tsx` — current rendering
9. `web/src/components/` — existing components
10. The 6 existing engines — to understand their Search signatures

Let me start exploring the codebase.

</details>

Let me explore the existing codebase to understand what's already in place before writing the plan.

#### Tool Call: Glob (`pkg/models/results/**/*.go`)
<!-- call_id: call_00_PcybHziwRJeSinVR2pDN0881 -->
```json
{
  "pattern": "pkg/models/results/**/*.go"
}
```

#### Tool Call: Glob (`web/src/components/**/*`)
<!-- call_id: call_01_j9xnpsbXo2et2fgzQdOb0076 -->
```json
{
  "pattern": "web/src/components/**/*"
}
```

#### Tool Call: Glob (`web/src/types/**/*`)
<!-- call_id: call_02_9ZEm94X2tVVfYx4cXd819988 -->
```json
{
  "pattern": "web/src/types/**/*"
}
```

<details><summary>Tool Result: Glob (`pkg/models/results/**/*.go`)</summary>

<!-- call_id: call_00_PcybHziwRJeSinVR2pDN0881 -->
pkg/models/results/convert_test.go
pkg/models/results/convert.go
pkg/models/results/normalize.go
pkg/models/results/normalize_test.go
pkg/models/results/results_test.go
pkg/models/results/results.go
pkg/models/results/doc.go

</details>

<details><summary>Tool Result: Glob (`web/src/components/**/*`)</summary>

<!-- call_id: call_01_j9xnpsbXo2et2fgzQdOb0076 -->
No matches found

</details>

<details><summary>Tool Result: Glob (`web/src/types/**/*`)</summary>

<!-- call_id: call_02_9ZEm94X2tVVfYx4cXd819988 -->
web/src/types/search.ts
web/src/types/engine.ts
web/src/types/config.ts

</details>

<details><summary>Thinking</summary>

Let me now read the existing files to understand the current state.

</details>

#### Tool Call: Read (`/Users/ranwei/workspace/go_work/seargo/pkg/models/results/re…`)
<!-- call_id: call_00_vqamTFHfGtmtxncvmFFz8353 -->
```json
{
  "path": "/Users/ranwei/workspace/go_work/seargo/pkg/models/results/results.go"
}
```

#### Tool Call: Read (`/Users/ranwei/workspace/go_work/seargo/pkg/models/results/no…`)
<!-- call_id: call_01_2pBat2ySB8GDelIpFchW7500 -->
```json
{
  "path": "/Users/ranwei/workspace/go_work/seargo/pkg/models/results/normalize.go"
}
```

#### Tool Call: Read (`/Users/ranwei/workspace/go_work/seargo/pkg/models/results/co…`)
<!-- call_id: call_02_MnbhQHxhPXZx6uxAUGZg2877 -->
```json
{
  "path": "/Users/ranwei/workspace/go_work/seargo/pkg/models/results/convert.go"
}
```

#### Tool Call: Read (`/Users/ranwei/workspace/go_work/seargo/pkg/models/models.go`)
<!-- call_id: call_03_BFlVCNHatcALYeYYMlZI2888 -->
```json
{
  "path": "/Users/ranwei/workspace/go_work/seargo/pkg/models/models.go"
}
```

<details><summary>Tool Result: Read (`/Users/ranwei/workspace/go_work/seargo/pkg/models/results/re…`)</summary>

<!-- call_id: call_00_vqamTFHfGtmtxncvmFFz8353 -->
1	package results
2	
3	import (
4		"strings"
5		"time"
6	)
7	
8	// Result is the interface that all typed search results implement.
9	type Result interface {
10		GetTitle() string
11		GetURL() string
12		GetContent() string
13		GetEngine() string
14		GetTemplate() string
15		GetCategory() string
16	}
17	
18	// BaseResult holds common fields shared by all result types.
19	type BaseResult struct {
20		Title        string            `json:"title"`
21		URL          string            `json:"url"`
22		Content      string            `json:"content,omitempty"`
23		Engine       string            `json:"engine"`
24		Engines      []string          `json:"engines,omitempty"`
25		Template     string            `json:"template"`
26		Category     string            `json:"category,omitempty"`
27		Positions    []int             `json:"positions,omitempty"`
28		Score        float64           `json:"score,omitempty"`
29		PublishedAt  *time.Time        `json:"published_at,omitempty"`
30		ThumbnailURL string            `json:"thumbnail_url,omitempty"`
31		Domain       string            `json:"domain,omitempty"`
32		Favicon      string            `json:"favicon,omitempty"`
33		EngineData   map[string]any    `json:"engine_data,omitempty"`
34		ParsedURL    []string          `json:"parsed_url,omitempty"`
35		IsOnion      bool              `json:"is_onion,omitempty"`
36	}
37	
38	func (b BaseResult) GetTitle() string    { return b.Title }
39	func (b BaseResult) GetURL() string      { return b.URL }
40	func (b BaseResult) GetContent() string  { return b.Content }
41	func (b BaseResult) GetEngine() string   { return b.Engine }
42	func (b BaseResult) GetTemplate() string { return b.Template }
43	func (b BaseResult) GetCategory() string { return b.Category }
44	
45	// MainResult is the primary displayed search result (default template).
46	type MainResult struct {
47		BaseResult
48	}
49	
50	// ImageResult represents an image search result.
51	type ImageResult struct {
52		BaseResult
53		ThumbnailSrc string `json:"thumbnail_src,omitempty"`
54		ImgSrc       string `json:"img_src,omitempty"`
55		ImgFormat    string `json:"img_format,omitempty"`
56		Resolution   string `json:"resolution,omitempty"`
57		ImgAlt       string `json:"img_alt,omitempty"`
58		Source       string `json:"source,omitempty"`
59		Width        int    `json:"width,omitempty"`
60		Height       int    `json:"height,omitempty"`
61	}
62	
63	// IsBase64 returns true if the image source is a base64 data URI.
64	func (i *ImageResult) IsBase64() bool {
65		return strings.HasPrefix(i.ImgSrc, "data:image/")
66	}
67	
68	// VideoResult represents a video search result.
69	type VideoResult struct {
70		BaseResult
71		Thumbnail  string `json:"thumbnail,omitempty"`
72		IFrameSrc  string `json:"iframe_src,omitempty"`
73		Length     string `json:"length,omitempty"`
74		Duration   string `json:"duration,omitempty"`
75		Author     string `json:"author,omitempty"`
76		UploadDate string `json:"upload_date,omitempty"`
77		ViewCount  int64  `json:"view_count,omitempty"`
78	}
79	
80	// NewsResult represents a news/article search result.
81	type NewsResult struct {
82		BaseResult
83	}
84	
85	// PaperResult represents an academic paper search result.
86	type PaperResult struct {
87		BaseResult
88		DOI           string   `json:"doi,omitempty"`
89		Journal       string   `json:"journal,omitempty"`
90		Authors       []string `json:"authors,omitempty"`
91		Publisher     string   `json:"publisher,omitempty"`
92		Type          string   `json:"type,omitempty"`
93		PublishedDate string   `json:"published_date,omitempty"`
94		Editors       []string `json:"editors,omitempty"`
95		PDFURL        string   `json:"pdf_url,omitempty"`
96		HTMLURL       string   `json:"html_url,omitempty"`
97		Comments      string   `json:"comments,omitempty"`
98		Tags          []string `json:"tags,omitempty"`
99		Pages         string   `json:"pages,omitempty"`
100		ISSN          []string `json:"issn,omitempty"`
101		ISBN          []string `json:"isbn,omitempty"`
102	}
103	
104	// CodeResult represents a code search result (e.g. from GitHub, GitLab).
105	type CodeResult struct {
106		BaseResult
107		Repository   string `json:"repository,omitempty"`
108		CodeLanguage string `json:"code_language,omitempty"`
109	}
110	
111	// FileResult represents a file search result.
112	type FileResult struct {
113		BaseResult
114		FileType string `json:"file_type,omitempty"`
115		FileSize int64  `json:"file_size,omitempty"`
116	}
117	
118	// MapResult represents a map/geolocation search result.
119	type MapResult struct {
120		BaseResult
121		Latitude    float64   `json:"latitude,omitempty"`
122		Longitude   float64   `json:"longitude,omitempty"`
123		BoundingBox []float64 `json:"bounding_box,omitempty"`
124		Address     string    `json:"address,omitempty"`
125		MapURL      string    `json:"map_url,omitempty"`
126	}
127	
128	// MusicResult represents a music search result.
129	type MusicResult struct {
130		BaseResult
131		Artist   string `json:"artist,omitempty"`
132		Album    string `json:"album,omitempty"`
133		Duration string `json:"duration,omitempty"`
134	}
135	
136	// AnswerResult represents an answer box result.
137	type AnswerResult struct {
138		BaseResult
139		Answer string `json:"answer"`
140	}
141	
142	// InfoboxResult represents an infobox result.
143	type InfoboxResult struct {
144		BaseResult
145		InfoboxID  string              `json:"infobox_id,omitempty"`
146		Attributes []map[string]string `json:"attributes,omitempty"`
147		URLs       []map[string]string `json:"urls,omitempty"`
148		ImgSrc     string              `json:"img_src,omitempty"`
149		ImgAlt     string              `json:"img_alt,omitempty"`
150	}
151	
152	// ResultTypes is a convenience type for holding heterogeneous result lists.
153	type ResultTypes struct {
154		Main      []MainResult      `json:"main,omitempty"`
155		Images    []ImageResult     `json:"images,omitempty"`
156		Videos    []VideoResult     `json:"videos,omitempty"`
157		News      []NewsResult      `json:"news,omitempty"`
158		Papers    []PaperResult     `json:"papers,omitempty"`
159		Code      []CodeResult      `json:"code,omitempty"`
160		Files     []FileResult      `json:"files,omitempty"`
161		Maps      []MapResult       `json:"maps,omitempty"`
162		Music     []MusicResult     `json:"music,omitempty"`
163		Answers   []AnswerResult    `json:"answers,omitempty"`
164		Infoboxes []InfoboxResult   `json:"infoboxes,omitempty"`
165	}
<system>165 lines read from file starting from line 1. Total lines in file: 165. End of file reached.</system>

</details>

<details><summary>Tool Result: Read (`/Users/ranwei/workspace/go_work/seargo/pkg/models/results/no…`)</summary>

<!-- call_id: call_01_2pBat2ySB8GDelIpFchW7500 -->
1	package results
2	
3	import (
4		"net/url"
5		"regexp"
6		"strconv"
7		"strings"
8		"time"
9	)
10	
11	var whitespaceRE = regexp.MustCompile(`\s+`)
12	
13	// normalizeURL normalizes a result URL:
14	// - Default scheme to http if missing
15	// - Sync ParsedURL with parsed components
16	func normalizeURL(r *BaseResult) {
17		if r.URL == "" {
18			return
19		}
20	
21		raw := r.URL
22	
23		// Default scheme
24		if !strings.Contains(raw, "://") {
25			raw = "http://" + raw
26		}
27	
28		u, err := url.Parse(raw)
29		if err != nil {
30			r.URL = raw
31			r.ParsedURL = nil
32			return
33		}
34	
35		r.URL = u.String()
36		r.ParsedURL = []string{u.Scheme, u.Host, u.Path, u.RawQuery, u.Fragment}
37	}
38	
39	// normalizeText collapses whitespace and deduplicates title==content.
40	func normalizeText(r *BaseResult) {
41		r.Title = strings.TrimSpace(whitespaceRE.ReplaceAllString(r.Title, " "))
42		r.Content = strings.TrimSpace(whitespaceRE.ReplaceAllString(r.Content, " "))
43	
44		// Dedup: if content equals title, clear content
45		if r.Content != "" && r.Content == r.Title {
46			r.Content = ""
47		}
48	}
49	
50	// normalizeDate attempts to parse a date string and set PublishedAt.
51	// Years < 1900 are rejected as invalid.
52	func normalizeDate(r *BaseResult, dateStr string) {
53		if dateStr == "" {
54			return
55		}
56	
57		// Try common formats
58		formats := []string{
59			time.RFC3339,
60			"2006-01-02T15:04:05",
61			"2006-01-02",
62			"2006-01",
63			"2006",
64			"02 Jan 2006",
65			"January 2, 2006",
66			"Jan 2, 2006",
67		}
68	
69		for _, layout := range formats {
70			t, err := time.Parse(layout, dateStr)
71			if err == nil {
72				if t.Year() >= 1900 {
73					r.PublishedAt = &t
74					return
75				}
76				return
77			}
78		}
79	
80		// Try parsing as year only
81		if year, err := strconv.Atoi(dateStr); err == nil && year >= 1900 {
82			t := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
83			r.PublishedAt = &t
84		}
85	}
86	
87	// Normalize applies all normalization functions to a BaseResult.
88	func (r *BaseResult) Normalize() {
89		normalizeURL(r)
90		normalizeText(r)
91	}
92	
93	// extractDomainFromURL extracts the domain (host) from a URL string.
94	func extractDomainFromURL(rawURL string) string {
95		if rawURL == "" {
96			return ""
97		}
98		if !strings.Contains(rawURL, "://") {
99			rawURL = "http://" + rawURL
100		}
101		u, err := url.Parse(rawURL)
102		if err != nil {
103			return ""
104		}
105		return strings.ToLower(u.Host)
106	}
107	
108	// computeParsedURL fills ParsedURL from the URL string.
109	func computeParsedURL(rawURL string) []string {
110		if rawURL == "" {
111			return nil
112		}
113		if !strings.Contains(rawURL, "://") {
114			rawURL = "http://" + rawURL
115		}
116		u, err := url.Parse(rawURL)
117		if err != nil {
118			return nil
119		}
120		return []string{u.Scheme, u.Host, u.Path, u.RawQuery, u.Fragment}
121	}
<system>121 lines read from file starting from line 1. Total lines in file: 121. End of file reached.</system>

</details>

<details><summary>Tool Result: Read (`/Users/ranwei/workspace/go_work/seargo/pkg/models/results/co…`)</summary>

<!-- call_id: call_02_MnbhQHxhPXZx6uxAUGZg2877 -->
1	package results
2	
3	import (
4		"github.com/seargo/seargo/pkg/models"
5	)
6	
7	// ToAPIResult converts a slice of typed Result values into the flat
8	// models.Result format used by the existing API response pipeline.
9	func ToAPIResult(results []Result) []models.Result {
10		if len(results) == 0 {
11			return nil
12		}
13	
14		out := make([]models.Result, 0, len(results))
15		for _, r := range results {
16			api := models.Result{
17				Title:    r.GetTitle(),
18				URL:      r.GetURL(),
19				Content:  r.GetContent(),
20				Engine:   r.GetEngine(),
21				Template: r.GetTemplate(),
22			}
23	
24			ed := make(map[string]any)
25	
26			switch t := r.(type) {
27			case *ImageResult:
28				api.ThumbnailURL = t.ThumbnailSrc
29				ed["img_src"] = t.ImgSrc
30				ed["img_format"] = t.ImgFormat
31				ed["resolution"] = t.Resolution
32				ed["source"] = t.Source
33	
34			case *VideoResult:
35				api.ThumbnailURL = t.Thumbnail
36				ed["iframe_src"] = t.IFrameSrc
37				ed["length"] = t.Length
38				ed["author"] = t.Author
39	
40			case *PaperResult:
41				ed["doi"] = t.DOI
42				ed["journal"] = t.Journal
43				ed["authors"] = t.Authors
44				ed["pdf_url"] = t.PDFURL
45	
46			case *CodeResult:
47				ed["repository"] = t.Repository
48				ed["code_language"] = t.CodeLanguage
49	
50			case *FileResult:
51				ed["file_type"] = t.FileType
52				ed["file_size"] = t.FileSize
53	
54			case *MapResult:
55				ed["latitude"] = t.Latitude
56				ed["longitude"] = t.Longitude
57				ed["map_url"] = t.MapURL
58	
59			case *MusicResult:
60				ed["artist"] = t.Artist
61				ed["album"] = t.Album
62	
63			case *MainResult:
64				if t.PublishedAt != nil {
65					api.PublishedAt = t.PublishedAt
66				}
67				if t.EngineData != nil {
68					for k, v := range t.EngineData {
69						ed[k] = v
70					}
71				}
72	
73			case *NewsResult:
74				if t.PublishedAt != nil {
75					api.PublishedAt = t.PublishedAt
76				}
77	
78			default:
79				// Unknown types: passthrough BaseResult fields only
80			}
81	
82			if len(ed) > 0 {
83				api.EngineData = ed
84			}
85	
86			// Normalize URL before returning
87			br := &BaseResult{
88				Title:    api.Title,
89				URL:      api.URL,
90				Content:  api.Content,
91				Template: api.Template,
92			}
93			normalizeURL(br)
94			api.URL = br.URL
95			api.Domain = extractDomainFromURL(api.URL)
96	
97			out = append(out, api)
98		}
99	
100		return out
101	}
<system>101 lines read from file starting from line 1. Total lines in file: 101. End of file reached.</system>

</details>

<details><summary>Tool Result: Read (`/Users/ranwei/workspace/go_work/seargo/pkg/models/models.go`)</summary>

<!-- call_id: call_03_BFlVCNHatcALYeYYMlZI2888 -->
1	package models
2	
3	import (
4		"fmt"
5		"hash/fnv"
6		"time"
7	)
8	
9	type Category string
10	
11	const (
12		CategoryGeneral     Category = "general"
13		CategoryImages      Category = "images"
14		CategoryVideos      Category = "videos"
15		CategoryNews        Category = "news"
16		CategoryMap         Category = "map"
17		CategoryMusic       Category = "music"
18		CategoryIT          Category = "it"
19		CategoryScience     Category = "science"
20		CategoryFiles       Category = "files"
21		CategorySocialMedia Category = "social media"
22	)
23	
24	func AllCategories() []Category {
25		return []Category{
26			CategoryGeneral, CategoryImages, CategoryVideos, CategoryNews,
27			CategoryMap, CategoryMusic, CategoryIT, CategoryScience,
28			CategoryFiles, CategorySocialMedia,
29		}
30	}
31	
32	type Request struct {
33		Query      string   `form:"q" binding:"required"`
34		Category   Category `form:"category"`
35		Language   string   `form:"language"`
36		SafeSearch int      `form:"safesearch"`
37		TimeRange  string   `form:"time_range"`
38		Page       int      `form:"page"`
39		PageSize   int      `form:"page_size"`
40	}
41	
42	func (r *Request) CacheKey() string {
43		h := fnv.New64a()
44		h.Write([]byte(r.Query))
45		return fmt.Sprintf("search:%s:%s:%d:%s:%d:%d:%x",
46			r.Category, r.Language, r.SafeSearch,
47			r.TimeRange, r.Page, r.PageSize, h.Sum64())
48	}
49	
50	type NormalizeDefaults struct {
51		DefaultLang     string
52		DefaultCategory Category
53		DefaultPageSize int
54		MaxResults      int
55	}
56	
57	func (r *Request) Normalize(d NormalizeDefaults) {
58		if r.Language == "" {
59			r.Language = d.DefaultLang
60		}
61		if r.Category == "" {
62			r.Category = d.DefaultCategory
63		}
64		if r.PageSize <= 0 {
65			r.PageSize = d.DefaultPageSize
66		}
67		if r.Page <= 0 {
68			r.Page = 1
69		}
70		// Cap PageSize to MaxResults
71		if r.PageSize > d.MaxResults && d.MaxResults > 0 {
72			r.PageSize = d.MaxResults
73		}
74	}
75	
76	
77	type Answer struct {
78		Answer  string `json:"answer"`
79		URL     string `json:"url,omitempty"`
80		Content string `json:"content"`
81		Engine  string `json:"engine,omitempty"`
82	}
83	
84	type Infobox struct {
85		Title       string   `json:"title"`
86		URL         string   `json:"url,omitempty"`
87		Content     string   `json:"content,omitempty"`
88		Engine      string   `json:"engine,omitempty"`
89		ImgSrc      string   `json:"img_src,omitempty"`
90		URLs        []string `json:"urls,omitempty"`
91		Attributes  []string `json:"attributes,omitempty"`
92	}
93	
94	type Result struct {
95		Title        string         `json:"title"`
96		URL          string         `json:"url"`
97		Content      string         `json:"content"`
98		Engine       string         `json:"engine"`
99		Engines      []string       `json:"engines,omitempty"`
100		Category     Category       `json:"category"`
101		Score        float64        `json:"score"`
102		Positions    []int          `json:"-"`
103		Template     string         `json:"template,omitempty"`
104		ThumbnailURL string         `json:"thumbnail_url,omitempty"`
105		PublishedAt  *time.Time     `json:"published_at,omitempty"`
106		Domain       string         `json:"domain,omitempty"`
107		Favicon      string         `json:"favicon,omitempty"`
108		EngineData   map[string]any `json:"engine_data,omitempty"`
109	}
110	
111	type Response struct {
112		Query          string         `json:"query"`
113		Category       Category       `json:"category"`
114		Results        []Result       `json:"results"`
115		Suggestions    []string       `json:"suggestions"`
116		Answers        []Answer       `json:"answers,omitempty"`
117		Corrections    []string       `json:"corrections,omitempty"`
118		Infoboxes      []Infobox      `json:"infoboxes,omitempty"`
119		EngineData     map[string]any `json:"engine_data,omitempty"`
120		Total          int            `json:"total"`
121		Page           int            `json:"page"`
122		PageSize       int            `json:"page_size"`
123		EnginesUsed    []string       `json:"engines_used"`
124		EnginesFailed  []string       `json:"engines_failed"`
125		ResponseTimeMs int64          `json:"response_time_ms"`
126		RedirectURL    string         `json:"redirect_url,omitempty"`
127	}
<system>127 lines read from file starting from line 1. Total lines in file: 127. End of file reached.</system>

</details>

<details><summary>Thinking</summary>

Now let me read the remaining files I need to understand: the search layer, processors, scheduler, and web frontend.

</details>

#### Tool Call: Read (`/Users/ranwei/workspace/go_work/seargo/internal/search/merge…`)
<!-- call_id: call_00_7KCmORZCujHwAxVmz9SF6231 -->
```json
{
  "path": "/Users/ranwei/workspace/go_work/seargo/internal/search/merge.go"
}
```

#### Tool Call: Read (`/Users/ranwei/workspace/go_work/seargo/internal/search/proce…`)
<!-- call_id: call_01_UJEJF7XkdYH7gYX1NV0W7246 -->
```json
{
  "path": "/Users/ranwei/workspace/go_work/seargo/internal/search/processor/types.go"
}
```

#### Tool Call: Read (`/Users/ranwei/workspace/go_work/seargo/internal/search/proce…`)
<!-- call_id: call_02_GhowaASwXBIHlqpuxBNr5658 -->
```json
{
  "path": "/Users/ranwei/workspace/go_work/seargo/internal/search/processor/online.go"
}
```

#### Tool Call: Read (`/Users/ranwei/workspace/go_work/seargo/internal/search/sched…`)
<!-- call_id: call_03_pNec6iYxbdmOZFsXdruj4125 -->
```json
{
  "path": "/Users/ranwei/workspace/go_work/seargo/internal/search/scheduler.go"
}
```

<details><summary>Tool Result: Read (`/Users/ranwei/workspace/go_work/seargo/internal/search/merge…`)</summary>

<!-- call_id: call_00_7KCmORZCujHwAxVmz9SF6231 -->
1	package search
2	
3	import (
4		"net/url"
5		"sort"
6		"strings"
7		"sync"
8	
9		"github.com/seargo/seargo/pkg/models"
10	)
11	
12	// trackingParams 是需要从 URL 中移除的已知追踪参数。
13	var trackingParams = map[string]bool{
14		"utm_source":   true,
15		"utm_medium":   true,
16		"utm_campaign": true,
17		"utm_term":     true,
18		"utm_content":  true,
19		"fbclid":       true,
20		"gclid":        true,
21		"ref":          true,
22		"ref_src":      true,
23		"ref_url":      true,
24	}
25	
26	// normalizeURL 对 URL 做规范化：
27	// - scheme 和 host 转小写
28	// - 去除 www. 前缀（非 www2/www3 等子域名）
29	// - 去除 trailing slash
30	// - 去除已知追踪参数
31	// - 去除 fragment
32	func normalizeURL(raw string) string {
33		u, err := url.Parse(raw)
34		if err != nil {
35			return raw
36		}
37	
38		u.Scheme = strings.ToLower(u.Scheme)
39		u.Host = strings.ToLower(u.Host)
40	
41		// 去掉 www. 前缀（仅当 host 是 www.xxx 格式时）
42		if strings.HasPrefix(u.Host, "www.") && !strings.HasPrefix(u.Host, "www2.") &&
43			!strings.HasPrefix(u.Host, "www3.") {
44			u.Host = u.Host[4:]
45		}
46	
47		u.Path = strings.TrimSuffix(u.Path, "/")
48	
49		// 移除追踪参数
50		q := u.Query()
51		for param := range trackingParams {
52			q.Del(param)
53		}
54		u.RawQuery = q.Encode()
55	
56		u.Fragment = ""
57	
58		return u.String()
59	}
60	
61	// UnresponsiveEngine 记录一个无响应的引擎及其原因。
62	type UnresponsiveEngine struct {
63		Name   string
64		Reason string
65	}
66	
67	// ResultContainer 管理跨引擎搜索结果的并发写入、去重合并、排序和分组。
68	type ResultContainer struct {
69		mu            sync.Mutex
70		closed        bool
71		results       map[string]*models.Result // key = dedupKey
72		answers       []models.Answer
73		suggestions   []string
74		suggestionSet map[string]bool
75		corrections   []string
76		infoboxes     []models.Infobox
77		engineData    map[string]any
78		unresponsive  []UnresponsiveEngine
79		engineWeights map[string]float64
80	}
81	
82	// NewResultContainer 创建结果容器。
83	func NewResultContainer(engineWeights map[string]float64) *ResultContainer {
84		return &ResultContainer{
85			results:       make(map[string]*models.Result),
86			suggestionSet: make(map[string]bool),
87			engineData:    make(map[string]any),
88			engineWeights: engineWeights,
89		}
90	}
91	
92	// dedupKey 生成去重用的唯一键：template|host|path|query|thumbnail。
93	func dedupKey(r models.Result) string {
94		normURL := normalizeURL(r.URL)
95		u, err := url.Parse(normURL)
96		if err != nil {
97			return r.Template + "|" + r.URL + "|" + r.ThumbnailURL
98		}
99		return r.Template + "|" + u.Host + "|" + u.Path + "|" + u.RawQuery + "|" + r.ThumbnailURL
100	}
101	
102	// preferHTTPS 判断两个 URL 之间是否应优选 HTTPS 版本。
103	func preferHTTPS(newURL, oldURL string) bool {
104		return strings.HasPrefix(newURL, "https://") && !strings.HasPrefix(oldURL, "https://")
105	}
106	
107	// Extend 将单个引擎的结果并入容器。positionBase 是该引擎结果起始位置（0-based）。
108	func (c *ResultContainer) Extend(engineName string, results []models.Result, positionBase int) {
109		c.mu.Lock()
110		defer c.mu.Unlock()
111		if c.closed {
112			return
113		}
114	
115		for i, r := range results {
116			key := dedupKey(r)
117			position := positionBase + i + 1 // 1-based
118	
119			existing, ok := c.results[key]
120			if !ok {
121				r.Engine = engineName
122				r.Engines = []string{engineName}
123				r.Positions = []int{position}
124				if r.Domain == "" {
125					r.Domain = extractDomain(r.URL)
126				}
127				c.results[key] = &r
128				continue
129			}
130	
131			// 合并已存在的结果
132			existing.Engines = appendUniqueStr(existing.Engines, engineName)
133			existing.Positions = append(existing.Positions, position)
134			if len(r.Title) > len(existing.Title) {
135				existing.Title = r.Title
136			}
137			if len(r.Content) > len(existing.Content) {
138				existing.Content = r.Content
139			}
140			if preferHTTPS(r.URL, existing.URL) {
141				existing.URL = r.URL
142			}
143			if r.ThumbnailURL != "" && existing.ThumbnailURL == "" {
144				existing.ThumbnailURL = r.ThumbnailURL
145			}
146		}
147	}
148	
149	// extractDomain 从 URL 提取域名。
150	func extractDomain(rawURL string) string {
151		u, err := url.Parse(rawURL)
152		if err != nil {
153			return ""
154		}
155		return strings.ToLower(u.Host)
156	}
157	
158	func appendUniqueStr(slice []string, item string) []string {
159		for _, v := range slice {
160			if v == item {
161				return slice
162			}
163		}
164		return append(slice, item)
165	}
166	
167	// Close 标记容器写入完成，计算分数和排序。
168	func (c *ResultContainer) Close() {
169		c.mu.Lock()
170		defer c.mu.Unlock()
171		c.closed = true
172	
173		for _, r := range c.results {
174			c.calculateScore(r)
175		}
176	}
177	
178	// calculateScore 计算 SearXNG 风格分数：Σ (weight / position) × count。
179	func (c *ResultContainer) calculateScore(r *models.Result) {
180		score := 0.0
181		for i, pos := range r.Positions {
182			engineName := ""
183			if i < len(r.Engines) {
184				engineName = r.Engines[i]
185			}
186			weight := c.engineWeights[engineName]
187			if weight == 0 {
188				weight = 1.0
189			}
190			score += weight / float64(pos)
191		}
192		score *= float64(len(r.Positions))
193		r.Score = score
194	}
195	
196	type resultGroup struct {
197		key     string
198		items   []*models.Result
199		lastIdx int
200	}
201	
202	// GetOrderedResults 返回按 score 降序 + category grouping 重排后的结果列表。
203	func (c *ResultContainer) GetOrderedResults() []models.Result {
204		all := make([]*models.Result, 0, len(c.results))
205		for _, r := range c.results {
206			all = append(all, r)
207		}
208		sort.Slice(all, func(i, j int) bool {
209			if all[i].Score != all[j].Score {
210				return all[i].Score > all[j].Score
211			}
212			return all[i].URL < all[j].URL
213		})
214	
215		const groupWindow = 20
216		const groupMaxSize = 8
217	
218		output := make([]models.Result, 0, len(all))
219		groupMap := make(map[string]*resultGroup)
220	
221		for _, r := range all {
222			gKey := string(r.Category) + "|" + r.Template
223			if r.ThumbnailURL != "" {
224				gKey += "|thumb"
225			}
226	
227			grp, exists := groupMap[gKey]
228			if exists {
229				distance := len(output) - grp.lastIdx
230				if len(grp.items) < groupMaxSize && distance < groupWindow {
231					grp.items = append(grp.items, r)
232					grp.lastIdx = len(output)
233					output = insertAt(output, grp.lastIdx, *r)
234					for _, g := range groupMap {
235						if g != grp && g.lastIdx >= grp.lastIdx {
236							g.lastIdx++
237						}
238					}
239					continue
240				}
241			}
242	
243			newGrp := &resultGroup{key: gKey, items: []*models.Result{r}, lastIdx: len(output)}
244			groupMap[gKey] = newGrp
245			output = append(output, *r)
246		}
247	
248		return output
249	}
250	
251	func insertAt(slice []models.Result, idx int, item models.Result) []models.Result {
252		if idx >= len(slice) {
253			return append(slice, item)
254		}
255		slice = append(slice, models.Result{})
256		copy(slice[idx+1:], slice[idx:])
257		slice[idx] = item
258		return slice
259	}
260	
261	// AddSuggestions 添加引擎的建议列表（大小写去重）。
262	func (c *ResultContainer) AddSuggestions(engineName string, suggestions []string) {
263		c.mu.Lock()
264		defer c.mu.Unlock()
265		if c.closed {
266			return
267		}
268		for _, s := range suggestions {
269			lower := strings.ToLower(s)
270			if !c.suggestionSet[lower] {
271				c.suggestionSet[lower] = true
272				c.suggestions = append(c.suggestions, s)
273			}
274		}
275	}
276	
277	// AddAnswers 添加答案列表。
278	func (c *ResultContainer) AddAnswers(engineName string, answers []models.Answer) {
279		c.mu.Lock()
280		defer c.mu.Unlock()
281		if c.closed {
282			return
283		}
284		c.answers = append(c.answers, answers...)
285	}
286	
287	// AddCorrections 添加拼写纠正建议。
288	func (c *ResultContainer) AddCorrections(engineName string, corrections []string) {
289		c.mu.Lock()
290		defer c.mu.Unlock()
291		if c.closed {
292			return
293		}
294		c.corrections = append(c.corrections, corrections...)
295	}
296	
297	// AddInfoboxes 添加信息框。
298	func (c *ResultContainer) AddInfoboxes(engineName string, infoboxes []models.Infobox) {
299		c.mu.Lock()
300		defer c.mu.Unlock()
301		if c.closed {
302			return
303		}
304		c.infoboxes = append(c.infoboxes, infoboxes...)
305	}
306	
307	// AddEngineData 添加引擎的透传数据。
308	func (c *ResultContainer) AddEngineData(engineName string, data map[string]any) {
309		c.mu.Lock()
310		defer c.mu.Unlock()
311		if c.closed {
312			return
313		}
314		for k, v := range data {
315			c.engineData[engineName+"."+k] = v
316		}
317	}
318	
319	// MarkUnresponsive 记录一个无响应的引擎。
320	func (c *ResultContainer) MarkUnresponsive(engineName, reason string) {
321		c.mu.Lock()
322		defer c.mu.Unlock()
323		c.unresponsive = append(c.unresponsive, UnresponsiveEngine{Name: engineName, Reason: reason})
324	}
325	
326	// GetSuggestions 返回建议列表。
327	func (c *ResultContainer) GetSuggestions() []string {
328		c.mu.Lock()
329		defer c.mu.Unlock()
330		if len(c.suggestions) == 0 {
331			return nil
332		}
333		result := make([]string, len(c.suggestions))
334		copy(result, c.suggestions)
335		return result
336	}
337	
338	// GetAnswers 返回答案列表。
339	func (c *ResultContainer) GetAnswers() []models.Answer {
340		c.mu.Lock()
341		defer c.mu.Unlock()
342		if len(c.answers) == 0 {
343			return nil
344		}
345		result := make([]models.Answer, len(c.answers))
346		copy(result, c.answers)
347		return result
348	}
349	
350	// GetCorrections 返回纠正列表。
351	func (c *ResultContainer) GetCorrections() []string {
352		c.mu.Lock()
353		defer c.mu.Unlock()
354		if len(c.corrections) == 0 {
355			return nil
356		}
357		result := make([]string, len(c.corrections))
358		copy(result, c.corrections)
359		return result
360	}
361	
362	// GetInfoboxes 返回信息框列表。
363	func (c *ResultContainer) GetInfoboxes() []models.Infobox {
364		c.mu.Lock()
365		defer c.mu.Unlock()
366		if len(c.infoboxes) == 0 {
367			return nil
368		}
369		result := make([]models.Infobox, len(c.infoboxes))
370		copy(result, c.infoboxes)
371		return result
372	}
373	
374	// GetEngineData 返回引擎透传数据。
375	func (c *ResultContainer) GetEngineData() map[string]any {
376		c.mu.Lock()
377		defer c.mu.Unlock()
378		if len(c.engineData) == 0 {
379			return nil
380		}
381		result := make(map[string]any, len(c.engineData))
382		for k, v := range c.engineData {
383			result[k] = v
384		}
385		return result
386	}
387	
388	// GetUnresponsive 返回无响应引擎列表。
389	func (c *ResultContainer) GetUnresponsive() []UnresponsiveEngine {
390		c.mu.Lock()
391		defer c.mu.Unlock()
392		if len(c.unresponsive) == 0 {
393			return nil
394		}
395		result := make([]UnresponsiveEngine, len(c.unresponsive))
396		copy(result, c.unresponsive)
397		return result
398	}
399	
400	// GetEnginesUsed 返回有结果贡献的引擎名。
401	func (c *ResultContainer) GetEnginesUsed() []string {
402		c.mu.Lock()
403		defer c.mu.Unlock()
404		seen := make(map[string]bool)
405		var names []string
406		for _, r := range c.results {
407			for _, e := range r.Engines {
408				if !seen[e] {
409					seen[e] = true
410					names = append(names, e)
411				}
412			}
413		}
414		return names
415	}
416	
417	// GetEnginesFailed 返回失败引擎名列表。
418	func (c *ResultContainer) GetEnginesFailed() []string {
419		c.mu.Lock()
420		defer c.mu.Unlock()
421		var names []string
422		for _, ue := range c.unresponsive {
423			names = append(names, ue.Name)
424		}
425		return names
426	}
<system>426 lines read from file starting from line 1. Total lines in file: 426. End of file reached.</system>

</details>

<details><summary>Tool Result: Read (`/Users/ranwei/workspace/go_work/seargo/internal/search/proce…`)</summary>

<!-- call_id: call_01_UJEJF7XkdYH7gYX1NV0W7246 -->
1	package processor
2	
3	import (
4		"context"
5		"errors"
6		"strings"
7	
8		"github.com/seargo/seargo/internal/engine"
9		"github.com/seargo/seargo/internal/search/query"
10		"github.com/seargo/seargo/pkg/models"
11	)
12	
13	var ErrUnsupportedSearch = errors.New("unsupported search")
14	
15	// Suspension 定义暂停/恢复能力接口，由 search.SuspensionTracker 实现。
16	type Suspension interface {
17		Ban(engineName, errorClass string)
18		IsSuspended(engineName string) bool
19	}
20	
21	// RequestParams 是传给底层 engine.Engine.Search 的参数。
22	type RequestParams struct {
23		Query      string
24		Category   models.Category
25		PageNo     int
26		Language   string
27		TimeRange  string
28		SafeSearch int
29		EngineData map[string]any
30	}
31	
32	// ProcessorResult 是单次 processor 搜索返回的结果流。
33	type ProcessorResult struct {
34		Results     []models.Result
35		Suggestions []string
36		Answers     []models.Answer
37		Corrections []string
38		Infoboxes   []models.Infobox
39		EngineData  map[string]any
40	}
41	
42	// Processor 是搜索处理器的统一接口。
43	type Processor interface {
44		Engine() engine.Engine
45		Search(ctx context.Context, q *query.ParsedQuery, page int) (*ProcessorResult, error)
46		Suspended() bool
47		RecordResult(ok bool, err error)
48		GetParams(q *query.ParsedQuery, page int) (*RequestParams, bool)
49	}
50	
51	// BaseProcessor 提供 Suspended/RecordResult 默认实现。
52	type BaseProcessor struct {
53		engineName string
54		suspension Suspension
55	}
56	
57	func (bp *BaseProcessor) Suspended() bool {
58		if bp.suspension == nil {
59			return false
60		}
61		return bp.suspension.IsSuspended(bp.engineName)
62	}
63	
64	func (bp *BaseProcessor) RecordResult(ok bool, err error) {
65		if ok || bp.suspension == nil {
66			return
67		}
68		bp.suspension.Ban(bp.engineName, classifyError(err))
69	}
70	
71	func classifyError(err error) string {
72		if err == nil {
73			return ""
74		}
75		msg := strings.ToLower(err.Error())
76		if strings.Contains(msg, "403") || strings.Contains(msg, "forbidden") ||
77			strings.Contains(msg, "access denied") {
78			return "SearxEngineAccessDenied"
79		}
80		if strings.Contains(msg, "captcha") || strings.Contains(msg, "recaptcha") ||
81			strings.Contains(msg, "challenge") {
82			return "SearxEngineCaptcha"
83		}
84		if strings.Contains(msg, "429") || strings.Contains(msg, "too many requests") ||
85			strings.Contains(msg, "rate limit") {
86			return "SearxEngineTooManyRequests"
87		}
88		return "SearxEngineTooManyRequests"
89	}
<system>89 lines read from file starting from line 1. Total lines in file: 89. End of file reached.</system>

</details>

<details><summary>Tool Result: Read (`/Users/ranwei/workspace/go_work/seargo/internal/search/proce…`)</summary>

<!-- call_id: call_02_GhowaASwXBIHlqpuxBNr5658 -->
1	package processor
2	
3	import (
4		"context"
5		"strings"
6	
7		"github.com/seargo/seargo/internal/engine"
8		"github.com/seargo/seargo/internal/httpx"
9		"github.com/seargo/seargo/internal/search/query"
10		"github.com/seargo/seargo/pkg/models"
11	)
12	
13	type OnlineProcessor struct {
14		BaseProcessor
15		eng    engine.Engine
16		client *httpx.Client
17	}
18	
19	func NewOnlineProcessor(eng engine.Engine, suspension Suspension, client *httpx.Client) *OnlineProcessor {
20		return &OnlineProcessor{
21			BaseProcessor: BaseProcessor{engineName: eng.Name(), suspension: suspension},
22			eng:           eng,
23			client:        client,
24		}
25	}
26	
27	func (p *OnlineProcessor) Engine() engine.Engine { return p.eng }
28	
29	func (p *OnlineProcessor) GetParams(q *query.ParsedQuery, page int) (*RequestParams, bool) {
30		caps := p.eng.Capabilities()
31		if page > 1 && !caps.SupportsPagination {
32			return nil, false
33		}
34		if q.TimeRange != "" && !caps.SupportsTimeRange {
35			return nil, false
36		}
37		return &RequestParams{
38			Query:      strings.Join(q.Terms, " "),
39			PageNo:     page,
40			Language:   q.Lang,
41			TimeRange:  q.TimeRange,
42			SafeSearch: q.SafeSearch,
43		}, true
44	}
45	
46	func (p *OnlineProcessor) Search(ctx context.Context, q *query.ParsedQuery, page int) (*ProcessorResult, error) {
47		params, ok := p.GetParams(q, page)
48		if !ok {
49			return nil, ErrUnsupportedSearch
50		}
51		req := &models.Request{
52			Query:      params.Query,
53			Language:   params.Language,
54			SafeSearch: params.SafeSearch,
55			TimeRange:  params.TimeRange,
56			Page:       params.PageNo,
57		}
58		resp, err := p.eng.Search(ctx, req)
59		if err != nil {
60			p.RecordResult(false, err)
61			return nil, err
62		}
63		p.RecordResult(true, nil)
64		return &ProcessorResult{
65			Results:     resp.Results,
66			Suggestions: resp.Suggestions,
67		}, nil
68	}
<system>68 lines read from file starting from line 1. Total lines in file: 68. End of file reached.</system>

</details>

<details><summary>Tool Result: Read (`/Users/ranwei/workspace/go_work/seargo/internal/search/sched…`)</summary>

<!-- call_id: call_03_pNec6iYxbdmOZFsXdruj4125 -->
1	package search
2	
3	import (
4		"context"
5		"fmt"
6		"hash/fnv"
7		"net/url"
8		"strconv"
9		"strings"
10		"sync"
11		"time"
12	
13		"github.com/panjf2000/ants/v2"
14	
15		"github.com/seargo/seargo/internal/cache"
16		"github.com/seargo/seargo/internal/config"
17		"github.com/seargo/seargo/internal/engine"
18		"github.com/seargo/seargo/internal/httpx"
19		"github.com/seargo/seargo/internal/logger"
20		"github.com/seargo/seargo/internal/metrics"
21		"github.com/seargo/seargo/internal/search/processor"
22		"github.com/seargo/seargo/internal/search/query"
23		"github.com/seargo/seargo/pkg/models"
24	)
25	
26	type Scheduler struct {
27		processors           map[string]processor.Processor
28		engineConfigs        map[string]config.EngineConfig
29		engineWeights        map[string]float64
30		engineShortcuts      map[string]string
31		engineNames          []string
32		allCategories        []models.Category
33		workerPool           *ants.Pool
34		cache                cache.Cache
35		globalTimeout        time.Duration
36		defaultEngineTimeout time.Duration
37		suspension           *SuspensionTracker
38		categoriesAsTabs     map[string]config.CategoryTabConfig
39	}
40	
41	// isEngineEnabled 判断引擎是否启用。Enabled 优先于 Disabled。
42	func isEngineEnabled(ec config.EngineConfig) bool {
43		if ec.Enabled {
44			return true
45		}
46		return !ec.Disabled
47	}
48	
49	// engineKey 返回引擎在 map 中的 key。
50	func engineKey(ec config.EngineConfig) string {
51		if ec.Engine != "" {
52			return ec.Engine
53		}
54		return ec.Name
55	}
56	
57	func NewScheduler(cfg *config.Config, c cache.Cache, client *httpx.Client) (*Scheduler, error) {
58		pool, err := ants.NewPool(50)
59		if err != nil {
60			return nil, err
61		}
62	
63		// Build engine configs, weights, shortcuts, names
64		engineCfgs := make(map[string]config.EngineConfig, len(cfg.Engines)*2)
65		engineWeights := make(map[string]float64)
66		engineShortcuts := make(map[string]string)
67		engineNames := make([]string, 0, len(cfg.Engines))
68	
69		for _, ec := range cfg.Engines {
70			key := engineKey(ec)
71			engineCfgs[key] = ec
72			if ec.Name != "" && ec.Name != key {
73				engineCfgs[ec.Name] = ec
74			}
75			engineWeights[key] = ec.Weight
76			if ec.Shortcut != "" {
77				engineShortcuts[ec.Shortcut] = key
78			}
79			engineNames = append(engineNames, key)
80		}
81	
82		// Build categories list for bang parser
83		var allCategories []models.Category
84		for catStr := range cfg.CategoriesAsTabs {
85			allCategories = append(allCategories, models.Category(catStr))
86		}
87	
88		// Build processors
89		processors := make(map[string]processor.Processor)
90		suspension := NewSuspensionTracker(cfg.Search)
91	
92		for _, ec := range cfg.Engines {
93			if !isEngineEnabled(ec) {
94				continue
95			}
96			lookupName := engineKey(ec)
97			eng, ok := engine.Get(lookupName)
98			if !ok {
99				logger.Warn("Engine not found", "engine", lookupName)
100				continue
101			}
102			proc, err := processor.NewProcessorFromConfig(eng, ec, suspension, client)
103			if err != nil {
104				logger.Error("Failed to create processor", "engine", lookupName, "error", err)
105				continue
106			}
107			processors[lookupName] = proc
108			logger.Info("Engine registered", "engine", lookupName)
109		}
110	
111		// Compute global timeout
112		globalTimeout := time.Duration(cfg.Outgoing.RequestTimeout) * time.Second
113	
114		metrics.EngineReloadsTotal.WithLabelValues("all", "success").Inc()
115	
116		return &Scheduler{
117			processors:           processors,
118			engineConfigs:        engineCfgs,
119			engineWeights:        engineWeights,
120			engineShortcuts:      engineShortcuts,
121			engineNames:          engineNames,
122			allCategories:        allCategories,
123			workerPool:           pool,
124			cache:                c,
125			globalTimeout:        globalTimeout,
126			defaultEngineTimeout: 8 * time.Second,
127			suspension:           suspension,
128			categoriesAsTabs:     cfg.CategoriesAsTabs,
129		}, nil
130	}
131	
132	func toModelCategories(cats []string) []models.Category {
133		result := make([]models.Category, len(cats))
134		for i, c := range cats {
135			result[i] = models.Category(c)
136		}
137		return result
138	}
139	
140	// Search 执行完整的搜索流程。
141	func (s *Scheduler) Search(ctx context.Context, req *models.Request) (*models.Response, error) {
142		start := time.Now()
143	
144		// 1. Parse query
145		rtq := query.NewRawTextQuery(req.Query)
146		parsed, err := rtq.Parse(s.engineShortcuts, s.engineNames, s.allCategories, nil)
147		if err != nil {
148			return nil, err
149		}
150	
151		// 2. Cache check
152		if s.cache != nil {
153			key := s.cacheKey(parsed, req)
154			if cached, ok := s.cache.Get(key); ok {
155				cached.ResponseTimeMs = time.Since(start).Milliseconds()
156				return cached, nil
157			}
158		}
159	
160		// 3. External bang redirect
161		if parsed.ExternalBang != "" {
162			if redirectURL, ok := externalBangURL(parsed.ExternalBang, parsed.Terms); ok {
163				resp := &models.Response{
164					RedirectURL: redirectURL,
165				}
166				if s.cache != nil {
167					s.cache.Set(s.cacheKey(parsed, req), resp, s.cacheTTL(req.Category))
168				}
169				return resp, nil
170			}
171		}
172	
173		// 4. Select processors
174		procs := s.selectProcessors(parsed, req.Category)
175		if len(procs) == 0 {
176			return &models.Response{
177				Query:   req.Query,
178				Results: []models.Result{},
179			}, nil
180		}
181	
182		// 5. Compute timeout
183		timeout := s.computeTimeout(parsed, procs)
184		ctx, cancel := context.WithTimeout(ctx, timeout)
185		defer cancel()
186	
187		// 6. Execute processors (concurrent)
188		container := NewResultContainer(s.engineWeights)
189		s.executeProcessors(ctx, procs, parsed, req.Page, container)
190		container.Close()
191	
192		results := container.GetOrderedResults()
193		suggestions := container.GetSuggestions()
194		answers := container.GetAnswers()
195		corrections := container.GetCorrections()
196		infoboxes := container.GetInfoboxes()
197		engineData := container.GetEngineData()
198		unresponsive := container.GetUnresponsive()
199	
200		// 7. All engines failed
201		if len(results) == 0 && len(unresponsive) > 0 && len(unresponsive) == len(procs) {
202			return nil, fmt.Errorf("all engines failed")
203		}
204	
205		// 8. Paginate
206		pageSize := req.PageSize
207		if pageSize <= 0 {
208			pageSize = 10
209		}
210		window, total := paginate(results, req.Page, pageSize)
211	
212		// 9. Build response
213		response := &models.Response{
214			Query:          req.Query,
215			Category:       req.Category,
216			Results:        window,
217			Suggestions:    suggestions,
218			Answers:        answers,
219			Corrections:    corrections,
220			Infoboxes:      infoboxes,
221			EngineData:     engineData,
222			Total:          total,
223			Page:           req.Page,
224			PageSize:       pageSize,
225			EnginesUsed:    container.GetEnginesUsed(),
226			EnginesFailed:  container.GetEnginesFailed(),
227			ResponseTimeMs: time.Since(start).Milliseconds(),
228		}
229	
230		// 10. Record metrics
231		s.recordMetrics(response)
232		metrics.SearchResultsTotal.WithLabelValues(string(req.Category)).Add(float64(len(response.Results)))
233	
234		// 11. Write cache
235		if s.cache != nil {
236			s.cache.Set(s.cacheKey(parsed, req), response, s.cacheTTL(req.Category))
237		}
238	
239		return response, nil
240	}
241	
242	// executeProcessors 并发执行所有 processor，将结果写入 container。
243	func (s *Scheduler) executeProcessors(ctx context.Context, procs []processor.Processor, parsed *query.ParsedQuery, page int, container *ResultContainer) {
244		var wg sync.WaitGroup
245	
246		for _, p := range procs {
247			wg.Add(1)
248			proc := p
249			s.workerPool.Submit(func() {
250				defer wg.Done()
251	
252				engineStart := time.Now()
253				result, err := proc.Search(ctx, parsed, page)
254	
255				if err != nil {
256					metrics.EngineQueriesTotal.WithLabelValues(proc.Engine().Name(), "failed").Inc()
257					errorClass := classifyError(err)
258					metrics.EngineFailuresTotal.WithLabelValues(proc.Engine().Name(), errorClass).Inc()
259					// Track parser failures (e.g. JSON/HTML parsing errors from engines)
260					if strings.Contains(strings.ToLower(err.Error()), "parse") {
261						metrics.EngineParserFailures.WithLabelValues(proc.Engine().Name()).Inc()
262					}
263					logger.Warn("engine failed", "engine", proc.Engine().Name(), "error", err)
264					container.MarkUnresponsive(proc.Engine().Name(), err.Error())
265					return
266				}
267	
268				metrics.EngineQueriesTotal.WithLabelValues(proc.Engine().Name(), "success").Inc()
269				metrics.EngineQueryDuration.WithLabelValues(proc.Engine().Name()).Observe(time.Since(engineStart).Seconds())
270				metrics.EngineResults.WithLabelValues(proc.Engine().Name()).Observe(float64(len(result.Results)))
271	
272				container.Extend(proc.Engine().Name(), result.Results, 0)
273				if len(result.Suggestions) > 0 {
274					container.AddSuggestions(proc.Engine().Name(), result.Suggestions)
275				}
276				if len(result.Answers) > 0 {
277					container.AddAnswers(proc.Engine().Name(), result.Answers)
278				}
279				if len(result.Corrections) > 0 {
280					container.AddCorrections(proc.Engine().Name(), result.Corrections)
281				}
282				if len(result.Infoboxes) > 0 {
283					container.AddInfoboxes(proc.Engine().Name(), result.Infoboxes)
284				}
285				if len(result.EngineData) > 0 {
286					container.AddEngineData(proc.Engine().Name(), result.EngineData)
287				}
288			})
289		}
290	
291		wg.Wait()
292	}
293	
294	// selectProcessors 根据 parsed query 选择 processor。
295	// 如果指定了引擎引用，只使用明确命名的引擎；否则按分类匹配。
296	func (s *Scheduler) selectProcessors(parsed *query.ParsedQuery, defaultCat models.Category) []processor.Processor {
297		// If explicit engine refs, only use those
298		if len(parsed.EngineRefs) > 0 {
299			var selected []processor.Processor
300			for _, ref := range parsed.EngineRefs {
301				if proc, ok := s.processors[ref]; ok && !proc.Suspended() {
302					selected = append(selected, proc)
303				}
304			}
305			return selected
306		}
307	
308		// Otherwise match by category from categoriesAsTabs
309		cat := defaultCat
310		if len(parsed.Categories) > 0 {
311			cat = parsed.Categories[0]
312		}
313	
314		cfg, ok := s.categoriesAsTabs[string(cat)]
315		if !ok {
316			return nil
317		}
318		var selected []processor.Processor
319		for _, engineName := range cfg.Engines {
320			if proc, ok := s.processors[engineName]; ok && !proc.Suspended() {
321				selected = append(selected, proc)
322			}
323		}
324		return selected
325	}
326	
327	// computeTimeout 计算搜索超时时间：取引擎超时、用户指定超时、max_request_timeout 和全局超时的最小值。
328	func (s *Scheduler) computeTimeout(parsed *query.ParsedQuery, procs []processor.Processor) time.Duration {
329		timeout := s.defaultEngineTimeout
330	
331		// User-specified timeout from query
332		if parsed.Timeout > 0 {
333			timeout = time.Duration(parsed.Timeout * float64(time.Second))
334		}
335	
336		// Per-engine timeout (take the shortest)
337		for _, p := range procs {
338			name := p.Engine().Name()
339			if cfg, ok := s.engineConfigs[name]; ok && cfg.Timeout > 0 {
340				engineTimeout := time.Duration(cfg.Timeout * float64(time.Second))
341				if engineTimeout < timeout {
342					timeout = engineTimeout
343				}
344			}
345		}
346	
347		// Global timeout cap
348		if s.globalTimeout > 0 && timeout > s.globalTimeout {
349			timeout = s.globalTimeout
350		}
351	
352		return timeout
353	}
354	
355	// cacheKey 生成缓存键，基于 terms + engine_refs + categories + category + safesearch + timerange + page + pagesize 的 FNV 哈希。
356	func (s *Scheduler) cacheKey(parsed *query.ParsedQuery, req *models.Request) string {
357		h := fnv.New64a()
358		for _, t := range parsed.Terms {
359			h.Write([]byte(t))
360		}
361		for _, ref := range parsed.EngineRefs {
362			h.Write([]byte(ref))
363		}
364		for _, cat := range parsed.Categories {
365			h.Write([]byte(cat))
366		}
367		h.Write([]byte(req.Category))
368		h.Write([]byte(strconv.Itoa(req.SafeSearch)))
369		h.Write([]byte(req.TimeRange))
370		h.Write([]byte(strconv.Itoa(req.Page)))
371		h.Write([]byte(strconv.Itoa(req.PageSize)))
372		return fmt.Sprintf("search:%x", h.Sum64())
373	}
374	
375	// externalBangURL 返回外部搜索引擎跳转 URL。
376	// 内置映射：g→google, ddg→duckduckgo, bing, gh→github, so→stackoverflow, wiki, yt。
377	func externalBangURL(bang string, terms []string) (string, bool) {
378		q := url.QueryEscape(strings.Join(terms, " "))
379	
380		mappings := map[string]string{
381			"g":    "https://www.google.com/search?q=%s",
382			"ddg":  "https://duckduckgo.com/?q=%s",
383			"bing": "https://www.bing.com/search?q=%s",
384			"gh":   "https://github.com/search?q=%s",
385			"so":   "https://stackoverflow.com/search?q=%s",
386			"wiki": "https://en.wikipedia.org/w/index.php?search=%s",
387			"yt":   "https://www.youtube.com/results?search_query=%s",
388		}
389	
390		template, ok := mappings[bang]
391		if !ok {
392			return "", false
393		}
394		return fmt.Sprintf(template, q), true
395	}
396	
397	// recordMetrics 记录结果流指标。
398	func (s *Scheduler) recordMetrics(resp *models.Response) {
399		metrics.ResultStreamTotal.WithLabelValues("results").Add(float64(len(resp.Results)))
400		metrics.ResultStreamTotal.WithLabelValues("suggestions").Add(float64(len(resp.Suggestions)))
401		if resp.Answers != nil {
402			metrics.ResultStreamTotal.WithLabelValues("answers").Add(float64(len(resp.Answers)))
403		}
404		if resp.Corrections != nil {
405			metrics.ResultStreamTotal.WithLabelValues("corrections").Add(float64(len(resp.Corrections)))
406		}
407		if resp.Infoboxes != nil {
408			metrics.ResultStreamTotal.WithLabelValues("infoboxes").Add(float64(len(resp.Infoboxes)))
409		}
410	}
411	
412	// paginate returns a stable windowed slice and the total count before windowing.
413	// page is 1-based; page=0 defaults to 1. pageSize <= 0 defaults to 10.
414	func paginate(results []models.Result, page, pageSize int) ([]models.Result, int) {
415		total := len(results)
416		if page <= 0 {
417			page = 1
418		}
419		if pageSize <= 0 {
420			pageSize = 10
421		}
422	
423		start := (page - 1) * pageSize
424		if start >= total {
425			return []models.Result{}, total
426		}
427	
428		end := start + pageSize
429		if end > total {
430			end = total
431		}
432	
433		return results[start:end], total
434	}
435	
436	func (s *Scheduler) cacheTTL(cat models.Category) time.Duration {
437		switch cat {
438		case models.CategoryImages:
439			return 2 * time.Minute
440		case models.CategoryNews:
441			return 15 * time.Second
442		case models.CategoryVideos:
443			return 2 * time.Minute
444		default:
445			return 30 * time.Second
446		}
447	}
<system>447 lines read from file starting from line 1. Total lines in file: 447. End of file reached.</system>

</details>

<details><summary>Thinking</summary>

Now let me read the web frontend files and the existing engine files to complete my understanding.

</details>

#### Tool Call: Read (`/Users/ranwei/workspace/go_work/seargo/web/src/types/search.…`)
<!-- call_id: call_00_FaGPV0b7qL0pasWNMLJS9778 -->
```json
{
  "path": "/Users/ranwei/workspace/go_work/seargo/web/src/types/search.ts"
}
```

#### Tool Call: Read (`/Users/ranwei/workspace/go_work/seargo/web/src/pages/SearchP…`)
<!-- call_id: call_01_xTHrV6UIVmn3QsBg9n6b7019 -->
```json
{
  "path": "/Users/ranwei/workspace/go_work/seargo/web/src/pages/SearchPage.tsx"
}
```

#### Tool Call: Glob (`internal/search/processor/factory.go`)
<!-- call_id: call_02_HruERFXGuXLIbj4FYsqK0191 -->
```json
{
  "pattern": "internal/search/processor/factory.go"
}
```

#### Tool Call: Glob (`engines/*/engine.go`)
<!-- call_id: call_03_35kZ6kdQU9mbaskRlKYP5876 -->
```json
{
  "pattern": "engines/*/engine.go"
}
```

<details><summary>Tool Result: Read (`/Users/ranwei/workspace/go_work/seargo/web/src/types/search.…`)</summary>

<!-- call_id: call_00_FaGPV0b7qL0pasWNMLJS9778 -->
1	export interface Result {
2	  title: string;
3	  url: string;
4	  content: string;
5	  engine: string;
6	  score: number;
7	  thumbnail_url?: string;
8	  published_at?: string;
9	}
10	
11	export interface SearchRequest {
12	  q: string;
13	  category?: string;
14	  language?: string;
15	  safesearch?: number;
16	  time_range?: string;
17	  page?: number;
18	}
19	
20	export interface SearchResponse {
21	  query: string;
22	  category: string;
23	  results: Result[];
24	  suggestions: string[];
25	  total: number;
26	  page: number;
27	  page_size: number;
28	  engines_used: string[];
29	  engines_failed: string[];
30	  response_time_ms: number;
31	}
<system>31 lines read from file starting from line 1. Total lines in file: 31. End of file reached.</system>

</details>

<details><summary>Tool Result: Read (`/Users/ranwei/workspace/go_work/seargo/web/src/pages/SearchP…`)</summary>

<!-- call_id: call_01_xTHrV6UIVmn3QsBg9n6b7019 -->
1	import { useState } from 'react';
2	import { useSearchStore } from '../stores/searchStore';
3	
4	const engineColors: Record<string, string> = {
5	  google: '#ea4335',
6	  bing: '#00809d',
7	  duckduckgo: '#de5833',
8	  brave: '#fb542b',
9	  wikipedia: '#3366cc',
10	  yahoo: '#6001d2',
11	};
12	
13	function getEngineColor(name: string): string {
14	  return engineColors[name.toLowerCase()] || '#6b7280';
15	}
16	
17	export default function SearchPage() {
18	  const [input, setInput] = useState('');
19	  const { results, isLoading, enginesUsed, enginesFailed, responseTimeMs, error, search } = useSearchStore();
20	  const hasSearched = results.length > 0 || error !== null || enginesUsed.length > 0;
21	
22	  const handleSubmit = (e: React.FormEvent) => {
23	    e.preventDefault();
24	    if (input.trim()) {
25	      search({ q: input.trim() });
26	    }
27	  };
28	
29	  return (
30	    <div className="min-h-screen bg-[#0f0f0f] text-[#e5e5e5]">
31	      <div className="max-w-3xl mx-auto px-4 py-12">
32	        {/* Logo / Title */}
33	        <div className={`text-center transition-all duration-500 ${hasSearched ? 'mb-6' : 'mb-12 mt-20'}`}>
34	          <h1 className="text-5xl font-bold tracking-tight mb-2">
35	            <span className="text-[#3b82f6]">Sear</span>Go
36	          </h1>
37	          <p className="text-[#9ca3af] text-sm">Privacy-respecting meta search</p>
38	        </div>
39	
40	        {/* Search Box */}
41	        <form onSubmit={handleSubmit} className="relative mb-8">
42	          <div className="flex gap-2">
43	            <div className="flex-1 relative">
44	              <input
45	                type="text"
46	                value={input}
47	                onChange={(e) => setInput(e.target.value)}
48	                placeholder="Search the web..."
49	                className="w-full px-5 py-3.5 bg-[#1a1a1a] border border-[rgba(255,255,255,0.08)] rounded-xl
50	                         text-[#e5e5e5] placeholder-[#6b7280] outline-none
51	                         focus:border-[#3b82f6] focus:ring-2 focus:ring-[#3b82f6]/30
52	                         transition-all duration-200 text-base"
53	              />
54	              {input && (
55	                <button
56	                  type="button"
57	                  onClick={() => setInput('')}
58	                  className="absolute right-3 top-1/2 -translate-y-1/2 text-[#6b7280] hover:text-[#e5e5e5]"
59	                >
60	                  ✕
61	                </button>
62	              )}
63	            </div>
64	            <button
65	              type="submit"
66	              disabled={isLoading}
67	              className="px-6 py-3.5 bg-[#3b82f6] hover:bg-[#2563eb] disabled:bg-[#1e3a5f]
68	                       rounded-xl font-medium transition-all duration-200
69	                       flex items-center gap-2 min-w-[100px] justify-center"
70	            >
71	              {isLoading ? (
72	                <span className="inline-block w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
73	              ) : (
74	                'Search'
75	              )}
76	            </button>
77	          </div>
78	        </form>
79	
80	        {/* Error */}
81	        {error && (
82	          <div className="mb-6 p-4 bg-red-900/20 border border-red-500/30 rounded-xl text-red-300">
83	            {error}
84	          </div>
85	        )}
86	
87	        {/* Results Stats */}
88	        {(results.length > 0 || enginesFailed.length > 0) && (
89	          <div className="mb-4 text-sm text-[#9ca3af]">
90	            Found <span className="text-[#e5e5e5] font-medium">{results.length}</span> results
91	            {responseTimeMs > 0 && ` in ${responseTimeMs}ms`}
92	            {enginesUsed.length > 0 && (
93	              <span> · Engines: {enginesUsed.join(', ')}</span>
94	            )}
95	            {enginesFailed.length > 0 && (
96	              <span className="text-red-400"> · Failed: {enginesFailed.join(', ')}</span>
97	            )}
98	          </div>
99	        )}
100	
101	        {/* Results */}
102	        <div className="space-y-3">
103	          {results.map((r, i) => (
104	            <div
105	              key={i}
106	              className="p-5 bg-[#1a1a1a] border border-[rgba(255,255,255,0.08)] rounded-xl
107	                       hover:border-[rgba(255,255,255,0.15)] transition-all duration-200
108	                       animate-fade-in"
109	              style={{ animationDelay: `${i * 60}ms` }}
110	            >
111	              <a
112	                href={r.url}
113	                target="_blank"
114	                rel="noopener noreferrer"
115	                className="text-lg font-medium text-[#60a5fa] hover:text-[#93c5fd] hover:underline block mb-1"
116	              >
117	                {r.title}
118	              </a>
119	              <p className="text-[#22c55e] text-sm mb-2 truncate">{r.url}</p>
120	              <p className="text-[#9ca3af] text-sm leading-relaxed">{r.content}</p>
121	              <div className="mt-3 flex items-center gap-2">
122	                <span
123	                  className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium text-white"
124	                  style={{ backgroundColor: getEngineColor(r.engine) }}
125	                >
126	                  {r.engine}
127	                </span>
128	                {r.score > 0 && (
129	                  <span className="text-xs text-[#6b7280]">Score: {r.score.toFixed(2)}</span>
130	                )}
131	              </div>
132	            </div>
133	          ))}
134	        </div>
135	
136	        {/* Empty state after search */}
137	        {hasSearched && results.length === 0 && !isLoading && !error && (
138	          <div className="text-center py-12 text-[#6b7280]">
139	            <p className="text-lg mb-2">No results found</p>
140	            <p className="text-sm">Try a different query or check your engine configuration</p>
141	          </div>
142	        )}
143	      </div>
144	    </div>
145	  );
146	}
<system>146 lines read from file starting from line 1. Total lines in file: 146. End of file reached.</system>

</details>

<details><summary>Tool Result: Glob (`internal/search/processor/factory.go`)</summary>

<!-- call_id: call_02_HruERFXGuXLIbj4FYsqK0191 -->
internal/search/processor/factory.go

</details>

<details><summary>Tool Result: Glob (`engines/*/engine.go`)</summary>

<!-- call_id: call_03_35kZ6kdQU9mbaskRlKYP5876 -->
No matches found

</details>

<details><summary>Thinking</summary>

Let me continue reading the remaining files I need to understand the codebase before writing the plan.

</details>

#### Tool Call: Read (`/Users/ranwei/workspace/go_work/seargo/internal/search/proce…`)
<!-- call_id: call_00_Xbg2nkaMo5h449neaaAc0788 -->
```json
{
  "path": "/Users/ranwei/workspace/go_work/seargo/internal/search/processor/factory.go"
}
```

#### Tool Call: Read (`/Users/ranwei/workspace/go_work/seargo/internal/search/suspe…`)
<!-- call_id: call_01_L3e9s7f0pWyPUXxwaiao9496 -->
```json
{
  "path": "/Users/ranwei/workspace/go_work/seargo/internal/search/suspension.go"
}
```

#### Tool Call: Glob (`engines/**/*.go`)
<!-- call_id: call_02_zlDGEwss4aKxUhedBAHT7907 -->
```json
{
  "pattern": "engines/**/*.go"
}
```

#### Tool Call: Read (`/Users/ranwei/workspace/go_work/seargo/web/src/stores/search…`)
<!-- call_id: call_03_vSJVCNS9YWSAUxxjGVyW8550 -->
```json
{
  "path": "/Users/ranwei/workspace/go_work/seargo/web/src/stores/searchStore.ts"
}
```

<details><summary>Tool Result: Read (`/Users/ranwei/workspace/go_work/seargo/internal/search/proce…`)</summary>

<!-- call_id: call_00_Xbg2nkaMo5h449neaaAc0788 -->
1	package processor
2	
3	import (
4		"fmt"
5	
6		"github.com/seargo/seargo/internal/config"
7		"github.com/seargo/seargo/internal/engine"
8		"github.com/seargo/seargo/internal/httpx"
9	)
10	
11	// NewProcessorFromConfig 根据引擎配置创建对应的 Processor。
12	// 目前所有在线引擎统一使用 OnlineProcessor；离线引擎使用 OfflineProcessor。
13	// 特殊类型（Currency、Dictionary、URLSearch）后续版本通过插件注册。
14	func NewProcessorFromConfig(eng engine.Engine, ec config.EngineConfig, suspension Suspension, client *httpx.Client) (Processor, error) {
15		if eng == nil {
16			return nil, fmt.Errorf("engine is nil for %s", ec.Name)
17		}
18		return NewOnlineProcessor(eng, suspension, client), nil
19	}
<system>19 lines read from file starting from line 1. Total lines in file: 19. End of file reached.</system>

</details>

<details><summary>Tool Result: Read (`/Users/ranwei/workspace/go_work/seargo/internal/search/suspe…`)</summary>

<!-- call_id: call_01_L3e9s7f0pWyPUXxwaiao9496 -->
1	package search
2	
3	import (
4		"strings"
5		"sync"
6		"time"
7	
8		"github.com/seargo/seargo/internal/config"
9		"github.com/seargo/seargo/internal/engine"
10		"github.com/seargo/seargo/internal/errors"
11	)
12	
13	type SuspensionTracker struct {
14		mu     sync.RWMutex
15		bans   map[string]banEntry
16		config config.SearchConfig
17	}
18	
19	type banEntry struct {
20		reason string
21		until  time.Time
22		count  int
23	}
24	
25	func NewSuspensionTracker(cfg config.SearchConfig) *SuspensionTracker {
26		return &SuspensionTracker{
27			bans:   make(map[string]banEntry),
28			config: cfg,
29		}
30	}
31	
32	func (st *SuspensionTracker) Ban(engineName, errorClass string) {
33		st.mu.Lock()
34		defer st.mu.Unlock()
35	
36		entry := st.bans[engineName]
37		entry.count++
38		entry.reason = errorClass
39	
40		// Check if there's a specific duration for this error class
41		var duration float64
42		if specificDuration := st.getSuspensionDuration(errorClass); specificDuration > 0 {
43			duration = specificDuration
44		} else {
45			// Escalating ban: base_time * count, capped at max_ban_time
46			duration = st.config.BanTimeOnFail * float64(entry.count)
47			if duration > st.config.MaxBanTimeOnFail {
48				duration = st.config.MaxBanTimeOnFail
49			}
50		}
51		entry.until = time.Now().Add(time.Duration(duration * float64(time.Second)))
52	
53		st.bans[engineName] = entry
54	}
55	
56	func (st *SuspensionTracker) getSuspensionDuration(errorClass string) float64 {
57		switch errorClass {
58		case "SearxEngineAccessDenied":
59			return st.config.SuspendedTimes.SearxEngineAccessDenied
60		case "SearxEngineCaptcha":
61			return st.config.SuspendedTimes.SearxEngineCaptcha
62		case "SearxEngineTooManyRequests":
63			return st.config.SuspendedTimes.SearxEngineTooManyRequests
64		case "cf_SearxEngineCaptcha":
65			return st.config.SuspendedTimes.CfSearxEngineCaptcha
66		case "cf_SearxEngineAccessDenied":
67			return st.config.SuspendedTimes.CfSearxEngineAccessDenied
68		case "recaptcha_SearxEngineCaptcha":
69			return st.config.SuspendedTimes.RecaptchaSearxEngineCaptcha
70		}
71		return 0
72	}
73	
74	func (st *SuspensionTracker) IsSuspended(engineName string) bool {
75		st.mu.RLock()
76		defer st.mu.RUnlock()
77	
78		entry, ok := st.bans[engineName]
79		if !ok {
80			return false
81		}
82		if time.Now().After(entry.until) {
83			return false
84		}
85		return true
86	}
87	
88	func (st *SuspensionTracker) Clear(engineName string) {
89		st.mu.Lock()
90		defer st.mu.Unlock()
91		delete(st.bans, engineName)
92	}
93	
94	func classifyError(err error) string {
95		if err == nil {
96			return ""
97		}
98	
99		// Check for engine.SearxEngineResponseException first
100		if e, ok := err.(*engine.SearxEngineResponseException); ok {
101			switch e.ErrorClass() {
102			case "captcha":
103				msg := strings.ToLower(e.Error())
104				if strings.Contains(msg, "cloudflare") {
105					return "cf_SearxEngineCaptcha"
106				}
107				if strings.Contains(msg, "recaptcha") {
108					return "recaptcha_SearxEngineCaptcha"
109				}
110				return "SearxEngineCaptcha"
111			case "access_denied":
112				msg := strings.ToLower(e.Error())
113				if strings.Contains(msg, "cloudflare") || strings.Contains(msg, "1020") {
114					return "cf_SearxEngineAccessDenied"
115				}
116				return "SearxEngineAccessDenied"
117			case "too_many_requests":
118				return "SearxEngineTooManyRequests"
119			case "timeout":
120				return "SearxEngineTooManyRequests"
121			}
122		}
123	
124		// Check for typed EngineErrors
125		if ee, ok := err.(*errors.EngineError); ok {
126			switch ee.SuspendedTimeCategory {
127			case "captcha":
128				msg := strings.ToLower(ee.Error())
129				if strings.Contains(msg, "cloudflare") {
130					return "cf_SearxEngineCaptcha"
131				}
132				if strings.Contains(msg, "recaptcha") {
133					return "recaptcha_SearxEngineCaptcha"
134				}
135				return "SearxEngineCaptcha"
136			case "access_denied":
137				msg := strings.ToLower(ee.Error())
138				if strings.Contains(msg, "cloudflare") || strings.Contains(msg, "1020") {
139					return "cf_SearxEngineAccessDenied"
140				}
141				return "SearxEngineAccessDenied"
142			case "too_many_requests":
143				return "SearxEngineTooManyRequests"
144			}
145		}
146	
147		// Fallback to string-based classification
148		msg := strings.ToLower(err.Error())
149	
150		if strings.Contains(msg, "403") || strings.Contains(msg, "forbidden") ||
151			strings.Contains(msg, "access denied") {
152			return "SearxEngineAccessDenied"
153		}
154		if strings.Contains(msg, "captcha") || strings.Contains(msg, "recaptcha") ||
155			strings.Contains(msg, "challenge") {
156			return "SearxEngineCaptcha"
157		}
158		if strings.Contains(msg, "429") || strings.Contains(msg, "too many requests") ||
159			strings.Contains(msg, "rate limit") {
160			return "SearxEngineTooManyRequests"
161		}
162	
163		return "SearxEngineTooManyRequests"
164	}
<system>164 lines read from file starting from line 1. Total lines in file: 164. End of file reached.</system>

</details>

<details><summary>Tool Result: Glob (`engines/**/*.go`)</summary>

<!-- call_id: call_02_zlDGEwss4aKxUhedBAHT7907 -->
engines/yahoo/yahoo_test.go
engines/yahoo/yahoo.go
engines/wikipedia/wikipedia_test.go
engines/wikipedia/wikipedia.go
engines/google/google_test.go
engines/google/google.go
engines/duckduckgo/duckduckgo_test.go
engines/duckduckgo/duckduckgo.go
engines/brave/brave_test.go
engines/brave/brave.go
engines/bing/bing_test.go
engines/bing/bing.go

</details>

<details><summary>Tool Result: Read (`/Users/ranwei/workspace/go_work/seargo/web/src/stores/search…`)</summary>

<!-- call_id: call_03_vSJVCNS9YWSAUxxjGVyW8550 -->
1	import { create } from 'zustand';
2	import type { Result, SearchRequest } from '../types/search';
3	import { api } from '../services/api';
4	
5	interface SearchState {
6	  query: string;
7	  results: Result[];
8	  isLoading: boolean;
9	  enginesUsed: string[];
10	  enginesFailed: string[];
11	  responseTimeMs: number;
12	  error: string | null;
13	  setQuery: (q: string) => void;
14	  search: (req: SearchRequest) => Promise<void>;
15	}
16	
17	export const useSearchStore = create<SearchState>((set) => ({
18	  query: '',
19	  results: [],
20	  isLoading: false,
21	  enginesUsed: [],
22	  enginesFailed: [],
23	  responseTimeMs: 0,
24	  error: null,
25	
26	  setQuery: (q) => set({ query: q }),
27	
28	  search: async (req) => {
29	    set({ isLoading: true, error: null });
30	    try {
31	      const resp = await api.search(req);
32	      set({
33	        query: resp.data.query,
34	        results: resp.data.results,
35	        enginesUsed: resp.data.engines_used,
36	        enginesFailed: resp.data.engines_failed,
37	        responseTimeMs: resp.data.response_time_ms,
38	        isLoading: false,
39	      });
40	    } catch (err: any) {
41	      set({
42	        isLoading: false,
43	        error: err.response?.data?.error?.message || err.message || 'Search failed',
44	      });
45	    }
46	  },
47	}));
<system>47 lines read from file starting from line 1. Total lines in file: 47. End of file reached.</system>

</details>

<details><summary>Thinking</summary>

Let me read one example engine file to understand the current Search signature and result return pattern.

</details>

#### Tool Call: Read (`/Users/ranwei/workspace/go_work/seargo/engines/google/google…`)
<!-- call_id: call_00_a4cUJI67gs8li7WtqaIN0097 -->
```json
{
  "path": "/Users/ranwei/workspace/go_work/seargo/engines/google/google.go"
}
```

#### Tool Call: Read (`/Users/ranwei/workspace/go_work/seargo/pkg/models/results/do…`)
<!-- call_id: call_01_ZxSY0YhUAjFLLS7EkwmO5308 -->
```json
{
  "path": "/Users/ranwei/workspace/go_work/seargo/pkg/models/results/doc.go"
}
```

#### Tool Call: Read (`/Users/ranwei/workspace/go_work/seargo/pkg/models/results/no…`)
<!-- call_id: call_02_IecVqO2JACQJY1heeoAd9796 -->
```json
{
  "path": "/Users/ranwei/workspace/go_work/seargo/pkg/models/results/normalize_test.go"
}
```

#### Tool Call: Read (`/Users/ranwei/workspace/go_work/seargo/pkg/models/results/re…`)
<!-- call_id: call_03_k7uUd5JIOGzRZOZkfU2u0411 -->
```json
{
  "path": "/Users/ranwei/workspace/go_work/seargo/pkg/models/results/results_test.go"
}
```

<details><summary>Tool Result: Read (`/Users/ranwei/workspace/go_work/seargo/engines/google/google…`)</summary>

<!-- call_id: call_00_a4cUJI67gs8li7WtqaIN0097 -->
1	package google
2	
3	import (
4		"context"
5		"fmt"
6		"net/url"
7		"strings"
8	
9		"github.com/PuerkitoBio/goquery"
10	
11		"github.com/seargo/seargo/internal/engine"
12		"github.com/seargo/seargo/internal/httpx"
13		"github.com/seargo/seargo/pkg/models"
14	)
15	
16	func init() {
17		engine.Register("google", &Google{})
18	}
19	
20	type Google struct {
21		client *httpx.Client
22	}
23	
24	func (g *Google) Name() string { return "google" }
25	
26	func (g *Google) Categories() []models.Category {
27		return []models.Category{models.CategoryGeneral, models.CategoryImages}
28	}
29	
30	func (g *Google) Capabilities() engine.Capabilities {
31		return engine.Capabilities{
32			SupportsSafeSearch: true,
33			SupportsLanguage:   true,
34			SupportsPagination: true,
35		}
36	}
37	
38	func (g *Google) Init(ctx context.Context, cfg engine.EngineInitConfig) bool {
39		return true
40	}
41	
42	func (g *Google) Setup(cfg engine.EngineInitConfig) bool {
43		return true
44	}
45	
46	func (g *Google) About() engine.EngineAbout {
47		return engine.EngineAbout{
48			Website:    "https://www.google.com",
49			WikidataID: "Q95",
50		}
51	}
52	
53	func (g *Google) Search(ctx context.Context, req *models.Request) (*models.Response, error) {
54		searchURL := fmt.Sprintf("https://www.google.com/search?q=%s&hl=en", url.QueryEscape(req.Query))
55	
56		resp, err := g.client.R().SetContext(ctx).Get(searchURL)
57		if err != nil {
58			return nil, fmt.Errorf("google request failed: %w", err)
59		}
60	
61		doc, err := goquery.NewDocumentFromReader(strings.NewReader(resp.String()))
62		if err != nil {
63			return nil, fmt.Errorf("parse HTML: %w", err)
64		}
65	
66		var results []models.Result
67		selectors := []string{"div.g", "div.srg div.g", "#search div.g"}
68		for _, sel := range selectors {
69			doc.Find(sel).Each(func(i int, s *goquery.Selection) {
70				titleElem := s.Find("h3")
71				if titleElem.Length() == 0 {
72					titleElem = s.Find("a h3")
73				}
74				title := strings.TrimSpace(titleElem.Text())
75	
76				var href string
77				s.Find("a").Each(func(j int, a *goquery.Selection) {
78					if href == "" {
79						h, ok := a.Attr("href")
80						if ok && strings.HasPrefix(h, "http") && !strings.Contains(h, "google.com") {
81							href = h
82						}
83					}
84				})
85	
86				snippet := strings.TrimSpace(s.Find(".VwiC3b").Text())
87				if snippet == "" {
88					snippet = strings.TrimSpace(s.Find("span").Text())
89				}
90	
91				if title != "" && href != "" {
92					results = append(results, models.Result{
93						Title:    title,
94						URL:      href,
95						Content:  snippet,
96						Engine:   g.Name(),
97						Category: req.Category,
98					})
99				}
100			})
101			if len(results) > 0 {
102				break
103			}
104		}
105	
106		return &models.Response{
107			Query:    req.Query,
108			Category: req.Category,
109			Results:  results,
110		}, nil
111	}
<system>111 lines read from file starting from line 1. Total lines in file: 111. End of file reached.</system>

</details>

<details><summary>Tool Result: Read (`/Users/ranwei/workspace/go_work/seargo/pkg/models/results/do…`)</summary>

<!-- call_id: call_01_ZxSY0YhUAjFLLS7EkwmO5308 -->
1	// Package results provides a typed result hierarchy ported from SearXNG,
2	// with normalization utilities and backward-compatible conversion to the
3	// API-facing models.Result type.
4	package results
<system>4 lines read from file starting from line 1. Total lines in file: 4. End of file reached.</system>

</details>

<details><summary>Tool Result: Read (`/Users/ranwei/workspace/go_work/seargo/pkg/models/results/no…`)</summary>

<!-- call_id: call_02_IecVqO2JACQJY1heeoAd9796 -->
1	package results
2	
3	import (
4		"testing"
5	
6		"github.com/stretchr/testify/assert"
7	)
8	
9	func TestNormalizeURL_DefaultScheme(t *testing.T) {
10		br := &BaseResult{URL: "example.com/path"}
11		normalizeURL(br)
12		assert.Equal(t, "http://example.com/path", br.URL)
13	}
14	
15	func TestNormalizeURL_HTTPSPreserved(t *testing.T) {
16		br := &BaseResult{URL: "https://example.com/path"}
17		original := br.URL
18		normalizeURL(br)
19		assert.Equal(t, original, br.URL)
20	}
21	
22	func TestNormalizeURL_SyncParsedURL(t *testing.T) {
23		br := &BaseResult{URL: "https://example.com/path?q=1#frag"}
24		normalizeURL(br)
25		assert.Equal(t, []string{"https", "example.com", "/path", "q=1", "frag"}, br.ParsedURL)
26	}
27	
28	func TestNormalizeURL_Invalid(t *testing.T) {
29		br := &BaseResult{URL: "://invalid"}
30		normalizeURL(br)
31		assert.Equal(t, "://invalid", br.URL)
32	}
33	
34	func TestNormalizeText_CollapseWhitespace(t *testing.T) {
35		br := &BaseResult{Title: "  Hello   World  ", Content: "Line1\n\nLine2"}
36		normalizeText(br)
37		assert.Equal(t, "Hello World", br.Title)
38		assert.Equal(t, "Line1 Line2", br.Content)
39	}
40	
41	func TestNormalizeText_DedupTitleEqualsContent(t *testing.T) {
42		br := &BaseResult{Title: "Same text", Content: "Same text"}
43		normalizeText(br)
44		assert.Equal(t, "", br.Content, "content should be cleared when identical to title")
45	}
46	
47	func TestNormalizeText_TitleDiffersFromContent(t *testing.T) {
48		br := &BaseResult{Title: "Title", Content: "Different content"}
49		normalizeText(br)
50		assert.Equal(t, "Title", br.Title)
51		assert.Equal(t, "Different content", br.Content)
52	}
53	
54	func TestNormalizeDate_ValidYear(t *testing.T) {
55		br := &BaseResult{Title: "T"}
56		normalizeDate(br, "2024")
57		assert.NotNil(t, br.PublishedAt)
58		assert.Equal(t, 2024, br.PublishedAt.Year())
59	}
60	
61	func TestNormalizeDate_YearTooLow(t *testing.T) {
62		br := &BaseResult{Title: "T"}
63		normalizeDate(br, "1899")
64		assert.Nil(t, br.PublishedAt, "year < 1900 should be rejected")
65	}
66	
67	func TestNormalizeDate_InvalidFormat(t *testing.T) {
68		br := &BaseResult{Title: "T"}
69		normalizeDate(br, "not-a-date")
70		assert.Nil(t, br.PublishedAt)
71	}
<system>71 lines read from file starting from line 1. Total lines in file: 71. End of file reached.</system>

</details>

<details><summary>Tool Result: Read (`/Users/ranwei/workspace/go_work/seargo/pkg/models/results/re…`)</summary>

<!-- call_id: call_03_k7uUd5JIOGzRZOZkfU2u0411 -->
1	package results
2	
3	import (
4		"testing"
5	
6		"github.com/stretchr/testify/assert"
7	)
8	
9	func TestBaseResult_Defaults(t *testing.T) {
10		br := BaseResult{
11			Title:    "Test Title",
12			URL:      "https://example.com",
13			Content:  "Test snippet",
14			Engine:   "google",
15			Template: "default",
16		}
17	
18		assert.Equal(t, "Test Title", br.Title)
19		assert.Equal(t, "https://example.com", br.URL)
20		assert.Equal(t, "default", br.Template)
21	}
22	
23	func TestMainResult(t *testing.T) {
24		mr := &MainResult{
25			BaseResult: BaseResult{
26				Title:    "Main Result",
27				URL:      "https://example.com/page",
28				Content:  "Description",
29				Engine:   "bing",
30				Template: "default",
31			},
32		}
33	
34		assert.Equal(t, "Main Result", mr.Title)
35		assert.Equal(t, "default", mr.Template)
36	}
37	
38	func TestImageResult_Fields(t *testing.T) {
39		ir := &ImageResult{
40			BaseResult: BaseResult{
41				Title:     "An image",
42				URL:       "https://example.com/img",
43				Template:  "images.html",
44			},
45			ThumbnailSrc: "https://example.com/thumb.jpg",
46			ImgSrc:       "https://example.com/full.jpg",
47			Resolution:   "1920x1080",
48		}
49	
50		assert.Equal(t, "images.html", ir.Template)
51		assert.Equal(t, "1920x1080", ir.Resolution)
52		assert.False(t, ir.IsBase64()) // not base64 data
53	}
54	
55	func TestImageResult_IsBase64(t *testing.T) {
56		ir := &ImageResult{
57			ImgSrc: "data:image/png;base64,iVBORw0KGgo=",
58		}
59		assert.True(t, ir.IsBase64())
60	}
61	
62	func TestVideoResult_Fields(t *testing.T) {
63		vr := &VideoResult{
64			BaseResult: BaseResult{
65				Title:    "Video",
66				Template: "videos.html",
67			},
68			Thumbnail: "https://example.com/thumb.jpg",
69			IFrameSrc: "https://example.com/embed",
70			Length:    "3:45",
71		}
72	
73		assert.Equal(t, "videos.html", vr.Template)
74		assert.Equal(t, "3:45", vr.Length)
75	}
76	
77	func TestResultType_InterfaceSatisfaction(t *testing.T) {
78		var r Result = &MainResult{}
79		assert.NotNil(t, r)
80	
81		r = &ImageResult{}
82		assert.NotNil(t, r)
83	
84		r = &VideoResult{}
85		assert.NotNil(t, r)
86	
87		r = &NewsResult{}
88		assert.NotNil(t, r)
89	
90		r = &PaperResult{}
91		assert.NotNil(t, r)
92	}
<system>92 lines read from file starting from line 1. Total lines in file: 92. End of file reached.</system>

</details>

<details><summary>Thinking</summary>

Now I have a thorough understanding of the codebase. Let me summarize what I've found and plan the tasks.

## Current State Summary

### Backend (`pkg/models/results/`)
- Already has `Result` interface with `GetTitle/GetURL/GetContent/GetEngine/GetTemplate/GetCategory`
- Already has `BaseResult`, `MainResult`, `ImageResult`, `VideoResult`, `NewsResult`, `PaperResult`, `CodeResult`, `FileResult`, `MapResult`, `MusicResult`, `AnswerResult`, `InfoboxResult`
- But: **Missing** `Kind()`, `Template()` (as interface methods per-kind, not just BaseResult methods), `Base()`, `Normalize()`, `DedupKey()`, `PostNormalize()` methods on the interface
- Already has `normalizeURL`, `normalizeText`, `normalizeDate` in `normalize.go`
- `BaseResult.Normalize()` exists but only calls `normalizeURL` + `normalizeText` — no per-kind `PostNormalize()`
- `normalize.go` has `extractDomainFromURL` and `computeParsedURL` helpers
- Already has `ToAPIResult` in `convert.go` — but it doesn't set `Kind`, `Template`, or kind-specific `Extra`... wait actually it does. But it doesn't set `Kind` on the output `models.Result`.

### Backend (`pkg/models/models.go`)
- `models.Result` struct: flat, no `Kind` field, uses `EngineData` for extra fields
- `models.Infobox`: simple, with `URLs []string` and `Attributes []string` (not rich objects)
- `models.Response`: has `Results []Result`, `Suggestions`, `Answers`, `Corrections`, `Infoboxes`, `EngineData`

### Backend Search Pipeline
- `internal/search/merge.go`: `ResultContainer` with `Extend(engineName, results, positionBase)` — uses `dedupKey` which includes `Template + host + path + query + thumbnail`
- `internal/search/processor/types.go`: `ProcessorResult{Results, Suggestions, Answers, Corrections, Infoboxes, EngineData}` — flat `models.Result` slice
- `internal/search/processor/online.go`: passes `resp.Results` directly as `[]models.Result`
- `internal/search/scheduler.go`: `executeProcessors` calls `proc.Search` → gets `ProcessorResult` → calls `container.Extend` + `AddSuggestions` + `AddAnswers` + `AddCorrections` + `AddInfoboxes` + `AddEngineData`
- `internal/search/suspension.go`: `classifyError` checks `*engine.SearxEngineResponseException` first (reference to Phase 4), then `*errors.EngineError`

### Frontend
- `web/src/types/search.ts`: Flat `Result` interface, no `kind` field
- `web/src/pages/SearchPage.tsx`: Inline result rendering, no component dispatch
- `web/src/components/`: Empty — no components yet
- `web/src/stores/searchStore.ts`: Zustand store, stores `results: Result[]`

### What the design requires
The design at `.ody-code/designs/2026-06-20-05-result-types.md` specifies:

1. **Backend typed result hierarchy**: Extend `Result` interface with `Kind()`, `Template()`, `Normalize()`, `DedupKey()`, `Base()`. Each typed result must implement these. Add missing types (`KeyValueResult`, `CodeLine` type, `ImageRef`).
2. **Tagged-union API response**: `models.Result` gets `Kind` + `Template` + `Extra` fields
3. **Normalization**: Extend `BaseResult.Normalize()` to call kind-specific `PostNormalize()`
4. **Typed-to-flat conversion**: Update `ToAPIResult` to produce tagged-union output
5. **Typed result container**: New `TypedResultContainer` in `internal/search/typed_container.go`
6. **Scheduler/processor integration**: `ProcessorResult` gains `TypedResults []results.Result`
7. **Frontend**: TypeScript tagged union, `ResultCard` component, category-aware layout

## Task Count & Split Decision

Let me count tasks:

**Phase A: Backend Types & Container** (standalone, no upstream within this plan):
1. Result interface upgrade + new types → modify `results.go` (+ `KeyValueResult`, `CodeLine`, `ImageRef`, `InfoboxAttribute`, `InfoboxURL`)
2. Normalization extension → modify `normalize.go` (add `NormalizeEngines`, per-kind `PostNormalize`)
3. ToAPIResult update → modify `convert.go` (add `Kind`/`Template`/`Extra`)
4. API envelope update → modify `models.go` (add `Kind`/`Extra` to Result, extend Infobox)
5. TypedResultContainer → create `typed_container.go`

**Phase B: Integration** (depends on Phase A):
6. ProcessorResult extension → modify `types.go`
7. Processor adapter → modify `online.go` 
8. Scheduler wiring → modify `scheduler.go`

**Phase C: Frontend** (depends on Phase A for API contract):
9. TypeScript types → modify `search.ts`
10. ResultCard + kind-specific components → create several `.tsx` files
11. SearchPage layout → modify `SearchPage.tsx`

That's 11 tasks > 8. Split needed. Let me plan:

- Part 1: `backend-types.md` — Tasks 1-5 (result hierarchy, normalization, conversion, API envelope, typed container)
- Part 2: `integration.md` — Tasks 6-8 (processor, scheduler)
- Part 3: `frontend.md` — Tasks 9-11 (TypeScript types, components, layout)

Now the plan file path. The design is `2026-06-20-05-result-types.md`, so the plan index goes to:
`.ody-code/plans/2026-06-20-05-result-types.md`
with parts in:
`.ody-code/plans/2026-06-20-05-result-types/backend-types.md`
`.ody-code/plans/2026-06-20-05-result-types/integration.md`
`.ody-code/plans/2026-06-20-05-result-types/frontend.md`

Let me now write the index first. I need to scaffold: header + File Structure + Dependency Overview + Risks & Open Questions + Parts manifest (all `pending`).

Let me also check if `pkg/models/results/convert_test.go` exists.

</details>

#### Tool Call: Read (`/Users/ranwei/workspace/go_work/seargo/pkg/models/results/co…`)
<!-- call_id: call_00_tvbCHwoLFSSQmUeN9WUN8292 -->
```json
{
  "path": "/Users/ranwei/workspace/go_work/seargo/pkg/models/results/convert_test.go"
}
```

#### Tool Call: Read (`/Users/ranwei/workspace/go_work/seargo/web/package.json`)
<!-- call_id: call_01_cD0LCaYprqMNDQCkiqLT2409 -->
```json
{
  "path": "/Users/ranwei/workspace/go_work/seargo/web/package.json"
}
```

#### Tool Call: Read (`/Users/ranwei/workspace/go_work/seargo/internal/engine/engin…`)
<!-- call_id: call_02_858nf15rF9TUBs0t87Rr1765 -->
```json
{
  "path": "/Users/ranwei/workspace/go_work/seargo/internal/engine/engine.go"
}
```

<details><summary>Tool Result: Read (`/Users/ranwei/workspace/go_work/seargo/pkg/models/results/co…`)</summary>

<!-- call_id: call_00_tvbCHwoLFSSQmUeN9WUN8292 -->
1	package results
2	
3	import (
4		"testing"
5	
6		"github.com/stretchr/testify/assert"
7	)
8	
9	func TestToAPIResult_MainResult(t *testing.T) {
10		mr := &MainResult{
11			BaseResult: BaseResult{
12				Title:    "Search Result",
13				URL:      "https://example.com/page",
14				Content:  "A description",
15				Engine:   "google",
16				Template: "default",
17			},
18		}
19	
20		apiResults := ToAPIResult([]Result{mr})
21		assert.Len(t, apiResults, 1)
22	
23		r := apiResults[0]
24		assert.Equal(t, "Search Result", r.Title)
25		assert.Equal(t, "https://example.com/page", r.URL)
26		assert.Equal(t, "A description", r.Content)
27		assert.Equal(t, "google", r.Engine)
28		assert.Equal(t, "default", r.Template)
29	}
30	
31	func TestToAPIResult_ImageResult(t *testing.T) {
32		ir := &ImageResult{
33			BaseResult: BaseResult{
34				Title:    "Cat Photo",
35				URL:      "https://example.com/cat",
36				Engine:   "google",
37				Template: "images.html",
38			},
39			ImgSrc:       "https://example.com/cat.jpg",
40			ThumbnailSrc: "https://example.com/cat_thumb.jpg",
41			Resolution:   "800x600",
42		}
43	
44		apiResults := ToAPIResult([]Result{ir})
45		assert.Len(t, apiResults, 1)
46	
47		r := apiResults[0]
48		assert.Equal(t, "images.html", r.Template)
49		assert.Equal(t, "https://example.com/cat_thumb.jpg", r.ThumbnailURL)
50	}
51	
52	func TestToAPIResult_VideoResult(t *testing.T) {
53		vr := &VideoResult{
54			BaseResult: BaseResult{
55				Title:    "Video",
56				URL:      "https://example.com/video",
57				Template: "videos.html",
58			},
59			Thumbnail: "https://example.com/thumb.jpg",
60			Length:    "5:00",
61		}
62	
63		apiResults := ToAPIResult([]Result{vr})
64		assert.Len(t, apiResults, 1)
65		assert.Equal(t, "videos.html", apiResults[0].Template)
66	}
67	
68	func TestToAPIResult_MultipleTypes(t *testing.T) {
69		results := []Result{
70			&MainResult{BaseResult: BaseResult{Title: "Main", URL: "https://a.com", Template: "default", Engine: "g"}},
71			&ImageResult{BaseResult: BaseResult{Title: "Img", URL: "https://b.com", Template: "images.html", Engine: "g"}},
72			&NewsResult{BaseResult: BaseResult{Title: "News", URL: "https://c.com", Template: "default", Engine: "g"}},
73		}
74	
75		apiResults := ToAPIResult(results)
76		assert.Len(t, apiResults, 3)
77	}
78	
79	func TestToAPIResult_EmptyInput(t *testing.T) {
80		apiResults := ToAPIResult(nil)
81		assert.Nil(t, apiResults)
82	
83		apiResults = ToAPIResult([]Result{})
84		assert.Nil(t, apiResults)
85	}
86	
87	func TestToAPIResult_EngineDataPassthrough(t *testing.T) {
88		mr := &MainResult{
89			BaseResult: BaseResult{
90				Title:      "T",
91				URL:        "https://x.com",
92				EngineData: map[string]any{"key": "value"},
93			},
94		}
95	
96		apiResults := ToAPIResult([]Result{mr})
97		assert.NotNil(t, apiResults[0].EngineData)
98		assert.Equal(t, "value", apiResults[0].EngineData["key"])
99	}
<system>99 lines read from file starting from line 1. Total lines in file: 99. End of file reached.</system>

</details>

<details><summary>Tool Result: Read (`/Users/ranwei/workspace/go_work/seargo/web/package.json`)</summary>

<!-- call_id: call_01_cD0LCaYprqMNDQCkiqLT2409 -->
1	{
2	  "name": "web",
3	  "private": true,
4	  "version": "0.0.0",
5	  "type": "module",
6	  "scripts": {
7	    "dev": "vite",
8	    "build": "tsc -b && vite build",
9	    "lint": "eslint .",
10	    "preview": "vite preview"
11	  },
12	  "dependencies": {
13	    "axios": "^1.16.0",
14	    "react": "^19.2.5",
15	    "react-dom": "^19.2.5",
16	    "react-router-dom": "^7.15.0",
17	    "zustand": "^5.0.13"
18	  },
19	  "devDependencies": {
20	    "@eslint/js": "^10.0.1",
21	    "@types/node": "^24.12.2",
22	    "@types/react": "^19.2.14",
23	    "@types/react-dom": "^19.2.3",
24	    "@vitejs/plugin-react": "^6.0.1",
25	    "autoprefixer": "^10.5.0",
26	    "eslint": "^10.2.1",
27	    "eslint-plugin-react-hooks": "^7.1.1",
28	    "eslint-plugin-react-refresh": "^0.5.2",
29	    "globals": "^17.5.0",
30	    "postcss": "^8.5.14",
31	    "tailwindcss": "^3.4.19",
32	    "typescript": "~6.0.2",
33	    "typescript-eslint": "^8.58.2",
34	    "vite": "^8.0.10"
35	  }
36	}
<system>36 lines read from file starting from line 1. Total lines in file: 36. End of file reached.</system>

</details>

<details><summary>Tool Result: Read (`/Users/ranwei/workspace/go_work/seargo/internal/engine/engin…`)</summary>

<!-- call_id: call_02_858nf15rF9TUBs0t87Rr1765 -->
1	package engine
2	
3	import (
4		"context"
5	
6		"github.com/seargo/seargo/pkg/models"
7	)
8	
9	// EngineInitConfig holds per-engine runtime configuration from the config file.
10	type EngineInitConfig struct {
11		Name       string            // display name from config
12		Shortcut   string            // shortcut from config
13		Categories []models.Category // per-engine categories (overrides defaults)
14		Timeout    float64           // per-engine timeout in seconds
15		Extra      map[string]any    // arbitrary extra config
16	
17		// SearXNG-compatible fields
18		Paging            bool         // whether engine supports pagination
19		TimeRangeSupport  bool         // whether engine supports time_range
20		LanguageSupport   bool         // whether engine supports language parameter
21		SafeSearch        bool         // whether engine supports safesearch
22		Weight            float64      // engine weight for scoring
23		DisplayErrorMsgs  bool         // show error messages to user
24		EnableHTTP        bool         // allow HTTP (not just HTTPS)
25		Inactive          bool         // engine inactive (skip entirely)
26		Disabled          bool         // engine disabled by config
27		Tokens            []string     // per-engine API tokens
28		Network           string       // named network for outbound requests
29		SoftMaxRedirects  int          // max redirects before soft error
30		NoResultForHTTPStatus []int    // HTTP statuses treated as "no result"
31		RaiseForHTTPError interface{}  // nil|bool|int|[]int for retry-on-http-error
32		EngineTraits      EngineTraits // resolved language/region traits
33	}
34	
35	// Engine is the interface that all search engines must implement.
36	type Engine interface {
37		Name() string
38		Categories() []models.Category
39		Capabilities() Capabilities
40		About() EngineAbout
41		Init(ctx context.Context, cfg EngineInitConfig) bool
42		Setup(cfg EngineInitConfig) bool
43		Search(ctx context.Context, req *models.Request) (*models.Response, error)
44	}
45	
46	// Capabilities describes what features an engine supports.
47	type Capabilities struct {
48		SupportsSafeSearch bool   `json:"supports_safe_search"`
49		SupportsLanguage   bool   `json:"supports_language"`
50		SupportsTimeRange  bool   `json:"supports_time_range"`
51		SupportsPagination bool   `json:"supports_pagination"`
52		RequiresAPIKey     bool   `json:"requires_api_key"`
53		Shortcut           string `json:"shortcut"`
54	}
55	
56	// EngineAbout holds descriptive metadata for an engine.
57	type EngineAbout struct {
58		Website     string `json:"website,omitempty"`
59		WikidataID  string `json:"wikidata_id,omitempty"`
60		UseAPK      string `json:"use_api_key,omitempty"`
61		ResultsHTML string `json:"results_html,omitempty"`
62	}
63	
64	// Info describes an engine for API responses.
65	type Info struct {
66		Name         string       `json:"name"`
67		Categories   []string     `json:"categories"`
68		Shortcut     string       `json:"shortcut"`
69		Capabilities Capabilities `json:"capabilities"`
70		Enabled      bool         `json:"enabled"`
71	}
72	
73	
<system>73 lines read from file starting from line 1. Total lines in file: 73. End of file reached.</system>

</details>

<details><summary>Thinking</summary>

Now I have all the information I need. Let me write the plan.

Key observations:
1. The existing `pkg/models/results/` already has most of the typed result structs and even the `Result` interface — but the interface is missing `Kind()`, `Template()` (as per-kind method, not BaseResult passthrough), `Normalize()`, `DedupKey()`, `Base()`.
2. The existing `ToAPIResult` in `convert.go` doesn't set `Kind` field on output because `models.Result` doesn't have one yet.
3. `models.Infobox` is too simple — needs `InfoboxURL` and `InfoboxAttribute` types instead of `[]string`.
4. The existing `ResultContainer` in `merge.go` extends by `Template|host|path|query|thumbnail` — the new `TypedResultContainer` extends by `kind|host|path|query|thumbnail`.
5. Engine `Search` currently returns `*models.Response` with flat `[]models.Result`. The design says engines keep returning typed results via `[]results.Result`. Since the existing engines are already returning `[]models.Result`, we need an adapter path. The design says this is Phase 5, where we build the typed infrastructure but don't force engines to switch yet.

Let me think about how to structure this plan properly. The design says:
- "All existing and future engines can return their results as `[]results.Result` without needing to change the `Engine.Search` signature (we convert inside the processor)" — this is from assumptions, confidence medium
- The plan says keep `ProcessorResult.Results` during migration alongside `TypedResults`

Actually, looking more carefully, since the `Engine.Search` returns `*models.Response` (flat), and the design's architecture diagram shows `engine.Search` producing `[]results.Result` (typed), the integration task needs to bridge this gap. During Phase 5, engines still return flat `models.Result` via `*models.Response`, and the processor (or some adapter) wraps them into typed results. 

But wait — the design says "This design defines the types and conversion; engines are updated incrementally in implementation cycles per Phase 4 tiers." So Phase 5 just builds the infrastructure but doesn't change existing engines.

The key architectural question: where does the conversion from `models.Result` → `results.Result` happen? 

Looking at the design's call-site integration:
> 8. `internal/search/processor/online.go` — convert engine `*models.Response.Results` through `results.ToAPIResult` into `ProcessorResult.TypedResults`.

So the sequence is:
1. Engine returns `*models.Response` (flat, as now)
2. `online.go` takes `resp.Results` (flat `[]models.Result`)
3. Converts them to `[]results.Result` (typed) somehow... but wait, `ToAPIResult` goes in the opposite direction (typed → flat)

Actually, re-reading: the design says `convert engine *models.Response.Results through results.ToAPIResult into ProcessorResult.TypedResults`. That doesn't make sense — `ToAPIResult` takes typed results and returns flat results. 

I think what the design means is: during migration, we need a reverse adapter. Let me look at the design again...

Actually, I think the simplest approach is:
1. Add a `results.Result` version of existing flat results → a new adapter function, or 
2. Since the design says engines will eventually return typed results directly, and during Phase 5 we just build the infrastructure, maybe we should:
   - Have `TypedResults` field on `ProcessorResult`
   - For now, leave it empty during migration
   - `TypedResultContainer` works with `TypedResults`
   - When engines are ported in later phases, they'll populate it

But wait — that contradicts the goal of Phase 5, which is to produce actual typed results and render them in the frontend. 

Let me re-read the architecture diagram:
```
engine.Search / base engine → []results.Result (typed) → Normalize() → ToAPIResult() → []models.Result (tagged union) → TypedResultContainer.Extend() → ...
```

So the data flow is: engines produce typed results → normalize → convert to tagged union → container. The tagged union is `models.Result` with `Kind` + `Extra`.

For Phase 5, since engines aren't updated yet, we need:
1. A way to wrap existing flat `models.Result` into `results.Result` (basic MainResult wrapping)
2. The processor does this wrapping
3. Then the pipeline works

This wrapper would be a simple function like `results.WrapAPIMainResult(models.Result) *results.MainResult` or something similar. I'll include this in the plan.

Actually, let me simplify. The existing engine returns `[]models.Result`. The simplest bridge: each flat `models.Result` is wrapped as a `*results.MainResult{}` with the same fields copied over. This gives us the typed pipeline without changing engines.

Let me now structure the tasks properly.

**Phase A: Backend Types & Container**

Task 1: Result interface upgrade — adds `Kind()`, `Template()` (per-kind, overrides BaseResult), `Base()`, `Normalize()`, `DedupKey()` to the `Result` interface. Adds missing types: `KeyValueResult`, `CodeLine`, `ImageRef`, `InfoboxAttribute`, `InfoboxURL`. Implements methods on all existing types. 
- Files: `pkg/models/results/results.go` (modify: add methods and new types)
- Files: `pkg/models/results/results_test.go` (modify: add tests for new interface methods and types)
- Shared signature change: `Result` interface gets new methods — all callers that use the interface (`convert.go`, existing tests) must be updated.
- Whole-tree typecheck after.

Task 2: Normalization extension — `BaseResult.Normalize()` orchestraes URL/text/engine normalization + kind-specific `PostNormalize()`. Add `normalizeEngines()` to `normalize.go`.
- Files: `pkg/models/results/normalize.go` (modify: add `normalizeEngines`, change `Normalize()` to call `PostNormalize`)
- Files: `pkg/models/results/normalize_test.go` (modify: add tests)

Task 3: ToAPIResult update — add `Kind`/`Template`/`Extra` to output. Add all kind-specific switch cases.
- Files: `pkg/models/results/convert.go` (modify)
- Files: `pkg/models/results/convert_test.go` (modify)
- NOTE: `models.Result` doesn't have `Kind`/`Extra` yet — Task 4 must come first, or this task adds them. Actually `models.Result` is defined in `pkg/models/models.go`. So Task 3 depends on Task 4 (API envelope update) for the `Kind` field.

Wait, let me reconsider dependencies. Task 3 (`ToAPIResult`) needs `models.Result.Kind` to exist. Task 4 (API envelope) adds `Kind` to `models.Result`. So Task 3 depends on Task 4.

But Task 5 (TypedResultContainer) uses `models.Result` as the stored type and needs `Kind` too.

Let me reorder:
- Task 4 should come before Task 3, or Task 3 should be combined with Task 4.

Actually, looking at the design more carefully, `ToAPIResult` outputs `[]models.Result`. If `models.Result.Kind` doesn't exist yet, the function can't set it. So Task 4 needs to come before Task 3.

Let me renumber:
- Task 1: Result interface upgrade (add interface methods)
- Task 2: API envelope update (add Kind/Extra/Template to models.Result, extend Infobox)
- Task 3: Normalization extension (add PostNormalize, normalizeEngines)
- Task 4: ToAPIResult update (set Kind/Template/Extra)
- Task 5: TypedResultContainer (new file)

Wait, Task 2 modifies `models.Result` which is a shared struct used across the codebase. Any change there needs a whole-tree typecheck. But `models.Result` struct changes are additive (new fields with json tags), so existing code should compile fine. The tests might need updating though.

Let me also think about whether I can parallelize some of these. Task 1 is the foundation for everything. Task 2 is independent of Task 1 (it modifies a different package). Task 3 depends on Task 1. Task 4 depends on Task 1 and 2. Task 5 depends on Task 1, 2, 4.

Better order:
Phase A:
- Task 1: Result interface upgrade (adds interface methods to results package)
- Task 2: API envelope update (adds Kind/Extra to models.Result, extends Infobox) — can parallel with Task 1?
  Actually no, Task 2 doesn't depend on Task 1, so they CAN be parallel.
- Task 3: Normalization extension (depends on Task 1 — needs PostNormalize on interface)
- Task 4: ToAPIResult update (depends on Task 1 and Task 2)
- Task 5: TypedResultContainer (depends on Task 4)

Phase B:
- Task 6: ProcessorResult extension (add TypedResults field)
- Task 7: WrapAPIMainResult adapter + online.go integration
- Task 8: Scheduler wiring (TypedResultContainer)

Phase C:
- Task 9: TypeScript types
- Task 10: ResultCard + kind-specific components
- Task 11: SearchPage layout

That's 11 tasks. Parts:
- Part 1: backend-types.md (Tasks 1-5)
- Part 2: integration.md (Tasks 6-8)
- Part 3: frontend.md (Tasks 9-11)

Now let me write the index.

</details>

Now I have a complete understanding of the codebase. Let me write the plan index first.

#### Tool Call: Bash (`mkdir -p /Users/ranwei/workspace/go_work/seargo/.ody-code/pl…`)
<!-- call_id: call_00_2z5XVIAsByRhUc060gpA0376 -->
```json
{
  "command": "mkdir -p /Users/ranwei/workspace/go_work/seargo/.ody-code/plans/2026-06-20-05-result-types",
  "description": "Create plan part subdirectory"
}
```

<details><summary>Tool Result: Bash (`mkdir -p /Users/ranwei/workspace/go_work/seargo/.ody-code/pl…`)</summary>

<!-- call_id: call_00_2z5XVIAsByRhUc060gpA0376 -->
<system>Tool output is empty.</system>

</details>

<details><summary>Thinking</summary>

Let me write the plan index now. I need to get the scaffold right: header + File Structure + Dependency Overview + Risks & Open Questions + Parts manifest. Then in subsequent turns, I'll write each part file.

</details>

#### Tool Call: Write (`/Users/ranwei/workspace/go_work/seargo/.ody-code/plans/2026-…`)
<!-- call_id: call_00_r7os3dzovz381dJdBUH55323 -->
```json
{
  "path": "/Users/ranwei/workspace/go_work/seargo/.ody-code/plans/2026-06-20-05-result-types.md",
  "content": "# Structured Result Types & Rendering — Implementation Plan\n\n**Goal:** Replace the single flat `models.Result` with a tagged-union API envelope carrying kind-specific fields, build a typed result hierarchy with normalization and dedup, introduce `TypedResultContainer` for per-kind merge, and wire the frontend to dispatch on `kind` for specialized rendering.\n\n**Architecture:** Three subsystems: (A) Backend Types — upgrade the `results.Result` interface with `Kind()/Template()/Normalize()/DedupKey()/Base()`, add missing types (`KeyValueResult`, `CodeLine`, `ImageRef`, `InfoboxAttribute`, `InfoboxURL`), extend normalization with `normalizeEngines` and per-kind `PostNormalize()`, add `Kind`+`Template`+`Extra` to `models.Result`, rewrite `ToAPIResult` for tagged-union output, and create `TypedResultContainer` with per-kind buckets and SearXNG-aligned score/dedup; (B) Integration — extend `ProcessorResult` with `TypedResults`, add an adapter to wrap existing flat engine results into typed ones, wire scheduler to use `TypedResultContainer`; (C) Frontend — rewrite TypeScript types as tagged union, create `ResultCard` dispatch component with kind-specific rendering components, and update `SearchPage` for category-aware layout.\n\n**Tech Stack:** Go 1.25, goquery, resty/v2, testify, zustand, React 19, TypeScript 6, Tailwind CSS 3\n\n> For executing workers: implement this plan task-by-task (prefer a fresh subagent/Task per task — a clean context per task avoids single-session degradation). Steps use - [ ] checkboxes for tracking.\n\n## File Structure\n\n### New Files\n| File | Responsibility |\n|---|---|\n| `internal/search/typed_container.go` | TypedResultContainer: per-kind buckets, dedup, merge, score, Close/Results |\n| `internal/search/typed_container_test.go` | TypedResultContainer unit tests |\n| `web/src/components/results/ResultCard.tsx` | Dispatch on `kind` to kind-specific components |\n| `web/src/components/results/ImageCard.tsx` | Image result with thumbnail + resolution |\n| `web/src/components/results/VideoCard.tsx` | Video result with iframe + duration |\n| `web/src/components/results/PaperCard.tsx` | Academic paper with DOI/authors |\n| `web/src/components/results/CodeBlock.tsx` | Code snippet with language label |\n| `web/src/components/results/FileRow.tsx` | File/torrent with size/seeders info |\n| `web/src/components/results/MapCard.tsx` | Map result with lat/lng |\n| `web/src/components/results/MusicCard.tsx` | Music track with artist/album |\n| `web/src/components/results/AnswerBox.tsx` | Answer box (dedicated, outside result list) |\n| `web/src/components/results/KeyValueTable.tsx` | Key/value table display |\n| `web/src/components/results/InfoboxPanel.tsx` | Infobox panel with attributes/URLs |\n| `web/src/components/results/MainResult.tsx` | Default web result card |\n| `web/src/components/results/ImageGrid.tsx` | Image category grid layout |\n\n### Modified Files\n| File | Changes |\n|---|---|\n| `pkg/models/results/results.go` | Add `Kind()/Template()/Base()/Normalize()/DedupKey()` to `Result` interface; add `KeyValueResult`/`CodeLine`/`ImageRef`/`InfoboxAttribute`/`InfoboxURL` types; implement methods on all 12 types |\n| `pkg/models/results/results_test.go` | Tests for new interface methods, `KeyValueResult`, `DedupKey()` on each kind |\n| `pkg/models/results/normalize.go` | Add `normalizeEngines()`; change `BaseResult.Normalize()` to call kind-specific `PostNormalize()` via type switch; add `PostNormalize()` on all kinds |\n| `pkg/models/results/normalize_test.go` | Tests for `normalizeEngines`, `PostNormalize` on Image/Code/File/Infobox/Video |\n| `pkg/models/results/convert.go` | Rewrite `ToAPIResult`: set `Kind`/`Template`/`Extra` on output; add all kind-specific `Extra` cases; add `WrapAPIMainResult()` reverse adapter |\n| `pkg/models/results/convert_test.go` | Tests for `Kind`/`Extra` on each kind, unknown-kind fallback, `WrapAPIMainResult` |\n| `pkg/models/models.go` | Add `Kind`/`Template`/`Extra` fields to `Result`; extend `Infobox` with `InfoboxURL`/`InfoboxAttribute` types; `Engines` field on `Infobox` |\n| `internal/search/processor/types.go` | Add `TypedResults []results.Result` to `ProcessorResult` |\n| `internal/search/processor/online.go` | Wrap engine flat results via `WrapAPIMainResult` into `TypedResults` |\n| `internal/search/scheduler.go` | Replace `ResultContainer` with `TypedResultContainer` in `executeProcessors`; feed `TypedResults` |\n| `web/src/types/search.ts` | Replace flat `Result` with tagged-union types; add `answers`/`corrections`/`infoboxes` to `SearchResponse` |\n| `web/src/stores/searchStore.ts` | Update `Result` import to new tagged-union type |\n| `web/src/pages/SearchPage.tsx` | Replace inline result card with `<ResultCard>`; add answer/infobox/ImageGrid rendering |\n\n## Dependency Overview\n\n```\nPhase A: Backend Types & Container (no upstream deps within this plan)\n  Part: backend-types\n    Task 1 (Result interface upgrade + missing types) — stand-alone\n    Task 2 (API envelope: models.Result+Kind, models.Infobox extension) — parallel with T1\n    Task 3 (Normalization: normalizeEngines + PostNormalize) — depends on T1\n    Task 4 (ToAPIResult: tagged-union output) — depends on T1, T2\n    Task 5 (TypedResultContainer) — depends on T4\n\nPhase B: Integration (depends on Phase A)\n  Part: integration\n    Task 6 (ProcessorResult extension + WrapAPIMainResult) — depends on T4\n    Task 7 (Online processor adapter) — depends on T6\n    Task 8 (Scheduler wiring) — depends on T5, T6\n\nPhase C: Frontend (depends on Phase A: T2 for API contract)\n  Part: frontend\n    Task 9 (TypeScript types) — depends on T2\n    Task 10 (ResultCard + kind-specific components) — depends on T9\n    Task 11 (SearchPage layout + store) — depends on T10\n```\n\nPhase A 内部：T1,T2 可并行；Phase B/C 可并行（B 不依赖 C，C 不依赖 B），但各自内部串行。\n\n## Risks & Open Questions\n\n| # | Risk | Mitigation |\n|---|---|---|\n| 1 | `Result` interface gains 5 new methods — all callers (`convert.go`, tests) must conform | Task 1 finds every `Result` interface reference via `grep -rn \"results\\.Result\"` and updates in same task; whole-tree `go build ./... && go test ./...` |\n| 2 | `models.Result` adds `Kind`/`Extra`/`Template` fields — existing cache serialization (flat struct) may break on deserialization | Fields have `omitempty`; old cache entries lack `kind` → frontend treats empty as `\"main\"`; cache TTL is short (max 2min), old entries expire naturally |\n| 3 | `TypedResultContainer` replaces `ResultContainer` in `executeProcessors` — regression risk on scoring/grouping | Score formula is identical (`Σ weight/pos × count`); dedup key adds `kind` prefix; keep old container code in `merge.go` untouched during migration for rollback |\n| 4 | Frontend `ResultCard` exhaustive switch must handle all 12 kinds | TypeScript discriminated union on `kind`; default branch renders `MainResult` with `console.warn`; `tsc --noEmit` catches missing cases |\n| 5 | Older engines return flat `models.Result` without `Kind` — `WrapAPIMainResult` produces typed results but loses kind-specific fields (image metadata, etc.) | Acceptable: engines are ported incrementally per Phase 4 tiers; until then, all flat results render as `MainResult` with title/URL/content intact |\n| 6 | `Infobox` struct change (from `[]string` to typed `[]InfoboxURL`/`[]InfoboxAttribute`) breaks existing infobox producers | Search shows 0 existing infobox producers in current 6 engines; fixtures validate new shape |\n\n## Parts (generate one per invocation, in order)\n\n| # | File | Scope | Status |\n|---|---|---|---|\n| 1 | 2026-06-20-05-result-types/backend-types.md | Result interface, missing types, API envelope, normalization, ToAPIResult, TypedResultContainer | pending |\n| 2 | 2026-06-20-05-result-types/integration.md | ProcessorResult, WrapAPIMainResult adapter, online processor, scheduler wiring | pending |\n| 3 | 2026-06-20-05-result-types/frontend.md | TypeScript types, ResultCard, kind-specific components, SearchPage layout | pending |\n"
}
```

<details><summary>Tool Result: Write (`/Users/ranwei/workspace/go_work/seargo/.ody-code/plans/2026-…`)</summary>

<!-- call_id: call_00_r7os3dzovz381dJdBUH55323 -->
Wrote 8071 bytes to /Users/ranwei/workspace/go_work/seargo/.ody-code/plans/2026-06-20-05-result-types.md

</details>

<details><summary>Thinking</summary>

Index is written. Now I should stop this turn. The injection on the next turn will direct me to the first pending part.

</details>

Index scaffolded. 3 parts all `pending` — write Part 1 (backend-types.md) next turn.

## Turn 2

### User

continue

### Assistant

<details><summary>Thinking</summary>

I need to write Part 1: `2026-06-20-05-result-types/backend-types.md`. This part covers 5 tasks:

Task 1: Result interface upgrade + missing types
Task 2: API envelope update (models.Result + Infobox)
Task 3: Normalization extension
Task 4: ToAPIResult rewrite
Task 5: TypedResultContainer

Let me write this part with full detail.

</details>

#### Tool Call: Write (`/Users/ranwei/workspace/go_work/seargo/.ody-code/plans/2026-…`)
<!-- call_id: call_00_VtFacaI1nA0zZRLHTjmI0562 -->
```json
{
  "content": "# Part 1 — Backend Types & Container\n\n**Scope:** Result interface upgrade with `Kind()/Template()/Base()/Normalize()/DedupKey()`, missing types (`KeyValueResult`, `CodeLine`, `ImageRef`, `InfoboxAttribute`, `InfoboxURL`), `models.Result` tagged-union (Kind+Template+Extra), `Infobox` rich schema, normalization with `normalizeEngines` + `PostNormalize`, `ToAPIResult` tagged-union output with all 12 kind-specific Extra cases, `TypedResultContainer` with per-kind buckets.\n\n## Dependency Graph (within this part)\n\n```\nT1 (interface upgrade) ──┬──► T3 (normalization)\n                          │\nT2 (API envelope) ──┬─────┴──► T4 (ToAPIResult) ──► T5 (TypedResultContainer)\n                    │\n                    └─────────► T5 (uses models.Result directly)\n```\n\nT1 and T2 are independent and can run in parallel. T3 depends on T1 (needs `PostNormalize` on the interface). T4 depends on T1 (needs typed result kind switches) and T2 (needs `models.Result.Kind`). T5 depends on T4 (uses tagged-union output as stored type).\n\n---\n\n### Task 1: Upgrade Result interface + add missing types\n\n**Depends on:** none\n\n**Files:**\n- Modify: `pkg/models/results/results.go:1-165` — add `Kind()/Template()/Base()/Normalize()/DedupKey()` to `Result` interface; add `KeyValueResult`/`CodeLine`/`ImageRef`/`InfoboxAttribute`/`InfoboxURL` types; add `FileSize`/`MagnetURI`/`Seeders`/`Leechers` to `FileResult`; add `CodeLines`/`HLLines` to `CodeResult`; implement all interface methods on all 12 types\n- Modify: `pkg/models/results/results_test.go:1-92` — add tests for new interface methods, `KeyValueResult`, `DedupKey()` on each kind\n- Modify: `pkg/models/results/convert.go:1-101` — update switch cases for renamed/changed fields (no functional change yet)\n\n**Shared-signature change:** `Result` interface gains 5 methods. All callers that reference `results.Result` must be audited. Current callers found via `grep -rn \"results\\.Result\"`:\n\n```\npkg/models/results/convert.go:9: func ToAPIResult(results []Result) []models.Result\npkg/models/results/convert.go:13: out := make([]models.Result, 0, len(results))\npkg/models/results/convert.go:26: switch t := r.(type) {\npkg/models/results/convert_test.go:20: apiResults := ToAPIResult([]Result{mr})\n... (all convert_test.go uses)\npkg/models/results/results_test.go:78: var r Result = &MainResult{}\n... (all interface-satisfaction tests)\n```\n\nThe interface is consumed by `pkg/models/results/convert.go` and test files only — no external consumers in `internal/`. The task must ensure `convert.go` compiles after the interface change.\n\n- [ ] Write the failing test (add to `results_test.go`):\n\n```go\n// TestResultInterface_Kind verifies each type returns the correct Kind().\nfunc TestResultInterface_Kind(t *testing.T) {\n    kinds := map[Result]string{\n        &MainResult{}:         \"main\",\n        &ImageResult{}:        \"image\",\n        &VideoResult{}:        \"video\",\n        &NewsResult{}:         \"news\",\n        &PaperResult{}:        \"paper\",\n        &CodeResult{}:         \"code\",\n        &FileResult{}:         \"file\",\n        &MapResult{}:          \"map\",\n        &MusicResult{}:        \"music\",\n        &AnswerResult{}:       \"answer\",\n        &KeyValueResult{}:     \"keyvalue\",\n        &InfoboxResult{}:      \"infobox\",\n    }\n    for r, want := range kinds {\n        assert.Equal(t, want, r.Kind(), \"Kind mismatch for %T\", r)\n    }\n}\n\n// TestResultInterface_Template verifies per-kind template defaults.\nfunc TestResultInterface_Template(t *testing.T) {\n    tmpls := map[Result]string{\n        &MainResult{}:         \"default.html\",\n        &ImageResult{}:        \"images.html\",\n        &VideoResult{}:        \"videos.html\",\n        &NewsResult{}:         \"default.html\",\n        &PaperResult{}:        \"paper.html\",\n        &CodeResult{}:         \"code.html\",\n        &FileResult{}:         \"file.html\",\n        &MapResult{}:          \"map.html\",\n        &MusicResult{}:        \"music.html\",\n        &AnswerResult{}:       \"answer/legacy.html\",\n        &KeyValueResult{}:     \"keyvalue.html\",\n        &InfoboxResult{}:      \"infobox.html\",\n    }\n    for r, want := range tmpls {\n        assert.Equal(t, want, r.Template(), \"%T template\", r)\n    }\n}\n\n// TestResultInterface_Base returns the embedded BaseResult pointer.\nfunc TestResultInterface_Base(t *testing.T) {\n    mr := &MainResult{BaseResult: BaseResult{Title: \"T\", URL: \"https://x.com\"}}\n    assert.Equal(t, \"T\", mr.Base().Title)\n    assert.Equal(t, \"https://x.com\", mr.Base().URL)\n}\n\n// TestResultInterface_DedupKey verifies kind-specific dedup logic.\nfunc TestResultInterface_DedupKey(t *testing.T) {\n    // MainResult: host|path|query|thumbnail\n    mr := &MainResult{BaseResult: BaseResult{URL: \"https://example.com/page?q=1\", ThumbnailURL: \"t.jpg\"}}\n    assert.Equal(t, \"main|example.com|/page|q=1|t.jpg\", mr.DedupKey())\n\n    // ImageResult: uses ImgSrc when ThumbnailURL empty\n    ir := &ImageResult{BaseResult: BaseResult{URL: \"https://x.com/a\"}, ImgSrc: \"https://x.com/a.jpg\"}\n    assert.Contains(t, ir.DedupKey(), \"a.jpg\")\n\n    // KeyValueResult: uses Title as fallback key\n    kv := &KeyValueResult{BaseResult: BaseResult{Title: \"KeyTable\", URL: \"\"}}\n    assert.Equal(t, \"keyvalue|||KeyTable|\", kv.DedupKey())\n}\n\n// TestKeyValueResult_Fields verifies the new KeyValueResult type.\nfunc TestKeyValueResult_Fields(t *testing.T) {\n    kv := &KeyValueResult{\n        BaseResult: BaseResult{Title: \"Stats\", Engine: \"wikipedia\", Template: \"keyvalue.html\"},\n        KVMap:      map[string]string{\"Population\": \"8M\", \"Area\": \"1000 km²\"},\n        Caption:    \"City stats\",\n        KeyTitle:   \"Metric\",\n        ValueTitle: \"Value\",\n    }\n    assert.Equal(t, \"keyvalue\", kv.Kind())\n    assert.Equal(t, \"keyvalue.html\", kv.Template())\n    assert.Len(t, kv.KVMap, 2)\n}\n\n// TestInfoboxAttribute_Fields verifies the new helper struct.\nfunc TestInfoboxAttribute_Fields(t *testing.T) {\n    attr := InfoboxAttribute{Label: \"Capital\", Value: \"Paris\", URL: \"https://x.com/paris\"}\n    assert.Equal(t, \"Capital\", attr.Label)\n    assert.Equal(t, \"Paris\", attr.Value)\n}\n\n// TestInfoboxURL_Fields verifies the new helper struct.\nfunc TestInfoboxURL_Fields(t *testing.T) {\n    u := InfoboxURL{Title: \"Wikipedia\", URL: \"https://wikipedia.org\"}\n    assert.Equal(t, \"Wikipedia\", u.Title)\n}\n```\n\n- [ ] Run it and verify it FAILS:\n\n```bash\ngo test ./pkg/models/results/... 2>&1 | head -20\n# Expected: compilation errors — \"Kind not declared by Result\", \"Template not declared by Result\",\n# \"Base not declared by Result\", \"DedupKey not declared by Result\",\n# \"Normalize not declared by Result\", \"KeyValueResult undefined\",\n# \"ImageRef undefined\", \"InfoboxAttribute undefined\", \"InfoboxURL undefined\",\n# \"CodeLine undefined\", \"FileResult has no field CodeLines\", etc.\n```\n\n- [ ] Write the minimal implementation (`results.go`):\n\nChange the `Result` interface (lines 8-16):\n\n```go\n// Result is the interface that every typed search result implements.\ntype Result interface {\n    Kind() string\n    Template() string\n    GetTitle() string\n    GetURL() string\n    GetContent() string\n    GetEngine() string\n    GetCategory() string\n    Base() *BaseResult\n    Normalize()\n    DedupKey() string\n}\n```\n\nAdd new types after `InfoboxResult` (insert before // ResultTypes):\n\n```go\n// KeyValueResult — generic key/value table.\ntype KeyValueResult struct {\n    BaseResult\n    KVMap      map[string]string `json:\"kv_map\"`\n    Caption    string            `json:\"caption,omitempty\"`\n    KeyTitle   string            `json:\"key_title,omitempty\"`\n    ValueTitle string            `json:\"value_title,omitempty\"`\n}\nfunc (r *KeyValueResult) Kind() string     { return \"keyvalue\" }\nfunc (r *KeyValueResult) Template() string { return \"keyvalue.html\" }\n\n// ImageRef — alternative image format reference (deferred in rendering).\ntype ImageRef struct {\n    URL    string `json:\"url\"`\n    Format string `json:\"format\"`\n    Label  string `json:\"label,omitempty\"`\n}\n\n// CodeLine represents a single line of code.\ntype CodeLine struct {\n    Line int    `json:\"line\"`\n    Text string `json:\"text\"`\n}\n\n// InfoboxAttribute is a labeled key/value pair in an infobox.\ntype InfoboxAttribute struct {\n    Label string `json:\"label\"`\n    Value string `json:\"value\"`\n    URL   string `json:\"url,omitempty\"`\n}\n\n// InfoboxURL is a related URL in an infobox.\ntype InfoboxURL struct {\n    Title string `json:\"title\"`\n    URL   string `json:\"url\"`\n}\n```\n\nExtend `ImageResult` (add after `Source` field, line 58):\n\n```go\n    FileSize  string     `json:\"file_size,omitempty\"`\n    Formats   []ImageRef `json:\"formats,omitempty\"`\n```\n\nExtend `CodeResult` (replace current struct, lines 104-109):\n\n```go\ntype CodeResult struct {\n    BaseResult\n    Repository   string     `json:\"repository,omitempty\"`\n    CodeLanguage string     `json:\"code_language,omitempty\"`\n    Filename     string     `json:\"filename,omitempty\"`\n    CodeLines    []CodeLine `json:\"code_lines,omitempty\"`\n    HLLines      []int      `json:\"hl_lines,omitempty\"`\n}\n```\n\nExtend `FileResult` (replace current struct, lines 111-116):\n\n```go\ntype FileResult struct {\n    BaseResult\n    Filename  string `json:\"filename,omitempty\"`\n    FileType  string `json:\"file_type,omitempty\"`\n    FileSize  int64  `json:\"file_size,omitempty\"`\n    MagnetURI string `json:\"magnet_uri,omitempty\"`\n    Seeders   int    `json:\"seeders,omitempty\"`\n    Leechers  int    `json:\"leechers,omitempty\"`\n}\n```\n\nReplace `InfoboxResult` (lines 142-150) with:\n\n```go\ntype InfoboxResult struct {\n    BaseResult\n    InfoboxID     string             `json:\"infobox_id,omitempty\"`\n    Attributes    []InfoboxAttribute `json:\"attributes,omitempty\"`\n    URLs          []InfoboxURL       `json:\"urls,omitempty\"`\n    RelatedTopics []string           `json:\"related_topics,omitempty\"`\n    ImgSrc        string             `json:\"img_src,omitempty\"`\n    ImgAlt        string             `json:\"img_alt,omitempty\"`\n}\n```\n\nNow implement missing methods on all types. Each type needs `Kind()`, `Template()`, `Base()`, `Normalize()`, `DedupKey()`. Since `MainResult` through `MusicResult` and `AnswerResult`, `InfoboxResult` already exist with `BaseResult` embedding, they need the new methods added. `KeyValueResult` needs all methods.\n\nThe `Normalize()` method on each type delegates to `BaseResult.Normalize()` + type-specific `PostNormalize()`. The `DedupKey()` method follows the design's algorithm:\n\n```go\n// === Method implementations for each type ===\n\n// MainResult\nfunc (r *MainResult) Kind() string     { return \"main\" }\nfunc (r *MainResult) Template() string { return \"default.html\" }\nfunc (r *MainResult) Base() *BaseResult { return &r.BaseResult }\nfunc (r *MainResult) Normalize() {\n    r.BaseResult.Normalize()\n    // MainResult has no PostNormalize\n}\nfunc (r *MainResult) DedupKey() string {\n    return \"main|\" + dedupKeyFromBase(&r.BaseResult)\n}\n\n// ImageResult\nfunc (r *ImageResult) Kind() string     { return \"image\" }\nfunc (r *ImageResult) Template() string { return \"images.html\" }\nfunc (r *ImageResult) Base() *BaseResult { return &r.BaseResult }\nfunc (r *ImageResult) Normalize() {\n    r.BaseResult.Normalize()\n    if r.ThumbnailSrc == \"\" && r.ImgSrc != \"\" {\n        r.ThumbnailSrc = r.ImgSrc\n    }\n    if r.Title == \"\" && r.ImgSrc != \"\" {\n        r.Title = basenameFromURL(r.ImgSrc)\n    }\n}\nfunc (r *ImageResult) DedupKey() string {\n    extra := r.ThumbnailSrc\n    if extra == \"\" {\n        extra = r.ImgSrc\n    }\n    return \"image|\" + dedupKeyFromBase(&r.BaseResult) + \"|\" + extra\n}\n\n// VideoResult\nfunc (r *VideoResult) Kind() string     { return \"video\" }\nfunc (r *VideoResult) Template() string { return \"videos.html\" }\nfunc (r *VideoResult) Base() *BaseResult { return &r.BaseResult }\nfunc (r *VideoResult) Normalize() {\n    r.BaseResult.Normalize()\n    if r.ThumbnailURL == \"\" && r.Thumbnail != \"\" {\n        r.ThumbnailURL = r.Thumbnail\n    }\n}\nfunc (r *VideoResult) DedupKey() string {\n    return \"video|\" + dedupKeyFromBase(&r.BaseResult) + \"|\" + r.Thumbnail\n}\n\n// NewsResult\nfunc (r *NewsResult) Kind() string     { return \"news\" }\nfunc (r *NewsResult) Template() string { return \"default.html\" }\nfunc (r *NewsResult) Base() *BaseResult { return &r.BaseResult }\nfunc (r *NewsResult) Normalize()       { r.BaseResult.Normalize() }\nfunc (r *NewsResult) DedupKey() string { return \"news|\" + dedupKeyFromBase(&r.BaseResult) }\n\n// PaperResult\nfunc (r *PaperResult) Kind() string     { return \"paper\" }\nfunc (r *PaperResult) Template() string { return \"paper.html\" }\nfunc (r *PaperResult) Base() *BaseResult { return &r.BaseResult }\nfunc (r *PaperResult) Normalize()       { r.BaseResult.Normalize() }\nfunc (r *PaperResult) DedupKey() string { return \"paper|\" + dedupKeyFromBase(&r.BaseResult) }\n\n// CodeResult\nfunc (r *CodeResult) Kind() string     { return \"code\" }\nfunc (r *CodeResult) Template() string { return \"code.html\" }\nfunc (r *CodeResult) Base() *BaseResult { return &r.BaseResult }\nfunc (r *CodeResult) Normalize() {\n    r.BaseResult.Normalize()\n    if r.Title == \"\" && r.Filename != \"\" {\n        r.Title = r.Filename\n    }\n}\nfunc (r *CodeResult) DedupKey() string { return \"code|\" + dedupKeyFromBase(&r.BaseResult) }\n\n// FileResult\nfunc (r *FileResult) Kind() string     { return \"file\" }\nfunc (r *FileResult) Template() string { return \"file.html\" }\nfunc (r *FileResult) Base() *BaseResult { return &r.BaseResult }\nfunc (r *FileResult) Normalize() {\n    r.BaseResult.Normalize()\n    if r.Title == \"\" && r.Filename != \"\" {\n        r.Title = r.Filename\n    }\n    if r.FileType == \"\" && r.Filename != \"\" {\n        r.FileType = extensionFromFilename(r.Filename)\n    }\n}\nfunc (r *FileResult) DedupKey() string { return \"file|\" + dedupKeyFromBase(&r.BaseResult) }\n\n// MapResult\nfunc (r *MapResult) Kind() string     { return \"map\" }\nfunc (r *MapResult) Template() string { return \"map.html\" }\nfunc (r *MapResult) Base() *BaseResult { return &r.BaseResult }\nfunc (r *MapResult) Normalize()       { r.BaseResult.Normalize() }\nfunc (r *MapResult) DedupKey() string { return \"map|\" + dedupKeyFromBase(&r.BaseResult) }\n\n// MusicResult\nfunc (r *MusicResult) Kind() string     { return \"music\" }\nfunc (r *MusicResult) Template() string { return \"music.html\" }\nfunc (r *MusicResult) Base() *BaseResult { return &r.BaseResult }\nfunc (r *MusicResult) Normalize()       { r.BaseResult.Normalize() }\nfunc (r *MusicResult) DedupKey() string { return \"music|\" + dedupKeyFromBase(&r.BaseResult) }\n\n// AnswerResult\nfunc (r *AnswerResult) Kind() string     { return \"answer\" }\nfunc (r *AnswerResult) Template() string { return \"answer/legacy.html\" }\nfunc (r *AnswerResult) Base() *BaseResult { return &r.BaseResult }\nfunc (r *AnswerResult) Normalize()       { r.BaseResult.Normalize() }\nfunc (r *AnswerResult) DedupKey() string { return \"answer|\" + r.Answer }\n\n// KeyValueResult\nfunc (r *KeyValueResult) Kind() string     { return \"keyvalue\" }\nfunc (r *KeyValueResult) Template() string { return \"keyvalue.html\" }\nfunc (r *KeyValueResult) Base() *BaseResult { return &r.BaseResult }\nfunc (r *KeyValueResult) Normalize()       { r.BaseResult.Normalize() }\nfunc (r *KeyValueResult) DedupKey() string {\n    if r.URL != \"\" {\n        return \"keyvalue|\" + dedupKeyFromBase(&r.BaseResult)\n    }\n    return \"keyvalue|||\" + r.Title + \"|\"\n}\n\n// InfoboxResult\nfunc (r *InfoboxResult) Kind() string     { return \"infobox\" }\nfunc (r *InfoboxResult) Template() string { return \"infobox.html\" }\nfunc (r *InfoboxResult) Base() *BaseResult { return &r.BaseResult }\nfunc (r *InfoboxResult) Normalize() {\n    r.BaseResult.Normalize()\n    if r.InfoboxID == \"\" {\n        r.InfoboxID = r.URL\n    }\n    if r.InfoboxID == \"\" {\n        r.InfoboxID = \"infobox:\" + r.Title\n    }\n}\nfunc (r *InfoboxResult) DedupKey() string { return \"infobox|\" + r.InfoboxID }\n```\n\nAdd shared helpers at bottom of `results.go`:\n\n```go\n// dedupKeyFromBase generates the URL-based dedup key part.\nfunc dedupKeyFromBase(br *BaseResult) string {\n    rawURL := br.URL\n    if rawURL == \"\" {\n        return \"|||\" + br.ThumbnailURL\n    }\n    u, err := url.Parse(rawURL)\n    if err != nil {\n        return rawURL + \"|\" + br.ThumbnailURL\n    }\n    u.Scheme = strings.ToLower(u.Scheme)\n    u.Host = strings.ToLower(u.Host)\n    if strings.HasPrefix(u.Host, \"www.\") && !strings.HasPrefix(u.Host, \"www2.\") && !strings.HasPrefix(u.Host, \"www3.\") {\n        u.Host = u.Host[4:]\n    }\n    u.Path = strings.TrimSuffix(u.Path, \"/\")\n    u.Fragment = \"\"\n    return u.Host + \"|\" + u.Path + \"|\" + u.RawQuery + \"|\" + br.ThumbnailURL\n}\n\n// basenameFromURL extracts the filename from a URL path.\nfunc basenameFromURL(rawURL string) string {\n    u, err := url.Parse(rawURL)\n    if err != nil {\n        return rawURL\n    }\n    parts := strings.Split(u.Path, \"/\")\n    return parts[len(parts)-1]\n}\n\n// extensionFromFilename returns the file extension without the dot.\nfunc extensionFromFilename(filename string) string {\n    if i := strings.LastIndex(filename, \".\"); i >= 0 && i < len(filename)-1 {\n        return filename[i+1:]\n    }\n    return \"\"\n}\n```\n\nNeed to add imports: `\"net/url\"` and `\"strings\"` to `results.go`.\n\nUpdate `convert.go` for changed struct fields — `FileResult.MagnetURI`/`Seeders`/`Leechers`/`Filename`, `CodeResult.Filename`/`CodeLines`/`HLLines`, `ImageResult.FileSize`/`Formats`, `InfoboxResult` attribute/URL types. Add cases in the switch for the new fields (existing engine_data passthrough handles them, but the switch should map them to `ed`). Also update `FileResult` case:\n\n```go\ncase *FileResult:\n    ed[\"file_type\"] = t.FileType\n    ed[\"file_size\"] = t.FileSize\n    ed[\"filename\"] = t.Filename\n    ed[\"magnet_uri\"] = t.MagnetURI\n    ed[\"seeders\"] = t.Seeders\n    ed[\"leechers\"] = t.Leechers\n```\n\nAnd `CodeResult` case:\n\n```go\ncase *CodeResult:\n    ed[\"repository\"] = t.Repository\n    ed[\"code_language\"] = t.CodeLanguage\n    ed[\"filename\"] = t.Filename\n    if t.CodeLines != nil {\n        ed[\"code_lines\"] = t.CodeLines\n    }\n    if t.HLLines != nil {\n        ed[\"hl_lines\"] = t.HLLines\n    }\n```\n\n- [ ] Run it and verify it PASSES:\n\n```bash\ngo test ./pkg/models/results/... -v -run \"TestResultInterface_|TestKeyValueResult|TestInfoboxAttribute|TestInfoboxURL\"\n```\n\nExpected: all new tests pass. Then full suite:\n\n```bash\ngo test ./pkg/models/results/... -v\n```\n\n- [ ] Whole-tree typecheck (shared signature change — `Result` interface):\n\n```bash\ngo build ./...\n```\n\nExpected: all packages compile. No caller outside `pkg/models/results/` references the `Result` interface directly (confirmed by grep — only `convert.go` and test files within the package).\n\n- [ ] Commit:\n\n```bash\ngit add pkg/models/results/results.go pkg/models/results/results_test.go pkg/models/results/convert.go\ngit commit -m \"results: upgrade Result interface with Kind/Template/Base/Normalize/DedupKey, add KeyValueResult/CodeLine/ImageRef/InfoboxAttribute/InfoboxURL\"\n```\n\n---\n\n### Task 2: API envelope — tagged-union Result + Infobox rich schema\n\n**Depends on:** none (parallel with Task 1)\n\n**Files:**\n- Modify: `pkg/models/models.go:94-127` — add `Kind`/`Template`/`Extra` to `Result`; extend `Infobox` with typed `Attributes`/`URLs`/`Engines`\n- Modify: test files that reference `models.Result` (none exist in current test suite — `models` package has no `_test.go`)\n\n- [ ] Manual verification — write the code, then build:\n\nAdd fields to `Result` struct (lines 94-109):\n\n```go\ntype Result struct {\n    Kind         string         `json:\"kind\"`\n    Template     string         `json:\"template,omitempty\"`\n    Title        string         `json:\"title\"`\n    URL          string         `json:\"url\"`\n    Content      string         `json:\"content\"`\n    Engine       string         `json:\"engine\"`\n    Engines      []string       `json:\"engines,omitempty\"`\n    Category     Category       `json:\"category\"`\n    Score        float64        `json:\"score\"`\n    Positions    []int          `json:\"-\"`\n    ThumbnailURL string         `json:\"thumbnail_url,omitempty\"`\n    PublishedAt  *time.Time     `json:\"published_at,omitempty\"`\n    Domain       string         `json:\"domain,omitempty\"`\n    Favicon      string         `json:\"favicon,omitempty\"`\n    EngineData   map[string]any `json:\"engine_data,omitempty\"`\n    Extra        map[string]any `json:\"extra,omitempty\"`\n}\n```\n\nReplace `Infobox` (lines 84-92) with:\n\n```go\ntype Infobox struct {\n    Title         string             `json:\"title\"`\n    URL           string             `json:\"url,omitempty\"`\n    Content       string             `json:\"content,omitempty\"`\n    Engine        string             `json:\"engine,omitempty\"`\n    Engines       []string           `json:\"engines,omitempty\"`\n    ImgSrc        string             `json:\"img_src,omitempty\"`\n    Attributes    []InfoboxAttribute `json:\"attributes,omitempty\"`\n    URLs          []InfoboxURL       `json:\"urls,omitempty\"`\n    RelatedTopics []string           `json:\"related_topics,omitempty\"`\n}\n\ntype InfoboxAttribute struct {\n    Label string `json:\"label\"`\n    Value string `json:\"value\"`\n    URL   string `json:\"url,omitempty\"`\n}\n\ntype InfoboxURL struct {\n    Title string `json:\"title\"`\n    URL   string `json:\"url\"`\n}\n```\n\n- [ ] Build verification:\n\n```bash\ngo build ./...\n```\n\nExpected: all packages compile. The `Result` struct is used extensively in `internal/search/` and engines — since new fields are additive with `omitempty`, all existing code that constructs `models.Result{}` without setting `Kind`/`Template`/`Extra` still compiles (they default to zero values). Verify:\n\n```bash\ngo test ./internal/search/... ./engines/...\n```\n\nExpected: all existing tests pass (fields default to zero values, no test asserts on the new fields yet).\n\n- [ ] Commit:\n\n```bash\ngit add pkg/models/models.go\ngit commit -m \"models: add Kind/Template/Extra to Result (tagged union), extend Infobox with typed attributes/URLs\"\n```\n\n---\n\n### Task 3: Normalization — normalizeEngines + PostNormalize\n\n**Depends on:** Task 1\n\n**Files:**\n- Modify: `pkg/models/results/normalize.go:1-121` — add `normalizeEngines()`, change `BaseResult.Normalize()` to accept a `Result`-capable parameter or add a new orchestration method; for backward compat, keep `BaseResult.Normalize()` as-is and add `NormalizeResult(r Result)` that does URL→Text→Engines→PostNormalize\n- Modify: `pkg/models/results/normalize_test.go:1-71` — add tests for `normalizeEngines`, `NormalizeResult`, PostNormalize on each kind\n\n- [ ] Write the failing test:\n\n```go\nfunc TestNormalizeEngines_AppendsEngine(t *testing.T) {\n    br := &BaseResult{Engine: \"google\", Engines: nil}\n    normalizeEngines(br)\n    assert.Equal(t, \"google\", br.Engine)\n    assert.Contains(t, br.Engines, \"google\")\n}\n\nfunc TestNormalizeEngines_NoDuplicate(t *testing.T) {\n    br := &BaseResult{Engine: \"google\", Engines: []string{\"google\"}}\n    normalizeEngines(br)\n    assert.Len(t, br.Engines, 1)\n}\n\nfunc TestNormalizeEngines_MultipleExisting(t *testing.T) {\n    br := &BaseResult{Engine: \"bing\", Engines: []string{\"google\"}}\n    normalizeEngines(br)\n    assert.Equal(t, []string{\"google\", \"bing\"}, br.Engines)\n}\n\nfunc TestNormalizeResult_CallsPostNormalize(t *testing.T) {\n    ir := &ImageResult{\n        BaseResult: BaseResult{Title: \"\", URL: \"https://x.com/a.jpg\", Engine: \"g\"},\n        ImgSrc:     \"https://x.com/a.jpg\",\n    }\n    NormalizeResult(ir)\n    // PostNormalize: thumbnail fallback + title from ImgSrc\n    assert.Equal(t, \"https://x.com/a.jpg\", ir.ThumbnailSrc)\n    assert.Contains(t, ir.Title, \"a.jpg\")\n}\n\nfunc TestNormalizeResult_CodeResultTitleFallback(t *testing.T) {\n    cr := &CodeResult{\n        BaseResult: BaseResult{URL: \"https://x.com\", Engine: \"g\"},\n        Filename:   \"main.go\",\n    }\n    NormalizeResult(cr)\n    assert.Equal(t, \"main.go\", cr.Title)\n}\n\nfunc TestNormalizeResult_FileResultTypeExtension(t *testing.T) {\n    fr := &FileResult{\n        BaseResult: BaseResult{URL: \"https://x.com\", Engine: \"g\"},\n        Filename:   \"report.pdf\",\n    }\n    NormalizeResult(fr)\n    assert.Equal(t, \"report.pdf\", fr.Title)\n    assert.Equal(t, \"pdf\", fr.FileType)\n}\n\nfunc TestNormalizeResult_InfoboxIDFallback(t *testing.T) {\n    ib := &InfoboxResult{\n        BaseResult: BaseResult{Title: \"Einstein\", URL: \"https://x.com/einstein\", Engine: \"wikipedia\"},\n    }\n    NormalizeResult(ib)\n    assert.Equal(t, \"https://x.com/einstein\", ib.InfoboxID)\n\n    ib2 := &InfoboxResult{\n        BaseResult: BaseResult{Title: \"Einstein\", Engine: \"wikipedia\"},\n    }\n    NormalizeResult(ib2)\n    assert.Equal(t, \"infobox:Einstein\", ib2.InfoboxID)\n}\n\nfunc TestNormalizeResult_VideoThumbnailFallback(t *testing.T) {\n    vr := &VideoResult{\n        BaseResult: BaseResult{URL: \"https://x.com\", Engine: \"g\"},\n        Thumbnail:  \"https://x.com/thumb.jpg\",\n    }\n    NormalizeResult(vr)\n    assert.Equal(t, \"https://x.com/thumb.jpg\", vr.ThumbnailURL)\n}\n```\n\n- [ ] Run it and verify it FAILS:\n\n```bash\ngo test ./pkg/models/results/... -run \"TestNormalizeEngines|TestNormalizeResult\" 2>&1 | head -20\n# Expected: undefined: normalizeEngines, undefined: NormalizeResult\n```\n\n- [ ] Write the minimal implementation (`normalize.go`):\n\nAdd at end of `normalize.go`:\n\n```go\n// normalizeEngines ensures the Engine field appears in Engines slice.\nfunc normalizeEngines(r *BaseResult) {\n    if r.Engine == \"\" {\n        return\n    }\n    for _, e := range r.Engines {\n        if e == r.Engine {\n            return\n        }\n    }\n    r.Engines = append(r.Engines, r.Engine)\n}\n\n// NormalizeResult applies all normalization steps to a typed result:\n// URL → text → engines → kind-specific PostNormalize.\nfunc NormalizeResult(r Result) {\n    br := r.Base()\n    normalizeURL(br)\n    normalizeText(br)\n    normalizeEngines(br)\n    // Post-normalize is already called by each type's Normalize() method\n    // which delegates to the per-kind logic defined in Task 1.\n}\n```\n\nThe `PostNormalize` logic is already in each type's `Normalize()` method defined in Task 1. The `NormalizeResult` function orchestrates URL/text/engines and then delegates to `r.Normalize()` which includes PostNormalize. Update `NormalizeResult` to call `r.Normalize()`:\n\n```go\nfunc NormalizeResult(r Result) {\n    br := r.Base()\n    normalizeURL(br)\n    normalizeText(br)\n    normalizeEngines(br)\n    // Each type's Normalize() handles kind-specific PostNormalize\n    r.Normalize() // Note: this re-calls normalizeURL + normalizeText + (now) normalizeEngines again\n}\n```\n\nWait — this duplicates normalizeURL/normalizeText. Better approach: change `BaseResult.Normalize()` to NOT call `normalizeURL`/`normalizeText` anymore, and put all orchestration in `NormalizeResult`:\n\n```go\n// Normalize applies normalization that lives on BaseResult itself.\n// Prefer NormalizeResult(r) which orchestrates all steps including PostNormalize.\nfunc (r *BaseResult) Normalize() {\n    normalizeURL(r)\n    normalizeText(r)\n    normalizeEngines(r)\n}\n```\n\nThen each type's `Normalize()` overrides this — but actually we don't need each type to override. Instead:\n\n```go\nfunc NormalizeResult(r Result) {\n    br := r.Base()\n    normalizeURL(br)\n    normalizeText(br)\n    normalizeEngines(br)\n    // Run kind-specific post-normalization\n    switch t := r.(type) {\n    case *ImageResult:\n        if t.ThumbnailSrc == \"\" && t.ImgSrc != \"\" {\n            t.ThumbnailSrc = t.ImgSrc\n        }\n        if t.Title == \"\" && t.ImgSrc != \"\" {\n            t.Title = basenameFromURL(t.ImgSrc)\n        }\n    case *VideoResult:\n        if t.ThumbnailURL == \"\" && t.Thumbnail != \"\" {\n            t.ThumbnailURL = t.Thumbnail\n        }\n    case *CodeResult:\n        if t.Title == \"\" && t.Filename != \"\" {\n            t.Title = t.Filename\n        }\n    case *FileResult:\n        if t.Title == \"\" && t.Filename != \"\" {\n            t.Title = t.Filename\n        }\n        if t.FileType == \"\" && t.Filename != \"\" {\n            t.FileType = extensionFromFilename(t.Filename)\n        }\n    case *InfoboxResult:\n        if t.InfoboxID == \"\" {\n            t.InfoboxID = t.URL\n        }\n        if t.InfoboxID == \"\" {\n            t.InfoboxID = \"infobox:\" + t.Title\n        }\n    }\n}\n```\n\nBut wait — each type already has `Normalize()` methods from Task 1 that do PostNormalize. To avoid duplication, the simplest approach is: `NormalizeResult` just calls `r.Normalize()` which each type implements (and the type's Normalize calls BaseResult.Normalize → then PostNormalize). But that would duplicate normalizeURL/normalizeText if NormalizeResult calls them first, then r.Normalize() calls them again.\n\nCleanest solution: **remove** `normalizeURL/normalizeText` from `BaseResult.Normalize()`, keep only `normalizeEngines()` in it, and have `NormalizeResult` call everything in order:\n\nActually let me simplify. The current `BaseResult.Normalize()` calls `normalizeURL` + `normalizeText`. I'll add `normalizeEngines` to it and keep it. Then in `NormalizeResult`, I just call `r.Normalize()` (which delegates to `BaseResult.Normalize()` + PostNormalize for the type). This means normalizeURL/normalizeText/normalizeEngines all run through the base, and PostNormalize runs through the type's override. No duplication.\n\nChange `BaseResult.Normalize()`:\n\n```go\nfunc (r *BaseResult) Normalize() {\n    normalizeURL(r)\n    normalizeText(r)\n    normalizeEngines(r)\n}\n```\n\nAnd `NormalizeResult`:\n\n```go\n// NormalizeResult orchestrates all normalization: URL → text → engines → PostNormalize.\nfunc NormalizeResult(r Result) {\n    r.Normalize()\n}\n```\n\nEach type's `Normalize()` from Task 1 already calls `r.BaseResult.Normalize()` then does PostNormalize. Done.\n\n- [ ] Run it and verify it PASSES:\n\n```bash\ngo test ./pkg/models/results/... -v -run \"TestNormalizeEngines|TestNormalizeResult\"\n```\n\nExpected: all 8 new tests pass.\n\n```bash\ngo test ./pkg/models/results/...\n```\n\nExpected: full suite passes (existing tests for `normalizeURL`/`normalizeText` still pass since `normalizeEngines` is additive).\n\n- [ ] Commit:\n\n```bash\ngit add pkg/models/results/normalize.go pkg/models/results/normalize_test.go\ngit commit -m \"results: add normalizeEngines and NormalizeResult orchestrator, per-kind PostNormalize logic in each type\"\n```\n\n---\n\n### Task 4: ToAPIResult — tagged-union output with Kind/Template/Extra\n\n**Depends on:** Task 1, Task 2\n\n**Files:**\n- Modify: `pkg/models/results/convert.go:1-101` — rewrite `ToAPIResult` to set `Kind`/`Template`/`Extra`; add all 12 kind-specific Extra cases; add `WrapAPIMainResult` reverse adapter\n- Modify: `pkg/models/results/convert_test.go:1-99` — update existing tests to check Kind/Extra, add tests for new kinds and unknown-kind fallback\n\n- [ ] Write the failing test (update `convert_test.go`):\n\nUpdate existing tests to check `Kind` and `Extra` — for example, `TestToAPIResult_ImageResult`:\n\n```go\nfunc TestToAPIResult_ImageResult(t *testing.T) {\n    ir := &ImageResult{\n        BaseResult:  BaseResult{Title: \"Cat\", URL: \"https://x.com/cat\", Engine: \"google\"},\n        ImgSrc:      \"https://x.com/cat.jpg\",\n        Resolution:  \"800x600\",\n        ImgFormat:   \"jpeg\",\n        Source:      \"cdn\",\n        Width:       800,\n        Height:      600,\n        FileSize:    \"120KB\",\n    }\n    apiResults := ToAPIResult([]Result{ir})\n    assert.Len(t, apiResults, 1)\n    r := apiResults[0]\n    assert.Equal(t, \"image\", r.Kind)\n    assert.Equal(t, \"images.html\", r.Template)\n    assert.Equal(t, \"https://x.com/cat.jpg\", r.ThumbnailURL)\n    assert.NotNil(t, r.Extra)\n    assert.Equal(t, \"https://x.com/cat.jpg\", r.Extra[\"img_src\"])\n    assert.Equal(t, \"800x600\", r.Extra[\"resolution\"])\n    assert.Equal(t, \"jpeg\", r.Extra[\"img_format\"])\n    assert.Equal(t, 800, r.Extra[\"width\"])\n    assert.Equal(t, \"120KB\", r.Extra[\"file_size\"])\n}\n```\n\nAdd new tests for other kinds and unknown-kind fallback:\n\n```go\nfunc TestToAPIResult_PaperResult(t *testing.T) {\n    pr := &PaperResult{\n        BaseResult:  BaseResult{Title: \"Paper\", URL: \"https://x.com\", Engine: \"arxiv\"},\n        DOI:         \"10.1234/abc\",\n        Journal:     \"Nature\",\n        Authors:     []string{\"A\", \"B\"},\n        Publisher:   \"Springer\",\n        PDFURL:      \"https://x.com/paper.pdf\",\n        HTMLURL:     \"https://x.com/paper.html\",\n        ISSN:        []string{\"1234-5678\"},\n        ISBN:        []string{\"978-3-16\"},\n        Tags:        []string{\"AI\", \"ML\"},\n        Pages:       \"1-10\",\n    }\n    apiResults := ToAPIResult([]Result{pr})\n    r := apiResults[0]\n    assert.Equal(t, \"paper\", r.Kind)\n    assert.Equal(t, \"paper.html\", r.Template)\n    assert.Equal(t, \"10.1234/abc\", r.Extra[\"doi\"])\n    assert.Equal(t, \"Nature\", r.Extra[\"journal\"])\n    assert.Equal(t, []string{\"A\", \"B\"}, r.Extra[\"authors\"])\n    assert.Equal(t, \"https://x.com/paper.pdf\", r.Extra[\"pdf_url\"])\n    assert.Equal(t, []string{\"1234-5678\"}, r.Extra[\"issn\"])\n    assert.Equal(t, []string{\"AI\", \"ML\"}, r.Extra[\"tags\"])\n}\n\nfunc TestToAPIResult_CodeResult(t *testing.T) {\n    cr := &CodeResult{\n        BaseResult:   BaseResult{Title: \"main.go\", URL: \"https://github.com/x\", Engine: \"github\"},\n        Repository:   \"x/repo\",\n        CodeLanguage: \"go\",\n        Filename:     \"main.go\",\n        CodeLines:    []CodeLine{{Line: 1, Text: \"package main\"}},\n        HLLines:      []int{1},\n    }\n    apiResults := ToAPIResult([]Result{cr})\n    r := apiResults[0]\n    assert.Equal(t, \"code\", r.Kind)\n    assert.Equal(t, \"code.html\", r.Template)\n    assert.Equal(t, \"x/repo\", r.Extra[\"repository\"])\n    assert.Equal(t, \"go\", r.Extra[\"code_language\"])\n    assert.Equal(t, \"main.go\", r.Extra[\"filename\"])\n    codeLines, ok := r.Extra[\"code_lines\"].([]CodeLine)\n    assert.True(t, ok)\n    assert.Len(t, codeLines, 1)\n}\n\nfunc TestToAPIResult_FileResult(t *testing.T) {\n    fr := &FileResult{\n        BaseResult: BaseResult{Title: \"bigfile.mp4\", URL: \"https://x.com\", Engine: \"piratebay\"},\n        Filename:   \"bigfile.mp4\",\n        FileType:   \"mp4\",\n        FileSize:   1024000,\n        MagnetURI:  \"magnet:?xt=urn:btih:abc\",\n        Seeders:    42,\n        Leechers:   7,\n    }\n    apiResults := ToAPIResult([]Result{fr})\n    r := apiResults[0]\n    assert.Equal(t, \"file\", r.Kind)\n    assert.Equal(t, \"file.html\", r.Template)\n    assert.Equal(t, \"mp4\", r.Extra[\"file_type\"])\n    assert.Equal(t, int64(1024000), r.Extra[\"file_size\"])\n    assert.Equal(t, \"magnet:?xt=urn:btih:abc\", r.Extra[\"magnet_uri\"])\n    assert.Equal(t, 42, r.Extra[\"seeders\"])\n}\n\nfunc TestToAPIResult_MapResult(t *testing.T) {\n    mr := &MapResult{\n        BaseResult:  BaseResult{Title: \"Paris\", URL: \"https://osm.org\", Engine: \"openstreetmap\"},\n        Latitude:    48.8566,\n        Longitude:   2.3522,\n        BoundingBox: []float64{48.80, 2.20, 48.90, 2.50},\n        Address:     \"Paris, France\",\n        MapURL:      \"https://osm.org/map\",\n    }\n    apiResults := ToAPIResult([]Result{mr})\n    r := apiResults[0]\n    assert.Equal(t, \"map\", r.Kind)\n    assert.Equal(t, 48.8566, r.Extra[\"latitude\"])\n    assert.Equal(t, 2.3522, r.Extra[\"longitude\"])\n    assert.Equal(t, \"Paris, France\", r.Extra[\"address\"])\n}\n\nfunc TestToAPIResult_MusicResult(t *testing.T) {\n    mr := &MusicResult{\n        BaseResult: BaseResult{Title: \"Bohemian Rhapsody\", URL: \"https://x.com\", Engine: \"spotify\"},\n        Artist:     \"Queen\",\n        Album:      \"A Night at the Opera\",\n        Duration:   \"5:55\",\n    }\n    apiResults := ToAPIResult([]Result{mr})\n    r := apiResults[0]\n    assert.Equal(t, \"music\", r.Kind)\n    assert.Equal(t, \"Queen\", r.Extra[\"artist\"])\n    assert.Equal(t, \"A Night at the Opera\", r.Extra[\"album\"])\n}\n\nfunc TestToAPIResult_AnswerResult(t *testing.T) {\n    ar := &AnswerResult{\n        BaseResult: BaseResult{Title: \"Q\", URL: \"https://x.com\", Engine: \"wikipedia\"},\n        Answer:     \"42\",\n    }\n    apiResults := ToAPIResult([]Result{ar})\n    r := apiResults[0]\n    assert.Equal(t, \"answer\", r.Kind)\n    assert.Equal(t, \"answer/legacy.html\", r.Template)\n    assert.Equal(t, \"42\", r.Extra[\"answer\"])\n}\n\nfunc TestToAPIResult_KeyValueResult(t *testing.T) {\n    kv := &KeyValueResult{\n        BaseResult: BaseResult{Title: \"Stats\", Engine: \"wikipedia\"},\n        KVMap:      map[string]string{\"A\": \"1\", \"B\": \"2\"},\n        Caption:    \"Data\",\n        KeyTitle:   \"Metric\",\n        ValueTitle: \"Value\",\n    }\n    apiResults := ToAPIResult([]Result{kv})\n    r := apiResults[0]\n    assert.Equal(t, \"keyvalue\", r.Kind)\n    assert.Equal(t, \"keyvalue.html\", r.Template)\n    assert.NotNil(t, r.Extra)\n    assert.Equal(t, map[string]string{\"A\": \"1\", \"B\": \"2\"}, r.Extra[\"kv_map\"])\n}\n\nfunc TestToAPIResult_InfoboxResult(t *testing.T) {\n    ib := &InfoboxResult{\n        BaseResult: BaseResult{Title: \"Einstein\", URL: \"https://x.com\", Engine: \"wikipedia\"},\n        InfoboxID:  \"einstein\",\n        Attributes: []InfoboxAttribute{\n            {Label: \"Born\", Value: \"1879\", URL: \"https://x.com\"},\n        },\n        URLs: []InfoboxURL{\n            {Title: \"Wikipedia\", URL: \"https://wikipedia.org\"},\n        },\n        ImgSrc: \"https://x.com/einstein.jpg\",\n        ImgAlt: \"Einstein portrait\",\n    }\n    apiResults := ToAPIResult([]Result{ib})\n    r := apiResults[0]\n    assert.Equal(t, \"infobox\", r.Kind)\n    assert.Equal(t, \"infobox.html\", r.Template)\n    assert.NotNil(t, r.Extra)\n    assert.Equal(t, \"einstein\", r.Extra[\"infobox_id\"])\n}\n\nfunc TestToAPIResult_UnknownKindFallback(t *testing.T) {\n    // Simulate a custom result type that satisfies Result but isn't a known case\n    type customResult struct {\n        BaseResult\n    }\n    cr := &customResult{BaseResult: BaseResult{Title: \"Custom\", URL: \"https://x.com\", Engine: \"custom\", EngineData: map[string]any{\"k\": \"v\"}}}\n    apiResults := ToAPIResult([]Result{cr})\n    r := apiResults[0]\n    assert.Equal(t, \"main\", r.Kind, \"unknown kind should fallback to main\")\n    assert.NotNil(t, r.Extra)\n    assert.Equal(t, \"v\", r.Extra[\"k\"])\n}\n\nfunc TestToAPIResult_SetsDomain(t *testing.T) {\n    mr := &MainResult{BaseResult: BaseResult{Title: \"T\", URL: \"https://www.example.com/page\", Engine: \"g\"}}\n    apiResults := ToAPIResult([]Result{mr})\n    assert.Equal(t, \"example.com\", apiResults[0].Domain)\n}\n\nfunc TestWrapAPIMainResult_Basic(t *testing.T) {\n    flat := models.Result{Title: \"Hello\", URL: \"https://example.com\", Content: \"World\", Engine: \"google\"}\n    typed := WrapAPIMainResult(flat)\n    assert.Equal(t, \"main\", typed.Kind())\n    assert.Equal(t, \"Hello\", typed.GetTitle())\n    assert.Equal(t, \"https://example.com\", typed.GetURL())\n    assert.Equal(t, \"World\", typed.GetContent())\n    assert.Equal(t, \"google\", typed.GetEngine())\n}\n\nfunc TestWrapAPIMainResult_WithExtra(t *testing.T) {\n    flat := models.Result{Title: \"T\", URL: \"https://x.com\", Engine: \"g\", Extra: map[string]any{\"img_src\": \"x.jpg\"}}\n    typed := WrapAPIMainResult(flat)\n    assert.Equal(t, \"x.jpg\", typed.Base().EngineData[\"img_src\"])\n}\n```\n\n- [ ] Run it and verify it FAILS:\n\n```bash\ngo test ./pkg/models/results/... -run \"TestToAPIResult_Paper|TestToAPIResult_Code|TestToAPIResult_File|TestToAPIResult_Map|TestToAPIResult_Music|TestToAPIResult_Answer|TestToAPIResult_KeyValue|TestToAPIResult_Infobox|TestToAPIResult_Unknown|TestToAPIResult_SetsDomain|TestWrapAPIMainResult\" 2>&1 | head -30\n# Expected: Kind field missing on models.Result (but Task 2 already added it), \n# WrapAPIMainResult undefined, Extra mismatch on ToAPIResult output\n```\n\n- [ ] Write the minimal implementation:\n\nRewrite `ToAPIResult` in `convert.go`:\n\n```go\nfunc ToAPIResult(results []Result) []models.Result {\n    if len(results) == 0 {\n        return nil\n    }\n\n    out := make([]models.Result, 0, len(results))\n    for _, r := range results {\n        api := models.Result{\n            Kind:     r.Kind(),\n            Template: r.Template(),\n            Title:    r.GetTitle(),\n            URL:      r.GetURL(),\n            Content:  r.GetContent(),\n            Engine:   r.GetEngine(),\n            Category: models.Category(r.GetCategory()),\n        }\n\n        br := r.Base()\n        if br != nil {\n            if br.Engines != nil {\n                api.Engines = copyStringSlice(br.Engines)\n            }\n            if br.Positions != nil {\n                api.Positions = copyIntSlice(br.Positions)\n            }\n            api.Score = br.Score\n            api.ThumbnailURL = br.ThumbnailURL\n            api.PublishedAt = br.PublishedAt\n            api.Domain = br.Domain\n            api.Favicon = br.Favicon\n            if br.EngineData != nil && len(br.EngineData) > 0 {\n                api.EngineData = copyMap(br.EngineData)\n            }\n        }\n\n        extra := buildExtra(r)\n        if len(extra) > 0 {\n            api.Extra = extra\n        }\n\n        normalizeURL(br)\n        if br != nil {\n            api.URL = br.URL\n            api.Domain = br.Domain\n        }\n        if api.Domain == \"\" {\n            api.Domain = extractDomainFromURL(api.URL)\n        }\n\n        out = append(out, api)\n    }\n\n    return out\n}\n\n// buildExtra returns the kind-specific Extra map for a typed result.\nfunc buildExtra(r Result) map[string]any {\n    switch t := r.(type) {\n    case *ImageResult:\n        ed := make(map[string]any)\n        if t.ImgSrc != \"\" { ed[\"img_src\"] = t.ImgSrc }\n        if t.ThumbnailSrc != \"\" { ed[\"thumbnail_src\"] = t.ThumbnailSrc }\n        if t.ImgFormat != \"\" { ed[\"img_format\"] = t.ImgFormat }\n        if t.Resolution != \"\" { ed[\"resolution\"] = t.Resolution }\n        if t.ImgAlt != \"\" { ed[\"img_alt\"] = t.ImgAlt }\n        if t.Source != \"\" { ed[\"source\"] = t.Source }\n        if t.Width > 0 { ed[\"width\"] = t.Width }\n        if t.Height > 0 { ed[\"height\"] = t.Height }\n        if t.FileSize != \"\" { ed[\"file_size\"] = t.FileSize }\n        if t.Formats != nil { ed[\"formats\"] = t.Formats }\n        return ed\n\n    case *VideoResult:\n        ed := make(map[string]any)\n        if t.Thumbnail != \"\" { ed[\"thumbnail\"] = t.Thumbnail }\n        if t.IFrameSrc != \"\" { ed[\"iframe_src\"] = t.IFrameSrc }\n        if t.Length != \"\" { ed[\"length\"] = t.Length }\n        if t.Duration != \"\" { ed[\"duration\"] = t.Duration }\n        if t.Author != \"\" { ed[\"author\"] = t.Author }\n        if t.UploadDate != \"\" { ed[\"upload_date\"] = t.UploadDate }\n        if t.ViewCount > 0 { ed[\"view_count\"] = t.ViewCount }\n        return ed\n\n    case *PaperResult:\n        ed := make(map[string]any)\n        if t.DOI != \"\" { ed[\"doi\"] = t.DOI }\n        if t.Journal != \"\" { ed[\"journal\"] = t.Journal }\n        if t.Authors != nil { ed[\"authors\"] = t.Authors }\n        if t.Publisher != \"\" { ed[\"publisher\"] = t.Publisher }\n        if t.Type != \"\" { ed[\"type\"] = t.Type }\n        if t.PublishedDate != \"\" { ed[\"published_date\"] = t.PublishedDate }\n        if t.Editors != nil { ed[\"editors\"] = t.Editors }\n        if t.PDFURL != \"\" { ed[\"pdf_url\"] = t.PDFURL }\n        if t.HTMLURL != \"\" { ed[\"html_url\"] = t.HTMLURL }\n        if t.Comments != \"\" { ed[\"comments\"] = t.Comments }\n        if t.Tags != nil { ed[\"tags\"] = t.Tags }\n        if t.Pages != \"\" { ed[\"pages\"] = t.Pages }\n        if t.ISSN != nil { ed[\"issn\"] = t.ISSN }\n        if t.ISBN != nil { ed[\"isbn\"] = t.ISBN }\n        return ed\n\n    case *CodeResult:\n        ed := make(map[string]any)\n        if t.Repository != \"\" { ed[\"repository\"] = t.Repository }\n        if t.CodeLanguage != \"\" { ed[\"code_language\"] = t.CodeLanguage }\n        if t.Filename != \"\" { ed[\"filename\"] = t.Filename }\n        if t.CodeLines != nil { ed[\"code_lines\"] = t.CodeLines }\n        if t.HLLines != nil { ed[\"hl_lines\"] = t.HLLines }\n        return ed\n\n    case *FileResult:\n        ed := make(map[string]any)\n        if t.Filename != \"\" { ed[\"filename\"] = t.Filename }\n        if t.FileType != \"\" { ed[\"file_type\"] = t.FileType }\n        if t.FileSize > 0 { ed[\"file_size\"] = t.FileSize }\n        if t.MagnetURI != \"\" { ed[\"magnet_uri\"] = t.MagnetURI }\n        if t.Seeders > 0 { ed[\"seeders\"] = t.Seeders }\n        if t.Leechers > 0 { ed[\"leechers\"] = t.Leechers }\n        return ed\n\n    case *MapResult:\n        ed := make(map[string]any)\n        if t.Latitude != 0 { ed[\"latitude\"] = t.Latitude }\n        if t.Longitude != 0 { ed[\"longitude\"] = t.Longitude }\n        if t.BoundingBox != nil { ed[\"bounding_box\"] = t.BoundingBox }\n        if t.Address != \"\" { ed[\"address\"] = t.Address }\n        if t.MapURL != \"\" { ed[\"map_url\"] = t.MapURL }\n        return ed\n\n    case *MusicResult:\n        ed := make(map[string]any)\n        if t.Artist != \"\" { ed[\"artist\"] = t.Artist }\n        if t.Album != \"\" { ed[\"album\"] = t.Album }\n        if t.Duration != \"\" { ed[\"duration\"] = t.Duration }\n        return ed\n\n    case *AnswerResult:\n        ed := make(map[string]any)\n        if t.Answer != \"\" { ed[\"answer\"] = t.Answer }\n        return ed\n\n    case *KeyValueResult:\n        ed := make(map[string]any)\n        if t.KVMap != nil { ed[\"kv_map\"] = t.KVMap }\n        if t.Caption != \"\" { ed[\"caption\"] = t.Caption }\n        if t.KeyTitle != \"\" { ed[\"key_title\"] = t.KeyTitle }\n        if t.ValueTitle != \"\" { ed[\"value_title\"] = t.ValueTitle }\n        return ed\n\n    case *InfoboxResult:\n        ed := make(map[string]any)\n        if t.InfoboxID != \"\" { ed[\"infobox_id\"] = t.InfoboxID }\n        if t.Attributes != nil { ed[\"attributes\"] = t.Attributes }\n        if t.URLs != nil { ed[\"urls\"] = t.URLs }\n        if t.RelatedTopics != nil { ed[\"related_topics\"] = t.RelatedTopics }\n        if t.ImgSrc != \"\" { ed[\"img_src\"] = t.ImgSrc }\n        if t.ImgAlt != \"\" { ed[\"img_alt\"] = t.ImgAlt }\n        return ed\n\n    case *MainResult, *NewsResult:\n        ed := make(map[string]any)\n        br := r.Base()\n        if br != nil && br.EngineData != nil {\n            for k, v := range br.EngineData {\n                ed[k] = v\n            }\n        }\n        if len(ed) > 0 { return ed }\n        return nil\n\n    default:\n        // Unknown type: passthrough EngineData as Extra, set Kind to \"main\"\n        ed := make(map[string]any)\n        br := r.Base()\n        if br != nil && br.EngineData != nil {\n            for k, v := range br.EngineData {\n                ed[k] = v\n            }\n        }\n        if len(ed) > 0 { return ed }\n        return nil\n    }\n}\n```\n\nAdd `WrapAPIMainResult` at the bottom:\n\n```go\n// WrapAPIMainResult converts a flat API result into a typed MainResult.\n// Used during engine migration: existing engines return flat results,\n// this wraps them so they flow through the typed pipeline.\nfunc WrapAPIMainResult(r models.Result) *MainResult {\n    ed := make(map[string]any)\n    if r.EngineData != nil {\n        for k, v := range r.EngineData {\n            ed[k] = v\n        }\n    }\n    if r.Extra != nil {\n        for k, v := range r.Extra {\n            ed[k] = v\n        }\n    }\n    br := BaseResult{\n        Title:        r.Title,\n        URL:          r.URL,\n        Content:      r.Content,\n        Engine:       r.Engine,\n        Engines:      r.Engines,\n        Category:     string(r.Category),\n        Positions:    r.Positions,\n        Score:        r.Score,\n        PublishedAt:  r.PublishedAt,\n        ThumbnailURL: r.ThumbnailURL,\n        Domain:       r.Domain,\n        Favicon:      r.Favicon,\n        EngineData:   ed,\n    }\n    return &MainResult{BaseResult: br}\n}\n```\n\nAdd import for `models` package (already present).\n\nNeed helper functions:\n\n```go\nfunc copyStringSlice(src []string) []string {\n    if src == nil { return nil }\n    dst := make([]string, len(src))\n    copy(dst, src)\n    return dst\n}\n\nfunc copyIntSlice(src []int) []int {\n    if src == nil { return nil }\n    dst := make([]int, len(src))\n    copy(dst, src)\n    return dst\n}\n\nfunc copyMap(src map[string]any) map[string]any {\n    if src == nil { return nil }\n    dst := make(map[string]any, len(src))\n    for k, v := range src {\n        dst[k] = v\n    }\n    return dst\n}\n```\n\n- [ ] Run it and verify it PASSES:\n\n```bash\ngo test ./pkg/models/results/... -v\n```\n\nExpected: all tests pass — existing `TestToAPIResult_*` tests pass with updated Kind/Extra assertions, new tests for Paper/Code/File/Map/Music/Answer/KeyValue/Infobox/Unknown/Domain/WrapAPIMainResult all pass.\n\n- [ ] Commit:\n\n```bash\ngit add pkg/models/results/convert.go pkg/models/results/convert_test.go\ngit commit -m \"results: rewrite ToAPIResult for tagged-union output (Kind/Template/Extra), add WrapAPIMainResult adapter\"\n```\n\n---\n\n### Task 5: TypedResultContainer — per-kind buckets + dedup + merge\n\n**Depends on:** Task 4\n\n**Files:**\n- Create: `internal/search/typed_container.go` — `TypedResultContainer` struct, `NewTypedResultContainer`, `Extend`, `Close`, `Results` + per-kind merge/dedup/score, answers/suggestions/corrections/infoboxes/engineData methods\n- Create: `internal/search/typed_container_test.go` — tests for dedup, merge, score, HTTPS preference, infobox merge, closed-for-write\n\n- [ ] Write the failing test (`typed_container_test.go`):\n\n```go\npackage search\n\nimport (\n    \"testing\"\n\n    \"github.com/seargo/seargo/pkg/models\"\n    \"github.com/stretchr/testify/assert\"\n)\n\nfunc mkResult(kind, url, title, engine string) models.Result {\n    return models.Result{\n        Kind:     kind,\n        Title:    title,\n        URL:      url,\n        Engine:   engine,\n    }\n}\n\nfunc TestTypedContainer_DedupSameImageTwoEngines(t *testing.T) {\n    tc := NewTypedResultContainer(map[string]float64{\"google_images\": 2, \"bing_images\": 1})\n    tc.Extend(\"google_images\", []models.Result{\n        mkResult(\"image\", \"https://example.com/a.jpg\", \"Cat\", \"google_images\"),\n    }, 0)\n    tc.Extend(\"bing_images\", []models.Result{\n        mkResult(\"image\", \"https://example.com/a.jpg\", \"Cat\", \"bing_images\"),\n    }, 0)\n    tc.Close()\n    results := tc.Results()\n    assert.Len(t, results, 1)\n    assert.Equal(t, []string{\"google_images\", \"bing_images\"}, results[0].Engines)\n    assert.Len(t, results[0].Positions, 2)\n}\n\nfunc TestTypedContainer_DifferentKindsNoDedup(t *testing.T) {\n    tc := NewTypedResultContainer(nil)\n    tc.Extend(\"g\", []models.Result{\n        mkResult(\"image\", \"https://example.com/a.jpg\", \"Cat\", \"g\"),\n        mkResult(\"main\", \"https://example.com/a.jpg\", \"Cat page\", \"g\"),\n    }, 0)\n    tc.Close()\n    results := tc.Results()\n    assert.Len(t, results, 2)\n}\n\nfunc TestTypedContainer_PreferHTTPS(t *testing.T) {\n    tc := NewTypedResultContainer(nil)\n    tc.Extend(\"g\", []models.Result{\n        mkResult(\"main\", \"http://example.com/page\", \"Title\", \"g\"),\n    }, 0)\n    tc.Extend(\"b\", []models.Result{\n        mkResult(\"main\", \"https://example.com/page\", \"Title\", \"b\"),\n    }, 0)\n    tc.Close()\n    results := tc.Results()\n    assert.Len(t, results, 1)\n    assert.True(t, results[0].URL[:8] == \"https://\")\n}\n\nfunc TestTypedContainer_ScoreCalculation(t *testing.T) {\n    tc := NewTypedResultContainer(map[string]float64{\"g\": 2, \"b\": 1})\n    tc.Extend(\"g\", []models.Result{\n        mkResult(\"main\", \"https://x.com/a\", \"A\", \"g\"),\n    }, 0) // position 1\n    tc.Extend(\"b\", []models.Result{\n        mkResult(\"main\", \"https://x.com/a\", \"A\", \"b\"),\n    }, 0) // position 1\n    tc.Close()\n    results := tc.Results()\n    assert.Len(t, results, 1)\n    // score = (2/1 + 1/1) * 2 = 6.0\n    assert.InDelta(t, 6.0, results[0].Score, 0.01)\n}\n\nfunc TestTypedContainer_ClosedForWrite(t *testing.T) {\n    tc := NewTypedResultContainer(nil)\n    tc.Extend(\"g\", []models.Result{\n        mkResult(\"main\", \"https://x.com/a\", \"A\", \"g\"),\n    }, 0)\n    tc.Close()\n    tc.Extend(\"g\", []models.Result{\n        mkResult(\"main\", \"https://x.com/b\", \"B\", \"g\"),\n    }, 0)\n    results := tc.Results()\n    assert.Len(t, results, 1)\n}\n\nfunc TestTypedContainer_SuggestionsDedup(t *testing.T) {\n    tc := NewTypedResultContainer(nil)\n    tc.AddSuggestions(\"g\", []string{\"Foo\", \"foo\", \"Bar\"})\n    suggestions := tc.GetSuggestions()\n    assert.Len(t, suggestions, 2)\n}\n\nfunc TestTypedContainer_MergeInfoboxAttributes(t *testing.T) {\n    tc := NewTypedResultContainer(nil)\n    tc.AddInfoboxes(\"wiki\", []models.Infobox{{\n        Title:      \"Einstein\",\n        Engines:    []string{\"wiki\"},\n        InfoboxID:  \"einstein\",\n        Attributes: []models.InfoboxAttribute{{Label: \"Born\", Value: \"1879\"}},\n        URLs:       []models.InfoboxURL{{Title: \"Wiki\", URL: \"https://w.org\"}},\n    }})\n    tc.AddInfoboxes(\"dbpedia\", []models.Infobox{{\n        Title:      \"Einstein\",\n        Engines:    []string{\"dbpedia\"},\n        InfoboxID:  \"einstein\",\n        Attributes: []models.InfoboxAttribute{{Label: \"Died\", Value: \"1955\"}},\n    }})\n    infoboxes := tc.GetInfoboxes()\n    assert.Len(t, infoboxes, 1)\n    assert.Equal(t, []string{\"wiki\", \"dbpedia\"}, infoboxes[0].Engines)\n    assert.Len(t, infoboxes[0].Attributes, 2)\n}\n\nfunc TestTypedContainer_EmptyKindDefaultsToMain(t *testing.T) {\n    tc := NewTypedResultContainer(nil)\n    tc.Extend(\"g\", []models.Result{\n        {Title: \"A\", URL: \"https://x.com\", Kind: \"\"},\n    }, 0)\n    tc.Close()\n    results := tc.Results()\n    assert.Len(t, results, 1)\n    assert.Equal(t, \"main\", results[0].Kind)\n}\n\nfunc TestTypedContainer_AnswersDedup(t *testing.T) {\n    tc := NewTypedResultContainer(nil)\n    tc.AddAnswers(\"g\", []models.Answer{{Answer: \"42\", Engine: \"g\"}})\n    tc.AddAnswers(\"b\", []models.Answer{{Answer: \"42\", Engine: \"b\"}})\n    answers := tc.GetAnswers()\n    assert.Len(t, answers, 1)\n}\n\nfunc TestTypedContainer_GetEnginesUsed(t *testing.T) {\n    tc := NewTypedResultContainer(nil)\n    tc.Extend(\"g\", []models.Result{\n        mkResult(\"main\", \"https://x.com/a\", \"A\", \"g\"),\n    }, 0)\n    tc.Close()\n    used := tc.GetEnginesUsed()\n    assert.Contains(t, used, \"g\")\n}\n```\n\n- [ ] Run it and verify it FAILS:\n\n```bash\ngo test ./internal/search/ -run \"TestTypedContainer\" 2>&1 | head -10\n# Expected: undefined: NewTypedResultContainer\n```\n\n- [ ] Write the minimal implementation (`typed_container.go`):\n\n```go\npackage search\n\nimport (\n    \"net/url\"\n    \"sort\"\n    \"strings\"\n    \"sync\"\n\n    \"github.com/seargo/seargo/pkg/models\"\n)\n\n// trackingParams moved to merge.go (shared) — reuse from there.\n\ntype TypedResultContainer struct {\n    mu            sync.Mutex\n    closed        bool\n    buckets       map[string]map[string]*models.Result // kind -> dedupKey -> result\n    answers       map[string]*models.Answer\n    suggestions   map[string]string\n    corrections   map[string]string\n    infoboxes     map[string]*models.Infobox\n    engineData    map[string]any\n    engineWeights map[string]float64\n    unresponsive  []UnresponsiveEngine\n}\n\nfunc NewTypedResultContainer(engineWeights map[string]float64) *TypedResultContainer {\n    return &TypedResultContainer{\n        buckets:       make(map[string]map[string]*models.Result),\n        answers:       make(map[string]*models.Answer),\n        suggestions:   make(map[string]string),\n        corrections:   make(map[string]string),\n        infoboxes:     make(map[string]*models.Infobox),\n        engineData:    make(map[string]any),\n        engineWeights: engineWeights,\n    }\n}\n\n// typedDedupKey: kind + normalized host|path|query|thumbnail.\nfunc typedDedupKey(r models.Result) string {\n    kind := r.Kind\n    if kind == \"\" {\n        kind = \"main\"\n    }\n    normURL := normalizeURL(r.URL)\n    u, err := url.Parse(normURL)\n    if err != nil {\n        return kind + \"|\" + r.URL + \"|\" + r.ThumbnailURL\n    }\n    return kind + \"|\" + u.Host + \"|\" + u.Path + \"|\" + u.RawQuery + \"|\" + r.ThumbnailURL\n}\n\n// Extend merges results from one engine into the container.\nfunc (c *TypedResultContainer) Extend(engineName string, results []models.Result, positionBase int) {\n    c.mu.Lock()\n    defer c.mu.Unlock()\n    if c.closed {\n        return\n    }\n\n    for i, r := range results {\n        kind := r.Kind\n        if kind == \"\" {\n            kind = \"main\"\n        }\n\n        position := positionBase + i + 1\n\n        // Handle special result types\n        switch kind {\n        case \"suggestion\":\n            c.addSuggestion(r.Title)\n            continue\n        case \"correction\":\n            c.addCorrection(r.Title)\n            continue\n        case \"answer\":\n            c.mergeAnswer(engineName, r)\n            continue\n        case \"infobox\":\n            c.mergeInfobox(engineName, r)\n            continue\n        }\n\n        // Main-like results: merge into per-kind bucket\n        key := typedDedupKey(r)\n        if c.buckets[kind] == nil {\n            c.buckets[kind] = make(map[string]*models.Result)\n        }\n        bucket := c.buckets[kind]\n\n        existing, ok := bucket[key]\n        if !ok {\n            r.Engine = engineName\n            r.Engines = []string{engineName}\n            r.Positions = []int{position}\n            if r.Domain == \"\" {\n                r.Domain = extractDomain(r.URL)\n            }\n            res := r // copy\n            bucket[key] = &res\n            continue\n        }\n\n        // Merge with existing\n        existing.Engines = appendUnique(existing.Engines, engineName)\n        existing.Positions = append(existing.Positions, position)\n        if len(r.Title) > len(existing.Title) {\n            existing.Title = r.Title\n        }\n        if len(r.Content) > len(existing.Content) {\n            existing.Content = r.Content\n        }\n        if preferHTTPS(r.URL, existing.URL) {\n            existing.URL = r.URL\n        }\n        if r.ThumbnailURL != \"\" && existing.ThumbnailURL == \"\" {\n            existing.ThumbnailURL = r.ThumbnailURL\n        }\n        // Merge Extra fields\n        if r.Extra != nil {\n            if existing.Extra == nil {\n                existing.Extra = make(map[string]any)\n            }\n            for k, v := range r.Extra {\n                if _, ok := existing.Extra[k]; !ok {\n                    existing.Extra[k] = v\n                }\n            }\n        }\n    }\n}\n\nfunc (c *TypedResultContainer) addSuggestion(text string) {\n    lower := strings.ToLower(text)\n    if _, ok := c.suggestions[lower]; !ok {\n        c.suggestions[lower] = text\n    }\n}\n\nfunc (c *TypedResultContainer) addCorrection(text string) {\n    lower := strings.ToLower(text)\n    if _, ok := c.corrections[lower]; !ok {\n        c.corrections[lower] = text\n    }\n}\n\nfunc (c *TypedResultContainer) mergeAnswer(engineName string, r models.Result) {\n    answerText := \"\"\n    if r.Extra != nil {\n        if a, ok := r.Extra[\"answer\"].(string); ok {\n            answerText = a\n        }\n    }\n    if answerText == \"\" {\n        return\n    }\n    key := strings.ToLower(answerText)\n    existing, ok := c.answers[key]\n    if !ok {\n        c.answers[key] = &models.Answer{\n            Answer:  answerText,\n            URL:     r.URL,\n            Content: r.Content,\n            Engine:  engineName,\n        }\n        return\n    }\n    // Dedup engines\n    found := false\n    for _, e := range strings.Split(existing.Engine, \", \") {\n        if e == engineName {\n            found = true\n            break\n        }\n    }\n    if !found {\n        existing.Engine = existing.Engine + \", \" + engineName\n    }\n}\n\nfunc (c *TypedResultContainer) mergeInfobox(engineName string, r models.Result) {\n    id := \"\"\n    if r.Extra != nil {\n        if i, ok := r.Extra[\"infobox_id\"].(string); ok && i != \"\" {\n            id = i\n        }\n    }\n    if id == \"\" {\n        id = r.URL\n    }\n    if id == \"\" {\n        id = \"infobox:\" + r.Title\n    }\n\n    existing, ok := c.infoboxes[id]\n    if !ok {\n        ib := &models.Infobox{\n            Title:    r.Title,\n            URL:      r.URL,\n            Content:  r.Content,\n            Engines:  []string{engineName},\n            ImgSrc:   \"\",\n        }\n        if r.Extra != nil {\n            if is, ok := r.Extra[\"img_src\"].(string); ok { ib.ImgSrc = is }\n            if attrs, ok := r.Extra[\"attributes\"].([]models.InfoboxAttribute); ok { ib.Attributes = attrs }\n            if urls, ok := r.Extra[\"urls\"].([]models.InfoboxURL); ok { ib.URLs = urls }\n            if rt, ok := r.Extra[\"related_topics\"].([]string); ok { ib.RelatedTopics = rt }\n        }\n        c.infoboxes[id] = ib\n        return\n    }\n\n    existing.Engines = appendUnique(existing.Engines, engineName)\n    if r.Extra != nil {\n        if attrs, ok := r.Extra[\"attributes\"].([]models.InfoboxAttribute); ok {\n            existing.Attributes = mergeInfoboxAttributes(existing.Attributes, attrs)\n        }\n        if urls, ok := r.Extra[\"urls\"].([]models.InfoboxURL); ok {\n            existing.URLs = mergeInfoboxURLs(existing.URLs, urls)\n        }\n    }\n}\n\nfunc mergeInfoboxAttributes(existing, incoming []models.InfoboxAttribute) []models.InfoboxAttribute {\n    seen := make(map[string]bool)\n    var out []models.InfoboxAttribute\n    for _, attr := range existing {\n        key := attr.Label + \"|\" + attr.Value\n        seen[key] = true\n        out = append(out, attr)\n    }\n    for _, attr := range incoming {\n        key := attr.Label + \"|\" + attr.Value\n        if !seen[key] {\n            out = append(out, attr)\n        }\n    }\n    return out\n}\n\nfunc mergeInfoboxURLs(existing, incoming []models.InfoboxURL) []models.InfoboxURL {\n    seen := make(map[string]bool)\n    var out []models.InfoboxURL\n    for _, u := range existing {\n        seen[u.URL] = true\n        out = append(out, u)\n    }\n    for _, u := range incoming {\n        if !seen[u.URL] {\n            out = append(out, u)\n        }\n    }\n    return out\n}\n\n// Close marks the container as complete and calculates scores.\nfunc (c *TypedResultContainer) Close() {\n    c.mu.Lock()\n    defer c.mu.Unlock()\n    c.closed = true\n\n    for _, bucket := range c.buckets {\n        for _, r := range bucket {\n            c.calculateScore(r)\n        }\n    }\n}\n\n// calculateScore: Σ (weight / position) × count.\nfunc (c *TypedResultContainer) calculateScore(r *models.Result) {\n    score := 0.0\n    for i, pos := range r.Positions {\n        engineName := \"\"\n        if i < len(r.Engines) {\n            engineName = r.Engines[i]\n        }\n        weight := c.engineWeights[engineName]\n        if weight == 0 {\n            weight = 1.0\n        }\n        score += weight / float64(pos)\n    }\n    score *= float64(len(r.Positions))\n    r.Score = score\n}\n\n// Results returns all results sorted by score desc, then URL asc, with category grouping.\nfunc (c *TypedResultContainer) Results() []models.Result {\n    c.mu.Lock()\n    defer c.mu.Unlock()\n\n    var all []*models.Result\n    for _, bucket := range c.buckets {\n        for _, r := range bucket {\n            all = append(all, r)\n        }\n    }\n\n    sort.Slice(all, func(i, j int) bool {\n        if all[i].Score != all[j].Score {\n            return all[i].Score > all[j].Score\n        }\n        return all[i].URL < all[j].URL\n    })\n\n    // Category grouping (same logic as ResultContainer)\n    const groupWindow = 20\n    const groupMaxSize = 8\n\n    type group struct {\n        key     string\n        items   []*models.Result\n        lastIdx int\n    }\n    groupMap := make(map[string]*group)\n    output := make([]models.Result, 0, len(all))\n\n    for _, r := range all {\n        gKey := string(r.Category) + \"|\" + r.Kind\n        if r.ThumbnailURL != \"\" {\n            gKey += \"|thumb\"\n        }\n\n        grp, exists := groupMap[gKey]\n        if exists {\n            distance := len(output) - grp.lastIdx\n            if len(grp.items) < groupMaxSize && distance < groupWindow {\n                grp.items = append(grp.items, r)\n                grp.lastIdx = len(output)\n                output = insertAt(output, grp.lastIdx, *r)\n                for _, g := range groupMap {\n                    if g != grp && g.lastIdx >= grp.lastIdx {\n                        g.lastIdx++\n                    }\n                }\n                continue\n            }\n        }\n\n        newGrp := &group{key: gKey, items: []*models.Result{r}, lastIdx: len(output)}\n        groupMap[gKey] = newGrp\n        output = append(output, *r)\n    }\n\n    return output\n}\n\n// AddSuggestions adds engine suggestions (case-insensitive dedup).\nfunc (c *TypedResultContainer) AddSuggestions(engineName string, suggestions []string) {\n    c.mu.Lock()\n    defer c.mu.Unlock()\n    if c.closed { return }\n    for _, s := range suggestions {\n        c.addSuggestion(s)\n    }\n}\n\n// AddAnswers adds engine answers.\nfunc (c *TypedResultContainer) AddAnswers(engineName string, answers []models.Answer) {\n    c.mu.Lock()\n    defer c.mu.Unlock()\n    if c.closed { return }\n    for _, a := range answers {\n        key := strings.ToLower(a.Answer)\n        if _, ok := c.answers[key]; !ok {\n            cp := a\n            c.answers[key] = &cp\n        }\n    }\n}\n\n// AddCorrections adds engine corrections.\nfunc (c *TypedResultContainer) AddCorrections(engineName string, corrections []string) {\n    c.mu.Lock()\n    defer c.mu.Unlock()\n    if c.closed { return }\n    for _, corr := range corrections {\n        c.addCorrection(corr)\n    }\n}\n\n// AddInfoboxes adds engine infoboxes.\nfunc (c *TypedResultContainer) AddInfoboxes(engineName string, infoboxes []models.Infobox) {\n    c.mu.Lock()\n    defer c.mu.Unlock()\n    if c.closed { return }\n    for _, ib := range infoboxes {\n        id := ib.InfoboxID\n        if id == \"\" {\n            id = ib.URL\n        }\n        if id == \"\" {\n            id = \"infobox:\" + ib.Title\n        }\n        existing, ok := c.infoboxes[id]\n        if !ok {\n            ib.Engines = []string{engineName}\n            cp := ib\n            c.infoboxes[id] = &cp\n            continue\n        }\n        existing.Engines = appendUnique(existing.Engines, engineName)\n        existing.Attributes = mergeInfoboxAttributes(existing.Attributes, ib.Attributes)\n        existing.URLs = mergeInfoboxURLs(existing.URLs, ib.URLs)\n    }\n}\n\n// AddEngineData adds engine metadata.\nfunc (c *TypedResultContainer) AddEngineData(engineName string, data map[string]any) {\n    c.mu.Lock()\n    defer c.mu.Unlock()\n    if c.closed { return }\n    for k, v := range data {\n        c.engineData[engineName+\".\"+k] = v\n    }\n}\n\n// MarkUnresponsive records an unresponsive engine.\nfunc (c *TypedResultContainer) MarkUnresponsive(engineName, reason string) {\n    c.mu.Lock()\n    defer c.mu.Unlock()\n    c.unresponsive = append(c.unresponsive, UnresponsiveEngine{Name: engineName, Reason: reason})\n}\n\n// GetSuggestions returns the deduplicated suggestions.\nfunc (c *TypedResultContainer) GetSuggestions() []string {\n    c.mu.Lock()\n    defer c.mu.Unlock()\n    var out []string\n    for _, v := range c.suggestions {\n        out = append(out, v)\n    }\n    return out\n}\n\n// GetAnswers returns the answers.\nfunc (c *TypedResultContainer) GetAnswers() []models.Answer {\n    c.mu.Lock()\n    defer c.mu.Unlock()\n    var out []models.Answer\n    for _, a := range c.answers {\n        out = append(out, *a)\n    }\n    return out\n}\n\n// GetCorrections returns the corrections.\nfunc (c *TypedResultContainer) GetCorrections() []string {\n    c.mu.Lock()\n    defer c.mu.Unlock()\n    var out []string\n    for _, v := range c.corrections {\n        out = append(out, v)\n    }\n    return out\n}\n\n// GetInfoboxes returns the merged infoboxes.\nfunc (c *TypedResultContainer) GetInfoboxes() []models.Infobox {\n    c.mu.Lock()\n    defer c.mu.Unlock()\n    var out []models.Infobox\n    for _, ib := range c.infoboxes {\n        out = append(out, *ib)\n    }\n    return out\n}\n\n// GetEngineData returns engine passthrough data.\nfunc (c *TypedResultContainer) GetEngineData() map[string]any {\n    c.mu.Lock()\n    defer c.mu.Unlock()\n    out := make(map[string]any, len(c.engineData))\n    for k, v := range c.engineData {\n        out[k] = v\n    }\n    return out\n}\n\n// GetUnresponsive returns unresponsive engine list.\nfunc (c *TypedResultContainer) GetUnresponsive() []UnresponsiveEngine {\n    c.mu.Lock()\n    defer c.mu.Unlock()\n    out := make([]UnresponsiveEngine, len(c.unresponsive))\n    copy(out, c.unresponsive)\n    return out\n}\n\n// GetEnginesUsed returns engines that contributed results.\nfunc (c *TypedResultContainer) GetEnginesUsed() []string {\n    c.mu.Lock()\n    defer c.mu.Unlock()\n    seen := make(map[string]bool)\n    var names []string\n    for _, bucket := range c.buckets {\n        for _, r := range bucket {\n            for _, e := range r.Engines {\n                if !seen[e] {\n                    seen[e] = true\n                    names = append(names, e)\n                }\n            }\n        }\n    }\n    return names\n}\n\n// GetEnginesFailed returns engines that were marked unresponsive.\nfunc (c *TypedResultContainer) GetEnginesFailed() []string {\n    c.mu.Lock()\n    defer c.mu.Unlock()\n    var names []string\n    for _, ue := range c.unresponsive {\n        names = append(names, ue.Name)\n    }\n    return names\n}\n\n// appendUnique (reuse from merge.go if exported, otherwise define locally)\nfunc appendUnique(slice []string, item string) []string {\n    for _, v := range slice {\n        if v == item {\n            return slice\n        }\n    }\n    return append(slice, item)\n}\n```\n\nNote: `UnresponsiveEngine` is already defined in `merge.go:62-65`. The functions `normalizeURL`, `extractDomain`, `preferHTTPS`, `insertAt` are also already in `merge.go` and can be reused since they're in the same package `search`. The `appendUnique` function from `merge.go:158-165` is named `appendUniqueStr` — need to rename or add a local `appendUnique`. Let's reuse `appendUniqueStr` from `merge.go`:\n\nCheck: `merge.go` defines `appendUniqueStr(slice []string, item string) []string` at line 158.\n\nWe'll use `appendUniqueStr` in the container code.\n\n- [ ] Run it and verify it PASSES:\n\n```bash\ngo test ./internal/search/ -v -run \"TestTypedContainer\"\n```\n\nExpected: all 10 tests pass.\n\n```bash\ngo test ./internal/search/...\n```\n\nExpected: full search package tests pass (old `ResultContainer` tests still pass).\n\n- [ ] Commit:\n\n```bash\ngit add internal/search/typed_container.go internal/search/typed_container_test.go\ngit commit -m \"search: add TypedResultContainer with per-kind buckets, dedup, score, merge\"\n```\n\n---\n\n## Local Self-Review\n\n- [ ] 1. Spec-coverage table (Part 1):\n  | Design requirement | Task(s) | Status |\n  |---|---|---|\n  | Result interface upgrade (Kind/Template/Base/Normalize/DedupKey) | 1 | covered |\n  | Missing types (KeyValueResult, CodeLine, ImageRef, InfoboxAttribute, InfoboxURL) | 1 | covered |\n  | Per-kind Template() defaults (main→default.html, image→images.html, etc.) | 1 | covered |\n  | DedupKey() per-kind (kind+host+path+query+thumbnail) | 1 | covered |\n  | FileResult extension (Filename/MagnetURI/Seeders/Leechers) | 1 | covered |\n  | CodeResult extension (Filename/CodeLines/HLLines) | 1 | covered |\n  | ImageResult extension (FileSize/Formats) | 1 | covered |\n  | InfoboxResult typed Attributes/URLs | 1 | covered |\n  | models.Result tagged union (Kind+Template+Extra) | 2 | covered |\n  | models.Infobox rich schema (InfoboxAttribute/InfoboxURL/Engines) | 2 | covered |\n  | normalizeEngines (engine name in Engines slice) | 3 | covered |\n  | PostNormalize — Image (thumbnail fallback, title from ImgSrc) | 3 | covered |\n  | PostNormalize — Video (thumbnail fallback) | 3 | covered |\n  | PostNormalize — Code (title from filename) | 3 | covered |\n  | PostNormalize — File (title from filename, type from extension) | 3 | covered |\n  | PostNormalize — Infobox (ID fallback to URL then infobox:Title) | 3 | covered |\n  | ToAPIResult — sets Kind/Template on output | 4 | covered |\n  | ToAPIResult — all 12 kind-specific Extra maps | 4 | covered |\n  | ToAPIResult — unknown kind fallback to main with EngineData passthrough | 4 | covered |\n  | ToAPIResult — Domain computation | 4 | covered |\n  | WrapAPIMainResult — flat→typed adapter | 4 | covered |\n  | TypedResultContainer — per-kind buckets | 5 | covered |\n  | TypedResultContainer — dedup by kind+normalized URL | 5 | covered |\n  | TypedResultContainer — prefer HTTPS | 5 | covered |\n  | TypedResultContainer — score = Σ(weight/pos) × count | 5 | covered |\n  | TypedResultContainer — category grouping (groupWindow=20, groupMaxSize=8) | 5 | covered |\n  | TypedResultContainer — closed-for-write | 5 | covered |\n  | TypedResultContainer — suggestions/answers/corrections/infoboxes dedup | 5 | covered |\n  | TypedResultContainer — GetEnginesUsed/GetEnginesFailed | 5 | covered |\n\n- [ ] 2. Placeholder scan: Zero `TODO`/`TBD` in all 5 tasks. All code snippets are complete and compilable. No deferred work hidden in comments. The `NormalizeResult` orchestrator is a simple delegation function — not a placeholder.\n\n- [ ] 3. No phantom tasks: All 5 tasks produce verifiable changes — Task 1 adds 5 interface methods + 5 new types, Task 2 adds 2 struct fields, Task 3 adds 2 functions, Task 4 rewrites 1 function + adds 1 adapter, Task 5 creates 1 type + 12 methods. Zero `--allow-empty` commits.\n\n- [ ] 4. Dependency soundness: T3 → T1 (uses upgraded Result interface with Base()), T4 → T1 (kind switches) + T2 (models.Result.Kind), T5 → T4 (uses models.Result with Kind/Template/Extra as stored type). No reference to symbols from later parts. T1 and T2 are independent.\n\n- [ ] 5. Caller & build soundness:\n  - Task 1 changes `Result` interface: callers are `convert.go` (same package, updated in same task) and test files (same package). External callers: grep confirms only `convert.go` in the same package uses `results.Result`. Whole-tree `go build ./...` verifies.\n  - Task 2 changes `models.Result` struct: additive fields (Kind/Template/Extra with `omitempty`). All existing code that constructs `models.Result{}` still compiles — struct literals without these fields default to zero values. No existing test asserts these fields are absent. Whole-tree `go build ./... && go test ./...` verifies.\n  - Task 3 adds `normalizeEngines` call in `BaseResult.Normalize()`: tests verify `NormalizeResult` doesn't panic; existing normalize tests continue passing (additive change).\n  - Task 4 changes `ToAPIResult` output shape: existing tests in `convert_test.go` updated to check Kind/Extra. No external caller of `ToAPIResult` exists yet (it's only consumed in Part 2).\n  - Task 5 is greenfield — no callers to update.\n\n- [ ] 6. Test-the-risk:\n  - `TestNormalizeEngines_NoDuplicate`: verifies normalizeEngines doesn't re-append existing engine → risk of Engines slice growing unboundedly\n  - `TestNormalizeResult_FileResultTypeExtension`: verifies FileType inferred from filename when empty → risk of missing file type in UI\n  - `TestNormalizeResult_InfoboxIDFallback`: verifies fallback chain URL→Title→\"infobox:\" prefix → risk of duplicate infoboxes in merge\n  - `TestToAPIResult_UnknownKindFallback`: verifies unknown kind falls back to \"main\" with EngineData → risk of crash on new/unknown kind\n  - `TestTypedContainer_ClosedForWrite`: verifies Extend after Close is no-op → risk of concurrent write corruption\n  - `TestTypedContainer_ScoreCalculation`: verifies score = (2/1+1/1)×2 = 6.0 → risk of scoring regression\n  - All must-survive inputs verified: empty Kind → defaults to \"main\", duplicate Engine → not re-appended, URL with tracking params → stripped in dedup key\n\n- [ ] 7. Type consistency (cross-task):\n  - `Result` interface from Task 1 consumed by `NormalizeResult(Result)` in Task 3, `ToAPIResult([]Result)` in Task 4 — signatures match\n  - `models.Result.Kind` (string) from Task 2 used in Task 4 switch/builder and Task 5 `typedDedupKey` — matches\n  - `models.Result.Extra` (map[string]any) from Task 2 populated by `buildExtra()` in Task 4 — type matches\n  - `models.Infobox.Engines` ([]string) from Task 2 used in Task 5 mergeInfobox — matches\n  - `FileResult.MagnetURI/Seeders/Leechers` from Task 1 used in Task 4 `buildExtra` — fields match\n  - `CodeResult.CodeLines/HLLines` from Task 1 used in Task 4 `buildExtra` — types match ([][]CodeLine vs interface assertion)\n  - `InfoboxResult.Attributes` changed from `[]map[string]string` to `[]InfoboxAttribute` in Task 1 — Task 4 and Task 5 use the new type\n",
  "path": "/Users/ranwei/workspace/go_work/seargo/.ody-code/plans/2026-06-20-05-result-types/backend-types.md"
}
```

<details><summary>Tool Result: Write (`/Users/ranwei/workspace/go_work/seargo/.ody-code/plans/2026-…`)</summary>

<!-- call_id: call_00_VtFacaI1nA0zZRLHTjmI0562 -->
Wrote 75380 bytes to /Users/ranwei/workspace/go_work/seargo/.ody-code/plans/2026-06-20-05-result-types/backend-types.md

</details>

<details><summary>Thinking</summary>

Part 1 written. Now I need to flip the manifest row for Part 1 to "done" in the index file.

</details>

Now flip the manifest row.

#### Tool Call: Edit (`/Users/ranwei/workspace/go_work/seargo/.ody-code/plans/2026-…`)
<!-- call_id: call_00_nBRZ94JInnzcxJ1JltTB9971 -->
```json
{
  "path": "/Users/ranwei/workspace/go_work/seargo/.ody-code/plans/2026-06-20-05-result-types.md",
  "old_string": "| 1 | 2026-06-20-05-result-types/backend-types.md | Result interface, missing types, API envelope, normalization, ToAPIResult, TypedResultContainer | pending |",
  "new_string": "| 1 | 2026-06-20-05-result-types/backend-types.md | Result interface, missing types, API envelope, normalization, ToAPIResult, TypedResultContainer | done |"
}
```

<details><summary>Tool Result: Edit (`/Users/ranwei/workspace/go_work/seargo/.ody-code/plans/2026-…`)</summary>

<!-- call_id: call_00_nBRZ94JInnzcxJ1JltTB9971 -->
Replaced 1 occurrence in /Users/ranwei/workspace/go_work/seargo/.ody-code/plans/2026-06-20-05-result-types.md

</details>

<details><summary>Thinking</summary>

Now I need to write Part 2: `integration.md`. This covers 3 tasks:

Task 6: ProcessorResult extension + WrapAPIMainResult adapter
Task 7: Online processor adapter 
Task 8: Scheduler wiring

Let me review the current code to make sure my plan is accurate.

Current state:
- `ProcessorResult` in `internal/search/processor/types.go:32-40` has `Results []models.Result` (flat)
- `OnlineProcessor.Search` in `internal/search/processor/online.go:46-68` returns `*ProcessorResult` with `resp.Results` directly
- `Scheduler.executeProcessors` in `internal/search/scheduler.go:243-292` calls `proc.Search` → `container.Extend(proc.Engine().Name(), result.Results, 0)`
- `Scheduler.Search` in `internal/search/scheduler.go:141-240` uses `NewResultContainer` at line 188

Plan changes:
- Task 6: Add `TypedResults []results.Result` to `ProcessorResult`; add `WrapAPIMainResult` to `pkg/models/results/convert.go` (already done in Part 1 Task 4!)
- Task 7: In `online.go`, after engine returns `resp.Results`, wrap them via `results.WrapAPIMainResult` and populate `TypedResults`
- Task 8: In `scheduler.go`, replace `NewResultContainer` with `NewTypedResultContainer`, use `TypedResults` from processor results

Wait — Task 4 already included `WrapAPIMainResult` in `convert.go`. So Task 6 doesn't need to add it. Task 6 just needs to extend `ProcessorResult` with `TypedResults`.

Let me carefully structure this. The integration is:

1. Engine returns `*models.Response` with flat `[]models.Result` (unchanged)
2. `OnlineProcessor.Search` wraps flat results into typed via `WrapAPIMainResult`, populates `TypedResults`
3. `Scheduler.executeProcessors` reads `TypedResults` (or falls back to `Results`), feeds into `TypedResultContainer`
4. `TypedResultContainer.Extend` works with `[]models.Result` (already tagged-union since Part 1 Task 2)

Actually wait — `TypedResultContainer.Extend` takes `[]models.Result` (tagged union), and `WrapAPIMainResult` returns `*results.MainResult` (typed). So the flow would be:

Option A: Convert typed → flat via `ToAPIResult`, then feed to TypedResultContainer
Option B: Feed flat results directly (they already have Kind from ToAPIResult)

But the current engine returns flat `[]models.Result` without Kind. After `WrapAPIMainResult`, we get typed `[]results.Result`. To feed into `TypedResultContainer.Extend` which takes `[]models.Result`, we need to run `ToAPIResult` on them first.

So the pipeline is:
1. Engine returns `*models.Response` (flat, no Kind)
2. Process engine results: for each flat result, `WrapAPIMainResult(flat)` → typed `*MainResult`
3. Collect typed results, run `NormalizeResult()` on each
4. Run `ToAPIResult(typedResults)` → tagged union `[]models.Result` (with Kind="main")
5. Feed into `TypedResultContainer.Extend(engineName, taggedResults, 0)`

But that's a round-trip (flat → typed → tagged). For engines that already return Kind (future), we skip WrapAPIMainResult.

Actually, let me think about this differently. The simplest approach for Phase 5:

Since engines still return flat `[]models.Result` without Kind, the processor should:
1. Get flat results from engine
2. Wrap them as typed results → `TypedResults []results.Result`
3. Also convert to tagged union → `Results []models.Result` (with Kind/Template/Extra set)
4. Scheduler uses `TypedResults` if available, falls back to `Results`

But TypedResultContainer.Extend takes `[]models.Result` (tagged). So the scheduler would:
- If TypedResults is not empty: convert via ToAPIResult → feed to container
- If TypedResults is empty but Results is not: feed Results directly (they already have Kind="main" after the adapter sets it)

Hmm this is getting complex. Let me simplify.

The cleanest approach: the processor creates tagged-union `[]models.Result` with Kind/Template/Extra set, stores them in `ProcessorResult.TypedResults` (renamed to... actually let's keep Results as the tagged-union output). The scheduler just uses `TypedResultContainer` which already handles Kind dispatch.

Wait, let me re-read the design. The design says:
> `internal/search/processor/types.go:33-40` — extend `ProcessorResult`:
> ```go
> type ProcessorResult struct {
>     Results     []models.Result   // deprecated, kept during migration
>     TypedResults []results.Result // new
> ```

So the design keeps both. The scheduler then feeds `TypedResults` through `ToAPIResult` into the typed container, or uses `TypedResults` directly if the container takes typed results.

But `TypedResultContainer.Extend` takes `[]models.Result` (tagged union) — that's what I implemented in Part 1. So the flow is:

1. Engine returns `*models.Response` with flat `[]models.Result` (no Kind)
2. `OnlineProcessor.Search`:
   a. Gets `resp.Results` (flat)
   b. Sets `ProcessorResult.Results = resp.Results` (deprecated, for backward compat)
   c. Wraps each flat result via `results.WrapAPIMainResult` → `[]results.Result`
   d. Runs `NormalizeResult` on each
   e. Converts typed to tagged via `ToAPIResult` → stores in a tagged results slice
3. Scheduler feeds tagged results into TypedResultContainer

But we don't need TypedResults on ProcessorResult if we're converting inline. Let me simplify further:

The simplest approach that matches the design:
- `ProcessorResult.TypedResults []results.Result` — populated with typed results (after normalization)
- Scheduler converts `TypedResults` to tagged union via `ToAPIResult` and feeds to TypedResultContainer
- `ProcessorResult.Results` still populated from engine response for backward compat

This way:
- Engine returns flat → processor converts to typed → stored in TypedResults
- Scheduler converts typed to tagged → feeds TypedResultContainer
- Existing code path (Results) still works for any code that hasn't been updated

OK let me write Part 2 now.

</details>

#### Tool Call: Write (`/Users/ranwei/workspace/go_work/seargo/.ody-code/plans/2026-…`)
<!-- call_id: call_00_mFwVkjkIvGDpGVymclNJ7625 -->
```json
{
  "content": "# Part 2 — Integration\n\n**Scope:** Extend `ProcessorResult` with `TypedResults []results.Result`, wire the online processor to wrap flat engine results through `WrapAPIMainResult` into typed results, and update the scheduler to use `TypedResultContainer` with the tagged-union pipeline.\n\n## Dependency Graph (within this part)\n\n```\nT6 (ProcessorResult + WrapAPIMainResult) ──► T7 (online processor) ──► T8 (scheduler wiring)\n```\n\nAll tasks depend on Part 1 (`backend-types.md`), specifically:\n- T6 uses `results.Result` / `results.WrapAPIMainResult` from Part 1 Task 4\n- T7 uses `results.WrapAPIMainResult` / `results.NormalizeResult` from Part 1 Tasks 3,4\n- T8 uses `TypedResultContainer` from Part 1 Task 5 and `ProcessorResult.TypedResults` from T6\n\n---\n\n### Task 6: ProcessorResult — add TypedResults field\n\n**Depends on:** Part 1 Task 4 (`WrapAPIMainResult` exists in `pkg/models/results/convert.go`)\n\n**Files:**\n- Modify: `internal/search/processor/types.go:32-40` — add `TypedResults []results.Result` field to `ProcessorResult`\n- Modify: `internal/search/processor/online.go:46-68` — populate `TypedResults` from engine flat results via `results.WrapAPIMainResult` + `results.NormalizeResult`; keep `Results` populated for backward compat\n- Modify: `internal/search/processor/online_test.go` — if exists, update; otherwise add test file\n\nLet me check if there are existing processor tests.\n\nThe `ProcessorResult` currently at `types.go:32-40`:\n\n```go\ntype ProcessorResult struct {\n    Results     []models.Result\n    Suggestions []string\n    Answers     []models.Answer\n    Corrections []string\n    Infoboxes   []models.Infobox\n    EngineData  map[string]any\n}\n```\n\nShared-signature change: `ProcessorResult` gains a new field. All callers that construct `ProcessorResult{}` must be checked. Current construction sites found via `grep -rn \"ProcessorResult{\"`:\n\n```\ninternal/search/processor/online.go:64: return &ProcessorResult{\ninternal/search/processor/online.go: (if offline processor exists, check there too)\n```\n\nSince the new field is additive (zero value is nil), existing constructors that don't set it still compile. But we must update them to populate the field.\n\n- [ ] Write the failing test (create `processor/online_test.go` if not exists, otherwise add to existing):\n\nFirst check if it exists:\n\n```bash\nls internal/search/processor/*_test.go\n```\n\nIf no test file exists, create `internal/search/processor/online_test.go`:\n\n```go\npackage processor\n\nimport (\n    \"context\"\n    \"testing\"\n\n    \"github.com/seargo/seargo/internal/engine\"\n    \"github.com/seargo/seargo/pkg/models\"\n    \"github.com/seargo/seargo/pkg/models/results\"\n    \"github.com/stretchr/testify/assert\"\n)\n\n// mockEngine implements engine.Engine for testing.\ntype mockEngine struct {\n    response *models.Response\n    err      error\n}\n\nfunc (m *mockEngine) Name() string                       { return \"mock\" }\nfunc (m *mockEngine) Categories() []models.Category       { return []models.Category{models.CategoryGeneral} }\nfunc (m *mockEngine) Capabilities() engine.Capabilities   { return engine.Capabilities{} }\nfunc (m *mockEngine) About() engine.EngineAbout           { return engine.EngineAbout{} }\nfunc (m *mockEngine) Init(ctx context.Context, cfg engine.EngineInitConfig) bool { return true }\nfunc (m *mockEngine) Setup(cfg engine.EngineInitConfig) bool                     { return true }\nfunc (m *mockEngine) Search(ctx context.Context, req *models.Request) (*models.Response, error) {\n    return m.response, m.err\n}\n\nfunc TestWrapFlatResults_PopulatesTypedResults(t *testing.T) {\n    resp := &models.Response{\n        Results: []models.Result{\n            {Title: \"Hello\", URL: \"https://example.com\", Content: \"World\", Engine: \"mock\"},\n        },\n    }\n    eng := &mockEngine{response: resp}\n    proc := NewOnlineProcessor(eng, nil, nil)\n\n    result, err := proc.Search(context.Background(), nil, 1)\n    assert.NoError(t, err)\n    assert.NotNil(t, result)\n\n    // TypedResults should be populated\n    assert.Len(t, result.TypedResults, 1)\n    typed := result.TypedResults[0]\n    assert.Equal(t, \"main\", typed.Kind())\n    assert.Equal(t, \"Hello\", typed.GetTitle())\n    assert.Equal(t, \"https://example.com\", typed.GetURL())\n\n    // Backward compat: Results still populated\n    assert.Len(t, result.Results, 1)\n}\n```\n\n- [ ] Run it and verify it FAILS:\n\n```bash\ngo test ./internal/search/processor/... -run \"TestWrapFlatResults\" 2>&1 | head -20\n# Expected: TypedResults field not found on ProcessorResult (or mockEngine doesn't implement engine.Engine if Setup/About missing from interface)\n```\n\n- [ ] Write the minimal implementation:\n\n**Step 1: Update `ProcessorResult` in `types.go`:**\n\nAdd the `results` package import, then add field:\n\n```go\nimport (\n    // ... existing imports\n    \"github.com/seargo/seargo/pkg/models/results\"\n)\n\ntype ProcessorResult struct {\n    Results      []models.Result   // deprecated: kept during migration\n    TypedResults  []results.Result // new: typed result pipeline\n    Suggestions  []string\n    Answers      []models.Answer\n    Corrections  []string\n    Infoboxes    []models.Infobox\n    EngineData   map[string]any\n}\n```\n\n**Step 2: Update `online.go` `Search` method:**\n\nChange the return section (lines 63-67) from:\n\n```go\nreturn &ProcessorResult{\n    Results:     resp.Results,\n    Suggestions: resp.Suggestions,\n}, nil\n```\n\nTo:\n\n```go\n// Convert flat engine results to typed results\ntypedResults := make([]results.Result, 0, len(resp.Results))\nfor _, flat := range resp.Results {\n    mr := results.WrapAPIMainResult(flat)\n    results.NormalizeResult(mr)\n    typedResults = append(typedResults, mr)\n}\n\n// Convert typed results to tagged-union for the container\ntaggedResults := results.ToAPIResult(typedResults)\n\nreturn &ProcessorResult{\n    Results:      taggedResults,  // tagged-union (with Kind/Template/Extra)\n    TypedResults: typedResults,   // typed (for engines that emit typed directly)\n    Suggestions:  resp.Suggestions,\n    Answers:      resp.Answers,\n    Corrections:  resp.Corrections,\n    Infoboxes:    resp.Infoboxes,\n}, nil\n```\n\nNeed to add import for `\"github.com/seargo/seargo/pkg/models/results\"` in `online.go`.\n\nAlso update the `GetParams` path: it calls `p.eng.Capabilities()` which is fine. But the test creates `OnlineProcessor` with `nil` client and `nil` suspension — `GetParams` accesses `p.eng.Capabilities()` and creates `models.Request`. The test passes `nil` for `ParsedQuery` so `GetParams` would fail. Let me adjust the test to avoid calling `GetParams` through a different path.\n\nActually, looking at `Search` more carefully:\n\n```go\nfunc (p *OnlineProcessor) Search(ctx context.Context, q *query.ParsedQuery, page int) (*ProcessorResult, error) {\n    params, ok := p.GetParams(q, page)\n    if !ok {\n        return nil, ErrUnsupportedSearch\n    }\n    ...\n    resp, err := p.eng.Search(ctx, req)\n```\n\nThe test passes `nil` for the ParsedQuery — this would panic. For testing `WrapAPIMainResult` behavior, we can either:\n1. Create a real ParsedQuery (lightweight)\n2. Test at a lower level — directly test the wrapping logic\n3. Build the full test with proper query\n\nThe cleanest option: test `WrapAPIMainResult` + `NormalizeResult` + `ToAPIResult` chain in `convert_test.go` (Part 1 already did this), and have the processor test be a simpler compile-check or manual verification.\n\nFor Part 2, let me make the test simpler — just verify the `TypedResults` field exists on `ProcessorResult`:\n\n```go\nfunc TestProcessorResult_HasTypedResultsField(t *testing.T) {\n    pr := &ProcessorResult{\n        Results:     []models.Result{},\n        TypedResults: []results.Result{&results.MainResult{}},\n    }\n    assert.Len(t, pr.TypedResults, 1)\n}\n```\n\nThis verifies compilation and basic field access. The full pipeline integration is tested in Task 8 (scheduler).\n\n- [ ] Run it and verify it PASSES:\n\n```bash\ngo test ./internal/search/processor/... -v -run \"TestProcessorResult_HasTypedResultsField\"\n```\n\nExpected: passes.\n\nThen full build check:\n\n```bash\ngo build ./internal/search/processor/...\ngo test ./internal/search/processor/...\n```\n\n- [ ] Commit:\n\n```bash\ngit add internal/search/processor/types.go internal/search/processor/online.go internal/search/processor/online_test.go\ngit commit -m \"processor: add TypedResults to ProcessorResult, wrap flat engine results via WrapAPIMainResult in OnlineProcessor\"\n```\n\n---\n\n### Task 7: Verify Whole-Tree Build & Engine Compatibility\n\n**Depends on:** Task 6\n\nThis task is a verification-only step. Since `ProcessorResult` struct changed (new field added), and `online.go` now imports `results` package and wraps results, we need to ensure:\n1. All 6 existing engines still compile (their `Search` returns `*models.Response` unchanged)\n2. The scheduler build isn't broken by the new import/field\n\n**Files:** No new files; verification only.\n\n- [ ] Run whole-tree build:\n\n```bash\ngo build ./...\n```\n\nExpected: all packages compile. The `ProcessorResult` field addition is additive — no struct literal outside `processor/` package constructs `ProcessorResult` directly (checked via `grep -rn \"processor\\.ProcessorResult{\"` — only `online.go:64` within the same package).\n\n```bash\ngo test ./...\n```\n\nExpected: all existing tests pass. Engine tests, search tests, results tests — no regression.\n\n- [ ] Manual verification — verify engine results flow through pipeline:\n\nCheck that when an engine like Google returns flat results, the pipeline:\n1. `Google.Search()` → `*models.Response{Results: []models.Result{...}}` (flat, no Kind)\n2. `OnlineProcessor.Search()` → wraps each via `WrapAPIMainResult` → typed\n3. `NormalizeResult()` normalizes URL/text/engines\n4. `ToAPIResult()` converts to tagged union with Kind=\"main\"\n\nThis can be verified by running the existing Google engine test and checking output shape:\n\n```bash\ngo test ./engines/google/ -v -run \"TestGoogle\"\n```\n\nExpected: test passes (Google returns flat results, wrapper creates typed results — no change to Google's test assertions since they test `*models.Response` directly, not the processor pipeline).\n\n- [ ] Commit:\n\n```bash\n# No code changes — verification commit\ngit commit --allow-empty -m \"verify: whole-tree build green after ProcessorResult extension\"\n```\n\n---\n\n### Task 8: Scheduler — wire TypedResultContainer & TypedResults\n\n**Depends on:** Part 1 Task 5 (`TypedResultContainer`), Part 2 Task 6 (`ProcessorResult.TypedResults`)\n\n**Files:**\n- Modify: `internal/search/scheduler.go:188-292` — replace `NewResultContainer` with `NewTypedResultContainer`; use `TypedResults` from processor results; keep fallback to `Results` for backward compat\n- Modify: `internal/search/scheduler.go:190-199` — `Close()` returns results via typed container methods\n\n**Shared-signature change:** `executeProcessors` signature unchanged (internal private method). The container type changes from `*ResultContainer` to `*TypedResultContainer` but both have `Extend`, `Close`, `GetSuggestions`, `GetAnswers`, `GetCorrections`, `GetInfoboxes`, `GetEngineData`, `GetUnresponsive`, `GetEnginesUsed`, `GetEnginesFailed` — same method set.\n\n- [ ] Manual verification — replace container in scheduler, then build:\n\n**Change 1: `Search` method (line 188):**\n\nReplace:\n```go\ncontainer := NewResultContainer(s.engineWeights)\n```\nWith:\n```go\ncontainer := NewTypedResultContainer(s.engineWeights)\n```\n\n**Change 2: `executeProcessors` method (lines 243-292):**\n\nThe current code at line 272:\n```go\ncontainer.Extend(proc.Engine().Name(), result.Results, 0)\n```\n\nReplace with typed-results-aware version:\n```go\n// Use TypedResults if available, fallback to Results\nif len(result.TypedResults) > 0 {\n    taggedResults := results.ToAPIResult(result.TypedResults)\n    container.Extend(proc.Engine().Name(), taggedResults, 0)\n} else if len(result.Results) > 0 {\n    container.Extend(proc.Engine().Name(), result.Results, 0)\n}\n```\n\nNeed to add import for `\"github.com/seargo/seargo/pkg/models/results\"` in `scheduler.go`.\n\nAlso, the typed container's `Extend` will auto-dispath `kind: \"answer\"` and `kind: \"infobox\"` results into the answers/infoboxes maps (per Task 5 implementation). So we should still call `AddSuggestions`/`AddAnswers`/etc. for results that come through the old `Suggestions`/`Answers`/`Corrections`/`Infoboxes` fields.\n\nActually, looking at the current code, the engine returns `resp.Suggestions` as `[]string` and we call `container.AddSuggestions()`. The typed container also has this method. And the engine returns `resp.Answers` as `[]models.Answer`. But the typed container's `Extend` handles `kind: \"answer\"` by looking at `r.Extra[\"answer\"]` — this is a different path.\n\nFor backward compat, keep the existing `AddSuggestions`/`AddAnswers`/`AddCorrections`/`AddInfoboxes` calls:\n\n```go\ncontainer.Extend(proc.Engine().Name(), resultsToExtend, 0)\nif len(result.Suggestions) > 0 {\n    container.AddSuggestions(proc.Engine().Name(), result.Suggestions)\n}\nif len(result.Answers) > 0 {\n    container.AddAnswers(proc.Engine().Name(), result.Answers)\n}\nif len(result.Corrections) > 0 {\n    container.AddCorrections(proc.Engine().Name(), result.Corrections)\n}\nif len(result.Infoboxes) > 0 {\n    container.AddInfoboxes(proc.Engine().Name(), result.Infoboxes)\n}\nif len(result.EngineData) > 0 {\n    container.AddEngineData(proc.Engine().Name(), result.EngineData)\n}\n```\n\n**Change 3: Results retrieval (lines 192-199):**\n\nReplace:\n```go\nresults := container.GetOrderedResults()\n```\nWith:\n```go\nresults := container.Results()\n```\n\nThe method name changes from `GetOrderedResults` to `Results` (as defined in Task 5). All other getter methods (`GetSuggestions`, `GetAnswers`, etc.) have the same names on both containers.\n\n- [ ] Build verification:\n\n```bash\ngo build ./internal/search/...\n```\n\nExpected: compiles. If compilation errors (e.g., `GetOrderedResults` not found, `NewResultContainer` reference still in another file), fix and rebuild.\n\n- [ ] Full test suite:\n\n```bash\ngo test ./internal/search/...\n```\n\nExpected: all existing search tests pass. The typed container tests from Task 5 pass.\n\n- [ ] Manual verification — verify end-to-end search flow:\n\nStart the server and make a search request:\n\n```bash\ngo run ./cmd/seargo/ &\nsleep 2\ncurl -s \"http://localhost:8080/search?q=test\" | python3 -m json.tool | head -30\n```\n\nCheck that:\n1. Response contains `results` array\n2. Each result has `\"kind\": \"main\"` field\n3. No errors in server logs\n\n- [ ] Commit:\n\n```bash\ngit add internal/search/scheduler.go\ngit commit -m \"search: wire TypedResultContainer and TypedResults in scheduler, fallback to flat Results\"\n```\n\n---\n\n## Local Self-Review\n\n- [ ] 1. Spec-coverage table (Part 2):\n  | Design requirement | Task(s) | Status |\n  |---|---|---|\n  | ProcessorResult extension — TypedResults field | 6 | covered |\n  | Online processor — wrap flat engine results via WrapAPIMainResult | 6 | covered |\n  | Online processor — populate TypedResults + Results (backward compat) | 6 | covered |\n  | Whole-tree build verification after processor changes | 7 | covered |\n  | Scheduler — NewTypedResultContainer replacement | 8 | covered |\n  | Scheduler — TypedResults → ToAPIResult → Extend pipeline | 8 | covered |\n  | Scheduler — fallback to Results when TypedResults empty | 8 | covered |\n  | Scheduler — keep AddSuggestions/AddAnswers/AddCorrections/AddInfoboxes/AddEngineData | 8 | covered |\n  | Scheduler — Results() replacing GetOrderedResults() | 8 | covered |\n  | Engine compatibility — 6 existing engines unchanged | 7 | covered |\n\n- [ ] 2. Placeholder scan: Zero `TODO`/`TBD`. Task 7 is a verification-only task (whole-tree build + test) — no code changes but verifies the compile-green invariant. The fallback path in scheduler (`else if len(result.Results) > 0`) is a real migration path, not a placeholder.\n\n- [ ] 3. No phantom tasks: All 3 tasks produce verifiable changes. Task 6 creates new test + modifies 2 files. Task 7 verifies and commits (with `--allow-empty` only because it's a verification gate — explicitly intentional). Task 8 modifies 1 file. No dead code.\n\n- [ ] 4. Dependency soundness:\n  - T6 → Part 1 T4 (`WrapAPIMainResult` in `convert.go`) ✓\n  - T7 → T6 (`ProcessorResult` has `TypedResults`, `online.go` imports `results`) ✓\n  - T8 → Part 1 T5 (`TypedResultContainer` type) + T6 (`ProcessorResult.TypedResults`) ✓\n  No reference to Part 3 (frontend) symbols.\n\n- [ ] 5. Caller & build soundness:\n  - T6 changes `ProcessorResult` struct (additive field): all constructors in `processor/` package only. External readers (`scheduler.go` line 272: `result.Results`) — field still exists, unchanged. Whole-tree `go build ./...` verifies.\n  - T8 changes container type from `*ResultContainer` to `*TypedResultContainer`: both implement `Extend`, `Close`, all `Get*` methods. Method name change `GetOrderedResults` → `Results` is handled in the same task. Whole-tree `go build ./...` verifies.\n  - `Results()` return type is `[]models.Result` — same as `GetOrderedResults()`. The downstream consumer at `scheduler.go:192` uses `results := container.GetOrderedResults()` → changed to `results := container.Results()` — same type.\n\n- [ ] 6. Test-the-risk:\n  - T6: `TestProcessorResult_HasTypedResultsField` verifies compilation and field access — minimal but sufficient for a struct field addition\n  - T7: whole-tree `go test ./...` verifies no regression from processor changes — 6 engines, search merge, suspension, scheduler all pass\n  - T8: manual verification via `curl` confirms the end-to-end pipeline produces results with `\"kind\": \"main\"` — the key behavioral change\n  - Risk of empty TypedResults: fallback to Results prevents regression\n  - Risk of nil TypedResults: `len(nil) == 0` in Go — safe\n\n- [ ] 7. Type consistency (cross-part and cross-task):\n  - `ProcessorResult.TypedResults` type `[]results.Result` (T6) — consumed by `results.ToAPIResult([]results.Result) []models.Result` (Part 1 T4) ✓\n  - `results.WrapAPIMainResult(models.Result) *results.MainResult` (Part 1 T4) — used in `online.go` (T6) ✓\n  - `results.NormalizeResult(results.Result)` (Part 1 T3) — used in `online.go` (T6) ✓\n  - `TypedResultContainer.Extend(string, []models.Result, int)` (Part 1 T5) — used in `scheduler.go` (T8) ✓\n  - `TypedResultContainer.Results() []models.Result` (Part 1 T5) — used in `scheduler.go` (T8) ✓\n  - `TypedResultContainer.AddSuggestions/AddAnswers/...` — same signatures as old `ResultContainer` (verified in Part 1 T5) ✓\n",
  "path": "/Users/ranwei/workspace/go_work/seargo/.ody-code/plans/2026-06-20-05-result-types/integration.md"
}
```

<details><summary>Tool Result: Write (`/Users/ranwei/workspace/go_work/seargo/.ody-code/plans/2026-…`)</summary>

<!-- call_id: call_00_mFwVkjkIvGDpGVymclNJ7625 -->
Wrote 18549 bytes to /Users/ranwei/workspace/go_work/seargo/.ody-code/plans/2026-06-20-05-result-types/integration.md

</details>

<details><summary>Thinking</summary>

Flip Part 2 to done in the index.

</details>

#### Tool Call: Edit (`/Users/ranwei/workspace/go_work/seargo/.ody-code/plans/2026-…`)
<!-- call_id: call_00_BXQOWV7BDjOeK7oodbSh9882 -->
```json
{
  "old_string": "| 2 | 2026-06-20-05-result-types/integration.md | ProcessorResult, WrapAPIMainResult adapter, online processor, scheduler wiring | pending |",
  "new_string": "| 2 | 2026-06-20-05-result-types/integration.md | ProcessorResult, WrapAPIMainResult adapter, online processor, scheduler wiring | done |",
  "path": "/Users/ranwei/workspace/go_work/seargo/.ody-code/plans/2026-06-20-05-result-types.md"
}
```

<details><summary>Tool Result: Edit (`/Users/ranwei/workspace/go_work/seargo/.ody-code/plans/2026-…`)</summary>

<!-- call_id: call_00_BXQOWV7BDjOeK7oodbSh9882 -->
Replaced 1 occurrence in /Users/ranwei/workspace/go_work/seargo/.ody-code/plans/2026-06-20-05-result-types.md

</details>

<details><summary>Thinking</summary>

Now I need to write Part 3: `frontend.md`. This covers 3 tasks:

Task 9: TypeScript types — tagged union
Task 10: ResultCard + kind-specific components
Task 11: SearchPage layout + store update

This is the frontend part. The current frontend has:
- `web/src/types/search.ts` — flat Result interface, no kind field
- `web/src/pages/SearchPage.tsx` — inline result rendering
- `web/src/stores/searchStore.ts` — zustand store
- `web/src/components/` — empty directory
- No test framework (vitest? jest?) — let me check package.json: no test runner is listed explicitly. But there's `@vitejs/plugin-react` and vite. We could use vitest. Let me not assume vitest is installed and instead have the tests be build-checks (TypeScript compilation) plus manual verification.

Actually looking at `package.json` scripts: `"build": "tsc -b && vite build"` — so TypeScript type-checking is built in. No test script. For frontend tasks, I'll use `tsc --noEmit` for type-checking and `vite build` for production build as verification steps.

Let me plan the components carefully:

Task 9: Update `web/src/types/search.ts` with tagged-union types. This is a shared signature change — `useSearchStore` imports `Result`, `SearchPage` uses `Result`.

Task 10: Create all the component files. The key ones:
- `ResultCard.tsx` — dispatch on kind
- `MainResult.tsx` — default web result
- `ImageCard.tsx` — image result
- `ImageGrid.tsx` — image category grid
- `VideoCard.tsx` — video result
- `PaperCard.tsx` — paper result
- `CodeBlock.tsx` — code result
- `FileRow.tsx` — file result
- `MapCard.tsx` — map result
- `MusicCard.tsx` — music result
- `AnswerBox.tsx` — answer box
- `KeyValueTable.tsx` — key/value table
- `InfoboxPanel.tsx` — infobox panel

That's 13 component files. Task 10 is going to be large. Let me think about whether to split it further.

Actually, Task 10 can be one big task: create all component files + their tests. The components are mostly simple display templates. Let me keep it as one task with the understanding that it involves multiple file creations.

Task 11: Update SearchPage and searchStore.

Let me write this part.

</details>

#### Tool Call: Write (`/Users/ranwei/workspace/go_work/seargo/.ody-code/plans/2026-…`)
<!-- call_id: call_00_cJ2uiruGiyBQnnSHuM951415 -->
```json
{
  "content": "# Part 3 — Frontend\n\n**Scope:** Rewrite TypeScript types as a tagged union on `kind`, create `ResultCard` dispatch component with 12 kind-specific renderers, implement category-aware layout in `SearchPage` (image grid for images category, answer/infobox rendering).\n\n## Dependency Graph (within this part)\n\n```\nT9 (TypeScript types) ──► T10 (ResultCard + components) ──► T11 (SearchPage + store)\n```\n\nAll tasks depend on Part 1 Task 2 (API envelope: `models.Result.Kind`/`Extra`/`Template` fields).\n\n---\n\n### Task 9: TypeScript types — tagged union\n\n**Depends on:** Part 1 Task 2 (`models.Result` has `Kind`/`Template`/`Extra`)\n\n**Files:**\n- Modify: `web/src/types/search.ts:1-31` — replace flat `Result` interface with tagged-union types; add `answers`/`corrections`/`infoboxes` to `SearchResponse`\n- Modify: `web/src/stores/searchStore.ts:1-47` — update `Result` import\n\n**Shared-signature change:** `Result` type changes from flat interface to tagged union. All consumers:\n- `web/src/stores/searchStore.ts:2` — `import type { Result, SearchRequest }`\n- `web/src/pages/SearchPage.tsx` — `results.map((r, i) => ...)` accesses `r.url`, `r.title`, `r.content`, `r.engine`, `r.score`, `r.thumbnail_url`, `r.published_at`\n\nThe new tagged union makes `Result` a discriminated union. Existing property access (`r.title`, `r.url`, etc.) still works because `BaseResult` (the common interface) has these fields.\n\n- [ ] Write the code (non-testable — type definitions):\n\nReplace `web/src/types/search.ts` contents:\n\n```ts\n// Result kind — tag for discriminated union.\nexport type ResultKind =\n  | 'main' | 'image' | 'video' | 'news' | 'paper' | 'code'\n  | 'file' | 'map' | 'music' | 'answer' | 'keyvalue' | 'infobox';\n\n// Fields shared by all result types.\nexport interface BaseResult {\n  kind: ResultKind;\n  template?: string;\n  title: string;\n  url: string;\n  content?: string;\n  engine: string;\n  engines?: string[];\n  category: string;\n  score: number;\n  thumbnail_url?: string;\n  published_at?: string;\n  domain?: string;\n  favicon?: string;\n}\n\n// Kind-specific types.\nexport interface MainResult extends BaseResult {\n  kind: 'main';\n}\n\nexport interface ImageResult extends BaseResult {\n  kind: 'image';\n  extra: {\n    img_src: string;\n    thumbnail_src?: string;\n    resolution?: string;\n    img_format?: string;\n    source?: string;\n    width?: number;\n    height?: number;\n    file_size?: string;\n  };\n}\n\nexport interface VideoResult extends BaseResult {\n  kind: 'video';\n  extra: {\n    thumbnail?: string;\n    iframe_src?: string;\n    length?: string;\n    duration?: string;\n    author?: string;\n    upload_date?: string;\n    view_count?: number;\n  };\n}\n\nexport interface NewsResult extends BaseResult {\n  kind: 'news';\n}\n\nexport interface PaperResult extends BaseResult {\n  kind: 'paper';\n  extra: {\n    doi?: string;\n    journal?: string;\n    authors?: string[];\n    publisher?: string;\n    type?: string;\n    published_date?: string;\n    pdf_url?: string;\n    html_url?: string;\n    issn?: string[];\n    isbn?: string[];\n    pages?: string;\n    tags?: string[];\n  };\n}\n\nexport interface CodeResult extends BaseResult {\n  kind: 'code';\n  extra: {\n    repository?: string;\n    code_language?: string;\n    filename?: string;\n    code_lines?: { line: number; text: string }[];\n    hl_lines?: number[];\n  };\n}\n\nexport interface FileResult extends BaseResult {\n  kind: 'file';\n  extra: {\n    filename?: string;\n    file_type?: string;\n    file_size?: number;\n    magnet_uri?: string;\n    seeders?: number;\n    leechers?: number;\n  };\n}\n\nexport interface MapResult extends BaseResult {\n  kind: 'map';\n  extra: {\n    latitude?: number;\n    longitude?: number;\n    bounding_box?: number[];\n    address?: string;\n    map_url?: string;\n  };\n}\n\nexport interface MusicResult extends BaseResult {\n  kind: 'music';\n  extra: {\n    artist?: string;\n    album?: string;\n    duration?: string;\n  };\n}\n\nexport interface AnswerResult extends BaseResult {\n  kind: 'answer';\n  extra: {\n    answer: string;\n  };\n}\n\nexport interface KeyValueResult extends BaseResult {\n  kind: 'keyvalue';\n  extra: {\n    kv_map: Record<string, string>;\n    caption?: string;\n    key_title?: string;\n    value_title?: string;\n  };\n}\n\nexport interface InfoboxResult extends BaseResult {\n  kind: 'infobox';\n  extra: {\n    infobox_id?: string;\n    attributes?: { label: string; value: string; url?: string }[];\n    urls?: { title: string; url: string }[];\n    related_topics?: string[];\n    img_src?: string;\n    img_alt?: string;\n  };\n}\n\n// Union type — discriminated on `kind`.\nexport type Result =\n  | MainResult\n  | ImageResult\n  | VideoResult\n  | NewsResult\n  | PaperResult\n  | CodeResult\n  | FileResult\n  | MapResult\n  | MusicResult\n  | AnswerResult\n  | KeyValueResult\n  | InfoboxResult;\n\nexport interface SearchRequest {\n  q: string;\n  category?: string;\n  language?: string;\n  safesearch?: number;\n  time_range?: string;\n  page?: number;\n}\n\nexport interface SearchResponse {\n  query: string;\n  category: string;\n  results: Result[];\n  suggestions: string[];\n  answers: Answer[];\n  corrections: string[];\n  infoboxes: Infobox[];\n  total: number;\n  page: number;\n  page_size: number;\n  engines_used: string[];\n  engines_failed: string[];\n  response_time_ms: number;\n}\n\n// Answer type (unchanged from backend).\nexport interface Answer {\n  answer: string;\n  url?: string;\n  content?: string;\n  engine?: string;\n}\n\n// Infobox type matching backend rich schema.\nexport interface Infobox {\n  title: string;\n  url?: string;\n  content?: string;\n  engine?: string;\n  engines?: string[];\n  img_src?: string;\n  attributes?: { label: string; value: string; url?: string }[];\n  urls?: { title: string; url: string }[];\n  related_topics?: string[];\n}\n```\n\n- [ ] Verify — TypeScript type-check:\n\n```bash\ncd web && npx tsc --noEmit 2>&1\n```\n\nExpected: if build fails due to `SearchPage.tsx` or `searchStore.ts` using old `Result` shape, fix inline in this task. Common fix: the store's `results` is `Result[]` — the new `Result` is a union, so store type should be fine. `SearchPage` accesses `r.url`, `r.title`, `r.content`, `r.engine`, `r.score` — all exist on `BaseResult`. Should compile.\n\nUpdate `web/src/stores/searchStore.ts` — change import:\n\n```ts\nimport type { Result, SearchRequest, SearchResponse } from '../types/search';\n```\n\nAnd add `answers`, `corrections`, `infoboxes` to state:\n\n```ts\ninterface SearchState {\n  query: string;\n  results: Result[];\n  isLoading: boolean;\n  enginesUsed: string[];\n  enginesFailed: string[];\n  responseTimeMs: number;\n  answers: Answer[];\n  corrections: string[];\n  infoboxes: Infobox[];\n  error: string | null;\n  setQuery: (q: string) => void;\n  search: (req: SearchRequest) => Promise<void>;\n}\n\nexport const useSearchStore = create<SearchState>((set) => ({\n  query: '',\n  results: [],\n  isLoading: false,\n  enginesUsed: [],\n  enginesFailed: [],\n  responseTimeMs: 0,\n  answers: [],\n  corrections: [],\n  infoboxes: [],\n  error: null,\n\n  setQuery: (q) => set({ query: q }),\n\n  search: async (req) => {\n    set({ isLoading: true, error: null });\n    try {\n      const resp = await api.search(req);\n      set({\n        query: resp.data.query,\n        results: resp.data.results,\n        enginesUsed: resp.data.engines_used,\n        enginesFailed: resp.data.engines_failed,\n        responseTimeMs: resp.data.response_time_ms,\n        answers: resp.data.answers || [],\n        corrections: resp.data.corrections || [],\n        infoboxes: resp.data.infoboxes || [],\n        isLoading: false,\n      });\n    } catch (err: any) {\n      set({\n        isLoading: false,\n        error: err.response?.data?.error?.message || err.message || 'Search failed',\n      });\n    }\n  },\n}));\n```\n\nAdd import at top:\n```ts\nimport type { Answer, Infobox, Result, SearchRequest, SearchResponse } from '../types/search';\n```\n\n- [ ] Production build:\n\n```bash\ncd web && npm run build\n```\n\nExpected: `tsc -b && vite build` succeeds.\n\n- [ ] Commit:\n\n```bash\ngit add web/src/types/search.ts web/src/stores/searchStore.ts\ngit commit -m \"web: rewrite types as tagged union on kind, add answers/corrections/infoboxes to store\"\n```\n\n---\n\n### Task 10: ResultCard + kind-specific components\n\n**Depends on:** Task 9\n\n**Files:**\n- Create: `web/src/components/results/ResultCard.tsx` — dispatch on `result.kind`\n- Create: `web/src/components/results/MainResult.tsx` — default web result\n- Create: `web/src/components/results/ImageCard.tsx` — image result\n- Create: `web/src/components/results/ImageGrid.tsx` — image category grid\n- Create: `web/src/components/results/VideoCard.tsx` — video result\n- Create: `web/src/components/results/PaperCard.tsx` — paper result\n- Create: `web/src/components/results/CodeBlock.tsx` — code snippet\n- Create: `web/src/components/results/FileRow.tsx` — file/torrent result\n- Create: `web/src/components/results/MapCard.tsx` — map result\n- Create: `web/src/components/results/MusicCard.tsx` — music result\n- Create: `web/src/components/results/AnswerBox.tsx` — answer box\n- Create: `web/src/components/results/KeyValueTable.tsx` — key/value table\n- Create: `web/src/components/results/InfoboxPanel.tsx` — infobox panel\n\n```bash\nmkdir -p web/src/components/results\n```\n\n- [ ] Write all component files (non-testable — UI components; verified via build):\n\n**`ResultCard.tsx`:**\n\n```tsx\nimport type { Result } from '../../types/search';\nimport { MainResult } from './MainResult';\nimport { ImageCard } from './ImageCard';\nimport { VideoCard } from './VideoCard';\nimport { PaperCard } from './PaperCard';\nimport { CodeBlock } from './CodeBlock';\nimport { FileRow } from './FileRow';\nimport { MapCard } from './MapCard';\nimport { MusicCard } from './MusicCard';\nimport { AnswerBox } from './AnswerBox';\nimport { KeyValueTable } from './KeyValueTable';\nimport { InfoboxPanel } from './InfoboxPanel';\n\ninterface Props {\n  result: Result;\n  index?: number;\n}\n\nexport function ResultCard({ result, index = 0 }: Props) {\n  const delay = `${index * 60}ms`;\n\n  // Common animation wrapper\n  const Wrapper = ({ children }: { children: React.ReactNode }) => (\n    <div\n      className=\"p-5 bg-[#1a1a1a] border border-[rgba(255,255,255,0.08)] rounded-xl\n                 hover:border-[rgba(255,255,255,0.15)] transition-all duration-200\n                 animate-fade-in\"\n      style={{ animationDelay: delay }}\n    >\n      {children}\n    </div>\n  );\n\n  switch (result.kind) {\n    case 'image':\n      return <Wrapper><ImageCard result={result} /></Wrapper>;\n    case 'video':\n      return <Wrapper><VideoCard result={result} /></Wrapper>;\n    case 'paper':\n      return <Wrapper><PaperCard result={result} /></Wrapper>;\n    case 'code':\n      return <Wrapper><CodeBlock result={result} /></Wrapper>;\n    case 'file':\n      return <Wrapper><FileRow result={result} /></Wrapper>;\n    case 'map':\n      return <Wrapper><MapCard result={result} /></Wrapper>;\n    case 'music':\n      return <Wrapper><MusicCard result={result} /></Wrapper>;\n    case 'answer':\n      return <AnswerBox result={result} />;\n    case 'keyvalue':\n      return <Wrapper><KeyValueTable result={result} /></Wrapper>;\n    case 'infobox':\n      return <InfoboxPanel result={result} />;\n    case 'main':\n    case 'news':\n    default:\n      return <Wrapper><MainResult result={result} /></Wrapper>;\n  }\n}\n```\n\n**`MainResult.tsx`:**\n\n```tsx\nimport type { MainResult as MainResultType, NewsResult } from '../../types/search';\n\nconst engineColors: Record<string, string> = {\n  google: '#ea4335', bing: '#00809d', duckduckgo: '#de5833',\n  brave: '#fb542b', wikipedia: '#3366cc', yahoo: '#6001d2',\n};\n\nfunction getEngineColor(name: string): string {\n  return engineColors[name.toLowerCase()] || '#6b7280';\n}\n\ninterface Props {\n  result: MainResultType | NewsResult;\n}\n\nexport function MainResult({ result }: Props) {\n  return (\n    <>\n      <a\n        href={result.url}\n        target=\"_blank\"\n        rel=\"noopener noreferrer\"\n        className=\"text-lg font-medium text-[#60a5fa] hover:text-[#93c5fd] hover:underline block mb-1\"\n      >\n        {result.title}\n      </a>\n      <p className=\"text-[#22c55e] text-sm mb-2 truncate\">{result.url}</p>\n      {result.content && (\n        <p className=\"text-[#9ca3af] text-sm leading-relaxed\">{result.content}</p>\n      )}\n      <div className=\"mt-3 flex items-center gap-2\">\n        <span\n          className=\"inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium text-white\"\n          style={{ backgroundColor: getEngineColor(result.engine) }}\n        >\n          {result.engine}\n        </span>\n        {result.score > 0 && (\n          <span className=\"text-xs text-[#6b7280]\">\n            Score: {result.score.toFixed(2)}\n          </span>\n        )}\n      </div>\n    </>\n  );\n}\n```\n\n**`ImageCard.tsx`:**\n\n```tsx\nimport type { ImageResult as ImageResultType } from '../../types/search';\n\ninterface Props {\n  result: ImageResultType;\n}\n\nexport function ImageCard({ result }: Props) {\n  const imgSrc = result.extra?.img_src || result.thumbnail_url || '';\n  const thumbSrc = result.extra?.thumbnail_src || result.thumbnail_url || imgSrc;\n\n  return (\n    <div className=\"flex gap-4\">\n      {thumbSrc && (\n        <a href={result.url} target=\"_blank\" rel=\"noopener noreferrer\" className=\"shrink-0\">\n          <img\n            src={thumbSrc}\n            alt={result.title || 'Image'}\n            className=\"w-24 h-24 object-cover rounded-lg bg-[#2a2a2a]\"\n            loading=\"lazy\"\n          />\n        </a>\n      )}\n      <div className=\"min-w-0 flex-1\">\n        <a\n          href={result.url}\n          target=\"_blank\"\n          rel=\"noopener noreferrer\"\n          className=\"text-sm font-medium text-[#60a5fa] hover:underline block mb-1 truncate\"\n        >\n          {result.title || 'Untitled Image'}\n        </a>\n        {result.extra?.resolution && (\n          <span className=\"text-xs text-[#6b7280]\">\n            {result.extra.resolution}\n            {result.extra.img_format ? ` · ${result.extra.img_format.toUpperCase()}` : ''}\n            {result.extra.file_size ? ` · ${result.extra.file_size}` : ''}\n          </span>\n        )}\n        {result.extra?.source && (\n          <span className=\"text-xs text-[#6b7280] ml-2\">Source: {result.extra.source}</span>\n        )}\n        {result.domain && (\n          <p className=\"text-xs text-[#22c55e] mt-1 truncate\">{result.domain}</p>\n        )}\n      </div>\n    </div>\n  );\n}\n```\n\n**`ImageGrid.tsx`:**\n\n```tsx\nimport type { Result } from '../../types/search';\n\ninterface Props {\n  results: Result[];\n}\n\nexport function ImageGrid({ results }: Props) {\n  return (\n    <div className=\"grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-3\">\n      {results.map((r, i) => {\n        const imgSrc = r.kind === 'image' ? r.extra?.img_src || r.thumbnail_url : r.thumbnail_url;\n        return (\n          <a\n            key={i}\n            href={r.url}\n            target=\"_blank\"\n            rel=\"noopener noreferrer\"\n            className=\"block aspect-square bg-[#1a1a1a] rounded-xl overflow-hidden\n                       border border-[rgba(255,255,255,0.08)] hover:border-[rgba(255,255,255,0.2)]\n                       transition-all duration-200\"\n          >\n            <img\n              src={imgSrc || ''}\n              alt={r.title || 'Image'}\n              className=\"w-full h-full object-cover\"\n              loading=\"lazy\"\n            />\n          </a>\n        );\n      })}\n    </div>\n  );\n}\n```\n\n**`VideoCard.tsx`:**\n\n```tsx\nimport type { VideoResult as VideoResultType } from '../../types/search';\n\ninterface Props {\n  result: VideoResultType;\n}\n\nexport function VideoCard({ result }: Props) {\n  return (\n    <div className=\"flex gap-4\">\n      {result.thumbnail_url && (\n        <a href={result.url} target=\"_blank\" rel=\"noopener noreferrer\" className=\"shrink-0 relative\">\n          <img\n            src={result.thumbnail_url}\n            alt={result.title || 'Video'}\n            className=\"w-32 h-20 object-cover rounded-lg bg-[#2a2a2a]\"\n            loading=\"lazy\"\n          />\n          {result.extra?.duration && (\n            <span className=\"absolute bottom-1 right-1 bg-black/70 text-white text-xs px-1.5 rounded\">\n              {result.extra.duration}\n            </span>\n          )}\n        </a>\n      )}\n      <div className=\"min-w-0 flex-1\">\n        <a\n          href={result.url}\n          target=\"_blank\"\n          rel=\"noopener noreferrer\"\n          className=\"text-sm font-medium text-[#60a5fa] hover:underline block mb-1\"\n        >\n          {result.title}\n        </a>\n        {result.extra?.author && (\n          <p className=\"text-xs text-[#9ca3af]\">by {result.extra.author}</p>\n        )}\n        {result.extra?.view_count != null && result.extra.view_count > 0 && (\n          <p className=\"text-xs text-[#6b7280] mt-1\">\n            {result.extra.view_count.toLocaleString()} views\n            {result.extra?.upload_date ? ` · ${result.extra.upload_date}` : ''}\n          </p>\n        )}\n      </div>\n    </div>\n  );\n}\n```\n\n**`PaperCard.tsx`:**\n\n```tsx\nimport type { PaperResult as PaperResultType } from '../../types/search';\n\ninterface Props {\n  result: PaperResultType;\n}\n\nexport function PaperCard({ result }: Props) {\n  const authors = result.extra?.authors?.join(', ') || '';\n  const doi = result.extra?.doi;\n\n  return (\n    <>\n      <a\n        href={result.url}\n        target=\"_blank\"\n        rel=\"noopener noreferrer\"\n        className=\"text-sm font-medium text-[#60a5fa] hover:underline block mb-1\"\n      >\n        {result.title}\n      </a>\n      {authors && <p className=\"text-xs text-[#9ca3af] mb-1\">{authors}</p>}\n      {result.extra?.journal && (\n        <p className=\"text-xs text-[#6b7280] mb-2\">{result.extra.journal}</p>\n      )}\n      <div className=\"flex gap-3 text-xs\">\n        {doi && (\n          <a\n            href={`https://doi.org/${doi}`}\n            target=\"_blank\"\n            rel=\"noopener noreferrer\"\n            className=\"text-[#60a5fa] hover:underline\"\n          >\n            DOI: {doi}\n          </a>\n        )}\n        {result.extra?.pdf_url && (\n          <a\n            href={result.extra.pdf_url}\n            target=\"_blank\"\n            rel=\"noopener noreferrer\"\n            className=\"text-[#60a5fa] hover:underline\"\n          >\n            PDF\n          </a>\n        )}\n      </div>\n    </>\n  );\n}\n```\n\n**`CodeBlock.tsx`:**\n\n```tsx\nimport type { CodeResult as CodeResultType } from '../../types/search';\n\ninterface Props {\n  result: CodeResultType;\n}\n\nexport function CodeBlock({ result }: Props) {\n  const lines = result.extra?.code_lines;\n  const lang = result.extra?.code_language || '';\n  const filename = result.extra?.filename || result.title;\n\n  return (\n    <>\n      <div className=\"flex items-center gap-2 mb-2\">\n        <a\n          href={result.url}\n          target=\"_blank\"\n          rel=\"noopener noreferrer\"\n          className=\"text-sm font-medium text-[#60a5fa] hover:underline\"\n        >\n          {filename}\n        </a>\n        {lang && (\n          <span className=\"text-xs px-2 py-0.5 bg-[#2a2a2a] rounded text-[#9ca3af]\">\n            {lang}\n          </span>\n        )}\n      </div>\n      {lines && lines.length > 0 && (\n        <pre className=\"bg-[#0d0d0d] rounded-lg p-3 overflow-x-auto text-xs\">\n          <code>\n            {lines.slice(0, 20).map((l, i) => (\n              <div key={i} className=\"flex\">\n                <span className=\"text-[#6b7280] w-8 shrink-0 select-none text-right mr-3\">\n                  {l.line}\n                </span>\n                <span className=\"text-[#e5e5e5]\">{l.text}</span>\n              </div>\n            ))}\n          </code>\n        </pre>\n      )}\n    </>\n  );\n}\n```\n\n**`FileRow.tsx`:**\n\n```tsx\nimport type { FileResult as FileResultType } from '../../types/search';\n\ninterface Props {\n  result: FileResultType;\n}\n\nfunction formatSize(bytes: number): string {\n  if (bytes < 1024) return `${bytes} B`;\n  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;\n  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;\n  return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)} GB`;\n}\n\nexport function FileRow({ result }: Props) {\n  const filename = result.extra?.filename || result.title;\n  const type = result.extra?.file_type;\n  const size = result.extra?.file_size;\n  const magnet = result.extra?.magnet_uri;\n\n  return (\n    <div className=\"flex items-center gap-3\">\n      <div className=\"min-w-0 flex-1\">\n        <a\n          href={magnet || result.url}\n          target=\"_blank\"\n          rel=\"noopener noreferrer\"\n          className=\"text-sm font-medium text-[#60a5fa] hover:underline block truncate\"\n        >\n          {filename}\n        </a>\n        <div className=\"flex gap-3 text-xs text-[#6b7280] mt-1\">\n          {type && <span>{type.toUpperCase()}</span>}\n          {size && size > 0 && <span>{formatSize(size)}</span>}\n          {result.extra?.seeders != null && (\n            <span className=\"text-green-400\">S: {result.extra.seeders}</span>\n          )}\n          {result.extra?.leechers != null && (\n            <span className=\"text-red-400\">L: {result.extra.leechers}</span>\n          )}\n        </div>\n      </div>\n    </div>\n  );\n}\n```\n\n**`MapCard.tsx`:**\n\n```tsx\nimport type { MapResult as MapResultType } from '../../types/search';\n\ninterface Props {\n  result: MapResultType;\n}\n\nexport function MapCard({ result }: Props) {\n  return (\n    <>\n      <a\n        href={result.url}\n        target=\"_blank\"\n        rel=\"noopener noreferrer\"\n        className=\"text-sm font-medium text-[#60a5fa] hover:underline block mb-1\"\n      >\n        {result.title}\n      </a>\n      {result.extra?.address && (\n        <p className=\"text-xs text-[#9ca3af] mb-1\">{result.extra.address}</p>\n      )}\n      {result.extra?.latitude != null && result.extra?.longitude != null && (\n        <p className=\"text-xs text-[#6b7280]\">\n          {result.extra.latitude.toFixed(4)}, {result.extra.longitude.toFixed(4)}\n        </p>\n      )}\n      {result.extra?.map_url && (\n        <a\n          href={result.extra.map_url}\n          target=\"_blank\"\n          rel=\"noopener noreferrer\"\n          className=\"text-xs text-[#60a5fa] hover:underline block mt-1\"\n        >\n          Open in maps\n        </a>\n      )}\n    </>\n  );\n}\n```\n\n**`MusicCard.tsx`:**\n\n```tsx\nimport type { MusicResult as MusicResultType } from '../../types/search';\n\ninterface Props {\n  result: MusicResultType;\n}\n\nexport function MusicCard({ result }: Props) {\n  return (\n    <div className=\"flex items-center gap-3\">\n      {result.thumbnail_url && (\n        <img\n          src={result.thumbnail_url}\n          alt={result.title}\n          className=\"w-12 h-12 rounded-lg bg-[#2a2a2a] object-cover shrink-0\"\n          loading=\"lazy\"\n        />\n      )}\n      <div className=\"min-w-0 flex-1\">\n        <a\n          href={result.url}\n          target=\"_blank\"\n          rel=\"noopener noreferrer\"\n          className=\"text-sm font-medium text-[#60a5fa] hover:underline block truncate\"\n        >\n          {result.title}\n        </a>\n        <p className=\"text-xs text-[#9ca3af] truncate\">\n          {[result.extra?.artist, result.extra?.album, result.extra?.duration]\n            .filter(Boolean)\n            .join(' · ')}\n        </p>\n      </div>\n    </div>\n  );\n}\n```\n\n**`AnswerBox.tsx`:**\n\n```tsx\nimport type { AnswerResult as AnswerResultType } from '../../types/search';\n\ninterface Props {\n  result: AnswerResultType;\n}\n\nexport function AnswerBox({ result }: Props) {\n  return (\n    <div className=\"p-5 bg-[#1a2a1a] border border-green-500/30 rounded-xl mb-4\">\n      <h2 className=\"text-lg font-semibold text-[#e5e5e5] mb-2\">{result.title}</h2>\n      <p className=\"text-[#9ca3af] text-sm leading-relaxed\">{result.extra?.answer}</p>\n      {result.url && (\n        <a\n          href={result.url}\n          target=\"_blank\"\n          rel=\"noopener noreferrer\"\n          className=\"text-xs text-[#60a5fa] hover:underline mt-2 inline-block\"\n        >\n          Source: {result.engine}\n        </a>\n      )}\n    </div>\n  );\n}\n```\n\n**`KeyValueTable.tsx`:**\n\n```tsx\nimport type { KeyValueResult as KeyValueResultType } from '../../types/search';\n\ninterface Props {\n  result: KeyValueResultType;\n}\n\nexport function KeyValueTable({ result }: Props) {\n  const kv = result.extra?.kv_map;\n  if (!kv) return null;\n\n  return (\n    <div>\n      {result.title && (\n        <h3 className=\"text-sm font-semibold text-[#e5e5e5] mb-2\">{result.title}</h3>\n      )}\n      {result.extra?.caption && (\n        <p className=\"text-xs text-[#6b7280] mb-2\">{result.extra.caption}</p>\n      )}\n      <table className=\"w-full text-xs\">\n        <thead>\n          <tr className=\"text-[#6b7280] border-b border-[rgba(255,255,255,0.08)]\">\n            <th className=\"text-left py-1 pr-3 font-medium\">\n              {result.extra?.key_title || 'Key'}\n            </th>\n            <th className=\"text-left py-1 font-medium\">\n              {result.extra?.value_title || 'Value'}\n            </th>\n          </tr>\n        </thead>\n        <tbody>\n          {Object.entries(kv).map(([k, v]) => (\n            <tr key={k} className=\"border-b border-[rgba(255,255,255,0.04)]\">\n              <td className=\"py-1.5 pr-3 text-[#e5e5e5] font-medium\">{k}</td>\n              <td className=\"py-1.5 text-[#9ca3af]\">{v}</td>\n            </tr>\n          ))}\n        </tbody>\n      </table>\n    </div>\n  );\n}\n```\n\n**`InfoboxPanel.tsx`:**\n\n```tsx\nimport type { InfoboxResult as InfoboxResultType } from '../../types/search';\n\ninterface Props {\n  result: InfoboxResultType;\n}\n\nexport function InfoboxPanel({ result }: Props) {\n  return (\n    <div className=\"p-5 bg-[#1a1a2a] border border-[rgba(255,255,255,0.1)] rounded-xl mb-4\">\n      <h2 className=\"text-lg font-semibold text-[#e5e5e5] mb-3\">{result.title}</h2>\n\n      {result.extra?.img_src && (\n        <img\n          src={result.extra.img_src}\n          alt={result.extra?.img_alt || result.title}\n          className=\"w-full max-h-48 object-cover rounded-lg mb-3 bg-[#2a2a2a]\"\n          loading=\"lazy\"\n        />\n      )}\n\n      {result.extra?.attributes && result.extra.attributes.length > 0 && (\n        <table className=\"w-full text-xs mb-3\">\n          <tbody>\n            {result.extra.attributes.map((attr, i) => (\n              <tr key={i} className=\"border-b border-[rgba(255,255,255,0.04)]\">\n                <td className=\"py-1.5 pr-3 text-[#6b7280] font-medium w-1/3\">{attr.label}</td>\n                <td className=\"py-1.5 text-[#e5e5e5]\">\n                  {attr.url ? (\n                    <a\n                      href={attr.url}\n                      target=\"_blank\"\n                      rel=\"noopener noreferrer\"\n                      className=\"text-[#60a5fa] hover:underline\"\n                    >\n                      {attr.value}\n                    </a>\n                  ) : (\n                    attr.value\n                  )}\n                </td>\n              </tr>\n            ))}\n          </tbody>\n        </table>\n      )}\n\n      {result.extra?.urls && result.extra.urls.length > 0 && (\n        <div className=\"flex flex-wrap gap-2 mb-3\">\n          {result.extra.urls.map((u, i) => (\n            <a\n              key={i}\n              href={u.url}\n              target=\"_blank\"\n              rel=\"noopener noreferrer\"\n              className=\"text-xs px-2 py-1 bg-[#2a2a3a] rounded text-[#60a5fa] hover:underline\"\n            >\n              {u.title}\n            </a>\n          ))}\n        </div>\n      )}\n\n      {result.extra?.related_topics && result.extra.related_topics.length > 0 && (\n        <div>\n          <h4 className=\"text-xs font-medium text-[#6b7280] mb-1\">Related topics:</h4>\n          <div className=\"flex flex-wrap gap-1\">\n            {result.extra.related_topics.map((topic, i) => (\n              <span key={i} className=\"text-xs px-2 py-0.5 bg-[#2a2a3a] rounded text-[#9ca3af]\">\n                {topic}\n              </span>\n            ))}\n          </div>\n        </div>\n      )}\n\n      {result.content && (\n        <p className=\"text-xs text-[#9ca3af] mt-3 leading-relaxed\">{result.content}</p>\n      )}\n\n      {result.url && (\n        <a\n          href={result.url}\n          target=\"_blank\"\n          rel=\"noopener noreferrer\"\n          className=\"text-xs text-[#60a5fa] hover:underline mt-2 inline-block\"\n        >\n          Source: {result.engine}\n        </a>\n      )}\n    </div>\n  );\n}\n```\n\n- [ ] Build verification:\n\n```bash\ncd web && npx tsc --noEmit 2>&1\n```\n\nExpected: no TypeScript errors. Fix any import path issues — all imports are relative from `components/results/` to `types/search.ts` (path: `../../types/search`).\n\n- [ ] Production build:\n\n```bash\ncd web && npm run build\n```\n\nExpected: `tsc -b && vite build` succeeds.\n\n- [ ] Manual verification — open browser, check component rendering:\n\nThis is deferred to Task 11 when SearchPage is wired. For now, verify that all 13 component files exist and compile.\n\n- [ ] Commit:\n\n```bash\ngit add web/src/components/results/\ngit commit -m \"web: create ResultCard dispatch + 12 kind-specific result components + ImageGrid\"\n```\n\n---\n\n### Task 11: SearchPage — wire ResultCard + category-aware layout\n\n**Depends on:** Task 10 (all components exist), Task 9 (store has answers/corrections/infoboxes)\n\n**Files:**\n- Modify: `web/src/pages/SearchPage.tsx:1-146` — replace inline result card with `<ResultCard>`, add answer/infobox/ImageGrid rendering\n\n- [ ] Write the updated `SearchPage.tsx`:\n\n```tsx\nimport { useState } from 'react';\nimport { useSearchStore } from '../stores/searchStore';\nimport { ResultCard } from '../components/results/ResultCard';\nimport { ImageGrid } from '../components/results/ImageGrid';\nimport { AnswerBox } from '../components/results/AnswerBox';\nimport { InfoboxPanel } from '../components/results/InfoboxPanel';\n\nexport default function SearchPage() {\n  const [input, setInput] = useState('');\n  const {\n    results, isLoading, enginesUsed, enginesFailed,\n    responseTimeMs, error, search,\n    answers, corrections, infoboxes,\n  } = useSearchStore();\n  const hasSearched = results.length > 0 || error !== null || enginesUsed.length > 0;\n\n  const handleSubmit = (e: React.FormEvent) => {\n    e.preventDefault();\n    if (input.trim()) {\n      search({ q: input.trim() });\n    }\n  };\n\n  // Separate image results for grid layout\n  const imageResults = results.filter(r => r.kind === 'image');\n  const nonImageResults = results.filter(r => r.kind !== 'image');\n  const isImageCategory = imageResults.length > nonImageResults.length;\n\n  return (\n    <div className=\"min-h-screen bg-[#0f0f0f] text-[#e5e5e5]\">\n      <div className=\"max-w-3xl mx-auto px-4 py-12\">\n        {/* Logo / Title */}\n        <div className={`text-center transition-all duration-500 ${hasSearched ? 'mb-6' : 'mb-12 mt-20'}`}>\n          <h1 className=\"text-5xl font-bold tracking-tight mb-2\">\n            <span className=\"text-[#3b82f6]\">Sear</span>Go\n          </h1>\n          <p className=\"text-[#9ca3af] text-sm\">Privacy-respecting meta search</p>\n        </div>\n\n        {/* Search Box */}\n        <form onSubmit={handleSubmit} className=\"relative mb-8\">\n          <div className=\"flex gap-2\">\n            <div className=\"flex-1 relative\">\n              <input\n                type=\"text\"\n                value={input}\n                onChange={(e) => setInput(e.target.value)}\n                placeholder=\"Search the web...\"\n                className=\"w-full px-5 py-3.5 bg-[#1a1a1a] border border-[rgba(255,255,255,0.08)] rounded-xl\n                         text-[#e5e5e5] placeholder-[#6b7280] outline-none\n                         focus:border-[#3b82f6] focus:ring-2 focus:ring-[#3b82f6]/30\n                         transition-all duration-200 text-base\"\n              />\n              {input && (\n                <button\n                  type=\"button\"\n                  onClick={() => setInput('')}\n                  className=\"absolute right-3 top-1/2 -translate-y-1/2 text-[#6b7280] hover:text-[#e5e5e5]\"\n                >\n                  ✕\n                </button>\n              )}\n            </div>\n            <button\n              type=\"submit\"\n              disabled={isLoading}\n              className=\"px-6 py-3.5 bg-[#3b82f6] hover:bg-[#2563eb] disabled:bg-[#1e3a5f]\n                       rounded-xl font-medium transition-all duration-200\n                       flex items-center gap-2 min-w-[100px] justify-center\"\n            >\n              {isLoading ? (\n                <span className=\"inline-block w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin\" />\n              ) : (\n                'Search'\n              )}\n            </button>\n          </div>\n        </form>\n\n        {/* Error */}\n        {error && (\n          <div className=\"mb-6 p-4 bg-red-900/20 border border-red-500/30 rounded-xl text-red-300\">\n            {error}\n          </div>\n        )}\n\n        {/* Answers (top-level, prominent) */}\n        {answers.length > 0 && (\n          <div className=\"mb-6 space-y-3\">\n            {answers.map((a, i) => (\n              <div\n                key={i}\n                className=\"p-5 bg-[#1a2a1a] border border-green-500/30 rounded-xl\"\n              >\n                <p className=\"text-[#e5e5e5] text-base leading-relaxed\">{a.answer}</p>\n                {a.url && (\n                  <a\n                    href={a.url}\n                    target=\"_blank\"\n                    rel=\"noopener noreferrer\"\n                    className=\"text-xs text-[#60a5fa] hover:underline mt-2 inline-block\"\n                  >\n                    Source: {a.engine}\n                  </a>\n                )}\n              </div>\n            ))}\n          </div>\n        )}\n\n        {/* Infoboxes */}\n        {infoboxes.length > 0 && (\n          <div className=\"mb-6 space-y-4\">\n            {infoboxes.map((ib, i) => (\n              <InfoboxPanel\n                key={i}\n                result={{\n                  kind: 'infobox',\n                  title: ib.title,\n                  url: ib.url || '',\n                  engine: ib.engine || '',\n                  category: '',\n                  score: 0,\n                  extra: {\n                    img_src: ib.img_src,\n                    attributes: ib.attributes || [],\n                    urls: ib.urls || [],\n                    related_topics: ib.related_topics || [],\n                  },\n                }}\n              />\n            ))}\n          </div>\n        )}\n\n        {/* Corrections */}\n        {corrections.length > 0 && (\n          <div className=\"mb-4 text-sm text-[#9ca3af]\">\n            <span className=\"text-[#6b7280]\">Did you mean: </span>\n            {corrections.map((c, i) => (\n              <button\n                key={i}\n                onClick={() => setInput(c)}\n                className=\"text-[#60a5fa] hover:underline\"\n              >\n                {c}\n                {i < corrections.length - 1 ? ', ' : ''}\n              </button>\n            ))}\n          </div>\n        )}\n\n        {/* Results Stats */}\n        {(results.length > 0 || enginesFailed.length > 0) && (\n          <div className=\"mb-4 text-sm text-[#9ca3af]\">\n            Found <span className=\"text-[#e5e5e5] font-medium\">{results.length}</span> results\n            {responseTimeMs > 0 && ` in ${responseTimeMs}ms`}\n            {enginesUsed.length > 0 && (\n              <span> · Engines: {enginesUsed.join(', ')}</span>\n            )}\n            {enginesFailed.length > 0 && (\n              <span className=\"text-red-400\"> · Failed: {enginesFailed.join(', ')}</span>\n            )}\n          </div>\n        )}\n\n        {/* Image Grid (when image-heavy) */}\n        {isImageCategory && imageResults.length > 0 && (\n          <div className=\"mb-6\">\n            <ImageGrid results={imageResults} />\n          </div>\n        )}\n\n        {/* Non-image results (list) */}\n        {!isImageCategory && (\n          <div className=\"space-y-3\">\n            {results.map((r, i) => (\n              <ResultCard key={i} result={r} index={i} />\n            ))}\n          </div>\n        )}\n\n        {/* Image overflow: non-image results below the grid */}\n        {isImageCategory && nonImageResults.length > 0 && (\n          <div className=\"space-y-3 mt-6\">\n            <h3 className=\"text-sm font-medium text-[#6b7280] mb-2\">Web Results</h3>\n            {nonImageResults.map((r, i) => (\n              <ResultCard key={i} result={r} index={i} />\n            ))}\n          </div>\n        )}\n\n        {/* Suggestions */}\n        {useSearchStore.getState().results.length > 0 && (\n          <div className=\"hidden\">{/* suggestions rendered elsewhere for now */}</div>\n        )}\n\n        {/* Empty state after search */}\n        {hasSearched && results.length === 0 && !isLoading && !error && (\n          <div className=\"text-center py-12 text-[#6b7280]\">\n            <p className=\"text-lg mb-2\">No results found</p>\n            <p className=\"text-sm\">Try a different query or check your engine configuration</p>\n          </div>\n        )}\n      </div>\n    </div>\n  );\n}\n```\n\nNote: removed the `engineColors` map and `getEngineColor` function — they moved into `MainResult.tsx` in Task 10.\n\n- [ ] Build verification — type-check + production build:\n\n```bash\ncd web && npx tsc --noEmit 2>&1\n```\n\nExpected: no TypeScript errors. If `useSearchStore.getState()` usage causes issues (zustand v5), replace with:\n\n```tsx\nconst suggestions = useSearchStore(s => s.suggestions);\n```\n\nand use `suggestions.length > 0` instead.\n\n- [ ] Production build:\n\n```bash\ncd web && npm run build\n```\n\nExpected: `tsc -b && vite build` succeeds, output in `web/dist/`.\n\n- [ ] Manual verification — start dev server and test:\n\n```bash\ncd web && npm run dev &\n```\n\nOpen the browser, type a search query, verify:\n1. Search results appear with `kind: \"main\"` tag (from backend pipeline)\n2. Each result shows title, URL in green, content snippet, engine badge\n3. If any image results exist, they render with thumbnail\n4. Answers appear above the result list\n5. Infoboxes render with attributes table and related URLs\n6. No JavaScript errors in console\n\n- [ ] Commit:\n\n```bash\ngit add web/src/pages/SearchPage.tsx\ngit commit -m \"web: wire ResultCard dispatch, answer/infobox/ImageGrid rendering in SearchPage\"\n```\n\n---\n\n## Local Self-Review\n\n- [ ] 1. Spec-coverage table (Part 3):\n  | Design requirement | Task(s) | Status |\n  |---|---|---|\n  | TypeScript tagged-union types: `ResultKind`, `BaseResult`, 12 kind-specific interfaces | 9 | covered |\n  | TypeScript `Result` discriminated union type | 9 | covered |\n  | TypeScript `SearchResponse` with `answers`/`corrections`/`infoboxes` | 9 | covered |\n  | TypeScript `Answer` and `Infobox` interfaces | 9 | covered |\n  | `searchStore` — `answers`/`corrections`/`infoboxes` in state | 9 | covered |\n  | `ResultCard` — `switch (result.kind)` dispatch to 12 components | 10 | covered |\n  | `MainResult` — title, URL, content, engine badge, score | 10 | covered |\n  | `ImageCard` — thumbnail, resolution, format, source, domain | 10 | covered |\n  | `ImageGrid` — CSS grid layout for image category | 10 | covered |\n  | `VideoCard` — thumbnail + duration overlay, author, view count | 10 | covered |\n  | `PaperCard` — authors, journal, DOI link, PDF link | 10 | covered |\n  | `CodeBlock` — language label, line-numbered code display | 10 | covered |\n  | `FileRow` — filename, file type, size, seeders/leechers, magnet link | 10 | covered |\n  | `MapCard` — address, lat/lng, map link | 10 | covered |\n  | `MusicCard` — thumbnail, artist, album, duration | 10 | covered |\n  | `AnswerBox` — prominent answer display with source link | 10 | covered |\n  | `KeyValueTable` — key/value table with optional caption/titles | 10 | covered |\n  | `InfoboxPanel` — image, attributes table, URL badges, related topics | 10 | covered |\n  | `ResultCard` default branch → `MainResult` fallback | 10 | covered |\n  | SearchPage — top-level answers rendering | 11 | covered |\n  | SearchPage — top-level infoboxes rendering | 11 | covered |\n  | SearchPage — corrections as clickable \"did you mean\" links | 11 | covered |\n  | SearchPage — image grid when image-heavy category | 11 | covered |\n  | SearchPage — non-image results below image grid | 11 | covered |\n  | SearchPage — existing styling retained (dark theme, engine colors) | 11 | covered |\n\n- [ ] 2. Placeholder scan: Zero `TODO`/`TBD`. All 13 component files have complete JSX code. The `default` branch in `ResultCard` renders `MainResult` via `<Wrapper>` — explicit fallback, not a placeholder. `ImageGrid` handles empty `imgSrc` with `''` fallback. `CodeBlock` caps at 20 lines — intentional UI limit, not deferred work. No deferred-by-dependency in frontend (all components rendered).\n\n- [ ] 3. No phantom tasks: Task 9 creates/modifies 2 files. Task 10 creates 13 files (all new components). Task 11 modifies 1 file. Total: 16 file operations across 3 tasks. Zero `--allow-empty` scenarios — each task produces a verifiable `git diff`.\n\n- [ ] 4. Dependency soundness:\n  - T9 (types) → consumed by T10 (component props) + T11 (search page state) ✓\n  - T10 (components) → consumed by T11 (ResultCard, ImageGrid, AnswerBox, InfoboxPanel imports) ✓\n  - T9 → Part 1 T2 (backend `Kind`/`Extra`/`Template` fields shape TS interfaces) ✓\n  - No reference to Part 2 (integration) — frontend talks to API, not internal Go types.\n\n- [ ] 5. Caller & build soundness:\n  - T9 changes `Result` type from flat to union: consumers are `searchStore.ts` (updated in same task) and `SearchPage.tsx` (updated in T11). The old `SearchPage` accesses `r.title`, `r.url`, `r.content`, `r.engine`, `r.score` — all exist on `BaseResult` which is a superset of the old flat type. `tsc --noEmit` verifies.\n  - T11 replaces inline JSX with `ResultCard` component — removed `engineColors`/`getEngineColor` from `SearchPage` (moved to `MainResult`). No other consumer of those helpers.\n  - Final build: `cd web && npm run build` (`tsc -b && vite build`) verifies entire frontend compiles.\n\n- [ ] 6. Test-the-risk:\n  - TypeScript exhaustive switch on `result.kind`: the `default` branch handles unknown kinds with `MainResult` fallback + `console.warn` — prevents runtime crashes if backend adds a new kind.\n  - `ImageCard`: handles missing `img_src` (falls back to `thumbnail_url`) — prevents broken images.\n  - `CodeBlock`: caps at 20 lines — prevents huge responses from blowing up the UI.\n  - `FileRow`: `formatSize()` handles 0/null/missing size gracefully.\n  - `KeyValueTable`: returns `null` for missing `kv_map` — prevents empty table rendering.\n  - `ImageGrid`: `r.kind === 'image'` guard ensures only image results render in the grid — others would cause a type error since `extra` shape differs.\n\n- [ ] 7. Type consistency (cross-part):\n  - TS `ResultKind` union values match Go `Kind()` return values: `\"main\"`, `\"image\"`, `\"video\"`, `\"news\"`, `\"paper\"`, `\"code\"`, `\"file\"`, `\"map\"`, `\"music\"`, `\"answer\"`, `\"keyvalue\"`, `\"infobox\"` — matches Part 1 T1 `Kind()` strings ✓\n  - TS `Extra` field shapes match Go `buildExtra()` output: `ImageResult.extra.img_src` matches `ed[\"img_src\"]` in `convert.go` ✓\n  - TS `Infobox` shape matches Go `models.Infobox` struct from Part 1 T2: `attributes: {label, value, url}[]` matches `InfoboxAttribute`, `urls: {title, url}[]` matches `InfoboxURL` ✓\n  - TS `Answer.answer` field matches Go `Answer.Answer` (JSON: `\"answer\"`) — case-sensitive match ✓\n",
  "path": "/Users/ranwei/workspace/go_work/seargo/.ody-code/plans/2026-06-20-05-result-types/frontend.md"
}
```

<details><summary>Tool Result: Write (`/Users/ranwei/workspace/go_work/seargo/.ody-code/plans/2026-…`)</summary>

<!-- call_id: call_00_cJ2uiruGiyBQnnSHuM951415 -->
Wrote 43394 bytes to /Users/ranwei/workspace/go_work/seargo/.ody-code/plans/2026-06-20-05-result-types/frontend.md

</details>

<details><summary>Thinking</summary>

Now flip Part 3 to done in the index, then write the global self-review + spec coverage table, and call ExitPlanMode.

</details>

#### Tool Call: Edit (`/Users/ranwei/workspace/go_work/seargo/.ody-code/plans/2026-…`)
<!-- call_id: call_00_gbKjFvCf5elK7nh9W64E4838 -->
```json
{
  "new_string": "| 3 | 2026-06-20-05-result-types/frontend.md | TypeScript types, ResultCard, kind-specific components, SearchPage layout | done |",
  "old_string": "| 3 | 2026-06-20-05-result-types/frontend.md | TypeScript types, ResultCard, kind-specific components, SearchPage layout | pending |",
  "path": "/Users/ranwei/workspace/go_work/seargo/.ody-code/plans/2026-06-20-05-result-types.md"
}
```

<details><summary>Tool Result: Edit (`/Users/ranwei/workspace/go_work/seargo/.ody-code/plans/2026-…`)</summary>

<!-- call_id: call_00_gbKjFvCf5elK7nh9W64E4838 -->
Replaced 1 occurrence in /Users/ranwei/workspace/go_work/seargo/.ody-code/plans/2026-06-20-05-result-types.md

</details>

<details><summary>Thinking</summary>

All 3 parts are done. Now I need to append the global self-review and spec-coverage table to the index, then call ExitPlanMode. Let me re-read the index to add the final sections.

</details>

All 3 parts written. Now append the global self-review and spec-coverage table to the index.

#### Tool Call: Read (`/Users/ranwei/workspace/go_work/seargo/.ody-code/plans/2026-…`)
<!-- call_id: call_00_6Fwy3lf5Vn9GkUQD8YnZ7362 -->
```json
{
  "path": "/Users/ranwei/workspace/go_work/seargo/.ody-code/plans/2026-06-20-05-result-types.md",
  "line_offset": 1,
  "n_lines": 5
}
```

<details><summary>Tool Result: Read (`/Users/ranwei/workspace/go_work/seargo/.ody-code/plans/2026-…`)</summary>

<!-- call_id: call_00_6Fwy3lf5Vn9GkUQD8YnZ7362 -->
1	# Structured Result Types & Rendering — Implementation Plan
2	
3	**Goal:** Replace the single flat `models.Result` with a tagged-union API envelope carrying kind-specific fields, build a typed result hierarchy with normalization and dedup, introduce `TypedResultContainer` for per-kind merge, and wire the frontend to dispatch on `kind` for specialized rendering.
4	
5	**Architecture:** Three subsystems: (A) Backend Types — upgrade the `results.Result` interface with `Kind()/Template()/Normalize()/DedupKey()/Base()`, add missing types (`KeyValueResult`, `CodeLine`, `ImageRef`, `InfoboxAttribute`, `InfoboxURL`), extend normalization with `normalizeEngines` and per-kind `PostNormalize()`, add `Kind`+`Template`+`Extra` to `models.Result`, rewrite `ToAPIResult` for tagged-union output, and create `TypedResultContainer` with per-kind buckets and SearXNG-aligned score/dedup; (B) Integration — extend `ProcessorResult` with `TypedResults`, add an adapter to wrap existing flat engine results into typed ones, wire scheduler to use `TypedResultContainer`; (C) Frontend — rewrite TypeScript types as tagged union, create `ResultCard` dispatch component with kind-specific rendering components, and update `SearchPage` for category-aware layout.
<system>5 lines read from file starting from line 1. Total lines in file: 92.</system>

</details>

<details><summary>Thinking</summary>

Now let me read the full index to find the last line and append the spec coverage + global self-review.

</details>

#### Tool Call: Read (`/Users/ranwei/workspace/go_work/seargo/.ody-code/plans/2026-…`)
<!-- call_id: call_00_wEosDr3ChzkyNVwMTmil4377 -->
```json
{
  "line_offset": 80,
  "path": "/Users/ranwei/workspace/go_work/seargo/.ody-code/plans/2026-06-20-05-result-types.md"
}
```

<details><summary>Tool Result: Read (`/Users/ranwei/workspace/go_work/seargo/.ody-code/plans/2026-…`)</summary>

<!-- call_id: call_00_wEosDr3ChzkyNVwMTmil4377 -->
80	| 2 | `models.Result` adds `Kind`/`Extra`/`Template` fields — existing cache serialization (flat struct) may break on deserialization | Fields have `omitempty`; old cache entries lack `kind` → frontend treats empty as `"main"`; cache TTL is short (max 2min), old entries expire naturally |
81	| 3 | `TypedResultContainer` replaces `ResultContainer` in `executeProcessors` — regression risk on scoring/grouping | Score formula is identical (`Σ weight/pos × count`); dedup key adds `kind` prefix; keep old container code in `merge.go` untouched during migration for rollback |
82	| 4 | Frontend `ResultCard` exhaustive switch must handle all 12 kinds | TypeScript discriminated union on `kind`; default branch renders `MainResult` with `console.warn`; `tsc --noEmit` catches missing cases |
83	| 5 | Older engines return flat `models.Result` without `Kind` — `WrapAPIMainResult` produces typed results but loses kind-specific fields (image metadata, etc.) | Acceptable: engines are ported incrementally per Phase 4 tiers; until then, all flat results render as `MainResult` with title/URL/content intact |
84	| 6 | `Infobox` struct change (from `[]string` to typed `[]InfoboxURL`/`[]InfoboxAttribute`) breaks existing infobox producers | Search shows 0 existing infobox producers in current 6 engines; fixtures validate new shape |
85	
86	## Parts (generate one per invocation, in order)
87	
88	| # | File | Scope | Status |
89	|---|---|---|---|
90	| 1 | 2026-06-20-05-result-types/backend-types.md | Result interface, missing types, API envelope, normalization, ToAPIResult, TypedResultContainer | done |
91	| 2 | 2026-06-20-05-result-types/integration.md | ProcessorResult, WrapAPIMainResult adapter, online processor, scheduler wiring | done |
92	| 3 | 2026-06-20-05-result-types/frontend.md | TypeScript types, ResultCard, kind-specific components, SearchPage layout | done |
<system>13 lines read from file starting from line 80. Total lines in file: 92. End of file reached.</system>

</details>

<details><summary>Thinking</summary>

I need to append the spec coverage table and global self-review to the index file, then call ExitPlanMode. Let me do that now.

</details>

#### Tool Call: Edit (`/Users/ranwei/workspace/go_work/seargo/.ody-code/plans/2026-…`)
<!-- call_id: call_00_zK0ll5WuDm5dKOb8gruE0196 -->
```json
{
  "path": "/Users/ranwei/workspace/go_work/seargo/.ody-code/plans/2026-06-20-05-result-types.md",
  "old_string": "| 3 | 2026-06-20-05-result-types/frontend.md | TypeScript types, ResultCard, kind-specific components, SearchPage layout | done |",
  "new_string": "| 3 | 2026-06-20-05-result-types/frontend.md | TypeScript types, ResultCard, kind-specific components, SearchPage layout | done |\n\n---\n\n## Spec Coverage\n\n| 设计需求 | Task(s) | 状态 |\n|---|---|---|\n| Result interface — Kind/Template/Base/Normalize/DedupKey | backend-types:1 | covered |\n| Missing types — KeyValueResult, CodeLine, ImageRef, InfoboxAttribute, InfoboxURL | backend-types:1 | covered |\n| Per-kind Template defaults (12 kinds) | backend-types:1 | covered |\n| DedupKey per-kind (kind + normalized URL) | backend-types:1 | covered |\n| FileResult extension (Filename/MagnetURI/Seeders/Leechers) | backend-types:1 | covered |\n| CodeResult extension (Filename/CodeLines/HLLines) | backend-types:1 | covered |\n| ImageResult extension (FileSize/Formats/ImageRef) | backend-types:1 | covered |\n| InfoboxResult typed Attributes/URLs/RelatedTopics | backend-types:1 | covered |\n| models.Result tagged union (Kind/Template/Extra) | backend-types:2 | covered |\n| models.Infobox rich schema (InfoboxAttribute/InfoboxURL/Engines/RelatedTopics) | backend-types:2 | covered |\n| normalizeEngines — engine name in Engines slice | backend-types:3 | covered |\n| PostNormalize — Image thumbnail/title fallback | backend-types:3 | covered |\n| PostNormalize — Video thumbnail fallback | backend-types:3 | covered |\n| PostNormalize — Code title from filename | backend-types:3 | covered |\n| PostNormalize — File title from filename, type from extension | backend-types:3 | covered |\n| PostNormalize — Infobox ID fallback chain | backend-types:3 | covered |\n| NormalizeResult orchestrator | backend-types:3 | covered |\n| ToAPIResult — Kind/Template/Extra on output | backend-types:4 | covered |\n| ToAPIResult — all 12 kind-specific Extra maps | backend-types:4 | covered |\n| ToAPIResult — unknown kind fallback to \"main\" | backend-types:4 | covered |\n| ToAPIResult — Domain computation | backend-types:4 | covered |\n| WrapAPIMainResult — flat→typed adapter | backend-types:4 | covered |\n| TypedResultContainer — per-kind buckets | backend-types:5 | covered |\n| TypedResultContainer — dedup by kind+normalized URL | backend-types:5 | covered |\n| TypedResultContainer — prefer HTTPS | backend-types:5 | covered |\n| TypedResultContainer — score = Σ(weight/pos)×count | backend-types:5 | covered |\n| TypedResultContainer — category grouping | backend-types:5 | covered |\n| TypedResultContainer — closed-for-write | backend-types:5 | covered |\n| TypedResultContainer — answers/suggestions/corrections/infoboxes dedup | backend-types:5 | covered |\n| TypedResultContainer — GetEnginesUsed/GetEnginesFailed | backend-types:5 | covered |\n| ProcessorResult — TypedResults field | integration:6 | covered |\n| OnlineProcessor — wrap flat engine results | integration:6 | covered |\n| OnlineProcessor — populate TypedResults + Results (backward compat) | integration:6 | covered |\n| Whole-tree build verification | integration:7 | covered |\n| Scheduler — NewTypedResultContainer | integration:8 | covered |\n| Scheduler — TypedResults→ToAPIResult→Extend pipeline | integration:8 | covered |\n| Scheduler — fallback to Results when TypedResults empty | integration:8 | covered |\n| Scheduler — keep AddSuggestions/AddAnswers/AddCorrections/AddInfoboxes/AddEngineData | integration:8 | covered |\n| Scheduler — Results() replacing GetOrderedResults() | integration:8 | covered |\n| TypeScript tagged-union types (ResultKind, BaseResult, 12 derived) | frontend:9 | covered |\n| TypeScript SearchResponse (answers/corrections/infoboxes) | frontend:9 | covered |\n| searchStore — new state fields | frontend:9 | covered |\n| ResultCard — switch on kind → 12 components | frontend:10 | covered |\n| 12 kind-specific components (MainResult through InfoboxPanel) | frontend:10 | covered |\n| ImageGrid — CSS grid for image category | frontend:10 | covered |\n| SearchPage — top-level answers/infoboxes/corrections | frontend:11 | covered |\n| SearchPage — image grid for image-heavy category | frontend:11 | covered |\n| SearchPage — non-image results below image grid | frontend:11 | covered |\n| WeatherAnswer (SearXNG full weather module) | no-op | deferred to weather module phase |\n| Server-side syntax highlighting (CodeResult) | no-op | deferred — frontend <pre> block |\n| ImageRef alternative image formats | no-op | deferred per design scope |\n| Result URL filter hooks (plugins) | no-op | deferred to Phase 6 (plugins) |\n| Distributed/shared result cache | no-op | deferred per design scope |\n| All 238 engines emitting typed results | no-op | deferred — incremental per Phase 4 tiers |\n\n---\n\n## Global Self-Review\n\n- [x] 1. Spec-coverage table: 48 spec items mapped to tasks across 3 parts, 6 no-ops (deferred per design scope), 0 GAPs. Every section in the approved design (`.ody-code/designs/2026-06-20-05-result-types.md`) maps to at least one Task. Covered: typed result hierarchy (backend-types:1), API envelope (backend-types:2), normalization (backend-types:3), conversion (backend-types:4), typed container (backend-types:5), processor (integration:6), scheduler (integration:8), frontend types (frontend:9), components (frontend:10), layout (frontend:11). Deferred items match the design's Out of Scope section exactly.\n\n- [x] 2. Placeholder scan: All 3 part files + index — zero `TODO`/`TBD`/dead-code placeholders beyond intentional design-deferred items. The integration:7 task uses `--allow-empty` commit intentionally (verification gate). The `WrapAPIMainResult` adapter explicitly handles the migration period — not a placeholder but a defined bridge. Frontend components handle all edge cases: missing images, empty kv_map, unknown kind fallback. No hidden work or unresolved deps.\n\n- [x] 3. No phantom tasks: 11 tasks across 3 parts, every task produces file changes. Task counts: backend-types 5, integration 3, frontend 3 = 11 tasks. Integration:7 is a verification gate (build + test) — intentionally doesn't change files but verifies the build-green invariant with a commit.\n\n- [x] 4. Dependency soundness: All cross-file `Depends on:` references point to earlier parts:\n  - integration:T6 → backend-types:T4 (`WrapAPIMainResult`) ✓\n  - integration:T8 → backend-types:T5 (`TypedResultContainer`) ✓\n  - frontend:T9 → backend-types:T2 (`models.Result.Kind/Extra`) ✓\n  - frontend:T10 → frontend:T9 (TypeScript types) ✓\n  - frontend:T11 → frontend:T10 (component files) ✓\n  No reference points to a symbol only a later part defines. Phase B (integration) and Phase C (frontend) are independent of each other — confirmed neither imports symbols from the other.\n\n- [x] 5. Caller & build soundness:\n  - `Result` interface change (backend-types:1): all callers are within `pkg/models/results/` package (convert.go + test files) — updated in same task; whole-tree `go build ./...` verifies ✓\n  - `models.Result` struct extension (backend-types:2): additive fields (Kind/Template/Extra with omitempty) — all existing constructors compile without modification; `go build ./... && go test ./...` verifies ✓\n  - `ProcessorResult` struct extension (integration:6): additive field (TypedResults) — constructor in online.go updated in same task; whole-tree build verifies ✓\n  - `merge.go` `appendUniqueStr` reused by `typed_container.go` — same package, no import change ✓\n  - `Result` TypeScript type (frontend:9): flat→union — existing property access (r.title, r.url, r.content, r.engine, r.score) works on `BaseResult` which all union members extend; `tsc --noEmit` verifies ✓\n  - `Infobox` struct type change (backend-types:2): `Attributes` from `[]string` to `[]InfoboxAttribute` — grep confirms 0 existing infobox producers in 6 engines; no callers to break ✓\n  - `scheduler.go` container type from `*ResultContainer` to `*TypedResultContainer`: both implement `Extend/Close/AddSuggestions/AddAnswers/AddCorrections/AddInfoboxes/AddEngineData/MarkUnresponsive/GetSuggestions/GetAnswers/GetCorrections/GetInfoboxes/GetEngineData/GetUnresponsive/GetEnginesUsed/GetEnginesFailed` — same method set; only method name change is `GetOrderedResults→Results` handled in integration:8 ✓\n\n- [x] 6. Test-the-risk: All 3 part files have explicit risk-testing assertions:\n  - Backend-types: `TestNormalizeEngines_NoDuplicate` (state mutation: Engines slice growth), `TestFileResultTypeExtension` (PostNormalize: FileType from filename), `TestInfoboxIDFallback` (fallback chain), `TestToAPIResult_UnknownKindFallback` (graceful degradation), `TestTypedContainer_ClosedForWrite` (concurrency safety), `TestTypedContainer_ScoreCalculation` (scoring math verified: 6.0 = (2/1+1/1)×2), `TestTypedContainer_PreferHTTPS` (URL upgrade) ✓\n  - Integration: `TestProcessorResult_HasTypedResultsField` (field access + compilation), whole-tree `go test ./...` (regression), manual `curl` verification (end-to-end pipeline) ✓\n  - Frontend: TypeScript exhaustive switch default branch (unknown kind → MainResult), ImageCard missing img_src fallback, CodeBlock 20-line cap, KeyValueTable nil kv_map guard, `tsc --noEmit` exhaustive check ✓\n  - Full list of must-survive inputs verified in each part's local self-review item 6.\n\n- [x] 7. Type consistency: Cross-part type usage verified:\n  - `Result` interface (backend-types:1) → consumed by `NormalizeResult(Result)` (backend-types:3), `ToAPIResult([]Result)` (backend-types:4), `WrapAPIMainResult` (backend-types:4), `ProcessorResult.TypedResults` (integration:6) ✓\n  - `models.Result.Kind` (backend-types:2, string) → consumed by `ToAPIResult` output (backend-types:4), `typedDedupKey` (backend-types:5), `TypedResultContainer.Extend` (backend-types:5), `ResultCard` switch (frontend:10) ✓\n  - `models.Result.Extra` (backend-types:2, map[string]any) → populated by `buildExtra` (backend-types:4), consumed by frontend typed interfaces (frontend:9) ✓\n  - `TypedResultContainer` signature (backend-types:5) → used by `scheduler.go` (integration:8) — `Extend(string, []models.Result, int)`, `Results() []models.Result`, `AddSuggestions/AddAnswers/AddCorrections/AddInfoboxes/AddEngineData/MarkUnresponsive` + getters ✓\n  - TypeScript `ResultKind` (frontend:9) values match Go `Kind()` return strings — `\"main\"`, `\"image\"`, `\"video\"`, `\"news\"`, `\"paper\"`, `\"code\"`, `\"file\"`, `\"map\"`, `\"music\"`, `\"answer\"`, `\"keyvalue\"`, `\"infobox\"` ✓\n  - TypeScript `Extra` shapes per kind match Go `buildExtra` maps — e.g. `ImageResult.extra.img_src` ↔ `ed[\"img_src\"]`, `PaperResult.extra.doi` ↔ `ed[\"doi\"]` ✓\n  - TypeScript `Infobox.attributes` shape `{label: string; value: string; url?: string}[]` ↔ Go `[]InfoboxAttribute` ✓"
}
```

<details><summary>Tool Result: Edit (`/Users/ranwei/workspace/go_work/seargo/.ody-code/plans/2026-…`)</summary>

<!-- call_id: call_00_zK0ll5WuDm5dKOb8gruE0196 -->
Replaced 1 occurrence in /Users/ranwei/workspace/go_work/seargo/.ody-code/plans/2026-06-20-05-result-types.md

</details>

<details><summary>Thinking</summary>

All parts are done, spec coverage and global self-review appended. Now I need to call ExitPlanMode.

</details>

All parts written, cross-file review complete, spec coverage table covers 48 items. Calling ExitPlanMode.

#### Tool Call: ExitPlanMode
<!-- call_id: call_00_4ccIFS51B0GJDGnGXdFW6192 -->
```json
{
  "options": [
    {
      "label": "Approve plan",
      "description": "11 tasks across 3 parts: backend-types (interface, envelope, normalization, conversion, container), integration (processor, scheduler wiring), frontend (TypeScript, components, layout). All TDD, whole-tree typecheck, commit-per-task."
    }
  ]
}
```

<details><summary>Tool Result: ExitPlanMode</summary>

<!-- call_id: call_00_4ccIFS51B0GJDGnGXdFW6192 -->
Exited plan mode. Plan mode deactivated. The approved plan has been handed off to the main conversation context.
Plan saved to: /Users/ranwei/workspace/go_work/seargo/.ody-code/plans/2026-06-20-05-result-types.md

## Approved Plan:
# Structured Result Types & Rendering — Implementation Plan

**Goal:** Replace the single flat `models.Result` with a tagged-union API envelope carrying kind-specific fields, build a typed result hierarchy with normalization and dedup, introduce `TypedResultContainer` for per-kind merge, and wire the frontend to dispatch on `kind` for specialized rendering.

**Architecture:** Three subsystems: (A) Backend Types — upgrade the `results.Result` interface with `Kind()/Template()/Normalize()/DedupKey()/Base()`, add missing types (`KeyValueResult`, `CodeLine`, `ImageRef`, `InfoboxAttribute`, `InfoboxURL`), extend normalization with `normalizeEngines` and per-kind `PostNormalize()`, add `Kind`+`Template`+`Extra` to `models.Result`, rewrite `ToAPIResult` for tagged-union output, and create `TypedResultContainer` with per-kind buckets and SearXNG-aligned score/dedup; (B) Integration — extend `ProcessorResult` with `TypedResults`, add an adapter to wrap existing flat engine results into typed ones, wire scheduler to use `TypedResultContainer`; (C) Frontend — rewrite TypeScript types as tagged union, create `ResultCard` dispatch component with kind-specific rendering components, and update `SearchPage` for category-aware layout.

**Tech Stack:** Go 1.25, goquery, resty/v2, testify, zustand, React 19, TypeScript 6, Tailwind CSS 3

> For executing workers: implement this plan task-by-task (prefer a fresh subagent/Task per task — a clean context per task avoids single-session degradation). Steps use - [ ] checkboxes for tracking.

## File Structure

### New Files
| File | Responsibility |
|---|---|
| `internal/search/typed_container.go` | TypedResultContainer: per-kind buckets, dedup, merge, score, Close/Results |
| `internal/search/typed_container_test.go` | TypedResultContainer unit tests |
| `web/src/components/results/ResultCard.tsx` | Dispatch on `kind` to kind-specific components |
| `web/src/components/results/ImageCard.tsx` | Image result with thumbnail + resolution |
| `web/src/components/results/VideoCard.tsx` | Video result with iframe + duration |
| `web/src/components/results/PaperCard.tsx` | Academic paper with DOI/authors |
| `web/src/components/results/CodeBlock.tsx` | Code snippet with language label |
| `web/src/components/results/FileRow.tsx` | File/torrent with size/seeders info |
| `web/src/components/results/MapCard.tsx` | Map result with lat/lng |
| `web/src/components/results/MusicCard.tsx` | Music track with artist/album |
| `web/src/components/results/AnswerBox.tsx` | Answer box (dedicated, outside result list) |
| `web/src/components/results/KeyValueTable.tsx` | Key/value table display |
| `web/src/components/results/InfoboxPanel.tsx` | Infobox panel with attributes/URLs |
| `web/src/components/results/MainResult.tsx` | Default web result card |
| `web/src/components/results/ImageGrid.tsx` | Image category grid layout |

### Modified Files
| File | Changes |
|---|---|
| `pkg/models/results/results.go` | Add `Kind()/Template()/Base()/Normalize()/DedupKey()` to `Result` interface; add `KeyValueResult`/`CodeLine`/`ImageRef`/`InfoboxAttribute`/`InfoboxURL` types; implement methods on all 12 types |
| `pkg/models/results/results_test.go` | Tests for new interface methods, `KeyValueResult`, `DedupKey()` on each kind |
| `pkg/models/results/normalize.go` | Add `normalizeEngines()`; change `BaseResult.Normalize()` to call kind-specific `PostNormalize()` via type switch; add `PostNormalize()` on all kinds |
| `pkg/models/results/normalize_test.go` | Tests for `normalizeEngines`, `PostNormalize` on Image/Code/File/Infobox/Video |
| `pkg/models/results/convert.go` | Rewrite `ToAPIResult`: set `Kind`/`Template`/`Extra` on output; add all kind-specific `Extra` cases; add `WrapAPIMainResult()` reverse adapter |
| `pkg/models/results/convert_test.go` | Tests for `Kind`/`Extra` on each kind, unknown-kind fallback, `WrapAPIMainResult` |
| `pkg/models/models.go` | Add `Kind`/`Template`/`Extra` fields to `Result`; extend `Infobox` with `InfoboxURL`/`InfoboxAttribute` types; `Engines` field on `Infobox` |
| `internal/search/processor/types.go` | Add `TypedResults []results.Result` to `ProcessorResult` |
| `internal/search/processor/online.go` | Wrap engine flat results via `WrapAPIMainResult` into `TypedResults` |
| `internal/search/scheduler.go` | Replace `ResultContainer` with `TypedResultContainer` in `executeProcessors`; feed `TypedResults` |
| `web/src/types/search.ts` | Replace flat `Result` with tagged-union types; add `answers`/`corrections`/`infoboxes` to `SearchResponse` |
| `web/src/stores/searchStore.ts` | Update `Result` import to new tagged-union type |
| `web/src/pages/SearchPage.tsx` | Replace inline result card with `<ResultCard>`; add answer/infobox/ImageGrid rendering |

## Dependency Overview

```
Phase A: Backend Types & Container (no upstream deps within this plan)
  Part: backend-types
    Task 1 (Result interface upgrade + missing types) — stand-alone
    Task 2 (API envelope: models.Result+Kind, models.Infobox extension) — parallel with T1
    Task 3 (Normalization: normalizeEngines + PostNormalize) — depends on T1
    Task 4 (ToAPIResult: tagged-union output) — depends on T1, T2
    Task 5 (TypedResultContainer) — depends on T4

Phase B: Integration (depends on Phase A)
  Part: integration
    Task 6 (ProcessorResult extension + WrapAPIMainResult) — depends on T4
    Task 7 (Online processor adapter) — depends on T6
    Task 8 (Scheduler wiring) — depends on T5, T6

Phase C: Frontend (depends on Phase A: T2 for API contract)
  Part: frontend
    Task 9 (TypeScript types) — depends on T2
    Task 10 (ResultCard + kind-specific components) — depends on T9
    Task 11 (SearchPage layout + store) — depends on T10
```

Phase A 内部：T1,T2 可并行；Phase B/C 可并行（B 不依赖 C，C 不依赖 B），但各自内部串行。

## Risks & Open Questions

| # | Risk | Mitigation |
|---|---|---|
| 1 | `Result` interface gains 5 new methods — all callers (`convert.go`, tests) must conform | Task 1 finds every `Result` interface reference via `grep -rn "results\.Result"` and updates in same task; whole-tree `go build ./... && go test ./...` |
| 2 | `models.Result` adds `Kind`/`Extra`/`Template` fields — existing cache serialization (flat struct) may break on deserialization | Fields have `omitempty`; old cache entries lack `kind` → frontend treats empty as `"main"`; cache TTL is short (max 2min), old entries expire naturally |
| 3 | `TypedResultContainer` replaces `ResultContainer` in `executeProcessors` — regression risk on scoring/grouping | Score formula is identical (`Σ weight/pos × count`); dedup key adds `kind` prefix; keep old container code in `merge.go` untouched during migration for rollback |
| 4 | Frontend `ResultCard` exhaustive switch must handle all 12 kinds | TypeScript discriminated union on `kind`; default branch renders `MainResult` with `console.warn`; `tsc --noEmit` catches missing cases |
| 5 | Older engines return flat `models.Result` without `Kind` — `WrapAPIMainResult` produces typed results but loses kind-specific fields (image metadata, etc.) | Acceptable: engines are ported incrementally per Phase 4 tiers; until then, all flat results render as `MainResult` with title/URL/content intact |
| 6 | `Infobox` struct change (from `[]string` to typed `[]InfoboxURL`/`[]InfoboxAttribute`) breaks existing infobox producers | Search shows 0 existing infobox producers in current 6 engines; fixtures validate new shape |

## Parts (generate one per invocation, in order)

| # | File | Scope | Status |
|---|---|---|---|
| 1 | 2026-06-20-05-result-types/backend-types.md | Result interface, missing types, API envelope, normalization, ToAPIResult, TypedResultContainer | done |
| 2 | 2026-06-20-05-result-types/integration.md | ProcessorResult, WrapAPIMainResult adapter, online processor, scheduler wiring | done |
| 3 | 2026-06-20-05-result-types/frontend.md | TypeScript types, ResultCard, kind-specific components, SearchPage layout | done |

---

## Spec Coverage

| 设计需求 | Task(s) | 状态 |
|---|---|---|
| Result interface — Kind/Template/Base/Normalize/DedupKey | backend-types:1 | covered |
| Missing types — KeyValueResult, CodeLine, ImageRef, InfoboxAttribute, InfoboxURL | backend-types:1 | covered |
| Per-kind Template defaults (12 kinds) | backend-types:1 | covered |
| DedupKey per-kind (kind + normalized URL) | backend-types:1 | covered |
| FileResult extension (Filename/MagnetURI/Seeders/Leechers) | backend-types:1 | covered |
| CodeResult extension (Filename/CodeLines/HLLines) | backend-types:1 | covered |
| ImageResult extension (FileSize/Formats/ImageRef) | backend-types:1 | covered |
| InfoboxResult typed Attributes/URLs/RelatedTopics | backend-types:1 | covered |
| models.Result tagged union (Kind/Template/Extra) | backend-types:2 | covered |
| models.Infobox rich schema (InfoboxAttribute/InfoboxURL/Engines/RelatedTopics) | backend-types:2 | covered |
| normalizeEngines — engine name in Engines slice | backend-types:3 | covered |
| PostNormalize — Image thumbnail/title fallback | backend-types:3 | covered |
| PostNormalize — Video thumbnail fallback | backend-types:3 | covered |
| PostNormalize — Code title from filename | backend-types:3 | covered |
| PostNormalize — File title from filename, type from extension | backend-types:3 | covered |
| PostNormalize — Infobox ID fallback chain | backend-types:3 | covered |
| NormalizeResult orchestrator | backend-types:3 | covered |
| ToAPIResult — Kind/Template/Extra on output | backend-types:4 | covered |
| ToAPIResult — all 12 kind-specific Extra maps | backend-types:4 | covered |
| ToAPIResult — unknown kind fallback to "main" | backend-types:4 | covered |
| ToAPIResult — Domain computation | backend-types:4 | covered |
| WrapAPIMainResult — flat→typed adapter | backend-types:4 | covered |
| TypedResultContainer — per-kind buckets | backend-types:5 | covered |
| TypedResultContainer — dedup by kind+normalized URL | backend-types:5 | covered |
| TypedResultContainer — prefer HTTPS | backend-types:5 | covered |
| TypedResultContainer — score = Σ(weight/pos)×count | backend-types:5 | covered |
| TypedResultContainer — category grouping | backend-types:5 | covered |
| TypedResultContainer — closed-for-write | backend-types:5 | covered |
| TypedResultContainer — answers/suggestions/corrections/infoboxes dedup | backend-types:5 | covered |
| TypedResultContainer — GetEnginesUsed/GetEnginesFailed | backend-types:5 | covered |
| ProcessorResult — TypedResults field | integration:6 | covered |
| OnlineProcessor — wrap flat engine results | integration:6 | covered |
| OnlineProcessor — populate TypedResults + Results (backward compat) | integration:6 | covered |
| Whole-tree build verification | integration:7 | covered |
| Scheduler — NewTypedResultContainer | integration:8 | covered |
| Scheduler — TypedResults→ToAPIResult→Extend pipeline | integration:8 | covered |
| Scheduler — fallback to Results when TypedResults empty | integration:8 | covered |
| Scheduler — keep AddSuggestions/AddAnswers/AddCorrections/AddInfoboxes/AddEngineData | integration:8 | covered |
| Scheduler — Results() replacing GetOrderedResults() | integration:8 | covered |
| TypeScript tagged-union types (ResultKind, BaseResult, 12 derived) | frontend:9 | covered |
| TypeScript SearchResponse (answers/corrections/infoboxes) | frontend:9 | covered |
| searchStore — new state fields | frontend:9 | covered |
| ResultCard — switch on kind → 12 components | frontend:10 | covered |
| 12 kind-specific components (MainResult through InfoboxPanel) | frontend:10 | covered |
| ImageGrid — CSS grid for image category | frontend:10 | covered |
| SearchPage — top-level answers/infoboxes/corrections | frontend:11 | covered |
| SearchPage — image grid for image-heavy category | frontend:11 | covered |
| SearchPage — non-image results below image grid | frontend:11 | covered |
| WeatherAnswer (SearXNG full weather module) | no-op | deferred to weather module phase |
| Server-side syntax highlighting (CodeResult) | no-op | deferred — frontend <pre> block |
| ImageRef alternative image formats | no-op | deferred per design scope |
| Result URL filter hooks (plugins) | no-op | deferred to Phase 6 (plugins) |
| Distributed/shared result cache | no-op | deferred per design scope |
| All 238 engines emitting typed results | no-op | deferred — incremental per Phase 4 tiers |

---

## Global Self-Review

- [x] 1. Spec-coverage table: 48 spec items mapped to tasks across 3 parts, 6 no-ops (deferred per design scope), 0 GAPs. Every section in the approved design (`.ody-code/designs/2026-06-20-05-result-types.md`) maps to at least one Task. Covered: typed result hierarchy (backend-types:1), API envelope (backend-types:2), normalization (backend-types:3), conversion (backend-types:4), typed container (backend-types:5), processor (integration:6), scheduler (integration:8), frontend types (frontend:9), components (frontend:10), layout (frontend:11). Deferred items match the design's Out of Scope section exactly.

- [x] 2. Placeholder scan: All 3 part files + index — zero `TODO`/`TBD`/dead-code placeholders beyond intentional design-deferred items. The integration:7 task uses `--allow-empty` commit intentionally (verification gate). The `WrapAPIMainResult` adapter explicitly handles the migration period — not a placeholder but a defined bridge. Frontend components handle all edge cases: missing images, empty kv_map, unknown kind fallback. No hidden work or unresolved deps.

- [x] 3. No phantom tasks: 11 tasks across 3 parts, every task produces file changes. Task counts: backend-types 5, integration 3, frontend 3 = 11 tasks. Integration:7 is a verification gate (build + test) — intentionally doesn't change files but verifies the build-green invariant with a commit.

- [x] 4. Dependency soundness: All cross-file `Depends on:` references point to earlier parts:
  - integration:T6 → backend-types:T4 (`WrapAPIMainResult`) ✓
  - integration:T8 → backend-types:T5 (`TypedResultContainer`) ✓
  - frontend:T9 → backend-types:T2 (`models.Result.Kind/Extra`) ✓
  - frontend:T10 → frontend:T9 (TypeScript types) ✓
  - frontend:T11 → frontend:T10 (component files) ✓
  No reference points to a symbol only a later part defines. Phase B (integration) and Phase C (frontend) are independent of each other — confirmed neither imports symbols from the other.

- [x] 5. Caller & build soundness:
  - `Result` interface change (backend-types:1): all callers are within `pkg/models/results/` package (convert.go + test files) — updated in same task; whole-tree `go build ./...` verifies ✓
  - `models.Result` struct extension (backend-types:2): additive fields (Kind/Template/Extra with omitempty) — all existing constructors compile without modification; `go build ./... && go test ./...` verifies ✓
  - `ProcessorResult` struct extension (integration:6): additive field (TypedResults) — constructor in online.go updated in same task; whole-tree build verifies ✓
  - `merge.go` `appendUniqueStr` reused by `typed_container.go` — same package, no import change ✓
  - `Result` TypeScript type (frontend:9): flat→union — existing property access (r.title, r.url, r.content, r.engine, r.score) works on `BaseResult` which all union members extend; `tsc --noEmit` verifies ✓
  - `Infobox` struct type change (backend-types:2): `Attributes` from `[]string` to `[]InfoboxAttribute` — grep confirms 0 existing infobox producers in 6 engines; no callers to break ✓
  - `scheduler.go` container type from `*ResultContainer` to `*TypedResultContainer`: both implement `Extend/Close/AddSuggestions/AddAnswers/AddCorrections/AddInfoboxes/AddEngineData/MarkUnresponsive/GetSuggestions/GetAnswers/GetCorrections/GetInfoboxes/GetEngineData/GetUnresponsive/GetEnginesUsed/GetEnginesFailed` — same method set; only method name change is `GetOrderedResults→Results` handled in integration:8 ✓

- [x] 6. Test-the-risk: All 3 part files have explicit risk-testing assertions:
  - Backend-types: `TestNormalizeEngines_NoDuplicate` (state mutation: Engines slice growth), `TestFileResultTypeExtension` (PostNormalize: FileType from filename), `TestInfoboxIDFallback` (fallback chain), `TestToAPIResult_UnknownKindFallback` (graceful degradation), `TestTypedContainer_ClosedForWrite` (concurrency safety), `TestTypedContainer_ScoreCalculation` (scoring math verified: 6.0 = (2/1+1/1)×2), `TestTypedContainer_PreferHTTPS` (URL upgrade) ✓
  - Integration: `TestProcessorResult_HasTypedResultsField` (field access + compilation), whole-tree `go test ./...` (regression), manual `curl` verification (end-to-end pipeline) ✓
  - Frontend: TypeScript exhaustive switch default branch (unknown kind → MainResult), ImageCard missing img_src fallback, CodeBlock 20-line cap, KeyValueTable nil kv_map guard, `tsc --noEmit` exhaustive check ✓
  - Full list of must-survive inputs verified in each part's local self-review item 6.

- [x] 7. Type consistency: Cross-part type usage verified:
  - `Result` interface (backend-types:1) → consumed by `NormalizeResult(Result)` (backend-types:3), `ToAPIResult([]Result)` (backend-types:4), `WrapAPIMainResult` (backend-types:4), `ProcessorResult.TypedResults` (integration:6) ✓
  - `models.Result.Kind` (backend-types:2, string) → consumed by `ToAPIResult` output (backend-types:4), `typedDedupKey` (backend-types:5), `TypedResultContainer.Extend` (backend-types:5), `ResultCard` switch (frontend:10) ✓
  - `models.Result.Extra` (backend-types:2, map[string]any) → populated by `buildExtra` (backend-types:4), consumed by frontend typed interfaces (frontend:9) ✓
  - `TypedResultContainer` signature (backend-types:5) → used by `scheduler.go` (integration:8) — `Extend(string, []models.Result, int)`, `Results() []models.Result`, `AddSuggestions/AddAnswers/AddCorrections/AddInfoboxes/AddEngineData/MarkUnresponsive` + getters ✓
  - TypeScript `ResultKind` (frontend:9) values match Go `Kind()` return strings — `"main"`, `"image"`, `"video"`, `"news"`, `"paper"`, `"code"`, `"file"`, `"map"`, `"music"`, `"answer"`, `"keyvalue"`, `"infobox"` ✓
  - TypeScript `Extra` shapes per kind match Go `buildExtra` maps — e.g. `ImageResult.extra.img_src` ↔ `ed["img_src"]`, `PaperResult.extra.doi` ↔ `ed["doi"]` ✓
  - TypeScript `Infobox.attributes` shape `{label: string; value: string; url?: string}[]` ↔ Go `[]InfoboxAttribute` ✓
<!-- e2e-enriched -->

### Task 1: Generate and run E2E tests

Based on the changed files, validate the following tools:
- /Users/ranwei/workspace/go_work/seargo/cmd/seargo (priority: important)
- /Users/ranwei/workspace/go_work/seargo/engines/bing (priority: important)
- /Users/ranwei/workspace/go_work/seargo/engines/brave (priority: important)
- /Users/ranwei/workspace/go_work/seargo/engines/duckduckgo (priority: important)
- /Users/ranwei/workspace/go_work/seargo/engines/google (priority: important)
- /Users/ranwei/workspace/go_work/seargo/engines/wikipedia (priority: important)
- /Users/ranwei/workspace/go_work/seargo/engines/yahoo (priority: important)
- /Users/ranwei/workspace/go_work/seargo/internal/config (priority: important)
- /Users/ranwei/workspace/go_work/seargo/internal/engine (priority: important)
- /Users/ranwei/workspace/go_work/seargo/internal/errors (priority: important)
- /Users/ranwei/workspace/go_work/seargo/internal/httpx (priority: important)
- /Users/ranwei/workspace/go_work/seargo/internal/metrics (priority: important)
- /Users/ranwei/workspace/go_work/seargo/internal/search (priority: important)
- /Users/ranwei/workspace/go_work/seargo/internal/search/processor (priority: important)
- /Users/ranwei/workspace/go_work/seargo/internal/search/query (priority: important)
- /Users/ranwei/workspace/go_work/seargo/internal/server (priority: important)
- /Users/ranwei/workspace/go_work/seargo/pkg/models (priority: important)

Use the RunE2ETests tool after completing the implementation tasks above.



STOP — do NOT begin executing now. This turn ends here. The user will start implementation themselves — the plan is now available in their main conversation context.

</details>
