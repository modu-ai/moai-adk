package measure

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// Axis is the measurement axis of an efficiency comparison (design §4.7): exactly one per record.
type Axis string

const (
	AxisTokensCost   Axis = "tokens-cost"
	AxisTime         Axis = "time"
	AxisCallsRetries Axis = "api-retries"
	AxisFailRecovery Axis = "failure-recovery"
	AxisInformation  Axis = "information-use"
	AxisQuality      Axis = "quality-safety-correctness"
)

var axes = map[Axis]bool{AxisTokensCost: true, AxisTime: true, AxisCallsRetries: true, AxisFailRecovery: true, AxisInformation: true, AxisQuality: true}

// ValueKind keeps the four kinds of value apart; nothing converts one into another.
type ValueKind string

const (
	KindMeasured  ValueKind = "measured"
	KindEstimated ValueKind = "estimated"
	KindUnknown   ValueKind = "UNKNOWN"
	KindNotRun    ValueKind = "NOT_RUN"
)

// ThresholdState says whether the threshold and sample design are only proposed or approved (OD-1).
type ThresholdState string

const (
	ThresholdProposed ThresholdState = "proposed"
	ThresholdApproved ThresholdState = "approved"
)

// Errors a record can be refused with.
var (
	ErrFieldMissing      = errors.New("measure: an efficiency record field is missing")
	ErrEstimateAsMeasure = errors.New("measure: an estimate was filed as a measurement")
	ErrGainHidesRegress  = errors.New("measure: an efficiency gain was reported next to a quality, safety or correctness regression")
	ErrCheckpointCost    = errors.New("measure: a checkpoint cost must be a measured value or a recorded Gap")
)

var treeSHA = regexp.MustCompile(`^[0-9a-f]{40}$`)

// QualityGuard is the quality, safety and correctness result of the same comparison.
type QualityGuard struct {
	Metric    string
	Result    string
	Regressed bool
}

// Record is one efficiency comparison record with the fields of design §4.7.
type Record struct {
	Axis           Axis
	BaselineSource string // a preserved log or an approved isolated fixture; NOT_RUN or "not comparable" otherwise
	Comparability  string // how task, model/effort, permissions, revision, environment and window line up, with limits
	ValueKind      ValueKind
	Value          string
	Unit           string
	Command        string
	Tree           string // the measured tree's commit SHA
	// EstimatedFrom says what an estimate was derived from. It is required for an estimated value and refused
	// for a measured one.
	EstimatedFrom  string
	Quality        QualityGuard
	ThresholdState ThresholdState
}

func missing(fields ...string) error {
	return fmt.Errorf("%w: %s", ErrFieldMissing, strings.Join(fields, ", "))
}

// Validate refuses a record with a missing field, an estimate filed as a measurement, or a measured value
// without the evidence of its measurement.
func (r Record) Validate() error {
	var absent []string
	if !axes[r.Axis] {
		absent = append(absent, "axis")
	}
	if strings.TrimSpace(r.BaselineSource) == "" {
		absent = append(absent, "baseline_source")
	}
	if strings.TrimSpace(r.Comparability) == "" {
		absent = append(absent, "comparability")
	}
	if strings.TrimSpace(r.Quality.Metric) == "" || strings.TrimSpace(r.Quality.Result) == "" {
		absent = append(absent, "quality_guard")
	}
	if r.ThresholdState != ThresholdProposed && r.ThresholdState != ThresholdApproved {
		absent = append(absent, "threshold_state")
	}
	switch r.ValueKind {
	case KindMeasured:
		if strings.TrimSpace(r.EstimatedFrom) != "" {
			return fmt.Errorf("%w: a measured value carries an estimate basis (%q)", ErrEstimateAsMeasure, r.EstimatedFrom)
		}
		if strings.TrimSpace(r.Value) == "" {
			absent = append(absent, "value")
		}
		if strings.TrimSpace(r.Unit) == "" {
			absent = append(absent, "unit")
		}
		if strings.TrimSpace(r.Command) == "" {
			absent = append(absent, "command")
		}
		if !treeSHA.MatchString(r.Tree) {
			absent = append(absent, "tree")
		}
	case KindEstimated:
		if strings.TrimSpace(r.Value) == "" {
			absent = append(absent, "value")
		}
		if strings.TrimSpace(r.Unit) == "" {
			absent = append(absent, "unit")
		}
		if strings.TrimSpace(r.EstimatedFrom) == "" {
			absent = append(absent, "estimated_from")
		}
	case KindUnknown, KindNotRun:
		if strings.TrimSpace(r.Value) != "" {
			return fmt.Errorf("%w: a %s record carries the value %q", ErrEstimateAsMeasure, r.ValueKind, r.Value)
		}
	default:
		absent = append(absent, "value_kind")
	}
	if len(absent) > 0 {
		return missing(absent...)
	}
	return nil
}

// ReportGain refuses to report an efficiency gain for a record whose quality guard shows a regression, and
// refuses one whose value is not measured. It returns nil when no gain is claimed.
func (r Record) ReportGain(claimed bool) error {
	if !claimed {
		return nil
	}
	if err := r.Validate(); err != nil {
		return err
	}
	if r.Quality.Regressed {
		return fmt.Errorf("%w: %s %s", ErrGainHidesRegress, r.Quality.Metric, r.Quality.Result)
	}
	if r.ValueKind != KindMeasured {
		return fmt.Errorf("%w: only a measured value may back a reported gain, this one is %s", ErrEstimateAsMeasure, r.ValueKind)
	}
	return nil
}

// CheckpointCost validates the cost field of a checkpoint (API, token or time). It is either a measured value
// (a record that validates as measured) or a recorded Gap; an estimate, UNKNOWN or an empty cell is neither.
func CheckpointCost(r *Record, gap string) error {
	switch {
	case r != nil && r.ValueKind == KindMeasured:
		return r.Validate()
	case strings.TrimSpace(gap) != "" && (r == nil || r.ValueKind != KindEstimated):
		return nil
	}
	return ErrCheckpointCost
}
