package jsonld

import (
	"encoding/json"
	"time"
)

// Product is something for sale: schema.org/Product. It covers what search
// engines read for a product snippet and for a merchant listing.
//
// ShippingDetails and ReturnPolicy describe the offers rather than the product,
// and that is where schema.org puts them. Here they are the product's defaults:
// each is written into every one of Offers that sets no Shipping or
// ReturnPolicy of its own. A product without offers has nowhere to put them, and
// they are left out.
type Product struct {
	// ID and Key are as on Article.
	ID          string
	Key         string
	Name        string
	Description string
	// ImageURLs are the product's images; search engines prefer several, in
	// different aspect ratios.
	ImageURLs []string
	// Brand is the brand's name.
	Brand string
	SKU   string
	// GTIN is the product's barcode number, whichever length: GTIN-8, -12, -13
	// or -14.
	GTIN            string
	MPN             string
	Color           string
	Material        string
	Size            string
	Offers          []Offer
	ShippingDetails []ShippingDetails
	ReturnPolicy    *ReturnPolicy
	AggregateRating *AggregateRating
	Reviews         []Review
}

func (Product) Type() string { return "Product" }

func (p Product) nodeKey() string { return p.Key }

func (p Product) MarshalJSON() ([]byte, error) {
	out := productJSON{
		Context:         schemaContext,
		ID:              p.ID,
		SchemaType:      "Product",
		Name:            p.Name,
		Description:     p.Description,
		Image:           p.ImageURLs,
		SKU:             p.SKU,
		GTIN:            p.GTIN,
		MPN:             p.MPN,
		Color:           p.Color,
		Material:        p.Material,
		Size:            p.Size,
		AggregateRating: aggregate(p.AggregateRating),
		Review:          reviews(p.Reviews),
	}
	if p.Brand != "" {
		out.Brand = &productBrandJSON{SchemaType: "Brand", Name: p.Brand}
	}
	var shipping []productShippingJSON
	for _, s := range p.ShippingDetails {
		shipping = append(shipping, s.wire())
	}
	returns := p.ReturnPolicy.wire()
	for _, o := range p.Offers {
		offer := o.wire()
		if offer.ShippingDetails == nil {
			offer.ShippingDetails = shipping
		}
		if offer.ReturnPolicy == nil {
			offer.ReturnPolicy = returns
		}
		out.Offers = append(out.Offers, offer)
	}
	return json.Marshal(out)
}

// ShippingDetails is what shipping costs and how long it takes to one country:
// schema.org/OfferShippingDetails.
type ShippingDetails struct {
	// Rate is the shipping cost; a Value of "0" is free shipping.
	Rate *Money
	// Country is the destination, an ISO 3166-1 alpha-2 code: "TR".
	Country string
	// HandlingDays and TransitDays are the least and the most business days the
	// order is handled and in transit. Nil is left out; {0, 0} is the same day.
	HandlingDays *DayRange
	TransitDays  *DayRange
}

// DayRange is a least and a most number of days.
type DayRange struct {
	Min int
	Max int
}

func (s ShippingDetails) wire() productShippingJSON {
	out := productShippingJSON{SchemaType: "OfferShippingDetails"}
	out.ShippingRate = s.Rate.wire()
	if s.Country != "" {
		out.ShippingDestination = &productRegionJSON{SchemaType: "DefinedRegion", AddressCountry: s.Country}
	}
	handling, transit := productDays(s.HandlingDays), productDays(s.TransitDays)
	if handling != nil || transit != nil {
		out.DeliveryTime = &productDeliveryJSON{SchemaType: "ShippingDeliveryTime", HandlingTime: handling, TransitTime: transit}
	}
	return out
}

// productDays is a min/max range of days, nil for nil.
func productDays(days *DayRange) *productQuantityJSON {
	if days == nil {
		return nil
	}
	return &productQuantityJSON{SchemaType: "QuantitativeValue", MinValue: days.Min, MaxValue: days.Max, UnitCode: "DAY"}
}

// ReturnPolicy is how a product can be returned: schema.org/MerchantReturnPolicy.
type ReturnPolicy struct {
	// Country is where the policy applies, an ISO 3166-1 alpha-2 code: "TR".
	Country  string
	Category ReturnPolicyCategory
	// Days is the return window, for ReturnFiniteWindow.
	Days   int
	Method ReturnMethod
	Fees   ReturnFees
	// ShippingFee is what returning costs the customer, for ReturnShippingFees.
	ShippingFee *Money
}

func (r *ReturnPolicy) wire() *productReturnJSON {
	if r == nil {
		return nil
	}
	return &productReturnJSON{
		SchemaType:           "MerchantReturnPolicy",
		ApplicableCountry:    r.Country,
		ReturnPolicyCategory: string(r.Category),
		MerchantReturnDays:   r.Days,
		ReturnMethod:         string(r.Method),
		ReturnFees:           string(r.Fees),
		ShippingFeesAmount:   r.ShippingFee.wire(),
	}
}

