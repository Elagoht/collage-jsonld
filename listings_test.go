package jsonld_test

import (
	"strings"
	"testing"
	"time"

	"github.com/Elagoht/collage/pkg/collage"

	jsonld "github.com/Elagoht/collage-jsonld"
)

var listingIstanbul = time.FixedZone("Europe/Istanbul", 3*60*60)

func TestListings_Marshal(t *testing.T) {
	cases := []struct {
		name string
		node jsonld.Node
		want string
	}{
		{
			name: "local business, everything",
			node: jsonld.LocalBusiness{
				BusinessType:  "Restaurant",
				ID:            "https://kahve.example/#place",
				Name:          "Kahve Durağı",
				Description:   "Üçüncü dalga kahve.",
				URL:           "https://kahve.example",
				ImageURLs:     []string{"https://kahve.example/1x1.jpg", "https://kahve.example/16x9.jpg"},
				LogoURL:       "https://kahve.example/logo.png",
				Telephone:     "+90 216 000 00 00",
				Email:         "merhaba@kahve.example",
				Address:       &jsonld.PostalAddress{StreetAddress: "Moda Cd. 1", Locality: "Kadıköy", Region: "İstanbul", PostalCode: "34710", Country: "TR"},
				Geo:           jsonld.GeoCoordinates{Latitude: 40.987, Longitude: 29.026},
				PriceRange:    "₺₺",
				SameAs:        []string{"https://instagram.com/kahve"},
				Menu:          "https://kahve.example/menu",
				ServesCuisine: []string{"Kahve", "Tatlı"},
				OpeningHours: []jsonld.OpeningHours{
					{Days: []time.Weekday{time.Monday, time.Tuesday}, Opens: "08:00", Closes: "20:00"},
					{Days: []time.Weekday{time.Saturday}, Opens: "00:00", Closes: "00:00",
						ValidFrom: time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC), ValidThrough: time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)},
				},
				AcceptsReservations: true,
				AggregateRating:     &jsonld.AggregateRating{Value: 4.6, RatingCount: 120},
				Reviews:             []jsonld.Review{{Author: jsonld.Agent{Name: "Ada"}, Rating: &jsonld.Rating{Value: 5}, Body: "Harika."}},
			},
			want: "{\"@context\":\"https://schema.org\",\"@id\":\"https://kahve.example/#place\",\"@type\":\"Restaurant\",\"name\":\"Kahve Durağı\",\"description\":\"Üçüncü dalga kahve.\",\"url\":\"https://kahve.example\",\"image\":[\"https://kahve.example/1x1.jpg\",\"https://kahve.example/16x9.jpg\"],\"logo\":{\"@type\":\"ImageObject\",\"url\":\"https://kahve.example/logo.png\"},\"telephone\":\"+90 216 000 00 00\",\"email\":\"merhaba@kahve.example\",\"address\":{\"@type\":\"PostalAddress\",\"streetAddress\":\"Moda Cd. 1\",\"addressLocality\":\"Kadıköy\",\"addressRegion\":\"İstanbul\",\"postalCode\":\"34710\",\"addressCountry\":\"TR\"},\"geo\":{\"@type\":\"GeoCoordinates\",\"latitude\":40.987,\"longitude\":29.026},\"priceRange\":\"₺₺\",\"openingHoursSpecification\":[{\"@type\":\"OpeningHoursSpecification\",\"dayOfWeek\":[\"Monday\",\"Tuesday\"],\"opens\":\"08:00\",\"closes\":\"20:00\"},{\"@type\":\"OpeningHoursSpecification\",\"dayOfWeek\":[\"Saturday\"],\"opens\":\"00:00\",\"closes\":\"00:00\",\"validFrom\":\"2026-12-31\",\"validThrough\":\"2027-01-01\"}],\"sameAs\":[\"https://instagram.com/kahve\"],\"menu\":\"https://kahve.example/menu\",\"servesCuisine\":[\"Kahve\",\"Tatlı\"],\"acceptsReservations\":true,\"aggregateRating\":{\"@type\":\"AggregateRating\",\"ratingValue\":4.6,\"ratingCount\":120},\"review\":[{\"@type\":\"Review\",\"author\":{\"@type\":\"Person\",\"name\":\"Ada\"},\"reviewRating\":{\"@type\":\"Rating\",\"ratingValue\":5},\"reviewBody\":\"Harika.\"}]}",
		},
		{
			name: "job posting, everything",
			node: jsonld.JobPosting{
				ID:                       "https://ose.example/jobs/1#job",
				Title:                    "Oyun Programcısı",
				Description:              "<p>Unity ile <b>oyun</b> geliştir.</p>",
				DatePosted:               time.Date(2026, 10, 1, 0, 0, 0, 0, listingIstanbul),
				ValidThrough:             time.Date(2026, 11, 1, 0, 0, 0, 0, listingIstanbul),
				EmploymentType:           []jsonld.EmploymentType{jsonld.EmploymentFullTime, jsonld.EmploymentContractor},
				HiringOrganization:       jsonld.Agent{Name: "OSE", URL: "https://ose.example"},
				HiringOrganizationLogo:   "https://ose.example/logo.png",
				HiringOrganizationSameAs: "https://linkedin.com/company/ose",
				Locations: []jsonld.PostalAddress{
					{Locality: "Kadıköy", Country: "TR"},
					{Locality: "Çankaya", Country: "TR"},
				},
				Remote:                true,
				ApplicantCountries:    []string{"TR"},
				Salary:                &jsonld.Salary{Currency: "TRY", Min: 80000, Max: 120000, Unit: jsonld.SalaryPerMonth},
				Identifier:            &jsonld.JobIdentifier{Name: "OSE", Value: "J-1"},
				DirectApply:           true,
				EducationRequirements: "bachelor degree",
				ExperienceMonths:      24,
			},
			want: "{\"@context\":\"https://schema.org\",\"@id\":\"https://ose.example/jobs/1#job\",\"@type\":\"JobPosting\",\"title\":\"Oyun Programcısı\",\"description\":\"\\u003cp\\u003eUnity ile \\u003cb\\u003eoyun\\u003c/b\\u003e geliştir.\\u003c/p\\u003e\",\"identifier\":{\"@type\":\"PropertyValue\",\"name\":\"OSE\",\"value\":\"J-1\"},\"datePosted\":\"2026-10-01\",\"validThrough\":\"2026-11-01T00:00:00+03:00\",\"employmentType\":[\"FULL_TIME\",\"CONTRACTOR\"],\"hiringOrganization\":{\"@type\":\"Organization\",\"name\":\"OSE\",\"sameAs\":\"https://linkedin.com/company/ose\",\"url\":\"https://ose.example\",\"logo\":\"https://ose.example/logo.png\"},\"jobLocation\":[{\"@type\":\"Place\",\"address\":{\"@type\":\"PostalAddress\",\"addressLocality\":\"Kadıköy\",\"addressCountry\":\"TR\"}},{\"@type\":\"Place\",\"address\":{\"@type\":\"PostalAddress\",\"addressLocality\":\"Çankaya\",\"addressCountry\":\"TR\"}}],\"jobLocationType\":\"TELECOMMUTE\",\"applicantLocationRequirements\":[{\"@type\":\"Country\",\"name\":\"TR\"}],\"baseSalary\":{\"@type\":\"MonetaryAmount\",\"currency\":\"TRY\",\"value\":{\"@type\":\"QuantitativeValue\",\"minValue\":80000,\"maxValue\":120000,\"unitText\":\"MONTH\"}},\"directApply\":true,\"educationRequirements\":\"bachelor degree\",\"experienceRequirements\":{\"@type\":\"OccupationalExperienceRequirements\",\"monthsOfExperience\":24}}",
		},
		{
			name: "forum posting, everything",
			node: jsonld.DiscussionForumPosting{
				ID:            "https://forum.example/t/1#post",
				Headline:      "En iyi indie oyun?",
				Text:          "Önerileriniz?",
				URL:           "https://forum.example/t/1",
				Author:        jsonld.Agent{Name: "Ada", URL: "https://forum.example/u/ada"},
				DatePublished: time.Date(2026, 10, 1, 12, 0, 0, 0, listingIstanbul),
				DateModified:  time.Date(2026, 10, 2, 12, 0, 0, 0, listingIstanbul),
				ImageURLs:     []string{"https://forum.example/a.png"},
				VideoURL:      "https://forum.example/v.mp4",
				Comments: []jsonld.Comment{{
					Text: "Hades.", Author: jsonld.Agent{Name: "Kemal"}, URL: "https://forum.example/t/1#c1", UpvoteCount: 3,
					DatePublished: time.Date(2026, 10, 1, 13, 0, 0, 0, listingIstanbul),
					Comments:      []jsonld.Comment{{Text: "Katılıyorum.", Author: jsonld.Agent{Name: "Ada"}}},
				}},
				UpvoteCount:  10,
				CommentCount: 2,
				ViewCount:    300,
			},
			want: "{\"@context\":\"https://schema.org\",\"@id\":\"https://forum.example/t/1#post\",\"@type\":\"DiscussionForumPosting\",\"headline\":\"En iyi indie oyun?\",\"text\":\"Önerileriniz?\",\"url\":\"https://forum.example/t/1\",\"author\":{\"@type\":\"Person\",\"name\":\"Ada\",\"url\":\"https://forum.example/u/ada\"},\"datePublished\":\"2026-10-01T12:00:00+03:00\",\"dateModified\":\"2026-10-02T12:00:00+03:00\",\"image\":[\"https://forum.example/a.png\"],\"video\":{\"@type\":\"VideoObject\",\"contentUrl\":\"https://forum.example/v.mp4\"},\"interactionStatistic\":[{\"@type\":\"InteractionCounter\",\"interactionType\":\"https://schema.org/LikeAction\",\"userInteractionCount\":10},{\"@type\":\"InteractionCounter\",\"interactionType\":\"https://schema.org/CommentAction\",\"userInteractionCount\":2},{\"@type\":\"InteractionCounter\",\"interactionType\":\"https://schema.org/ViewAction\",\"userInteractionCount\":300}],\"comment\":[{\"@type\":\"Comment\",\"text\":\"Hades.\",\"author\":{\"@type\":\"Person\",\"name\":\"Kemal\"},\"datePublished\":\"2026-10-01T13:00:00+03:00\",\"url\":\"https://forum.example/t/1#c1\",\"interactionStatistic\":[{\"@type\":\"InteractionCounter\",\"interactionType\":\"https://schema.org/LikeAction\",\"userInteractionCount\":3}],\"comment\":[{\"@type\":\"Comment\",\"text\":\"Katılıyorum.\",\"author\":{\"@type\":\"Person\",\"name\":\"Ada\"}}]}]}",
		},
		{
			name: "qa page, everything",
			node: jsonld.QAPage{
				ID: "https://qa.example/q/1#page",
				Question: jsonld.QAQuestion{
					Name: "Unity mi Godot mu?", Text: "Yeni başlıyorum.", AnswerCount: 2, UpvoteCount: 4,
					DateCreated: time.Date(2026, 10, 1, 9, 0, 0, 0, listingIstanbul), Author: jsonld.Agent{Name: "Ada"},
					Accepted:  []jsonld.QAAnswer{{Text: "Godot.", URL: "https://qa.example/q/1#a1", UpvoteCount: 9, Author: jsonld.Agent{Name: "Kemal"}}},
					Suggested: []jsonld.QAAnswer{{Text: "Unity.", DateCreated: time.Date(2026, 10, 1, 10, 0, 0, 0, listingIstanbul)}},
				},
			},
			want: "{\"@context\":\"https://schema.org\",\"@id\":\"https://qa.example/q/1#page\",\"@type\":\"QAPage\",\"mainEntity\":{\"@type\":\"Question\",\"name\":\"Unity mi Godot mu?\",\"text\":\"Yeni başlıyorum.\",\"answerCount\":2,\"upvoteCount\":4,\"dateCreated\":\"2026-10-01T09:00:00+03:00\",\"author\":{\"@type\":\"Person\",\"name\":\"Ada\"},\"acceptedAnswer\":[{\"@type\":\"Answer\",\"text\":\"Godot.\",\"url\":\"https://qa.example/q/1#a1\",\"upvoteCount\":9,\"author\":{\"@type\":\"Person\",\"name\":\"Kemal\"}}],\"suggestedAnswer\":[{\"@type\":\"Answer\",\"text\":\"Unity.\",\"dateCreated\":\"2026-10-01T10:00:00+03:00\"}]}}",
		},
		{
			name: "vacation rental, everything",
			node: jsonld.VacationRental{
				ID:            "https://ev.example/1#rental",
				Identifier:    "ev-1",
				Name:          "Deniz Evi",
				Description:   "Denize sıfır.",
				ImageURLs:     []string{"https://ev.example/1.jpg", "https://ev.example/2.jpg"},
				Address:       &jsonld.PostalAddress{Locality: "Bodrum", Region: "Muğla", Country: "TR"},
				Geo:           jsonld.GeoCoordinates{Latitude: 37.03, Longitude: 27.43},
				Brand:         "Ev Kiralama",
				CheckinTime:   "15:00",
				CheckoutTime:  "11:00",
				KnowsLanguage: []string{"tr", "en"},
				Accommodation: jsonld.Accommodation{
					Kind: "House", Occupancy: 6, Bedrooms: 3, Bathrooms: 2, FloorSizeM2: 120,
					Beds:      []jsonld.Bed{{Type: "Double", Count: 2}, {Type: "Single", Count: 2}},
					Amenities: []string{"wifi", "pool"},
				},
				AggregateRating: &jsonld.AggregateRating{Value: 4.9, ReviewCount: 30},
			},
			want: "{\"@context\":\"https://schema.org\",\"@id\":\"https://ev.example/1#rental\",\"@type\":\"VacationRental\",\"identifier\":\"ev-1\",\"name\":\"Deniz Evi\",\"description\":\"Denize sıfır.\",\"image\":[\"https://ev.example/1.jpg\",\"https://ev.example/2.jpg\"],\"address\":{\"@type\":\"PostalAddress\",\"addressLocality\":\"Bodrum\",\"addressRegion\":\"Muğla\",\"addressCountry\":\"TR\"},\"latitude\":37.03,\"longitude\":27.43,\"brand\":{\"@type\":\"Brand\",\"name\":\"Ev Kiralama\"},\"checkinTime\":\"15:00\",\"checkoutTime\":\"11:00\",\"knowsLanguage\":[\"tr\",\"en\"],\"containsPlace\":{\"@type\":\"House\",\"occupancy\":{\"@type\":\"QuantitativeValue\",\"value\":6},\"numberOfBedrooms\":3,\"numberOfBathroomsTotal\":2,\"floorSize\":{\"@type\":\"QuantitativeValue\",\"value\":120,\"unitCode\":\"MTK\"},\"bed\":[{\"@type\":\"BedDetails\",\"numberOfBeds\":2,\"typeOfBed\":\"Double\"},{\"@type\":\"BedDetails\",\"numberOfBeds\":2,\"typeOfBed\":\"Single\"}],\"amenityFeature\":[{\"@type\":\"LocationFeatureSpecification\",\"name\":\"wifi\",\"value\":true},{\"@type\":\"LocationFeatureSpecification\",\"name\":\"pool\",\"value\":true}]},\"aggregateRating\":{\"@type\":\"AggregateRating\",\"ratingValue\":4.9,\"reviewCount\":30}}",
		},
		{
			name: "local business, a name alone",
			node: jsonld.LocalBusiness{Name: "Bakkal"},
			want: "{\"@context\":\"https://schema.org\",\"@type\":\"LocalBusiness\",\"name\":\"Bakkal\"}",
		},
		{
			name: "job posting, a title alone",
			node: jsonld.JobPosting{Title: "Stajyer"},
			want: "{\"@context\":\"https://schema.org\",\"@type\":\"JobPosting\",\"title\":\"Stajyer\"}",
		},
		{
			name: "forum posting, text alone",
			node: jsonld.DiscussionForumPosting{Text: "Merhaba"},
			want: "{\"@context\":\"https://schema.org\",\"@type\":\"DiscussionForumPosting\",\"text\":\"Merhaba\"}",
		},
		{
			name: "qa page, a question alone",
			node: jsonld.QAPage{Question: jsonld.QAQuestion{Name: "Neden?"}},
			want: "{\"@context\":\"https://schema.org\",\"@type\":\"QAPage\",\"mainEntity\":{\"@type\":\"Question\",\"name\":\"Neden?\"}}",
		},
		{
			name: "vacation rental, a name alone",
			node: jsonld.VacationRental{Name: "Oda"},
			want: "{\"@context\":\"https://schema.org\",\"@type\":\"VacationRental\",\"name\":\"Oda\"}",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := marshal(t, tc.node); got != tc.want {
				t.Errorf("got  %s\nwant %s", got, tc.want)
			}
		})
	}
}

