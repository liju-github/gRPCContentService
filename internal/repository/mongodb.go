package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/liju-github/ContentService/internal/models"
)

type Repository interface {
	PostQuestion(ctx context.Context, question *models.Question) error
	GetQuestionsByUserID(ctx context.Context, userID string) ([]models.Question, error)
	GetQuestionsByTags(ctx context.Context, tags []string) ([]models.Question, error)
	GetQuestionsByWord(ctx context.Context, word string) ([]models.Question, error)
	DeleteQuestion(ctx context.Context, questionID string) error
	GetQuestionByID(ctx context.Context, questionID string) (*models.Question, error)
	PostAnswer(ctx context.Context, questionID string, answer *models.Answer) error
	DeleteAnswer(ctx context.Context, questionID, answerID string) error
	FlagQuestion(ctx context.Context, questionID, userID, reason string) error
	FlagAnswer(ctx context.Context, questionID, answerID, userID, reason string) error
	MarkQuestionAsAnswered(ctx context.Context, questionID string) error
	GetUserFeed(ctx context.Context, userID string) ([]models.Question, error)
	GetFlaggedQuestions(ctx context.Context) ([]models.Question, int32, error)
	GetFlaggedAnswers(ctx context.Context) ([]models.Answer, int32, error)

	//additional methods...
	AddTag(ctx context.Context, tag *models.Tag) error
	RemoveTag(ctx context.Context, tagName string) error
	UpvoteAnswer(ctx context.Context, questionID, answerID, userID string) error
	DownvoteAnswer(ctx context.Context, questionID, answerID string) error
	SearchQuestionsAnswersUsers(ctx context.Context, keyword string) (*models.SearchResult, error)
	// Additional methods for vote tracking
	HasUserVotedOnAnswer(ctx context.Context, questionID, answerID, userID string) (bool, string, error)
	GetAnswerOwnerID(ctx context.Context, questionID, answerID string) (string, error)
	GetUserIDFromQuestionID(ctx context.Context, questionID string) (string, error)
}

type MongoRepository struct {
	client    *mongo.Client
	database  string
	questions *mongo.Collection
	tags      *mongo.Collection
}

func NewMongoRepository(cfg *models.MongoConfig) (*MongoRepository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.URI))
	if err != nil {
		return nil, err
	}

	// Ping database to verify connection
	if err = client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	db := client.Database(cfg.Database)

	// Create indexes
	_, err = db.Collection("questions").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "tags", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "question", Value: "text"}},
		},
	})
	if err != nil {
		return nil, err
	}

	return &MongoRepository{
		client:    client,
		database:  cfg.Database,
		questions: db.Collection("questions"),
		tags:      db.Collection("tags"),
	}, nil
}

func (r *MongoRepository) PostQuestion(ctx context.Context, question *models.Question) error {
	question.ID = primitive.NewObjectID()
	question.CreatedAt = time.Now()
	question.IsAnswered = false

	_, err := r.questions.InsertOne(ctx, question)
	return err
}

func (r *MongoRepository) GetQuestionsByUserID(ctx context.Context, userID string) ([]models.Question, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.questions.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var questions []models.Question
	if err = cursor.All(ctx, &questions); err != nil {
		return nil, err
	}

	return questions, nil
}

func (r *MongoRepository) GetQuestionsByTags(ctx context.Context, tags []string) ([]models.Question, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.questions.Find(ctx, bson.M{"tags": bson.M{"$in": tags}}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var questions []models.Question
	if err = cursor.All(ctx, &questions); err != nil {
		return nil, err
	}

	return questions, nil
}

func (r *MongoRepository) GetQuestionsByWord(ctx context.Context, word string) ([]models.Question, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.questions.Find(ctx, bson.M{
		"$text": bson.M{
			"$search": word,
		},
	}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var questions []models.Question
	if err = cursor.All(ctx, &questions); err != nil {
		return nil, err
	}

	return questions, nil
}

func (r *MongoRepository) DeleteQuestion(ctx context.Context, questionID string) error {
	id, err := primitive.ObjectIDFromHex(questionID)
	if err != nil {
		return err
	}

	result, err := r.questions.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("question not found")
	}

	return nil
}

func (r *MongoRepository) GetQuestionByID(ctx context.Context, questionID string) (*models.Question, error) {
	id, err := primitive.ObjectIDFromHex(questionID)
	if err != nil {
		return nil, err
	}

	var question models.Question
	err = r.questions.FindOne(ctx, bson.M{"_id": id}).Decode(&question)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("question not found")
		}
		return nil, err
	}

	return &question, nil
}

