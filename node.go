package jsonld

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

// Node is one piece of structured data, ready to be emitted.
//
// It is an interface over concrete types rather than a map, and that is the whole
// design decision in this file. JSON-LD is open-ended, so the obvious shape is
// map[string]any — and then nothing checks a misspelled property, the compiler has
// nothing to say about a missing one, and every application invents its own spelling
// of the same vocabulary. The types below cover what a content site actually emits;
// Raw is the escape hatch for everything else, and it makes the escape explicit
// rather than making it the default.
type Node interface {
	json.Marshaler
	// Type is the schema.org type this node declares, used in error messages and
	// to keep the interface from being satisfied by accident.
	Type() string
}

// schemaContext is the JSON-LD context every node declares. It is not named
// "context" because this package also imports Go's.
const schemaContext = "https://schema.org"

// Raw is structured data the application marshalled itself, for a schema.org type
// this package does not model.
//
// It is checked for being valid JSON when it is marshalled, because invalid JSON
// inside a <script> tag is worse than absent JSON: a consumer that fails to parse it
// may discard every other node on the page along with it.
type Raw struct {
	// SchemaType is what the data declares as its "@type", for diagnostics.
	SchemaType string
	// JSON is the node, already marshalled, as a JSON object. It needs no
	// "@context" property: MarshalJSON puts one first, replacing any the object
	// has, so every node on a page declares it the same way. The other members
	// keep the order they are written in.
	JSON json.RawMessage
	// Key is as on the other nodes: never written, it tells Emit which
	// declarations are the same node. An "@id" goes in JSON like any other member.
	Key string
}

func (r Raw) Type() string { return r.SchemaType }

func (r Raw) nodeKey() string { return r.Key }

// MarshalJSON writes "@context" first and then the caller's members in the order
// the caller wrote them.
//
// It walks the object member by member rather than decoding it into a map, because
// a map comes back out in alphabetical order: "@type" would land after
// "description", which is valid JSON-LD and miserable to read in a page's source.
// Each value is copied through verbatim. A member the caller named "@context" is
// dropped rather than repeated, so the node declares the same context as every other
// node on the page — and a duplicate key is one a consumer may resolve either way.
func (r Raw) MarshalJSON() ([]byte, error) {
	// Valid first, so the walk below only has to understand the shape and never
	// decides what counts as JSON.
	if !json.Valid(r.JSON) {
		return nil, fmt.Errorf("jsonld: raw %s node is not valid JSON", r.SchemaType)
	}
	dec := json.NewDecoder(bytes.NewReader(r.JSON))
	if open, err := dec.Token(); err != nil || open != json.Delim('{') {
		// An array, a string or null has nowhere to put "@context". null in
		// particular used to decode into a nil map and panic on the assignment.
		return nil, fmt.Errorf("jsonld: raw %s node is not a JSON object", r.SchemaType)
	}

	var out bytes.Buffer
	out.WriteString(`{"@context":"` + schemaContext + `"`)
	for dec.More() {
		keyToken, err := dec.Token()
		if err != nil {
			return nil, err
		}
		key, _ := keyToken.(string) // json.Valid held, so an object key is a string
		var value json.RawMessage
		if err := dec.Decode(&value); err != nil {
			return nil, err
		}
		if key == "@context" {
			continue
		}
		encodedKey, err := json.Marshal(key)
		if err != nil {
			return nil, err
		}
		out.WriteByte(',')
		out.Write(encodedKey)
		out.WriteByte(':')
		out.Write(value)
	}
	out.WriteByte('}')

	// The verbatim values are the caller's bytes, which may hold a literal "<".
	// Escaping here keeps Raw holding the same two layers every other node does,
	// rather than leaning on json.Marshal's pass alone.
	var escaped bytes.Buffer
	json.HTMLEscape(&escaped, out.Bytes())
	return escaped.Bytes(), nil
}

