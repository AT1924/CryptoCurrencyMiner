package test

import (
	"fmt"
	liteminer "liteminer/pkg"
	"testing"
)

func Equal(a, b []liteminer.Interval) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func TestBasicInterval(t *testing.T) {
	fmt.Println("starting interval test")
	upper := 13
	numIntervals := 4
	intervals := liteminer.GenerateIntervals(uint64(upper), numIntervals)

	var ans []liteminer.Interval
	interval1 := liteminer.Interval{
		Lower: 0,
		Upper: 4,
	}
	interval2 := liteminer.Interval{
		Lower: 4,
		Upper: 7,
	}
	interval3 := liteminer.Interval{
		Lower: 7,
		Upper: 10,
	}
	interval4 := liteminer.Interval{
		Lower: 10,
		Upper: 13,
	}
	ans = append(ans, interval1)
	ans = append(ans, interval2)
	ans = append(ans, interval3)
	ans = append(ans, interval4)
	if !Equal(ans, intervals) {
		t.Errorf("Expected  %d, but received %d", ans, intervals)
	}

}

func TestLargeRemainder(t *testing.T) {
	fmt.Println("starting interval second test")
	upper := 179
	numIntervals := 30
	intervals := liteminer.GenerateIntervals(uint64(upper), numIntervals)
	//fmt.Println("intervals", intervals)

	ans := liteminer.Interval{Lower: 174, Upper: 179}

	if ans != intervals[len(intervals)-1] {
		t.Errorf("Expected  %d, but received %d", ans, intervals[len(intervals)-1])
	}

}
