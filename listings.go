package jsonld

import (
	"encoding/json"
	"time"
)

// The nodes in this file describe things a page lists for someone to visit, apply
// to, answer or book: a business, a job, a forum thread, a question, a stay.

// LocalBusiness is a business with a place people go to: schema.org/LocalBusiness,
// or one of its more specific types.
//
// Google wants a name and an address before it shows one; the rest refines it.
type LocalBusiness struct {
	// Kind is the schema.org type, any of LocalBusiness's subtypes:
	// "Restaurant", "Store", "Dentist". Empty is "LocalBusiness". The more
	// specific type is the more useful one.
	Kind BusinessKind
	// ID and Key are as on Article.
	ID          string
	Key         string
	Name        string
	Description string
	URL         string
	// ImageURLs are photos of the business. Several, in different aspect ratios,
	// are what Google asks for.
	ImageURLs []string
	LogoURL   string
	// Telephone is written as given; the international form, "+90 212 …", is
	// the one consumers read without guessing.
	Telephone string
	Email     string
	Address   *PostalAddress
	Geo       GeoCoordinates
	// PriceRange is relative cost as the business states it: "$$", "₺₺", or a
	// range such as "100-300 TRY".
	PriceRange   string
	OpeningHours []OpeningHours
	// SameAs are the business's profiles elsewhere.
	SameAs []string
	// Menu is the URL of the menu, for a business that serves food.
	Menu          string
	ServesCuisine []string
	// AcceptsReservations is written only when true.
	AcceptsReservations bool
	AggregateRating     *AggregateRating
	Reviews             []Review
}

// OpeningHours is when a business is open on some days of the week:
// schema.org/OpeningHoursSpecification.
//
// A day open in two stretches is two entries. A business open around the clock
// opens at "00:00" and closes at "23:59"; one closed all day opens and closes at
// "00:00". ValidFrom and ValidThrough limit an entry to a period, which is how
// holiday hours are written.
type OpeningHours struct {
	Days []time.Weekday
	// Opens and Closes are local wall-clock times: "09:00", "18:30".
	Opens        string
	Closes       string
	ValidFrom    time.Time
	ValidThrough time.Time
}

// BusinessKind is the schema.org type of a LocalBusiness. It is free-form,
// because schema.org has over a hundred subtypes: "Restaurant", "Store",
// "Dentist", "MovieTheater".
type BusinessKind string

func (b LocalBusiness) Type() string {
	if b.Kind == "" {
		return "LocalBusiness"
	}
	return string(b.Kind)
}

func (b LocalBusiness) nodeKey() string { return b.Key }

func (b LocalBusiness) MarshalJSON() ([]byte, error) {
	out := listingLocalBusinessJSON{
		Context:         schemaContext,
		ID:              b.ID,
		SchemaType:      b.Type(),
		Name:            b.Name,
		Description:     b.Description,
		URL:             b.URL,
		Image:           b.ImageURLs,
		Telephone:       b.Telephone,
		Email:           b.Email,
		Address:         b.Address.wire(),
		Geo:             b.Geo.wire(),
		PriceRange:      b.PriceRange,
		SameAs:          b.SameAs,
		Menu:            b.Menu,
		ServesCuisine:   b.ServesCuisine,
		AggregateRating: aggregate(b.AggregateRating),
		Review:          reviews(b.Reviews),
	}
	if b.LogoURL != "" {
		out.Logo = &imageObject{SchemaType: "ImageObject", URL: b.LogoURL}
	}
	if b.AcceptsReservations {
		out.AcceptsReservations = true
	}
	for _, hours := range b.OpeningHours {
		out.OpeningHoursSpecification = append(out.OpeningHoursSpecification, hours.wire())
	}
	return json.Marshal(out)
}

