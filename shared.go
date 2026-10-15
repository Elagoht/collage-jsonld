package jsonld

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// The pieces below are parts of nodes rather than nodes: an author, a price, a
// rating, a point on a map. AggregateRating and Review are both — a node of their
// own when a page is about one, and a property of a Product or a Recipe.

// Agent is whoever made, wrote, performed or organized something: a
// schema.org/Person, Organization or PerformingGroup, by name and, optionally, a
// URL.
type Agent struct {
	// Kind is the agent's schema.org type; empty is AgentPerson, except where a
	// field says otherwise — a hiring organization is an Organization unless told.
	Kind AgentKind
	Name string
	URL  string
	// SameAs are the agent's profiles elsewhere, which tie it to the same person
	// or organization on other sites.
	SameAs []string
	// LogoURL is an organization's logo; it is written as "logo".
	LogoURL string
}

// AgentKind is the schema.org type of an Agent.
type AgentKind string

const (
	AgentPerson          AgentKind = "Person"
	AgentOrganization    AgentKind = "Organization"
	AgentPerformingGroup AgentKind = "PerformingGroup"
)

type agentJSON struct {
	SchemaType string       `json:"@type"`
	Name       string       `json:"name"`
	URL        string       `json:"url,omitempty"`
	SameAs     []string     `json:"sameAs,omitempty"`
	Logo       *imageObject `json:"logo,omitempty"`
}

func (a Agent) wire() agentJSON {
	return a.wireAs(AgentPerson)
}

// wireAs is the agent's wire form with fallback as its type when Kind is empty.
func (a Agent) wireAs(fallback AgentKind) agentJSON {
	kind := a.Kind
	if kind == "" {
		kind = fallback
	}
	out := agentJSON{SchemaType: string(kind), Name: a.Name, URL: a.URL, SameAs: a.SameAs}
	if a.LogoURL != "" {
		out.Logo = &imageObject{SchemaType: "ImageObject", URL: a.LogoURL}
	}
	return out
}

// organizationAgent is a's wire form as an Organization unless Kind says
// otherwise, nil without a name.
func organizationAgent(a Agent) *agentJSON {
	if a.Name == "" {
		return nil
	}
	out := a.wireAs(AgentOrganization)
	return &out
}

// Money is an amount in a currency: schema.org/MonetaryAmount.
type Money struct {
	// Value is a string so that "0" is written; use a dot for decimals.
	Value string
	// Currency is an ISO 4217 code: "TRY".
	Currency string
}

type moneyJSON struct {
	SchemaType string `json:"@type"`
	Value      string `json:"value"`
	Currency   string `json:"currency,omitempty"`
}

// wire is nil for nil, so the property is left out.
func (m *Money) wire() *moneyJSON {
	if m == nil {
		return nil
	}
	return &moneyJSON{SchemaType: "MonetaryAmount", Value: m.Value, Currency: m.Currency}
}

// agents is the wire form of a list of agents, leaving out any without a name,
// and nil when none is left, so the property is left out.
func agents(list []Agent) []agentJSON {
	var out []agentJSON
	for _, a := range list {
		if a.Name != "" {
			out = append(out, a.wire())
		}
	}
	return out
}

// agentPtr is one agent's wire form, nil without a name.
func agentPtr(a Agent) *agentJSON {
	if a.Name == "" {
		return nil
	}
	out := a.wire()
	return &out
}

// Offer is a way to get something — a ticket, a product, a course, a stay:
// schema.org/Offer.
type Offer struct {
	URL string
	// Price is a string so that "0", something free, is written rather than
	// mistaken for no price. Use a dot for decimals: "149.90".
	Price string
	// Currency is an ISO 4217 code: "TRY".
	Currency        string
	Availability    Availability
	ItemCondition   ItemCondition
	ValidFrom       time.Time
	PriceValidUntil time.Time
	// Category is what kind of offer it is: "Free", "Paid", "Subscription".
	Category string
	// ShippingDetails and ReturnPolicy are this offer's, for a merchant listing. A
	// Product's own ShippingDetails and ReturnPolicy fill in for an offer that
	// sets none.
	ShippingDetails []ShippingDetails
	ReturnPolicy    *ReturnPolicy
}

