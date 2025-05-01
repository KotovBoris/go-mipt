//go:build !solution

package genericsum

import (
	"math/cmplx"
	"sort"
	"sync"

	"golang.org/x/exp/constraints"
)

func Min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}

	return b
}

func SortSlice[T constraints.Ordered](a []T) {
	sort.Slice(a, func(i, j int) bool {
		return a[i] < a[j]
	})
}

func MapsEqual[K comparable, V comparable](a, b map[K]V) bool {
	if len(a) != len(b) {
		return false
	}

	for key := range a {
		if a[key] != b[key] {
			return false
		}
	}

	return true
}

func SliceContains[T comparable](s []T, v T) bool {
	for _, elem := range s {
		if elem == v {
			return true
		}
	}

	return false
}

func MergeChans[T any](chs ...<-chan T) <-chan T {
	res := make(chan T)
	closed := make([]bool, len(chs))

	go func() {
		var wg sync.WaitGroup
		done_cnt := 0

		for done_cnt < len(chs) {
			for i, ch := range chs {
				if closed[i] {
					continue
				}

				select {
				case value, ok := <-ch:
					if !ok {
						closed[i] = true
						done_cnt++
						continue
					}

					wg.Add(1)
					go func() {
						res <- value
						wg.Done()
					}()
				default:
				}
			}
		}

		wg.Wait()
		close(res)
	}()

	return res
}

func SensefullIsHermitianMatrix(m [][]complex128) bool {
	for i := range m {
		for j := range m[i] {
			if j >= len(m) || i >= len(m[j]) || m[i][j] != cmplx.Conj(m[j][i]) {
				return false
			}
		}
	}

	return true
}

func UnSensefullIsHermitianMatrix(m [][]complex64) bool {
	for i := range m {
		for j := range m[i] {
			if j >= len(m) || i >= len(m[j]) || complex128(m[i][j]) != cmplx.Conj(complex128(m[j][i])) {
				return false
			}
		}
	}

	return true
}

type Number interface {
	int | int8 | int16 | int32 | int64 |
		uint | uint8 | uint16 | uint32 | uint64 | uintptr |
		float32 | float64 |
		complex64 | complex128
}

func IsSymmetricalMatrix[T comparable](m [][]T) bool {
	for i := range m {
		for j := range m[i] {
			if j >= len(m) || i >= len(m[j]) || m[i][j] != m[j][i] {
				return false
			}
		}
	}

	return true
}

func IsHermitianMatrix[T Number](m [][]T) bool {
	switch v := any(m).(type) {
	case [][]complex128:
		return SensefullIsHermitianMatrix(v)
	case [][]complex64:
		return UnSensefullIsHermitianMatrix(v)
	default:
		return IsSymmetricalMatrix(m)
	}
}
