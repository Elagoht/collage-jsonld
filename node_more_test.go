package jsonld_test

import (
	"testing"
	"time"

	jsonld "github.com/Elagoht/collage-jsonld"
)

func TestArticle_MoreFields(t *testing.T) {
	cases := []struct {
		name string
		node jsonld.Node
		want string
	}{
		{
			name: "authors, images, speakable and paywall",
			node: jsonld.Article{
				Headline:   "x",
				AuthorName: "Noor",
				Authors:    []jsonld.Agent{{Name: "Ada"}, {Kind: jsonld.AgentOrganization, Name: "Desk", URL: "https://desk.example"}},
				ImageURL:   "https://x.example/a.jpg",
				ImageURLs:  []string{"https://x.example/b.jpg"},
				Speakable:  []string{".lead"},
				Paywalled:  true,
				PaywallCSS: []string{".paid"},
			},
			want: `{"@context":"https://schema.org","@type":"Article","headline":"x",` +
				`"author":[{"@type":"Person","name":"Noor"},{"@type":"Person","name":"Ada"},` +
				`{"@type":"Organization","name":"Desk","url":"https://desk.example"}],` +
				`"image":["https://x.example/a.jpg","https://x.example/b.jpg"],` +
				`"speakable":{"@type":"SpeakableSpecification","cssSelector":[".lead"]},` +
				`"isAccessibleForFree":false,` +
				`"hasPart":[{"@type":"WebPageElement","isAccessibleForFree":false,"cssSelector":".paid"}]}`,
		},
		{
			name: "authors alone are a list, even of one",
			node: jsonld.Article{Headline: "x", Authors: []jsonld.Agent{{Name: "Ada"}}},
			want: `{"@context":"https://schema.org","@type":"Article","headline":"x","author":[{"@type":"Person","name":"Ada"}]}`,
		},
		{
			name: "author name alone is the object it always was",
			node: jsonld.Article{Headline: "x", AuthorName: "Noor", AuthorURL: "https://n.example"},
			want: `{"@context":"https://schema.org","@type":"Article","headline":"x","author":{"@type":"Person","name":"Noor","url":"https://n.example"}}`,
		},
		{
			name: "image URLs alone",
			node: jsonld.Article{Headline: "x", ImageURLs: []string{"https://x.example/b.jpg"}},
			want: `{"@context":"https://schema.org","@type":"Article","headline":"x","image":["https://x.example/b.jpg"]}`,
		},
		{
			name: "news article",
			node: jsonld.NewsArticle{ID: "https://x.example/n#news", Headline: "x", Paywalled: true},
			want: `{"@context":"https://schema.org","@id":"https://x.example/n#news","@type":"NewsArticle","headline":"x","isAccessibleForFree":false}`,
		},
		{
			name: "a blog post takes the new fields too",
			node: jsonld.BlogPosting{Headline: "y", Speakable: []string{"h1"}},
			want: `{"@context":"https://schema.org","@type":"BlogPosting","headline":"y","speakable":{"@type":"SpeakableSpecification","cssSelector":["h1"]}}`,
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

func TestPerson_MoreFields(t *testing.T) {
	person := jsonld.Person{
		Name:          "Noor Haddad",
		GivenName:     "Noor",
		FamilyName:    "Haddad",
		AlternateName: "N.H.",
		Telephone:     "+90 212 000 00 00",
		WorksFor:      jsonld.Agent{Name: "The Wire", URL: "https://wire.example"},
		Address:       &jsonld.PostalAddress{Locality: "Kadıköy", Country: "TR"},
	}
	want := `{"@context":"https://schema.org","@type":"Person","name":"Noor Haddad",` +
		`"givenName":"Noor","familyName":"Haddad","alternateName":"N.H.","telephone":"+90 212 000 00 00",` +
		`"worksFor":{"@type":"Organization","name":"The Wire","url":"https://wire.example"},` +
		`"address":{"@type":"PostalAddress","addressLocality":"Kadıköy","addressCountry":"TR"}}`
	if got := marshal(t, person); got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}

	// An employer given as a person stays one.
	person = jsonld.Person{Name: "Ada", WorksFor: jsonld.Agent{Kind: jsonld.AgentPerson, Name: "Noor"}}
	want = `{"@context":"https://schema.org","@type":"Person","name":"Ada","worksFor":{"@type":"Person","name":"Noor"}}`
	if got := marshal(t, person); got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}
}

func TestOrganization_MoreFields(t *testing.T) {
	org := jsonld.Organization{
		Name:                 "OSE",
		AlternateName:        "Oyun Sektörü Etkinlikleri",
		LegalName:            "OSE A.Ş.",
		FoundingDate:         time.Date(2020, 3, 1, 0, 0, 0, 0, time.UTC),
		NumberOfEmployees:    12,
		TaxID:                "1234567890",
		VatID:                "TR1234567890",
		DUNS:                 "123456789",
		LEICode:              "5493001KJTIIGC8Y1R12",
		NAICS:                "711310",
		GlobalLocationNumber: "8690000000000",
		ISO6523Code:          "0199:5493001KJTIIGC8Y1R12",
		ContactPoints: []jsonld.ContactPoint{{
			Telephone:         "+90 212 000 00 00",
			Email:             "destek@ose.example",
			ContactType:       "customer service",
			AreaServed:        []string{"TR"},
			AvailableLanguage: []string{"Turkish", "English"},
		}},
	}
	want := `{"@context":"https://schema.org","@type":"Organization","name":"OSE",` +
		`"alternateName":"Oyun Sektörü Etkinlikleri","legalName":"OSE A.Ş.","foundingDate":"2020-03-01",` +
		`"numberOfEmployees":{"@type":"QuantitativeValue","value":12},` +
		`"taxID":"1234567890","vatID":"TR1234567890","duns":"123456789","leiCode":"5493001KJTIIGC8Y1R12",` +
		`"naics":"711310","globalLocationNumber":"8690000000000","iso6523Code":"0199:5493001KJTIIGC8Y1R12",` +
		`"contactPoint":[{"@type":"ContactPoint","telephone":"+90 212 000 00 00","email":"destek@ose.example",` +
		`"contactType":"customer service","areaServed":["TR"],"availableLanguage":["Turkish","English"]}]}`
	if got := marshal(t, org); got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}
}

func TestEvent_PerformersAndPreviousStart(t *testing.T) {
	event := jsonld.Event{
		Name:              "Jam",
		Status:            jsonld.EventRescheduled,
		Performers:        []jsonld.Agent{{Name: "Ada"}, {Kind: jsonld.AgentPerformingGroup, Name: "Band", URL: "https://band.example"}},
		PreviousStartDate: time.Date(2026, 11, 1, 19, 0, 0, 0, time.UTC),
	}
	want := `{"@context":"https://schema.org","@type":"Event","name":"Jam",` +
		`"eventStatus":"https://schema.org/EventRescheduled",` +
		`"performer":[{"@type":"Person","name":"Ada"},{"@type":"PerformingGroup","name":"Band","url":"https://band.example"}],` +
		`"previousStartDate":"2026-11-01T19:00:00Z"}`
	if got := marshal(t, event); got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}
}
