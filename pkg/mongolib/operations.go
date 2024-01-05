package mongolib

import (
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (c *Connection) InsertOne(coll string, data any) (res *mongo.InsertOneResult, err error) {
	ctx, cancel := c.getTimeoutContext()
	defer cancel()
	return c.Coll(coll).InsertOne(ctx, data)
}

func (c *Connection) UpsertOne(coll string, filter any, data any) (res *mongo.UpdateResult, err error) {
	ctx, cancel := c.getTimeoutContext()
	defer cancel()
	opts := options.Update().SetUpsert(true)
	return c.Coll(coll).UpdateOne(ctx, filter, data, opts)
}
