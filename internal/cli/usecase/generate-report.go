package usecase

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sonarbridge-go/internal/logging"
	"sonarbridge-go/pkg/httpclient"
	string2 "sonarbridge-go/pkg/string"
	"time"
)

type GenerateSonarReportRequest struct {
	TaskId       string `json:"taskId"`
	SonarProject struct {
		Key string `json:"key"`
	} `json:"sonarProject"`
	Gitlab struct {
		ProjectId  string `json:"projectId"`
		CiToken    string `json:"ciToken"`
		BranchName string `json:"branchName"`
		BranchUrl  string `json:"branchUrl"`
		CommitSha  string `json:"commitSha"`
	} `json:"gitlab"`
	MergeRequest struct {
		IID int `json:"iid"`
	} `json:"mergeRequest"`
}

func (g *GenerateSonarReportUseCase) ExecuteSonarRequest(
	request GenerateSonarReportRequest,
	sharedKey string,
) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ts := time.Now().Unix()

	sig := computeSignature("/webhook/sonar", string2.ToJSON(request), sharedKey, ts)

	resp, err := g.client.Post(ctx, "/webhook/sonar", &httpclient.RequestOptions{
		Headers: []httpclient.OptionParamItem{
			httpclient.NewRequestOption("ContentType", "application/json"),
			httpclient.NewRequestOption("x-sonar-webhook-Timestamp", ts),
			httpclient.NewRequestOption("x-sonar-webhook-hmac-sign", sig),
		},
		Body: request,
	})

	if err != nil {
		logging.Error("échec de communication avec le server", "error", err)
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logging.Error("échec d'encodage de la response http", "error", err)
		return err
	}

	var o map[string]any
	err = json.Unmarshal(bodyBytes, &o)
	if err != nil {
		logging.Error("échec d'encodage de la response http", "error", err)
		return err
	}

	if !o["Mergeable"].(bool) {
		return fmt.Errorf("the project is not mergeable, because the sonar analysis failed. QG STATUS=%s", o["QualityGateStatus"])
	}

	return err
}

type GenerateSonarReportUseCase struct {
	client  *httpclient.Client
	baseURL string
}

func NewGenerateSonarReport(baseUrl string, opts map[string]any) *GenerateSonarReportUseCase {
	var tlsOpt httpclient.Option
	if certFile, ok := opts["server-cert-file"]; ok {
		tlsOpt = httpclient.WithTLSCACertFile(certFile.(string))
	}
	client := httpclient.NewClientHttp(
		httpclient.WithBaseURL(baseUrl),
		httpclient.WithTimeout(60*time.Second),
		httpclient.WithRetry(3, 1*time.Second),
		tlsOpt,
	)

	return &GenerateSonarReportUseCase{
		client:  client,
		baseURL: baseUrl,
	}
}

func computeSignature(path, body, sharedSecret string, timestamp int64) string {
	return signRequest("POST", path, body, timestamp, sharedSecret)
}

func signRequest(method, path, body string, timestamp int64, sharedSecret string) string {
	mac := hmac.New(sha256.New, []byte(sharedSecret))
	minifiedBody, err := minifyJson(body)
	if err != nil {
		return ""
	}
	payload := fmt.Sprintf("%s\n%s\n%s\n%d", method, path, minifiedBody, timestamp)
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func minifyJson(jsonStr string) (string, error) {
	var buf bytes.Buffer
	if err := json.Compact(&buf, []byte(jsonStr)); err != nil {
		return "", err
	}
	return buf.String(), nil
}
