package jsonld

import (
	"encoding/json"
	"time"
)

// WebPage is one page of the site: schema.org/WebPage, or one of its more
// specific kinds through PageType.
//
// Its Type, and so its key, is the PageType: an AboutPage and a WebPage emitted on
// one page are two nodes, as an Article and a BlogPosting are.
type WebPage struct {
	// ID and Key are as on Article.
	ID  string
	Key string
	// PageType is the kind of page; empty is PageWebPage.
	PageType    WebPageType
	Name        string
	Description string
	URL         string
	// InLanguage is the page's language as an IETF tag: "tr", "en-GB".
	InLanguage    string
	DatePublished time.Time
	DateModified  time.Time
	ImageURL      string
	// IsPartOf is the "@id" of the WebSite node the page belongs to.
	IsPartOf string
	// Speakable, Paywalled and PaywallCSS are as on Article.
	Speakable  []string
	Paywalled  bool
	PaywallCSS []string
}

// WebPageType is the schema.org type of a WebPage.
type WebPageType string

const (
	PageWebPage       WebPageType = "WebPage"
	PageAbout         WebPageType = "AboutPage"
	PageContact       WebPageType = "ContactPage"
	PageCollection    WebPageType = "CollectionPage"
	PageItem          WebPageType = "ItemPage"
	PageSearchResults WebPageType = "SearchResultsPage"
	PageCheckout      WebPageType = "CheckoutPage"
)

func (w WebPage) Type() string {
	if w.PageType == "" {
		return string(PageWebPage)
	}
	return string(w.PageType)
}

func (w WebPage) nodeKey() string { return w.Key }

func (w WebPage) MarshalJSON() ([]byte, error) {
	out := webPageJSON{
		Context:       schemaContext,
		ID:            w.ID,
		SchemaType:    w.Type(),
		Name:          w.Name,
		Description:   w.Description,
		URL:           w.URL,
		InLanguage:    w.InLanguage,
		DatePublished: dateTime(w.DatePublished),
		DateModified:  dateTime(w.DateModified),
		Speakable:     nodeSpeakable(w.Speakable),
		HasPart:       nodePaywallParts(w.PaywallCSS),
	}
	if w.ImageURL != "" {
		out.Image = &imageObject{SchemaType: "ImageObject", URL: w.ImageURL}
	}
	if w.IsPartOf != "" {
		out.IsPartOf = &pagesRefJSON{ID: w.IsPartOf}
	}
	if w.Paywalled {
		free := false
		out.IsAccessibleForFree = &free
	}
	return json.Marshal(out)
}

// ProfilePage is a page about one person or organization — an author's page, a
// member's profile: schema.org/ProfilePage.
type ProfilePage struct {
	// ID and Key are as on Article.
	ID  string
	Key string
	// MainEntity is who the page is about, a Person or an Organization. It is
	// written inside the page, without an "@context" of its own.
	MainEntity   Node
	DateCreated  time.Time
	DateModified time.Time
}

func (ProfilePage) Type() string { return "ProfilePage" }

func (p ProfilePage) nodeKey() string { return p.Key }

func (p ProfilePage) MarshalJSON() ([]byte, error) {
	out := profilePageJSON{
		Context:      schemaContext,
		ID:           p.ID,
		SchemaType:   "ProfilePage",
		DateCreated:  dateTime(p.DateCreated),
		DateModified: dateTime(p.DateModified),
	}
	if p.MainEntity != nil {
		entity, err := nested(p.MainEntity)
		if err != nil {
			return nil, err
		}
		out.MainEntity = entity
	}
	return json.Marshal(out)
}

type webPageJSON struct {
	Context             string                `json:"@context"`
	ID                  string                `json:"@id,omitempty"`
	SchemaType          string                `json:"@type"`
	Name                string                `json:"name,omitempty"`
	Description         string                `json:"description,omitempty"`
	URL                 string                `json:"url,omitempty"`
	InLanguage          string                `json:"inLanguage,omitempty"`
	DatePublished       string                `json:"datePublished,omitempty"`
	DateModified        string                `json:"dateModified,omitempty"`
	Image               *imageObject          `json:"image,omitempty"`
	IsPartOf            *pagesRefJSON         `json:"isPartOf,omitempty"`
	Speakable           *nodeSpeakableJSON    `json:"speakable,omitempty"`
	IsAccessibleForFree *bool                 `json:"isAccessibleForFree,omitempty"`
	HasPart             []nodePaywallPartJSON `json:"hasPart,omitempty"`
}

// pagesRefJSON points at a node by its "@id" alone.
type pagesRefJSON struct {
	ID string `json:"@id"`
}

type profilePageJSON struct {
	Context      string          `json:"@context"`
	ID           string          `json:"@id,omitempty"`
	SchemaType   string          `json:"@type"`
	DateCreated  string          `json:"dateCreated,omitempty"`
	DateModified string          `json:"dateModified,omitempty"`
	MainEntity   json.RawMessage `json:"mainEntity,omitempty"`
}

var (
	_ keyed = WebPage{}
	_ keyed = ProfilePage{}
)
