package jsonld

import (
	"encoding/json"
	"time"
)

// Course is a course of study: schema.org/Course. Search engines list a site's
// courses from these; each needs a name, a description and a provider.
type Course struct {
	// ID and Key are as on Article.
	ID          string
	Key         string
	Name        string
	Description string
	URL         string
	// Provider is who offers the course; an empty Kind is an Organization here,
	// not a Person, since that is what offers most courses.
	Provider        Agent
	Offers          []Offer
	AggregateRating *AggregateRating
	Instances       []CourseInstance
}

// CourseInstance is one run of a Course: schema.org/CourseInstance.
type CourseInstance struct {
	// Mode is how it is attended: "Online", "Onsite" or "Blended".
	Mode string
	// Workload is the time it takes, as a duration: 20 hours is PT20H.
	Workload time.Duration
	// Location is where an onsite run happens, as written.
	Location string
}

func (Course) Type() string { return "Course" }

func (c Course) nodeKey() string { return c.Key }

func (c Course) MarshalJSON() ([]byte, error) {
	provider := c.Provider
	if provider.Kind == "" {
		provider.Kind = AgentOrganization
	}
	out := courseJSON{
		Context:         schemaContext,
		ID:              c.ID,
		SchemaType:      "Course",
		Name:            c.Name,
		Description:     c.Description,
		URL:             c.URL,
		Provider:        agentPtr(provider),
		Offers:          offers(c.Offers),
		AggregateRating: aggregate(c.AggregateRating),
	}
	for _, in := range c.Instances {
		out.HasCourseInstance = append(out.HasCourseInstance, courseInstanceJSON{
			SchemaType:     "CourseInstance",
			CourseMode:     in.Mode,
			CourseWorkload: isoDuration(in.Workload),
			Location:       in.Location,
		})
	}
	return json.Marshal(out)
}

type courseJSON struct {
	Context           string               `json:"@context"`
	ID                string               `json:"@id,omitempty"`
	SchemaType        string               `json:"@type"`
	Name              string               `json:"name"`
	Description       string               `json:"description,omitempty"`
	URL               string               `json:"url,omitempty"`
	Provider          *agentJSON           `json:"provider,omitempty"`
	Offers            []offerJSON          `json:"offers,omitempty"`
	AggregateRating   *aggregateRatingJSON `json:"aggregateRating,omitempty"`
	HasCourseInstance []courseInstanceJSON `json:"hasCourseInstance,omitempty"`
}

type courseInstanceJSON struct {
	SchemaType     string `json:"@type"`
	CourseMode     string `json:"courseMode,omitempty"`
	CourseWorkload string `json:"courseWorkload,omitempty"`
	Location       string `json:"location,omitempty"`
}

// Quiz is a set of flashcards, which is what an education Q&A result is built
// from: schema.org/Quiz.
type Quiz struct {
	// ID and Key are as on Article.
	ID   string
	Key  string
	Name string
	// About is the subject: "Photosynthesis".
	About string
	// EducationalAlignment places the quiz in a curriculum: its level, its
	// subject.
	EducationalAlignment []Alignment
	Questions            []Flashcard
}

// Alignment is where a quiz sits in a curriculum: schema.org/AlignmentObject.
type Alignment struct {
	Kind AlignmentKind
	// Target is the level or subject, as written: "Lise", "Biyoloji".
	Target string
}

// AlignmentKind is what an Alignment places the quiz by.
type AlignmentKind string

const (
	AlignLevel   AlignmentKind = "educationalLevel"
	AlignSubject AlignmentKind = "educationalSubject"
)

type alignmentJSON struct {
	SchemaType    string `json:"@type"`
	AlignmentType string `json:"alignmentType"`
	TargetName    string `json:"targetName"`
}

func learningAlignments(list []Alignment) []alignmentJSON {
	var out []alignmentJSON
	for _, a := range list {
		out = append(out, alignmentJSON{SchemaType: "AlignmentObject", AlignmentType: string(a.Kind), TargetName: a.Target})
	}
	return out
}

