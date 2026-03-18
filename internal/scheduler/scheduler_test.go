package scheduler

import (
	"testing"
	"time"

	"github.com/jeerrry/steamctl/internal/steam"
)

func TestBuildSchedule_SortsByPercentDescending(t *testing.T) {
	t.Parallel()

	global := []steam.GlobalAchievement{
		{Name: "RARE", Percent: 5.0},
		{Name: "COMMON", Percent: 90.0},
		{Name: "MID", Percent: 50.0},
	}

	schedule := BuildSchedule(
		global,
		nil,
		time.Now(),
		10*time.Hour,
		1*time.Minute,
		30*time.Minute,
	)

	if len(schedule) != 3 {
		t.Fatalf("len(schedule) = %d, want 3", len(schedule))
	}
	if schedule[0].Name != "COMMON" {
		t.Errorf("schedule[0].Name = %q, want COMMON", schedule[0].Name)
	}
	if schedule[1].Name != "MID" {
		t.Errorf("schedule[1].Name = %q, want MID", schedule[1].Name)
	}
	if schedule[2].Name != "RARE" {
		t.Errorf("schedule[2].Name = %q, want RARE", schedule[2].Name)
	}
}

func TestBuildSchedule_SkipsUnlocked(t *testing.T) {
	t.Parallel()

	global := []steam.GlobalAchievement{
		{Name: "A", Percent: 80.0},
		{Name: "B", Percent: 60.0},
		{Name: "C", Percent: 40.0},
	}

	unlocked := map[string]bool{"A": true, "C": true}

	schedule := BuildSchedule(
		global,
		unlocked,
		time.Now(),
		5*time.Hour,
		1*time.Minute,
		30*time.Minute,
	)

	if len(schedule) != 1 {
		t.Fatalf("len(schedule) = %d, want 1", len(schedule))
	}
	if schedule[0].Name != "B" {
		t.Errorf("schedule[0].Name = %q, want B", schedule[0].Name)
	}
}

func TestBuildSchedule_AllUnlocked(t *testing.T) {
	t.Parallel()

	global := []steam.GlobalAchievement{
		{Name: "A", Percent: 80.0},
	}

	unlocked := map[string]bool{"A": true}

	schedule := BuildSchedule(
		global,
		unlocked,
		time.Now(),
		5*time.Hour,
		1*time.Minute,
		30*time.Minute,
	)

	if schedule != nil {
		t.Errorf("expected nil schedule, got %d entries", len(schedule))
	}
}

func TestBuildSchedule_IntervalsWithinBounds(t *testing.T) {
	t.Parallel()

	global := make([]steam.GlobalAchievement, 20)
	for i := range global {
		global[i] = steam.GlobalAchievement{
			Name:    "ACH_" + string(rune('A'+i)),
			Percent: float64(100 - i*5),
		}
	}

	lo := 5 * time.Minute
	hi := 1 * time.Hour

	schedule := BuildSchedule(
		global,
		nil,
		time.Now(),
		24*time.Hour,
		lo,
		hi,
	)

	if len(schedule) != 20 {
		t.Fatalf("len(schedule) = %d, want 20", len(schedule))
	}

	for i, s := range schedule {
		if s.DelayAfter < lo {
			t.Errorf("schedule[%d].DelayAfter = %v, below min %v", i, s.DelayAfter, lo)
		}
		if s.DelayAfter > hi {
			t.Errorf("schedule[%d].DelayAfter = %v, above max %v", i, s.DelayAfter, hi)
		}
	}
}

func TestBuildSchedule_ChronologicalOrder(t *testing.T) {
	t.Parallel()

	global := []steam.GlobalAchievement{
		{Name: "A", Percent: 80.0},
		{Name: "B", Percent: 60.0},
		{Name: "C", Percent: 40.0},
		{Name: "D", Percent: 20.0},
	}

	start := time.Now()
	schedule := BuildSchedule(
		global,
		nil,
		start,
		10*time.Hour,
		1*time.Minute,
		2*time.Hour,
	)

	for i := 1; i < len(schedule); i++ {
		if schedule[i].UnlockAt.Before(schedule[i-1].UnlockAt) {
			t.Errorf("schedule[%d].UnlockAt (%v) is before schedule[%d].UnlockAt (%v)",
				i, schedule[i].UnlockAt, i-1, schedule[i-1].UnlockAt)
		}
	}
}