func (h OpeningHours) wire() listingOpeningHoursJSON {
	out := listingOpeningHoursJSON{
		SchemaType:   "OpeningHoursSpecification",
		Opens:        h.Opens,
		Closes:       h.Closes,
		ValidFrom:    date(h.ValidFrom),
		ValidThrough: date(h.ValidThrough),
	}
	for _, day := range h.Days {
		// time.Weekday's String is the English day name, which is schema.org's
		// DayOfWeek member name too.
		out.DayOfWeek = append(out.DayOfWeek, day.String())
	}
	return out
}

type listingLocalBusinessJSON struct {
	Context                   string                    `json:"@context"`
	ID                        string                    `json:"@id,omitempty"`
	SchemaType                string                    `json:"@type"`
	Name                      string                    `json:"name"`
	Description               string                    `json:"description,omitempty"`
	URL                       string                    `json:"url,omitempty"`
	Image                     []string                  `json:"image,omitempty"`
	Logo                      *imageObject              `json:"logo,omitempty"`
	Telephone                 string                    `json:"telephone,omitempty"`
	Email                     string                    `json:"email,omitempty"`
	Address                   *postalAddressJSON        `json:"address,omitempty"`
	Geo                       *geoJSON                  `json:"geo,omitempty"`
	PriceRange                string                    `json:"priceRange,omitempty"`
	OpeningHoursSpecification []listingOpeningHoursJSON `json:"openingHoursSpecification,omitempty"`
	SameAs                    []string                  `json:"sameAs,omitempty"`
	Menu                      string                    `json:"menu,omitempty"`
	ServesCuisine             []string                  `json:"servesCuisine,omitempty"`
	AcceptsReservations       bool                      `json:"acceptsReservations,omitempty"`
	AggregateRating           *aggregateRatingJSON      `json:"aggregateRating,omitempty"`
	Review                    []reviewJSON              `json:"review,omitempty"`
}

type listingOpeningHoursJSON struct {
	SchemaType   string   `json:"@type"`
	DayOfWeek    []string `json:"dayOfWeek,omitempty"`
	Opens        string   `json:"opens,omitempty"`
	Closes       string   `json:"closes,omitempty"`
	ValidFrom    string   `json:"validFrom,omitempty"`
	ValidThrough string   `json:"validThrough,omitempty"`
}

// JobPosting is a job opening: schema.org/JobPosting.
//
// Google wants a title, a description, a posting date, the hiring organization
// and a location — a Locations entry, or Remote with ApplicantCountries — before
// it shows one.
type JobPosting struct {
	// ID and Key are as on Article.
	ID    string
	Key   string
	Title string
	// Description is the full description of the job. It may hold HTML — <p>,
	// <ul>, <li>, <br> — which consumers render; it is still escaped for the
	// script element it is written into, like every other string here.
	Description  string
	DatePosted   time.Time
	ValidThrough time.Time
	// EmploymentType is one or more of the Employment constants.
	EmploymentType []EmploymentType
	// HiringOrganization is the employer, with its logo and profiles; its Kind
	// defaults to AgentOrganization here rather than to a person.
	HiringOrganization Agent
	// Locations are where the work is done. One is written as an object, more as
	// a list.
	Locations []PostalAddress
	// Remote marks a fully remote job ("TELECOMMUTE"). Google then wants
	// ApplicantCountries, the countries applicants may work from.
	Remote             bool
	ApplicantCountries []string
	Salary             *Salary
	// Identifier is the employer's own identifier for the opening.
	Identifier *JobIdentifier
	// DirectApply says the posting's URL is where one applies, without a detour
	// to another site. It is written only when true.
	DirectApply           bool
	EducationRequirements string
	// ExperienceMonths is the months of experience required; zero leaves it out.
	ExperienceMonths int
}

// EmploymentType is a kind of employment, as Google's job postings name them.
type EmploymentType string

const (
	EmploymentFullTime   EmploymentType = "FULL_TIME"
	EmploymentPartTime   EmploymentType = "PART_TIME"
	EmploymentContractor EmploymentType = "CONTRACTOR"
	EmploymentTemporary  EmploymentType = "TEMPORARY"
	EmploymentIntern     EmploymentType = "INTERN"
	EmploymentVolunteer  EmploymentType = "VOLUNTEER"
	EmploymentPerDiem    EmploymentType = "PER_DIEM"
	EmploymentOther      EmploymentType = "OTHER"
)