// Article is a piece of writing: schema.org/Article.
type Article struct {
	// ID is the node's "@id": an IRI naming the thing itself, which is how a
	// consumer tells two mentions of one entity from two entities. Left empty, the
	// node has no "@id".
	ID string
	// Key tells Emit which declarations are the same node. It is never written to
	// the page. Nodes of one type with different keys all appear; the same key
	// declared twice keeps the innermost. Empty is a key too: a type emitted
	// without one is a single node per page, as WebSite is.
	Key           string
	Headline      string
	Description   string
	URL           string
	Section       string
	Keywords      []string
	DatePublished time.Time
	DateModified  time.Time
	AuthorName    string
	AuthorURL     string
	PublisherName string
	PublisherLogo string
	ImageURL      string
	WordCount     int
	// Authors are the article's authors beyond AuthorName, or all of them when
	// AuthorName is empty. With more than AuthorName, "author" is a list.
	Authors []Agent
	// ImageURLs are more images of the article — schema.org consumers prefer
	// several aspect ratios. With any, "image" is a list of URLs, ImageURL first.
	ImageURLs []string
	// Speakable are CSS selectors for the parts of the page suited to being read
	// aloud.
	Speakable []string
	// Paywalled marks the article as not free to read; PaywallCSS are the CSS
	// selectors of the parts behind the paywall. Search engines tell paywalled
	// content from cloaking by these.
	Paywalled  bool
	PaywallCSS []string
}

func (Article) Type() string { return "Article" }

func (a Article) nodeKey() string { return a.Key }

func (a Article) MarshalJSON() ([]byte, error) {
	out := a.wire("Article")
	out.Context = schemaContext
	return json.Marshal(out)
}

// wire is the article's schema.org shape under schemaType, without "@context" —
// which only a top-level node declares, so a post nested in a Blog leaves it out.
func (a Article) wire(schemaType string) articleJSON {
	out := articleJSON{
		ID:          a.ID,
		SchemaType:  schemaType,
		Headline:    a.Headline,
		Description: a.Description,
		URL:         a.URL,
		Section:     a.Section,
		Keywords:    a.Keywords,
		WordCount:   a.WordCount,
		Author:      nodeAuthors(a.AuthorName, a.AuthorURL, a.Authors),
		Publisher:   publisherNode(a.PublisherName, a.PublisherLogo),
		Speakable:   nodeSpeakable(a.Speakable),
		HasPart:     nodePaywallParts(a.PaywallCSS),
	}
	if a.Paywalled {
		free := false
		out.IsAccessibleForFree = &free
	}
	if !a.DatePublished.IsZero() {
		out.DatePublished = a.DatePublished.Format(time.RFC3339)
	}
	if !a.DateModified.IsZero() {
		out.DateModified = a.DateModified.Format(time.RFC3339)
	}
	out.Image = nodeImages(a.ImageURL, a.ImageURLs)
	return out
}

// nodeAuthorField is an article's "author": the single object it always was
// when only AuthorName is set, and a list once Authors adds to it.
type nodeAuthorField struct {
	single bool
	list   []agentJSON
}

func (f nodeAuthorField) MarshalJSON() ([]byte, error) {
	if f.single {
		return json.Marshal(f.list[0])
	}
	return json.Marshal(f.list)
}

func nodeAuthors(name, url string, more []Agent) *nodeAuthorField {
	var list []agentJSON
	if name != "" {
		list = append(list, Agent{Name: name, URL: url}.wire())
	}
	list = append(list, agents(more)...)
	if len(list) == 0 {
		return nil
	}
	return &nodeAuthorField{single: len(more) == 0, list: list}
}

// nodeImageField is an article's "image": the ImageObject it always was with
// ImageURL alone, and a list of URLs once ImageURLs adds to it.
type nodeImageField struct {
	object *imageObject
	urls   []string
}

func (f nodeImageField) MarshalJSON() ([]byte, error) {
	if f.object != nil {
		return json.Marshal(f.object)
	}
	return json.Marshal(f.urls)
}

func nodeImages(first string, more []string) *nodeImageField {
	if len(more) == 0 {
		if first == "" {
			return nil
		}
		return &nodeImageField{object: &imageObject{SchemaType: "ImageObject", URL: first}}
	}
	var urls []string
	if first != "" {
		urls = append(urls, first)
	}
	return &nodeImageField{urls: append(urls, more...)}
}

type nodeSpeakableJSON struct {
	SchemaType  string   `json:"@type"`
	CSSSelector []string `json:"cssSelector"`
}

func nodeSpeakable(selectors []string) *nodeSpeakableJSON {
	if len(selectors) == 0 {
		return nil
	}
	return &nodeSpeakableJSON{SchemaType: "SpeakableSpecification", CSSSelector: selectors}
}

type nodePaywallPartJSON struct {
	SchemaType          string `json:"@type"`
	IsAccessibleForFree bool   `json:"isAccessibleForFree"`
	CSSSelector         string `json:"cssSelector"`
}

func nodePaywallParts(selectors []string) []nodePaywallPartJSON {
	var out []nodePaywallPartJSON
	for _, sel := range selectors {
		out = append(out, nodePaywallPartJSON{SchemaType: "WebPageElement", CSSSelector: sel})
	}
	return out
}

