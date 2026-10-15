package jsonld_test

import (
	"strings"
	"testing"
	"time"

	"github.com/Elagoht/collage/pkg/collage"

	jsonld "github.com/Elagoht/collage-jsonld"
)

func TestProduct_Marshal(t *testing.T) {
	cases := []struct {
		name    string
		product jsonld.Product
		want    string
	}{
		{
			name: "everything",
			product: jsonld.Product{
				ID:          "https://shop.example/p/mouse#product",
				Key:         "mouse",
				Name:        "Oyuncu Faresi",
				Description: "Kablosuz.",
				ImageURLs:   []string{"https://shop.example/m-1x1.jpg", "https://shop.example/m-16x9.jpg"},
				Brand:       "Acme",
				SKU:         "M-1",
				GTIN:        "8690000000001",
				MPN:         "AC-M1",
				Color:       "Siyah",
				Material:    "Plastik",
				Size:        "M",
				Offers: []jsonld.Offer{{
					URL:             "https://shop.example/p/mouse",
					Price:           "899.90",
					Currency:        "TRY",
					Availability:    jsonld.InStock,
					ItemCondition:   jsonld.NewCondition,
					PriceValidUntil: time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
				}},
				ShippingDetails: []jsonld.ShippingDetails{{
					Rate:         &jsonld.Money{Value: "0", Currency: "TRY"},
					Country:      "TR",
					HandlingDays: &jsonld.DayRange{Min: 0, Max: 1},
					TransitDays:  &jsonld.DayRange{Min: 1, Max: 3},
				}},
				ReturnPolicy: &jsonld.ReturnPolicy{
					Country:  "TR",
					Category: jsonld.ReturnFiniteWindow,
					Days:     14,
					Method:   jsonld.ReturnByMail,
					Fees:     jsonld.FreeReturn,
				},
				AggregateRating: &jsonld.AggregateRating{Value: 4.4, ReviewCount: 89},
				Reviews: []jsonld.Review{{
					Author: jsonld.Agent{Name: "Ada"},
					Rating: &jsonld.Rating{Value: 5},
					Body:   "Çok iyi.",
				}},
			},
			want: `{"@context":"https://schema.org","@id":"https://shop.example/p/mouse#product","@type":"Product",` +
				`"name":"Oyuncu Faresi","description":"Kablosuz.",` +
				`"image":["https://shop.example/m-1x1.jpg","https://shop.example/m-16x9.jpg"],` +
				`"brand":{"@type":"Brand","name":"Acme"},"sku":"M-1","gtin":"8690000000001","mpn":"AC-M1",` +
				`"color":"Siyah","material":"Plastik","size":"M",` +
				`"offers":[{"@type":"Offer","url":"https://shop.example/p/mouse","price":"899.90","priceCurrency":"TRY",` +
				`"availability":"https://schema.org/InStock","itemCondition":"https://schema.org/NewCondition",` +
				`"priceValidUntil":"2026-12-31",` +
				`"shippingDetails":[{"@type":"OfferShippingDetails",` +
				`"shippingRate":{"@type":"MonetaryAmount","value":"0","currency":"TRY"},` +
				`"shippingDestination":{"@type":"DefinedRegion","addressCountry":"TR"},` +
				`"deliveryTime":{"@type":"ShippingDeliveryTime",` +
				`"handlingTime":{"@type":"QuantitativeValue","minValue":0,"maxValue":1,"unitCode":"DAY"},` +
				`"transitTime":{"@type":"QuantitativeValue","minValue":1,"maxValue":3,"unitCode":"DAY"}}}],` +
				`"hasMerchantReturnPolicy":{"@type":"MerchantReturnPolicy","applicableCountry":"TR",` +
				`"returnPolicyCategory":"https://schema.org/MerchantReturnFiniteReturnWindow","merchantReturnDays":14,` +
				`"returnMethod":"https://schema.org/ReturnByMail","returnFees":"https://schema.org/FreeReturn"}}],` +
				`"aggregateRating":{"@type":"AggregateRating","ratingValue":4.4,"reviewCount":89},` +
				`"review":[{"@type":"Review","author":{"@type":"Person","name":"Ada"},` +
				`"reviewRating":{"@type":"Rating","ratingValue":5},"reviewBody":"Çok iyi."}]}`,
		},
		{
			name:    "a name alone",
			product: jsonld.Product{Name: "Fare"},
			want:    `{"@context":"https://schema.org","@type":"Product","name":"Fare"}`,
		},
		{
			name: "shipping without offers has nowhere to go",
			product: jsonld.Product{Name: "Fare", ShippingDetails: []jsonld.ShippingDetails{{Rate: &jsonld.Money{Value: "0"}, Country: "TR"}},
				ReturnPolicy: &jsonld.ReturnPolicy{Category: jsonld.ReturnNotPermitted}},
			want: `{"@context":"https://schema.org","@type":"Product","name":"Fare"}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := marshal(t, tc.product); got != tc.want {
				t.Errorf("got  %s\nwant %s", got, tc.want)
			}
		})
	}
}

func TestSoftwareApplication_Marshal(t *testing.T) {
	cases := []struct {
		name string
		app  jsonld.SoftwareApplication
		want string
	}{
		{
			name: "everything",
			app: jsonld.SoftwareApplication{
				Kind:                jsonld.AppVideoGame,
				ID:                  "https://game.example/#game",
				Name:                "Kale",
				Description:         "Bir strateji oyunu.",
				URL:                 "https://game.example",
				ImageURLs:           []string{"https://game.example/cover.png"},
				OperatingSystem:     "Windows, macOS",
				ApplicationCategory: "GameApplication",
				Offers:              []jsonld.Offer{{Price: "149.90", Currency: "TRY"}},
				AggregateRating:     &jsonld.AggregateRating{Value: 4.8, RatingCount: 1200},
			},
			want: `{"@context":"https://schema.org","@id":"https://game.example/#game","@type":"VideoGame",` +
				`"name":"Kale","description":"Bir strateji oyunu.","url":"https://game.example",` +
				`"image":["https://game.example/cover.png"],` +
				`"operatingSystem":"Windows, macOS","applicationCategory":"GameApplication",` +
				`"offers":[{"@type":"Offer","price":"149.90","priceCurrency":"TRY"}],` +
				`"aggregateRating":{"@type":"AggregateRating","ratingValue":4.8,"ratingCount":1200}}`,
		},
		{
			name: "no kind is SoftwareApplication",
			app:  jsonld.SoftwareApplication{Name: "Araç"},
			want: `{"@context":"https://schema.org","@type":"SoftwareApplication","name":"Araç"}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := marshal(t, tc.app); got != tc.want {
				t.Errorf("got  %s\nwant %s", got, tc.want)
			}
		})
	}
	if got := (jsonld.SoftwareApplication{Kind: jsonld.AppMobile}).Type(); got != "MobileApplication" {
		t.Errorf("Type() = %q, want MobileApplication", got)
	}
}

