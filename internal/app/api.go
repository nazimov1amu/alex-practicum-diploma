package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/config"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/db"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/routes"
	"go.uber.org/zap"
)

type App struct {
	logger  *zap.SugaredLogger
	db      *sql.DB
	echo    *echo.Echo
	address string
	ctx     context.Context
	cancel  context.CancelFunc
	done    chan struct{}
}

func NewApp() (*App, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("logger: %w", err)
	}
	sugar := logger.Sugar()
	defer sugar.Sync()

	cfg, err := config.NewAppConfig()
	if err != nil {
		return nil, err
	}

	database, err := db.NewDB(cfg.DatabaseURL)
	if err != nil {
		sugar.Errorw("connect database", "error", err)
		return nil, err
	}

	address := cfg.Address
	if address == "" {
		address = ":8080"
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &App{
		logger:  sugar,
		db:      database,
		echo:    routes.Router(nil),
		address: address,
		ctx:     ctx,
		cancel:  cancel,
		done:    make(chan struct{}),
	}, nil
}

func (a *App) Run() error {
	defer close(a.done)

	sc := echo.StartConfig{
		Address:         a.address,
		GracefulTimeout: 10 * time.Second,
	}
	if err := sc.Start(a.ctx, a.echo); err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, http.ErrServerClosed) {
		a.logger.Errorw("server", "error", err)
		return err
	}

	return nil
}

func (a *App) Stop() error {
	a.cancel()
	<-a.done

	err := a.db.Close()
	return err
}