// BlogPosting is a post on a blog: schema.org/BlogPosting, which schema.org
// derives from Article and which describes a blog post more precisely than
// Article does.
//
// It is Article's own struct under another name rather than a copy of its
// fields, so the two cannot drift apart: a property added to Article is a
// property of a post too. Being a separate type is what gives it its own
// "@type" — and its own key, so a page can emit both.
type BlogPosting Article

func (BlogPosting) Type() string { return "BlogPosting" }

func (b BlogPosting) nodeKey() string { return b.Key }

func (b BlogPosting) MarshalJSON() ([]byte, error) {
	out := Article(b).wire("BlogPosting")
	out.Context = schemaContext
	return json.Marshal(out)
}

// NewsArticle is a news story: schema.org/NewsArticle. It is Article's own
// struct under another name, as BlogPosting is, with its own "@type" and key.
type NewsArticle Article

func (NewsArticle) Type() string { return "NewsArticle" }

func (n NewsArticle) nodeKey() string { return n.Key }

func (n NewsArticle) MarshalJSON() ([]byte, error) {
	out := Article(n).wire("NewsArticle")
	out.Context = schemaContext
	return json.Marshal(out)
}

// Blog is the blog itself, typically on its index page: schema.org/Blog.
type Blog struct {
	// ID and Key are as on Article.
	ID            string
	Key           string
	Name          string
	Description   string
	URL           string
	AuthorName    string
	AuthorURL     string
	PublisherName string
	PublisherLogo string
	ImageURL      string
	// Posts are the posts the page lists, emitted as the blog's "blogPost"
	// property. Each carries no "@context" of its own; the blog's covers it.
	Posts []BlogPosting
}

func (Blog) Type() string { return "Blog" }

func (b Blog) nodeKey() string { return b.Key }

func (b Blog) MarshalJSON() ([]byte, error) {
	out := blogJSON{
		Context:     schemaContext,
		ID:          b.ID,
		SchemaType:  "Blog",
		Name:        b.Name,
		Description: b.Description,
		URL:         b.URL,
		Author:      authorNode(b.AuthorName, b.AuthorURL),
		Publisher:   publisherNode(b.PublisherName, b.PublisherLogo),
	}
	if b.ImageURL != "" {
		out.Image = &imageObject{SchemaType: "ImageObject", URL: b.ImageURL}
	}
	for _, post := range b.Posts {
		out.BlogPost = append(out.BlogPost, Article(post).wire("BlogPosting"))
	}
	return json.Marshal(out)
}

// Person is someone in their own right — an author's page, an about page:
// schema.org/Person. An article's author is described by Article's own fields;
// this is for a page about the person.
type Person struct {
	// ID and Key are as on Article.
	ID          string
	Key         string
	Name        string
	Description string
	URL         string
	ImageURL    string
	JobTitle    string
	// Email is written as given; schema.org takes the plain address.
	Email string
	// KnowsAbout are the subjects the person is known for: "Go", "Distributed
	// systems".
	KnowsAbout []string
	// SameAs are the person's profiles elsewhere, which is how a consumer ties
	// this page to the same person on other sites.
	SameAs        []string
	GivenName     string
	FamilyName    string
	AlternateName string
	Telephone     string
	// WorksFor is the person's employer; its Kind defaults to Organization.
	WorksFor Agent
	Address  *PostalAddress
}

func (Person) Type() string { return "Person" }

func (p Person) nodeKey() string { return p.Key }

func (p Person) MarshalJSON() ([]byte, error) {
	out := personJSON{
		Context:     schemaContext,
		ID:          p.ID,
		SchemaType:  "Person",
		Name:        p.Name,
		Description: p.Description,
		URL:         p.URL,
		JobTitle:    p.JobTitle,
		Email:       p.Email,
		KnowsAbout:  p.KnowsAbout,
		SameAs:      p.SameAs,
		GivenName:   p.GivenName,
		FamilyName:  p.FamilyName,
		Alternate:   p.AlternateName,
		Telephone:   p.Telephone,
		Address:     p.Address.wire(),
	}
	if p.WorksFor.Name != "" {
		employer := p.WorksFor
		if employer.Kind == "" {
			employer.Kind = AgentOrganization
		}
		out.WorksFor = agentPtr(employer)
	}
	if p.ImageURL != "" {
		out.Image = &imageObject{SchemaType: "ImageObject", URL: p.ImageURL}
	}
	return json.Marshal(out)
}

