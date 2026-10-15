package jsonld

import (
	"encoding/json"
	"time"
)

// Event is something that happens at a time and a place, or online:
// schema.org/Event.
//
// Search engines want a name, a start date and a location before they show an
// event; the others refine it. Times are written with their own offset. A
// wall-clock time read without a zone — "2026-11-07 19:00" parsed as UTC — is
// three hours off in Istanbul; parse it in the event's location instead.
type Event struct {
	// ID and Key are as on Article.
	ID          string
	Key         string
	Name        string
	Description string
	URL         string
	ImageURL    string
	StartDate   time.Time
	EndDate     time.Time
	// Status is left out when empty, which consumers read as EventScheduled.
	Status EventStatus
	// AttendanceMode is left out when empty. A mixed event wants both a Place and
	// an OnlineURL.
	AttendanceMode AttendanceMode
	// Place is where the event happens in person.
	Place *Place
	// OnlineURL is where the event happens online, written as a VirtualLocation.
	// With a Place as well, the location is both.
	OnlineURL     string
	OrganizerName string
	OrganizerURL  string
	Offers        []Offer
}

func (Event) Type() string { return "Event" }

func (e Event) nodeKey() string { return e.Key }

func (e Event) MarshalJSON() ([]byte, error) {
	out := eventJSON{
		Context:             schemaContext,
		ID:                  e.ID,
		SchemaType:          "Event",
		Name:                e.Name,
		Description:         e.Description,
		URL:                 e.URL,
		EventStatus:         string(e.Status),
		EventAttendanceMode: string(e.AttendanceMode),
		Organizer:           organizerNode(e.OrganizerName, e.OrganizerURL),
	}
	if !e.StartDate.IsZero() {
		out.StartDate = e.StartDate.Format(time.RFC3339)
	}
	if !e.EndDate.IsZero() {
		out.EndDate = e.EndDate.Format(time.RFC3339)
	}
	if e.ImageURL != "" {
		out.Image = &imageObject{SchemaType: "ImageObject", URL: e.ImageURL}
	}

	// One location is an object and two are an array: schema.org takes either,
	// and a single place written as a one-element array reads oddly in a page.
	var locations []json.RawMessage
	if e.Place != nil {
		place, err := json.Marshal(e.Place.wire())
		if err != nil {
			return nil, err
		}
		locations = append(locations, place)
	}
	if e.OnlineURL != "" {
		online, err := json.Marshal(virtualLocation{SchemaType: "VirtualLocation", URL: e.OnlineURL})
		if err != nil {
			return nil, err
		}
		locations = append(locations, online)
	}
	switch len(locations) {
	case 1:
		out.Location = locations[0]
	case 2:
		both, err := json.Marshal(locations)
		if err != nil {
			return nil, err
		}
		out.Location = both
	}

	for _, offer := range e.Offers {
		out.Offers = append(out.Offers, offer.wire())
	}
	return json.Marshal(out)
}

// EventStatus is schema.org's EventStatusType, written as its full IRI.
type EventStatus string

const (
	EventScheduled   EventStatus = "https://schema.org/EventScheduled"
	EventCancelled   EventStatus = "https://schema.org/EventCancelled"
	EventPostponed   EventStatus = "https://schema.org/EventPostponed"
	EventRescheduled EventStatus = "https://schema.org/EventRescheduled"
	EventMovedOnline EventStatus = "https://schema.org/EventMovedOnline"
)

// AttendanceMode is schema.org's EventAttendanceModeEnumeration, written as its
// full IRI.
type AttendanceMode string

const (
	AttendanceOffline AttendanceMode = "https://schema.org/OfflineEventAttendanceMode"
	AttendanceOnline  AttendanceMode = "https://schema.org/OnlineEventAttendanceMode"
	AttendanceMixed   AttendanceMode = "https://schema.org/MixedEventAttendanceMode"
)

// Availability is schema.org's ItemAvailability for an Offer, written as its
// full IRI.
type Availability string

const (
	InStock  Availability = "https://schema.org/InStock"
	SoldOut  Availability = "https://schema.org/SoldOut"
	PreOrder Availability = "https://schema.org/PreOrder"
)

