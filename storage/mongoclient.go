package storage

import (
	"context"
	"strings"
	"time"

	"github.com/kickback-app/common/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type cursorDecoder struct {
	ctx    context.Context
	cursor *mongo.Cursor
}

// Decode reads from the cursor and unmarshalls the data into the given
// object pointer
func (cd cursorDecoder) Decode(v interface{}) error {
	return cd.cursor.All(cd.ctx, v)
}

var retryableErrs = []string{}

type mongoClient struct {
	client      *mongo.Client
	database    *mongo.Database
	maxRetries  int
	retryPolicy func(int) time.Duration
}

type CallContext struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// NewCallContext is used when a client wants to make a call to the data store and provide a
// context object
func NewCallContext() *CallContext {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	return &CallContext{ctx: ctx, cancel: cancel}
}

// NewMongoClient returns a new mongoDB client
func NewMongoClient(client *mongo.Client, database *mongo.Database) *mongoClient {
	retries := 3
	retryPolicy := func(i int) time.Duration {
		return time.Duration(10*i) * time.Second
	}
	return &mongoClient{
		client:      client,
		database:    database,
		maxRetries:  retries,
		retryPolicy: retryPolicy,
	}
}

func (mc *mongoClient) Collection(collection string) *mongo.Collection {
	return mc.database.Collection(collection)
}

func (mc *mongoClient) Close(l log.Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer func() {
		if err := mc.client.Disconnect(ctx); err != nil {
			// panic(err) // TODO: should this be a panic
			l.Error("unable to successfully disconnect from db: %v", err)
		}
		l.Debug("successfully disconnected from mongo client...✅👍")
	}()
}

type FindOneParams struct {
	Collection     string
	Filter         interface{}
	AdditionalOpts []*options.FindOneOptions
}

func (fop *FindOneParams) valid() bool {
	return fop.Collection != "" && fop.Filter != nil
}

func (mc *mongoClient) FindOne(l log.Logger, cc *CallContext, params *FindOneParams) (Decoder, error) {
	attempt := 0
	for attempt < mc.maxRetries {
		res, err := mc.findOne(l, cc, params)
		if err != nil {
			if isRetryableErr(err) {
				attempt++
				l.Warn("retryable error %s, attempting again (attempt: %v)", err, attempt)
				continue
			}
			return res, err
		}
		return res, err
	}
	return nil, MaxRetriesExceededError{}
}

func (mc *mongoClient) findOne(l log.Logger, cc *CallContext, params *FindOneParams) (Decoder, error) {
	if ok := params.valid(); !ok {
		l.Error("invalid parameters")
		return nil, MissingRequiredParameterError{}
	}
	collection := mc.Collection(params.Collection)
	resp := collection.FindOne(cc.ctx, params.Filter, params.AdditionalOpts...)
	err := resp.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, NotFoundError{}
		}
		l.Error("error finding doc in %s: %v", params.Collection, err)
		return nil, err
	}
	var result interface{}
	err = resp.Decode(&result)
	if err != nil {
		l.Error("unable to decode response: %v", err)
		return nil, err
	}
	return resp, nil
}

type FindManyParams struct {
	Collection     string
	Filter         interface{}
	AdditionalOpts []*options.FindOptions
}

func (fmp *FindManyParams) valid() bool {
	return fmp.Collection != "" && fmp.Filter != nil
}

func (mc *mongoClient) FindMany(l log.Logger, cc *CallContext, params *FindManyParams) (Decoder, error) {
	attempt := 0
	for attempt < mc.maxRetries {
		res, err := mc.findMany(l, cc, params)
		if err != nil {
			if isRetryableErr(err) {
				attempt++
				l.Warn("retryable error %s, attempting again (attempt: %v)", err, attempt)
				continue
			}
			return res, err
		}
		return res, err
	}
	return nil, MaxRetriesExceededError{}
}

