// File: contentService.go
package service

import (
    "context"
    "errors"
    "time"
    "strings"

    "github.com/liju-github/ContentService/internal/models"
    mongodb "github.com/liju-github/ContentService/internal/repository"
    contentPB "github.com/liju-github/ContentService/proto/content"
)

type ContentService struct {
    contentPB.UnimplementedContentServiceServer
    repo mongodb.Repository
}

func NewContentService(repo mongodb.Repository) *ContentService {
    return &ContentService{
        repo: repo,
    }
}

func (s *ContentService) PostQuestion(ctx context.Context, req *contentPB.PostQuestionRequest) (*contentPB.PostQuestionResponse, error) {
    if err := validatePostQuestion(req); err != nil {
        return &contentPB.PostQuestionResponse{
            Success: false,
            Message: err.Error(),
        }, err
    }

    question := &models.Question{
        UserID:     req.UserID,
        Question:   strings.TrimSpace(req.Question),
        Details:    strings.TrimSpace(req.Details),
        Tags:       sanitizeTags(req.Tags),
        CreatedAt:  time.Now(),
        UpdatedAt:  time.Now(),
        IsAnswered: false,
        IsFlagged:  false,
        Answers:    []models.Answer{},
        Flags:      []models.Flag{},
    }

    err := s.repo.PostQuestion(ctx, question)
    if err != nil {
        return &contentPB.PostQuestionResponse{
            Success: false,
            Message: "Failed to create question: " + err.Error(),
        }, err
    }

    return &contentPB.PostQuestionResponse{
        Success: true,
        Message: "Question created successfully",
    }, nil
}

func (s *ContentService) GetQuestionsByUserID(ctx context.Context, req *contentPB.GetQuestionsByUserIDRequest) (*contentPB.GetQuestionsByUserIDResponse, error) {
    if req.UserID == "" {
        return nil, errors.New("user_id is required")
    }

    questions, err := s.repo.GetQuestionsByUserID(ctx, req.UserID)
    if err != nil {
        return nil, err
    }

    return &contentPB.GetQuestionsByUserIDResponse{
        Questions: convertToProtoQuestions(questions),
    }, nil
}

func (s *ContentService) GetQuestionsByTags(ctx context.Context, req *contentPB.GetQuestionsByTagsRequest) (*contentPB.GetQuestionsByTagsResponse, error) {
    if len(req.Tags) == 0 {
        return nil, errors.New("at least one tag is required")
    }

    questions, err := s.repo.GetQuestionsByTags(ctx, sanitizeTags(req.Tags))
    if err != nil {
        return nil, err
    }

    return &contentPB.GetQuestionsByTagsResponse{
        Questions: convertToProtoQuestions(questions),
    }, nil
}

func (s *ContentService) GetQuestionsByWord(ctx context.Context, req *contentPB.GetQuestionsByWordRequest) (*contentPB.GetQuestionsByWordResponse, error) {
    if req.SearchWord == "" {
        return nil, errors.New("search word is required")
    }

    questions, err := s.repo.GetQuestionsByWord(ctx, strings.TrimSpace(req.SearchWord))
    if err != nil {
        return nil, err
    }

    return &contentPB.GetQuestionsByWordResponse{
        Questions: convertToProtoQuestions(questions),
    }, nil
}

func (s *ContentService) DeleteQuestion(ctx context.Context, req *contentPB.DeleteQuestionRequest) (*contentPB.DeleteQuestionResponse, error) {
    if req.QuestionID == "" || req.UserID == "" {
        return &contentPB.DeleteQuestionResponse{
            Success: false,
            Message: "question_id and user_id are required",
        }, errors.New("question_id and user_id are required")
    }

    // Verify user owns the question
    questionOwnerID, err := s.repo.GetUserIDFromQuestionID(ctx, req.QuestionID)
    if err != nil {
        return &contentPB.DeleteQuestionResponse{
            Success: false,
            Message: "Failed to verify question ownership",
        }, err
    }

    if questionOwnerID != req.UserID {
        return &contentPB.DeleteQuestionResponse{
            Success: false,
            Message: "Unauthorized: only question owner can delete the question",
        }, errors.New("unauthorized deletion attempt")
    }

    err = s.repo.DeleteQuestion(ctx, req.QuestionID)
    if err != nil {
        return &contentPB.DeleteQuestionResponse{
            Success: false,
            Message: "Failed to delete question: " + err.Error(),
        }, err
    }

    return &contentPB.DeleteQuestionResponse{
        Success: true,
        Message: "Question deleted successfully",
    }, nil
}

