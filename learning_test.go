package jsonld_test

import (
	"strings"
	"testing"
	"time"

	"github.com/Elagoht/collage/pkg/collage"

	jsonld "github.com/Elagoht/collage-jsonld"
)

func TestCourse_Marshal(t *testing.T) {
	cases := []struct {
		name   string
		course jsonld.Course
		want   string
	}{
		{
			name: "everything",
			course: jsonld.Course{
				ID:              "https://ose.example/c/unity#course",
				Name:            "Unity'ye giriş",
				Description:     "Sıfırdan ilk oyun.",
				URL:             "https://ose.example/c/unity",
				Provider:        jsonld.Agent{Name: "OSE", URL: "https://ose.example"},
				Offers:          []jsonld.Offer{{Price: "0", Currency: "TRY", Category: "Free"}},
				AggregateRating: &jsonld.AggregateRating{Value: 4.5, ReviewCount: 12},
				Instances:       []jsonld.CourseInstance{{Mode: "Online", Workload: 20 * time.Hour}, {Mode: "Onsite", Location: "Kadıköy"}},
			},
			want: `{"@context":"https://schema.org","@id":"https://ose.example/c/unity#course","@type":"Course",` +
				`"name":"Unity'ye giriş","description":"Sıfırdan ilk oyun.","url":"https://ose.example/c/unity",` +
				`"provider":{"@type":"Organization","name":"OSE","url":"https://ose.example"},` +
				`"offers":[{"@type":"Offer","price":"0","priceCurrency":"TRY","category":"Free"}],` +
				`"aggregateRating":{"@type":"AggregateRating","ratingValue":4.5,"reviewCount":12},` +
				`"hasCourseInstance":[{"@type":"CourseInstance","courseMode":"Online","courseWorkload":"PT20H"},` +
				`{"@type":"CourseInstance","courseMode":"Onsite","location":"Kadıköy"}]}`,
		},
		{
			name:   "a person as provider",
			course: jsonld.Course{Name: "Pixel art", Provider: jsonld.Agent{Kind: jsonld.AgentPerson, Name: "Göktuğ Saylam"}},
			want: `{"@context":"https://schema.org","@type":"Course","name":"Pixel art",` +
				`"provider":{"@type":"Person","name":"Göktuğ Saylam"}}`,
		},
		{
			name:   "a name alone",
			course: jsonld.Course{Name: "Pixel art"},
			want:   `{"@context":"https://schema.org","@type":"Course","name":"Pixel art"}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := marshal(t, tc.course); got != tc.want {
				t.Errorf("got  %s\nwant %s", got, tc.want)
			}
		})
	}
}

func TestQuiz_Marshal(t *testing.T) {
	cases := []struct {
		name string
		quiz jsonld.Quiz
		want string
	}{
		{
			name: "everything",
			quiz: jsonld.Quiz{
				ID:               "https://ose.example/q#quiz",
				Name:             "Oyun tarihi",
				About:            "Video oyunları",
				EducationalLevel: "Lise",
				Questions:        []jsonld.Flashcard{{Question: "Pong hangi yıl çıktı?", Answer: "1972"}},
			},
			want: `{"@context":"https://schema.org","@id":"https://ose.example/q#quiz","@type":"Quiz",` +
				`"name":"Oyun tarihi","about":{"@type":"Thing","name":"Video oyunları"},"educationalLevel":"Lise",` +
				`"hasPart":[{"@type":"Question","eduQuestionType":"Flashcard","text":"Pong hangi yıl çıktı?",` +
				`"acceptedAnswer":{"@type":"Answer","text":"1972"}}]}`,
		},
		{
			name: "empty",
			quiz: jsonld.Quiz{},
			want: `{"@context":"https://schema.org","@type":"Quiz"}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := marshal(t, tc.quiz); got != tc.want {
				t.Errorf("got  %s\nwant %s", got, tc.want)
			}
		})
	}
}

