package registry_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"sonarbridge-go/internal/plugin/registry"
)

func fakeLocalRegistry(t *testing.T, binaryContent []byte) *httptest.Server {
	t.Helper()

	assetName := fmt.Sprintf("ci-bridge-vcs-gitlab_%s_%s", currentGOOS(), currentGOARCH())
	sum := sha256.Sum256(binaryContent)
	checksums := hex.EncodeToString(sum[:]) + "  " + assetName + "\n"

	// Declared before the mux so the handler closures below can
	// reference srv.URL — they only run at request time, by which
	// point srv is assigned (Go closures capture by reference).
	var srv *httptest.Server

	mux := http.NewServeMux()

	releaseHandler := func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Accept"); got != "application/vnd.github+json" {
			t.Errorf("unexpected Accept header: %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"tag_name": "v0.1.0",
			"assets": []map[string]any{
				{
					"name":          "vcs-gitlab-plugin",
					"path":          "bin.local/assets/binaries/vcs-gitlab-plugin_linux_arm_6/vcs-gitlab-plugin",
					"goos":          "linux",
					"goarch":        "arm",
					"goarm":         "6",
					"target":        "linux_arm_6",
					"internal_type": 4,
					"type":          "Binary",
				},
				{
					"name":          "sonarbridge-go-sonarbridge-go_Windows_i386.zip.sha256_checksums.txt",
					"path":          "bin.local/assets/binaries/sonarbridge-go-sonarbridge-go_Windows_i386.zip.sha256_checksums.txt",
					"internal_type": 12,
					"type":          "Checksum",
				},
			},
		})
	}

	mux.HandleFunc("/repos/nivekalara237/ci-bridge-vcs-gitlab/releases/latest", releaseHandler)
	mux.HandleFunc("/repos/nivekalara237/ci-bridge-vcs-gitlab/releases/tags/v0.1.0", releaseHandler)
	mux.HandleFunc("/repos/nivekalara237/ci-bridge-vcs-gitlab/releases/tags/v9.9.9", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	mux.HandleFunc("/asset/1", func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Accept"); got != "application/octet-stream" {
			t.Errorf("unexpected Accept header on asset download: %q", got)
		}
		w.Write(binaryContent)
	})
	mux.HandleFunc("/asset/2", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(checksums))
	})

	srv = httptest.NewServer(mux)
	return srv
}

func TestGitHubRegistry_ResolveLatestAndFetch(t *testing.T) {
	binaryContent := []byte("#!/bin/sh\necho pretend-gitlab-plugin-binary\n")
	srv := fakeLocalRegistry(t, binaryContent)
	defer srv.Close()

	reg := registry.NewLocalRegistry(srv.URL)

	artifact, err := reg.Resolve(context.Background(), "gitlab", "latest")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if artifact.Version != "v0.1.0" {
		t.Fatalf("unexpected version: %+v", artifact)
	}
	wantSum := sha256.Sum256(binaryContent)
	if artifact.SHA256 != hex.EncodeToString(wantSum[:]) {
		t.Fatalf("checksum mismatch: got %s, want %x", artifact.SHA256, wantSum)
	}

	destDir := t.TempDir()
	path, err := reg.Fetch(context.Background(), artifact, destDir)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(binaryContent) {
		t.Fatalf("downloaded content mismatch")
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&0o111 == 0 {
		t.Fatalf("expected the fetched binary to be executable, mode=%v", info.Mode())
	}
}

func TestGitHubRegistry_ResolveExactTag(t *testing.T) {
	binaryContent := []byte("v0.1.0 content")
	srv := fakeLocalRegistry(t, binaryContent)
	defer srv.Close()

	reg := registry.NewLocalRegistry(srv.URL)

	artifact, err := reg.Resolve(context.Background(), "gitlab", "v0.1.0")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if artifact.Version != "v0.1.0" {
		t.Fatalf("unexpected version: %+v", artifact)
	}
}

func TestGitHubRegistry_UnknownVersionIsAClearError(t *testing.T) {
	srv := fakeLocalRegistry(t, []byte("x"))
	defer srv.Close()

	reg := registry.NewLocalRegistry(srv.URL)

	_, err := reg.Resolve(context.Background(), "gitlab", "v9.9.9")
	if err == nil {
		t.Fatal("expected an error for a release that doesn't exist")
	}
	if !strings.Contains(err.Error(), "no release found") {
		t.Fatalf("expected a clear 'no release found' error, got: %v", err)
	}
}

func TestGitHubRegistry_TokenSentOnAPIRequests(t *testing.T) {
	var gotAuth string
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/nivekalara237/ci-bridge-vcs-gitlab/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusNotFound)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	reg := registry.NewLocalRegistry(srv.URL)
	reg.Resolve(context.Background(), "gitlab", "latest")

	if gotAuth != "Bearer my-secret-token" {
		t.Fatalf("expected the token on the release lookup, got %q", gotAuth)
	}
}

// currentGOOS/currentGOARCH avoid importing "runtime" under the name
// used by the package under test's own internal import, keeping this
// test file simple to read.
func currentGOOS() string   { return goos }
func currentGOARCH() string { return goarch }
