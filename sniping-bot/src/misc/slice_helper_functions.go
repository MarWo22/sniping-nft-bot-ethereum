package misc

import (
	"math/rand"
	"time"
)

// Returns true if the slice contains element
func ContainsInt(slice []int, element int) bool {
	for _, elem := range slice {
		if elem == element {
			return true
		}
	}
	return false
}

// Returns true if the slice contains element
func ContainsString(slice []string, element string) bool {
	for _, elem := range slice {
		if elem == element {
			return true
		}
	}
	return false
}

// Returns true if the slice contains the given element, else returns false
func BinarySearchInt(needle int, haystack []int) int {

	low := 0
	high := len(haystack) - 1

	for low <= high {
		median := (low + high) / 2

		if haystack[median] < needle {
			low = median + 1
		} else {
			high = median - 1
		}
	}

	if low == len(haystack) || haystack[low] != needle {
		return -1
	}

	return low
}

// Removes the given index from an int slice
func RemoveIntOrdered(slice []int, index int) []int {
	return append(slice[:index], slice[index+1:]...)
}

func RemoveIntUnordered(s []int, i int) []int {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

// Returns the index of an int in a int slice
func GetIndexInt(slice []int, element int) int {
	for i, elem := range slice {
		if elem == element {
			return i
		}
	}
	return -1
}

// Shuffles a slice of ints
func ShuffleInt(slice []int) []int {
	sliceCopy := make([]int, len(slice))
	copy(sliceCopy, slice)

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(sliceCopy), func(i, j int) { sliceCopy[i], sliceCopy[j] = sliceCopy[j], sliceCopy[i] })
	return sliceCopy
}
