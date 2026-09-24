package sonar

import (
	"sonarbridge-go/internal/core/domain/sonar"
	"time"
)

type QualityGateStatus string

const (
	QG_OK    QualityGateStatus = "OK"
	QG_ERROR QualityGateStatus = "ERROR"
	QG_WARN  QualityGateStatus = "WARN"
	QG_NONE  QualityGateStatus = "NONE"
)

type PagingResponse struct {
	PageIndex int `json:"pageIndex"`
	PageSize  int `json:"pageSize"`
	Total     int `json:"total"`
}

type TaskResponse struct {
	Task struct {
		AnalysisID string `json:"analysisId"`
		Status     string `json:"status"`
	} `json:"task"`
}

type TasksResponse struct {
	Tasks []struct {
		AnalysisID string `json:"analysisId,omitempty"`
		Status     string `json:"status"`
	} `json:"tasks"`
}

type TasksCompleteResponse struct {
	Tasks []struct {
		ID                 string    `json:"id"`
		Type               string    `json:"type,omitempty"`
		ComponentID        string    `json:"componentId,omitempty"`
		ComponentKey       string    `json:"componentKey,omitempty"`
		ComponentName      string    `json:"componentName,omitempty"`
		ComponentQualifier string    `json:"componentQualifier,omitempty"`
		AnalysisID         string    `json:"analysisId,omitempty"`
		Status             string    `json:"status"`
		SubmittedAt        time.Time `json:"submittedAt"`
		SubmitterLogin     string    `json:"submitterLogin,omitempty"`
		StartedAt          time.Time `json:"startedAt"`
		ExecutedAt         time.Time `json:"executedAt"`
		ExecutionTimeMs    int       `json:"executionTimeMs,omitempty"`
		HasScannerContext  bool      `json:"hasScannerContext"`
		WarningCount       int       `json:"warningCount"`
		Warnings           []string  `json:"warnings,omitempty"`
		InfoMessages       []string  `json:"infoMessages,omitempty"`
	} `json:"tasks"`
	Paging PagingResponse `json:"paging"`
}

type QualityGateCondition struct {
	Status         QualityGateStatus `json:"status"`
	MetricKey      string            `json:"metricKey"`
	Comparator     string            `json:"comparator"`
	PeriodIndex    *string           `json:"periodIndex,omitempty"`
	ErrorThreshold string            `json:"errorThreshold"`
	ActualValue    string            `json:"actualValue"`
}

type QualityGateStatusResponse struct {
	ProjectStatus struct {
		Status            QualityGateStatus      `json:"status"`
		Conditions        []QualityGateCondition `json:"conditions"`
		IgnoredConditions bool                   `json:"ignoredConditions"`
		CaycStatus        string                 `json:"caycStatus,omitempty"`
		Period            *struct {
			Index     int    `json:"index"`
			Mode      string `json:"mode"`
			Date      string `json:"date"`
			Parameter string `json:"parameter"`
		} `json:"period,omitempty"`
	} `json:"projectStatus"`
}

type MeasureComponentResponse struct {
	Key       string `json:"key"`
	Name      string `json:"name"`
	Qualifier string `json:"qualifier"`
	Language  string `json:"Language,omitempty"`
	Path      string `json:"path,omitempty"`
	Measures  []struct {
		MetricName string `json:"Metric"`
		Value      string `json:"value,omitempty"`
		Period     *struct {
			Value     string `json:"value"`
			BestValue bool   `json:"bestValue"`
		} `json:"period,omitempty"`
	} `json:"measures"`
}
type MeasureMetricItemResponse struct {
	Key                   string `json:"key"`
	Name                  string `json:"name"`
	Description           string `json:"description"`
	Domain                string `json:"domain,omitempty"`
	MetricType            string `json:"type,omitempty"`
	HigherValuesAreBetter bool   `json:"higherValuesAreBetter,omitempty"`
	Qualitative           bool   `json:"qualitative,omitempty"`
	Hidden                bool   `json:"hidden,omitempty"`
}

// /go:generate stringer -type=MeasuresResponse
type MeasuresResponse struct {
	Component MeasureComponentResponse    `json:"component"`
	Metrics   []MeasureMetricItemResponse `json:"metrics"`
	Period    *struct {
		Mode      string `json:"mode"`
		Date      string `json:"date"`
		Parameter string `json:"parameter"`
	} `json:"period"`
}

type TextRange struct {
	StartLine   int `json:"startLine"`
	EndLine     int `json:"endLine"`
	StartOffset int `json:"startOffset"`
	EndOffset   int `json:"endOffset"`
}

