/*
 *  Brown University, CS138, Spring 2020
 *
 *  Purpose: contains all interval related logic.
 */

package pkg

// Interval represents [Lower, Upper)
type Interval struct {
	Lower uint64 // Inclusive
	Upper uint64 // Exclusive
}

// GenerateIntervals divides the range [0, upperBound] into numIntervals intervals.
func GenerateIntervals(upperBound uint64, numIntervals int) (intervals []Interval) {
	// TODO: Students should implement this.

	remainder := int(upperBound) % numIntervals
	//comment out below conditional to include remainder functionality
	//var intUpperBound int
	//if remainder != 0{
	//	intUpperBound = int(upperBound)-remainder
	//} else {
	//	intUpperBound = int(upperBound)
	//}

	//intUpperBound := int(upperBound)

	stepSize := int(upperBound) / numIntervals

	for i := 0; i < int(upperBound); i += stepSize {
		lower := uint64(i)
		upper := uint64(i + stepSize)

		// uncomment below to add remainder functionality to intervals
		// without this will fail interval testing
		if remainder != 0 {
			upper = upper + 1
			remainder = remainder - 1
			i = i + 1
		}

		currentInterval := Interval{Lower: lower, Upper: upper}

		slice := append(intervals, currentInterval)
		intervals = slice

	}

	return
}
