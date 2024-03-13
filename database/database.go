package database

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DbConnection struct {
	ConnectionUri string
	Client        *mongo.Client
}

func (dbObj *DbConnection) NewDbConnection(connectionUri string) {
	dbObj.ConnectionUri = connectionUri
}

func (dbObj *DbConnection) Connect() (err error) {
	clientOptions := options.Client().ApplyURI(dbObj.ConnectionUri)
	dbObj.Client, err = mongo.Connect(context.TODO(), clientOptions)
	return err
}

func (dbObj *DbConnection) GetCollection(database string, collection string) *mongo.Collection {
	return dbObj.Client.Database(database).Collection(collection)
}
