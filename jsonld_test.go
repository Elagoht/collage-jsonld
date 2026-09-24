package jsonld_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/Elagoht/collage/pkg/collage"

	"github.com/Elagoht/collage-jsonld"
)

var templates = fstest.MapFS{
	"layouts/main.html": &fstest.MapFile{Data: []byte(
		`<!DOCTYPE html><html><head><title>t</title>{{hoist "head"}}</head><body>{{slot "content"}}</body></html>`)},
	"pages/article.html": &fstest.MapFile{Data: []byte(`<article>{{.Headline}}</article>`)},
	"layouts/bare.html":  &fstest.MapFile{Data: []byte(`<p>no head here</p>{{slot "content"}}</p>`)},
	"pages/bare.html":    &fstest.MapFile{Data: []byte(`<p>content</p>`)},
}

type articleView struct{ Headline string }

// newSite builds a site whose article page emits whatever emit returns.
func newSite(t *testing.T, p collage.Plugin, emit func(*collage.RenderContext)) http.Handler {
	t.Helper()

	app, err := collage.New(&collage.Config{
		Logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
		Template: collage.TemplateConfig{FS: templates, Extension: ".html"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if p != nil {
		if err := app.RegisterPlugin(p); err != nil {
			t.Fatalf("RegisterPlugin: %v", err)
		}
	}

	layout := collage.NewFragment("layout", "layouts/main.html").WithSlot("content", true, false).Build()
	article := collage.NewPage("article").
		WithLayout(layout).
		WithContent(collage.NewFragment("article", "pages/article.html").
			WithDataHandler(func(_ context.Context, rc *collage.RenderContext) (any, []string, error) { // any: the framework's own handler signature
				if emit != nil {
					emit(rc)
				}
				return articleView{Headline: "A Headline"}, nil, nil
			}).
			Build()).
		WithPath("en", "/article").
		Build()
	if err := app.RegisterPage(article); err != nil {
		t.Fatalf("RegisterPage: %v", err)
	}

	// A page with no <head> at all, to show the plugin declining rather than
	// inventing a place to write.
	bare := collage.NewPage("bare").
		WithLayout(collage.NewFragment("bare-layout", "layouts/bare.html").
			WithSlot("content", true, false).
			Build()).
		WithContent(collage.NewFragment("bare", "pages/bare.html").
			WithDataHandler(func(_ context.Context, rc *collage.RenderContext) (any, []string, error) { // any: the framework's own handler signature
				jsonld.Emit(rc, jsonld.Article{Headline: "unreachable"})
				return nil, nil, nil
			}).
			Build()).
		WithPath("en", "/bare").
		Build()
	if err := app.RegisterPage(bare); err != nil {
		t.Fatalf("RegisterPage(bare): %v", err)
	}

	return app.Handler()
}

var scriptBlock = regexp.MustCompile(`(?s)<script type="application/ld\+json">(.*?)</script>`)

// blocks returns the decoded JSON-LD nodes on a page, failing the test if any of
// them does not parse — which is the whole point of emitting them.
func blocks(t *testing.T, body string) []map[string]any { // any: what encoding/json decodes into
	t.Helper()
	var out []map[string]any // any: what encoding/json decodes into
	for _, m := range scriptBlock.FindAllStringSubmatch(body, -1) {
		var node map[string]any // any: what encoding/json decodes into
		if err := json.Unmarshal([]byte(m[1]), &node); err != nil {
			t.Fatalf("a JSON-LD block does not parse: %v\n%s", err, m[1])
		}
		out = append(out, node)
	}
	return out
}

func get(t *testing.T, h http.Handler, target string) string {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s = %d: %s", target, rec.Code, rec.Body.String())
	}
	return rec.Body.String()
}

func TestPlugin_EmitsWhatTheRenderContributed(t *testing.T) {
	published := time.Date(2026, 9, 20, 8, 0, 0, 0, time.UTC)
	site := newSite(t, jsonld.New(), func(rc *collage.RenderContext) {
		jsonld.Emit(rc, jsonld.Article{
			Headline:      "Seawalls Buy Time",
			DatePublished: published,
			AuthorName:    "Noor Haddad",
			Keywords:      []string{"climate", "adaptation"},
		})
	})

	nodes := blocks(t, get(t, site, "/article"))
	if len(nodes) != 1 {
		t.Fatalf("got %d nodes, want 1", len(nodes))
	}
	node := nodes[0]
	if node["@type"] != "Article" || node["@context"] != "https://schema.org" {
		t.Errorf("node = %v, want a schema.org Article", node)
	}
	if node["headline"] != "Seawalls Buy Time" {
		t.Errorf("headline = %v", node["headline"])
	}
	if node["datePublished"] != published.Format(time.RFC3339) {
		t.Errorf("datePublished = %v, want RFC3339", node["datePublished"])
	}
	author, ok := node["author"].(map[string]any) // any: what encoding/json decodes into
	if !ok || author["name"] != "Noor Haddad" {
		t.Errorf("author = %v", node["author"])
	}
}

func TestPlugin_EmitsIntoTheHead(t *testing.T) {
	site := newSite(t, jsonld.New(), func(rc *collage.RenderContext) {
		jsonld.Emit(rc, jsonld.Article{Headline: "x"})
	})

	body := get(t, site, "/article")
	head := strings.Index(body, "</head>")
	script := strings.Index(body, `<script type="application/ld+json">`)
	if script < 0 || script > head {
		t.Errorf("the block is not inside <head>: script at %d, </head> at %d", script, head)
	}
}

func TestPlugin_ClosingScriptTagInContentCannotEscape(t *testing.T) {
	// The HTML parser ends a script element at the first "</script" in its text,
	// wherever it appears — including inside a JSON string. A headline containing
	// one would close the block early and spill the rest of the JSON into the page
	// as markup: an injection, not a formatting bug.
	//
	// What prevents it is that json.Marshal escapes "<", ">" and "&" by default.
	// This test is what stands between that default and someone reaching for an
	// Encoder with SetEscapeHTML(false), which produces valid JSON and this hole.
	site := newSite(t, jsonld.New(), func(rc *collage.RenderContext) {
		jsonld.Emit(rc, jsonld.Article{Headline: `</script><img src=x onerror=alert(1)>`})
	})

	body := get(t, site, "/article")

	if strings.Contains(body, "<img src=x") {
		t.Fatalf("the payload escaped the script block:\n%s", body)
	}
	nodes := blocks(t, body)
	if len(nodes) != 1 {
		t.Fatalf("got %d nodes, want 1", len(nodes))
	}
	// Escaped, and still the same string once parsed.
	if nodes[0]["headline"] != `</script><img src=x onerror=alert(1)>` {
		t.Errorf("headline decoded to %q, want the original", nodes[0]["headline"])
	}
}

func TestPlugin_EmitsNothingWhenThereIsNothingToSay(t *testing.T) {
	// An empty block is worse than no block: a consumer that fails to parse one
	// may distrust the whole page.
	site := newSite(t, jsonld.New(), nil)

	if body := get(t, site, "/article"); strings.Contains(body, "application/ld+json") {
		t.Errorf("a script block was emitted for a page with no data:\n%s", body)
	}
}

func TestPlugin_DeclinesALayoutWithNoHoistArea(t *testing.T) {
	// A layout that never calls {{hoist "head"}} gets nothing. That is the layout
	// deciding, which is the point of hoisting: the plugin no longer searches the
	// finished HTML for somewhere to put itself.
	site := newSite(t, jsonld.New(), nil)

	if body := get(t, site, "/bare"); strings.Contains(body, "application/ld+json") {
		t.Errorf("data was written into a layout that asked for none:\n%s", body)
	}
}

func TestPlugin_AddsTheConfiguredWebSiteNode(t *testing.T) {
	site := newSite(t, jsonld.NewWith(jsonld.Config{
		SiteName:  "The Wire",
		SiteURL:   "https://thewire.example",
		SearchURL: "https://thewire.example/search?q={query}",
	}), func(rc *collage.RenderContext) {
		jsonld.Emit(rc, jsonld.Article{Headline: "x"})
	})

	nodes := blocks(t, get(t, site, "/article"))
	if len(nodes) != 2 {
		t.Fatalf("got %d nodes, want the article and the site", len(nodes))
	}

	var found bool
	for _, node := range nodes {
		if node["@type"] != "WebSite" {
			continue
		}
		found = true
		if node["name"] != "The Wire" {
			t.Errorf("site name = %v", node["name"])
		}
		action, ok := node["potentialAction"].(map[string]any) // any: what encoding/json decodes into
		if !ok || action["@type"] != "SearchAction" {
			t.Errorf("potentialAction = %v, want a SearchAction", node["potentialAction"])
		}
	}
	if !found {
		t.Error("no WebSite node was emitted")
	}
}

func TestPlugin_DoesNotDuplicateAPagesOwnWebSiteNode(t *testing.T) {
	site := newSite(t, jsonld.NewWith(jsonld.Config{SiteName: "The Wire"}), func(rc *collage.RenderContext) {
		jsonld.Emit(rc, jsonld.WebSite{Name: "Something Else"})
	})

	nodes := blocks(t, get(t, site, "/article"))
	if len(nodes) != 1 || nodes[0]["name"] != "Something Else" {
		t.Errorf("nodes = %v, want only the page's own WebSite", nodes)
	}
}

func TestPlugin_Disabled(t *testing.T) {
	site := newSite(t, jsonld.NewWith(jsonld.Config{SiteName: "The Wire", Disabled: true}), func(rc *collage.RenderContext) {
		jsonld.Emit(rc, jsonld.Article{Headline: "x"})
	})

	if body := get(t, site, "/article"); strings.Contains(body, "application/ld+json") {
		t.Error("data was emitted although the plugin is disabled")
	}
}

func TestPlugin_MultipleFragmentsEachContribute(t *testing.T) {
	// Emit appends, so two fragments can contribute without knowing about each
	// other — which is the only way a breadcrumb fragment and a content fragment
	// can both have a say.
	site := newSite(t, jsonld.New(), func(rc *collage.RenderContext) {
		jsonld.Emit(rc, jsonld.Article{Headline: "x"})
		jsonld.Emit(rc, jsonld.BreadcrumbList{Items: []jsonld.Breadcrumb{
			{Name: "Home", URL: "/"},
			{Name: "Climate", URL: "/category/climate"},
		}})
	})

	nodes := blocks(t, get(t, site, "/article"))
	if len(nodes) != 2 {
		t.Fatalf("got %d nodes, want 2", len(nodes))
	}
	for _, node := range nodes {
		if node["@type"] != "BreadcrumbList" {
			continue
		}
		items, ok := node["itemListElement"].([]any) // any: what encoding/json decodes into
		if !ok || len(items) != 2 {
			t.Fatalf("itemListElement = %v", node["itemListElement"])
		}
		first, _ := items[0].(map[string]any) // any: what encoding/json decodes into
		if first["position"] != float64(1) {
			t.Errorf("first position = %v, want 1 — schema.org positions are 1-based", first["position"])
		}
	}
}

func TestRaw_RejectsInvalidJSON(t *testing.T) {
	// Invalid JSON inside a script tag is worse than absent JSON: a consumer that
	// fails to parse it may discard every other node on the page with it.
	if _, err := json.Marshal(jsonld.Raw{SchemaType: "Event", JSON: []byte(`{"broken"`)}); err == nil {
		t.Fatal("Marshal accepted invalid JSON")
	}
}

func TestRaw_AddsTheContext(t *testing.T) {
	out, err := json.Marshal(jsonld.Raw{SchemaType: "Event", JSON: []byte(`{"@type":"Event","name":"x"}`)})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var node map[string]any // any: what encoding/json decodes into
	if err := json.Unmarshal(out, &node); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if node["@context"] != "https://schema.org" {
		t.Errorf("@context = %v", node["@context"])
	}
}

func TestRaw_KeepsTheCallersKeyOrder(t *testing.T) {
	// Decoding through a map sorted the keys, which put "description" before
	// "@type". Valid JSON-LD, and hard to read in a page's source.
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "context goes first, the rest as written",
			in:   `{"@type":"Event","name":"x","description":"d","startDate":"2026-10-01"}`,
			want: `{"@context":"https://schema.org","@type":"Event","name":"x","description":"d","startDate":"2026-10-01"}`,
		},
		{
			name: "nested values are copied through untouched",
			in:   `{"@type":"Event","location":{"name":"Hall","@type":"Place"},"offers":[{"price":"0","@type":"Offer"}]}`,
			want: `{"@context":"https://schema.org","@type":"Event","location":{"name":"Hall","@type":"Place"},"offers":[{"price":"0","@type":"Offer"}]}`,
		},
		{
			name: "whitespace around and inside the object",
			in:   "\n  { \"@type\" : \"Event\" ,\n\t\"name\" : \"x\" }  \n",
			want: `{"@context":"https://schema.org","@type":"Event","name":"x"}`,
		},
		{
			name: "an empty object",
			in:   `{}`,
			want: `{"@context":"https://schema.org"}`,
		},
		{
			name: "the caller's own context is replaced, not repeated",
			in:   `{"@type":"Event","@context":"http://example.org/other","name":"x"}`,
			want: `{"@context":"https://schema.org","@type":"Event","name":"x"}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := json.Marshal(jsonld.Raw{SchemaType: "Event", JSON: []byte(tc.in)})
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			if string(out) != tc.want {
				t.Errorf("got  %s\nwant %s", out, tc.want)
			}
		})
	}
}

func TestRaw_RejectsWhatIsNotAnObject(t *testing.T) {
	// There is nowhere to put "@context" in anything but an object. null used to
	// decode into a nil map and panic, which took the render down with it.
	for _, in := range []string{``, `null`, `[]`, `[{"@type":"Event"}]`, `"Event"`, `42`, `{} {}`, `{"@type":"Event",}`} {
		if out, err := json.Marshal(jsonld.Raw{SchemaType: "Event", JSON: []byte(in)}); err == nil {
			t.Errorf("Marshal(%q) = %s, want an error", in, out)
		}
	}
}

func TestRaw_ClosingScriptTagCannotEscape(t *testing.T) {
	// Raw copies the caller's values through verbatim, so a literal "</script>"
	// in one is the caller's bytes reaching the page unless Raw escapes them.
	payload := `</script><img src=x onerror=alert(1)>`
	// Hand-written, as a caller's literal would be: encoding/json would have
	// escaped the "<" before Raw ever saw it.
	raw := []byte(`{"@type":"Event","name":"` + payload + `"}`)

	out, err := jsonld.Raw{SchemaType: "Event", JSON: raw}.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}
	if strings.Contains(string(out), "<") {
		t.Errorf("MarshalJSON left a literal '<' in place: %s", out)
	}

	site := newSite(t, jsonld.New(), func(rc *collage.RenderContext) {
		jsonld.Emit(rc, jsonld.Raw{SchemaType: "Event", JSON: raw})
	})
	body := get(t, site, "/article")
	if strings.Contains(body, "<img src=x") {
		t.Fatalf("the payload escaped the script block:\n%s", body)
	}
	if nodes := blocks(t, body); len(nodes) != 1 || nodes[0]["name"] != payload {
		t.Errorf("nodes = %v, want one Event named with the original string", nodes)
	}
}

// marshal is json.Marshal for a node, failing the test on an error.
func marshal(t *testing.T, node jsonld.Node) string {
	t.Helper()
	out, err := json.Marshal(node)
	if err != nil {
		t.Fatalf("Marshal(%s): %v", node.Type(), err)
	}
	return string(out)
}

func TestBlogPosting_IsAnArticleUnderItsOwnType(t *testing.T) {
	// Every property Article has, a post has, spelled the same way: only the
	// "@type" differs.
	article := jsonld.Article{
		Headline:      "Seawalls Buy Time",
		Description:   "What a decade of seawalls bought.",
		URL:           "https://blog.example/seawalls",
		Section:       "Climate",
		Keywords:      []string{"climate", "adaptation"},
		DatePublished: time.Date(2026, 9, 20, 8, 0, 0, 0, time.UTC),
		DateModified:  time.Date(2026, 9, 21, 8, 0, 0, 0, time.UTC),
		AuthorName:    "Noor Haddad",
		AuthorURL:     "https://blog.example/about",
		PublisherName: "Noor's Notes",
		PublisherLogo: "https://blog.example/logo.png",
		ImageURL:      "https://blog.example/seawalls.jpg",
		WordCount:     1200,
	}

	got := marshal(t, jsonld.BlogPosting(article))
	want := strings.Replace(marshal(t, article), `"@type":"Article"`, `"@type":"BlogPosting"`, 1)
	if got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}
	if !strings.HasPrefix(got, `{"@context":"https://schema.org","@type":"BlogPosting",`) {
		t.Errorf("got %s, want a schema.org BlogPosting", got)
	}
}

func TestPerson_Marshal(t *testing.T) {
	cases := []struct {
		name   string
		person jsonld.Person
		want   string
	}{
		{
			name:   "only a name",
			person: jsonld.Person{Name: "Noor Haddad"},
			want:   `{"@context":"https://schema.org","@type":"Person","name":"Noor Haddad"}`,
		},
		{
			name: "every property",
			person: jsonld.Person{
				Name:        "Noor Haddad",
				Description: "Writes about the coast.",
				URL:         "https://blog.example/about",
				ImageURL:    "https://blog.example/noor.jpg",
				JobTitle:    "Climate reporter",
				SameAs:      []string{"https://github.com/noor", "https://mastodon.example/@noor"},
			},
			want: `{"@context":"https://schema.org","@type":"Person","name":"Noor Haddad",` +
				`"description":"Writes about the coast.","url":"https://blog.example/about",` +
				`"image":{"@type":"ImageObject","url":"https://blog.example/noor.jpg"},` +
				`"jobTitle":"Climate reporter",` +
				`"sameAs":["https://github.com/noor","https://mastodon.example/@noor"]}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := marshal(t, tc.person); got != tc.want {
				t.Errorf("got  %s\nwant %s", got, tc.want)
			}
		})
	}
}

func TestPlugin_BlogTypesEachGetTheirOwnBlock(t *testing.T) {
	// One key per schema.org type, so a BlogPosting does not replace an Article
	// or the Person beside it.
	site := newSite(t, jsonld.New(), func(rc *collage.RenderContext) {
		jsonld.Emit(rc,
			jsonld.Article{Headline: "x"},
			jsonld.BlogPosting{Headline: "y"},
			jsonld.Person{Name: "Noor Haddad"},
		)
	})

	nodes := blocks(t, get(t, site, "/article"))
	seen := map[string]bool{}
	for _, node := range nodes {
		if node["@context"] != "https://schema.org" {
			t.Errorf("node %v declares no schema.org context", node)
		}
		schemaType, _ := node["@type"].(string)
		seen[schemaType] = true
	}
	for _, want := range []string{"Article", "BlogPosting", "Person"} {
		if !seen[want] {
			t.Errorf("no %s block among %v", want, nodes)
		}
	}
	if len(nodes) != 3 {
		t.Errorf("got %d nodes, want 3", len(nodes))
	}
}
