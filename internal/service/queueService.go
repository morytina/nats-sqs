package service

import (
	"context"
	"nats/internal/context/traces"
	"nats/internal/entity"
	"nats/internal/repo"
	"nats/pkg/config"
	"strings"
)

type QueueService interface {
	CreateQueue(ctx context.Context, name, account string) (entity.Queue, error)
	DeleteQueue(ctx context.Context, name string) error
	ListQueues(ctx context.Context, account string) ([]entity.Queue, error)
}

type queueService struct {
	natsRepo repo.NatsRepo
	cfg      *config.Config
}

func NewQueueService(natsRepo repo.NatsRepo, cfg *config.Config) QueueService {
	return &queueService{natsRepo: natsRepo, cfg: cfg}
}

func (s *queueService) CreateQueue(ctx context.Context, name, account string) (entity.Queue, error) {
	_, err := s.natsRepo.CreateStream(ctx, name)
	queue := makeQueueSrn(s.cfg.Region, account, name)
	return queue, err
}

func (s *queueService) DeleteQueue(ctx context.Context, name string) error {
	return s.natsRepo.DeleteStream(ctx, name)
}

func (s *queueService) ListQueues(ctx context.Context, account string) ([]entity.Queue, error) {
	ctx, span := traces.StartSpan(ctx, "listQueues")
	defer span.End()

	namesCh, err := s.natsRepo.ListStreamNames(ctx)
	if err != nil {
		traces.RecordSpanError(ctx, span, "natsRepo.ListStreamNames error", err)
		return nil, err
	}

	var queues []entity.Queue
	for name := range namesCh {
		queues = append(queues, makeQueueSrn(s.cfg.Region, account, name))
	}
	return queues, nil
}

func makeQueueSrn(region, account, name string) entity.Queue {
	var sb strings.Builder
	sb.Grow(len("srn:scp:sns:::") + len(region) + len(account) + len(name))
	sb.WriteString("srn:scp:sns:")
	sb.WriteString(region)
	sb.WriteByte(':')
	sb.WriteString(account)
	sb.WriteByte(':')
	sb.WriteString(name)
	return entity.Queue{QueueSrn: sb.String()}
}