func TestMovie_Marshal(t *testing.T) {
	cases := []struct {
		name  string
		movie jsonld.Movie
		want  string
	}{
		{
			name: "everything",
			movie: jsonld.Movie{
				ID:          "https://film.example/m/kis#movie",
				Name:        "Kış Uykusu",
				URL:         "https://film.example/m/kis",
				ImageURLs:   []string{"https://film.example/kis.jpg"},
				DateCreated: time.Date(2014, 5, 16, 0, 0, 0, 0, time.UTC),
				Director:    jsonld.Agent{Name: "Nuri Bilge Ceylan"},
				Actors:      []jsonld.Agent{{Name: "Haluk Bilginer"}, {Name: ""}, {Name: "Melisa Sözen"}},
				Duration:    3*time.Hour + 16*time.Minute,
				Reviews:     []jsonld.Review{{Author: jsonld.Agent{Name: "Ada"}, Rating: &jsonld.Rating{Value: 9, BestRating: 10}}},
			},
			want: `{"@context":"https://schema.org","@id":"https://film.example/m/kis#movie","@type":"Movie",` +
				`"name":"Kış Uykusu","url":"https://film.example/m/kis",` +
				`"image":["https://film.example/kis.jpg"],"dateCreated":"2014-05-16",` +
				`"director":{"@type":"Person","name":"Nuri Bilge Ceylan"},` +
				`"actor":[{"@type":"Person","name":"Haluk Bilginer"},{"@type":"Person","name":"Melisa Sözen"}],` +
				`"duration":"PT3H16M",` +
				`"review":[{"@type":"Review","author":{"@type":"Person","name":"Ada"},` +
				`"reviewRating":{"@type":"Rating","ratingValue":9,"bestRating":10}}]}`,
		},
		{
			name:  "a name alone",
			movie: jsonld.Movie{Name: "Kış Uykusu"},
			want:  `{"@context":"https://schema.org","@type":"Movie","name":"Kış Uykusu"}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := marshal(t, tc.movie); got != tc.want {
				t.Errorf("got  %s\nwant %s", got, tc.want)
			}
		})
	}
}

func TestEmployerAggregateRating_Marshal(t *testing.T) {
	cases := []struct {
		name   string
		rating jsonld.EmployerAggregateRating
		want   string
	}{
		{
			name: "everything",
			rating: jsonld.EmployerAggregateRating{
				ID:          "https://jobs.example/acme#rating",
				Employer:    jsonld.Agent{Name: "Acme", URL: "https://acme.example", SameAs: []string{"https://linkedin.com/company/acme"}},
				Value:       3.8,
				RatingCount: 120,
				ReviewCount: 40,
				BestRating:  5,
				WorstRating: 1,
			},
			want: `{"@context":"https://schema.org","@id":"https://jobs.example/acme#rating","@type":"EmployerAggregateRating",` +
				`"itemReviewed":{"@type":"Organization","name":"Acme","url":"https://acme.example",` +
				`"sameAs":["https://linkedin.com/company/acme"]},` +
				`"ratingValue":3.8,"ratingCount":120,"reviewCount":40,"bestRating":5,"worstRating":1}`,
		},
		{
			name:   "minimal",
			rating: jsonld.EmployerAggregateRating{Employer: jsonld.Agent{Name: "Acme"}, Value: 4, RatingCount: 3},
			want: `{"@context":"https://schema.org","@type":"EmployerAggregateRating",` +
				`"itemReviewed":{"@type":"Organization","name":"Acme"},"ratingValue":4,"ratingCount":3}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := marshal(t, tc.rating); got != tc.want {
				t.Errorf("got  %s\nwant %s", got, tc.want)
			}
		})
	}
}

