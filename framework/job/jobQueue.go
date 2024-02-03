package job

import (
	"sync"
)

// JobQueue is a FIFO job queue
type JobQueue struct {
	jobs  []Job
	mutex sync.Mutex
}

func NewJobQueue() *JobQueue {
	return &JobQueue{}
}

// AddJob adds a new job to the queue
func (jobQueue *JobQueue) AddJob(job Job) {
	jobQueue.mutex.Lock()
	defer jobQueue.mutex.Unlock()

	jobQueue.jobs = append(jobQueue.jobs, job)
}

// PopQueue returns the next item from the queue
func (jobQueue *JobQueue) PopQueue() Job {
	jobQueue.mutex.Lock()
	defer jobQueue.mutex.Unlock()

	if len(jobQueue.jobs) == 0 {
		return nil
	}

	job := jobQueue.jobs[0]
	jobQueue.jobs = jobQueue.jobs[1:]
	return job
}

func (jobQueue *JobQueue) Size() int {
	jobQueue.mutex.Lock()
	defer jobQueue.mutex.Unlock()
	return len(jobQueue.jobs)
}