// Salary is a job's base pay: schema.org/MonetaryAmount holding a
// QuantitativeValue. Give Value for a fixed amount, or Min and Max for a range.
type Salary struct {
	// Currency is an ISO 4217 code: "TRY".
	Currency string
	Value    float64
	Min      float64
	Max      float64
	Unit     SalaryUnit
}

// SalaryUnit is the period a Salary is paid for.
type SalaryUnit string

const (
	SalaryPerHour  SalaryUnit = "HOUR"
	SalaryPerDay   SalaryUnit = "DAY"
	SalaryPerWeek  SalaryUnit = "WEEK"
	SalaryPerMonth SalaryUnit = "MONTH"
	SalaryPerYear  SalaryUnit = "YEAR"
)

// JobIdentifier is an employer's identifier for an opening:
// schema.org/PropertyValue, Name being the employer and Value the identifier.
type JobIdentifier struct {
	Name  string
	Value string
}

func (JobPosting) Type() string { return "JobPosting" }

func (j JobPosting) nodeKey() string { return j.Key }

func (j JobPosting) MarshalJSON() ([]byte, error) {
	out := listingJobPostingJSON{
		Context:               schemaContext,
		ID:                    j.ID,
		SchemaType:            "JobPosting",
		Title:                 j.Title,
		Description:           j.Description,
		DatePosted:            date(j.DatePosted),
		ValidThrough:          dateTime(j.ValidThrough),
		EducationRequirements: j.EducationRequirements,
	}
	for _, kind := range j.EmploymentType {
		out.EmploymentType = append(out.EmploymentType, string(kind))
	}
	out.HiringOrganization = organizationAgent(j.HiringOrganization)

	switch len(j.Locations) {
	case 0:
	case 1:
		location, err := json.Marshal(listingPlace(j.Locations[0]))
		if err != nil {
			return nil, err
		}
		out.JobLocation = location
	default:
		places := make([]listingPlaceJSON, 0, len(j.Locations))
		for _, address := range j.Locations {
			places = append(places, listingPlace(address))
		}
		locations, err := json.Marshal(places)
		if err != nil {
			return nil, err
		}
		out.JobLocation = locations
	}
	if j.Remote {
		out.JobLocationType = "TELECOMMUTE"
	}
	for _, country := range j.ApplicantCountries {
		out.ApplicantLocationRequirements = append(out.ApplicantLocationRequirements,
			listingCountryJSON{SchemaType: "Country", Name: country})
	}
	if j.Salary != nil {
		out.BaseSalary = &listingMonetaryAmountJSON{
			SchemaType: "MonetaryAmount",
			Currency:   j.Salary.Currency,
		}
	}
	if j.Salary != nil && (j.Salary.Value != 0 || j.Salary.Min != 0 || j.Salary.Max != 0) {
		out.BaseSalary.Value = &listingQuantitativeValueJSON{
			SchemaType: "QuantitativeValue",
			Value:      j.Salary.Value,
			MinValue:   j.Salary.Min,
			MaxValue:   j.Salary.Max,
			UnitText:   string(j.Salary.Unit),
		}
	}
	if j.Identifier != nil {
		out.Identifier = &listingPropertyValueJSON{SchemaType: "PropertyValue", Name: j.Identifier.Name, Value: j.Identifier.Value}
	}
	if j.DirectApply {
		out.DirectApply = true
	}
	if j.ExperienceMonths > 0 {
		out.ExperienceRequirements = &listingExperienceJSON{
			SchemaType:         "OccupationalExperienceRequirements",
			MonthsOfExperience: j.ExperienceMonths,
		}
	}
	return json.Marshal(out)
}