// Place is a physical location: schema.org/Place. It is part of an Event, not a
// node of its own.
type Place struct {
	Name    string
	Address *PostalAddress
}

func (p Place) wire() placeJSON {
	return placeJSON{SchemaType: "Place", Name: p.Name, Address: p.Address.wire()}
}

// PostalAddress is a mailing address: schema.org/PostalAddress. It is part of a
// Place or an Organization, not a node of its own.
type PostalAddress struct {
	StreetAddress string
	// Locality is the city or town.
	Locality string
	// Region is the state, province or, in Turkey, the il.
	Region     string
	PostalCode string
	// Country is an ISO 3166-1 alpha-2 code, "TR", or the country's name.
	Country string
}

// wire is nil for a nil address, so the property is left out.
func (a *PostalAddress) wire() *postalAddressJSON {
	if a == nil {
		return nil
	}
	return &postalAddressJSON{
		SchemaType:      "PostalAddress",
		StreetAddress:   a.StreetAddress,
		AddressLocality: a.Locality,
		AddressRegion:   a.Region,
		PostalCode:      a.PostalCode,
		AddressCountry:  a.Country,
	}
}

// Offer is a way to attend an Event — a ticket, a registration:
// schema.org/Offer.
type Offer struct {
	URL string
	// Price is a string so that "0", a free event, is written rather than
	// mistaken for no price. Use a dot for decimals: "149.90".
	Price string
	// Currency is an ISO 4217 code: "TRY".
	Currency     string
	Availability Availability
	ValidFrom    time.Time
}

func (o Offer) wire() offerJSON {
	out := offerJSON{
		SchemaType:    "Offer",
		URL:           o.URL,
		Price:         o.Price,
		PriceCurrency: o.Currency,
		Availability:  string(o.Availability),
	}
	if !o.ValidFrom.IsZero() {
		out.ValidFrom = o.ValidFrom.Format(time.RFC3339)
	}
	return out
}

// organizerNode is an event's organizer, flattened into name fields as Article's
// author is, and nil when the name is empty.
func organizerNode(name, url string) *organization {
	if name == "" {
		return nil
	}
	return &organization{SchemaType: "Organization", Name: name, URL: url}
}

type eventJSON struct {
	Context             string          `json:"@context"`
	ID                  string          `json:"@id,omitempty"`
	SchemaType          string          `json:"@type"`
	Name                string          `json:"name"`
	Description         string          `json:"description,omitempty"`
	URL                 string          `json:"url,omitempty"`
	Image               *imageObject    `json:"image,omitempty"`
	StartDate           string          `json:"startDate,omitempty"`
	EndDate             string          `json:"endDate,omitempty"`
	EventStatus         string          `json:"eventStatus,omitempty"`
	EventAttendanceMode string          `json:"eventAttendanceMode,omitempty"`
	Location            json.RawMessage `json:"location,omitempty"`
	Organizer           *organization   `json:"organizer,omitempty"`
	Offers              []offerJSON     `json:"offers,omitempty"`
}

type placeJSON struct {
	SchemaType string             `json:"@type"`
	Name       string             `json:"name,omitempty"`
	Address    *postalAddressJSON `json:"address,omitempty"`
}

type virtualLocation struct {
	SchemaType string `json:"@type"`
	URL        string `json:"url"`
}

type postalAddressJSON struct {
	SchemaType      string `json:"@type"`
	StreetAddress   string `json:"streetAddress,omitempty"`
	AddressLocality string `json:"addressLocality,omitempty"`
	AddressRegion   string `json:"addressRegion,omitempty"`
	PostalCode      string `json:"postalCode,omitempty"`
	AddressCountry  string `json:"addressCountry,omitempty"`
}

type offerJSON struct {
	SchemaType    string `json:"@type"`
	URL           string `json:"url,omitempty"`
	Price         string `json:"price,omitempty"`
	PriceCurrency string `json:"priceCurrency,omitempty"`
	Availability  string `json:"availability,omitempty"`
	ValidFrom     string `json:"validFrom,omitempty"`
}
