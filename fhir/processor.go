package fhir

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"log/slog"
	"pseudonymous/config"
	"strings"
	"sync"
	"time"
)

type Processor struct {
	provider    Provider
	client      *PsnClient
	project     string
	concurrency int
}

func NewProcessor(config *config.AppConfig, project string) *Processor {
	concurrency := config.App.Concurrency
	if concurrency == 0 {
		concurrency = 1
	}

	return &Processor{
		provider:    NewProvider(config.Fhir.Provider, project),
		client:      NewClient(config.Fhir.Pseudonymizer),
		project:     project,
		concurrency: concurrency,
	}
}

func (p *Processor) Close() error {
	return p.provider.Close()
}

func (p *Processor) Pseudonymize(resource bson.M) ([]byte, error) {

	resData, err := json.Marshal(resource)
	if err != nil {
		slog.Error("Unable to marshal resource to JSON", "error", err.Error())
		return nil, err
	}

	resp, err := p.client.Send(resData, p.project+"-")
	if err != nil {
		slog.Error("Failed to pseudonymize resource", "error", err.Error())
		return nil, err
	}

	return resp, nil
}

func (p *Processor) Run() error {
	start := time.Now()

	wg := new(sync.WaitGroup)
	jobs := make(chan MongoResource)
	results := make(chan string)

	concurrency := p.concurrency
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go p.createWorker(wg, jobs, results)
	}
	slog.Info("Worker created", "concurrency", concurrency)

	go func() {
		slog.Info("Reading resources", "provider", p.provider.Name())
		err := p.provider.Read(jobs)
		if err != nil {
			slog.Error("Failed to read data", "error", err.Error())
		}

		// wait for resources to be processed
		close(jobs)
		// wait for results
		wg.Wait()
		close(results)
	}()

	// read results
	m := make(map[string]int)
	for r := range results {
		m[r]++
	}

	slog.Info("Finished processing results", "count", convertToString(m), "time", time.Since(start))

	return p.provider.Close()
}

func (p *Processor) createWorker(wg *sync.WaitGroup, jobs <-chan MongoResource, results chan string) {
	defer wg.Done()

	for r := range jobs {

		// pseudonymize
		psnResource, _ := p.Pseudonymize(r.Fhir)

		// unmarshal result
		var fhirBson bson.M
		err := bson.UnmarshalExtJSON(psnResource, true, &fhirBson)
		if err != nil {
			slog.Error("Failed to convert psn data to BSON", "error", err.Error())
			continue
		}

		// save result
		psnResult := MongoResource{
			Id:         r.Id,
			Fhir:       fhirBson,
			Collection: r.Collection,
		}
		err = p.provider.Write(psnResult)
		if err != nil {
			slog.Error("Failed to save psn data to database collection",
				"id", psnResult.Id,
				"collection", psnResult.Collection.Name(),
				"error", err.Error())
			continue
		}

		slog.Debug("Successfully processed resource", "_id", psnResult.Id, "collections", psnResult.Collection.Name())

		// send result
		results <- psnResult.Collection.Name()

	}
}

func convertToString(m map[string]int) string {
	b := new(bytes.Buffer)
	for key, value := range m {
		_, err := fmt.Fprintf(b, "%s=%d ", key, value)
		if err != nil {
			return ""
		}
	}

	return strings.TrimSpace(b.String())
}
