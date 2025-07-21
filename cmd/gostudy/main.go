package main

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// This controls the maxprocs environment variable in container runtimes.
// see https://martin.baillie.id/wrote/gotchas-in-the-go-network-packages-defaults/#bonus-gomaxprocs-containers-and-the-cfs

func main() {

	// // Logger configuration
	// logger := log.New(
	// 	log.WithLevel(os.Getenv("LOG_LEVEL")),
	// 	log.WithSource(),
	// )

	// #51 배열과 슬라이스를 명확히 구분하라
	// a := [3]int{0, 1, 2}
	// for i, v := range a {
	// 	a[2] = 10
	// 	if i == 2 {
	// 		fmt.Println(v)
	// 	}
	// }

	// #52 에러를 두 번 처리하지 마라
	// if err := run(logger); err != nil {
	// 	logger.ErrorContext(context.Background(), "an error occurred", slog.String("error", err.Error()))
	// 	os.Exit(1)
	// }

	// #56 동시성이 무조건 빠르다고 착각하지 마라
	// benchmarkMergeSort(sequentialMergeSort)
	// benchmarkMergeSort(parallelMergeSortV1)
	// benchmarkMergeSort(parallelMergeSortV2)

	// #58 경쟁 문제에 대해 완전히 이해하라

	// data race
	var wg sync.WaitGroup

	i := 0

	wg.Add(2)

	go func() {
		defer wg.Done()
		i++
	}()

	go func() {
		defer wg.Done()
		i++
	}()

	wg.Wait()
	fmt.Println(i)

	// atomic
	var i2 int64

	var wg2 sync.WaitGroup

	wg2.Add(2)

	go func() {
		defer wg2.Done()
		atomic.AddInt64(&i2, 1)
	}()

	go func() {
		defer wg2.Done()
		atomic.AddInt64(&i2, 1)
	}()

	wg2.Wait()
	fmt.Println(i2)

	// mutex
	i3 := 0
	mutex := sync.Mutex{}

	w3 := sync.WaitGroup{}

	w3.Add(2)

	go func() {
		defer w3.Done()
		mutex.Lock()
		i3++
		mutex.Unlock()
	}()

	go func() {
		defer w3.Done()
		mutex.Lock()
		i3++
		mutex.Unlock()
	}()

	w3.Wait()
	fmt.Println(i3)

	// channel
	i4 := 0
	ch := make(chan int)

	wg4 := sync.WaitGroup{}
	wg4.Add(2)
	go func() {
		defer wg4.Done()
		ch <- 1
	}()

	go func() {
		defer wg4.Done()
		ch <- 1
	}()

	i4 += <-ch
	i4 += <-ch

	wg4.Wait()
	fmt.Println(i4)

	// race condition

	i5 := 0
	wg5 := sync.WaitGroup{}
	wg5.Add(2)

	go func() {
		defer wg5.Done()
		i5 = 1
	}()

	go func() {
		defer wg5.Done()
		i5 = 2
	}()

	wg5.Wait()
	fmt.Println(i5)
}

// func run(logger *slog.Logger) error {
// 	ctx := context.Background()

// 	_, err := maxprocs.Set(maxprocs.Logger(func(s string, i ...interface{}) {
// 		logger.DebugContext(ctx, fmt.Sprintf(s, i...))
// 	}))
// 	if err != nil {
// 		return fmt.Errorf("setting max procs: %w", err)
// 	}

// 	logger.InfoContext(ctx, "Hello world!", slog.String("location", "world"))

// 	return nil
// }

func benchmarkMergeSort(sortFunc func([]int)) {
	const benchmarkRuns = 5
	const arraySize = 10000

	var totalTime time.Duration
	for run := 0; run < benchmarkRuns; run++ {
		arr := make([]int, arraySize)
		for i := 0; i < arraySize; i++ {
			arr[i] = rand.Intn(arraySize)
		}

		start := time.Now()
		sortFunc(arr)
		elapsed := time.Since(start)
		totalTime += elapsed
	}

	avgTime := totalTime / benchmarkRuns
	fmt.Printf("\nAverage time over %d runs: %s\n", benchmarkRuns, avgTime)
}
