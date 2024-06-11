package fhir

import (
	"context"
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
	Collection *mongo.Collection  `json:"-" bson:"-"`
}

type Provider interface {
	Read() ([]MongoResource, error)
	Write(resource MongoResource) error
	Close() error
}

type MongoFhirProvider struct {
	Client      *mongo.Client
	Context     context.Context
	Source      *mongo.Database
	Destination *mongo.Database
}

func (p *MongoFhirProvider) Close() error {
	return p.Disconnect()
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

	// TODO configurable prefixes
	source := client.Database("idat_fhir_" + database)
	dest := client.Database("psn_fhir_" + database)

	return &MongoFhirProvider{
		Client:      client,
		Context:     ctx,
		Source:      source,
		Destination: dest,
	}
}

func (p *MongoFhirProvider) Disconnect() error {
	return p.Client.Disconnect(p.Context)
}

func (p *MongoFhirProvider) Read() ([]MongoResource, error) {
	// get collections
	collectionNames, err := p.Source.ListCollectionNames(context.Background(), bson.M{})
	if err != nil {
		slog.Error("Failed to list collections from source database", "database", p.Source.Name(), "error", err.Error())
		return nil, err
	}

	if len(collectionNames) == 0 {
		slog.Error("No collections found in source database", "database", p.Source.Name())
		return nil, err
	}

	var results []MongoResource
	for _, colName := range collectionNames {
		// get resources
		collection := p.Source.Collection(colName)
		cur, err := collection.Find(context.Background(), bson.M{})
		if err != nil {
			slog.Error("Failed to create cursor on database collection", "database", p.Source.Name(), "collection", colName, "error", err.Error())
			return nil, err
		}

		// TODO read in batches
		var resources []MongoResource

		if err = cur.All(context.Background(), &resources); err != nil {
			slog.Error("Failed to list collections from source database", "database", p.Source.Name(), "error", err.Error())
			return nil, err
		}
		for i := range resources {
			resources[i].Collection = collection
		}

		slog.Info("Successfully read resources from database collection", "database", p.Source.Name(), "collection", colName, "count", len(resources))

		results = slices.Concat(results, resources)
	}
	return results, nil
}

func (p *MongoFhirProvider) Write(res MongoResource) error {

	coll := p.Destination.Collection(res.Collection.Name())
	opts := options.Update().SetUpsert(true)
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "fhir", Value: res.Fhir}}}}

	_, err := coll.UpdateByID(context.Background(), res.Id, update, opts)
	if err == nil {
		slog.Debug("Document written", "_id", res.Id.Hex())
	}

	return err
}