// listingPlace is an address as the Place a job is located at.
func listingPlace(address PostalAddress) listingPlaceJSON {
	return listingPlaceJSON{SchemaType: "Place", Address: address.wire()}
}

type listingJobPostingJSON struct {
	Context                       string                     `json:"@context"`
	ID                            string                     `json:"@id,omitempty"`
	SchemaType                    string                     `json:"@type"`
	Title                         string                     `json:"title"`
	Description                   string                     `json:"description,omitempty"`
	Identifier                    *listingPropertyValueJSON  `json:"identifier,omitempty"`
	DatePosted                    string                     `json:"datePosted,omitempty"`
	ValidThrough                  string                     `json:"validThrough,omitempty"`
	EmploymentType                []string                   `json:"employmentType,omitempty"`
	HiringOrganization            *agentJSON                 `json:"hiringOrganization,omitempty"`
	JobLocation                   json.RawMessage            `json:"jobLocation,omitempty"`
	JobLocationType               string                     `json:"jobLocationType,omitempty"`
	ApplicantLocationRequirements []listingCountryJSON       `json:"applicantLocationRequirements,omitempty"`
	BaseSalary                    *listingMonetaryAmountJSON `json:"baseSalary,omitempty"`
	DirectApply                   bool                       `json:"directApply,omitempty"`
	EducationRequirements         string                     `json:"educationRequirements,omitempty"`
	ExperienceRequirements        *listingExperienceJSON     `json:"experienceRequirements,omitempty"`
}

type listingPlaceJSON struct {
	SchemaType string             `json:"@type"`
	Address    *postalAddressJSON `json:"address,omitempty"`
}

type listingCountryJSON struct {
	SchemaType string `json:"@type"`
	Name       string `json:"name"`
}

type listingMonetaryAmountJSON struct {
	SchemaType string                        `json:"@type"`
	Currency   string                        `json:"currency,omitempty"`
	Value      *listingQuantitativeValueJSON `json:"value,omitempty"`
}

type listingQuantitativeValueJSON struct {
	SchemaType string  `json:"@type"`
	Value      float64 `json:"value,omitempty"`
	MinValue   float64 `json:"minValue,omitempty"`
	MaxValue   float64 `json:"maxValue,omitempty"`
	UnitText   string  `json:"unitText,omitempty"`
}

type listingPropertyValueJSON struct {
	SchemaType string `json:"@type"`
	Name       string `json:"name,omitempty"`
	Value      string `json:"value"`
}

type listingExperienceJSON struct {
	SchemaType         string `json:"@type"`
	MonthsOfExperience int    `json:"monthsOfExperience"`
}

// DiscussionForumPosting is a thread on a forum, with its replies:
// schema.org/DiscussionForumPosting.
//
// Google wants an author, a publication date and the post's text, image or
// video. The counts are left out at zero.
type DiscussionForumPosting struct {
	// ID and Key are as on Article.
	ID            string
	Key           string
	Headline      string
	Text          string
	URL           string
	Author        Agent
	DatePublished time.Time
	DateModified  time.Time
	ImageURLs     []string
	// VideoURL is a video the post consists of or carries.
	VideoURL string
	// Comments are the replies, each of which may hold replies of its own.
	Comments     []Comment
	UpvoteCount  int
	CommentCount int
	ViewCount    int
}

// Comment is a reply in a thread: schema.org/Comment.
type Comment struct {
	Text          string
	Author        Agent
	DatePublished time.Time
	URL           string
	UpvoteCount   int
	// Comments are the replies to this one.
	Comments []Comment
}

func (DiscussionForumPosting) Type() string { return "DiscussionForumPosting" }

func (d DiscussionForumPosting) nodeKey() string { return d.Key }

