package aur

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gogrlx/snack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func withRPCServer(t *testing.T, handler http.HandlerFunc) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	oldBaseURL := rpcBaseURL
	rpcBaseURL = srv.URL
	t.Cleanup(func() {
		rpcBaseURL = oldBaseURL
	})
}

func TestRPCSearch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := rpcResponse{
			ResultCount: 2,
			Results: []rpcResult{
				{Name: "yay", Version: "12.5.7-1", Description: "AUR helper"},
				{Name: "yay-bin", Version: "12.5.7-1", Description: "AUR helper (binary)"},
			},
			Type:    "search",
			Version: 5,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	// Override the base URL for testing — we need to test the parsing
	// Since we can't easily override the const, test the JSON parsing directly
	var resp rpcResponse
	httpResp, err := http.Get(srv.URL)
	require.NoError(t, err)
	defer httpResp.Body.Close()
	require.NoError(t, json.NewDecoder(httpResp.Body).Decode(&resp))

	assert.Equal(t, 2, resp.ResultCount)
	assert.Equal(t, "yay", resp.Results[0].Name)
	assert.Equal(t, "12.5.7-1", resp.Results[0].Version)

	pkg := resp.Results[0].toPackage()
	assert.Equal(t, "yay", pkg.Name)
	assert.Equal(t, "12.5.7-1", pkg.Version)
	assert.Equal(t, "aur", pkg.Repository)
}

func TestRPCInfo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := rpcResponse{
			ResultCount: 1,
			Results: []rpcResult{
				{
					Name:        "yay",
					Version:     "12.5.7-1",
					Description: "Yet another yogurt",
					Depends:     []string{"pacman>6.1", "git"},
					MakeDepends: []string{"go>=1.24"},
				},
			},
			Type:    "multiinfo",
			Version: 5,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	var resp rpcResponse
	httpResp, err := http.Get(srv.URL)
	require.NoError(t, err)
	defer httpResp.Body.Close()
	require.NoError(t, json.NewDecoder(httpResp.Body).Decode(&resp))

	assert.Equal(t, 1, resp.ResultCount)
	r := resp.Results[0]
	assert.Equal(t, "yay", r.Name)
	assert.Equal(t, []string{"pacman>6.1", "git"}, r.Depends)
	assert.Equal(t, []string{"go>=1.24"}, r.MakeDepends)
}

func TestRPCError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := rpcResponse{
			Error:   "Incorrect request type specified.",
			Type:    "error",
			Version: 5,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	var resp rpcResponse
	httpResp, err := http.Get(srv.URL)
	require.NoError(t, err)
	defer httpResp.Body.Close()
	require.NoError(t, json.NewDecoder(httpResp.Body).Decode(&resp))

	assert.Equal(t, "Incorrect request type specified.", resp.Error)
}

func TestRPCInfoMulti(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := rpcResponse{
			ResultCount: 2,
			Results: []rpcResult{
				{Name: "yay", Version: "12.5.7-1"},
				{Name: "paru", Version: "2.0.4-1"},
			},
			Type:    "multiinfo",
			Version: 5,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	var resp rpcResponse
	httpResp, err := http.Get(srv.URL)
	require.NoError(t, err)
	defer httpResp.Body.Close()
	require.NoError(t, json.NewDecoder(httpResp.Body).Decode(&resp))

	assert.Equal(t, 2, resp.ResultCount)

	// Simulate rpcInfoMulti result building
	result := make(map[string]rpcResult, len(resp.Results))
	for _, r := range resp.Results {
		result[r.Name] = r
	}
	assert.Equal(t, "12.5.7-1", result["yay"].Version)
	assert.Equal(t, "2.0.4-1", result["paru"].Version)
}

func TestToPackage(t *testing.T) {
	r := rpcResult{
		Name:        "paru",
		Version:     "2.0.4-1",
		Description: "Feature packed AUR helper",
		URL:         "https://github.com/Morganamilo/paru",
	}
	pkg := r.toPackage()
	assert.Equal(t, "paru", pkg.Name)
	assert.Equal(t, "2.0.4-1", pkg.Version)
	assert.Equal(t, "Feature packed AUR helper", pkg.Description)
	assert.Equal(t, "aur", pkg.Repository)
	assert.False(t, pkg.Installed) // AUR search results aren't installed
}

func TestRPCSearchLive(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live AUR API test")
	}
	pkgs, err := rpcSearch(context.Background(), "yay")
	require.NoError(t, err)
	assert.NotEmpty(t, pkgs)

	found := false
	for _, p := range pkgs {
		if p.Name == "yay" {
			found = true
			assert.NotEmpty(t, p.Version)
			assert.Equal(t, "aur", p.Repository)
			break
		}
	}
	assert.True(t, found, "expected to find 'yay' in AUR search results")
}

