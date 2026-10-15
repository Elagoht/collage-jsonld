package jsonld

import (
	"encoding/json"
	"strings"
	"time"
)

// VideoObject is a video: schema.org/VideoObject.
//
// Search engines want a name, a thumbnail and an upload date before they show
// one; ContentURL or EmbedURL is how they reach the video itself.
type VideoObject struct {
	// ID and Key are as on Article.
	ID            string
	Key           string
	Name          string
	Description   string
	ThumbnailURLs []string
	UploadDate    time.Time
	Duration      time.Duration
	// ContentURL is the video file; EmbedURL is a player for it.
	ContentURL string
	EmbedURL   string
	// Expires is when the video stops being available, left out at zero.
	Expires time.Time
	// WatchCount is how many times it was watched, left out at zero.
	WatchCount int
	// RegionsAllowed are ISO 3166-1 alpha-2 codes.
	RegionsAllowed []string
	// Clips are the video's chapters, each a start and end in seconds.
	Clips []Clip
	// Live marks the video as a live stream and when it runs.
	Live *BroadcastEvent
}

// Clip is a chapter of a VideoObject: schema.org/Clip. Offsets are seconds from
// the video's start.
type Clip struct {
	Name        string
	StartOffset int
	EndOffset   int
	// URL points into the video at the clip's start: "…/watch?t=90".
	URL string
}

// BroadcastEvent is a live stream's schedule: schema.org/BroadcastEvent.
type BroadcastEvent struct {
	StartDate time.Time
	EndDate   time.Time
}

func (VideoObject) Type() string { return "VideoObject" }

func (v VideoObject) nodeKey() string { return v.Key }

func (v VideoObject) MarshalJSON() ([]byte, error) {
	out := v.wire()
	out.Context = schemaContext
	return json.Marshal(out)
}

// wire is the video's shape without "@context", which a video nested in a
// Recipe leaves out.
func (v VideoObject) wire() videoJSON {
	out := videoJSON{
		ID:             v.ID,
		SchemaType:     "VideoObject",
		Name:           v.Name,
		Description:    v.Description,
		ThumbnailURL:   v.ThumbnailURLs,
		UploadDate:     dateTime(v.UploadDate),
		Duration:       isoDuration(v.Duration),
		ContentURL:     v.ContentURL,
		EmbedURL:       v.EmbedURL,
		Expires:        dateTime(v.Expires),
		RegionsAllowed: v.RegionsAllowed,
	}
	if v.WatchCount > 0 {
		out.InteractionStatistic = &interactionCounterJSON{
			SchemaType:           "InteractionCounter",
			InteractionType:      mediaAction{SchemaType: "WatchAction"},
			UserInteractionCount: v.WatchCount,
		}
	}
	for _, c := range v.Clips {
		out.HasPart = append(out.HasPart, clipJSON{
			SchemaType:  "Clip",
			Name:        c.Name,
			StartOffset: c.StartOffset,
			EndOffset:   c.EndOffset,
			URL:         c.URL,
		})
	}
	if v.Live != nil {
		out.Publication = &broadcastJSON{
			SchemaType:      "BroadcastEvent",
			IsLiveBroadcast: true,
			StartDate:       dateTime(v.Live.StartDate),
			EndDate:         dateTime(v.Live.EndDate),
		}
	}
	return out
}

type videoJSON struct {
	Context              string                  `json:"@context,omitempty"`
	ID                   string                  `json:"@id,omitempty"`
	SchemaType           string                  `json:"@type"`
	Name                 string                  `json:"name"`
	Description          string                  `json:"description,omitempty"`
	ThumbnailURL         []string                `json:"thumbnailUrl,omitempty"`
	UploadDate           string                  `json:"uploadDate,omitempty"`
	Duration             string                  `json:"duration,omitempty"`
	ContentURL           string                  `json:"contentUrl,omitempty"`
	EmbedURL             string                  `json:"embedUrl,omitempty"`
	Expires              string                  `json:"expires,omitempty"`
	RegionsAllowed       []string                `json:"regionsAllowed,omitempty"`
	InteractionStatistic *interactionCounterJSON `json:"interactionStatistic,omitempty"`
	HasPart              []clipJSON              `json:"hasPart,omitempty"`
	Publication          *broadcastJSON          `json:"publication,omitempty"`
}

type interactionCounterJSON struct {
	SchemaType           string      `json:"@type"`
	InteractionType      mediaAction `json:"interactionType"`
	UserInteractionCount int         `json:"userInteractionCount"`
}

// mediaAction is an action named by type alone: {"@type":"WatchAction"}.
type mediaAction struct {
	SchemaType string `json:"@type"`
}

type clipJSON struct {
	SchemaType  string `json:"@type"`
	Name        string `json:"name"`
	StartOffset int    `json:"startOffset"`
	EndOffset   int    `json:"endOffset,omitempty"`
	URL         string `json:"url,omitempty"`
}

type broadcastJSON struct {
	SchemaType      string `json:"@type"`
	IsLiveBroadcast bool   `json:"isLiveBroadcast"`
	StartDate       string `json:"startDate,omitempty"`
	EndDate         string `json:"endDate,omitempty"`
}

