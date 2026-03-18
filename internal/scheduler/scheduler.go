// Package scheduler generates randomized achievement unlock schedules.
package scheduler

import (
	"math/rand/v2"
	"slices"
	"time"

	"github.com/jeerrry/steamctl/internal/steam"
)

// ScheduledUnlock pairs an achievement with its planned unlock time.
type ScheduledUnlock struct {
	Name       string
	Percent    float64
	UnlockAt   time.Time
	DelayAfter time.Duration
}

// BuildSchedule generates a schedule of achievement unlocks sorted by global
// percentage (most common first), with randomized intervals between min and max,
// fitting within the total time range. Achievements already unlocked are skipped.
func BuildSchedule(
	global []steam.GlobalAchievement,
	unlocked map[string]bool,
	startTime time.Time,
	timeRange time.Duration,
	minInterval time.Duration,
	maxInterval time.Duration,
) []ScheduledUnlock {
	// Sort by percentage descending (most common first)
	sorted := make([]steam.GlobalAchievement, len(global))
	copy(sorted, global)
	slices.SortFunc(sorted, func(a, b steam.GlobalAchievement) int {
		if a.Percent > b.Percent {
			return -1
		}
		if a.Percent < b.Percent {
			return 1
		}
		return 0
	})

	// Filter out already unlocked
	var pending []steam.GlobalAchievement
	for _, a := range sorted {
		if !unlocked[a.Name] {
			pending = append(pending, a)
		}
	}

	if len(pending) == 0 {
		return nil
	}

	intervals := generateIntervals(len(pending), timeRange, minInterval, maxInterval)

	schedule := make([]ScheduledUnlock, len(pending))
	cursor := startTime
	for i, a := range pending {
		schedule[i] = ScheduledUnlock{
			Name:       a.Name,
			Percent:    a.Percent,
			UnlockAt:   cursor,
			DelayAfter: intervals[i],
		}
		cursor = cursor.Add(intervals[i])
	}

	return schedule
}

// generateIntervals creates randomized delays that sum to approximately the
// total time range. Each interval is clamped to [lo, hi].
func generateIntervals(n int, total, lo, hi time.Duration) []time.Duration {
	if n <= 0 {
		return nil
	}

	intervals := make([]time.Duration, n)

	for i := range intervals {
		jitter := hi - lo
		if jitter <= 0 {
			intervals[i] = lo
			continue
		}
		intervals[i] = lo + time.Duration(rand.Int64N(int64(jitter)))
	}

	// Scale intervals to fit within total time range
	var sum time.Duration
	for _, d := range intervals {
		sum += d
	}

	if sum > 0 {
		scale := float64(total) / float64(sum)
		for i := range intervals {
			intervals[i] = time.Duration(float64(intervals[i]) * scale)
			if intervals[i] < lo {
				intervals[i] = lo
			}
			if intervals[i] > hi {
				intervals[i] = hi
			}
		}
	}

	return intervals
}
