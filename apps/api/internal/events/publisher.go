package events

import (
	"context"
	"log/slog"
	"math"
	"sync"
	"time"

	antenne "github.com/FacileStudio/antenne-client/go"
	enveloppe "github.com/FacileStudio/enveloppe/go"
	"gorm.io/gorm"
)

// Transport defines the underlying event bus transport methods used by Publisher.
type Transport interface {
	Connect(ctx context.Context) error
	EmitNow(channel string, payload any) error
	Disconnect()
	IsConnected() bool
	Listen(channel string, handler antenne.EventHandler) func()
}

// Emitter defines the public contract for publishing domain events.
type Emitter interface {
	EmitTask(ctx context.Context, channel string, payload TaskPayload) error
	EmitProject(ctx context.Context, channel string, payload ProjectPayload) error
	EmitComment(ctx context.Context, channel string, payload CommentPayload) error
	EmitWorkspace(ctx context.Context, channel string, payload WorkspacePayload) error
	Close()
}

// Option configures a Publisher instance.
type Option func(*Publisher)

// WithTransport sets a custom transport for testing.
func WithTransport(t Transport) Option {
	return func(p *Publisher) {
		p.transport = t
	}
}

// WithLogger sets a custom logger for Publisher.
func WithLogger(l *slog.Logger) Option {
	return func(p *Publisher) {
		p.logger = l
	}
}

// WithStore queues events in the antenne_outbox table of db instead of
// sending them directly, and lets the publisher describe each project's space.
func WithStore(db *gorm.DB) Option {
	return func(p *Publisher) {
		p.db = db
	}
}

// Publisher publishes Opus domain events to Antenne.
type Publisher struct {
	config    Config
	transport Transport
	logger    *slog.Logger
	db        *gorm.DB
	mu        sync.Mutex
	dialed    bool
	published int64
	dropped   int64
	lastError string
}

// New creates a Publisher. If configuration is enabled and no transport
// was provided, an antenne-client instance is constructed.
func New(cfg Config, opts ...Option) *Publisher {
	p := &Publisher{
		config: cfg,
		logger: slog.Default(),
	}
	for _, opt := range opts {
		opt(p)
	}
	if p.transport == nil && cfg.Enabled() {
		p.transport = antenne.NewClient(&antenne.Config{
			App:        string(enveloppe.AppOpus),
			Instance:   cfg.URL,
			Secret:     cfg.Secret,
			InstanceID: cfg.InstanceID,
			Events: antenne.EventConfig{
				Emit:   AllChannels(),
				Listen: InboundChannels(),
			},
		},
			antenne.WithMaxReconnect(math.MaxInt32),
			antenne.WithOnError(func(err error) {
				p.logger.Warn("antenne connection error", slog.Any("error", err))
			}),
		)
	}
	return p
}

// Listen registers a handler for an inbound channel. Call it before Start.
func (p *Publisher) Listen(channel string, handler antenne.EventHandler) {
	if p == nil || p.transport == nil {
		return
	}
	p.transport.Listen(channel, handler)
}

// Start initiates background connection to Antenne and, with a store, the
// outbox worker.
func (p *Publisher) Start(ctx context.Context) {
	if !p.active() {
		return
	}
	go p.dial(ctx)
	if p.db != nil {
		go p.runOutbox(ctx, outboxInterval)
	}
}

func (p *Publisher) dial(ctx context.Context) {
	backoff := 5 * time.Second
	for {
		attempt, cancel := context.WithTimeout(ctx, 10*time.Second)
		err := p.transport.Connect(attempt)
		cancel()
		if err == nil {
			p.mu.Lock()
			p.dialed = true
			p.lastError = ""
			p.mu.Unlock()
			p.logger.Info("connected to antenne", slog.String("url", p.config.URL))
			return
		}
		p.mu.Lock()
		p.lastError = err.Error()
		p.mu.Unlock()
		p.logger.Warn("antenne dial failed, retrying", slog.Any("error", err), slog.Duration("retry_in", backoff))
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}
		if backoff *= 2; backoff > 60*time.Second {
			backoff = 60 * time.Second
		}
	}
}

// Close disconnects from the event bus.
func (p *Publisher) Close() {
	if p == nil || p.transport == nil {
		return
	}
	p.transport.Disconnect()
}

// IsConnected reports whether the publisher transport is connected.
func (p *Publisher) IsConnected() bool {
	if p == nil || p.transport == nil {
		return false
	}
	return p.transport.IsConnected()
}

// Stats returns the number of published and dropped events.
func (p *Publisher) Stats() (int64, int64) {
	if p == nil {
		return 0, 0
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.published, p.dropped
}
