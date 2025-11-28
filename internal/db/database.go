package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	ID        string    `bson:"_id"`
	Email     string    `bson:"email"`
	Name      string    `bson:"name"`
	AvatarURL string    `bson:"avatar_url"`
	CreatedAt time.Time `bson:"created_at"`
}

var client *mongo.Client
var db *mongo.Database

func Connect(uri string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var err error
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return err
	}
	db = client.Database("gossg")
	return nil
}

func GetUser(id string) (*User, error) {
	var user User
	err := db.Collection("users").FindOne(context.Background(), bson.M{"_id": id}).Decode(&user)
	return &user, err
}

func SaveUser(user *User) error {
	opts := options.Replace().SetUpsert(true)
	_, err := db.Collection("users").ReplaceOne(context.Background(), bson.M{"_id": user.ID}, user, opts)
	return err
}
