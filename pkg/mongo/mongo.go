package mongodb

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// type MongoDB struct {
// 	*mongo.Database
// }

type Config struct {
	Host     string
	Port     uint
	DBName   string
	Username string
	Password string
}

func ConnectMongoDB(cnf Config) (*mongo.Database, error) {
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d", cnf.Username, cnf.Password, cnf.Host, cnf.Port)
	clientOptions := options.Client().ApplyURI(uri)
	clientOptions.SetTimeout(time.Second * 10)
	clientOptions.Auth.Username = cnf.Username
	clientOptions.Auth.Password = cnf.Password

	client, err := mongo.Connect(clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %v", err)
	}

	log.Println("Connected to MongoDB successfully!")
	return client.Database(cnf.DBName), nil
}