func TestReviewAndAggregateRating_StandaloneNestTheItemWithoutContext(t *testing.T) {
	item := jsonld.Product{Name: "Fare", Brand: "Acme"}
	const nestedItem = `"itemReviewed":{"@type":"Product","name":"Fare","brand":{"@type":"Brand","name":"Acme"}}`

	review := jsonld.Review{
		ID:            "https://shop.example/r/1",
		ItemReviewed:  item,
		Author:        jsonld.Agent{Name: "Ada"},
		Rating:        &jsonld.Rating{Value: 4, BestRating: 5, WorstRating: 1},
		Name:          "İyi",
		Body:          "Rahat.",
		DatePublished: time.Date(2026, 10, 1, 9, 0, 0, 0, time.FixedZone("", 3*60*60)),
		Publisher:     jsonld.Agent{Kind: jsonld.AgentOrganization, Name: "Shop"},
	}
	wantReview := `{"@context":"https://schema.org","@id":"https://shop.example/r/1","@type":"Review",` + nestedItem + `,` +
		`"name":"İyi","author":{"@type":"Person","name":"Ada"},` +
		`"reviewRating":{"@type":"Rating","ratingValue":4,"bestRating":5,"worstRating":1},` +
		`"reviewBody":"Rahat.","datePublished":"2026-10-01T09:00:00+03:00",` +
		`"publisher":{"@type":"Organization","name":"Shop"}}`
	if got := marshal(t, review); got != wantReview {
		t.Errorf("review\ngot  %s\nwant %s", got, wantReview)
	}

	rating := jsonld.AggregateRating{ItemReviewed: item, Value: 4.5, RatingCount: 10}
	wantRating := `{"@context":"https://schema.org","@type":"AggregateRating",` + nestedItem + `,` +
		`"ratingValue":4.5,"ratingCount":10}`
	if got := marshal(t, rating); got != wantRating {
		t.Errorf("aggregate rating\ngot  %s\nwant %s", got, wantRating)
	}
}

