package fhir

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-resty/resty/v2"
	models "github.com/samply/golang-fhir-models/fhir-models/fhir"
	"log/slog"
	"pseudonymous/config"
	"time"
)

type PsnClient struct {
	rest   *resty.Client
	config config.Pseudonymizer
}

func NewClient(cfg config.Pseudonymizer) *PsnClient {
	client := resty.New().
		SetLogger(config.DefaultLogger()).
		SetRetryCount(cfg.Retry.Count).
		SetTimeout(time.Duration(cfg.Retry.Timeout) * time.Second).
		SetRetryWaitTime(time.Duration(cfg.Retry.Wait) * time.Second).
		SetRetryMaxWaitTime(time.Duration(cfg.Retry.MaxWait) * time.Second)

	// TODO
	// if cfg.Auth != nil {
	// 	 client = client.SetBasicAuth(cfg.Auth.User, cfg.Auth.Password)
	// }

	return &PsnClient{rest: client, config: cfg}
}

func (c *PsnClient) Send(fhir []byte, domain string) ([]byte, error) {

	resource := json.RawMessage{}
	err := resource.UnmarshalJSON(fhir)
	if err != nil {
		slog.Error("Failed to unmarshal FHIR JSON payload", "error", err)
		return nil, err
	}

	params := models.Parameters{
		Parameter: []models.ParametersParameter{
			{
				Name: "settings",
				Part: []models.ParametersParameter{
					{
						Name: "domain-prefix", ValueString: &domain,
					},
				}}, {
				Name:     "resource",
				Resource: fhir,
			},
		},
	}

	resp, err := c.rest.R().
		SetBody(params).
		SetHeader("Content-Type", "application/fhir+json").
		Post(c.config.Url + "/$de-identify")
	if err != nil {
		slog.Error("Failed to send request to the FHIR pseudonymizer", "error", err)
		return nil, err
	}

	// http response status
	success := resp.IsSuccess()

	if success {
		slog.Log(context.Background(), slog.LevelDebug, "FHIR pseudonymizer response", "status", resp.Status(), "body", string(resp.Body()))
		return resp.Body(), nil
	}
	slog.Log(context.Background(), slog.LevelError, "FHIR pseudonymizer response", "status", resp.Status(), "body", string(resp.Body()))
	return nil, errors.New("FHIR pseudonymizer request returned no success")

}
