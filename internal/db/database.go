package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

type Post struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	UserID      string             `bson:"user_id"`
	Slug        string             `bson:"slug"`
	Title       string             `bson:"title"`
	Content     string             `bson:"content"`
	IsPublic    bool               `bson:"is_public"`
	CreatedAt   time.Time          `bson:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"`
}

var client *mongo.Client
var database *mongo.Database

func Connect(uri string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var err error
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return err
	}
	database = client.Database("gossg")
	return nil
}

func GetUser(id string) (*User, error) {
	var user User
	err := database.Collection("users").FindOne(context.Background(), bson.M{"_id": id}).Decode(&user)
	return &user, err
}

func GetUserByEmail(email string) (*User, error) {
	var user User
	err := database.Collection("users").FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	return &user, err
}

func SaveUser(user *User) error {
	opts := options.Replace().SetUpsert(true)
	_, err := database.Collection("users").ReplaceOne(context.Background(), bson.M{"_id": user.ID}, user, opts)
	return err
}

func CreatePost(post *Post) error {
	post.CreatedAt = time.Now()
	post.UpdatedAt = time.Now()
	result, err := database.Collection("posts").InsertOne(context.Background(), post)
	if err != nil {
		return err
	}
	post.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func GetPost(id string) (*Post, error) {
	var post Post
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	err = database.Collection("posts").FindOne(context.Background(), bson.M{"_id": objID}).Decode(&post)
	return &post, err
}

func GetPostBySlug(userID, slug string) (*Post, error) {
	var post Post
	err := database.Collection("posts").FindOne(context.Background(), bson.M{"user_id": userID, "slug": slug}).Decode(&post)
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func GetPublicPostBySlug(userID, slug string) (*Post, error) {
	var post Post
	err := database.Collection("posts").FindOne(context.Background(), bson.M{"user_id": userID, "slug": slug, "is_public": true}).Decode(&post)
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func GetPublicPostBySlug(userID, slug string) (*Post, error) {
	var post Post
	err := database.Collection("posts").FindOne(context.Background(), bson.M{"user_id": userID, "slug": slug, "is_public": true}).Decode(&post)
	return &post, err
}

func GetUserPosts(userID string) ([]*Post, error) {
	cursor, err := database.Collection("posts").Find(context.Background(), bson.M{"user_id": userID}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	var posts []*Post
	if err := cursor.All(context.Background(), &posts); err != nil {
		return nil, err
	}
	return posts, nil
}

func GetPublicPosts(userID string) ([]*Post, error) {
	cursor, err := database.Collection("posts").Find(context.Background(), bson.M{"user_id": userID, "is_public": true}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	var posts []*Post
	if err := cursor.All(context.Background(), &posts); err != nil {
		return nil, err
	}
	return posts, nil
}

func UpdatePost(post *Post) error {
	post.UpdatedAt = time.Now()
	_, err := database.Collection("posts").UpdateOne(
		context.Background(),
		bson.M{"_id": post.ID},
		bson.M{"$set": bson.M{"title": post.Title, "content": post.Content, "is_public": post.IsPublic, "updated_at": post.UpdatedAt}},
	)
	return err
}

func DeletePost(id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = database.Collection("posts").DeleteOne(context.Background(), bson.M{"_id": objID})
	return err
}

func TogglePostVisibility(id string) (*Post, error) {
	var post Post
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	err = database.Collection("posts").FindOne(context.Background(), bson.M{"_id": objID}).Decode(&post)
	if err != nil {
		return nil, err
	}

	newVisibility := !post.IsPublic
	_, err = database.Collection("posts").UpdateOne(
		context.Background(),
		bson.M{"_id": objID},
		bson.M{"$set": bson.M{"is_public": newVisibility, "updated_at": time.Now()}},
	)
	if err != nil {
		return nil, err
	}

	post.IsPublic = newVisibility
	return &post, nil
}
