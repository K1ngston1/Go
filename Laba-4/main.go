package main

import (
	"fmt"
	"sync"
	"time"
)

type Job struct {
	run func()
}

type WorkerPool struct {
	jobs chan Job
	wg   *sync.WaitGroup
}

func NewWorkerPool(numWorkers int) *WorkerPool {
	wp := &WorkerPool{
		jobs: make(chan Job),
		wg:   &sync.WaitGroup{},
	}

	for i := 1; i <= numWorkers; i++ {
		go func(id int) {
			for job := range wp.jobs {
				fmt.Printf("Воркер %d взяв задачу\n", id)
				job.run()
				wp.wg.Done()
			}
		}(i)
	}

	return wp
}

func (wp *WorkerPool) AddJob(job Job) {
	wp.wg.Add(1)
	wp.jobs <- job
}

func (wp *WorkerPool) Close() {
	wp.wg.Wait()
	close(wp.jobs)
}

func main() {
	counter := 0
	var mu sync.Mutex
	pool := NewWorkerPool(3)

	for i := 0; i < 10; i++ {
		pool.AddJob(Job{
			run: func() {
				time.Sleep(300 * time.Millisecond)
				mu.Lock()
				counter++
				fmt.Printf("Воркер виконав задачу, лічильник = %d\n", counter)
				mu.Unlock()
			},
		})
	}
	pool.Close()
	fmt.Println("Усі задачі виконано. Кінцеве значення:", counter)
}