// Flashcard is one question of a Quiz and its answer, both as plain text.
type Flashcard struct {
	Question string
	Answer   string
}

func (Quiz) Type() string { return "Quiz" }

func (q Quiz) nodeKey() string { return q.Key }

func (q Quiz) MarshalJSON() ([]byte, error) {
	out := quizJSON{
		Context:              schemaContext,
		ID:                   q.ID,
		SchemaType:           "Quiz",
		Name:                 q.Name,
		EducationalAlignment: learningAlignments(q.EducationalAlignment),
	}
	if q.About != "" {
		out.About = &learningThing{SchemaType: "Thing", Name: q.About}
	}
	for _, card := range q.Questions {
		out.HasPart = append(out.HasPart, flashcardJSON{
			SchemaType:      "Question",
			EduQuestionType: "Flashcard",
			Text:            card.Question,
			AcceptedAnswer:  answerJSON{SchemaType: "Answer", Text: card.Answer},
		})
	}
	return json.Marshal(out)
}

type quizJSON struct {
	Context              string          `json:"@context"`
	ID                   string          `json:"@id,omitempty"`
	SchemaType           string          `json:"@type"`
	Name                 string          `json:"name,omitempty"`
	About                *learningThing  `json:"about,omitempty"`
	EducationalAlignment []alignmentJSON `json:"educationalAlignment,omitempty"`
	HasPart              []flashcardJSON `json:"hasPart,omitempty"`
}

// learningThing is a schema.org/Thing by name alone.
type learningThing struct {
	SchemaType string `json:"@type"`
	Name       string `json:"name"`
}

type flashcardJSON struct {
	SchemaType      string     `json:"@type"`
	EduQuestionType string     `json:"eduQuestionType"`
	Text            string     `json:"text"`
	AcceptedAnswer  answerJSON `json:"acceptedAnswer"`
}

// MathSolver is a tool that solves math problems: schema.org/MathSolver, written
// as both a MathSolver and a LearningResource, as search engines ask.
type MathSolver struct {
	// ID and Key are as on Article.
	ID   string
	Key  string
	Name string
	URL  string
	// UsageInfo is a URL to the solver's terms of use.
	UsageInfo string
	// InLanguage is a BCP 47 code: "tr".
	InLanguage string
	// SolveURL is the solver with "{math_expression_string}" where the problem
	// goes.
	SolveURL string
	// QuestionTypes are the kinds of problem it solves: "Polynomial Equation",
	// "Derivative".
	QuestionTypes []string
}

func (MathSolver) Type() string { return "MathSolver" }

func (m MathSolver) nodeKey() string { return m.Key }

func (m MathSolver) MarshalJSON() ([]byte, error) {
	out := mathSolverJSON{
		Context:              schemaContext,
		ID:                   m.ID,
		SchemaType:           []string{"MathSolver", "LearningResource"},
		Name:                 m.Name,
		URL:                  m.URL,
		UsageInfo:            m.UsageInfo,
		InLanguage:           m.InLanguage,
		LearningResourceType: "Math solver",
	}
	if m.SolveURL != "" {
		out.PotentialAction = &solveMathActionJSON{
			SchemaType:          "SolveMathAction",
			Target:              m.SolveURL,
			MathExpressionInput: "required name=math_expression_string",
			EduQuestionType:     m.QuestionTypes,
		}
	}
	return json.Marshal(out)
}

type mathSolverJSON struct {
	Context              string               `json:"@context"`
	ID                   string               `json:"@id,omitempty"`
	SchemaType           []string             `json:"@type"`
	Name                 string               `json:"name"`
	URL                  string               `json:"url,omitempty"`
	UsageInfo            string               `json:"usageInfo,omitempty"`
	InLanguage           string               `json:"inLanguage,omitempty"`
	PotentialAction      *solveMathActionJSON `json:"potentialAction,omitempty"`
	LearningResourceType string               `json:"learningResourceType"`
}

