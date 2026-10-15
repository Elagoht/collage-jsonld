package jsonld_test

import (
	"strings"
	"testing"
	"time"

	"github.com/Elagoht/collage/pkg/collage"

	jsonld "github.com/Elagoht/collage-jsonld"
)

func TestOrganization_Marshal(t *testing.T) {
	cases := []struct {
		name string
		org  jsonld.Organization
		want string
	}{
		{
			name: "everything",
			org: jsonld.Organization{
				ID:          "https://ose.example/#org",
				Key:         "ose",
				Name:        "OSE",
				Description: "Oyun sektörü etkinlikleri.",
				URL:         "https://ose.example",
				LogoURL:     "https://ose.example/logo.png",
				Email:       "info@ose.example",
				Telephone:   "+90 212 000 00 00",
				SameAs:      []string{"https://instagram.com/ose", "https://linkedin.com/company/ose"},
				Address: &jsonld.PostalAddress{
					StreetAddress: "Bir Sokak 1",
					Locality:      "Kadıköy",
					Region:        "İstanbul",
					PostalCode:    "34710",
					Country:       "TR",
				},
			},
			want: `{"@context":"https://schema.org","@id":"https://ose.example/#org","@type":"Organization",` +
				`"name":"OSE","description":"Oyun sektörü etkinlikleri.","url":"https://ose.example",` +
				`"logo":{"@type":"ImageObject","url":"https://ose.example/logo.png"},` +
				`"email":"info@ose.example","telephone":"+90 212 000 00 00",` +
				`"sameAs":["https://instagram.com/ose","https://linkedin.com/company/ose"],` +
				`"address":{"@type":"PostalAddress","streetAddress":"Bir Sokak 1","addressLocality":"Kadıköy",` +
				`"addressRegion":"İstanbul","postalCode":"34710","addressCountry":"TR"}}`,
		},
		{
			name: "a name alone",
			org:  jsonld.Organization{Name: "OSE"},
			want: `{"@context":"https://schema.org","@type":"Organization","name":"OSE"}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := marshal(t, tc.org); got != tc.want {
				t.Errorf("got  %s\nwant %s", got, tc.want)
			}
		})
	}
}

func TestEvent_Marshal(t *testing.T) {
	istanbul := time.FixedZone("Europe/Istanbul", 3*60*60)
	start := time.Date(2026, 11, 7, 19, 0, 0, 0, istanbul)
	end := time.Date(2026, 11, 7, 22, 0, 0, 0, istanbul)
	place := &jsonld.Place{Name: "Kültür Merkezi", Address: &jsonld.PostalAddress{Locality: "Kadıköy", Country: "TR"}}
	const placeJSON = `{"@type":"Place","name":"Kültür Merkezi","address":{"@type":"PostalAddress","addressLocality":"Kadıköy","addressCountry":"TR"}}`
	const onlineJSON = `{"@type":"VirtualLocation","url":"https://stream.example/live"}`

	cases := []struct {
		name  string
		event jsonld.Event
		want  string
	}{
		{
			name: "in person, with everything",
			event: jsonld.Event{
				ID:             "https://ose.example/e/jam#event",
				Name:           "Game Jam",
				Description:    "48 saat.",
				URL:            "https://ose.example/e/jam",
				ImageURLs:      []string{"https://ose.example/jam.png"},
				StartDate:      start,
				EndDate:        end,
				Status:         jsonld.EventScheduled,
				AttendanceMode: jsonld.AttendanceOffline,
				Place:          place,
				Organizer:      jsonld.Agent{Name: "OSE", URL: "https://ose.example"},
				Offers: []jsonld.Offer{{
					URL:          "https://ose.example/e/jam/kayit",
					Price:        "0",
					Currency:     "TRY",
					Availability: jsonld.InStock,
					ValidFrom:    time.Date(2026, 10, 1, 0, 0, 0, 0, istanbul),
				}},
			},
			want: `{"@context":"https://schema.org","@id":"https://ose.example/e/jam#event","@type":"Event",` +
				`"name":"Game Jam","description":"48 saat.","url":"https://ose.example/e/jam",` +
				`"image":["https://ose.example/jam.png"],` +
				`"startDate":"2026-11-07T19:00:00+03:00","endDate":"2026-11-07T22:00:00+03:00",` +
				`"eventStatus":"https://schema.org/EventScheduled",` +
				`"eventAttendanceMode":"https://schema.org/OfflineEventAttendanceMode",` +
				`"location":` + placeJSON + `,` +
				`"organizer":{"@type":"Organization","name":"OSE","url":"https://ose.example"},` +
				`"offers":[{"@type":"Offer","url":"https://ose.example/e/jam/kayit","price":"0","priceCurrency":"TRY",` +
				`"availability":"https://schema.org/InStock","validFrom":"2026-10-01T00:00:00+03:00"}]}`,
		},
		{
			name:  "online",
			event: jsonld.Event{Name: "Meetup", StartDate: start, AttendanceMode: jsonld.AttendanceOnline, OnlineURL: "https://stream.example/live"},
			want: `{"@context":"https://schema.org","@type":"Event","name":"Meetup","startDate":"2026-11-07T19:00:00+03:00",` +
				`"eventAttendanceMode":"https://schema.org/OnlineEventAttendanceMode","location":` + onlineJSON + `}`,
		},
		{
			name: "mixed is both locations",
			event: jsonld.Event{Name: "Konferans", StartDate: start, AttendanceMode: jsonld.AttendanceMixed,
				Place: place, OnlineURL: "https://stream.example/live"},
			want: `{"@context":"https://schema.org","@type":"Event","name":"Konferans","startDate":"2026-11-07T19:00:00+03:00",` +
				`"eventAttendanceMode":"https://schema.org/MixedEventAttendanceMode",` +
				`"location":[` + placeJSON + `,` + onlineJSON + `]}`,
		},
		{
			name:  "a name alone",
			event: jsonld.Event{Name: "Meetup"},
			want:  `{"@context":"https://schema.org","@type":"Event","name":"Meetup"}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := marshal(t, tc.event); got != tc.want {
				t.Errorf("got  %s\nwant %s", got, tc.want)
			}
		})
	}
}

