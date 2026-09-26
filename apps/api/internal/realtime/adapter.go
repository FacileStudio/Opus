package realtime

import "context"

// Handler processes envelopes received from an adapter bus.
type Handler func(Envelope)

// Adapter distributes broadcast envelopes across instances.
type Adapter interface {
	Publish(ctx context.Context, env Envelope) error
	Subscribe(ctx context.Context, handler Handler) error
	Close() error
}
