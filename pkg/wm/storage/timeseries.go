package storage

import (
	"sort"
	"time"

	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
)

const (
	tsRemoveBelow   = 0.25
	tsNoChangeAbove = 0.9
)

// sortTimeseries sort given timeseries in ascending order
func sortTimeseries(series []*wm.TimeseriesValue) {
	sort.Slice(series, func(i, j int) bool {
		return float64(series[i].Timestamp) < float64(series[j].Timestamp)
	})
}

// toSparkline converts the given timeseries to sparkline
func toSparkline(series []*wm.TimeseriesValue) ([]float64, error) {
	sparkline := make([]float64, 0)
	for _, s := range series {
		sparkline = append(sparkline, s.Value)
	}
	return sparkline, nil
}

// deepCloneTs creates a deep copy of the given time series, series
func deepCloneTs(series []*wm.TimeseriesValue) []*wm.TimeseriesValue {
	var newSeries []*wm.TimeseriesValue
	for _, t := range series {
		newSeries = append(newSeries, &wm.TimeseriesValue{Timestamp: t.Timestamp, Value: t.Value})
	}
	return newSeries
}

// Note: computeCoverage and correctIncompleteTimeseries is ported from incomplete-data-detection.js from casuemos repo

// computeCoverage computes the temporal coverage of the final aggregated data point using the final
// raw data point rawLastTimestamp, raw and aggregated temporal resolutions, rawRes and aggRes.
//
// The raw temporal resolution has to be finer than the aggregated resolution.
// If the raw resolution is coarser or equal to the aggregated resolution, the coverage
// is assumed to be 100%. The raw resolution 'Other' is assumed to always have 100% coverage.
//
// This function uses the following approximations and assumptions:
// - All years have 12 months
// - All years have 36 dekad
// - All years have 52 weeks
// - All years have 365 days
// - All months have 3 dekad
// - All months have 4 weeks
// - All months have 30 days
//
// returns the percentage of the final month/year that was covered by the raw data.
// Negative values indicate inconsistent raw data and aggregated data timestamps.
// Since approximations are used, the value could be greater than 1.
func computeCoverage(rawLastTimestamp int64, rawRes wm.TemporalResolution, aggRes wm.TemporalResolutionOption) float64 {
	lastRawTime := time.UnixMilli(rawLastTimestamp)
	if aggRes == wm.TemporalResolutionOptionYear {
		month := float64(lastRawTime.Month())
		dayOfYear := float64(lastRawTime.YearDay())
		switch rawRes {
		case wm.TemporalResolutionOther:
		case wm.TemporalResolutionAnnual:
			return 1
		case wm.TemporalResolutionMonthly:
			return month / 12
		case wm.TemporalResolutionDekad: // 36 dekad in a year
			return dayOfYear / 10 / 36
		case wm.TemporalResolutionWeekly: // 52 weeks in a year
			return dayOfYear / 7 / 52
		case wm.TemporalResolutionDaily: // 365 days in a year
			return dayOfYear / 365
		}
	}
	dayOfMonth := float64(lastRawTime.Day())
	switch rawRes {
	case wm.TemporalResolutionOther:
	case wm.TemporalResolutionAnnual:
	case wm.TemporalResolutionMonthly:
		return 1
	case wm.TemporalResolutionDekad: // 3 dekad in a month
		return dayOfMonth / 10 / 3
	case wm.TemporalResolutionWeekly: // 4 weeks in a month
		return dayOfMonth / 7 / 4
	case wm.TemporalResolutionDaily: // 30 days in a month
		return dayOfMonth / 30
	}
	return 1
}

// correctIncompleteTimeseries checks the final point in the timeseries and adjusts it based on information about the raw data.
//
// If the raw data covers less than tsRemoveBelow of the last timeframe the last point is removed.
// If the raw data covers more than tsNoChangeAbove of the last timeframe there is no change.
// If the coverage falls between these values and the aggregation is set to 'Sum' the data is scaled
// by the reciprocal of the coverage.
//
// timeseries - List of points
// rawRes - Resolution of the raw data
// aggRes - Resolution of the aggregated data
// aggOpt - Temporal aggregation function
// rawLastTimestamp - Final date from the raw data
func correctIncompleteTimeseries(timeseries []*wm.TimeseriesValue, aggOpt wm.AggregationOption, aggRes wm.TemporalResolutionOption, rawRes wm.TemporalResolution, rawLastTimestamp int64) []*wm.TimeseriesValue {
	series := deepCloneTs(timeseries)

	if aggOpt != wm.AggregationOptionSum {
		// No change is required
		return series
	}
	if len(series) == 0 || series[0].Timestamp > rawLastTimestamp {
		// Data is out of scope
		return series
	}

	lastRawTime := time.UnixMilli(rawLastTimestamp).UTC()
	lastAggTime := time.UnixMilli(series[len(series)-1].Timestamp).UTC()

	isSameYear := lastAggTime.Year() == lastRawTime.Year()
	isSameYearMonth := isSameYear && lastAggTime.Month() == lastRawTime.Month()

	areDatesValid := (aggRes == wm.TemporalResolutionOptionYear && isSameYear) ||
		(aggRes == wm.TemporalResolutionOptionMonth && isSameYearMonth)

	if !areDatesValid {
		return series
	}

	// Dates are valid and the aggregation is sum.
	// Run extrapolation if the temporal coverage of the final aggregated point is between tsRemoveBelow and tsNoChangeAbove
	coverage := computeCoverage(rawLastTimestamp, rawRes, aggRes)
	if coverage < tsRemoveBelow {
		return series[:len(series)-1]
	}
	if coverage < tsNoChangeAbove {
		series[len(series)-1].Value *= 1 / coverage
		return series
	}
	return series
}