// ReturnPolicyCategory is schema.org's MerchantReturnEnumeration, written as its
// full IRI.
type ReturnPolicyCategory string

const (
	ReturnFiniteWindow    ReturnPolicyCategory = "https://schema.org/MerchantReturnFiniteReturnWindow"
	ReturnNotPermitted    ReturnPolicyCategory = "https://schema.org/MerchantReturnNotPermitted"
	ReturnUnlimitedWindow ReturnPolicyCategory = "https://schema.org/MerchantReturnUnlimitedWindow"
)

// ReturnMethod is schema.org's ReturnMethodEnumeration, written as its full IRI.
type ReturnMethod string

const (
	ReturnByMail  ReturnMethod = "https://schema.org/ReturnByMail"
	ReturnInStore ReturnMethod = "https://schema.org/ReturnInStore"
	ReturnAtKiosk ReturnMethod = "https://schema.org/ReturnAtKiosk"
)

// ReturnFees is schema.org's ReturnFeesEnumeration, written as its full IRI.
type ReturnFees string

const (
	FreeReturn                       ReturnFees = "https://schema.org/FreeReturn"
	ReturnFeesCustomerResponsibility ReturnFees = "https://schema.org/ReturnFeesCustomerResponsibility"
	ReturnShippingFees               ReturnFees = "https://schema.org/ReturnShippingFees"
)

// SoftwareApplication is an app, a program or a game:
// schema.org/SoftwareApplication, or the more specific type its Kind names.
type SoftwareApplication struct {
	// Kind is the schema.org type; empty is AppSoftware.
	Kind AppKind
	// ID and Key are as on Article.
	ID              string
	Key             string
	Name            string
	Description     string
	URL             string
	ImageURLs       []string
	OperatingSystem string
	// ApplicationCategory is schema.org's category name: "GameApplication",
	// "BusinessApplication".
	ApplicationCategory string
	Offers              []Offer
	AggregateRating     *AggregateRating
	Reviews             []Review
}

// AppKind is the schema.org type of a SoftwareApplication.
type AppKind string

const (
	AppSoftware  AppKind = "SoftwareApplication"
	AppMobile    AppKind = "MobileApplication"
	AppWeb       AppKind = "WebApplication"
	AppVideoGame AppKind = "VideoGame"
)

func (s SoftwareApplication) Type() string {
	if s.Kind == "" {
		return string(AppSoftware)
	}
	return string(s.Kind)
}

func (s SoftwareApplication) nodeKey() string { return s.Key }

func (s SoftwareApplication) MarshalJSON() ([]byte, error) {
	out := productSoftwareJSON{
		Context:             schemaContext,
		ID:                  s.ID,
		SchemaType:          s.Type(),
		Name:                s.Name,
		Description:         s.Description,
		URL:                 s.URL,
		OperatingSystem:     s.OperatingSystem,
		ApplicationCategory: s.ApplicationCategory,
		Offers:              offers(s.Offers),
		AggregateRating:     aggregate(s.AggregateRating),
		Review:              reviews(s.Reviews),
	}
	out.Image = s.ImageURLs
	return json.Marshal(out)
}

// Movie is a film: schema.org/Movie.
type Movie struct {
	// ID and Key are as on Article.
	ID              string
	Key             string
	Name            string
	URL             string
	ImageURLs       []string
	DateCreated     time.Time
	Director        Agent
	Actors          []Agent
	Duration        time.Duration
	AggregateRating *AggregateRating
	Reviews         []Review
}

func (Movie) Type() string { return "Movie" }

func (m Movie) nodeKey() string { return m.Key }

func (m Movie) MarshalJSON() ([]byte, error) {
	out := productMovieJSON{
		Context:         schemaContext,
		ID:              m.ID,
		SchemaType:      "Movie",
		Name:            m.Name,
		URL:             m.URL,
		DateCreated:     date(m.DateCreated),
		Director:        agentPtr(m.Director),
		Actor:           agents(m.Actors),
		Duration:        isoDuration(m.Duration),
		AggregateRating: aggregate(m.AggregateRating),
		Review:          reviews(m.Reviews),
	}
	out.Image = m.ImageURLs
	return json.Marshal(out)
}

// EmployerAggregateRating is how employees rate working for an organization:
// schema.org/EmployerAggregateRating.
type EmployerAggregateRating struct {
	// ID and Key are as on Article.
	ID  string
	Key string
	// Employer is the organization rated; an empty Kind is AgentOrganization.
	Employer    Agent
	Value       float64
	RatingCount int
	ReviewCount int
	BestRating  float64
	WorstRating float64
}

func (EmployerAggregateRating) Type() string { return "EmployerAggregateRating" }

func (e EmployerAggregateRating) nodeKey() string { return e.Key }

