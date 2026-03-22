package metrics

import "time"

// percentileFromSorted returns the pth percentile from an ascending-sorted
// latency slice using a discrete rank rule: k = ceil(p*n/100), then sorted[k-1].
func percentileFromSorted(sorted []time.Duration, p int) time.Duration {
	n := len(sorted)
	if n == 0 {
		return 0
	}
	if p <= 0 {
		return sorted[0]
	}
	if p >= 100 {
		return sorted[n-1]
	}
	k := (p*n + 99) / 100 // ceil(p*n/100)
	if k < 1 {
		k = 1
	}
	if k > n {
		k = n
	}
	return sorted[k-1]
}
