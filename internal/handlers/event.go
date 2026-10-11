package handlers

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/common"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/dto"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/middleware"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/models"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/services"
)

type eventService interface {
	Create(ctx context.Context, event models.Event) (models.Event, error)
	ListWithdrawals(ctx context.Context, userID uuid.UUID) ([]models.Event, error)
}

type EventHandler struct {
	events eventService
}

func NewEventHandler(events *services.EventService) *EventHandler {
	return &EventHandler{events: events}
}

func (h *EventHandler) Register(g *echo.Group) {
	user := g.Group("/user", middleware.JWT())
	user.POST("/balance/withdraw", h.withdraw)
	user.GET("/withdrawals", h.withdrawals)
}

func (h *EventHandler) withdraw(c *echo.Context) error {
	userID, err := common.UserIDFromContext(c)
	if err != nil {
		return mapAppError(err)
	}

	var req dto.WithdrawRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, invalidRequestBody)
	}
	if err := req.Validate(); err != nil {
		return mapAppError(err)
	}

	if _, err := h.events.Create(c.Request().Context(), models.Event{
		UserID:      userID,
		OrderNumber: req.Order,
		Amount:      req.Sum,
		EventType:   models.EventTypeWithdrawal,
	}); err != nil {
		return mapAppError(err)
	}
	return c.NoContent(http.StatusOK)
}

func (h *EventHandler) withdrawals(c *echo.Context) error {
	userID, err := common.UserIDFromContext(c)
	if err != nil {
		return mapAppError(err)
	}

	events, err := h.events.ListWithdrawals(c.Request().Context(), userID)
	if err != nil {
		return mapAppError(err)
	}
	if len(events) == 0 {
		return c.NoContent(http.StatusNoContent)
	}

	response := make([]dto.WithdrawalResponse, 0, len(events))
	for _, event := range events {
		response = append(response, dto.WithdrawalResponse{
			Order:       event.OrderNumber,
			Sum:         event.Amount,
			ProcessedAt: event.CreatedAt,
		})
	}
	return c.JSON(http.StatusOK, response)
}
