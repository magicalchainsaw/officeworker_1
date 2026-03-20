package handler

import (
	"errors"
	"net/http"
	"officeworker/internal/pkg/response"
	"officeworker/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SessionHandler struct {
	sessionService *service.SessionService
}

func NewSessionHandler(sessionService *service.SessionService) *SessionHandler {
	return &SessionHandler{sessionService: sessionService}
}

func (h *SessionHandler) Create(c *gin.Context) {
	userID, err := currentUserID(c)
	if err != nil {
		response.AbortWithError(c, http.StatusUnauthorized, err.Error())
		return
	}

	var req service.CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err.Error())
		return
	}

	resp, err := h.sessionService.Create(userID, &req)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, resp)
}

func (h *SessionHandler) List(c *gin.Context) {
	userID, err := currentUserID(c)
	if err != nil {
		response.AbortWithError(c, http.StatusUnauthorized, err.Error())
		return
	}

	resp, err := h.sessionService.List(userID)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, resp)
}

func (h *SessionHandler) ListActive(c *gin.Context) {
	userID, err := currentUserID(c)
	if err != nil {
		response.AbortWithError(c, http.StatusUnauthorized, err.Error())
		return
	}

	resp, err := h.sessionService.ListActive(userID)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, resp)
}

func (h *SessionHandler) Get(c *gin.Context) {
	userID, err := currentUserID(c)
	if err != nil {
		response.AbortWithError(c, http.StatusUnauthorized, err.Error())
		return
	}

	sessionID, err := parseSessionID(c.Param("id"))
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	resp, err := h.sessionService.Get(userID, sessionID)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, resp)
}

func (h *SessionHandler) Delete(c *gin.Context) {
	userID, err := currentUserID(c)
	if err != nil {
		response.AbortWithError(c, http.StatusUnauthorized, err.Error())
		return
	}

	sessionID, err := parseSessionID(c.Param("id"))
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	if err := h.sessionService.Delete(userID, sessionID); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *SessionHandler) Update(c *gin.Context) {
	userID, err := currentUserID(c)
	if err != nil {
		response.AbortWithError(c, http.StatusUnauthorized, err.Error())
		return
	}

	sessionID, err := parseSessionID(c.Param("id"))
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	var req service.UpdateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err.Error())
		return
	}

	resp, err := h.sessionService.Update(userID, sessionID, &req)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, resp)
}

func (h *SessionHandler) Activate(c *gin.Context) {
	userID, err := currentUserID(c)
	if err != nil {
		response.AbortWithError(c, http.StatusUnauthorized, err.Error())
		return
	}

	sessionID, err := parseSessionID(c.Param("id"))
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	resp, err := h.sessionService.Activate(userID, sessionID)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, resp)
}

func (h *SessionHandler) Deactivate(c *gin.Context) {
	userID, err := currentUserID(c)
	if err != nil {
		response.AbortWithError(c, http.StatusUnauthorized, err.Error())
		return
	}

	sessionID, err := parseSessionID(c.Param("id"))
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	resp, err := h.sessionService.Deactivate(userID, sessionID)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, resp)
}

func (h *SessionHandler) SendMessage(c *gin.Context) {
	userID, err := currentUserID(c)
	if err != nil {
		response.AbortWithError(c, http.StatusUnauthorized, err.Error())
		return
	}

	sessionID, err := parseSessionID(c.Param("id"))
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	var req service.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err.Error())
		return
	}

	resp, err := h.sessionService.SendMessage(userID, sessionID, &req)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, resp)
}

func currentUserID(c *gin.Context) (uint, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, errors.New("user not authenticated")
	}

	id, ok := userID.(uint)
	if !ok {
		return 0, errors.New("invalid user id")
	}

	return id, nil
}

func parseSessionID(raw string) (uint, error) {
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, errors.New("invalid session id")
	}

	return uint(id), nil
}
