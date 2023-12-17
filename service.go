package main

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type RequestPayload struct {
	ToSort [][]int `json:"to_sort"`
}

type ResponsePayload struct {
	SortedArrays [][]int `json:"sorted_arrays"`
	TimeNS       int64   `json:"time_ns"`
}

// Sequential code-------------------------
const minMerge = 32

func Timsort1(arr []int) {
	n := len(arr)

	// Perform insertion sort for small chunks
	for i := 0; i < n; i += minMerge {
		end := i + minInt(minMerge, n-i)
		insertionSort(arr[i:end])
	}

	// Merge the sorted chunks
	for size := minMerge; size < n; size *= 2 {
		for left := 0; left < n; left += 2 * size {
			mid := left + size
			right := minInt(left+2*size, n)
			merge1(arr, left, mid, right)
		}
	}
}

func insertionSort(arr []int) {
	for i := 1; i < len(arr); i++ {
		j := i
		for j > 0 && arr[j] < arr[j-1] {
			arr[j], arr[j-1] = arr[j-1], arr[j]
			j--
		}
	}
}

func merge1(arr []int, left, mid, right int) {
	lenLeft := mid - left
	lenRight := right - mid

	leftArr := make([]int, lenLeft)
	rightArr := make([]int, lenRight)

	copy(leftArr, arr[left:mid])
	copy(rightArr, arr[mid:right])

	i, j, k := 0, 0, left

	for i < lenLeft && j < lenRight {
		if leftArr[i] <= rightArr[j] {
			arr[k] = leftArr[i]
			i++
		} else {
			arr[k] = rightArr[j]
			j++
		}
		k++
	}

	for i < lenLeft {
		arr[k] = leftArr[i]
		i++
		k++
	}

	for j < lenRight {
		arr[k] = rightArr[j]
		j++
		k++
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func sortSeq(input [][]int) [][]int {
	for i := range input {
		Timsort1(input[i])
	}
	return input
}

// concurrent sorting -----------------------------
func merge2(left, right []int) []int {
	result := make([]int, len(left)+len(right))
	i, j, k := 0, 0, 0

	for i < len(left) && j < len(right) {
		if left[i] <= right[j] {
			result[k] = left[i]
			i++
		} else {
			result[k] = right[j]
			j++
		}
		k++
	}

	for i < len(left) {
		result[k] = left[i]
		i++
		k++
	}

	for j < len(right) {
		result[k] = right[j]
		j++
		k++
	}

	return result
}

func mergeSort(arr []int, resultChan chan []int) {
	if len(arr) <= 1 {
		resultChan <- arr
		return
	}

	mid := len(arr) / 2

	leftChan := make(chan []int)
	rightChan := make(chan []int)

	go mergeSort(arr[:mid], leftChan)
	go mergeSort(arr[mid:], rightChan)

	left := <-leftChan
	right := <-rightChan

	close(leftChan)
	close(rightChan)

	resultChan <- merge2(left, right)
}

func sortCon(input [][]int) [][]int {
	var wg sync.WaitGroup

	for i := range input {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			resultChan := make(chan []int)
			go mergeSort(input[i], resultChan)
			input[i] = <-resultChan
		}(i)
	}

	wg.Wait()
	return input
}

func mySingleHandler(w http.ResponseWriter, r *http.Request) {
	var reqPayload RequestPayload
	if err := json.NewDecoder(r.Body).Decode(&reqPayload); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	startTime := time.Now()
	sortedArrays := sortSeq(reqPayload.ToSort)
	timeTaken := time.Since(startTime)

	response := ResponsePayload{
		SortedArrays: sortedArrays,
		TimeNS:       timeTaken.Nanoseconds(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func myConcurrentHandler(w http.ResponseWriter, r *http.Request) {
	var reqPayload RequestPayload
	if err := json.NewDecoder(r.Body).Decode(&reqPayload); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	startTime := time.Now()
	sortedArrays := sortCon(reqPayload.ToSort)
	timeTaken := time.Since(startTime)

	response := ResponsePayload{
		SortedArrays: sortedArrays,
		TimeNS:       timeTaken.Nanoseconds(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/process-single", mySingleHandler)
	http.HandleFunc("/process-concurrent", myConcurrentHandler)
	// bil()
	port := ":8000"
	if err := http.ListenAndServe(port, nil); err != nil {
		panic(err)
	}
}
