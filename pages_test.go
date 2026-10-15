package jsonld_test

import (
	"strings"
	"testing"
	"time"

	"github.com/Elagoht/collage/pkg/collage"

	jsonld "github.com/Elagoht/collage-jsonld"
)

func TestWebPage_Marshal(t *testing.T) {
	when := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	cases := []struct {
		name string
		page jsonld.WebPage
		want string
	}{
		{
			name: "everything",
			page: jsonld.WebPage{
				ID:            "https://x.example/about#page",
				PageType:      jsonld.PageAbout,
				Name:          "Hakkımızda",
				Description:   "Biz kimiz.",
				URL:           "https://x.example/about",
				InLanguage:    "tr",
				DatePublished: when,
				DateModified:  when,
				ImageURL:      "https://x.example/a.png",
				IsPartOf:      "https://x.example/#site",
				Speakable:     []string{".lead"},
				Paywalled:     true,
				PaywallCSS:    []string{".paid"},
			},
			want: `{"@context":"https://schema.org","@id":"https://x.example/about#page","@type":"AboutPage",` +
				`"name":"Hakkımızda","description":"Biz kimiz.","url":"https://x.example/about","inLanguage":"tr",` +
				`"datePublished":"2026-01-02T03:04:05Z","dateModified":"2026-01-02T03:04:05Z",` +
				`"image":{"@type":"ImageObject","url":"https://x.example/a.png"},` +
				`"isPartOf":{"@id":"https://x.example/#site"},` +
				`"speakable":{"@type":"SpeakableSpecification","cssSelector":[".lead"]},` +
				`"isAccessibleForFree":false,` +
				`"hasPart":[{"@type":"WebPageElement","isAccessibleForFree":false,"cssSelector":".paid"}]}`,
		},
		{
			name: "no page type is a WebPage",
			page: jsonld.WebPage{Name: "x"},
			want: `{"@context":"https://schema.org","@type":"WebPage","name":"x"}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := marshal(t, tc.page); got != tc.want {
				t.Errorf("got  %s\nwant %s", got, tc.want)
			}
		})
	}
	if got := (jsonld.WebPage{PageType: jsonld.PageContact}).Type(); got != "ContactPage" {
		t.Errorf("Type() = %q, want ContactPage", got)
	}
}

func TestProfilePage_Marshal(t *testing.T) {
	when := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	cases := []struct {
		name string
		page jsonld.ProfilePage
		want string
	}{
		{
			name: "a person, without a context of its own",
			page: jsonld.ProfilePage{
				MainEntity:   jsonld.Person{ID: "https://x.example/#noor", Name: "Noor"},
				DateCreated:  when,
				DateModified: when,
			},
			want: `{"@context":"https://schema.org","@type":"ProfilePage",` +
				`"dateCreated":"2026-01-02T03:04:05Z","dateModified":"2026-01-02T03:04:05Z",` +
				`"mainEntity":{"@id":"https://x.example/#noor","@type":"Person","name":"Noor"}}`,
		},
		{
			name: "an organization",
			page: jsonld.ProfilePage{ID: "https://x.example/ose#page", MainEntity: jsonld.Organization{Name: "OSE"}},
			want: `{"@context":"https://schema.org","@id":"https://x.example/ose#page","@type":"ProfilePage",` +
				`"mainEntity":{"@type":"Organization","name":"OSE"}}`,
		},
		{
			name: "no entity",
			page: jsonld.ProfilePage{},
			want: `{"@context":"https://schema.org","@type":"ProfilePage"}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := marshal(t, tc.page); got != tc.want {
				t.Errorf("got  %s\nwant %s", got, tc.want)
			}
		})
	}
}

func TestPlugin_PageTypesTakeKeys(t *testing.T) {
	// Keyed loops keep every node, and a WebPage's key is its PageType: an
	// AboutPage and a WebPage are two nodes.
	site := newSite(t, jsonld.New(jsonld.Options{}), func(rc *collage.RenderContext) {
		for _, k := range []string{"a", "b"} {
			jsonld.Emit(rc,
				jsonld.NewsArticle{Key: k, Headline: k},
				jsonld.WebPage{Key: k, Name: k},
				jsonld.ProfilePage{Key: k, MainEntity: jsonld.Person{Name: k}},
			)
		}
		jsonld.Emit(rc, jsonld.WebPage{Name: "plain"}, jsonld.WebPage{PageType: jsonld.PageAbout, Name: "about"})
	})

	count := map[string]int{}
	for _, node := range blocks(t, get(t, site, "/article")) {
		schemaType, _ := node["@type"].(string)
		count[schemaType]++
	}
	if count["NewsArticle"] != 2 || count["WebPage"] != 3 || count["AboutPage"] != 1 || count["ProfilePage"] != 2 {
		t.Errorf("blocks per type = %v, want 2 NewsArticle, 3 WebPage, 1 AboutPage, 2 ProfilePage", count)
	}
}

func TestPlugin_ClosingScriptTagInPageFieldsCannotEscape(t *testing.T) {
	const payload = `</script><img src=x onerror=alert(1)>`
	site := newSite(t, jsonld.New(jsonld.Options{}), func(rc *collage.RenderContext) {
		jsonld.Emit(rc,
			jsonld.WebPage{Name: payload, PaywallCSS: []string{payload}},
			jsonld.ProfilePage{MainEntity: jsonld.Person{Name: payload, GivenName: payload}},
			jsonld.NewsArticle{Headline: "x", Authors: []jsonld.Agent{{Name: payload}}},
			jsonld.Organization{Name: "x", ContactPoints: []jsonld.ContactPoint{{ContactType: payload}}},
		)
	})

	body := get(t, site, "/article")
	if strings.Contains(body, "<img src=x") {
		t.Fatalf("the payload escaped the script block:\n%s", body)
	}
	nodes := blocks(t, body)
	if len(nodes) != 4 {
		t.Fatalf("got %d nodes, want 4", len(nodes))
	}
	for _, node := range nodes {
		if node["@type"] != "ProfilePage" {
			continue
		}
		entity, _ := node["mainEntity"].(map[string]any) // any: what encoding/json decodes into
		if entity["name"] != payload {
			t.Errorf("mainEntity name decoded to %q, want the original", entity["name"])
		}
	}
}
