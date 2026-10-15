// Package jsonld is a collage plugin that emits schema.org structured data into the
// document head.
//
//	app, err := collage.New(&collage.Config{
//		Plugins: []collage.Plugin{jsonld.New(jsonld.Options{SiteName: "The Wire"})},
//	})
//
// A page contributes its own structured data from a data handler, by hoisting nodes
// into the page's head:
//
//	func articleData(ctx context.Context, rc *collage.RenderContext) (view, []string, error) {
//		article, err := client.Article(ctx, rc.Param("slug"))
//		// ...
//		jsonld.Emit(rc, jsonld.Article{
//			Headline:      article.Title,
//			DatePublished: article.PublishedAt,
//			AuthorName:    article.Author,
//		})
//		return view{Article: article}, nil, nil
//	}
//
//	collage.NewFragment("article", "pages/article.html").
//		WithData(collage.DataHandler(articleData))
//
// The plugin reads them after the render and writes one <script> tag per node. It
// reads the render's data rather than the rendered HTML deliberately: the data
// handler already fetched the article, and parsing a headline back out of the markup
// to describe the markup is both slower and wrong the first time a template changes.
package jsonld

import (
	"context"
	"encoding/json"
	"html/template"
	"log/slog"
	"strconv"
	"strings"

	"github.com/Elagoht/collage/pkg/collage"
)

// Name is the plugin's name, and the key its configuration is found under.
const Name = "elagoht/jsonld"

// hoistArea is where the script tags go. Structured data belongs in the head, and
// the layout decides where in it with {{hoist "head"}}.
const hoistArea = "head"

// Emit adds nodes to the structured data for the render rc belongs to.
//
// It marshals and hoists immediately rather than collecting for later, because
// hoisting is what decides placement and it only works while the tree is rendering.
// An earlier version collected here and spliced the finished HTML in AfterRender,
// which worked and quietly decided a layout question on the layout's behalf.
//
// A node's key is its schema.org type plus its Key field, so a nested fragment's
// Article replaces the one a layout declared rather than sitting beside it — the
// innermost declaration of a key wins, which is what "more specific" means here.
// Two nodes of different types both appear, and so do two of one type with
// different Keys: the people on an about page are each a Person. A Node this
// package does not define has no Key and is keyed by its type alone.
func Emit(rc *collage.RenderContext, nodes ...Node) {
	if rc == nil || disabledFor(rc) {
		return
	}
	for _, node := range nodes {
		encoded, err := marshalNode(node)
		if err != nil {
			// One bad node must not cost the page its other nodes, or the page
			// itself. Structured data is an enhancement; failing a render over a
			// malformed one would be this package deciding otherwise on the site's
			// behalf. It is logged, so the missing node has a trace.
			loggerFor(rc).WarnContext(rc.Context(), "jsonld: a node was left out of the page",
				"type", node.Type(), "error", err)
			continue
		}
		rc.Hoist(hoistArea, hoistKey(node), encoded)
	}
}

// keyed is the nodes of this package, which carry a Key. It is unexported, and
// not part of Node, so adding it took nothing from a Node an application wrote.
type keyed interface {
	nodeKey() string
}

// Every node this package defines is keyed. One that is not would collapse a loop
// of them into the last, which is the bug Key exists to fix.
var (
	_ keyed = Raw{}
	_ keyed = Article{}
	_ keyed = BlogPosting{}
	_ keyed = Blog{}
	_ keyed = Person{}
	_ keyed = Organization{}
	_ keyed = WebSite{}
	_ keyed = BreadcrumbList{}
	_ keyed = Event{}
	_ keyed = FAQPage{}
)

// hoistKey is the key node is hoisted under. The Key is quoted, so whatever it
// holds stays one Key and cannot run into the type's half.
func hoistKey(node Node) string {
	key := Name + ":" + node.Type()
	if k, ok := node.(keyed); ok && k.nodeKey() != "" {
		key += "#" + strconv.Quote(k.nodeKey())
	}
	return key
}

// disabledKey marks a render the plugin was told to stay out of.
//
// Through the render's values rather than a package-level flag, because a
// package-level flag is state two applications in one process would fight over and
// tests would have to reset. The plugin writes it in OnBeforeRender; Emit reads it.
// An application that calls Emit without registering the plugin has nothing to read
// and emits, which is right: there is no configuration saying otherwise.
var disabledKey = collage.NewKey[bool](Name + ":disabled")

