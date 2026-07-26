package aur

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gogrlx/snack"
)

var rpcBaseURL = "https://aur.archlinux.org/rpc/v5"

// rpcResponse is the top-level AUR RPC response.
type rpcResponse struct {
	ResultCount int         `json:"resultcount"`
	Results     []rpcResult `json:"results"`
	Type        string      `json:"type"`
	Error       string      `json:"error,omitempty"`
	Version     int         `json:"version"`
}

// rpcResult is a single package from the AUR RPC API.
type rpcResult struct {
	Name           string   `json:"Name"`
	Version        string   `json:"Version"`
	Description    string   `json:"Description"`
	URL            string   `json:"URL"`
	URLPath        string   `json:"URLPath"`
	PackageBase    string   `json:"PackageBase"`
	PackageBaseID  int      `json:"PackageBaseID"`
	NumVotes       int      `json:"NumVotes"`
	Popularity     float64  `json:"Popularity"`
	OutOfDate      *int64   `json:"OutOfDate"`
	Maintainer     string   `json:"Maintainer"`
	FirstSubmitted int64    `json:"FirstSubmitted"`
	LastModified   int64    `json:"LastModified"`
	Depends        []string `json:"Depends"`
	MakeDepends    []string `json:"MakeDepends"`
	OptDepends     []string `json:"OptDepends"`
	License        []string `json:"License"`
	Keywords       []string `json:"Keywords"`
}

func (r *rpcResult) toPackage() snack.Package {
	return snack.Package{
		Name:        r.Name,
		Version:     r.Version,
		Description: r.Description,
		Repository:  "aur",
	}
}

// rpcGet performs an AUR RPC API request.
func rpcGet(ctx context.Context, endpoint string) (*rpcResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("aur rpc: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("aur rpc: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("aur rpc: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("aur rpc: reading response: %w", err)
	}

	var result rpcResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("aur rpc: parsing response: %w", err)
	}

	if result.Error != "" {
		return nil, fmt.Errorf("aur rpc: %s", result.Error)
	}

	return &result, nil
}

// rpcSearch queries the AUR for packages matching the query string.
func rpcSearch(ctx context.Context, query string) ([]snack.Package, error) {
	endpoint := rpcBaseURL + "/search/" + url.PathEscape(query)
	resp, err := rpcGet(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	pkgs := make([]snack.Package, 0, len(resp.Results))
	for _, r := range resp.Results {
		pkgs = append(pkgs, r.toPackage())
	}
	return pkgs, nil
}

// rpcInfo returns info about a specific AUR package.
func rpcInfo(ctx context.Context, pkg string) (*snack.Package, error) {
	endpoint := rpcBaseURL + "/info?arg[]=" + url.QueryEscape(pkg)
	resp, err := rpcGet(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	if resp.ResultCount == 0 {
		return nil, fmt.Errorf("aur info %s: %w", pkg, snack.ErrNotFound)
	}
	p := resp.Results[0].toPackage()
	return &p, nil
}

// rpcInfoMulti returns info about multiple AUR packages in a single request.
func rpcInfoMulti(ctx context.Context, pkgs []string) (map[string]rpcResult, error) {
	if len(pkgs) == 0 {
		return nil, nil
	}
	params := make([]string, len(pkgs))
	for i, p := range pkgs {
		params[i] = "arg[]=" + url.QueryEscape(p)
	}
	endpoint := rpcBaseURL + "/info?" + strings.Join(params, "&")
	resp, err := rpcGet(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	result := make(map[string]rpcResult, len(resp.Results))
	for _, r := range resp.Results {
		result[r.Name] = r
	}
	return result, nil
}
