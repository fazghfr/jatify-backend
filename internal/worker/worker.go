package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type WorkerPool struct {
	concurrency int
	deps ProcessorDeps
	jobCh chan int
	ctx context.Context
	wg sync.WaitGroup
}

func NewPool(ctx context.Context, jobCh chan int, deps ProcessorDeps) *WorkerPool {
	return &WorkerPool{
		concurrency: 3,
		ctx: ctx,
		deps: deps,
		jobCh: jobCh,
		wg: sync.WaitGroup{},
	}
}

func (p *WorkerPool) Start () {
	err := p.deps.JobRepo.ResetStale()
	if err != nil {
		fmt.Println("warn: failed to reset stale jobs: %s", err.Error())
	}
	// spawning goroutines
	for i := 0; i < p.concurrency; i++ {
		p.wg.Add(1)
		go p.runWorker()
	}
}

func (p *WorkerPool) Wait() {
    p.wg.Wait()
}


func (p *WorkerPool) runWorker() {
    defer p.wg.Done()
    backoff := 1 * time.Second

    for {
        select {
        case <-p.ctx.Done():
            return

        case _, ok := <-p.jobCh:
            if !ok { return }
            p.processNext()
            backoff = 1 * time.Second  // reset backoff on success

        case <-time.After(backoff):
            // slow path: poll DB
            claimed := p.processNext()
            if !claimed {
                // nothing in DB either — back off
                backoff = min(backoff*2, 30*time.Second)
            } else {
                backoff = 1 * time.Second
            }
        }
    }
}

func (p *WorkerPool) processNext() bool {
    job, err := p.deps.JobRepo.ClaimNext()
    if err != nil {
        return false  // nothing to process
    }

    result, err := ProcessJob(p.ctx, *job, p.deps)
    if err != nil {
        p.deps.JobRepo.MarkFailed(job, err.Error())
        return true
    }


    // cleaning result here
    type Res struct {
		Summary               string   `json:"summary"`
		Skills                []string `json:"skills"`
		ExperienceYears       int      `json:"experience_years"`
		Strengths             []string `json:"strengths"`
		Weaknesses            []string `json:"weaknesses"`
		ImprovementSuggestions []string `json:"improvement_suggestions"`
		OverallScore          int      `json:"overall_score"`
    }

    var res Res
    if err := json.Unmarshal([]byte(result), &res); err != nil {
    	p.deps.JobRepo.MarkFailed(job, err.Error())
    	return false
    }

    cleaned, err := json.Marshal(res)
    if err != nil {
        p.deps.JobRepo.MarkFailed(job, err.Error())
        return true
    }

    job.ResultJSON = string(cleaned)
    p.deps.JobRepo.MarkDone(job)
    return true
}
