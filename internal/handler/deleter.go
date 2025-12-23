package handler

import (
	"context"

	"github.com/Oleg2210/goshortener/internal/service"
	"go.uber.org/zap"
)

type DeleteTask struct {
	UserID string
	Short  string
}

type Deleter struct {
	ctx     context.Context
	queue   chan DeleteTask
	service *service.ShortenerService
	logger  *zap.Logger
}

func NewDeleter(ctx context.Context, logger *zap.Logger, service *service.ShortenerService, workers int) *Deleter {
	d := &Deleter{
		ctx:     ctx,
		queue:   make(chan DeleteTask),
		service: service,
		logger:  logger,
	}

	for i := 0; i < workers; i++ {
		go d.worker()
	}

	return d
}

func (d *Deleter) worker() {
	for {
		select {
		case <-d.ctx.Done():
			return
		case task, ok := <-d.queue:
			if !ok {
				return
			}

			err := d.service.MarkDelete(d.ctx, task.Short, task.UserID)
			d.logger.Error("failed to mark delete "+task.Short, zap.Error(err))
		}
	}
}
