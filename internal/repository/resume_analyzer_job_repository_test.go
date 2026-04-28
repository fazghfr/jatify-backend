package repository

import (
	"job-tracker/internal/entity"
	"testing"
	"time"
)

func makeJob(retryCount, maxRetries int) *entity.ResumeAnalysisJob {
	return &entity.ResumeAnalysisJob{
		RetryCount: retryCount,
		MaxRetries: maxRetries,
		Status:     "processing",
	}
}

func TestApplyRetryLogic_ShouldRetry(t *testing.T) {
	job := makeJob(0, 3)
	now := time.Now()

	applyRetryLogic(job, "some error", now)

	if job.Status != "pending" {
		t.Errorf("expected status=pending, got %s", job.Status)
	}
	if job.RetryCount != 1 {
		t.Errorf("expected retry_count=1, got %d", job.RetryCount)
	}
	if job.NextRetryAt == nil {
		t.Error("expected next_retry_at to be set")
	}
	// delay for retry 0 = 30s * 2^0 = 30s
	expectedDelay := 30 * time.Second
	actualDelay := job.NextRetryAt.Sub(now)
	if actualDelay < expectedDelay-time.Second || actualDelay > expectedDelay+time.Second {
		t.Errorf("expected ~30s delay, got %v", actualDelay)
	}
	if job.TimeFinished != nil {
		t.Error("expected time_finished to be nil on retry")
	}
	if job.ErrorMsg != nil {
		t.Error("expected error_msg to be nil on retry")
	}
}

func TestApplyRetryLogic_DelayDoublesEachRetry(t *testing.T) {
	cases := []struct {
		retryCount    int
		expectedDelay time.Duration
	}{
		{0, 30 * time.Second},
		{1, 60 * time.Second},
		{2, 120 * time.Second},
	}

	for _, tc := range cases {
		job := makeJob(tc.retryCount, 10) // high max so it always retries
		now := time.Now()

		applyRetryLogic(job, "error", now)

		actual := job.NextRetryAt.Sub(now)
		if actual < tc.expectedDelay-time.Second || actual > tc.expectedDelay+time.Second {
			t.Errorf("retry %d: expected ~%v delay, got %v", tc.retryCount, tc.expectedDelay, actual)
		}
	}
}

func TestApplyRetryLogic_PermanentFailOnLastRetry(t *testing.T) {
	job := makeJob(2, 3) // one more attempt would hit max
	now := time.Now()
	errMsg := "llm timeout"

	applyRetryLogic(job, errMsg, now)

	if job.Status != "failed" {
		t.Errorf("expected status=failed, got %s", job.Status)
	}
	if job.ErrorMsg == nil || *job.ErrorMsg != errMsg {
		t.Errorf("expected error_msg=%q", errMsg)
	}
	if job.TimeFinished == nil {
		t.Error("expected time_finished to be set")
	}
	if job.NextRetryAt != nil {
		t.Error("expected next_retry_at to be nil on permanent failure")
	}
	if job.RetryCount != 2 {
		t.Errorf("expected retry_count unchanged at 2, got %d", job.RetryCount)
	}
}

func TestApplyRetryLogic_MaxRetriesZero(t *testing.T) {
	job := makeJob(0, 0) // never retry
	now := time.Now()

	applyRetryLogic(job, "error", now)

	if job.Status != "failed" {
		t.Errorf("expected immediate failure, got %s", job.Status)
	}
}

func TestApplyRetryLogic_MaxRetriesOne(t *testing.T) {
	job := makeJob(0, 1) // only one attempt allowed, no retries
	now := time.Now()

	applyRetryLogic(job, "error", now)

	if job.Status != "failed" {
		t.Errorf("expected status=failed, got %s", job.Status)
	}
}