// authorNode and publisherNode describe the author and publisher that Article
// and Blog both flatten into name fields, and are nil when the name is empty so
// the property is left out rather than emitted nameless.

func authorNode(name, url string) *person {
	if name == "" {
		return nil
	}
	return &person{SchemaType: "Person", Name: name, URL: url}
}

func publisherNode(name, logo string) *organization {
	if name == "" {
		return nil
	}
	out := &organization{SchemaType: "Organization", Name: name}
	if logo != "" {
		out.Logo = &imageObject{SchemaType: "ImageObject", URL: logo}
	}
	return out
}

// WebSite describes the site itself: schema.org/WebSite.
type WebSite struct {
	// ID and Key are as on Article.
	ID   string
	Key  string
	Name string
	URL  string
	// SearchURL is a search endpoint with "{query}" where the term goes. Setting
	// it emits a SearchAction, which is what a search box in a result listing is
	// built from.
	SearchURL string
}

func (WebSite) Type() string { return "WebSite" }

func (w WebSite) nodeKey() string { return w.Key }

func (w WebSite) MarshalJSON() ([]byte, error) {
	out := webSiteJSON{Context: schemaContext, ID: w.ID, SchemaType: "WebSite", Name: w.Name, URL: w.URL}
	if w.SearchURL != "" {
		out.PotentialAction = &searchAction{
			SchemaType: "SearchAction",
			Target:     w.SearchURL,
			QueryInput: "required name=query",
		}
	}
	return json.Marshal(out)
}

// Organization is a company, a community, a publisher in its own right:
// schema.org/Organization. Article's publisher is described by Article's own
// fields; this is for the organization behind the site, or a page about one.
type Organization struct {
	// ID and Key are as on Article.
	ID          string
	Key         string
	Name        string
	Description string
	URL         string
	LogoURL     string
	// Email is written as given; schema.org takes the plain address.
	Email string
	// Telephone is written as given; the international form, "+90 212 …", is
	// the one consumers read without guessing.
	Telephone string
	// SameAs are the organization's profiles elsewhere — its Instagram, its
	// LinkedIn — which is how a consumer ties this page to the same organization
	// on other sites.
	SameAs        []string
	Address       *PostalAddress
	AlternateName string
	LegalName     string
	// FoundingDate is written as a date.
	FoundingDate      time.Time
	NumberOfEmployees int
	// The identifiers below are written as given: tax and VAT numbers, the D-U-N-S
	// number, the Legal Entity Identifier, the NAICS code, the GS1 Global Location
	// Number and the ISO 6523 identifier.
	TaxID                string
	VatID                string
	DUNS                 string
	LEICode              string
	NAICS                string
	GlobalLocationNumber string
	ISO6523Code          string
	// ContactPoints are the ways to reach the organization for a purpose —
	// customer service, sales.
	ContactPoints []ContactPoint
}

// ContactPoint is one way to reach an organization: schema.org/ContactPoint.
type ContactPoint struct {
	Telephone string
	Email     string
	// ContactType is what the contact point is for: "customer service".
	ContactType string
	// AreaServed are the regions served, as ISO 3166-1 codes or names.
	AreaServed []string
	// AvailableLanguage are the languages spoken, as names or IETF tags.
	AvailableLanguage []string
}

type contactPointJSON struct {
	SchemaType        string   `json:"@type"`
	Telephone         string   `json:"telephone,omitempty"`
	Email             string   `json:"email,omitempty"`
	ContactType       string   `json:"contactType,omitempty"`
	AreaServed        []string `json:"areaServed,omitempty"`
	AvailableLanguage []string `json:"availableLanguage,omitempty"`
}

func (Organization) Type() string { return "Organization" }

func (o Organization) nodeKey() string { return o.Key }

