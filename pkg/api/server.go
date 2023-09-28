package api

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/cors"
	"go.uber.org/zap"

	"github.com/tonkeeper/tonkeeper-twa-api/pkg/api/oas"
)

type Server struct {
	logger     *zap.Logger
	httpServer *http.Server
}

type ServerOptions struct {
}

type ServerOption func(options *ServerOptions)

func NewServer(log *zap.Logger, pool *pgxpool.Pool, handler *Handler, address string, opts ...ServerOption) (*Server, error) {
	options := &ServerOptions{}
	for _, o := range opts {
		o(options)
	}
	ogenServer, err := oas.NewServer(handler)
	if err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	mux.Handle("/", ogenServer)
	mux.HandleFunc("/healthz", healthzHandler(pool))

	serv := Server{
		logger: log,
		httpServer: &http.Server{
			Addr:    address,
			Handler: cors.Default().Handler(mux),
		},
	}
	return &serv, nil
}

func (s *Server) Run() {
	err := s.httpServer.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		s.logger.Info("tonkeeper-twa-api quit")
		return
	}
	s.logger.Fatal("ListedAndServe() failed", zap.Error(err))
}

func healthzHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := pool.Ping(r.Context()); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
