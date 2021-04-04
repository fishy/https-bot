package similarity

import (
	"math"
)

// Similarity returns how similar a and b are.
//
// The return value is the total length of the chunks a and b have in common.
// For example, when a is "abcdef" and b is "abcfoodef",
// they have 2 chunks in common: "abc" and "def", thus 6 is returned.
func Similarity(a, b []byte) int {
	common, indexA, indexB := LCS(a, b)
	if common == 0 || common == len(a) || common == len(b) {
		return common
	}
	total := common
	total += Similarity(a[0:indexA], b[0:indexB])
	total += Similarity(a[common+indexA:], b[common+indexB:])
	return total
}

func similarity(a, b []byte) (float64, float64) {
	if len(a) == 0 && len(b) == 0 {
		return 1, 1
	}
	if len(a) == 0 || len(b) == 0 {
		return 0, 0
	}
	sim := float64(Similarity(a, b))
	return sim / float64(len(a)), sim / float64(len(b))
}

// MinSimilarity returns the smaller number between Similarity(a, b) / len(a) and
// Similarity(a, b) / len(b).
//
// 1 means they are identical, 0 means they have nothing in common.
func MinSimilarity(a, b []byte) float64 {
	return math.Min(similarity(a, b))
}

// MaxSimilarity returns the larger number between Similarity(a, b) / len(a) and
// Similarity(a, b) / len(b).
//
// 1 means either they are identical, or one is superset of the other.
// (for example, a = "abcdef" and b = "abcfoodef")
func MaxSimilarity(a, b []byte) float64 {
	return math.Max(similarity(a, b))
}

// LCS is an implementation of longest common subsequence problem[1] optimized
// for space.
//
// Most LCS implementations are optimized for time,
// do a lot of allocations to memoize,
// making them unsuitable for larger inputs.
//
// Since in our use case we only need the length of the common chunk,
// we can avoid most of the allocations.
//
// In worst case scenario (there's almost nothing in common between a and b)
// the time complexity is O(N^2).
// In best case scenario (a == b) the time complexity is O(N).
//
// [1]: https://en.wikipedia.org/wiki/Longest_common_subsequence_problem
func LCS(a, b []byte) (max, indexA, indexB int) {
	for i := 0; i < len(a)-max; i++ {
		for j := 0; j < len(b)-max; j++ {
			if a[i] == b[j] {
				k := 1
				for i+k < len(a) && j+k < len(b) && a[i+k] == b[j+k] {
					k++
				}
				if k > max {
					max = k
					indexA = i
					indexB = j
				}
			}
		}
	}
	return
}
