package jsonld_test

import (
	"strings"
	"testing"
	"time"

	"github.com/Elagoht/collage/pkg/collage"

	jsonld "github.com/Elagoht/collage-jsonld"
)

var mediaTR = time.FixedZone("Europe/Istanbul", 3*60*60)

func TestVideoObject_Marshal(t *testing.T) {
	cases := []struct {
		name  string
		video jsonld.VideoObject
		want  string
	}{
		{
			name: "everything",
			video: jsonld.VideoObject{
				ID:             "https://ose.example/v/jam#video",
				Name:           "Jam özeti",
				Description:    "48 saat, 3 dakikada.",
				ThumbnailURLs:  []string{"https://ose.example/t1.jpg", "https://ose.example/t2.jpg"},
				UploadDate:     time.Date(2026, 11, 9, 12, 0, 0, 0, mediaTR),
				Duration:       3*time.Minute + 20*time.Second,
				ContentURL:     "https://ose.example/jam.mp4",
				EmbedURL:       "https://ose.example/embed/jam",
				Expires:        time.Date(2027, 11, 9, 0, 0, 0, 0, mediaTR),
				WatchCount:     1200,
				RegionsAllowed: []string{"TR", "DE"},
				Clips:          []jsonld.Clip{{Name: "Açılış", StartOffset: 0, EndOffset: 30, URL: "https://ose.example/v/jam?t=0"}},
				Live: &jsonld.BroadcastEvent{
					StartDate: time.Date(2026, 11, 7, 19, 0, 0, 0, mediaTR),
					EndDate:   time.Date(2026, 11, 7, 22, 0, 0, 0, mediaTR),
				},
			},
			want: `{"@context":"https://schema.org","@id":"https://ose.example/v/jam#video","@type":"VideoObject",` +
				`"name":"Jam özeti","description":"48 saat, 3 dakikada.",` +
				`"thumbnailUrl":["https://ose.example/t1.jpg","https://ose.example/t2.jpg"],` +
				`"uploadDate":"2026-11-09T12:00:00+03:00","duration":"PT3M20S",` +
				`"contentUrl":"https://ose.example/jam.mp4","embedUrl":"https://ose.example/embed/jam",` +
				`"expires":"2027-11-09T00:00:00+03:00","regionsAllowed":"TR,DE",` +
				`"interactionStatistic":{"@type":"InteractionCounter","interactionType":{"@type":"WatchAction"},"userInteractionCount":1200},` +
				`"hasPart":[{"@type":"Clip","name":"Açılış","startOffset":0,"endOffset":30,"url":"https://ose.example/v/jam?t=0"}],` +
				`"publication":{"@type":"BroadcastEvent","isLiveBroadcast":true,` +
				`"startDate":"2026-11-07T19:00:00+03:00","endDate":"2026-11-07T22:00:00+03:00"}}`,
		},
		{
			name:  "a name alone",
			video: jsonld.VideoObject{Name: "Jam"},
			want:  `{"@context":"https://schema.org","@type":"VideoObject","name":"Jam"}`,
		},
		{
			name:  "hours",
			video: jsonld.VideoObject{Name: "Jam", Duration: 2*time.Hour + 5*time.Minute},
			want:  `{"@context":"https://schema.org","@type":"VideoObject","name":"Jam","duration":"PT2H5M"}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := marshal(t, tc.video); got != tc.want {
				t.Errorf("got  %s\nwant %s", got, tc.want)
			}
		})
	}
}

func TestImageObject_Marshal(t *testing.T) {
	cases := []struct {
		name  string
		image jsonld.ImageObject
		want  string
	}{
		{
			name: "everything",
			image: jsonld.ImageObject{
				ID:                 "https://ose.example/i/1#image",
				ContentURL:         "https://ose.example/i/1.jpg",
				Name:               "Sahne",
				Caption:            "Açılış konuşması",
				License:            "https://creativecommons.org/licenses/by/4.0/",
				AcquireLicensePage: "https://ose.example/lisans",
				CreditText:         "OSE",
				Creator:            jsonld.Agent{Name: "Senem Aksakal"},
				CopyrightNotice:    "© OSE",
			},
			want: `{"@context":"https://schema.org","@id":"https://ose.example/i/1#image","@type":"ImageObject",` +
				`"contentUrl":"https://ose.example/i/1.jpg","name":"Sahne","caption":"Açılış konuşması",` +
				`"license":"https://creativecommons.org/licenses/by/4.0/","acquireLicensePage":"https://ose.example/lisans",` +
				`"creditText":"OSE","creator":{"@type":"Person","name":"Senem Aksakal"},"copyrightNotice":"© OSE"}`,
		},
		{
			name:  "a URL alone",
			image: jsonld.ImageObject{ContentURL: "https://ose.example/i/1.jpg"},
			want:  `{"@context":"https://schema.org","@type":"ImageObject","contentUrl":"https://ose.example/i/1.jpg"}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := marshal(t, tc.image); got != tc.want {
				t.Errorf("got  %s\nwant %s", got, tc.want)
			}
		})
	}
}