func TestFAQPage_Marshal(t *testing.T) {
	faq := jsonld.FAQPage{Questions: []jsonld.Question{
		{Question: "Katılım ücretli mi?", Answer: "Hayır."},
		{Question: "Online izlenebilir mi?", Answer: "Evet."},
	}}
	want := `{"@context":"https://schema.org","@type":"FAQPage","mainEntity":[` +
		`{"@type":"Question","name":"Katılım ücretli mi?","acceptedAnswer":{"@type":"Answer","text":"Hayır."}},` +
		`{"@type":"Question","name":"Online izlenebilir mi?","acceptedAnswer":{"@type":"Answer","text":"Evet."}}]}`
	if got := marshal(t, faq); got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}
}

func TestPlugin_NewTypesTakeKeys(t *testing.T) {
	// A loop of events or organizations keeps them all, as a loop of people does.
	site := newSite(t, jsonld.New(jsonld.Options{}), func(rc *collage.RenderContext) {
		for _, name := range []string{"Jam", "Meetup", "Konferans"} {
			jsonld.Emit(rc, jsonld.Event{Key: name, Name: name})
		}
		for _, name := range []string{"OSE", "Partner"} {
			jsonld.Emit(rc, jsonld.Organization{Key: name, Name: name})
		}
		jsonld.Emit(rc, jsonld.FAQPage{Key: "a"}, jsonld.FAQPage{Key: "b"})
	})

	count := map[string]int{}
	for _, node := range blocks(t, get(t, site, "/article")) {
		schemaType, _ := node["@type"].(string)
		count[schemaType]++
	}
	if count["Event"] != 3 || count["Organization"] != 2 || count["FAQPage"] != 2 {
		t.Errorf("blocks per type = %v, want 3 Event, 2 Organization, 2 FAQPage", count)
	}
}

func TestPlugin_ClosingScriptTagInAnEventCannotEscape(t *testing.T) {
	// The new nodes marshal through their own wire structs; the escaping that
	// keeps an Article's headline inside its block holds for them too.
	const payload = `</script><img src=x onerror=alert(1)>`
	site := newSite(t, jsonld.New(jsonld.Options{}), func(rc *collage.RenderContext) {
		jsonld.Emit(rc,
			jsonld.Event{Name: "x", Description: payload, Place: &jsonld.Place{Name: payload}},
			jsonld.FAQPage{Questions: []jsonld.Question{{Question: payload, Answer: payload}}},
			jsonld.Organization{Name: payload},
		)
	})

	body := get(t, site, "/article")
	if strings.Contains(body, "<img src=x") {
		t.Fatalf("the payload escaped the script block:\n%s", body)
	}
	if nodes := blocks(t, body); len(nodes) != 3 {
		t.Fatalf("got %d nodes, want 3", len(nodes))
	}
}
