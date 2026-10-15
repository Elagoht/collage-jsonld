package jsonld_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/Elagoht/collage/pkg/collage"

	jsonld "github.com/Elagoht/collage-jsonld"
)

func TestEmit_SkipsNilNodes(t *testing.T) {
	// A nil node, and a typed nil behind the interface, are left out with a
	// warning; asking either its type would panic, and the nodes beside them
	// still appear.
	var person *jsonld.Person
	site := newSite(t, jsonld.New(jsonld.Options{}), func(rc *collage.RenderContext) {
		jsonld.Emit(rc, nil, person, jsonld.Organization{Name: "OSE"})
	})

	nodes := blocks(t, get(t, site, "/article"))
	if len(nodes) != 1 || nodes[0]["@type"] != "Organization" {
		t.Fatalf("nodes = %v, want only the Organization", nodes)
	}
}

func TestNested_NilNodeIsLeftOut(t *testing.T) {
	var person *jsonld.Person
	cases := []struct {
		name string
		node jsonld.Node
		want string
	}{
		{
			name: "profile page with a typed-nil entity",
			node: jsonld.ProfilePage{MainEntity: person},
			want: `{"@context":"https://schema.org","@type":"ProfilePage"}`,
		},
		{
			name: "review of a typed-nil item",
			node: jsonld.Review{ItemReviewed: person, Body: "İyi."},
			want: `{"@context":"https://schema.org","@type":"Review","reviewBody":"İyi."}`,
		},
		{
			name: "aggregate rating of a typed-nil item",
			node: jsonld.AggregateRating{ItemReviewed: person, Value: 4, RatingCount: 2},
			want: `{"@context":"https://schema.org","@type":"AggregateRating","ratingValue":4,"ratingCount":2}`,
		},
		{
			name: "list entry with a typed-nil node",
			node: jsonld.ItemList{Items: []jsonld.ListEntry{{URL: "https://x.example/a", Node: person}}},
			want: `{"@context":"https://schema.org","@type":"ItemList","itemListElement":[{"@type":"ListItem","position":1,"url":"https://x.example/a"}]}`,
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

func TestQAPage_AnswerCountDefaultsToTheAnswersListed(t *testing.T) {
	page := jsonld.QAPage{Question: jsonld.QAQuestion{
		Name:      "Neden?",
		Accepted:  []jsonld.QAAnswer{{Text: "Çünkü."}},
		Suggested: []jsonld.QAAnswer{{Text: "Bilmem."}, {Text: "Belki."}},
	}}
	var decoded struct {
		MainEntity struct {
			AnswerCount int `json:"answerCount"`
		} `json:"mainEntity"`
	}
	if err := json.Unmarshal([]byte(marshal(t, page)), &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.MainEntity.AnswerCount != 3 {
		t.Errorf("answerCount = %d, want 3", decoded.MainEntity.AnswerCount)
	}
}

func TestPaywallCSS_ImpliesNotFree(t *testing.T) {
	cases := []struct {
		name string
		node jsonld.Node
		want string
	}{
		{
			name: "article",
			node: jsonld.Article{Headline: "x", PaywallCSS: []string{".paid"}},
			want: `{"@context":"https://schema.org","@type":"Article","headline":"x","isAccessibleForFree":false,` +
				`"hasPart":[{"@type":"WebPageElement","isAccessibleForFree":false,"cssSelector":".paid"}]}`,
		},
		{
			name: "web page",
			node: jsonld.WebPage{Name: "x", PaywallCSS: []string{".paid"}},
			want: `{"@context":"https://schema.org","@type":"WebPage","name":"x","isAccessibleForFree":false,` +
				`"hasPart":[{"@type":"WebPageElement","isAccessibleForFree":false,"cssSelector":".paid"}]}`,
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

func TestFixes_Marshal(t *testing.T) {
	cases := []struct {
		name string
		node jsonld.Node
		want string
	}{
		{
			name: "an event organized by a person, at an empty place",
			node: jsonld.Event{Name: "Meetup", Place: &jsonld.Place{},
				Organizer: jsonld.Agent{Kind: jsonld.AgentPerson, Name: "Ada", SameAs: []string{"https://x.example/ada"}}},
			want: `{"@context":"https://schema.org","@type":"Event","name":"Meetup",` +
				`"organizer":{"@type":"Person","name":"Ada","sameAs":["https://x.example/ada"]}}`,
		},
		{
			name: "one named author beside a nameless one is still one author",
			node: jsonld.Article{Headline: "x", AuthorName: "Ada", Authors: []jsonld.Agent{{}}},
			want: `{"@context":"https://schema.org","@type":"Article","headline":"x","author":{"@type":"Person","name":"Ada"}}`,
		},
		{
			name: "same-day handling, and a fee for returning",
			node: jsonld.Product{Name: "Fare", Offers: []jsonld.Offer{{Price: "100", Currency: "TRY"}},
				ShippingDetails: []jsonld.ShippingDetails{{HandlingDays: &jsonld.DayRange{}}},
				ReturnPolicy: &jsonld.ReturnPolicy{Fees: jsonld.ReturnShippingFees,
					ShippingFee: &jsonld.Money{Value: "30", Currency: "TRY"}}},
			want: `{"@context":"https://schema.org","@type":"Product","name":"Fare","offers":[{"@type":"Offer","price":"100","priceCurrency":"TRY",` +
				`"shippingDetails":[{"@type":"OfferShippingDetails","deliveryTime":{"@type":"ShippingDeliveryTime",` +
				`"handlingTime":{"@type":"QuantitativeValue","minValue":0,"maxValue":0,"unitCode":"DAY"}}}],` +
				`"hasMerchantReturnPolicy":{"@type":"MerchantReturnPolicy","returnFees":"https://schema.org/ReturnShippingFees",` +
				`"returnShippingFeesAmount":{"@type":"MonetaryAmount","value":"30","currency":"TRY"}}}]}`,
		},
		{
			name: "a salary with no amount has no value",
			node: jsonld.JobPosting{Title: "x", Salary: &jsonld.Salary{Currency: "TRY", Unit: jsonld.SalaryPerMonth}},
			want: `{"@context":"https://schema.org","@type":"JobPosting","title":"x",` +
				`"baseSalary":{"@type":"MonetaryAmount","currency":"TRY"}}`,
		},
		{
			name: "a recipe's url and reviews",
			node: jsonld.Recipe{Name: "Menemen", URL: "https://x.example/menemen",
				Reviews: []jsonld.Review{{Author: jsonld.Agent{Name: "Ada"}, Rating: &jsonld.Rating{Value: 5},
					ContentReferenceTime: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)}}},
			want: `{"@context":"https://schema.org","@type":"Recipe","name":"Menemen","url":"https://x.example/menemen",` +
				`"review":[{"@type":"Review","author":{"@type":"Person","name":"Ada"},` +
				`"reviewRating":{"@type":"Rating","ratingValue":5},"contentReferenceTime":"2026-09-01T00:00:00Z"}]}`,
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
