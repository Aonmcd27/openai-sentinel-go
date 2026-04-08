package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	sentinel "openai-sentinel-go"
)

func TestHealthzReturnsOK(t *testing.T) {
	srv := server{}
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	srv.routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if got := rec.Body.String(); got != "{\"ok\":true}\n" {
		t.Fatalf("unexpected body: %q", got)
	}
}

func TestBuildRequiresBearerTokenWhenConfigured(t *testing.T) {
	srv := server{
		apiBearerToken: "secret",
		buildToken: func(ctx context.Context, session *sentinel.Session, flow, referer, turnstileToken string) (sentinel.Token, error) {
			t.Fatal("builder should not be called without auth")
			return sentinel.Token{}, nil
		},
	}
	body := []byte(`{"flow":"username_password_create","session":{"deviceId":"dev","userAgent":"ua"}}`)
	req := httptest.NewRequest(http.MethodPost, "/build", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestBuildRejectsInvalidJSON(t *testing.T) {
	srv := server{}
	req := httptest.NewRequest(http.MethodPost, "/build", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestBuildSuccessReturnsToken(t *testing.T) {
	var gotFlow, gotReferer, gotTurnstile string
	var gotSession *sentinel.Session
	srv := server{
		apiBearerToken:  "secret",
		clientTimeoutMs: 1234,
		buildToken: func(ctx context.Context, session *sentinel.Session, flow, referer, turnstileToken string) (sentinel.Token, error) {
			gotSession = session
			gotFlow = flow
			gotReferer = referer
			gotTurnstile = turnstileToken
			return sentinel.Token{
				P:    "p-token",
				T:    "t-token",
				C:    "c-token",
				ID:   "dev-1",
				Flow: flow,
			}, nil
		},
	}
	reqBody := buildRequest{
		Flow:           "username_password_create",
		Referer:        "https://auth.openai.com/create-account/password",
		TurnstileToken: "turnstile",
		Session: sessionPayload{
			DeviceID:            "dev-1",
			UserAgent:           "Mozilla/5.0",
			ScreenWidth:         1920,
			ScreenHeight:        1080,
			HeapLimit:           4294705152,
			HardwareConcurrency: 8,
			Language:            "zh-CN",
			LanguagesJoin:       "zh-CN,en-US",
			Persona: personaPayload{
				Platform:   "Win32",
				Vendor:     "Google Inc.",
				SessionID:  "session-1",
				TimeOrigin: 1775190798250,
			},
		},
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/build", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer secret")
	rec := httptest.NewRecorder()

	srv.routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if gotFlow != reqBody.Flow || gotReferer != reqBody.Referer || gotTurnstile != reqBody.TurnstileToken {
		t.Fatalf("unexpected builder args: flow=%q referer=%q turnstile=%q", gotFlow, gotReferer, gotTurnstile)
	}
	if gotSession == nil {
		t.Fatal("expected session")
	}
	if gotSession.DeviceID != "dev-1" || gotSession.UserAgent != "Mozilla/5.0" {
		t.Fatalf("unexpected session basics: %+v", gotSession)
	}
	if gotSession.Client == nil {
		t.Fatal("expected http client on session")
	}
	var token sentinel.Token
	if err := json.Unmarshal(rec.Body.Bytes(), &token); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if token.P != "p-token" || token.T != "t-token" || token.C != "c-token" {
		t.Fatalf("unexpected token: %+v", token)
	}
}