func (d DiscussionForumPosting) MarshalJSON() ([]byte, error) {
	out := listingForumPostingJSON{
		Context:              schemaContext,
		ID:                   d.ID,
		SchemaType:           "DiscussionForumPosting",
		Headline:             d.Headline,
		Text:                 d.Text,
		URL:                  d.URL,
		Author:               agentPtr(d.Author),
		DatePublished:        dateTime(d.DatePublished),
		DateModified:         dateTime(d.DateModified),
		Image:                d.ImageURLs,
		Comment:              listingComments(d.Comments),
		InteractionStatistic: listingInteractions(d.UpvoteCount, d.CommentCount, d.ViewCount),
	}
	if d.VideoURL != "" {
		out.Video = &listingVideoJSON{SchemaType: "VideoObject", ContentURL: d.VideoURL}
	}
	return json.Marshal(out)
}

// listingComments is the wire form of a list of comments and their replies, nil
// for none.
func listingComments(list []Comment) []listingCommentJSON {
	var out []listingCommentJSON
	for _, c := range list {
		out = append(out, listingCommentJSON{
			SchemaType:           "Comment",
			Text:                 c.Text,
			Author:               agentPtr(c.Author),
			DatePublished:        dateTime(c.DatePublished),
			URL:                  c.URL,
			InteractionStatistic: listingInteractions(c.UpvoteCount, 0, 0),
			Comment:              listingComments(c.Comments),
		})
	}
	return out
}

// listingInteractions is a post's counts as schema.org InteractionCounters, the
// zero ones left out, and nil when all are zero.
func listingInteractions(upvotes, comments, views int) []listingInteractionCounterJSON {
	var out []listingInteractionCounterJSON
	for _, counter := range []struct {
		action string
		count  int
	}{
		{"https://schema.org/LikeAction", upvotes},
		{"https://schema.org/CommentAction", comments},
		{"https://schema.org/ViewAction", views},
	} {
		if counter.count > 0 {
			out = append(out, listingInteractionCounterJSON{
				SchemaType:           "InteractionCounter",
				InteractionType:      counter.action,
				UserInteractionCount: counter.count,
			})
		}
	}
	return out
}

type listingForumPostingJSON struct {
	Context              string                          `json:"@context"`
	ID                   string                          `json:"@id,omitempty"`
	SchemaType           string                          `json:"@type"`
	Headline             string                          `json:"headline,omitempty"`
	Text                 string                          `json:"text,omitempty"`
	URL                  string                          `json:"url,omitempty"`
	Author               *agentJSON                      `json:"author,omitempty"`
	DatePublished        string                          `json:"datePublished,omitempty"`
	DateModified         string                          `json:"dateModified,omitempty"`
	Image                []string                        `json:"image,omitempty"`
	Video                *listingVideoJSON               `json:"video,omitempty"`
	InteractionStatistic []listingInteractionCounterJSON `json:"interactionStatistic,omitempty"`
	Comment              []listingCommentJSON            `json:"comment,omitempty"`
}

type listingCommentJSON struct {
	SchemaType           string                          `json:"@type"`
	Text                 string                          `json:"text,omitempty"`
	Author               *agentJSON                      `json:"author,omitempty"`
	DatePublished        string                          `json:"datePublished,omitempty"`
	URL                  string                          `json:"url,omitempty"`
	InteractionStatistic []listingInteractionCounterJSON `json:"interactionStatistic,omitempty"`
	Comment              []listingCommentJSON            `json:"comment,omitempty"`
}

type listingInteractionCounterJSON struct {
	SchemaType           string `json:"@type"`
	InteractionType      string `json:"interactionType"`
	UserInteractionCount int    `json:"userInteractionCount"`
}

type listingVideoJSON struct {
	SchemaType string `json:"@type"`
	ContentURL string `json:"contentUrl"`
}

// QAPage is a page with one question and the answers to it:
// schema.org/QAPage. A page of questions the site answers itself is an
// FAQPage instead.
type QAPage struct {
	// ID and Key are as on Article.
	ID       string
	Key      string
	Question QAQuestion
}