func (e EmployerAggregateRating) MarshalJSON() ([]byte, error) {
	out := productEmployerRatingJSON{
		Context:     schemaContext,
		ID:          e.ID,
		SchemaType:  "EmployerAggregateRating",
		RatingValue: e.Value,
		RatingCount: e.RatingCount,
		ReviewCount: e.ReviewCount,
		BestRating:  e.BestRating,
		WorstRating: e.WorstRating,
	}
	out.ItemReviewed = organizationAgent(e.Employer)
	return json.Marshal(out)
}

var (
	_ keyed = Product{}
	_ keyed = SoftwareApplication{}
	_ keyed = Movie{}
	_ keyed = EmployerAggregateRating{}
)

type productJSON struct {
	Context         string               `json:"@context"`
	ID              string               `json:"@id,omitempty"`
	SchemaType      string               `json:"@type"`
	Name            string               `json:"name"`
	Description     string               `json:"description,omitempty"`
	Image           []string             `json:"image,omitempty"`
	Brand           *productBrandJSON    `json:"brand,omitempty"`
	SKU             string               `json:"sku,omitempty"`
	GTIN            string               `json:"gtin,omitempty"`
	MPN             string               `json:"mpn,omitempty"`
	Color           string               `json:"color,omitempty"`
	Material        string               `json:"material,omitempty"`
	Size            string               `json:"size,omitempty"`
	Offers          []offerJSON          `json:"offers,omitempty"`
	AggregateRating *aggregateRatingJSON `json:"aggregateRating,omitempty"`
	Review          []reviewJSON         `json:"review,omitempty"`
}

type productBrandJSON struct {
	SchemaType string `json:"@type"`
	Name       string `json:"name"`
}

type productShippingJSON struct {
	SchemaType          string               `json:"@type"`
	ShippingRate        *moneyJSON           `json:"shippingRate,omitempty"`
	ShippingDestination *productRegionJSON   `json:"shippingDestination,omitempty"`
	DeliveryTime        *productDeliveryJSON `json:"deliveryTime,omitempty"`
}

type productRegionJSON struct {
	SchemaType     string `json:"@type"`
	AddressCountry string `json:"addressCountry"`
}

type productDeliveryJSON struct {
	SchemaType   string               `json:"@type"`
	HandlingTime *productQuantityJSON `json:"handlingTime,omitempty"`
	TransitTime  *productQuantityJSON `json:"transitTime,omitempty"`
}

type productQuantityJSON struct {
	SchemaType string `json:"@type"`
	MinValue   int    `json:"minValue"`
	MaxValue   int    `json:"maxValue"`
	UnitCode   string `json:"unitCode"`
}

type productReturnJSON struct {
	SchemaType           string `json:"@type"`
	ApplicableCountry    string `json:"applicableCountry,omitempty"`
	ReturnPolicyCategory string `json:"returnPolicyCategory,omitempty"`
	MerchantReturnDays   int    `json:"merchantReturnDays,omitempty"`
	ReturnMethod         string `json:"returnMethod,omitempty"`
	ReturnFees           string `json:"returnFees,omitempty"`

	ShippingFeesAmount *moneyJSON `json:"returnShippingFeesAmount,omitempty"`
}

type productSoftwareJSON struct {
	Context             string               `json:"@context"`
	ID                  string               `json:"@id,omitempty"`
	SchemaType          string               `json:"@type"`
	Name                string               `json:"name"`
	Description         string               `json:"description,omitempty"`
	URL                 string               `json:"url,omitempty"`
	Image               []string             `json:"image,omitempty"`
	OperatingSystem     string               `json:"operatingSystem,omitempty"`
	ApplicationCategory string               `json:"applicationCategory,omitempty"`
	Offers              []offerJSON          `json:"offers,omitempty"`
	AggregateRating     *aggregateRatingJSON `json:"aggregateRating,omitempty"`
	Review              []reviewJSON         `json:"review,omitempty"`
}

type productMovieJSON struct {
	Context         string               `json:"@context"`
	ID              string               `json:"@id,omitempty"`
	SchemaType      string               `json:"@type"`
	Name            string               `json:"name"`
	URL             string               `json:"url,omitempty"`
	Image           []string             `json:"image,omitempty"`
	DateCreated     string               `json:"dateCreated,omitempty"`
	Director        *agentJSON           `json:"director,omitempty"`
	Actor           []agentJSON          `json:"actor,omitempty"`
	Duration        string               `json:"duration,omitempty"`
	AggregateRating *aggregateRatingJSON `json:"aggregateRating,omitempty"`
	Review          []reviewJSON         `json:"review,omitempty"`
}

type productEmployerRatingJSON struct {
	Context      string     `json:"@context"`
	ID           string     `json:"@id,omitempty"`
	SchemaType   string     `json:"@type"`
	ItemReviewed *agentJSON `json:"itemReviewed,omitempty"`
	RatingValue  float64    `json:"ratingValue"`
	RatingCount  int        `json:"ratingCount,omitempty"`
	ReviewCount  int        `json:"reviewCount,omitempty"`
	BestRating   float64    `json:"bestRating,omitempty"`
	WorstRating  float64    `json:"worstRating,omitempty"`
}