// loggerKey carries the application's logger to Emit, set by the plugin in
// OnBeforeRender. A render without the plugin logs through slog.Default.
var loggerKey = collage.NewKey[*slog.Logger](Name + ":logger")

func loggerFor(rc *collage.RenderContext) *slog.Logger {
	if log, ok := loggerKey.Get(rc); ok && log != nil {
		return log
	}
	return slog.Default()
}

func disabledFor(rc *collage.RenderContext) bool {
	disabled, _ := disabledKey.Get(rc)
	return disabled
}

// marshalNode renders one node as a script element.
//
// json.Marshal, not an Encoder with SetEscapeHTML(false): Marshal escapes "<", ">"
// and "&" to their \u form, and that is what keeps a node safe inside a <script>
// element. The HTML parser ends a script at the first "</script" in its text,
// wherever it appears — including inside a JSON string — so a headline containing
// one would close the block early and spill the rest of the JSON into the document
// as markup.
//
// The escaping is held twice over, which is worth knowing before changing either
// half: each node's own MarshalJSON escapes when it marshals its wire struct, and
// Marshal re-escapes a Marshaler's output when it compacts it. Disabling one layer
// changes nothing; disabling both opens the hole, which is what
// TestPlugin_ClosingScriptTagInContentCannotEscape demonstrates.
func marshalNode(node Node) (template.HTML, error) {
	encoded, err := json.Marshal(node)
	if err != nil {
		return "", err
	}
	var block strings.Builder
	block.WriteString(`<script type="application/ld+json">`)
	block.Write(encoded)
	block.WriteString("</script>")
	return template.HTML(block.String()), nil // any: a script element this package assembled from escaped JSON
}

// Options configures the plugin. Its section of the application's plugin
// configuration is decoded over it.
type Options struct {
	// SiteName names the site. When it is set, a WebSite node is declared on
	// every page — as a default, which a page emitting its own WebSite replaces.
	// Its url is the application's origin, collage.BaseURL, and is left out when
	// the application has none.
	SiteName string `json:"siteName"`
	// SearchURL is a search endpoint with "{query}" where the term goes. It is part
	// of the WebSite node, and ignored without SiteName.
	SearchURL string `json:"searchURL"`
	// Disabled turns structured data off without removing the plugin from the
	// application: the site-wide node is not declared, and every Emit in the render
	// returns without doing anything.
	Disabled bool `json:"disabled"`
}

// Plugin declares the site-wide node. Everything else a page emits goes through
// Emit, which needs no plugin at all — registering this one is how a site gets its
// WebSite node without every page repeating it.
type Plugin struct {
	opts Options
	log  *slog.Logger
}

// New returns a plugin with opts as its starting point, which the application's
// own configuration is then decoded over.
func New(opts Options) *Plugin { return &Plugin{opts: opts} }

func (p *Plugin) Name() string    { return Name }
func (p *Plugin) Version() string { return "0.5.0" }

func (p *Plugin) Init(_ context.Context, host collage.Host) error {
	p.log = host.Logger()
	var err error
	p.opts, err = collage.PluginConfig(host, p.opts)
	return err
}

func (p *Plugin) Shutdown(context.Context) error { return nil }

// OnBeforeRender declares the site-wide WebSite node.
//
// Before the tree renders, so it lands at depth zero and any page emitting its own
// WebSite wins — the plugin provides a default, the page provides the specific
// thing. Declaring it afterwards would mean splicing finished HTML, which is what
// hoisting exists to stop.
func (p *Plugin) OnBeforeRender(_ context.Context, ev *collage.BeforeRenderEvent) error {
	if ev.Context == nil {
		return nil
	}
	if p.log != nil {
		loggerKey.Set(ev.Context, p.log)
	}
	if p.opts.Disabled {
		// Marked before any fragment runs, so every Emit this render sees it.
		disabledKey.Set(ev.Context, true)
		return nil
	}
	if p.opts.SiteName == "" {
		return nil
	}
	Emit(ev.Context, WebSite{Name: p.opts.SiteName, URL: collage.BaseURL(ev.Context), SearchURL: p.opts.SearchURL})
	return nil
}
