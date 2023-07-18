package repo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Connection struct {
	client *mongo.Client
	db     *mongo.Database
}

func Connect(uri string, db string) (*Connection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	return &Connection{
		client: client,
		db:     client.Database(db),
	}, err
}

type RepoIdGeneratorFunc[T any, K any] func() (K, error)

type Repo[T any, K any] struct {
	connection *Connection
	collection mongo.Collection
	generateId RepoIdGeneratorFunc[T, K]
	timeout    int
}

func (r *Repo[T, K]) getContext() (context.Context, context.CancelFunc) {
	if r.timeout > 0 {
		return context.WithTimeout(context.Background(), time.Duration(r.timeout)*time.Millisecond)
	}
	return context.Background(), func() {}
}

func (r *Repo[T, K]) New(ids ...K) (data T, err error) {
	var id K
	if len(ids) > 0 {
		id = ids[0]
	} else if r.generateId != nil {
		id, err = r.generateId()
	}
	if err == nil {
		_, err = setIdValue[T, K](&data, &id)
	}
	return
}

func (r *Repo[T, K]) ApplyId(data *T) (err error) {
	var id K
	if r.generateId != nil {
		id, err = r.generateId()
	}
	if err == nil {
		_, err = setIdValue[T, K](data, &id)
	}
	return
}

func (r *Repo[T, K]) Find(filter map[string]interface{}) (result []T, err error) {
	ctx, cancel := r.getContext()
	defer cancel()
	cur, err := r.collection.Find(ctx, filter)
	if err != nil {
		return
	}
	err = cur.All(ctx, &result)
	return
}

func (r *Repo[T, K]) Save(data T) (uid K, err error) {
	ctx, cancel := r.getContext()
	defer cancel()
	id, ok := getIdValue[T, K](&data)
	if !ok {
		return r.Insert(data)
	}
	res, err := r.collection.UpdateOne(ctx, bson.M{"_id": *id}, bson.M{"$set": data}, options.Update().SetUpsert(true))
	if err != nil {
		return
	}
	uid = res.UpsertedID.(K)
	return
}

func (r *Repo[T, K]) Update(id K, data T) (uid K, err error) {
	ctx, cancel := r.getContext()
	defer cancel()
	res, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": data})
	if err != nil {
		return
	}
	uid = res.UpsertedID.(K)
	return
}

func (r *Repo[T, K]) Insert(data T) (id K, err error) {
	ctx, cancel := r.getContext()
	defer cancel()
	res, err := r.collection.InsertOne(ctx, data)
	if err != nil {
		return
	}
	id = res.InsertedID.(K)
	return
}

func (r *Repo[T, K]) Get(id K) (data T, err error) {
	ctx, cancel := r.getContext()
	defer cancel()
	res := r.collection.FindOne(ctx, bson.M{"_id": id})
	err = res.Err()
	if err != nil {
		return
	}
	err = res.Decode(&data)
	return
}

func (r *Repo[T, K]) Delete(data T) (count int64, err error) {
	ctx, cancel := r.getContext()
	defer cancel()
	id, ok := getIdValue[T, K](&data)
	if !ok {
		err = fmt.Errorf("_id field not found for %v", data)
		return
	}
	res, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return
	}
	count = res.DeletedCount
	return
}

func (r *Repo[T, K]) DeleteById(id K) (count int64, err error) {
	ctx, cancel := r.getContext()
	defer cancel()
	res, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return
	}
	count = res.DeletedCount
	return
}

func (r *Repo[T, K]) SetIdGenerator(applyId RepoIdGeneratorFunc[T, K]) {
	r.generateId = applyId
}

func (r *Repo[T, K]) SetTimeout(timeout int) {
	r.timeout = timeout
}

func NewRepo[T any, K any](connection *Connection, collection string) *Repo[T, K] {
	return &Repo[T, K]{
		connection: connection,
		collection: *connection.db.Collection(collection),
	}
}