// QAQuestion is the question a QAPage is about: schema.org/Question.
//
// Google wants its name and at least one answer, accepted or suggested.
// AnswerCount is all the answers the page has, not only those listed here; left
// zero, it is the number listed. It is always written, because Google requires it.
type QAQuestion struct {
	// Name is the question in short, the way a title is.
	Name          string
	Text          string
	AnswerCount   int
	UpvoteCount   int
	DatePublished time.Time
	DateModified  time.Time
	Author        Agent
	// Accepted are the answers the asker or the site accepted; Suggested are
	// the others.
	Accepted  []QAAnswer
	Suggested []QAAnswer
}

// QAAnswer is an answer to a QAQuestion: schema.org/Answer.
type QAAnswer struct {
	Text string
	// URL links to the answer itself, typically the page with a fragment.
	URL           string
	UpvoteCount   int
	DatePublished time.Time
	DateModified  time.Time
	Author        Agent
}

func (QAPage) Type() string { return "QAPage" }

func (q QAPage) nodeKey() string { return q.Key }

func (q QAPage) MarshalJSON() ([]byte, error) {
	question := q.Question
	answerCount := question.AnswerCount
	if answerCount == 0 {
		answerCount = len(question.Accepted) + len(question.Suggested)
	}
	return json.Marshal(listingQaPageJSON{
		Context:    schemaContext,
		ID:         q.ID,
		SchemaType: "QAPage",
		MainEntity: listingQaQuestionJSON{
			SchemaType:      "Question",
			Name:            question.Name,
			Text:            question.Text,
			AnswerCount:     answerCount,
			UpvoteCount:     question.UpvoteCount,
			DatePublished:   dateTime(question.DatePublished),
			DateModified:    dateTime(question.DateModified),
			Author:          agentPtr(question.Author),
			AcceptedAnswer:  listingAnswers(question.Accepted),
			SuggestedAnswer: listingAnswers(question.Suggested),
		},
	})
}

// listingAnswers is the wire form of a list of answers, nil for none.
func listingAnswers(list []QAAnswer) []listingQaAnswerJSON {
	var out []listingQaAnswerJSON
	for _, a := range list {
		out = append(out, listingQaAnswerJSON{
			SchemaType:    "Answer",
			Text:          a.Text,
			URL:           a.URL,
			UpvoteCount:   a.UpvoteCount,
			DatePublished: dateTime(a.DatePublished),
			DateModified:  dateTime(a.DateModified),
			Author:        agentPtr(a.Author),
		})
	}
	return out
}

type listingQaPageJSON struct {
	Context    string                `json:"@context"`
	ID         string                `json:"@id,omitempty"`
	SchemaType string                `json:"@type"`
	MainEntity listingQaQuestionJSON `json:"mainEntity"`
}

type listingQaQuestionJSON struct {
	SchemaType      string                `json:"@type"`
	Name            string                `json:"name"`
	Text            string                `json:"text,omitempty"`
	AnswerCount     int                   `json:"answerCount"`
	UpvoteCount     int                   `json:"upvoteCount,omitempty"`
	DatePublished   string                `json:"datePublished,omitempty"`
	DateModified    string                `json:"dateModified,omitempty"`
	Author          *agentJSON            `json:"author,omitempty"`
	AcceptedAnswer  []listingQaAnswerJSON `json:"acceptedAnswer,omitempty"`
	SuggestedAnswer []listingQaAnswerJSON `json:"suggestedAnswer,omitempty"`
}

type listingQaAnswerJSON struct {
	SchemaType    string     `json:"@type"`
	Text          string     `json:"text"`
	URL           string     `json:"url,omitempty"`
	UpvoteCount   int        `json:"upvoteCount,omitempty"`
	DatePublished string     `json:"datePublished,omitempty"`
	DateModified  string     `json:"dateModified,omitempty"`
	Author        *agentJSON `json:"author,omitempty"`
}