func TestItemList_Marshal(t *testing.T) {
	cases := []struct {
		name string
		list jsonld.ItemList
		want string
	}{
		{
			name: "summary: URLs to each item's own page",
			list: jsonld.ItemList{Items: []jsonld.ListEntry{{URL: "https://ose.example/r/1"}, {URL: "https://ose.example/r/2"}}},
			want: `{"@context":"https://schema.org","@type":"ItemList","itemListElement":[` +
				`{"@type":"ListItem","position":1,"url":"https://ose.example/r/1"},` +
				`{"@type":"ListItem","position":2,"url":"https://ose.example/r/2"}]}`,
		},
		{
			name: "all in one: each item in full, without its own @context",
			list: jsonld.ItemList{ID: "https://ose.example/r#list", Name: "Tarifler", Items: []jsonld.ListEntry{
				{Node: jsonld.Recipe{Name: "Menemen"}},
				{URL: "https://ose.example/r/2", Node: jsonld.Recipe{Name: "Pilav"}},
			}},
			want: `{"@context":"https://schema.org","@id":"https://ose.example/r#list","@type":"ItemList","name":"Tarifler","itemListElement":[` +
				`{"@type":"ListItem","position":1,"item":{"@type":"Recipe","name":"Menemen"}},` +
				`{"@type":"ListItem","position":2,"url":"https://ose.example/r/2","item":{"@type":"Recipe","name":"Pilav"}}]}`,
		},
		{
			name: "empty",
			list: jsonld.ItemList{},
			want: `{"@context":"https://schema.org","@type":"ItemList","itemListElement":[]}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := marshal(t, tc.list); got != tc.want {
				t.Errorf("got  %s\nwant %s", got, tc.want)
			}
		})
	}
}

func TestRecipe_Marshal(t *testing.T) {
	cases := []struct {
		name   string
		recipe jsonld.Recipe
		want   string
	}{
		{
			name: "everything",
			recipe: jsonld.Recipe{
				ID:            "https://ose.example/r/menemen#recipe",
				Name:          "Menemen",
				ImageURLs:     []string{"https://ose.example/m1.jpg", "https://ose.example/m2.jpg"},
				Description:   "Soğanlı mı soğansız mı.",
				Author:        jsonld.Agent{Name: "Göktuğ Saylam"},
				DatePublished: time.Date(2026, 10, 1, 9, 0, 0, 0, mediaTR),
				PrepTime:      10 * time.Minute,
				CookTime:      15 * time.Minute,
				TotalTime:     25 * time.Minute,
				Yield:         "2 kişilik",
				Category:      "Kahvaltı",
				Cuisine:       "Türk",
				Keywords:      []string{"yumurta", "domates"},
				Calories:      "270 kalori",
				Ingredients:   []string{"3 yumurta", "2 domates"},
				Instructions: []jsonld.HowToStep{
					{Name: "Doğra", Text: "Domatesleri doğrayın.", URL: "https://ose.example/r/menemen#1", ImageURL: "https://ose.example/s1.jpg"},
					{Text: "Yumurtaları kırın."},
				},
				AggregateRating: &jsonld.AggregateRating{Value: 4.7, RatingCount: 31},
				Video:           &jsonld.VideoObject{Name: "Menemen", UploadDate: time.Date(2026, 10, 1, 9, 0, 0, 0, mediaTR)},
			},
			want: `{"@context":"https://schema.org","@id":"https://ose.example/r/menemen#recipe","@type":"Recipe",` +
				`"name":"Menemen","image":["https://ose.example/m1.jpg","https://ose.example/m2.jpg"],` +
				`"description":"Soğanlı mı soğansız mı.","author":{"@type":"Person","name":"Göktuğ Saylam"},` +
				`"datePublished":"2026-10-01T09:00:00+03:00","prepTime":"PT10M","cookTime":"PT15M","totalTime":"PT25M",` +
				`"recipeYield":"2 kişilik","recipeCategory":"Kahvaltı","recipeCuisine":"Türk","keywords":"yumurta, domates",` +
				`"nutrition":{"@type":"NutritionInformation","calories":"270 kalori"},` +
				`"recipeIngredient":["3 yumurta","2 domates"],` +
				`"recipeInstructions":[{"@type":"HowToStep","name":"Doğra","text":"Domatesleri doğrayın.",` +
				`"url":"https://ose.example/r/menemen#1","image":"https://ose.example/s1.jpg"},` +
				`{"@type":"HowToStep","text":"Yumurtaları kırın."}],` +
				`"aggregateRating":{"@type":"AggregateRating","ratingValue":4.7,"ratingCount":31},` +
				`"video":{"@type":"VideoObject","name":"Menemen","uploadDate":"2026-10-01T09:00:00+03:00"}}`,
		},
		{
			name:   "a name alone",
			recipe: jsonld.Recipe{Name: "Menemen"},
			want:   `{"@context":"https://schema.org","@type":"Recipe","name":"Menemen"}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := marshal(t, tc.recipe); got != tc.want {
				t.Errorf("got  %s\nwant %s", got, tc.want)
			}
		})
	}
}