func TestPlugin_CommerceTypesTakeKeys(t *testing.T) {
	// A loop of products, apps, films or employer ratings keeps them all.
	site := newSite(t, jsonld.New(jsonld.Options{}), func(rc *collage.RenderContext) {
		for _, name := range []string{"Fare", "Klavye", "Kulaklık"} {
			jsonld.Emit(rc, jsonld.Product{Key: name, Name: name})
		}
		jsonld.Emit(rc,
			jsonld.SoftwareApplication{Key: "a", Name: "a"}, jsonld.SoftwareApplication{Key: "b", Name: "b"},
			jsonld.Movie{Key: "a", Name: "a"}, jsonld.Movie{Key: "b", Name: "b"},
			jsonld.EmployerAggregateRating{Key: "a", Employer: jsonld.Agent{Name: "a"}},
			jsonld.EmployerAggregateRating{Key: "b", Employer: jsonld.Agent{Name: "b"}},
		)
	})

	count := map[string]int{}
	for _, node := range blocks(t, get(t, site, "/article")) {
		schemaType, _ := node["@type"].(string)
		count[schemaType]++
	}
	if count["Product"] != 3 || count["SoftwareApplication"] != 2 || count["Movie"] != 2 || count["EmployerAggregateRating"] != 2 {
		t.Errorf("blocks per type = %v, want 3 Product, 2 SoftwareApplication, 2 Movie, 2 EmployerAggregateRating", count)
	}
}

func TestPlugin_ClosingScriptTagInAProductCannotEscape(t *testing.T) {
	const payload = `</script><img src=x onerror=alert(1)>`
	site := newSite(t, jsonld.New(jsonld.Options{}), func(rc *collage.RenderContext) {
		jsonld.Emit(rc,
			jsonld.Product{Name: payload, Brand: payload, Offers: []jsonld.Offer{{Price: "1"}},
				ShippingDetails: []jsonld.ShippingDetails{{Rate: &jsonld.Money{Value: payload}}}},
			jsonld.SoftwareApplication{Name: payload},
			jsonld.Movie{Name: payload, Director: jsonld.Agent{Name: payload}},
			jsonld.EmployerAggregateRating{Employer: jsonld.Agent{Name: payload}},
			jsonld.Review{ItemReviewed: jsonld.Product{Name: payload}, Body: payload},
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

func TestProduct_AnOffersOwnPolicyWinsOverTheProducts(t *testing.T) {
	// The product's shipping and returns are defaults: an offer that sets its own
	// keeps them, and one that sets none takes the product's.
	product := jsonld.Product{
		Name: "Gamepad",
		Offers: []jsonld.Offer{
			{Price: "100", Currency: "TRY", ReturnPolicy: &jsonld.ReturnPolicy{Country: "TR", Category: jsonld.ReturnNotPermitted},
				ShippingDetails: []jsonld.ShippingDetails{{Rate: &jsonld.Money{Value: "0", Currency: "TRY"}, Country: "TR"}}},
			{Price: "90", Currency: "TRY"},
		},
		ShippingDetails: []jsonld.ShippingDetails{{Rate: &jsonld.Money{Value: "50", Currency: "TRY"}, Country: "TR"}},
		ReturnPolicy:    &jsonld.ReturnPolicy{Country: "TR", Category: jsonld.ReturnFiniteWindow, Days: 14},
	}
	want := `{"@context":"https://schema.org","@type":"Product","name":"Gamepad","offers":[` +
		`{"@type":"Offer","price":"100","priceCurrency":"TRY",` +
		`"shippingDetails":[{"@type":"OfferShippingDetails","shippingRate":{"@type":"MonetaryAmount","value":"0","currency":"TRY"},"shippingDestination":{"@type":"DefinedRegion","addressCountry":"TR"}}],` +
		`"hasMerchantReturnPolicy":{"@type":"MerchantReturnPolicy","applicableCountry":"TR","returnPolicyCategory":"https://schema.org/MerchantReturnNotPermitted"}},` +
		`{"@type":"Offer","price":"90","priceCurrency":"TRY",` +
		`"shippingDetails":[{"@type":"OfferShippingDetails","shippingRate":{"@type":"MonetaryAmount","value":"50","currency":"TRY"},"shippingDestination":{"@type":"DefinedRegion","addressCountry":"TR"}}],` +
		`"hasMerchantReturnPolicy":{"@type":"MerchantReturnPolicy","applicableCountry":"TR","returnPolicyCategory":"https://schema.org/MerchantReturnFiniteReturnWindow","merchantReturnDays":14}}]}`
	if got := marshal(t, product); got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}
}
