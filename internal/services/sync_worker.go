package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/models"
	"go.uber.org/zap"
)

const accrualPollInterval = time.Second

type accrualOrderStore interface {
	ListPending(ctx context.Context) ([]models.Order, error)
	ApplyAccrual(ctx context.Context, tx *sql.Tx, number, status string, accrual float64) (uuid.UUID, bool, error)
}

type accrualResult struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

type tooManyRequestsError struct {
	retryAfter time.Duration
}

func (e tooManyRequestsError) Error() string {
	return "accrual too many requests"
}

type SyncWorker struct {
	baseURL string
	db      *sql.DB
	orders  accrualOrderStore
	users   userBalanceStore
	client  *http.Client
	logger  *zap.SugaredLogger
}

func NewSyncWorker(baseURL string, db *sql.DB, orders accrualOrderStore, users userBalanceStore, logger *zap.SugaredLogger) *SyncWorker {
	return &SyncWorker{
		baseURL: strings.TrimRight(baseURL, "/"),
		db:      db,
		orders:  orders,
		users:   users,
		client:  &http.Client{Timeout: 10 * time.Second},
		logger:  logger,
	}
}

func (w *SyncWorker) Run(ctx context.Context) {
	if w.baseURL == "" {
		return
	}

	ticker := time.NewTicker(accrualPollInterval)
	defer ticker.Stop()

	w.tick(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.tick(ctx)
		}
	}
}

func (w *SyncWorker) tick(ctx context.Context) {
	orders, err := w.orders.ListPending(ctx)
	if err != nil {
		w.logger.Errorw("list pending orders", "error", err)
		return
	}

	for _, order := range orders {
		if err := w.syncOrder(ctx, order); err != nil {
			var limited tooManyRequestsError
			if errors.As(err, &limited) {
				timer := time.NewTimer(limited.retryAfter)
				defer timer.Stop()
				select {
				case <-ctx.Done():
				case <-timer.C:
				}
				return
			}
			w.logger.Errorw("sync order", "order", order.OrderNumber, "error", err)
		}
	}
}

func (w *SyncWorker) syncOrder(ctx context.Context, order models.Order) error {
	result, err := w.fetchAccrual(ctx, order.OrderNumber)
	if err != nil || result == nil {
		return err
	}

	status := mapAccrualStatus(result.Status)
	if status == "" {
		return nil
	}

	tx, err := w.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	userID, updated, err := w.orders.ApplyAccrual(ctx, tx, order.OrderNumber, status, result.Accrual)
	if err != nil {
		return err
	}
	if updated && status == models.OrderStatusProcessed && result.Accrual > 0 {
		if err := w.users.AddBalance(ctx, tx, userID, result.Accrual); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (w *SyncWorker) fetchAccrual(ctx context.Context, number string) (*accrualResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, w.baseURL+"/api/orders/"+number, nil)
	if err != nil {
		return nil, err
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNoContent:
		return nil, nil
	case http.StatusTooManyRequests:
		retryAfter := time.Second
		if seconds, err := strconv.Atoi(resp.Header.Get("Retry-After")); err == nil && seconds > 0 {
			retryAfter = time.Duration(seconds) * time.Second
		}
		io.Copy(io.Discard, resp.Body)
		return nil, tooManyRequestsError{retryAfter: retryAfter}
	case http.StatusOK:
		var result accrualResult
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("decode accrual: %w", err)
		}
		return &result, nil
	default:
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("accrual status %d: %s", resp.StatusCode, body)
	}
}

func mapAccrualStatus(status string) string {
	switch status {
	case models.OrderStatusRegistered, models.OrderStatusProcessing:
		return models.OrderStatusProcessing
	case models.OrderStatusProcessed:
		return models.OrderStatusProcessed
	case models.OrderStatusInvalid:
		return models.OrderStatusInvalid
	default:
		return ""
	}
}
