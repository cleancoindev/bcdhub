package channels

import (
	"github.com/baking-bad/bcdhub/cmd/api/handlers"
	"github.com/baking-bad/bcdhub/cmd/api/ws/datasources"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/pkg/errors"
)

// ChannelOption -
type ChannelOption func(*DefaultChannel)

// WithSource -
func WithSource(sources []datasources.DataSource, typ string) ChannelOption {
	return func(c *DefaultChannel) {
		source, err := getSourceByType(sources, typ)
		if err != nil {
			logger.Error(err)
			return
		}
		c.sources = append(c.sources, source)
	}
}

// WithContext -
func WithContext(ctx *handlers.Context) ChannelOption {
	return func(c *DefaultChannel) {
		c.ctx = ctx
	}
}

func getSourceByType(sources []datasources.DataSource, typ string) (datasources.DataSource, error) {
	for i := range sources {
		if sources[i].GetType() == typ {
			return sources[i], nil
		}
	}
	return nil, errors.Errorf("unknown source type: %s", typ)
}