func (o Offer) wire() offerJSON {
	out := offerJSON{
		SchemaType:    "Offer",
		URL:           o.URL,
		Price:         o.Price,
		PriceCurrency: o.Currency,
		Availability:  string(o.Availability),
		ItemCondition: string(o.ItemCondition),
		Category:      o.Category,
		ReturnPolicy:  o.ReturnPolicy.wire(),
	}
	for _, s := range o.ShippingDetails {
		out.ShippingDetails = append(out.ShippingDetails, s.wire())
	}
	if !o.ValidFrom.IsZero() {
		out.ValidFrom = o.ValidFrom.Format(time.RFC3339)
	}
	if !o.PriceValidUntil.IsZero() {
		out.PriceValidUntil = o.PriceValidUntil.Format(time.DateOnly)
	}
	return out
}

// offers is the wire form of a list of offers, nil for none.
func offers(list []Offer) []offerJSON {
	var out []offerJSON
	for _, o := range list {
		out = append(out, o.wire())
	}
	return out
}

type offerJSON struct {
	SchemaType      string                `json:"@type"`
	URL             string                `json:"url,omitempty"`
	Price           string                `json:"price,omitempty"`
	PriceCurrency   string                `json:"priceCurrency,omitempty"`
	Availability    string                `json:"availability,omitempty"`
	ItemCondition   string                `json:"itemCondition,omitempty"`
	ValidFrom       string                `json:"validFrom,omitempty"`
	PriceValidUntil string                `json:"priceValidUntil,omitempty"`
	Category        string                `json:"category,omitempty"`
	ShippingDetails []productShippingJSON `json:"shippingDetails,omitempty"`
	ReturnPolicy    *productReturnJSON    `json:"hasMerchantReturnPolicy,omitempty"`
}

// Availability is schema.org's ItemAvailability, written as its full IRI.
type Availability string

const (
	InStock             Availability = "https://schema.org/InStock"
	OutOfStock          Availability = "https://schema.org/OutOfStock"
	SoldOut             Availability = "https://schema.org/SoldOut"
	PreOrder            Availability = "https://schema.org/PreOrder"
	PreSale             Availability = "https://schema.org/PreSale"
	BackOrder           Availability = "https://schema.org/BackOrder"
	LimitedAvailability Availability = "https://schema.org/LimitedAvailability"
	OnlineOnly          Availability = "https://schema.org/OnlineOnly"
	InStoreOnly         Availability = "https://schema.org/InStoreOnly"
	Discontinued        Availability = "https://schema.org/Discontinued"
)

// ItemCondition is schema.org's OfferItemCondition, written as its full IRI.
type ItemCondition string

const (
	NewCondition         ItemCondition = "https://schema.org/NewCondition"
	UsedCondition        ItemCondition = "https://schema.org/UsedCondition"
	RefurbishedCondition ItemCondition = "https://schema.org/RefurbishedCondition"
	DamagedCondition     ItemCondition = "https://schema.org/DamagedCondition"
)

// GeoCoordinates is a point on a map: schema.org/GeoCoordinates. Both zero is
// read as no point and left out; Null Island has no businesses.
type GeoCoordinates struct {
	Latitude  float64
	Longitude float64
}

