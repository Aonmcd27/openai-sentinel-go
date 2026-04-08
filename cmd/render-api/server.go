package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	sentinel "openai-sentinel-go"
)

type buildFn func(ctx context.Context, session *sentinel.Session, flow, referer, turnstileToken string) (sentinel.Token, error)

type server struct {
	apiBearerToken  string
	clientTimeoutMs int
	buildToken      buildFn
}

type buildRequest struct {
	Flow           string         `json:"flow"`
	Referer        string         `json:"referer"`
	TurnstileToken string         `json:"turnstileToken"`
	Session        sessionPayload `json:"session"`
}

type sessionPayload struct {
	DeviceID            string         `json:"deviceId"`
	UserAgent           string         `json:"userAgent"`
	ScreenWidth         int            `json:"screenWidth"`
	ScreenHeight        int            `json:"screenHeight"`
	HeapLimit           int64          `json:"heapLimit"`
	HardwareConcurrency int            `json:"hardwareConcurrency"`
	Language            string         `json:"language"`
	LanguagesJoin       string         `json:"languagesJoin"`
	Persona             personaPayload `json:"persona"`
}

type personaPayload struct {
	Platform              string  `json:"platform"`
	Vendor                string  `json:"vendor"`
	TimezoneOffsetMin     int     `json:"timezoneOffsetMin"`
	SessionID             string  `json:"sessionId"`
	TimeOrigin            float64 `json:"timeOrigin"`
	WindowFlags           [7]int  `json:"windowFlags"`
	WindowFlagsSet        bool    `json:"windowFlagsSet"`
	EntropyA              float64 `json:"entropyA"`
	EntropyB              float64 `json:"entropyB"`
	DateString            string  `json:"dateString"`
	RequirementsScriptURL string  `json:"requirementsScriptURL"`
	NavigatorProbe        string  `json:"navigatorProbe"`
	DocumentProbe         string  `json:"documentProbe"`
	WindowProbe           string  `json:"windowProbe"`
	PerformanceNow        float64 `json:"performanceNow"`
	RequirementsElapsed   float64 `json:"requirementsElapsed"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (s server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("POST /build", s.handleBuild)
	return mux
}

func (s server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s server) handleBuild(w http.ResponseWriter, r *http.Request) {
	if err := s.authorize(r); err != nil {
		writeError(w, http.StatusUnauthorized, err)
		return
	}
	var req buildRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if s.buildToken == nil {
		writeError(w, http.StatusInternalServerError, errors.New("build handler not configured"))
		return
	}
	token, err := s.buildToken(r.Context(), req.Session.toSession(s.clientTimeoutMs), req.Flow, req.Referer, req.TurnstileToken)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, token)
}

func (s server) authorize(r *http.Request) error {
	if strings.TrimSpace(s.apiBearerToken) == "" {
		return nil
	}
	header := strings.TrimSpace(r.Header.Get("Authorization"))
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return errors.New("missing bearer token")
	}
	if strings.TrimSpace(strings.TrimPrefix(header, prefix)) != s.apiBearerToken {
		return errors.New("invalid bearer token")
	}
	return nil
}

func (s sessionPayload) toSession(timeoutMs int) *sentinel.Session {
	timeout := 10 * time.Second
	if timeoutMs > 0 {
		timeout = time.Duration(timeoutMs) * time.Millisecond
	}
	return &sentinel.Session{
		Client:              &http.Client{Timeout: timeout},
		DeviceID:            s.DeviceID,
		UserAgent:           s.UserAgent,
		ScreenWidth:         s.ScreenWidth,
		ScreenHeight:        s.ScreenHeight,
		HeapLimit:           s.HeapLimit,
		HardwareConcurrency: s.HardwareConcurrency,
		Language:            s.Language,
		LanguagesJoin:       s.LanguagesJoin,
		Persona: sentinel.Persona{
			Platform:              s.Persona.Platform,
			Vendor:                s.Persona.Vendor,
			TimezoneOffsetMin:     s.Persona.TimezoneOffsetMin,
			SessionID:             s.Persona.SessionID,
			TimeOrigin:            s.Persona.TimeOrigin,
			WindowFlags:           s.Persona.WindowFlags,
			WindowFlagsSet:        s.Persona.WindowFlagsSet,
			EntropyA:              s.Persona.EntropyA,
			EntropyB:              s.Persona.EntropyB,
			DateString:            s.Persona.DateString,
			RequirementsScriptURL: s.Persona.RequirementsScriptURL,
			NavigatorProbe:        s.Persona.NavigatorProbe,
			DocumentProbe:         s.Persona.DocumentProbe,
			WindowProbe:           s.Persona.WindowProbe,
			PerformanceNow:        s.Persona.PerformanceNow,
			RequirementsElapsed:   s.Persona.RequirementsElapsed,
		},
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, errorResponse{Error: err.Error()})
}