func TestMathSolver_Marshal(t *testing.T) {
	cases := []struct {
		name   string
		solver jsonld.MathSolver
		want   string
	}{
		{
			name: "everything",
			solver: jsonld.MathSolver{
				ID:            "https://math.example/#solver",
				Name:          "Denklem çözücü",
				URL:           "https://math.example",
				UsageInfo:     "https://math.example/kullanim",
				InLanguage:    "tr",
				SolveURL:      "https://math.example/coz?q={math_expression_string}",
				QuestionTypes: []string{"Polynomial Equation", "Derivative"},
			},
			want: `{"@context":"https://schema.org","@id":"https://math.example/#solver","@type":["MathSolver","LearningResource"],` +
				`"name":"Denklem çözücü","url":"https://math.example","usageInfo":"https://math.example/kullanim","inLanguage":"tr",` +
				`"potentialAction":{"@type":"SolveMathAction","target":"https://math.example/coz?q={math_expression_string}",` +
				`"mathExpression-input":"required name=math_expression_string","eduQuestionType":["Polynomial Equation","Derivative"]},` +
				`"learningResourceType":"Math solver"}`,
		},
		{
			name:   "a name alone",
			solver: jsonld.MathSolver{Name: "Çözücü"},
			want:   `{"@context":"https://schema.org","@type":["MathSolver","LearningResource"],"name":"Çözücü","learningResourceType":"Math solver"}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := marshal(t, tc.solver); got != tc.want {
				t.Errorf("got  %s\nwant %s", got, tc.want)
			}
		})
	}
}

func TestDataset_Marshal(t *testing.T) {
	cases := []struct {
		name    string
		dataset jsonld.Dataset
		want    string
	}{
		{
			name: "everything",
			dataset: jsonld.Dataset{
				ID:               "https://ose.example/d/etkinlik#dataset",
				Name:             "Türkiye oyun etkinlikleri",
				Description:      "2020'den bu yana Türkiye'de düzenlenen oyun etkinliklerinin listesi.",
				URL:              "https://ose.example/d/etkinlik",
				SameAs:           []string{"https://data.example/ose"},
				Identifier:       []string{"https://doi.org/10.0000/ose"},
				Keywords:         []string{"oyun", "etkinlik"},
				License:          "https://creativecommons.org/licenses/by/4.0/",
				NotFree:          true,
				Creator:          []jsonld.Agent{{Kind: jsonld.AgentOrganization, Name: "OSE"}, {Name: ""}},
				Distribution:     []jsonld.DataDownload{{EncodingFormat: "text/csv", ContentURL: "https://ose.example/d/etkinlik.csv"}},
				TemporalCoverage: "2020-01-01/2026-12-31",
				SpatialCoverage:  "Türkiye",
				VariableMeasured: []string{"katılımcı sayısı"},
			},
			want: `{"@context":"https://schema.org","@id":"https://ose.example/d/etkinlik#dataset","@type":"Dataset",` +
				`"name":"Türkiye oyun etkinlikleri","description":"2020'den bu yana Türkiye'de düzenlenen oyun etkinliklerinin listesi.",` +
				`"url":"https://ose.example/d/etkinlik","sameAs":["https://data.example/ose"],` +
				`"identifier":["https://doi.org/10.0000/ose"],"keywords":["oyun","etkinlik"],` +
				`"license":"https://creativecommons.org/licenses/by/4.0/","isAccessibleForFree":false,` +
				`"creator":[{"@type":"Organization","name":"OSE"}],` +
				`"distribution":[{"@type":"DataDownload","encodingFormat":"text/csv","contentUrl":"https://ose.example/d/etkinlik.csv"}],` +
				`"temporalCoverage":"2020-01-01/2026-12-31","spatialCoverage":"Türkiye","variableMeasured":["katılımcı sayısı"]}`,
		},
		{
			name:    "a name alone says nothing about access",
			dataset: jsonld.Dataset{Name: "Etkinlikler"},
			want:    `{"@context":"https://schema.org","@type":"Dataset","name":"Etkinlikler"}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := marshal(t, tc.dataset); got != tc.want {
				t.Errorf("got  %s\nwant %s", got, tc.want)
			}
		})
	}
}

func TestPlugin_LearningTypesTakeKeys(t *testing.T) {
	site := newSite(t, jsonld.New(jsonld.Options{}), func(rc *collage.RenderContext) {
		for _, name := range []string{"a", "b"} {
			jsonld.Emit(rc,
				jsonld.Course{Key: name, Name: name},
				jsonld.Quiz{Key: name},
				jsonld.MathSolver{Key: name, Name: name},
				jsonld.Dataset{Key: name, Name: name},
			)
		}
	})

	count := map[string]int{}
	for _, node := range blocks(t, get(t, site, "/article")) {
		switch schemaType := node["@type"].(type) {
		case string:
			count[schemaType]++
		case []any: // any: what encoding/json decodes an array into
			count["MathSolver"]++
		}
	}
	for _, want := range []string{"Course", "Quiz", "MathSolver", "Dataset"} {
		if count[want] != 2 {
			t.Errorf("%d %s blocks, want 2 (%v)", count[want], want, count)
		}
	}
}

func TestPlugin_ClosingScriptTagInLearningCannotEscape(t *testing.T) {
	const payload = `</script><img src=x onerror=alert(1)>`
	site := newSite(t, jsonld.New(jsonld.Options{}), func(rc *collage.RenderContext) {
		jsonld.Emit(rc,
			jsonld.Course{Name: payload, Instances: []jsonld.CourseInstance{{Location: payload}}},
			jsonld.Quiz{About: payload, Questions: []jsonld.Flashcard{{Question: payload, Answer: payload}}},
			jsonld.MathSolver{Name: payload, SolveURL: payload},
			jsonld.Dataset{Name: payload, Description: payload},
		)
	})

	body := get(t, site, "/article")
	if strings.Contains(body, "<img src=x") {
		t.Fatalf("the payload escaped the script block:\n%s", body)
	}
	if nodes := blocks(t, body); len(nodes) != 4 {
		t.Fatalf("got %d nodes, want 4", len(nodes))
	}
}