// ImageObject is an image described in its own right — who made it, under what
// license: schema.org/ImageObject. It is what an image search shows as the
// image's licensing details.
type ImageObject struct {
	// ID and Key are as on Article.
	ID  string
	Key string
	// ContentURL is the image file.
	ContentURL string
	Name       string
	Caption    string
	// License is a URL to the license the image is used under;
	// AcquireLicensePage is where to get one.
	License            string
	AcquireLicensePage string
	CreditText         string
	Creator            Agent
	CopyrightNotice    string
}

func (ImageObject) Type() string { return "ImageObject" }

func (i ImageObject) nodeKey() string { return i.Key }

func (i ImageObject) MarshalJSON() ([]byte, error) {
	return json.Marshal(imageMetadataJSON{
		Context:            schemaContext,
		ID:                 i.ID,
		SchemaType:         "ImageObject",
		ContentURL:         i.ContentURL,
		Name:               i.Name,
		Caption:            i.Caption,
		License:            i.License,
		AcquireLicensePage: i.AcquireLicensePage,
		CreditText:         i.CreditText,
		Creator:            agentPtr(i.Creator),
		CopyrightNotice:    i.CopyrightNotice,
	})
}

type imageMetadataJSON struct {
	Context            string     `json:"@context"`
	ID                 string     `json:"@id,omitempty"`
	SchemaType         string     `json:"@type"`
	ContentURL         string     `json:"contentUrl"`
	Name               string     `json:"name,omitempty"`
	Caption            string     `json:"caption,omitempty"`
	License            string     `json:"license,omitempty"`
	AcquireLicensePage string     `json:"acquireLicensePage,omitempty"`
	CreditText         string     `json:"creditText,omitempty"`
	Creator            *agentJSON `json:"creator,omitempty"`
	CopyrightNotice    string     `json:"copyrightNotice,omitempty"`
}

// ItemList is an ordered list of things, which is what a search engine builds a
// carousel from: schema.org/ItemList.
//
// Each entry is either a URL, for a list page whose items each have a page of
// their own carrying their own structured data, or a Node, for a page that holds
// every item in full. Positions are written for you, starting at 1.
type ItemList struct {
	// ID and Key are as on Article.
	ID    string
	Key   string
	Name  string
	Items []ListEntry
}

// ListEntry is one item of an ItemList: a URL to the item's own page, or the
// item itself. With both, both are written.
type ListEntry struct {
	URL  string
	Node Node
}

func (ItemList) Type() string { return "ItemList" }

func (l ItemList) nodeKey() string { return l.Key }

func (l ItemList) MarshalJSON() ([]byte, error) {
	items := make([]mediaListItemJSON, 0, len(l.Items))
	for i, entry := range l.Items {
		item := mediaListItemJSON{SchemaType: "ListItem", Position: i + 1, URL: entry.URL}
		if entry.Node != nil {
			encoded, err := nested(entry.Node)
			if err != nil {
				return nil, err
			}
			item.Item = encoded
		}
		items = append(items, item)
	}
	return json.Marshal(itemListJSON{
		Context:         schemaContext,
		ID:              l.ID,
		SchemaType:      "ItemList",
		Name:            l.Name,
		ItemListElement: items,
	})
}

type itemListJSON struct {
	Context         string              `json:"@context"`
	ID              string              `json:"@id,omitempty"`
	SchemaType      string              `json:"@type"`
	Name            string              `json:"name,omitempty"`
	ItemListElement []mediaListItemJSON `json:"itemListElement"`
}

type mediaListItemJSON struct {
	SchemaType string          `json:"@type"`
	Position   int             `json:"position"`
	URL        string          `json:"url,omitempty"`
	Item       json.RawMessage `json:"item,omitempty"`
}

// HowToStep is one step of a HowTo or a Recipe: schema.org/HowToStep.
type HowToStep struct {
	Name     string
	Text     string
	URL      string
	ImageURL string
}

type howToStepJSON struct {
	SchemaType string `json:"@type"`
	Name       string `json:"name,omitempty"`
	Text       string `json:"text"`
	URL        string `json:"url,omitempty"`
	Image      string `json:"image,omitempty"`
}

func mediaSteps(steps []HowToStep) []howToStepJSON {
	var out []howToStepJSON
	for _, s := range steps {
		out = append(out, howToStepJSON{SchemaType: "HowToStep", Name: s.Name, Text: s.Text, URL: s.URL, Image: s.ImageURL})
	}
	return out
}

// Recipe is a dish and how to make it: schema.org/Recipe.
type Recipe struct {
	// ID and Key are as on Article.
	ID   string
	Key  string
	Name string
	// URL is the recipe's own page, which a carousel of recipes links to.
	URL           string
	ImageURLs     []string
	Description   string
	Author        Agent
	DatePublished time.Time
	PrepTime      time.Duration
	CookTime      time.Duration
	TotalTime     time.Duration
	// Yield is how much it makes: "4 servings".
	Yield    string
	Category string
	Cuisine  string
	Keywords []string
	// Calories per serving, as written: "270 calories".
	Calories        string
	Ingredients     []string
	Instructions    []HowToStep
	AggregateRating *AggregateRating
	Reviews         []Review
	Video           *VideoObject
}

