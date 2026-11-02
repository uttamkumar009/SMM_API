package main

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Store struct {
	users, orders, transactions *mongo.Collection
	cfg                         Config
}

func NewStore(db *mongo.Database, cfg Config) *Store {
	return &Store{db.Collection("users"), db.Collection("orders"), db.Collection("transactions"), cfg}
}
func (s *Store) User(ctx context.Context, email string) (User, error) {
	var user User
	err := s.users.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	return user, err
}
func (s *Store) UserByID(ctx context.Context, id primitive.ObjectID) (User, error) {
	var user User
	err := s.users.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	return user, err
}
func (s *Store) EnsureDemo(ctx context.Context, hash string) error {
	_, err := s.users.UpdateOne(ctx, bson.M{"email": s.cfg.DemoEmail}, bson.M{"$setOnInsert": bson.M{
		"name": "Alex Kumar", "email": s.cfg.DemoEmail, "passwordHash": hash, "balance": 248.90, "createdAt": time.Now(),
	}}, options.Update().SetUpsert(true))
	return err
}
func (s *Store) Orders(ctx context.Context, userID primitive.ObjectID) ([]Order, error) {
	cursor, err := s.orders.Find(ctx, bson.M{"userId": userID}, options.Find().SetSort(bson.D{{"createdAt", -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var result []Order
	err = cursor.All(ctx, &result)
	return result, err
}
func (s *Store) Transactions(ctx context.Context, userID primitive.ObjectID) ([]Transaction, error) {
	cursor, err := s.transactions.Find(ctx, bson.M{"userId": userID}, options.Find().SetSort(bson.D{{"createdAt", -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var result []Transaction
	err = cursor.All(ctx, &result)
	return result, err
}
func (s *Store) AddFunds(ctx context.Context, id primitive.ObjectID, amount float64) (User, error) {
	var user User
	err := s.users.FindOneAndUpdate(ctx, bson.M{"_id": id}, bson.M{"$inc": bson.M{"balance": amount}},
		options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&user)
	if err != nil {
		return user, err
	}
	_, err = s.transactions.InsertOne(ctx, Transaction{UserID: id, Description: "Added funds", Type: "Credit", Status: "Completed", Amount: amount, CreatedAt: time.Now()})
	return user, err
}
func (s *Store) CreateOrder(ctx context.Context, userID primitive.ObjectID, providerID, service, link string, quantity int, amount float64) (Order, error) {
	var user User
	if err := s.users.FindOne(ctx, bson.M{"_id": userID}).Decode(&user); err != nil {
		return Order{}, err
	}
	if user.Balance < amount {
		return Order{}, errors.New("insufficient wallet balance")
	}
	order := Order{ID: "#OR-" + primitive.NewObjectID().Hex()[:8], ProviderID: providerID, UserID: userID, Service: service, Link: link, Quantity: quantity, Status: "processing", Amount: amount, CreatedAt: time.Now()}
	session, err := s.users.Database().Client().StartSession()
	if err != nil {
		return Order{}, err
	}
	defer session.EndSession(ctx)
	_, err = session.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
		if _, e := s.users.UpdateOne(sc, bson.M{"_id": userID, "balance": bson.M{"$gte": amount}}, bson.M{"$inc": bson.M{"balance": -amount}}); e != nil {
			return nil, e
		}
		if _, e := s.orders.InsertOne(sc, order); e != nil {
			return nil, e
		}
		_, e := s.transactions.InsertOne(sc, Transaction{UserID: userID, Description: service + " order", Type: "Debit", Status: "Completed", Amount: -amount, CreatedAt: time.Now()})
		return nil, e
	})
	return order, err
}
