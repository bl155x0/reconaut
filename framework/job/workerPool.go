package job

import (
	"fmt"
	"reconaut/iobuffer"
	"sync"
	"time"
)

type WorkerStatus int

const (
	WorkerStatusIdle = iota
	WorkerStatusRunning
)

type WorkerPool struct {
	jobQueue          *JobQueue
	shutdown          chan struct{}
	waitGroup         sync.WaitGroup
	numberWorkers     int
	workerStatus      map[int]WorkerStatus
	workerStatusMutex sync.Mutex
}

func NewWorkerPool(numberWorkers int, jobQueue *JobQueue) *WorkerPool {
	return &WorkerPool{
		jobQueue:      jobQueue,
		shutdown:      make(chan struct{}),
		numberWorkers: numberWorkers,
		workerStatus:  make(map[int]WorkerStatus),
	}
}

// Start begins processing with the given amount of workers
func (workerPool *WorkerPool) Start() {
	for i := 0; i < workerPool.numberWorkers; i++ {
		workerPool.workerStatus[i] = WorkerStatusIdle
	}
	for i := 0; i < workerPool.numberWorkers; i++ {
		//Add the new worker go-routine to the waitgroup
		workerPool.waitGroup.Add(1)
		go func(workerId int) {
			//remove the from the waitgroup when the worker was shutdown
			defer workerPool.waitGroup.Done()

			//Start the worker
			workerPool.process(workerId)
		}(i)
	}
}

func (workerPool *WorkerPool) Stop() {
	close(workerPool.shutdown)
	workerPool.waitGroup.Wait()
}

// DoJob enqueues a job and executes it with the next free worker of the pool
func (workerPool *WorkerPool) DoJob(job Job) {
	workerPool.jobQueue.AddJob(job)
}

// DoJob enqueues a list of job and executes it with the next free worker of the pool
func (workerPool *WorkerPool) DoJobs(jobs []Job) {
	for _, job := range jobs {
		workerPool.jobQueue.AddJob(job)
	}
}

func (workerPool *WorkerPool) setWorkerStatus(workerId int, status WorkerStatus) {
	workerPool.workerStatusMutex.Lock()
	defer workerPool.workerStatusMutex.Unlock()
	workerPool.workerStatus[workerId] = status
}

func (workerPool *WorkerPool) hasWorkerRunning() bool {
	workerPool.workerStatusMutex.Lock()
	defer workerPool.workerStatusMutex.Unlock()
	for _, s := range workerPool.workerStatus {
		if s == WorkerStatusRunning {
			return true
		}
	}
	return false
}

func (workerPool *WorkerPool) hasJobsInQueue() bool {
	return workerPool.jobQueue.Size() > 0

}

func (workerPool *WorkerPool) Finished() bool {
	return workerPool.hasWorkerRunning() == false &&
		workerPool.hasJobsInQueue() == false
}

// process is a worker functin that processes jobs from the queue
func (workerPool *WorkerPool) process(workerId int) {
	for {
		select {
		case <-workerPool.shutdown:
			//Shutdown signal received: Let the worker exit
			return
		default:
			// Look for a new job to do:
			job := workerPool.jobQueue.PopQueue()
			if job == nil {
				// No more jobs. IDLE
				workerPool.setWorkerStatus(workerId, WorkerStatusIdle)
				continue
			}
			workerPool.setWorkerStatus(workerId, WorkerStatusRunning)
			nextJobs := job.Execute()
			if nextJobs != nil {
				workerPool.DoJobs(nextJobs)
			}
			iobuffer.GetIOBuffer().AddOutputVerbose(fmt.Sprintf("[Worker %d] Finished job %s", workerId, job.ID()))
		}
		time.Sleep(1 * time.Millisecond)
	}
}