func (Recipe) Type() string { return "Recipe" }

func (r Recipe) nodeKey() string { return r.Key }

func (r Recipe) MarshalJSON() ([]byte, error) {
	out := recipeJSON{
		Context:            schemaContext,
		ID:                 r.ID,
		SchemaType:         "Recipe",
		Name:               r.Name,
		URL:                r.URL,
		Image:              r.ImageURLs,
		Description:        r.Description,
		Author:             agentPtr(r.Author),
		DatePublished:      dateTime(r.DatePublished),
		PrepTime:           isoDuration(r.PrepTime),
		CookTime:           isoDuration(r.CookTime),
		TotalTime:          isoDuration(r.TotalTime),
		RecipeYield:        r.Yield,
		RecipeCategory:     r.Category,
		RecipeCuisine:      r.Cuisine,
		Keywords:           strings.Join(r.Keywords, ", "),
		RecipeIngredient:   r.Ingredients,
		RecipeInstructions: mediaSteps(r.Instructions),
		AggregateRating:    aggregate(r.AggregateRating),
		Review:             reviews(r.Reviews),
	}
	if r.Calories != "" {
		out.Nutrition = &nutritionJSON{SchemaType: "NutritionInformation", Calories: r.Calories}
	}
	if r.Video != nil {
		video := r.Video.wire()
		out.Video = &video
	}
	return json.Marshal(out)
}

type recipeJSON struct {
	Context            string               `json:"@context"`
	ID                 string               `json:"@id,omitempty"`
	SchemaType         string               `json:"@type"`
	Name               string               `json:"name"`
	URL                string               `json:"url,omitempty"`
	Image              []string             `json:"image,omitempty"`
	Description        string               `json:"description,omitempty"`
	Author             *agentJSON           `json:"author,omitempty"`
	DatePublished      string               `json:"datePublished,omitempty"`
	PrepTime           string               `json:"prepTime,omitempty"`
	CookTime           string               `json:"cookTime,omitempty"`
	TotalTime          string               `json:"totalTime,omitempty"`
	RecipeYield        string               `json:"recipeYield,omitempty"`
	RecipeCategory     string               `json:"recipeCategory,omitempty"`
	RecipeCuisine      string               `json:"recipeCuisine,omitempty"`
	Keywords           string               `json:"keywords,omitempty"`
	Nutrition          *nutritionJSON       `json:"nutrition,omitempty"`
	RecipeIngredient   []string             `json:"recipeIngredient,omitempty"`
	RecipeInstructions []howToStepJSON      `json:"recipeInstructions,omitempty"`
	AggregateRating    *aggregateRatingJSON `json:"aggregateRating,omitempty"`
	Review             []reviewJSON         `json:"review,omitempty"`
	Video              *videoJSON           `json:"video,omitempty"`
}

type nutritionJSON struct {
	SchemaType string `json:"@type"`
	Calories   string `json:"calories"`
}

// HowTo is instructions for doing something: schema.org/HowTo.
type HowTo struct {
	// ID and Key are as on Article.
	ID          string
	Key         string
	Name        string
	Description string
	ImageURL    string
	TotalTime   time.Duration
	// Supplies are used up by the task; Tools are not.
	Supplies []string
	Tools    []string
	Steps    []HowToStep
}

func (HowTo) Type() string { return "HowTo" }

func (h HowTo) nodeKey() string { return h.Key }

func (h HowTo) MarshalJSON() ([]byte, error) {
	return json.Marshal(howToJSON{
		Context:     schemaContext,
		ID:          h.ID,
		SchemaType:  "HowTo",
		Name:        h.Name,
		Description: h.Description,
		Image:       h.ImageURL,
		TotalTime:   isoDuration(h.TotalTime),
		Supply:      mediaThings("HowToSupply", h.Supplies),
		Tool:        mediaThings("HowToTool", h.Tools),
		Step:        mediaSteps(h.Steps),
	})
}

type howToJSON struct {
	Context     string          `json:"@context"`
	ID          string          `json:"@id,omitempty"`
	SchemaType  string          `json:"@type"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Image       string          `json:"image,omitempty"`
	TotalTime   string          `json:"totalTime,omitempty"`
	Supply      []mediaThing    `json:"supply,omitempty"`
	Tool        []mediaThing    `json:"tool,omitempty"`
	Step        []howToStepJSON `json:"step,omitempty"`
}

// mediaThing is a thing by type and name: a HowToSupply, a HowToTool.
type mediaThing struct {
	SchemaType string `json:"@type"`
	Name       string `json:"name"`
}

func mediaThings(schemaType string, names []string) []mediaThing {
	var out []mediaThing
	for _, n := range names {
		out = append(out, mediaThing{SchemaType: schemaType, Name: n})
	}
	return out
}

var (
	_ keyed = VideoObject{}
	_ keyed = ImageObject{}
	_ keyed = ItemList{}
	_ keyed = Recipe{}
	_ keyed = HowTo{}
)