func (s *ContentService) PostAnswer(ctx context.Context, req *contentPB.PostAnswerByQuestionIDRequest) (*contentPB.PostAnswerByQuestionIDResponse, error) {
    if err := validatePostAnswer(req); err != nil {
        return &contentPB.PostAnswerByQuestionIDResponse{
            Success: false,
            Message: err.Error(),
        }, err
    }

    // Check if user is trying to answer their own question
    questionOwnerID, err := s.repo.GetUserIDFromQuestionID(ctx, req.QuestionID)
    if err != nil {
        return &contentPB.PostAnswerByQuestionIDResponse{
            Success: false,
            Message: "Failed to verify question ownership",
        }, err
    }

    if questionOwnerID == req.UserID {
        return &contentPB.PostAnswerByQuestionIDResponse{
            Success: false,
            Message: "Cannot answer your own question",
        }, errors.New("self-answering not allowed")
    }

    answer := &models.Answer{
        UserID:    req.UserID,
        Answer:    strings.TrimSpace(req.Answer),
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
        Upvotes:   0,
        Downvotes: 0,
        IsFlagged: false,
        Vote:   []models.Vote{},
        Flags:     []models.Flag{},
    }

    err = s.repo.PostAnswer(ctx, req.QuestionID, answer)
    if err != nil {
        return &contentPB.PostAnswerByQuestionIDResponse{
            Success: false,
            Message: "Failed to post answer: " + err.Error(),
        }, err
    }

    return &contentPB.PostAnswerByQuestionIDResponse{
        Success: true,
        Message: "Answer posted successfully",
    }, nil
}

func (s *ContentService) UpvoteAnswer(ctx context.Context, req *contentPB.UpvoteAnswerByAnswerIDRequest) (*contentPB.UpvoteAnswerByAnswerIDResponse, error) {
    if req.AnswerID == "" || req.QuestionId == "" || req.UserID == "" {
        return &contentPB.UpvoteAnswerByAnswerIDResponse{
            Success: false,
            Message: "answer_id, question_id, and user_id are required",
        }, errors.New("missing required fields")
    }

    // Check if user has already voted
    hasVoted, voteType, err := s.repo.HasUserVotedOnAnswer(ctx, req.QuestionId, req.AnswerID, req.UserID)
    if err != nil {
        return &contentPB.UpvoteAnswerByAnswerIDResponse{
            Success: false,
            Message: "Failed to check voting status",
        }, err
    }

    if hasVoted && voteType == "upvote" {
        return &contentPB.UpvoteAnswerByAnswerIDResponse{
            Success: false,
            Message: "User has already upvoted this answer",
        }, errors.New("already upvoted")
    }

    // Check if user is trying to vote on their own answer
    answerOwnerID, err := s.repo.GetAnswerOwnerID(ctx, req.QuestionId, req.AnswerID)
    if err != nil {
        return &contentPB.UpvoteAnswerByAnswerIDResponse{
            Success: false,
            Message: "Failed to verify answer ownership",
        }, err
    }

    if answerOwnerID == req.UserID {
        return &contentPB.UpvoteAnswerByAnswerIDResponse{
            Success: false,
            Message: "Cannot vote on your own answer",
        }, errors.New("self-voting not allowed")
    }

    err = s.repo.UpvoteAnswer(ctx, req.QuestionId, req.AnswerID, req.UserID)
    if err != nil {
        return &contentPB.UpvoteAnswerByAnswerIDResponse{
            Success: false,
            Message: "Failed to upvote answer: " + err.Error(),
        }, err
    }

    return &contentPB.UpvoteAnswerByAnswerIDResponse{
        Success: true,
        Message: "Answer upvoted successfully",
    }, nil
}

// Helper functions

func validatePostQuestion(req *contentPB.PostQuestionRequest) error {
    if req.Question == "" || req.UserID == "" {
        return errors.New("question and user_id are required")
    }

    if len(req.Question) < 10 {
        return errors.New("question must be at least 10 characters long")
    }

    if len(req.Tags) > 5 {
        return errors.New("maximum 5 tags allowed")
    }

    return nil
}

func validatePostAnswer(req *contentPB.PostAnswerByQuestionIDRequest) error {
    if req.QuestionID == "" || req.Answer == "" || req.UserID == "" {
        return errors.New("question_id, answer, and user_id are required")
    }

    if len(req.Answer) < 20 {
        return errors.New("answer must be at least 20 characters long")
    }

    return nil
}

func sanitizeTags(tags []string) []string {
    sanitized := make([]string, 0, len(tags))
    for _, tag := range tags {
        tag = strings.TrimSpace(strings.ToLower(tag))
        if tag != "" && len(tag) <= 20 {
            sanitized = append(sanitized, tag)
        }
    }
    return sanitized
}

func convertToProtoQuestions(questions []models.Question) []*contentPB.Question {
    protoQuestions := make([]*contentPB.Question, len(questions))
    for i, q := range questions {
        protoQuestions[i] = &contentPB.Question{
            QuestionID: q.ID.Hex(),
            Question:   q.Question,
            UserID:     q.UserID,
            CreatedAt:  q.CreatedAt.Unix(),
            Tags:       q.Tags,
            IsAnswered: q.IsAnswered,
            Details:    q.Details,
        }
    }
    return protoQuestions
}