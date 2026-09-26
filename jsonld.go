// Package jsonld is a collage plugin that emits schema.org structured data into the
// document head.
//
//	app, err := collage.New(&collage.Config{
//		Plugins: []collage.Plugin{jsonld.New()},
//	})
//
// A page contributes its own structured data from a data handler, by putting nodes
// into the render's shared data:
//
//	func articleData(ctx context.Context, rc *collage.RenderContext) (any, []string, error) {
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
	"strings"

	"github.com/Elagoht/collage/pkg/collage"
)

// Name is the plugin's name, and the key its configuration is found under.
const Name = "elagoht/jsonld"

// HoistArea is where the script tags go. Structured data belongs in the head, and
// the layout decides where in it with {{hoist "head"}}.
const HoistArea = "head"

// Emit adds nodes to the structured data for the render rc belongs to.
//
// It marshals and hoists immediately rather than collecting for later, because
// hoisting is what decides placement and it only works while the tree is rendering.
// An earlier version collected here and spliced the finished HTML in AfterRender,
// which worked and quietly decided a layout question on the layout's behalf.
//
// One key per schema.org type, so a nested fragment's Article replaces the one a
// layout declared rather than sitting beside it — the innermost declaration of a key
// wins, which is what "more specific" means here. Two nodes of different types both
// appear.
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
			// behalf.
			continue
		}
		rc.Hoist(HoistArea, Name+":"+node.Type(), encoded)
	}
}

// disabledKey marks a render the plugin was told to stay out of.
//
// Through the render's shared data rather than a package variable, because a
// package variable is state two applications in one process would fight over and
// tests would have to reset. The plugin writes it in OnBeforeRender; Emit reads it.
// An application that calls Emit without registering the plugin has nothing to read
// and emits, which is right: there is no configuration saying otherwise.
const disabledKey = Name + ":disabled"

func disabledFor(rc *collage.RenderContext) bool {
	value, ok := rc.Get(disabledKey)
	if !ok {
		return false
	}
	disabled, _ := value.(bool) // any: SharedData's own value type
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

// Config is the plugin's configuration.
type Config struct {
	// SiteName and SiteURL describe the site. When SiteName is set, a WebSite node
	// is declared on every page — as a default, which a page emitting its own
	// WebSite replaces.
	SiteName string `json:"siteName"`
	SiteURL  string `json:"siteURL"`
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
	cfg Config
	log *slog.Logger
}

// New returns a plugin configured entirely from the application.
func New() *Plugin { return &Plugin{} }

// NewWith returns a plugin with cfg as its starting point, which the application's
// own configuration is then decoded over.
func NewWith(cfg Config) *Plugin { return &Plugin{cfg: cfg} }

func (p *Plugin) Name() string    { return Name }
func (p *Plugin) Version() string { return "0.2.3" }

func (p *Plugin) Init(_ context.Context, host collage.Host) error {
	p.log = host.Logger()
	return host.Config(&p.cfg)
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
	if p.cfg.Disabled {
		// Marked before any fragment runs, so every Emit this render sees it.
		ev.Context.Set(disabledKey, true)
		return nil
	}
	if p.cfg.SiteName == "" {
		return nil
	}
	Emit(ev.Context, WebSite{Name: p.cfg.SiteName, URL: p.cfg.SiteURL, SearchURL: p.cfg.SearchURL})
	return nil
}