type solveMathActionJSON struct {
	SchemaType          string   `json:"@type"`
	Target              string   `json:"target"`
	MathExpressionInput string   `json:"mathExpression-input"`
	EduQuestionType     []string `json:"eduQuestionType,omitempty"`
}

// Dataset is a body of data published for others to use: schema.org/Dataset.
// Search engines want a description between 50 and 5000 characters.
type Dataset struct {
	// ID and Key are as on Article.
	ID          string
	Key         string
	Name        string
	Description string
	URL         string
	SameAs      []string
	// Identifier are DOIs and other identifiers for the dataset.
	Identifier []string
	Keywords   []string
	// License is a URL to the license the data is published under.
	License string
	// Paywalled marks a dataset that is not free to access, as on Article; left
	// false, nothing is said either way.
	Paywalled    bool
	Creator      []Agent
	Distribution []DataDownload
	// TemporalCoverage is the period the data covers, in ISO 8601:
	// "2020-01-01/2025-12-31".
	TemporalCoverage string
	// SpatialCoverage is the place the data covers, as written.
	SpatialCoverage  string
	VariableMeasured []string
}

// DataDownload is one file a Dataset can be downloaded as:
// schema.org/DataDownload.
type DataDownload struct {
	// EncodingFormat is the file's media type: "text/csv".
	EncodingFormat string
	ContentURL     string
}

func (Dataset) Type() string { return "Dataset" }

func (d Dataset) nodeKey() string { return d.Key }

func (d Dataset) MarshalJSON() ([]byte, error) {
	out := datasetJSON{
		Context:          schemaContext,
		ID:               d.ID,
		SchemaType:       "Dataset",
		Name:             d.Name,
		Description:      d.Description,
		URL:              d.URL,
		SameAs:           d.SameAs,
		Identifier:       d.Identifier,
		Keywords:         d.Keywords,
		License:          d.License,
		Creator:          agents(d.Creator),
		TemporalCoverage: d.TemporalCoverage,
		SpatialCoverage:  d.SpatialCoverage,
		VariableMeasured: d.VariableMeasured,
	}
	if d.Paywalled {
		free := false
		out.IsAccessibleForFree = &free
	}
	for _, dl := range d.Distribution {
		out.Distribution = append(out.Distribution, dataDownloadJSON{
			SchemaType:     "DataDownload",
			EncodingFormat: dl.EncodingFormat,
			ContentURL:     dl.ContentURL,
		})
	}
	return json.Marshal(out)
}

type datasetJSON struct {
	Context             string             `json:"@context"`
	ID                  string             `json:"@id,omitempty"`
	SchemaType          string             `json:"@type"`
	Name                string             `json:"name"`
	Description         string             `json:"description,omitempty"`
	URL                 string             `json:"url,omitempty"`
	SameAs              []string           `json:"sameAs,omitempty"`
	Identifier          []string           `json:"identifier,omitempty"`
	Keywords            []string           `json:"keywords,omitempty"`
	License             string             `json:"license,omitempty"`
	IsAccessibleForFree *bool              `json:"isAccessibleForFree,omitempty"`
	Creator             []agentJSON        `json:"creator,omitempty"`
	Distribution        []dataDownloadJSON `json:"distribution,omitempty"`
	TemporalCoverage    string             `json:"temporalCoverage,omitempty"`
	SpatialCoverage     string             `json:"spatialCoverage,omitempty"`
	VariableMeasured    []string           `json:"variableMeasured,omitempty"`
}

type dataDownloadJSON struct {
	SchemaType     string `json:"@type"`
	EncodingFormat string `json:"encodingFormat,omitempty"`
	ContentURL     string `json:"contentUrl"`
}

var (
	_ keyed = Course{}
	_ keyed = Quiz{}
	_ keyed = MathSolver{}
	_ keyed = Dataset{}
)