func TestRPCInfoLive(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live AUR API test")
	}
	pkg, err := rpcInfo(context.Background(), "yay")
	require.NoError(t, err)
	assert.Equal(t, "yay", pkg.Name)
	assert.NotEmpty(t, pkg.Version)
}

func TestRPCInfoNotFound(t *testing.T) {
	withRPCServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/info", r.URL.Path)
		assert.Equal(t, "this-package-definitely-does-not-exist-12345", r.URL.Query().Get("arg[]"))

		resp := rpcResponse{
			ResultCount: 0,
			Type:        "multiinfo",
			Version:     5,
		}
		require.NoError(t, json.NewEncoder(w).Encode(resp))
	})

	_, err := rpcInfo(context.Background(), "this-package-definitely-does-not-exist-12345")
	require.Error(t, err)
	assert.True(t, errors.Is(err, snack.ErrNotFound))
}

func TestRPCSearchUsesEscapedPath(t *testing.T) {
	withRPCServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/search/foo%2Fbar", r.URL.EscapedPath())

		resp := rpcResponse{
			ResultCount: 1,
			Results: []rpcResult{
				{Name: "foo/bar", Version: "1.0.0-1", Description: "escaped name"},
			},
			Type:    "search",
			Version: 5,
		}
		require.NoError(t, json.NewEncoder(w).Encode(resp))
	})

	pkgs, err := rpcSearch(context.Background(), "foo/bar")
	require.NoError(t, err)
	require.Len(t, pkgs, 1)
	assert.Equal(t, snack.Package{
		Name:        "foo/bar",
		Version:     "1.0.0-1",
		Description: "escaped name",
		Repository:  "aur",
	}, pkgs[0])
}

func TestRPCInfoUsesEscapedQueryArg(t *testing.T) {
	const pkgName = "name with spaces"

	withRPCServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/info", r.URL.Path)
		assert.Equal(t, pkgName, r.URL.Query().Get("arg[]"))

		resp := rpcResponse{
			ResultCount: 1,
			Results: []rpcResult{
				{Name: pkgName, Version: "2.0.0-1", Description: "query package"},
			},
			Type:    "multiinfo",
			Version: 5,
		}
		require.NoError(t, json.NewEncoder(w).Encode(resp))
	})

	pkg, err := rpcInfo(context.Background(), pkgName)
	require.NoError(t, err)
	require.NotNil(t, pkg)
	assert.Equal(t, pkgName, pkg.Name)
	assert.Equal(t, "2.0.0-1", pkg.Version)
	assert.Equal(t, "aur", pkg.Repository)
}

func TestRPCInfoMultiBuildsRepeatedArgs(t *testing.T) {
	wantArgs := []string{"yay-bin", "pkg with spaces"}

	withRPCServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/info", r.URL.Path)
		assert.Equal(t, wantArgs, r.URL.Query()["arg[]"])

		resp := rpcResponse{
			ResultCount: 2,
			Results: []rpcResult{
				{Name: "yay-bin", Version: "12.5.7-1"},
				{Name: "pkg with spaces", Version: "1.2.3-1"},
			},
			Type:    "multiinfo",
			Version: 5,
		}
		require.NoError(t, json.NewEncoder(w).Encode(resp))
	})

	result, err := rpcInfoMulti(context.Background(), wantArgs)
	require.NoError(t, err)
	assert.Equal(t, "12.5.7-1", result["yay-bin"].Version)
	assert.Equal(t, "1.2.3-1", result["pkg with spaces"].Version)
}

func TestRPCInfoMultiEmptyDoesNotRequest(t *testing.T) {
	result, err := rpcInfoMulti(context.Background(), nil)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestRPCGetHandlesResponseFailures(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		wantErr string
	}{
		{
			name: "non-OK status",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				http.Error(w, "unavailable", http.StatusServiceUnavailable)
			},
			wantErr: "HTTP 503",
		},
		{
			name: "invalid JSON",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				_, err := w.Write([]byte("{"))
				require.NoError(t, err)
			},
			wantErr: "parsing response",
		},
		{
			name: "AUR error field",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				resp := rpcResponse{Error: "Incorrect request type specified.", Type: "error", Version: 5}
				require.NoError(t, json.NewEncoder(w).Encode(resp))
			},
			wantErr: "Incorrect request type specified.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(tt.handler)
			t.Cleanup(srv.Close)

			_, err := rpcGet(context.Background(), srv.URL)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestRPCGetRejectsInvalidEndpoint(t *testing.T) {
	_, err := rpcGet(context.Background(), string([]byte{0x7f}))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "aur rpc")
}
