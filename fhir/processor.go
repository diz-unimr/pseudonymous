package fhir

import (
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"log/slog"
	"pseudonymous/config"
	"sync"
)

type Processor struct {
	provider Provider
	client   *PsnClient
	project  string
}

func NewProcessor(config *config.AppConfig, project string) *Processor {
	return &Processor{
		provider: NewProvider(config.Fhir.Provider, project),
		client:   NewClient(config.Fhir.Pseudonymizer),
		project:  project,
	}
}

func (p *Processor) Close() error {
	return p.provider.Close()
}

func (p *Processor) Pseudonymize(resource bson.M) ([]byte, error) {

	resData, err := json.Marshal(resource)
	if err != nil {
		slog.Error("Unable to marshal resource to JSON", err.Error())
		return nil, err
	}

	resp, err := p.client.Send(resData, p.project+"-")
	if err != nil {
		slog.Error("Failed to pseudonymize resource", err.Error())
		return nil, err
	}

	return resp, nil
}

func (p *Processor) Run() error {
	// TODO
	slog.Info("Reading resources", "provider", p.provider)

	resources, err := p.provider.Read()
	if err != nil {
		slog.Error("Failed to read data", err.Error())
		return err
	}

	wg := new(sync.WaitGroup)
	jobs := make(chan MongoResource)
	results := make(chan MongoResource)

	// TODO
	numThreads := 5
	for i := 0; i < numThreads; i++ {
		wg.Add(1)
		go p.createWorker(wg, jobs, results)
	}
	slog.Info("Worker threads created", "threads", numThreads)

	// send resources to workers
	for _, r := range resources {
		jobs <- r
	}

	go func() {
		// wait for resources to be processed
		close(jobs)
		wg.Wait()
		close(results)
	}()

	// wait for results
	var res []MongoResource
	for result := range results {
		res = append(res, result)
	}

	slog.Info("Finished processing results", "count", len(res), "results", res)

	return p.provider.Close()
}

func (p *Processor) createWorker(wg *sync.WaitGroup, jobs <-chan MongoResource, results chan<- MongoResource) {
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

		// send result
		results <- psnResult

	}
}