func (r *MongoRepository) GetAnswersByQuestionID(ctx context.Context, questionID string) ([]models.Answer, error) {
	qID, err := primitive.ObjectIDFromHex(questionID)
	if err != nil {
		return nil, err
	}

	var result struct {
		Answers []models.Answer `bson:"answers"`
	}

	// Retrieve only the answers field from the question document
	err = r.questions.FindOne(ctx, bson.M{"_id": qID}, options.FindOne().SetProjection(bson.M{"answers": 1})).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("question not found")
		}
		return nil, err
	}

	return result.Answers, nil
}

func (r *MongoRepository) PostAnswer(ctx context.Context, questionID string, answer *models.Answer) error {
	qID, err := primitive.ObjectIDFromHex(questionID)
	if err != nil {
		return err
	}

	answer.ID = primitive.NewObjectID()
	answer.CreatedAt = time.Now()

	update := bson.M{
		"$push": bson.M{"answers": answer},
	}

	result, err := r.questions.UpdateOne(ctx, bson.M{"_id": qID}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("question not found")
	}

	return nil
}

func (r *MongoRepository) DeleteAnswer(ctx context.Context, questionID, answerID string) error {
	qID, err := primitive.ObjectIDFromHex(questionID)
	if err != nil {
		return err
	}

	aID, err := primitive.ObjectIDFromHex(answerID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$pull": bson.M{
			"answers": bson.M{"_id": aID},
		},
	}

	result, err := r.questions.UpdateOne(ctx, bson.M{"_id": qID}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("question not found")
	}

	return nil
}

func (r *MongoRepository) DownvoteAnswer(ctx context.Context, questionID, answerID string) error {
	qID, err := primitive.ObjectIDFromHex(questionID)
	if err != nil {
		return err
	}

	aID, err := primitive.ObjectIDFromHex(answerID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$inc": bson.M{"answers.$[elem].downvotes": 1},
	}

	arrayFilters := options.Update().SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{bson.M{"elem._id": aID}},
	})

	result, err := r.questions.UpdateOne(ctx, bson.M{"_id": qID}, update, arrayFilters)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("question or answer not found")
	}

	return nil
}

func (r *MongoRepository) FlagQuestion(ctx context.Context, questionID, userID, reason string) error {
	qID, err := primitive.ObjectIDFromHex(questionID)
	if err != nil {
		return err
	}

	flag := models.Flag{
		UserID:    userID,
		Reason:    reason,
		CreatedAt: time.Now(),
	}

	update := bson.M{
		"$push": bson.M{"flags": flag},
		"$set":  bson.M{"is_flagged": true},
	}

	result, err := r.questions.UpdateOne(ctx, bson.M{"_id": qID}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("question not found")
	}

	return nil
}

func (r *MongoRepository) FlagAnswer(ctx context.Context, questionID, answerID, userID, reason string) error {
	qID, err := primitive.ObjectIDFromHex(questionID)
	if err != nil {
		return err
	}

	aID, err := primitive.ObjectIDFromHex(answerID)
	if err != nil {
		return err
	}

	flag := models.Flag{
		UserID:    userID,
		Reason:    reason,
		CreatedAt: time.Now(),
	}

	update := bson.M{
		"$push": bson.M{"answers.$[elem].flags": flag},
		"$set":  bson.M{"answers.$[elem].is_flagged": true},
	}

	arrayFilters := options.Update().SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{bson.M{"elem._id": aID}},
	})

	result, err := r.questions.UpdateOne(ctx, bson.M{"_id": qID}, update, arrayFilters)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("question or answer not found")
	}

	return nil
}

