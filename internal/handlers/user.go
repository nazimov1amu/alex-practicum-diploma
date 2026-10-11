package handlers

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/common"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/dto"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/middleware"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/services"
)

const invalidRequestBody = "invalid_request_body"

type UserHandler struct {
	users *services.UserService
}

func NewUserHandler(users *services.UserService) *UserHandler {
	return &UserHandler{users: users}
}

func (h *UserHandler) Register(g *echo.Group) {
	users := g.Group("/user")
	users.POST("/register", h.register)
	users.POST("/login", h.login)
	users.POST("/refresh", h.refresh)
	users.GET("/balance", h.balance, middleware.JWT())
}

func (h *UserHandler) register(c *echo.Context) error {
	var req dto.Credentials
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, invalidRequestBody)
	}
	if err := req.Validate(); err != nil {
		return mapAppError(err)
	}

	tokens, err := h.users.Register(c.Request().Context(), req.Login, req.Password)
	if err != nil {
		return mapAppError(err)
	}
	return writeTokens(c, tokens)
}

func (h *UserHandler) login(c *echo.Context) error {
	var req dto.Credentials
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, invalidRequestBody)
	}
	if err := req.Validate(); err != nil {
		return mapAppError(err)
	}

	tokens, err := h.users.Login(c.Request().Context(), req.Login, req.Password)
	if err != nil {
		return mapAppError(err)
	}
	return writeTokens(c, tokens)
}

func (h *UserHandler) refresh(c *echo.Context) error {
	var req dto.RefreshRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, invalidRequestBody)
	}
	if err := req.Validate(); err != nil {
		return mapAppError(err)
	}

	tokens, err := h.users.Refresh(c.Request().Context(), req.RefreshToken)
	if err != nil {
		return mapAppError(err)
	}
	return writeTokens(c, tokens)
}

func (h *UserHandler) balance(c *echo.Context) error {
	userID, err := common.UserIDFromContext(c)
	if err != nil {
		return mapAppError(err)
	}

	balance, err := h.users.GetBalance(c.Request().Context(), userID)
	if err != nil {
		return mapAppError(err)
	}

	return c.JSON(http.StatusOK, dto.BalanceResponse{
		Current:   balance.Current,
		Withdrawn: balance.Withdrawn,
	})
}

func writeTokens(c *echo.Context, tokens services.TokenPair) error {
	c.Response().Header().Set(echo.HeaderAuthorization, "Bearer "+tokens.AccessToken)
	return c.JSON(http.StatusOK, dto.TokenPairResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}
