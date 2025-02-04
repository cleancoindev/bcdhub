package ws

import (
	"github.com/baking-bad/bcdhub/cmd/api/handlers"
	"github.com/baking-bad/bcdhub/cmd/api/ws/datasources"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/mq"
)

// HubOption -
type HubOption func(*Hub)

// WithSource -
func WithSource(source datasources.DataSource) HubOption {
	return func(h *Hub) {
		h.sources = append(h.sources, source)
	}
}

// WithRabbitSource -
func WithRabbitSource(messageQueue mq.Mediator) HubOption {
	return func(h *Hub) {
		rmq, err := datasources.NewRabbitMQ(messageQueue)
		if err != nil {
			logger.Error(err)
			return
		}
		h.sources = append(h.sources, rmq)
	}
}

// WithContext -
func WithContext(ctx *handlers.Context) HubOption {
	return func(h *Hub) {
		h.ctx = ctx
	}
}
