package mongolib

import (
	"context"
	"fmt"
	"slices"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CollectionDef struct {
	Name       string
	Timeseries *TimeseriesDef
	Indexes    []IndexDef
}

type TimeseriesDef struct {
	TimeField   string
	MetaField   string
	Granularity string
}

type IndexDef struct {
	Fields bson.D
	Unique bool
}

func (c *Connection) DefineCollections(defs []CollectionDef) error {
	ctx, cancel := c.getTimeoutContext()
	defer cancel()
	names, err := c.DB.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return fmt.Errorf("cannot obtain collection list: %w", &err)
	}

	for _, def := range defs {
		exists := slices.Contains(names, def.Name)

		if !exists {
			opts := options.CreateCollection()

			if def.Timeseries != nil {
				// Creates a time series collection that stores "temperature" values over time
				ts := def.Timeseries
				tso := options.TimeSeries()

				if ts.TimeField != "" {
					tso.SetTimeField(ts.TimeField)
				}

				if ts.MetaField != "" {
					tso.SetMetaField(ts.MetaField)
				}

				if ts.Granularity != "" {
					tso.SetGranularity(ts.Granularity)
				}

				opts.SetTimeSeriesOptions(tso)
			}

			ctx, cancel := c.getTimeoutContext()
			defer cancel()

			err := c.DB.CreateCollection(ctx, def.Name, opts)
			if err != nil {
				return fmt.Errorf("cannot define collection %s: %w", def.Name, err)
			}
		}

		if def.Indexes != nil {
			coll := c.DB.Collection(def.Name)
			indexes := coll.Indexes()
			for _, idx := range def.Indexes {
				ctx, cancel := c.getTimeoutContext()
				defer cancel()

				opts := options.Index()

				if idx.Unique {
					opts.SetUnique(true)
				}

				indexName, err := indexes.CreateOne(
					ctx,
					mongo.IndexModel{
						Keys:    idx.Fields,
						Options: opts,
					},
				)
				if err != nil {
					return fmt.Errorf("cannot create collection %s index: %w", def.Name, err)
				}

				c.Logger.Debug("index created", "collection", def.Name, "name", indexName, "index", idx)
			}
		}

	}
	return nil
}

func (c *Connection) getTimeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), c.timeout)
}

func (c *Connection) Coll(name string) *mongo.Collection {
	coll, ok := c.collections[name]
	if !ok {
		coll = c.DB.Collection(name)
		c.collections[name] = coll
	}
	return coll
}
