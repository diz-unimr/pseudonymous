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
	"time"
)

type MongoResource struct {
	Id         primitive.ObjectID `json:"id" bson:"_id"`
	Fhir       bson.M             `json:"fhir" bson:"fhir"`
	Collection *mongo.Collection  `json:"-" bson:"-"`
}

type Provider interface {
	Name() string
	Read(chan<- MongoResource) error
	Write(resource MongoResource) error
	Close() error
}

type MongoFhirProvider struct {
	Client      *mongo.Client
	Context     context.Context
	Source      *mongo.Database
	Destination *mongo.Database
	name        string
	batchSize   int
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
		slog.Error("Failed to connect to mongo", "connection", connection, "error", err.Error())
		os.Exit(1)
	}

	// prefix by convention
	source := client.Database("idat_fhir_" + database)
	dest := client.Database("psn_fhir_" + database)

	return &MongoFhirProvider{
		name:        "MongoFhirProvider",
		Client:      client,
		Context:     ctx,
		Source:      source,
		Destination: dest,
		batchSize:   config.MongoDb.BatchSize,
	}
}

func (p *MongoFhirProvider) Disconnect() error {
	return p.Client.Disconnect(p.Context)
}

func (p *MongoFhirProvider) Read(res chan<- MongoResource) error {
	// get collections
	collectionNames, err := p.Source.ListCollectionNames(context.Background(), bson.M{})
	if err != nil {
		slog.Error("Failed to list collections from source database", "database", p.Source.Name(), "error", err.Error())
		return err
	}

	if len(collectionNames) == 0 {
		slog.Error("No collections found in source database", "database", p.Source.Name())
		return err
	}

	// TODO context with timeout
	ctx := context.Background()
	batchSize := int32(p.batchSize)
	slog.Info("Fetching data from source database", "database", p.Source.Name(), "batchSize", batchSize)

	for _, colName := range collectionNames {
		// get resources
		collection := p.Source.Collection(colName)
		cur, err := collection.Find(ctx, bson.M{}, options.Find().SetBatchSize(batchSize))
		if err != nil {
			slog.Error("Failed to create cursor on database collection", "database", p.Source.Name(), "collection", colName, "error", err.Error())
			return err
		}

		count := 0
		for cur.Next(ctx) {
			var result MongoResource
			err = cur.Decode(&result)
			if err != nil {
				slog.Error("Failed to read next batch", "database", p.Source.Name(), "collection", colName, "error", err.Error())
				return err
			}
			count++
			result.Collection = collection
			res <- result
		}

		slog.Info("Successfully read resources from database collection", "database", p.Source.Name(), "collection", colName, "count", count)

	}
	return nil
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

func (p *MongoFhirProvider) Name() string {
	return p.name
}
