package worker

import (
	"context"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/gibson7780/go-project/internal/stats"
	"github.com/gibson7780/go-project/internal/urls"
	"github.com/redis/go-redis/v9"
)

type worker struct {
	Ctx          context.Context
	RedisClient  redis.UniversalClient
	UrlsService  urls.Service
	StatsService stats.Service
	JobWg        sync.WaitGroup
	OutsideWg    sync.WaitGroup
}

func NewWorker(ctx context.Context, redisClient redis.UniversalClient, urlsService urls.Service, statsService stats.Service) *worker {

	worker := &worker{
		Ctx:          ctx,
		RedisClient:  redisClient,
		UrlsService:  urlsService,
		StatsService: statsService,
	}

	worker.Start(ctx)

	return worker
}

func (w *worker) Start(ctx context.Context) {
	w.OutsideWg.Go(func() {
		w.SetFlushJob(ctx)
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

func (w *worker) SetFlushJob(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.JobWg.Go(func() {
				w.UpdateClicksStatJob(ctx)
			})

		case <-ctx.Done():
			w.JobWg.Wait()
			return
		}
	}
}

func (w *worker) Wait() {
	w.OutsideWg.Wait()
}
