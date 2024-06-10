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

func (p *Processor) Pseudonymize(resource bson.M) ([]byte, error) {

	resData, err := json.Marshal(resource)
	if err != nil {
		slog.Error("Unable to marshal resource to JSON", err)
		return nil, err
	}

	resp, err := p.client.Send(resData, p.project+"-")
	if err != nil {
		slog.Error("Failed to pseudonymize resource", err)
		return nil, err
	}

	return resp, nil
}

func (p *Processor) Run() error {
	// TODO
	slog.Info("Reading resources", "provider", p.provider)

	resources, err := p.provider.Read()
	if err != nil {
		slog.Error("Failed to read data", err)
		return err
	}

	wg := new(sync.WaitGroup)
	requests := make(chan MongoResource)
	results := make(chan MongoResource)

	// TODO
	numThreads := 5
	for i := 0; i < numThreads; i++ {
		wg.Add(1)
		go p.createWorker(wg, requests)
	}
	slog.Info("Worker threads created", "threads", numThreads)

	// send resources to workers
	for _, r := range resources {
		requests <- r
	}

	// wait for resources to be processed
	close(requests)
	wg.Wait()
	close(results)

	// TODO save from worker thread
	var res []MongoResource

	for result := range results {
		res = append(res, result)
	}

	return nil
}

func (p *Processor) createWorker(wg *sync.WaitGroup, requests <-chan MongoResource) {
	//, results chan<- MongoResource) {
	defer wg.Done()

	//for {
	//	select {
	//	case r := <-requests:
	//		// TODO error handling
	//		psnResource, _ := p.Pseudonymize(r.Fhir)
	//		var fhirBson bson.M
	//		err := bson.UnmarshalExtJSON(psnResource, true, &fhirBson)
	//		if err != nil {
	//			slog.Error("Failed to convert psn data to BSON", err)
	//			continue
	//		}
	//		results <- MongoResource{
	//			Id:         r.Id,
	//			Fhir:       fhirBson,
	//			collection: r.collection,
	//		}
	//		//case <- QuitChan:
	//		//	wg.Done()
	//		//	return
	//		//}
	//	}
	//}

	for r := range requests {
		// TODO error handling
		psnResource, _ := p.Pseudonymize(r.Fhir)
		var fhirBson bson.M
		err := bson.UnmarshalExtJSON(psnResource, true, &fhirBson)
		if err != nil {
			slog.Error("Failed to convert psn data to BSON", err)
			continue
		}
		//results <- MongoResource{
		//	Id:         r.Id,
		//	Fhir:       fhirBson,
		//	collection: r.collection,
		//}
	}
}