func (r *MongoRepository) MarkQuestionAsAnswered(ctx context.Context, questionID string) error {
	qID, err := primitive.ObjectIDFromHex(questionID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{"is_answered": true},
	}

	result, err := r.questions.UpdateOne(ctx, bson.M{"_id": qID}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("question not found")
	}

	return nil
}

func (r *MongoRepository) GetUserFeed(ctx context.Context, userID string) ([]models.Question, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(50)

	cursor, err := r.questions.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var questions []models.Question
	if err = cursor.All(ctx, &questions); err != nil {
		return nil, err
	}

	return questions, nil
}

func (r *MongoRepository) AddTag(ctx context.Context, tag *models.Tag) error {
	tag.ID = primitive.NewObjectID()
	_, err := r.tags.InsertOne(ctx, tag)
	return err
}

func (r *MongoRepository) RemoveTag(ctx context.Context, tagName string) error {
	result, err := r.tags.DeleteOne(ctx, bson.M{"name": tagName})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("tag not found")
	}

	return nil
}

func (r *MongoRepository) SearchQuestionsAnswersUsers(ctx context.Context, keyword string) (*models.SearchResult, error) {
	// Search in questions
	questionsCursor, err := r.questions.Find(ctx, bson.M{
		"$or": []bson.M{
			{"question": bson.M{"$regex": keyword, "$options": "i"}},
			{"answers.answer": bson.M{"$regex": keyword, "$options": "i"}},
		},
	})
	if err != nil {
		return nil, err
	}
	defer questionsCursor.Close(ctx)

	var questions []models.Question
	if err = questionsCursor.All(ctx, &questions); err != nil {
		return nil, err
	}

	return &models.SearchResult{
		Questions: questions,
	}, nil
}

func (r *MongoRepository) Close(ctx context.Context) error {
	return r.client.Disconnect(ctx)
}

func (r *MongoRepository) GetUserIDFromQuestionID(ctx context.Context, questionID string) (string, error) {
	qID, err := primitive.ObjectIDFromHex(questionID)
	if err != nil {
		return "", err
	}

	var result struct {
		UserID string `bson:"user_id"`
	}

	err = r.questions.FindOne(ctx, bson.M{"_id": qID}, options.FindOne().SetProjection(bson.M{"user_id": 1})).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", errors.New("question not found")
		}
		return "", err
	}

	return result.UserID, nil
}

func (r *MongoRepository) HasUserVotedOnAnswer(ctx context.Context, questionID, answerID, userID string) (bool, string, error) {
	qID, err := primitive.ObjectIDFromHex(questionID)
	if err != nil {
		return false, "", err
	}

	aID, err := primitive.ObjectIDFromHex(answerID)
	if err != nil {
		return false, "", err
	}

	// Find the question and the specific answer
	var question models.Question
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"_id": qID}}},
		{{Key: "$unwind", Value: "$answers"}},
		{{Key: "$match", Value: bson.M{"answers._id": aID}}},
	}

	cursor, err := r.questions.Aggregate(ctx, pipeline)
	if err != nil {
		return false, "", err
	}
	defer cursor.Close(ctx)

	if !cursor.Next(ctx) {
		return false, "", errors.New("answer not found")
	}

	if err := cursor.Decode(&question); err != nil {
		return false, "", err
	}

	// Check if user has voted
	for _, vote := range question.Answers[0].Vote {
		if vote.UserID == userID {
			return true, vote.VoteType, nil
		}
	}

	return false, "", nil
}