func (mc *mongoClient) findMany(l log.Logger, cc *CallContext, params *FindManyParams) (Decoder, error) {
	if ok := params.valid(); !ok {
		l.Error("invalid parameters for findMany operation")
		return nil, MissingRequiredParameterError{}
	}
	collection := mc.Collection(params.Collection)
	cursor, err := collection.Find(cc.ctx, params.Filter, params.AdditionalOpts...)
	if err != nil {
		l.Error("unable to find docs in %v: %v", params.Collection, err)
		return nil, err
	}
	return cursorDecoder{
		cursor: cursor,
		ctx:    cc.ctx, // @todo should we generate a new context here?
	}, nil
}

type InsertOneParams struct {
	Collection     string
	AdditionalOpts []*options.InsertOneOptions
}

func (iop *InsertOneParams) valid() bool {
	return iop.Collection != ""
}

type InsertOneResult struct {
	InsertedID interface{}
}

func (mc *mongoClient) InsertOne(l log.Logger, cc *CallContext, document interface{}, params *InsertOneParams) (*InsertOneResult, error) {
	attempt := 0
	for attempt < mc.maxRetries {
		res, err := mc.insertOne(l, cc, document, params)
		if err != nil {
			if isRetryableErr(err) {
				attempt++
				l.Warn("retryable error %s, attempting again (attempt: %v)", err, attempt)
				continue
			}
			return res, err
		}
		return res, err
	}
	return nil, MaxRetriesExceededError{}
}

func (mc *mongoClient) insertOne(l log.Logger, cc *CallContext, document interface{}, params *InsertOneParams) (*InsertOneResult, error) {
	var result *InsertOneResult
	if ok := params.valid(); !ok {
		l.Error("invalid parameters for insertOne operation")
		return nil, MissingRequiredParameterError{}
	}
	collection := mc.Collection(params.Collection)
	res, err := collection.InsertOne(cc.ctx, document, params.AdditionalOpts...)
	if err != nil {
		if isCollisionErr(err) {
			l.Error("collision found trying to insert into %v: %v", params.Collection, err)
			return nil, CollisionError{CollectionName: params.Collection}
		}
		l.Error("unable to insert document into %v", err)
		return nil, err
	}
	result.InsertedID = res.InsertedID
	return result, nil
}

type InsertManyParams struct {
	Collection     string
	AdditionalOpts []*options.InsertManyOptions
}

func (imp *InsertManyParams) valid() bool {
	return imp.Collection != ""
}

type InsertManyResult struct {
	InsertedIDs []interface{}
}

func (mc *mongoClient) InsertMany(l log.Logger, cc *CallContext, data []interface{}, params *InsertManyParams) (*InsertManyResult, error) {
	attempt := 0
	for attempt < mc.maxRetries {
		res, err := mc.insertMany(l, cc, data, params)
		if err != nil {
			if isRetryableErr(err) {
				attempt++
				l.Warn("retryable error %s, attempting again (attempt: %v)", err, attempt)
				continue
			}
			return res, err
		}
		return res, err
	}
	return nil, MaxRetriesExceededError{}
}

func (mc *mongoClient) insertMany(l log.Logger, cc *CallContext, data []interface{}, params *InsertManyParams) (*InsertManyResult, error) {
	var result *InsertManyResult
	if ok := params.valid(); !ok {
		l.Error("invalid parameters for insertMany operation")
		return nil, MissingRequiredParameterError{}
	}
	collection := mc.Collection(params.Collection)
	res, err := collection.InsertMany(cc.ctx, data, params.AdditionalOpts...)
	if err != nil {
		if isCollisionErr(err) {
			l.Error("collision found trying to insert many into %v: %v", params.Collection, err)
			return nil, CollisionError{CollectionName: params.Collection}
		}
		l.Error("unable to insert many into %v: %v", params.Collection, err)
		return nil, err
	}
	result.InsertedIDs = res.InsertedIDs
	return result, nil
}

type UpsertParams struct {
	Collection     string
	Filter         interface{}
	Multiple       bool
	Generic        bool
	Upsert         *bool
	AdditionalOpts []*options.UpdateOptions
}

func (up *UpsertParams) valid() bool {
	return up.Collection != "" && up.Filter != nil
}

