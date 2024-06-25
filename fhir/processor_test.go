package fhir

import (
	"context"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
	"io"
	"net/http"
	"pseudonymous/config"
	"testing"
)

func TestRun(t *testing.T) {

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {

		provider := &MongoFhirProvider{
			Client:      mt.Client,
			Context:     context.Background(),
			Source:      mt.DB,
			Destination: mt.DB,
			name:        "MongoDB Test Provider",
		}
		p := &Processor{
			provider:    provider,
			client:      NewClient(config.Pseudonymizer{}),
			project:     "test",
			concurrency: 1,
		}

		// test resources
		pat := MongoResource{
			Id:         primitive.ObjectID{},
			Fhir:       bson.M{"resourceType": "Patient"},
			Collection: nil,
		}
		obs := MongoResource{
			Id:         primitive.ObjectID{},
			Fhir:       bson.M{"resourceType": "Observation"},
			Collection: nil,
		}

		collNames := []bson.D{{{"name", "Patient"}}, {{"name", "Observation"}}}

		// expect one Patient and one Observation in results
		expResultCount := map[string]int{"Patient": 1, "Observation": 1}

		// setup mocks
		// mongodb
		mt.AddMockResponses(
			// list collections and read data
			mtest.CreateCursorResponse(1, "test.$cmd.listCollections", mtest.FirstBatch, collNames...),
			mtest.CreateSuccessResponse(), mtest.CreateSuccessResponse(),
			mtest.CreateCursorResponse(2, "test.Patient", mtest.FirstBatch, toDoc(pat)),
			mtest.CreateSuccessResponse(),
			mtest.CreateCursorResponse(3, "test.Observation", mtest.FirstBatch, toDoc(obs)),

			// save data back
			mtest.CreateSuccessResponse(), mtest.CreateSuccessResponse(),
			mtest.CreateSuccessResponse(),
		)

		// rest client (pseudonymization)
		httpmock.ActivateNonDefault(p.client.rest.GetClient())
		httpmock.RegisterResponder("POST", "/$de-identify", func(req *http.Request) (*http.Response, error) {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return httpmock.NewStringResponse(400, ""), nil
			}
			// just return the request body
			// it's a Parameters resource, but that doesn't matter here
			return httpmock.NewBytesResponse(200, body), nil
		})

		// act
		result, err := p.Run()

		assert.Nil(t, err)
		assert.Equal(t, expResultCount, result.count)
	})
}

func toDoc(v interface{}) (doc bson.D) {
	data, _ := bson.Marshal(v)

	_ = bson.Unmarshal(data, &doc)
	return
}
