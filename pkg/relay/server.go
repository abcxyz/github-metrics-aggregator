package relay

import (
	"context"
	"fmt"
	"net/http"

	"github.com/abcxyz/github-metrics-aggregator/pkg/version"
	"github.com/abcxyz/github-metrics-aggregator/pkg/webhook"
	"github.com/abcxyz/pkg/logging"
	"google.golang.org/api/option"
)

// Messenger defines the interface for sending messages to a relay destination.
type Messenger interface {
	Send(ctx context.Context, msg []byte) error
}

// Server acts as a HTTP server for handling relay requests.
type Server struct {
	config         *Config
	relayMessenger Messenger
}

// NewServer creates a new instance of the Server.
func NewServer(ctx context.Context, cfg *Config) (*Server, error) {

	agent := fmt.Sprintf("abcxyz:github-metrics-aggregator/%s", version.Version)
	relayMessenger, err := webhook.NewPubSubMessenger(ctx, cfg.RelayProjectID, cfg.RelayTopicID, option.WithUserAgent(agent))
	if err != nil {
		return nil, fmt.Errorf("failed to create event pubsub: %w", err)
	}

	return &Server{
		config:         cfg,
		relayMessenger: relayMessenger,
	}, nil
}

// Routes creates a ServeMux of all of the routes that
// this Router supports.
func (s *Server) Routes(ctx context.Context) http.Handler {
	logger := logging.FromContext(ctx)
	mux := http.NewServeMux()
	mux.Handle("/", s.handleRelay())

	// Middleware
	root := logging.HTTPInterceptor(logger, s.config.ProjectID)(mux)

	return root
}