// VacationRental is a place to stay rented out as a whole:
// schema.org/VacationRental.
//
// Google wants an Identifier, a name, images — at least eight — an address or
// coordinates, and the Accommodation with its occupancy.
type VacationRental struct {
	// ID and Key are as on Article.
	ID  string
	Key string
	// Identifier is the listing's own identifier, stable across updates.
	Identifier  string
	Name        string
	Description string
	ImageURLs   []string
	Address     *PostalAddress
	Geo         GeoCoordinates
	// AdditionalType is what the property is: "House", "Apartment", "Villa",
	// "Cottage".
	AdditionalType string
	// Brand is the rental brand or host company, when there is one.
	Brand string
	// CheckinTime and CheckoutTime are times of day with their offset, as Google
	// writes them: "14:30:00+03:00".
	CheckinTime  string
	CheckoutTime string
	// KnowsLanguage are the languages the host speaks, as IETF tags: "tr", "en".
	KnowsLanguage   []string
	Accommodation   Accommodation
	AggregateRating *AggregateRating
	Reviews         []Review
}

// Accommodation is what a VacationRental rents: schema.org/Accommodation.
type Accommodation struct {
	// AdditionalType is how much of the property is rented: "EntirePlace",
	// "PrivateRoom", "SharedRoom".
	AdditionalType string
	// Occupancy is the most guests it takes.
	Occupancy int
	Rooms     int
	Bedrooms  int
	// Bathrooms counts a half bathroom as 0.5.
	Bathrooms   float64
	FloorSizeM2 float64
	Beds        []Bed
	Amenities   []Amenity
}

// Amenity is something an Accommodation offers, by Google's amenity name:
// "wifi", "ac", "kitchen", "pool". An empty Value says it is there; the
// amenities that take a kind — "internetType", "parkingType", "poolType",
// "licenseNum" — say it in Value: "Free", "Indoor".
type Amenity struct {
	Name  string
	Value string
}

// Bed is a kind of bed and how many of them there are: schema.org/BedDetails.
type Bed struct {
	// Type is the bed's kind: "Double", "Queen", "Single", "SofaBed".
	Type  string
	Count int
}

func (VacationRental) Type() string { return "VacationRental" }

func (v VacationRental) nodeKey() string { return v.Key }

func (v VacationRental) MarshalJSON() ([]byte, error) {
	out := listingVacationRentalJSON{
		Context:         schemaContext,
		ID:              v.ID,
		SchemaType:      "VacationRental",
		Identifier:      v.Identifier,
		Name:            v.Name,
		Description:     v.Description,
		Image:           v.ImageURLs,
		Address:         v.Address.wire(),
		CheckinTime:     v.CheckinTime,
		CheckoutTime:    v.CheckoutTime,
		KnowsLanguage:   v.KnowsLanguage,
		AdditionalType:  v.AdditionalType,
		AggregateRating: aggregate(v.AggregateRating),
		Review:          reviews(v.Reviews),
	}
	if geo := v.Geo.wire(); geo != nil {
		out.Latitude = &geo.Latitude
		out.Longitude = &geo.Longitude
	}
	if !v.Accommodation.empty() {
		place := v.Accommodation.wire()
		out.ContainsPlace = &place
	}
	if v.Brand != "" {
		out.Brand = &listingBrandJSON{SchemaType: "Brand", Name: v.Brand}
	}
	return json.Marshal(out)
}

// empty reports whether nothing about the accommodation was given, so the
// property is left out rather than written as a bare type.
func (a Accommodation) empty() bool {
	return a.AdditionalType == "" && a.Occupancy == 0 && a.Rooms == 0 && a.Bedrooms == 0 && a.Bathrooms == 0 &&
		a.FloorSizeM2 == 0 && len(a.Beds) == 0 && len(a.Amenities) == 0
}