func (r *MongoRepository) GetAnswerOwnerID(ctx context.Context, questionID, answerID string) (string, error) {
	qID, err := primitive.ObjectIDFromHex(questionID)
	if err != nil {
		return "", err
	}

	aID, err := primitive.ObjectIDFromHex(answerID)
	if err != nil {
		return "", err
	}

	// Find the question and the specific answer
	var result struct {
		Answers []struct {
			ID     primitive.ObjectID `bson:"_id"`
			UserID string             `bson:"user_id"`
		} `bson:"answers"`
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"_id": qID}}},
		{{Key: "$project", Value: bson.M{
			"answers": bson.M{
				"$filter": bson.M{
					"input": "$answers",
					"as":    "answer",
					"cond":  bson.M{"$eq": []interface{}{"$$answer._id", aID}},
				},
			},
		}}},
	}

	cursor, err := r.questions.Aggregate(ctx, pipeline)
	if err != nil {
		return "", err
	}
	defer cursor.Close(ctx)

	if !cursor.Next(ctx) {
		return "", errors.New("question not found")
	}

	if err := cursor.Decode(&result); err != nil {
		return "", err
	}

	if len(result.Answers) == 0 {
		return "", errors.New("answer not found")
	}

	return result.Answers[0].UserID, nil
}

// Modify the UpvoteAnswer method to include vote tracking
func (r *MongoRepository) UpvoteAnswer(ctx context.Context, questionID, answerID, userID string) error {
	qID, err := primitive.ObjectIDFromHex(questionID)
	if err != nil {
		return err
	}

	aID, err := primitive.ObjectIDFromHex(answerID)
	if err != nil {
		return err
	}

	// Check if user has already voted
	hasVoted, voteType, err := r.HasUserVotedOnAnswer(ctx, questionID, answerID, userID)
	if err != nil {
		return err
	}

	if hasVoted {
		if voteType == "upvote" {
			return errors.New("already upvoted")
		}
		// Remove the previous downvote before adding upvote
		if voteType == "downvote" {
			update := bson.M{
				"$inc":  bson.M{"answers.$[elem].downvotes": -1},
				"$pull": bson.M{"answers.$[elem].voted_by": bson.M{"user_id": userID}},
			}

			arrayFilters := options.Update().SetArrayFilters(options.ArrayFilters{
				Filters: []interface{}{bson.M{"elem._id": aID}},
			})

			if _, err := r.questions.UpdateOne(ctx, bson.M{"_id": qID}, update, arrayFilters); err != nil {
				return err
			}
		}
	}

	// Add upvote and record the vote
	update := bson.M{
		"$inc": bson.M{"answers.$[elem].upvotes": 1},
		"$push": bson.M{
			"answers.$[elem].voted_by": models.Vote{
				UserID:   userID,
				VoteType: "upvote",
				VotedAt:  time.Now(),
			},
		},
	}

	arrayFilters := options.Update().SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{bson.M{"elem._id": aID}},
	})

	result, err := r.questions.UpdateOne(ctx, bson.M{"_id": qID}, update, arrayFilters)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("question or answer not found")
	}

	return nil
}

func (r *MongoRepository) GetFlaggedQuestions(ctx context.Context) ([]models.Question, int32, error) {

	// Create match stage for flagged questions
	match := bson.M{"is_flagged": true}

	// Get total count of flagged questions
	totalCount, err := r.questions.CountDocuments(ctx, match)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.questions.Find(ctx, match, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var questions []models.Question
	if err = cursor.All(ctx, &questions); err != nil {
		return nil, 0, err
	}

	return questions, int32(totalCount), nil
}

func (r *MongoRepository) GetFlaggedAnswers(ctx context.Context) ([]models.Answer, int32, error) {

	// Pipeline to unwind answers and match flagged ones
	pipeline := mongo.Pipeline{
		{{Key: "$unwind", Value: "$answers"}},
		{{Key: "$match", Value: bson.M{"answers.is_flagged": true}}},
		{{Key: "$group", Value: bson.M{
			"_id":     nil,
			"total":   bson.M{"$sum": 1},
			"answers": bson.M{"$push": "$answers"},
		}}},
	}

	// Execute pipeline to get total count and answers
	cursor, err := r.questions.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var result struct {
		Total   int32           `bson:"total"`
		Answers []models.Answer `bson:"answers"`
	}

	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return nil, 0, err
		}
	}

	return result.Answers, result.Total, nil
}
