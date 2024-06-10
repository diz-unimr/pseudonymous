package fhir

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log/slog"
	"os"
	"pseudonymous/config"
	"slices"
	"time"
)

type MongoResource struct {
	Id         primitive.ObjectID `json:"id" bson:"_id"`
	Fhir       bson.M             `json:"fhir" bson:"fhir"`
	collection mongo.Collection
}

type Provider interface {
	Read() ([]MongoResource, error)
	Save(resource MongoResource) error
}

type MongoFhirProvider struct {
	Client   *mongo.Client
	Context  context.Context
	Database *mongo.Database
}

func NewProvider(config config.Provider, database string) *MongoFhirProvider {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	connection := config.MongoDb.Connection
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connection))
	if err != nil {
		slog.Error("failed to connect to mongo", "connection", connection)
		os.Exit(1)
	}

	// TODO configurable prefix
	db := client.Database("idat_fhir_" + database)

	return &MongoFhirProvider{
		Client:   client,
		Context:  ctx,
		Database: db,
	}
}

func (p *MongoFhirProvider) disconnect() {
	if err := p.Client.Disconnect(p.Context); err != nil {
		panic(err)
	}
}

func (p *MongoFhirProvider) Read() ([]MongoResource, error) {
	// get collections
	collectionNames, err := p.Database.ListCollectionNames(context.Background(), bson.M{})
	if err != nil {
		slog.Error("Failed to list collections from source database", "database", p.Database.Name(), "error", err.Error())
		return nil, err
	}

	if len(collectionNames) == 0 {
		slog.Error("No collections found in source database", "database", p.Database.Name())
		return nil, err
	}

	var results []MongoResource
	for _, colName := range collectionNames {
		// get resources
		collection := p.Database.Collection(colName)
		cur, err := collection.Find(context.Background(), bson.M{})
		if err != nil {
			slog.Error("Failed to create cursor on database collection", "database", p.Database.Name(), "collection", colName, "error", err.Error())
			return nil, err
		}

		// TODO read in batches
		var resources []MongoResource

		if err = cur.All(context.Background(), &resources); err != nil {
			// TODO error handling
			slog.Error("Failed to list collections from source database", "database", p.Database.Name(), "error", err.Error())
			return nil, err
		}

		slog.Info("Successfully read resources from database collection", "database", p.Database.Name(), "collection", colName, "count", len(resources))

		results = slices.Concat(results, resources)
	}
	return results, nil
}

func (p *MongoFhirProvider) Save(resource MongoResource) error {
	// TODO
	return errors.New("failed to save data")
}
