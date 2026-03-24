package handler

import (
	"context"
	"strings"
	"sync"

	"github.com/Oleg2210/goshortener/internal/service"
	"go.uber.org/zap"
)

type DeleteTask struct {
	UserID string
	Shorts []string
}

type Deleter struct {
	ctx     context.Context
	queue   chan DeleteTask
	service *service.ShortenerService
	logger  *zap.Logger
	wg      *sync.WaitGroup
}

func NewDeleter(ctx context.Context, wg *sync.WaitGroup, logger *zap.Logger, service *service.ShortenerService, workers int) *Deleter {
	d := &Deleter{
		ctx:     ctx,
		queue:   make(chan DeleteTask, workers),
		service: service,
		logger:  logger,
		wg:      wg,
	}

	wg.Add(1)
	for i := 0; i < workers; i++ {
		go d.worker()
	}

	return d
}

func (d *Deleter) worker() {
	defer d.wg.Done()
	for {
		select {
		case <-d.ctx.Done():
			return
		case task, ok := <-d.queue:
			if !ok {
				return
			}

			err := d.service.MarkDelete(d.ctx, task.Shorts, task.UserID)
			d.logger.Error("failed to mark delete "+strings.Join(task.Shorts, ","), zap.Error(err))
		}
	}
}
