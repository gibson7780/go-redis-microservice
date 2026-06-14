package worker

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/gibson7780/go-project/internal/jobs"
	"github.com/gibson7780/go-project/internal/stats"
	"github.com/gibson7780/go-project/internal/urls"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var releaseLockScript = redis.NewScript(`
	local currentWorkerID = redis.call("GET", KEYS[1])
	if currentWorkerID == ARGV[1] then
		return redis.call("DEL", KEYS[1])
	end
	return 0
`)

type worker struct {
	Ctx          context.Context
	RedisClient  redis.UniversalClient
	UrlsService  urls.Service
	StatsService stats.Service
	JobsService  JobsWorker
	JobWg        sync.WaitGroup
	OutsideWg    sync.WaitGroup
	WorkerID     string
}

type JobsWorker interface {
	GetJobs(ctx context.Context) (*[]jobs.Job, error)
	SetStatus(ctx context.Context, job *jobs.Job) error
}

func NewWorker(ctx context.Context, redisClient redis.UniversalClient, urlsService urls.Service, statsService stats.Service, jobsService JobsWorker) *worker {

	worker := &worker{
		Ctx:          ctx,
		RedisClient:  redisClient,
		UrlsService:  urlsService,
		StatsService: statsService,
		JobsService:  jobsService,
		WorkerID:     uuid.New().String(),
	}

	worker.Start(ctx)

	return worker
}

func (w *worker) Start(ctx context.Context) {
	w.OutsideWg.Go(func() {
		w.Run(ctx)
	})
}

func (w *worker) UpdateClicksStatJob(ctx context.Context) {
	clicks, err := w.RedisClient.HGetAll(ctx, "clicks").Result()
	if err != nil {
		slog.Error("redis error", "err", err)
	}

	w.RedisClient.Rename(ctx, "clicks", "flush_clicks")

	urlCountData := map[string]int64{}

	for index, value := range clicks {
		count, _ := strconv.ParseInt(value, 10, 64)
		urlCountData[index] = count
	}

	err = w.StatsService.BatchStats(urlCountData)

	if err != nil {
		slog.Error("redis error", "err", err)
	}

	w.RedisClient.Del(ctx, "flush_clicks")
}

func (w *worker) Run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.JobWg.Go(func() {
				w.UpdateClicksStatJob(ctx)
			})
			w.JobWg.Go(func() {
				w.UpdateJobs(ctx)
			})
		case <-ctx.Done():
			w.JobWg.Wait()
			return
		}
	}
}

// job id 是每個pending的job的id
// worker id是這個worker instance的id
// 所以不同worker會有一樣的job id, 不同的worker id
func (w *worker) UpdateJobs(ctx context.Context) {
	pendingJobs, err := w.JobsService.GetJobs(ctx)
	if err != nil {
		return
	}
	for _, job := range *pendingJobs {
		locked, err := w.RedisClient.SetNX(ctx, fmt.Sprintf("lock:job:%s", job.ID.String()), w.WorkerID, 30*time.Second).Result()
		if err != nil {
			slog.Info("job err", "err", err)
			continue
		}
		if !locked {
			slog.Info("job duplicate", "err", err)
			continue
		}
		payload := &jobs.Job{
			ID:        job.ID,
			Status:    "success",
			UpdatedAt: time.Now(),
		}
		err = w.JobsService.SetStatus(ctx, payload)
		if err != nil {
			slog.Info("update error", "err", err)
			continue
		}
		// del
		jobID := fmt.Sprintf("lock:job:%s", job.ID.String())
		releaseLockScript.Run(ctx, w.RedisClient, []string{jobID}, w.WorkerID)
		// w.RedisClient.Del(ctx, fmt.Sprintf("lock:job:%s", job.ID.String()))
	}
}

func (w *worker) Wait() {
	w.OutsideWg.Wait()
}