func TestHowTo_Marshal(t *testing.T) {
	cases := []struct {
		name  string
		howTo jsonld.HowTo
		want  string
	}{
		{
			name: "everything",
			howTo: jsonld.HowTo{
				ID:          "https://ose.example/h#howto",
				Name:        "Stant kurmak",
				Description: "Fuar standı kurulumu.",
				ImageURL:    "https://ose.example/stant.jpg",
				TotalTime:   90 * time.Minute,
				Supplies:    []string{"Afiş"},
				Tools:       []string{"Matkap"},
				Steps:       []jsonld.HowToStep{{Text: "Afişi asın."}},
			},
			want: `{"@context":"https://schema.org","@id":"https://ose.example/h#howto","@type":"HowTo",` +
				`"name":"Stant kurmak","description":"Fuar standı kurulumu.","image":"https://ose.example/stant.jpg",` +
				`"totalTime":"PT1H30M","supply":[{"@type":"HowToSupply","name":"Afiş"}],` +
				`"tool":[{"@type":"HowToTool","name":"Matkap"}],"step":[{"@type":"HowToStep","text":"Afişi asın."}]}`,
		},
		{
			name:  "a name alone",
			howTo: jsonld.HowTo{Name: "Stant kurmak"},
			want:  `{"@context":"https://schema.org","@type":"HowTo","name":"Stant kurmak"}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := marshal(t, tc.howTo); got != tc.want {
				t.Errorf("got  %s\nwant %s", got, tc.want)
			}
		})
	}
}

func TestPlugin_MediaTypesTakeKeys(t *testing.T) {
	site := newSite(t, jsonld.New(jsonld.Options{}), func(rc *collage.RenderContext) {
		for _, name := range []string{"a", "b"} {
			jsonld.Emit(rc,
				jsonld.VideoObject{Key: name, Name: name},
				jsonld.ImageObject{Key: name, ContentURL: name},
				jsonld.ItemList{Key: name},
				jsonld.Recipe{Key: name, Name: name},
				jsonld.HowTo{Key: name, Name: name},
			)
		}
	})

	count := map[string]int{}
	for _, node := range blocks(t, get(t, site, "/article")) {
		schemaType, _ := node["@type"].(string)
		count[schemaType]++
	}
	for _, want := range []string{"VideoObject", "ImageObject", "ItemList", "Recipe", "HowTo"} {
		if count[want] != 2 {
			t.Errorf("%d %s blocks, want 2 (%v)", count[want], want, count)
		}
	}
}

func TestPlugin_ClosingScriptTagInMediaCannotEscape(t *testing.T) {
	const payload = `</script><img src=x onerror=alert(1)>`
	site := newSite(t, jsonld.New(jsonld.Options{}), func(rc *collage.RenderContext) {
		jsonld.Emit(rc,
			jsonld.VideoObject{Name: payload, Clips: []jsonld.Clip{{Name: payload}}},
			jsonld.ImageObject{ContentURL: payload, Caption: payload},
			jsonld.ItemList{Items: []jsonld.ListEntry{{Node: jsonld.Recipe{Name: payload}}}},
			jsonld.Recipe{Name: payload, Ingredients: []string{payload}, Video: &jsonld.VideoObject{Name: payload}},
			jsonld.HowTo{Name: payload, Steps: []jsonld.HowToStep{{Text: payload}}},
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
