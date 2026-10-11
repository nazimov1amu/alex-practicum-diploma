package handlers

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/common"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/dto"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/middleware"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/models"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/services"
)

type orderService interface {
	Create(ctx context.Context, order models.Order) (models.Order, error)
	List(ctx context.Context, userID uuid.UUID) ([]models.Order, error)
}

type OrderHandler struct {
	orders orderService
}

func NewOrderHandler(orders *services.OrderService) *OrderHandler {
	return &OrderHandler{orders: orders}
}

func (h *OrderHandler) Register(g *echo.Group) {
	orders := g.Group("/user/orders", middleware.JWT())
	orders.POST("", h.create)
	orders.GET("", h.list)
}

func (h *OrderHandler) create(c *echo.Context) error {
	userID, err := common.UserIDFromContext(c)
	if err != nil {
		return mapAppError(err)
	}

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, invalidRequestBody)
	}

	req := dto.UploadOrderRequest{Number: strings.TrimSpace(string(body))}
	if err := req.Validate(); err != nil {
		return mapAppError(err)
	}

	if _, err := h.orders.Create(c.Request().Context(), models.Order{
		OrderNumber: req.Number,
		UserID:      userID,
	}); errors.Is(err, services.ErrOrderAlreadyUploaded) {
		return c.NoContent(http.StatusOK)
	} else if err != nil {
		return mapAppError(err)
	}

	return c.NoContent(http.StatusAccepted)
}

func (h *OrderHandler) list(c *echo.Context) error {
	userID, err := common.UserIDFromContext(c)
	if err != nil {
		return mapAppError(err)
	}

	orders, err := h.orders.List(c.Request().Context(), userID)
	if err != nil {
		return mapAppError(err)
	}
	if len(orders) == 0 {
		return c.NoContent(http.StatusNoContent)
	}

	response := make([]dto.OrderResponse, 0, len(orders))
	for _, order := range orders {
		item := dto.OrderResponse{
			Number:     order.OrderNumber,
			Status:     order.Status,
			UploadedAt: order.CreatedAt,
		}
		if order.Accrual > 0 {
			item.Accrual = &order.Accrual
		}
		response = append(response, item)
	}

	if len(response) == 0 {
		return c.NoContent(http.StatusNoContent)
	}

	return c.JSON(http.StatusOK, response)
}