func (o Organization) MarshalJSON() ([]byte, error) {
	out := organizationJSON{
		Context:     schemaContext,
		ID:          o.ID,
		SchemaType:  "Organization",
		Name:        o.Name,
		Description: o.Description,
		URL:         o.URL,
		Email:       o.Email,
		Telephone:   o.Telephone,
		SameAs:      o.SameAs,
		Address:     o.Address.wire(),
		Alternate:   o.AlternateName,
		LegalName:   o.LegalName,
		Founding:    date(o.FoundingDate),
		TaxID:       o.TaxID,
		VatID:       o.VatID,
		DUNS:        o.DUNS,
		LEICode:     o.LEICode,
		NAICS:       o.NAICS,
		GLN:         o.GlobalLocationNumber,
		ISO6523:     o.ISO6523Code,
	}
	if o.NumberOfEmployees > 0 {
		out.Employees = &nodeQuantityJSON{SchemaType: "QuantitativeValue", Value: o.NumberOfEmployees}
	}
	for _, c := range o.ContactPoints {
		out.ContactPoint = append(out.ContactPoint, contactPointJSON{
			SchemaType:        "ContactPoint",
			Telephone:         c.Telephone,
			Email:             c.Email,
			ContactType:       c.ContactType,
			AreaServed:        c.AreaServed,
			AvailableLanguage: c.AvailableLanguage,
		})
	}
	if o.LogoURL != "" {
		out.Logo = &imageObject{SchemaType: "ImageObject", URL: o.LogoURL}
	}
	return json.Marshal(out)
}

// FAQPage is a page of questions and their answers: schema.org/FAQPage.
type FAQPage struct {
	// ID and Key are as on Article.
	ID        string
	Key       string
	Questions []Question
}

// Question is one entry on an FAQPage. Answer is text; search engines show it
// as written, so keep markup out of it.
type Question struct {
	Question string
	Answer   string
}

func (FAQPage) Type() string { return "FAQPage" }

func (f FAQPage) nodeKey() string { return f.Key }

func (f FAQPage) MarshalJSON() ([]byte, error) {
	entities := make([]questionJSON, 0, len(f.Questions))
	for _, q := range f.Questions {
		entities = append(entities, questionJSON{
			SchemaType:     "Question",
			Name:           q.Question,
			AcceptedAnswer: answerJSON{SchemaType: "Answer", Text: q.Answer},
		})
	}
	return json.Marshal(faqPageJSON{
		Context:    schemaContext,
		ID:         f.ID,
		SchemaType: "FAQPage",
		MainEntity: entities,
	})
}

// Breadcrumb is one step in a trail.
type Breadcrumb struct {
	Name string
	URL  string
}

// BreadcrumbList is the trail to the current page: schema.org/BreadcrumbList.
type BreadcrumbList struct {
	// ID and Key are as on Article.
	ID    string
	Key   string
	Items []Breadcrumb
}

func (BreadcrumbList) Type() string { return "BreadcrumbList" }

func (b BreadcrumbList) nodeKey() string { return b.Key }

func (b BreadcrumbList) MarshalJSON() ([]byte, error) {
	items := make([]listItem, 0, len(b.Items))
	for i, crumb := range b.Items {
		items = append(items, listItem{
			SchemaType: "ListItem",
			Position:   i + 1, // schema.org positions are 1-based
			Name:       crumb.Name,
			Item:       crumb.URL,
		})
	}
	return json.Marshal(breadcrumbJSON{
		Context:         schemaContext,
		ID:              b.ID,
		SchemaType:      "BreadcrumbList",
		ItemListElement: items,
	})
}

// The wire shapes. They are separate from the exported types so the exported ones
// can use Go's vocabulary — time.Time, a slice of Breadcrumb — while what goes over
// the wire uses schema.org's.

type articleJSON struct {
	Context             string                `json:"@context,omitempty"` // empty only when nested in a Blog
	ID                  string                `json:"@id,omitempty"`
	SchemaType          string                `json:"@type"`
	Headline            string                `json:"headline,omitempty"`
	Description         string                `json:"description,omitempty"`
	URL                 string                `json:"url,omitempty"`
	Section             string                `json:"articleSection,omitempty"`
	Keywords            []string              `json:"keywords,omitempty"`
	DatePublished       string                `json:"datePublished,omitempty"`
	DateModified        string                `json:"dateModified,omitempty"`
	WordCount           int                   `json:"wordCount,omitempty"`
	Author              *nodeAuthorField      `json:"author,omitempty"`
	Publisher           *organization         `json:"publisher,omitempty"`
	Image               *nodeImageField       `json:"image,omitempty"`
	Speakable           *nodeSpeakableJSON    `json:"speakable,omitempty"`
	IsAccessibleForFree *bool                 `json:"isAccessibleForFree,omitempty"`
	HasPart             []nodePaywallPartJSON `json:"hasPart,omitempty"`
}

type blogJSON struct {
	Context     string        `json:"@context"`
	ID          string        `json:"@id,omitempty"`
	SchemaType  string        `json:"@type"`
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	URL         string        `json:"url,omitempty"`
	Author      *person       `json:"author,omitempty"`
	Publisher   *organization `json:"publisher,omitempty"`
	Image       *imageObject  `json:"image,omitempty"`
	BlogPost    []articleJSON `json:"blogPost,omitempty"`
}

