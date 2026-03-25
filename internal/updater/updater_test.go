package updater

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDownloadFile_Success(t *testing.T) {
	const want = "fake binary content"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(want))
	}))
	defer srv.Close()

	out := filepath.Join(t.TempDir(), "binary")
	if err := downloadFile(context.Background(), srv.URL, out); err != nil {
		t.Fatalf("downloadFile: %v", err)
	}

	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("reading output: %v", err)
	}
	if string(got) != want {
		t.Errorf("content = %q, want %q", got, want)
	}
}

func TestDownloadFile_BadStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	err := downloadFile(context.Background(), srv.URL, filepath.Join(t.TempDir(), "out"))
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
}

func TestDownloadFile_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := downloadFile(ctx, srv.URL, filepath.Join(t.TempDir(), "out"))
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}