func TestJobPosting_OneLocationIsAnObject(t *testing.T) {
	// schema.org takes either; one place written as a one-element list reads
	// oddly, so one is an object and more are a list.
	got := marshal(t, jsonld.JobPosting{Title: "x", Locations: []jsonld.PostalAddress{{Locality: "Kadıköy"}}})
	want := `"jobLocation":{"@type":"Place","address":{"@type":"PostalAddress","addressLocality":"Kadıköy"}}`
	if !strings.Contains(got, want) {
		t.Errorf("got %s\nwant it to hold %s", got, want)
	}
}

func TestLocalBusiness_TypeIsItsBusinessType(t *testing.T) {
	if got := (jsonld.LocalBusiness{}).Type(); got != "LocalBusiness" {
		t.Errorf("empty BusinessType: Type() = %q, want LocalBusiness", got)
	}
	if got := (jsonld.LocalBusiness{BusinessType: "Dentist"}).Type(); got != "Dentist" {
		t.Errorf("Type() = %q, want Dentist", got)
	}
}

func TestPlugin_ListingsTakeKeys(t *testing.T) {
	// A loop of each keeps them all, as a loop of people does.
	site := newSite(t, jsonld.New(jsonld.Options{}), func(rc *collage.RenderContext) {
		for _, key := range []string{"a", "b"} {
			jsonld.Emit(rc,
				jsonld.LocalBusiness{Key: key, Name: key},
				jsonld.JobPosting{Key: key, Title: key},
				jsonld.DiscussionForumPosting{Key: key, Text: key},
				jsonld.QAPage{Key: key, Question: jsonld.QAQuestion{Name: key}},
				jsonld.VacationRental{Key: key, Name: key},
			)
		}
	})

	count := map[string]int{}
	for _, node := range blocks(t, get(t, site, "/article")) {
		schemaType, _ := node["@type"].(string)
		count[schemaType]++
	}
	for _, schemaType := range []string{"LocalBusiness", "JobPosting", "DiscussionForumPosting", "QAPage", "VacationRental"} {
		if count[schemaType] != 2 {
			t.Errorf("%s blocks = %d, want 2 (all: %v)", schemaType, count[schemaType], count)
		}
	}
}

func TestPlugin_ClosingScriptTagInAListingCannotEscape(t *testing.T) {
	// A job description may hold HTML by design; it is escaped for the script
	// element like everything else.
	const payload = `</script><img src=x onerror=alert(1)>`
	site := newSite(t, jsonld.New(jsonld.Options{}), func(rc *collage.RenderContext) {
		jsonld.Emit(rc,
			jsonld.LocalBusiness{Name: payload},
			jsonld.JobPosting{Title: "x", Description: payload},
			jsonld.DiscussionForumPosting{Text: payload, Comments: []jsonld.Comment{{Text: payload}}},
			jsonld.QAPage{Question: jsonld.QAQuestion{Name: payload, Accepted: []jsonld.QAAnswer{{Text: payload}}}},
			jsonld.VacationRental{Name: payload, Accommodation: jsonld.Accommodation{Amenities: []string{payload}}},
		)
	})

	body := get(t, site, "/article")
	if strings.Contains(body, "<img src=x") {
		t.Fatalf("the payload escaped the script block:\n%s", body)
	}
	if nodes := blocks(t, body); len(nodes) != 5 {
		t.Fatalf("got %d nodes, want 5", len(nodes))
	}
}