func (a Accommodation) wire() listingAccommodationJSON {
	out := listingAccommodationJSON{
		SchemaType:             "Accommodation",
		AdditionalType:         a.AdditionalType,
		NumberOfRooms:          a.Rooms,
		NumberOfBedrooms:       a.Bedrooms,
		NumberOfBathroomsTotal: a.Bathrooms,
	}
	if a.Occupancy > 0 {
		out.Occupancy = &listingQuantitativeValueJSON{SchemaType: "QuantitativeValue", Value: float64(a.Occupancy)}
	}
	if a.FloorSizeM2 > 0 {
		// MTK is UN/CEFACT's code for a square metre, which is how schema.org
		// names a unit.
		out.FloorSize = &listingFloorSizeJSON{SchemaType: "QuantitativeValue", Value: a.FloorSizeM2, UnitCode: "MTK"}
	}
	for _, bed := range a.Beds {
		out.Bed = append(out.Bed, listingBedJSON{SchemaType: "BedDetails", NumberOfBeds: bed.Count, TypeOfBed: bed.Type})
	}
	for _, amenity := range a.Amenities {
		value := json.RawMessage("true")
		if amenity.Value != "" {
			// A string marshals without fail; the error is json.Marshal's
			// signature, not a case.
			value, _ = json.Marshal(amenity.Value)
		}
		out.AmenityFeature = append(out.AmenityFeature, listingAmenityJSON{
			SchemaType: "LocationFeatureSpecification",
			Name:       amenity.Name,
			Value:      value,
		})
	}
	return out
}

type listingVacationRentalJSON struct {
	Context         string                    `json:"@context"`
	ID              string                    `json:"@id,omitempty"`
	SchemaType      string                    `json:"@type"`
	AdditionalType  string                    `json:"additionalType,omitempty"`
	Identifier      string                    `json:"identifier,omitempty"`
	Name            string                    `json:"name"`
	Description     string                    `json:"description,omitempty"`
	Image           []string                  `json:"image,omitempty"`
	Address         *postalAddressJSON        `json:"address,omitempty"`
	Latitude        *float64                  `json:"latitude,omitempty"`
	Longitude       *float64                  `json:"longitude,omitempty"`
	Brand           *listingBrandJSON         `json:"brand,omitempty"`
	CheckinTime     string                    `json:"checkinTime,omitempty"`
	CheckoutTime    string                    `json:"checkoutTime,omitempty"`
	KnowsLanguage   []string                  `json:"knowsLanguage,omitempty"`
	ContainsPlace   *listingAccommodationJSON `json:"containsPlace,omitempty"`
	AggregateRating *aggregateRatingJSON      `json:"aggregateRating,omitempty"`
	Review          []reviewJSON              `json:"review,omitempty"`
}

type listingBrandJSON struct {
	SchemaType string `json:"@type"`
	Name       string `json:"name"`
}

type listingAccommodationJSON struct {
	SchemaType             string                        `json:"@type"`
	AdditionalType         string                        `json:"additionalType,omitempty"`
	Occupancy              *listingQuantitativeValueJSON `json:"occupancy,omitempty"`
	NumberOfRooms          int                           `json:"numberOfRooms,omitempty"`
	NumberOfBedrooms       int                           `json:"numberOfBedrooms,omitempty"`
	NumberOfBathroomsTotal float64                       `json:"numberOfBathroomsTotal,omitempty"`
	FloorSize              *listingFloorSizeJSON         `json:"floorSize,omitempty"`
	Bed                    []listingBedJSON              `json:"bed,omitempty"`
	AmenityFeature         []listingAmenityJSON          `json:"amenityFeature,omitempty"`
}

type listingFloorSizeJSON struct {
	SchemaType string  `json:"@type"`
	Value      float64 `json:"value"`
	UnitCode   string  `json:"unitCode"`
}

type listingBedJSON struct {
	SchemaType   string `json:"@type"`
	NumberOfBeds int    `json:"numberOfBeds,omitempty"`
	TypeOfBed    string `json:"typeOfBed,omitempty"`
}

type listingAmenityJSON struct {
	SchemaType string          `json:"@type"`
	Name       string          `json:"name"`
	Value      json.RawMessage `json:"value"`
}

var (
	_ keyed = LocalBusiness{}
	_ keyed = JobPosting{}
	_ keyed = DiscussionForumPosting{}
	_ keyed = QAPage{}
	_ keyed = VacationRental{}
)
