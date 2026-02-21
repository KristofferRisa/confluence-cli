package models

// Page represents a Confluence page (v2 API)
type Page struct {
	ID         string   `json:"id"`
	Status     string   `json:"status"`
	Title      string   `json:"title"`
	SpaceID    string   `json:"spaceId"`
	ParentID   string   `json:"parentId,omitempty"`
	ParentType string   `json:"parentType,omitempty"`
	AuthorID   string   `json:"authorId,omitempty"`
	CreatedAt  string   `json:"createdAt,omitempty"`
	Version    *Version `json:"version,omitempty"`
	Body       *Body    `json:"body,omitempty"`
	Links      *Links   `json:"_links,omitempty"`
}

// Body contains page content in storage format
type Body struct {
	Storage *BodyContent `json:"storage,omitempty"`
}

// BodyContent holds the actual content value
type BodyContent struct {
	Value          string `json:"value"`
	Representation string `json:"representation"`
}

// Version tracks page version for optimistic locking
type Version struct {
	Number    int    `json:"number"`
	Message   string `json:"message,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
	AuthorID  string `json:"authorId,omitempty"`
	MinorEdit bool   `json:"minorEdit,omitempty"`
}

// Space represents a Confluence space (v2 API)
type Space struct {
	ID          string            `json:"id"`
	Key         string            `json:"key"`
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Status      string            `json:"status"`
	AuthorID    string            `json:"authorId,omitempty"`
	CreatedAt   string            `json:"createdAt,omitempty"`
	HomepageID  string            `json:"homepageId,omitempty"`
	Description *SpaceDescription `json:"description,omitempty"`
	Links       *Links            `json:"_links,omitempty"`
}

// SpaceDescription holds space description content
type SpaceDescription struct {
	Plain *BodyContent `json:"plain,omitempty"`
}

// Label represents a page label/tag (v1 API)
type Label struct {
	ID     string `json:"id"`
	Prefix string `json:"prefix"`
	Name   string `json:"name"`
}

// Attachment represents a file attached to a page (v1 API)
type Attachment struct {
	ID         string          `json:"id"`
	Type       string          `json:"type"`
	Title      string          `json:"title"`
	Metadata   AttachmentMeta  `json:"metadata,omitempty"`
	Extensions AttachmentExt   `json:"extensions,omitempty"`
	Links      *AttachmentLinks `json:"_links,omitempty"`
}

// AttachmentMeta holds attachment metadata
type AttachmentMeta struct {
	MediaType string `json:"mediaType"`
	Comment   string `json:"comment,omitempty"`
}

// AttachmentExt holds attachment extension data
type AttachmentExt struct {
	MediaType string `json:"mediaType"`
	FileSize  int64  `json:"fileSize"`
}

// AttachmentLinks holds attachment-specific links
type AttachmentLinks struct {
	Download string `json:"download"`
	WebUI    string `json:"webui,omitempty"`
}

// SearchResult holds CQL search results (v1 API)
type SearchResult struct {
	Results   []SearchEntry `json:"results"`
	Start     int           `json:"start"`
	Limit     int           `json:"limit"`
	Size      int           `json:"size"`
	TotalSize int           `json:"totalSize"`
}

// SearchEntry is a single search result
type SearchEntry struct {
	Content      SearchContent `json:"content"`
	Title        string        `json:"title"`
	Excerpt      string        `json:"excerpt"`
	URL          string        `json:"url"`
	LastModified string        `json:"lastModified"`
}

// SearchContent holds the content portion of a search result
type SearchContent struct {
	ID    string       `json:"id"`
	Type  string       `json:"type"`
	Title string       `json:"title"`
	Space *SearchSpace `json:"space,omitempty"`
	Links *Links       `json:"_links,omitempty"`
}

// SearchSpace holds space info within search results
type SearchSpace struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

// --- Request types ---

// CreatePageRequest is the payload for creating a page (v2 API)
type CreatePageRequest struct {
	SpaceID  string         `json:"spaceId"`
	Status   string         `json:"status"`
	Title    string         `json:"title"`
	ParentID string         `json:"parentId,omitempty"`
	Body     CreatePageBody `json:"body"`
}

// CreatePageBody holds the body for create/update requests
type CreatePageBody struct {
	Representation string `json:"representation"`
	Value          string `json:"value"`
}

// UpdatePageRequest is the payload for updating a page (v2 API)
type UpdatePageRequest struct {
	ID      string         `json:"id"`
	Status  string         `json:"status"`
	Title   string         `json:"title"`
	Body    CreatePageBody `json:"body"`
	Version UpdateVersion  `json:"version"`
}

// UpdateVersion specifies the new version number for updates
type UpdateVersion struct {
	Number  int    `json:"number"`
	Message string `json:"message,omitempty"`
}

// --- Pagination ---

// ListOptions provides pagination and filtering parameters
type ListOptions struct {
	Limit  int
	Cursor string
	Status string
}

// --- Response wrappers ---

// PageList is a paginated list of pages
type PageList struct {
	Results []Page `json:"results"`
	Links   *Links `json:"_links,omitempty"`
}

// SpaceList is a paginated list of spaces
type SpaceList struct {
	Results []Space `json:"results"`
	Links   *Links  `json:"_links,omitempty"`
}

// LabelList is a paginated list of labels (v1 API)
type LabelList struct {
	Results []Label `json:"results"`
	Start   int     `json:"start"`
	Limit   int     `json:"limit"`
	Size    int     `json:"size"`
}

// AttachmentList is a paginated list of attachments (v1 API)
type AttachmentList struct {
	Results []Attachment `json:"results"`
	Start   int          `json:"start"`
	Limit   int          `json:"limit"`
	Size    int          `json:"size"`
}

// Links holds pagination and navigation links
type Links struct {
	Next  string `json:"next,omitempty"`
	WebUI string `json:"webui,omitempty"`
}

// PageTree represents the page hierarchy for tree display
type PageTree struct {
	Page     Page       `json:"page"`
	Children []PageTree `json:"children,omitempty"`
}
