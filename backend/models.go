package main

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string             `bson:"name" json:"name"`
	Email        string             `bson:"email" json:"email"`
	PasswordHash string             `bson:"passwordHash" json:"-"`
	Balance      float64            `bson:"balance" json:"balance"`
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
}
type Order struct {
	ID         string             `bson:"_id" json:"id"`
	ProviderID string             `bson:"providerId,omitempty" json:"providerId,omitempty"`
	UserID     primitive.ObjectID `bson:"userId" json:"-"`
	Service    string             `bson:"service" json:"service"`
	Link       string             `bson:"link" json:"link"`
	Quantity   int                `bson:"quantity" json:"quantity"`
	Status     string             `bson:"status" json:"status"`
	Amount     float64            `bson:"amount" json:"amount"`
	CreatedAt  time.Time          `bson:"createdAt" json:"createdAt"`
}
type Transaction struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      primitive.ObjectID `bson:"userId" json:"-"`
	Description string             `bson:"description" json:"description"`
	Type        string             `bson:"type" json:"type"`
	Status      string             `bson:"status" json:"status"`
	Amount      float64            `bson:"amount" json:"amount"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
}
