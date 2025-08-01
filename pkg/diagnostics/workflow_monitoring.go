/*
Copyright 2023 The Dapr Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package diagnostics

import (
	"context"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"

	diagUtils "github.com/dapr/dapr/pkg/diagnostics/utils"
)

var (
	workflowNameKey = tag.MustNewKey("workflow_name")
	activityNameKey = tag.MustNewKey("activity_name")
)

const (
	StatusSuccess     = "success"
	StatusFailed      = "failed"
	StatusRecoverable = "recoverable"
	CreateWorkflow    = "create_workflow"
	GetWorkflow       = "get_workflow"
	AddEvent          = "add_event"
	PurgeWorkflow     = "purge_workflow"

	WorkflowEvent = "event"
	Timer         = "timer"
)

type workflowMetrics struct {
	// workflowOperationCount records count of Successful/Failed requests to Create/Get/Purge Workflow and Add Events.
	workflowOperationCount *stats.Int64Measure
	// workflowOperationLatency records latency of response for workflow operation requests.
	workflowOperationLatency *stats.Float64Measure
	// workflowExecutionCount records count of Successful/Failed/Recoverable workflow executions.
	workflowExecutionCount *stats.Int64Measure
	// activityOperationCount records count of Successful/Failed requests to create activities.
	activityOperationCount *stats.Int64Measure
	// activityOperationLatency records latency of response for activity operation requests.
	activityOperationLatency *stats.Float64Measure
	// activityExecutionCount records count of Successful/Failed/Recoverable activity executions.
	activityExecutionCount *stats.Int64Measure
	// activityExecutionLatency records time taken to run an activity to completion.
	activityExecutionLatency *stats.Float64Measure
	// workflowExecutionLatency records time taken to run a workflow to completion.
	workflowExecutionLatency *stats.Float64Measure
	// workflowSchedulingLatency records time taken between workflow execution request and actual workflow execution
	workflowSchedulingLatency *stats.Float64Measure
	appID                     string
	enabled                   bool
	namespace                 string
	meter                     stats.Recorder
}

func newWorkflowMetrics() *workflowMetrics {
	return &workflowMetrics{
		workflowOperationCount: stats.Int64(
			"runtime/workflow/operation/count",
			"The number of successful/failed workflow operation requests.",
			stats.UnitDimensionless),
		workflowOperationLatency: stats.Float64(
			"runtime/workflow/operation/latency",
			"The latencies of responses for workflow operation requests.",
			stats.UnitMilliseconds),
		activityOperationCount: stats.Int64(
			"runtime/workflow/activity/operation/count",
			"The number of successful/failed activity operation requests.",
			stats.UnitDimensionless),
		activityOperationLatency: stats.Float64(
			"runtime/workflow/activity/operation/latency",
			"The latencies of responses for activity operation requests.",
			stats.UnitMilliseconds),
		workflowExecutionCount: stats.Int64(
			"runtime/workflow/execution/count",
			"The number of successful/failed/recoverable workflow executions.",
			stats.UnitDimensionless),
		activityExecutionCount: stats.Int64(
			"runtime/workflow/activity/execution/count",
			"The number of successful/failed/recoverable activity executions.",
			stats.UnitDimensionless),
		activityExecutionLatency: stats.Float64(
			"runtime/workflow/activity/execution/latency",
			"The total time taken to run an activity to completion.",
			stats.UnitMilliseconds),
		workflowExecutionLatency: stats.Float64(
			"runtime/workflow/execution/latency",
			"The total time taken to run workflow to completion.",
			stats.UnitMilliseconds),
		workflowSchedulingLatency: stats.Float64(
			"runtime/workflow/scheduling/latency",
			"Interval between workflow execution request and workflow execution.",
			stats.UnitMilliseconds),
	}
}

func (w *workflowMetrics) IsEnabled() bool {
	return w != nil && w.enabled
}

// Init registers the workflow metrics views.
func (w *workflowMetrics) Init(meter view.Meter, appID, namespace string, latencyDistribution *view.Aggregation) error {
	w.appID = appID
	w.enabled = true
	w.namespace = namespace
	w.meter = meter

	return meter.Register(
		diagUtils.NewMeasureView(w.workflowOperationCount, []tag.Key{appIDKey, namespaceKey, operationKey, statusKey}, view.Count()),
		diagUtils.NewMeasureView(w.workflowOperationLatency, []tag.Key{appIDKey, namespaceKey, operationKey, statusKey}, latencyDistribution),
		diagUtils.NewMeasureView(w.workflowExecutionCount, []tag.Key{appIDKey, namespaceKey, workflowNameKey, statusKey}, view.Count()),
		diagUtils.NewMeasureView(w.activityOperationCount, []tag.Key{appIDKey, namespaceKey, activityNameKey, statusKey}, view.Count()),
		diagUtils.NewMeasureView(w.activityOperationLatency, []tag.Key{appIDKey, namespaceKey, activityNameKey, statusKey}, latencyDistribution),
		diagUtils.NewMeasureView(w.activityExecutionCount, []tag.Key{appIDKey, namespaceKey, activityNameKey, statusKey}, view.Count()),
		diagUtils.NewMeasureView(w.activityExecutionLatency, []tag.Key{appIDKey, namespaceKey, activityNameKey, statusKey}, latencyDistribution),
		diagUtils.NewMeasureView(w.workflowExecutionLatency, []tag.Key{appIDKey, namespaceKey, workflowNameKey, statusKey}, latencyDistribution),
		diagUtils.NewMeasureView(w.workflowSchedulingLatency, []tag.Key{appIDKey, namespaceKey, workflowNameKey}, latencyDistribution))
}

// WorkflowOperationEvent records total number of Successful/Failed workflow Operations requests. It also records latency for those requests.
func (w *workflowMetrics) WorkflowOperationEvent(ctx context.Context, operation, status string, elapsed float64) {
	if !w.IsEnabled() {
		return
	}

	stats.RecordWithOptions(ctx, stats.WithRecorder(w.meter), stats.WithTags(diagUtils.WithTags(w.workflowOperationCount.Name(), appIDKey, w.appID, namespaceKey, w.namespace, operationKey, operation, statusKey, status)...), stats.WithMeasurements(w.workflowOperationCount.M(1)))

	if elapsed > 0 {
		stats.RecordWithOptions(ctx, stats.WithRecorder(w.meter), stats.WithTags(diagUtils.WithTags(w.workflowOperationLatency.Name(), appIDKey, w.appID, namespaceKey, w.namespace, operationKey, operation, statusKey, status)...), stats.WithMeasurements(w.workflowOperationLatency.M(elapsed)))
	}
}

// WorkflowExecutionEvent records total number of Successful/Failed/Recoverable workflow executions.
// Execution latency for workflow is not supported yet.
func (w *workflowMetrics) WorkflowExecutionEvent(ctx context.Context, workflowName, status string) {
	if !w.IsEnabled() {
		return
	}

	stats.RecordWithOptions(ctx, stats.WithRecorder(w.meter), stats.WithTags(diagUtils.WithTags(w.workflowExecutionCount.Name(), appIDKey, w.appID, namespaceKey, w.namespace, workflowNameKey, workflowName, statusKey, status)...), stats.WithMeasurements(w.workflowExecutionCount.M(1)))
}

func (w *workflowMetrics) WorkflowExecutionLatency(ctx context.Context, workflowName, status string, elapsed float64) {
	if !w.IsEnabled() {
		return
	}

	if elapsed > 0 {
		stats.RecordWithOptions(ctx, stats.WithRecorder(w.meter), stats.WithTags(diagUtils.WithTags(w.workflowExecutionLatency.Name(), appIDKey, w.appID, namespaceKey, w.namespace, workflowNameKey, workflowName, statusKey, status)...), stats.WithMeasurements(w.workflowExecutionLatency.M(elapsed)))
	}
}

func (w *workflowMetrics) WorkflowSchedulingLatency(ctx context.Context, workflowName string, elapsed float64) {
	if !w.IsEnabled() {
		return
	}

	if elapsed > 0 {
		stats.RecordWithOptions(ctx, stats.WithRecorder(w.meter), stats.WithTags(diagUtils.WithTags(w.workflowSchedulingLatency.Name(), appIDKey, w.appID, namespaceKey, w.namespace, workflowNameKey, workflowName)...), stats.WithMeasurements(w.workflowSchedulingLatency.M(elapsed)))
	}
}

// ActivityExecutionEvent records total number of Successful/Failed/Recoverable actvity executions. It also records latency for these executions.
func (w *workflowMetrics) ActivityExecutionEvent(ctx context.Context, activityName, status string, elapsed float64) {
	if !w.IsEnabled() {
		return
	}

	stats.RecordWithOptions(ctx, stats.WithRecorder(w.meter), stats.WithTags(diagUtils.WithTags(w.activityExecutionCount.Name(), appIDKey, w.appID, namespaceKey, w.namespace, activityNameKey, activityName, statusKey, status)...), stats.WithMeasurements(w.activityExecutionCount.M(1)))

	if elapsed > 0 {
		stats.RecordWithOptions(ctx, stats.WithRecorder(w.meter), stats.WithTags(diagUtils.WithTags(w.activityExecutionLatency.Name(), appIDKey, w.appID, namespaceKey, w.namespace, activityNameKey, activityName, statusKey, status)...), stats.WithMeasurements(w.activityExecutionLatency.M(elapsed)))
	}
}

// ActivityOperationEvent records total number of Successful/Failed/Recoverable activity requests. It also records latency for these requests.
func (w *workflowMetrics) ActivityOperationEvent(ctx context.Context, activityName, status string, elapsed float64) {
	if !w.IsEnabled() {
		return
	}

	stats.RecordWithOptions(ctx, stats.WithRecorder(w.meter), stats.WithTags(diagUtils.WithTags(w.activityOperationCount.Name(), appIDKey, w.appID, namespaceKey, w.namespace, activityNameKey, activityName, statusKey, status)...), stats.WithMeasurements(w.activityOperationCount.M(1)))

	if elapsed > 0 {
		stats.RecordWithOptions(ctx, stats.WithRecorder(w.meter), stats.WithTags(diagUtils.WithTags(w.activityOperationLatency.Name(), appIDKey, w.appID, namespaceKey, w.namespace, activityNameKey, activityName, statusKey, status)...), stats.WithMeasurements(w.activityOperationLatency.M(elapsed)))
	}
}