type UpsertResult struct {
	MatchedCount  int64       // The number of documents matched by the filter.
	ModifiedCount int64       // The number of documents modified by the operation.
	UpsertedCount int64       // The number of documents upserted by the operation.
	UpsertedID    interface{} // The _id field of the upserted document, or nil if no upsert was done.
}

func (mc *mongoClient) Upsert(l log.Logger, cc *CallContext, updates interface{}, params *UpsertParams) (*UpsertResult, error) {
	attempt := 0
	for attempt < mc.maxRetries {
		res, err := mc.upsert(l, cc, updates, params)
		if err != nil {
			if isRetryableErr(err) {
				attempt++
				l.Warn("retryable error %s, attempting again (attempt: %v)", err, attempt)
				continue
			}
			return res, err
		}
		return res, err
	}
	return nil, MaxRetriesExceededError{}
}

func (mc *mongoClient) upsert(l log.Logger, cc *CallContext, updates interface{}, params *UpsertParams) (*UpsertResult, error) {
	var resp *UpsertResult
	if ok := params.valid(); !ok {
		l.Error("invalid parameters for upsert operation")
		return resp, MissingRequiredParameterError{}
	}
	collection := mc.Collection(params.Collection)
	updateCmd := updates
	if params.Generic {
		updateCmd = bson.D{{Key: "$set", Value: updates}}
	}
	upsert := true
	if params.Upsert != nil {
		upsert = *params.Upsert
	}
	params.AdditionalOpts = append(params.AdditionalOpts, options.Update().SetUpsert(upsert))
	var err error
	var res *mongo.UpdateResult
	if !params.Multiple {
		res, err = collection.UpdateOne(cc.ctx, params.Filter, updateCmd, params.AdditionalOpts...)
	} else {
		res, err = collection.UpdateMany(cc.ctx, params.Filter, updateCmd, params.AdditionalOpts...)
	}
	if err != nil {
		l.Error("unable to update doc(s): %v", err)
		return resp, err
	}
	resp.MatchedCount = res.MatchedCount
	resp.ModifiedCount = res.ModifiedCount
	resp.UpsertedCount = res.UpsertedCount
	resp.UpsertedID = res.UpsertedID
	return resp, nil
}

type DeleteParams struct {
	Collection     string
	Filter         interface{}
	Multiple       bool
	Generic        bool
	AdditionalOpts []*options.DeleteOptions
}

func (dp *DeleteParams) valid() bool {
	return dp.Collection != "" && dp.Filter != nil
}

type DeleteResult struct {
	DeletedCount int64
}

func (mc *mongoClient) Delete(l log.Logger, cc *CallContext, params *DeleteParams) (*DeleteResult, error) {
	attempt := 0
	for attempt < mc.maxRetries {
		res, err := mc.delete(l, cc, params)
		if err != nil {
			if isRetryableErr(err) {
				attempt++
				l.Warn("retryable error %s, attempting again (attempt: %v)", err, attempt)
				continue
			}
			return res, err
		}
		return res, err
	}
	return nil, MaxRetriesExceededError{}
}

func (mc *mongoClient) delete(l log.Logger, cc *CallContext, params *DeleteParams) (*DeleteResult, error) {
	var result *DeleteResult
	if ok := params.valid(); !ok {
		l.Error("invalid parameters for delete operation")
		return result, MissingRequiredParameterError{}
	}
	collection := mc.Collection(params.Collection)
	var err error
	var res *mongo.DeleteResult
	if !params.Multiple {
		res, err = collection.DeleteOne(cc.ctx, params.Filter, params.AdditionalOpts...)
	} else {
		res, err = collection.DeleteMany(cc.ctx, params.Filter, params.AdditionalOpts...)
	}
	if err != nil {
		l.Error("unable to delete doc(s): %v", err)
		return result, err
	}
	result.DeletedCount = res.DeletedCount
	return result, nil
}

func isRetryableErr(err error) bool {
	for _, eStr := range retryableErrs {
		if strings.Contains(err.Error(), eStr) {
			return true
		}
	}
	return false
}

type MaxRetriesExceededError struct{}

func (e MaxRetriesExceededError) Error() string {
	return "max retries exhausted trying to call database"
}
