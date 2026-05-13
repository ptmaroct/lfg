package installer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRunVerifiedCurl_HappyPath(t *testing.T) {
	body := []byte("#!/bin/sh\necho hello\n")
	sum := sha256.Sum256(body)
	expected := hex.EncodeToString(sum[:])

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	out := make(chan Line, 32)
	cmdline := "curl -fsSL " + srv.URL + " | sh"
	if err := runVerifiedCurl(context.Background(), "test", cmdline, expected, out); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	close(out)
	var sawVerified bool
	for ln := range out {
		if strings.Contains(ln.Text, "sha256 verified") {
			sawVerified = true
		}
	}
	if !sawVerified {
		t.Errorf("expected sha256-verified meta line in output")
	}
}

func TestRunVerifiedCurl_SHAMismatch(t *testing.T) {
	body := []byte("not the file you were expecting")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	out := make(chan Line, 32)
	cmdline := "curl -fsSL " + srv.URL + " | bash"
	bogus := strings.Repeat("0", 64)
	err := runVerifiedCurl(context.Background(), "test", cmdline, bogus, out)
	if err == nil {
		t.Fatal("expected supply-chain check failure, got nil error")
	}
	if !strings.Contains(err.Error(), "supply-chain check failed") {
		t.Errorf("wrong error: %v", err)
	}
}

func TestRunVerifiedCurl_NoURL(t *testing.T) {
	out := make(chan Line, 4)
	err := runVerifiedCurl(context.Background(), "test", "echo hi", "abc", out)
	if err == nil {
		t.Fatal("expected error when no URL in cmdline")
	}
	if !strings.Contains(err.Error(), "no https URL") {
		t.Errorf("wrong error: %v", err)
	}
}

func TestRunVerifiedCurl_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusInternalServerError)
	}))
	defer srv.Close()

	out := make(chan Line, 4)
	err := runVerifiedCurl(context.Background(), "test",
		"curl "+srv.URL+" | sh", strings.Repeat("0", 64), out)
	if err == nil {
		t.Fatal("expected fetch error")
	}
}
