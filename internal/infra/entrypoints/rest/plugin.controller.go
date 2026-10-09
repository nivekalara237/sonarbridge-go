package rest

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sonarbridge-go/configs"
	"sonarbridge-go/internal/infra/entrypoints/dto"
	"sonarbridge-go/internal/infra/entrypoints/rest/httpx"
	"sonarbridge-go/internal/plugin/manager"
	"sonarbridge-go/pkg/arrays"
	"sonarbridge-go/pkg/httpclient"
	"strings"
	"time"
)

type PluginController struct {
	manager    *manager.Manager
	restclient *httpclient.Client
}

func NewPluginController(mgr *manager.Manager) *PluginController {
	var encoded string
	if configs.AppConfig.Plugins.RemoteRegistry.Enabled && (configs.AppConfig.Plugins.RemoteRegistry.Password != "" ||
		configs.AppConfig.Plugins.RemoteRegistry.Username != "") {
		encoded = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s",
			configs.AppConfig.Plugins.RemoteRegistry.Username,
			configs.AppConfig.Plugins.RemoteRegistry.Password)))
	}
	return &PluginController{
		manager: mgr,
		restclient: httpclient.NewClientHttp(
			httpclient.WithBaseURL(configs.AppConfig.Plugins.RemoteRegistry.Url),
			httpclient.WithDefaultHeader("Authorization", "Basic "+encoded),
			httpclient.WithCache(httpclient.NewInMemoryCache()),
		),
	}
}

type AssetFile struct {
	Metadata  string `json:"metadata_json"`
	Artifacts string `json:"artifacts_json"`
	Checksums string `json:"checksums_text"`
	Provider  string `json:"provider"`
}

func (c *PluginController) GetInstalledPlugins(w http.ResponseWriter, req *http.Request) error {
	_, cancel := context.WithTimeout(req.Context(), 10*time.Second)
	defer cancel()

	if req.Method != http.MethodGet {
		return httpx.ErrMethodNotAllowed
	}

	var installedStates []dto.InstalledPluginStateResponse

	for _, status := range c.manager.List() {
		installedStates = append(installedStates, dto.InstalledPluginStateResponse{
			Name:            status.Name,
			State:           string(status.State),
			PluginVersion:   status.Info.Version,
			ProtocolVersion: status.Info.ProtocolVersion,
			Type:            status.Info.PluginType,
			Capabilities:    status.Info.Capabilities,
		})
	}
	httpx.WriteJSON(w, http.StatusOK, dto.OkListResponse(installedStates))
	return nil
}

func (c *PluginController) GetAvailablePlugins(w http.ResponseWriter, req *http.Request) error {
	ctx, cancel := context.WithTimeout(req.Context(), 10*time.Second)
	defer cancel()

	if req.Method != http.MethodGet {
		return httpx.ErrMethodNotAllowed
	}

	resp, err := c.restclient.Get(ctx, "/artifacts/all", nil)
	if err != nil || !httpclient.IsSuccess(resp) {
		return httpx.ErrInternal.Wrap(errors.New("unable to get artifacts"))
	}

	var assetStr []AssetFile
	if err := json.NewDecoder(resp.Body).Decode(&assetStr); err != nil {
		return httpx.ErrInternal
	}

	osQuery := req.URL.Query().Get("os")
	archQuery := req.URL.Query().Get("arch")

	assetGrouped := arrays.GroupBy[AssetFile, string](assetStr, func(file AssetFile) string {
		return file.Provider
	})

	var response []dto.AssetArtifactResponse

	for _, v := range assetGrouped {
		for _, item := range v {
			a := &dto.AssetArtifactResponse{}
			a.Provider = item.Provider

			allArtifacts, err := parseArtifacts(item.Artifacts)
			if err != nil {
				return httpx.ErrInternal.Wrap(err)
			}

			a.Artifacts = filterArtifacts(allArtifacts, osQuery, archQuery)

			if er := json.Unmarshal([]byte(item.Artifacts), &allArtifacts); er != nil {
				return httpx.ErrInternal.Wrap(er)
			}

			if er := json.Unmarshal([]byte(item.Metadata), &a.Metadata); er != nil {
				return httpx.ErrInternal.Wrap(er)
			}
			checksums := parseChecksums(item.Checksums)
			a.Checksums = filterChecksums(checksums, strings.ToLower(osQuery), strings.ToLower(archQuery))
			response = append(response, *a)
		}
	}

	httpx.WriteJSON(w, http.StatusOK, dto.OkListResponse(response))
	return nil
}

func parseChecksums(raw string) []dto.Checksum {
	if raw != "" {
		sc := bufio.NewScanner(strings.NewReader(raw))
		line := 0
		var resp []dto.Checksum
		for sc.Scan() {
			line++
			l := strings.TrimSpace(sc.Text())
			if l == "" {
				continue
			}

			fd := strings.Fields(l)
			if len(fd) > 1 {
				resp = append(resp,
					dto.Checksum{BinaryName: strings.Join(fd[1:], " "), Hash: fd[0]})
			}
		}

		return resp
	}

	return nil
}

func parseArtifacts(raw string) ([]map[string]any, error) {
	if raw == "" {
		return nil, nil
	}
	var all []map[string]any
	if err := json.Unmarshal([]byte(raw), &all); err != nil {
		return nil, err
	}

	var withoutMetadata []map[string]any
	for _, item := range all {
		if item["type"] != "Metadata" {
			withoutMetadata = append(withoutMetadata, item)
		}
	}

	return withoutMetadata, nil
}

func filterChecksums(checksums []dto.Checksum, osQuery, archQuery string) []dto.Checksum {
	if osQuery == "" && archQuery == "" {
		return checksums
	}

	if archQuery == "amd64" {
		archQuery = "x86_64"
	}

	out := make([]dto.Checksum, 0, len(checksums))
	for _, cs := range checksums {
		name := strings.ToLower(cs.BinaryName)
		if osQuery != "" && !strings.Contains(name, osQuery) {
			continue
		}
		if archQuery != "" && !strings.Contains(name, archQuery) {
			continue
		}
		out = append(out, cs)
	}
	return out
}
func filterArtifacts(arts []map[string]any, osQuery, archQuery string) []map[string]any {
	if osQuery == "" && archQuery == "" {
		return arts
	}

	out := make([]map[string]any, 0, len(arts))
	for _, art := range arts {
		if osQuery != "" && art["goos"] != osQuery {
			continue
		}
		if archQuery != "" && art["goarch"] != archQuery {
			continue
		}
		out = append(out, art)
	}
	return out
}
