package domain

import "sonarbridge-go/internal/core/domain/sonar"

/**
export interface SonarQubeTask {
    id: string;
    type: string;
    componentId: string;
    componentKey: string;
    componentName: string;
    analysisId: string;
    status: 'SUCCESS' | 'FAILED' | 'PENDING' | 'IN_PROGRESS' | 'CANCELED';
    submittedAt: string;
    startedAt: string;
    executedAt: string;
    executionTimeMs: number;
}
*/

type TaskStatus string

type SonarTaskDetails struct {
	Id              string
	Type            string
	ComponentId     string
	ComponentKey    string
	ComponentName   string
	AnalysisId      string
	Status          TaskStatus
	SubmittedAt     string
	ExecutedAt      string
	ExecutionTimeMs int32
}

/**
export interface AnalysisDetails {
    analysisId: string;
    dashboardUrl: string;
    metrics: Record<string, string | number>;
    qualityGate: {
        status: string;
        conditions: QualityGateCondition[];
    };
    totalIssues: number;
    newIssues: number;
    taskId: string;
}
*/

type QualityGateCondition struct {
	Status         string
	MetricKey      string
	Comparator     string
	PeriodIndex    *string
	ErrorThreshold string
	ActualValue    string
}

type AnalysisDetails struct {
	AnalysisId   string
	DashboardUrl string
	Metrics      map[string]any
	QualityGate  struct {
		Status     string
		Conditions []QualityGateCondition
	}
	TotalIssues int
	NewIssues   int
	TaskId      string
	TaskStatus  TaskStatus
	Issues      []sonar.Issue
}

const (
	TASK_SUCCESS     TaskStatus = "SUCCESS"
	TASK_FAILED      TaskStatus = "FAILED"
	TASK_PENDING     TaskStatus = "PENDING"
	TASK_IN_PROGRESS TaskStatus = "IN_PROGRESS"
	TASK_CANCELED    TaskStatus = "CANCELED"
	TASK_UNKNOWN     TaskStatus = "UNKNOWN"
)