type IssuesItemResponse struct {
	Key                string              `json:"key"`
	Rule               string              `json:"rule"`
	Severity           sonar.IssueSeverity `json:"severity"`
	Component          string              `json:"component"`
	Project            string              `json:"project"`
	Line               *int                `json:"line,omitempty"`
	Hash               *string             `json:"hash,omitempty"`
	TextRange          *TextRange          `json:"textRange,omitempty"`
	Flows              []any               `json:"flows"`
	Status             string              `json:"issueStatus"`
	LinkedTicketStatus string              `json:"linkedTicketStatus"`
	Message            string              `json:"message"`
	MessageFormattings []struct {
		Start int    `json:"start"`
		End   int    `json:"end"`
		Type  string `json:"type"`
	} `json:"messageFormattings,omitempty"`
	Effort            string          `json:"effort"`
	Debt              string          `json:"debt"`
	Assignee          *string         `json:"assignee,omitempty"`
	Author            *string         `json:"author,omitempty"`
	Tags              []string        `json:"tags"`
	CreationDate      string          `json:"creationDate"`
	UpdateDate        string          `json:"updateDate"`
	Type              sonar.IssueType `json:"type"`
	Scope             string          `json:"scope"`
	QuickFixAvailable bool            `json:"quickFixAvailable"`
	IsNew             bool            `json:"isNew"`
	PrioritizedRule   bool            `json:"prioritizedRule"`

	CleanCodeAttribute         string `json:"cleanCodeAttribute,omitempty"`
	CleanCodeAttributeCategory string `json:"cleanCodeAttributeCategory,omitempty"`
	Impacts                    []struct {
		SoftwareQuality string `json:"softwareQuality"`
		Severity        string `json:"severity"`
	} `json:"impacts,omitempty"`
	Comments []struct {
		Key       string `json:"key"`
		Login     string `json:"login"`
		HtmlText  string `json:"htmlText"`
		Markdown  string `json:"markdown"`
		Updatable string `json:"updatable"`
		CreatedAt string `json:"createdAt"`
	} `json:"comments,omitempty"`
	Transitions  []string `json:"transitions,omitempty"`
	Actions      []string `json:"actions,omitempty"`
	InternalTags []string `json:"internalTags,omitempty"`
}

type IssueSearchComponent struct {
	Key       string `json:"key"`
	Enabled   bool   `json:"enabled"`
	Qualifier string `json:"qualifier"`
	Name      string `json:"name"`
	LongName  string `json:"longName"`
	Path      string `json:"path,omitempty"`
}

type Rule struct {
	Key          string `json:"key"`
	Name         string `json:"name"`
	Status       string `json:"status"`
	Language     string `json:"lang"`
	LanguageName string `json:"langName"`
}

type IssueUser struct {
	Login  string `json:"login"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
	Avatar string `json:"avatar"`
}

// IssueFacet Dans le jargon SonarQube, une facet est un agrégat de comptage sur une dimension donnée.
// Concrètement, quand tu fais un /api/issues/search, tu peux demander facets=severities,types,rules et SonarQube te retourne en plus des issues un bloc
// Utilité pour tes rapports — au lieu d'appeler /api/measures/component + paginer toutes les issues, tu peux faire un seul appel avec
// ps=1&facets=severities,types pour avoir les totaux sans charger les issues elles-mêmes.
type IssueFacet struct {
	Property string `json:"property"`
	Values   []struct {
		Val   string `json:"val"`
		Count int    `json:"count"`
	} `json:"values"`
}

type IssuesResponse struct {
	Total       int                    `json:"total"`
	P           int                    `json:"p"`
	Ps          int                    `json:"ps"`
	EffortTotal float32                `json:"effortTotal"`
	Issues      []IssuesItemResponse   `json:"issues"`
	Components  []IssueSearchComponent `json:"components"`
	Facets      []IssueFacet           `json:"facets,omitempty"`
}

type IssueSearchResponse struct {
	Paging     PagingResponse         `json:"paging"`
	Issues     []IssuesItemResponse   `json:"issues"`
	Components []IssueSearchComponent `json:"components,omitempty"`
	Rules      []Rule                 `json:"rules,omitempty"`
	Facets     []IssueFacet           `json:"facets,omitempty"`
}

type MetricValue interface {
	~string | ~int | ~int64 | ~float32 | ~float64
}
type Metrics[T MetricValue] map[string]T