type personJSON struct {
	Context     string             `json:"@context"`
	ID          string             `json:"@id,omitempty"`
	SchemaType  string             `json:"@type"`
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	URL         string             `json:"url,omitempty"`
	Image       *imageObject       `json:"image,omitempty"`
	JobTitle    string             `json:"jobTitle,omitempty"`
	Email       string             `json:"email,omitempty"`
	KnowsAbout  []string           `json:"knowsAbout,omitempty"`
	SameAs      []string           `json:"sameAs,omitempty"`
	GivenName   string             `json:"givenName,omitempty"`
	FamilyName  string             `json:"familyName,omitempty"`
	Alternate   string             `json:"alternateName,omitempty"`
	Telephone   string             `json:"telephone,omitempty"`
	WorksFor    *agentJSON         `json:"worksFor,omitempty"`
	Address     *postalAddressJSON `json:"address,omitempty"`
}

// person is an author nested in another node; Person is one emitted in its own
// right, with the properties a page about someone has.
type person struct {
	SchemaType string `json:"@type"`
	Name       string `json:"name"`
	URL        string `json:"url,omitempty"`
}

type organization struct {
	SchemaType string       `json:"@type"`
	Name       string       `json:"name"`
	URL        string       `json:"url,omitempty"`
	Logo       *imageObject `json:"logo,omitempty"`
}

type imageObject struct {
	SchemaType string `json:"@type"`
	URL        string `json:"url"`
}

type webSiteJSON struct {
	Context         string        `json:"@context"`
	ID              string        `json:"@id,omitempty"`
	SchemaType      string        `json:"@type"`
	Name            string        `json:"name"`
	URL             string        `json:"url,omitempty"`
	PotentialAction *searchAction `json:"potentialAction,omitempty"`
}

type searchAction struct {
	SchemaType string `json:"@type"`
	Target     string `json:"target"`
	QueryInput string `json:"query-input"`
}

type breadcrumbJSON struct {
	Context         string     `json:"@context"`
	ID              string     `json:"@id,omitempty"`
	SchemaType      string     `json:"@type"`
	ItemListElement []listItem `json:"itemListElement"`
}

type listItem struct {
	SchemaType string `json:"@type"`
	Position   int    `json:"position"`
	Name       string `json:"name"`
	Item       string `json:"item,omitempty"`
}

type organizationJSON struct {
	Context      string             `json:"@context"`
	ID           string             `json:"@id,omitempty"`
	SchemaType   string             `json:"@type"`
	Name         string             `json:"name"`
	Description  string             `json:"description,omitempty"`
	URL          string             `json:"url,omitempty"`
	Logo         *imageObject       `json:"logo,omitempty"`
	Email        string             `json:"email,omitempty"`
	Telephone    string             `json:"telephone,omitempty"`
	SameAs       []string           `json:"sameAs,omitempty"`
	Address      *postalAddressJSON `json:"address,omitempty"`
	Alternate    string             `json:"alternateName,omitempty"`
	LegalName    string             `json:"legalName,omitempty"`
	Founding     string             `json:"foundingDate,omitempty"`
	Employees    *nodeQuantityJSON  `json:"numberOfEmployees,omitempty"`
	TaxID        string             `json:"taxID,omitempty"`
	VatID        string             `json:"vatID,omitempty"`
	DUNS         string             `json:"duns,omitempty"`
	LEICode      string             `json:"leiCode,omitempty"`
	NAICS        string             `json:"naics,omitempty"`
	GLN          string             `json:"globalLocationNumber,omitempty"`
	ISO6523      string             `json:"iso6523Code,omitempty"`
	ContactPoint []contactPointJSON `json:"contactPoint,omitempty"`
}

type nodeQuantityJSON struct {
	SchemaType string `json:"@type"`
	Value      int    `json:"value"`
}

type faqPageJSON struct {
	Context    string         `json:"@context"`
	ID         string         `json:"@id,omitempty"`
	SchemaType string         `json:"@type"`
	MainEntity []questionJSON `json:"mainEntity"`
}

type questionJSON struct {
	SchemaType     string     `json:"@type"`
	Name           string     `json:"name"`
	AcceptedAnswer answerJSON `json:"acceptedAnswer"`
}

type answerJSON struct {
	SchemaType string `json:"@type"`
	Text       string `json:"text"`
}

var _ keyed = NewsArticle{}
