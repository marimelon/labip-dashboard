package main

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	cfg    *Config
	pve    *PVEClient
	nebura *NeburaCliend
}

func NewHandler(cfg *Config) *Handler {
	handler := &Handler{cfg: cfg}
	if cfg.PVEEndpoint != "" {
		handler.pve = NewPVECliend(cfg.PVEEndpoint, cfg.PVEApiUser, cfg.PVEApiKey)
	}
	if cfg.NeburaEndpoint != "" {
		handler.nebura = NewNeburaCliend(cfg.NeburaUser, cfg.NeburaPass, cfg.NeburaEndpoint)
	}
	return handler
}

func (h *Handler) IndexHandler(c echo.Context) error {
	ipInfoList, err := h.get_ip_info()
	if err != nil {
		return err
	}

	now := time.Now()
	templateData := map[string]interface{}{
		"FetchDate": now.Format("2006-01-02 15:04:05"),
		"IpList":    ipInfoList,
	}
	return c.Render(http.StatusOK, "index", templateData)
}

func (h *Handler) IpHandler(c echo.Context) error {
	ipInfoList, err := h.get_ip_info()
	if err != nil {
		return err
	}

	now := time.Now()
	data := map[string]interface{}{
		"FetchDate": now.Format("2006-01-02 15:04:05"),
		"IpList":    ipInfoList,
	}
	c.JSON(200, data)
	return nil
}
