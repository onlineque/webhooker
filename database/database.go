package database

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strings"
)

type DbConnection struct {
	ConnectionUri string
	Database      string
	Client        *mongo.Client
}

func (dbObj *DbConnection) NewDbConnection(connectionUri string) {
	dbObj.ConnectionUri = connectionUri
}

func (dbObj *DbConnection) Connect() (err error) {
	splitDbUri := strings.Split(dbObj.ConnectionUri, "/")
	dbObj.Database = splitDbUri[len(splitDbUri)-1]

	clientOptions := options.Client().ApplyURI(dbObj.ConnectionUri)
	dbObj.Client, err = mongo.Connect(context.TODO(), clientOptions)
	return err
}

func (dbObj *DbConnection) GetCollection(collection string) *mongo.Collection {
	return dbObj.Client.Database(dbObj.Database).Collection(collection)
}
