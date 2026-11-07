package scheduler

import "github.com/codicecustode/inference-scheduler/internal/models"

type JobQueue []*models.Job

func (q JobQueue) Len() int{
	return len(q)
}

func (q JobQueue) Less(i,j int) bool{

	if q[i].Priority != q[j].Priority{
		return q[i].Priority > q[j].Priority
	}

	return q[i].CreatedAt.Before(q[j].CreatedAt)

}

func (q JobQueue) Swap (i,j int){
	q[i],q[j] = q[j], q[i]
}

func (q *JobQueue) Push(x interface{}) {
    *q = append(*q, x.(*models.Job))
}


func (q *JobQueue) Pop() interface{} {
    old := *q
    n := len(old)
    job := old[n-1]
    *q = old[:n-1]
    return job
}
