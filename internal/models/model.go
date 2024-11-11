package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MongoConfig struct {
	URI      string
	Database string
}

type Question struct {
	ID         primitive.ObjectID `bson:"_id" json:"id"`
	UserID     string             `bson:"user_id" json:"user_id"`
	Question   string             `bson:"question" json:"question"`
	Details    string             `bson:"details" json:"details"`
	Tags       []string           `bson:"tags" json:"tags"`
	Answers    []Answer           `bson:"answers" json:"answers"`
	IsAnswered bool               `bson:"is_answered" json:"is_answered"`
	IsFlagged  bool               `bson:"is_flagged" json:"is_flagged"`
	Flags      []Flag             `bson:"flags" json:"flags"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}

type Answer struct {
	ID         primitive.ObjectID `bson:"_id" json:"id"`
	QuestionID primitive.ObjectID `bson:"question_id" json:"question_id"`
	UserID     string             `bson:"user_id" json:"user_id"`
	Answer     string             `bson:"answer" json:"answer"`
	Upvotes    int                `bson:"upvotes" json:"upvotes"`
	Downvotes  int                `bson:"downvotes" json:"downvotes"`
	IsFlagged  bool               `bson:"is_flagged" json:"is_flagged"`
	Flags      []Flag             `bson:"flags" json:"flags"`
	Vote       []Vote             `bson:"votes" json:"votes"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}

type Flag struct {
	UserID    string    `bson:"user_id" json:"user_id"`
	Reason    string    `bson:"reason" json:"reason"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}

type Tag struct {
	ID          primitive.ObjectID `bson:"_id" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
}

type SearchResult struct {
	Questions []Question `json:"questions"`
}

type Vote struct {
	UserID   string    `bson:"user_id"`
	VoteType string    `bson:"vote_type"`
	VotedAt  time.Time `bson:"voted_at"`
}
