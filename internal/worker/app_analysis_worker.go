package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"job-tracker/internal/entity"
)

type AppAnalysisWorkerPool struct {
	concurrency int
	deps        AppAnalysisDeps
	jobCh       chan int
	ctx         context.Context
	wg          sync.WaitGroup
}

func NewAppAnalysisPool(ctx context.Context, jobCh chan int, deps AppAnalysisDeps) *AppAnalysisWorkerPool {
	return &AppAnalysisWorkerPool{
		concurrency: 3,
		ctx:         ctx,
		deps:        deps,
		jobCh:       jobCh,
		wg:          sync.WaitGroup{},
	}
}

func (p *AppAnalysisWorkerPool) Start() {
	if err := p.deps.JobRepo.ResetStale(); err != nil {
		fmt.Printf("warn: failed to reset stale app analysis jobs: %s\n", err.Error())
	}
	for i := 0; i < p.concurrency; i++ {
		p.wg.Add(1)
		go p.runWorker()
	}
}

func (p *AppAnalysisWorkerPool) Wait() {
	p.wg.Wait()
}

func (p *AppAnalysisWorkerPool) runWorker() {
	defer p.wg.Done()
	backoff := 1 * time.Second

	for {
		select {
		case <-p.ctx.Done():
			return

		case _, ok := <-p.jobCh:
			if !ok {
				return
			}
			p.processNext()
			backoff = 1 * time.Second

		case <-time.After(backoff):
			claimed := p.processNext()
			if !claimed {
				backoff = min(backoff*2, 30*time.Second)
			} else {
				backoff = 1 * time.Second
			}
		}
	}
}

func (p *AppAnalysisWorkerPool) processNext() bool {
	job, err := p.deps.JobRepo.ClaimNext()
	if err != nil {
		return false
	}

	result, err := ProcessAppAnalysisJob(p.ctx, *job, p.deps)
	if err != nil {
		p.deps.JobRepo.MarkFailed(job, err.Error())
		return true
	}

	result = strings.TrimSpace(result)
	if strings.HasPrefix(result, "```") {
		if i := strings.Index(result, "\n"); i != -1 {
			result = result[i+1:]
		}
		result = strings.TrimSuffix(strings.TrimSpace(result), "```")
		result = strings.TrimSpace(result)
	}

	var res entity.AppAnalysisResult
	if err := json.Unmarshal([]byte(result), &res); err != nil {
		p.deps.JobRepo.MarkFailed(job, err.Error())
		return true
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