type geoJSON struct {
	SchemaType string  `json:"@type"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
}

func (g GeoCoordinates) wire() *geoJSON {
	if g.Latitude == 0 && g.Longitude == 0 {
		return nil
	}
	return &geoJSON{SchemaType: "GeoCoordinates", Latitude: g.Latitude, Longitude: g.Longitude}
}

// Rating is one reviewer's rating: schema.org/Rating. BestRating and WorstRating
// are left out at zero, which consumers read as 5 and 1.
type Rating struct {
	Value       float64
	BestRating  float64
	WorstRating float64
}

type ratingJSON struct {
	SchemaType  string  `json:"@type"`
	RatingValue float64 `json:"ratingValue"`
	BestRating  float64 `json:"bestRating,omitempty"`
	WorstRating float64 `json:"worstRating,omitempty"`
}

func (r *Rating) wire() *ratingJSON {
	if r == nil {
		return nil
	}
	return &ratingJSON{SchemaType: "Rating", RatingValue: r.Value, BestRating: r.BestRating, WorstRating: r.WorstRating}
}

// AggregateRating is the average of many ratings: schema.org/AggregateRating.
//
// As a property of a Product, a Recipe or a LocalBusiness it rates that, and
// ItemReviewed is left empty. Emitted on its own it needs ItemReviewed, the
// thing it rates.
type AggregateRating struct {
	// ID and Key are as on Article.
	ID           string
	Key          string
	ItemReviewed Node
	Value        float64
	// RatingCount is how many ratings there are, ReviewCount how many of them
	// came with a review. One of the two is required.
	RatingCount int
	ReviewCount int
	BestRating  float64
	WorstRating float64
}

func (AggregateRating) Type() string { return "AggregateRating" }

func (a AggregateRating) nodeKey() string { return a.Key }

func (a AggregateRating) MarshalJSON() ([]byte, error) {
	out := a.wire()
	out.Context = schemaContext
	out.ID = a.ID
	if a.ItemReviewed != nil {
		item, err := nested(a.ItemReviewed)
		if err != nil {
			return nil, err
		}
		out.ItemReviewed = item
	}
	return json.Marshal(out)
}

func (a AggregateRating) wire() aggregateRatingJSON {
	return aggregateRatingJSON{
		SchemaType:  "AggregateRating",
		RatingValue: a.Value,
		RatingCount: a.RatingCount,
		ReviewCount: a.ReviewCount,
		BestRating:  a.BestRating,
		WorstRating: a.WorstRating,
	}
}

// aggregate is the embedded form, nil for nil, so the property is left out.
func aggregate(a *AggregateRating) *aggregateRatingJSON {
	if a == nil {
		return nil
	}
	out := a.wire()
	return &out
}

type aggregateRatingJSON struct {
	Context      string          `json:"@context,omitempty"`
	ID           string          `json:"@id,omitempty"`
	SchemaType   string          `json:"@type"`
	ItemReviewed json.RawMessage `json:"itemReviewed,omitempty"`
	RatingValue  float64         `json:"ratingValue"`
	RatingCount  int             `json:"ratingCount,omitempty"`
	ReviewCount  int             `json:"reviewCount,omitempty"`
	BestRating   float64         `json:"bestRating,omitempty"`
	WorstRating  float64         `json:"worstRating,omitempty"`
}

// Review is one person's review of something: schema.org/Review.
//
// As a property of a Product, a Recipe or a LocalBusiness it reviews that, and
// ItemReviewed is left empty. Emitted on its own it needs ItemReviewed.
type Review struct {
	// ID and Key are as on Article.
	ID            string
	Key           string
	ItemReviewed  Node
	Author        Agent
	Rating        *Rating
	Name          string
	Body          string
	DatePublished time.Time
	// Publisher is the site the review appeared on, when it is not the author's.
	Publisher Agent
	// ContentReferenceTime is when the reviewed experience happened — the stay,
	// the meal — which a vacation rental's reviews must carry in some countries.
	ContentReferenceTime time.Time
}

func (Review) Type() string { return "Review" }

func (r Review) nodeKey() string { return r.Key }

func (r Review) MarshalJSON() ([]byte, error) {
	out := r.wire()
	out.Context = schemaContext
	out.ID = r.ID
	if r.ItemReviewed != nil {
		item, err := nested(r.ItemReviewed)
		if err != nil {
			return nil, err
		}
		out.ItemReviewed = item
	}
	return json.Marshal(out)
}

func (r Review) wire() reviewJSON {
	out := reviewJSON{
		SchemaType:   "Review",
		Name:         r.Name,
		Author:       agentPtr(r.Author),
		ReviewRating: r.Rating.wire(),
		ReviewBody:   r.Body,
		Publisher:    agentPtr(r.Publisher),

		ContentReferenceTime: dateTime(r.ContentReferenceTime),
	}
	if !r.DatePublished.IsZero() {
		out.DatePublished = r.DatePublished.Format(time.RFC3339)
	}
	return out
}

// reviews is the embedded form of a list of reviews, nil for none.
func reviews(list []Review) []reviewJSON {
	var out []reviewJSON
	for _, r := range list {
		out = append(out, r.wire())
	}
	return out
}

type reviewJSON struct {
	Context       string          `json:"@context,omitempty"`
	ID            string          `json:"@id,omitempty"`
	SchemaType    string          `json:"@type"`
	ItemReviewed  json.RawMessage `json:"itemReviewed,omitempty"`
	Name          string          `json:"name,omitempty"`
	Author        *agentJSON      `json:"author,omitempty"`
	ReviewRating  *ratingJSON     `json:"reviewRating,omitempty"`
	ReviewBody    string          `json:"reviewBody,omitempty"`
	DatePublished string          `json:"datePublished,omitempty"`
	Publisher     *agentJSON      `json:"publisher,omitempty"`

	ContentReferenceTime string `json:"contentReferenceTime,omitempty"`
}

// nested is node marshalled for use inside another node: the same object without
// its "@context", which only a top-level node declares. Its other members keep
// their order.
//
// A nil node, a typed nil among them, is no node: nested returns nil and the
// property is left out. Its type is never asked for, because asking a typed nil
// panics.
func nested(node Node) (json.RawMessage, error) {
	if isNil(node) {
		return nil, nil
	}
	encoded, err := json.Marshal(node)
	if err != nil {
		return nil, err
	}
	if bytes.Equal(encoded, []byte("null")) {
		return nil, nil
	}
	dec := json.NewDecoder(bytes.NewReader(encoded))
	if open, err := dec.Token(); err != nil || open != json.Delim('{') {
		return nil, fmt.Errorf("jsonld: nested %T node is not a JSON object", node)
	}
	var out bytes.Buffer
	out.WriteByte('{')
	first := true
	for dec.More() {
		keyToken, err := dec.Token()
		if err != nil {
			return nil, err
		}
		key, _ := keyToken.(string)
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
		if !first {
			out.WriteByte(',')
		}
		first = false
		out.Write(encodedKey)
		out.WriteByte(':')
		out.Write(value)
	}
	out.WriteByte('}')
	return out.Bytes(), nil
}

// isoDuration is d as an ISO 8601 duration, "PT1H30M", the form schema.org's
// Duration takes; empty for zero or less, so the property is left out.
func isoDuration(d time.Duration) string {
	if d <= 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("PT")
	if h := d / time.Hour; h > 0 {
		b.WriteString(strconv.FormatInt(int64(h), 10) + "H")
		d -= h * time.Hour
	}
	if m := d / time.Minute; m > 0 {
		b.WriteString(strconv.FormatInt(int64(m), 10) + "M")
		d -= m * time.Minute
	}
	if d > 0 {
		b.WriteString(strconv.FormatFloat(d.Seconds(), 'f', -1, 64) + "S")
	}
	return b.String()
}

// dateTime is t in RFC 3339, empty for the zero time, so the property is left
// out.
func dateTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

// date is t as a calendar date, empty for the zero time.
func date(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.DateOnly)
}

var (
	_ keyed = AggregateRating{}
	_ keyed = Review{}
)
