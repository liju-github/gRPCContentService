package service

import (
	"context"
	"errors"
	"time"

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

// PostQuestion implements the PostQuestion RPC method
func (s *ContentService) PostQuestion(ctx context.Context, req *contentPB.PostQuestionRequest) (*contentPB.PostQuestionResponse, error) {
	if req.Question == "" || req.UserID == "" {
		return &contentPB.PostQuestionResponse{
			Success: false,
			Message: "question and user_id are required",
		}, errors.New("question and user_id are required")
	}

	question := &models.Question{
		UserID:     req.UserID,
		Question:   req.Question,
		Tags:       req.Tags,
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

// GetQuestionsByUserID implements the GetQuestionsByUserID RPC method
func (s *ContentService) GetQuestionsByUserID(ctx context.Context, req *contentPB.GetQuestionsByUserIDRequest) (*contentPB.GetQuestionsByUserIDResponse, error) {
	if req.UserID == "" {
		return nil, errors.New("user_id is required")
	}

	questions, err := s.repo.GetQuestionsByUserID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	pbQuestions := make([]*contentPB.Question, len(questions))
	for i, q := range questions {
		pbQuestions[i] = &contentPB.Question{
			QuestionID: q.ID.Hex(),
			Question:   q.Question,
			UserID:     q.UserID,
			CreatedAt:  q.CreatedAt.Unix(),
			Tags:       q.Tags,
			IsAnswered: q.IsAnswered,
		}
	}

	return &contentPB.GetQuestionsByUserIDResponse{
		Questions: pbQuestions,
	}, nil
}

// GetQuestionsByTags implements the GetQuestionsByTags RPC method
func (s *ContentService) GetQuestionsByTags(ctx context.Context, req *contentPB.GetQuestionsByTagsRequest) (*contentPB.GetQuestionsByTagsResponse, error) {
	if len(req.Tags) == 0 {
		return nil, errors.New("at least one tag is required")
	}

	questions, err := s.repo.GetQuestionsByTags(ctx, req.Tags)
	if err != nil {
		return nil, err
	}

	pbQuestions := make([]*contentPB.Question, len(questions))
	for i, q := range questions {
		pbQuestions[i] = &contentPB.Question{
			QuestionID: q.ID.Hex(),
			Question:   q.Question,
			UserID:     q.UserID,
			CreatedAt:  q.CreatedAt.Unix(),
			Tags:       q.Tags,
			IsAnswered: q.IsAnswered,
		}
	}

	return &contentPB.GetQuestionsByTagsResponse{
		Questions: pbQuestions,
	}, nil
}

// GetQuestionsByWord implements the GetQuestionsByWord RPC method
func (s *ContentService) GetQuestionsByWord(ctx context.Context, req *contentPB.GetQuestionsByWordRequest) (*contentPB.GetQuestionsByWordResponse, error) {
	if req.SearchWord == "" {
		return nil, errors.New("search word is required")
	}

	questions, err := s.repo.GetQuestionsByWord(ctx, req.SearchWord)
	if err != nil {
		return nil, err
	}

	pbQuestions := make([]*contentPB.Question, len(questions))
	for i, q := range questions {
		pbQuestions[i] = &contentPB.Question{
			QuestionID: q.ID.Hex(),
			Question:   q.Question,
			UserID:     q.UserID,
			CreatedAt:  q.CreatedAt.Unix(),
			Tags:       q.Tags,
			IsAnswered: q.IsAnswered,
		}
	}

	return &contentPB.GetQuestionsByWordResponse{
		Questions: pbQuestions,
	}, nil
}

// DeleteQuestion implements the DeleteQuestion RPC method
func (s *ContentService) DeleteQuestion(ctx context.Context, req *contentPB.DeleteQuestionRequest) (*contentPB.DeleteQuestionResponse, error) {
	if req.QuestionID == "" {
		return &contentPB.DeleteQuestionResponse{
			Success: false,
			Message: "question_id is required",
		}, errors.New("question_id is required")
	}

	err := s.repo.DeleteQuestion(ctx, req.QuestionID)
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

// GetQuestionByID implements the GetQuestionByID RPC method
func (s *ContentService) GetQuestionByID(ctx context.Context, req *contentPB.GetQuestionByIDRequest) (*contentPB.GetQuestionByIDResponse, error) {
	if req.QuestionID == "" {
		return nil, errors.New("question_id is required")
	}

	question, err := s.repo.GetQuestionByID(ctx, req.QuestionID)
	if err != nil {
		return nil, err
	}

	pbQuestion := &contentPB.Question{
		QuestionID: question.ID.Hex(),
		Question:   question.Question,
		UserID:     question.UserID,
		CreatedAt:  question.CreatedAt.Unix(),
		Tags:       question.Tags,
		IsAnswered: question.IsAnswered,
	}

	return &contentPB.GetQuestionByIDResponse{
		Question: pbQuestion,
	}, nil
}

// PostAnswerByQuestionID implements the PostAnswerByQuestionID RPC method
func (s *ContentService) PostAnswerByQuestionID(ctx context.Context, req *contentPB.PostAnswerByQuestionIDRequest) (*contentPB.PostAnswerByQuestionIDResponse, error) {
	if req.QuestionID == "" || req.Answer == "" || req.UserID == "" {
		return &contentPB.PostAnswerByQuestionIDResponse{
			Success: false,
			Message: "question_id, answer, and user_id are required",
		}, errors.New("question_id, answer, and user_id are required")
	}

	answer := &models.Answer{
		UserID:    req.UserID,
		Answer:    req.Answer,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Upvotes:   0,
		Downvotes: 0,
		IsFlagged: false,
		Flags:     []models.Flag{},
	}

	err := s.repo.PostAnswer(ctx, req.QuestionID, answer)
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

// DeleteAnswerByAnswerID implements the DeleteAnswerByAnswerID RPC method
func (s *ContentService) DeleteAnswerByAnswerID(ctx context.Context, req *contentPB.DeleteAnswerByAnswerIDRequest) (*contentPB.DeleteAnswerByAnswerIDResponse, error) {
	if req.AnswerID == "" {
		return &contentPB.DeleteAnswerByAnswerIDResponse{
			Success: false,
			Message: "answer_id is required",
		}, errors.New("answer_id is required")
	}

	err := s.repo.DeleteAnswer(ctx, req.QuestionID, req.AnswerID)
	if err != nil {
		return &contentPB.DeleteAnswerByAnswerIDResponse{
			Success: false,
			Message: "Failed to delete answer: " + err.Error(),
		}, err
	}

	return &contentPB.DeleteAnswerByAnswerIDResponse{
		Success: true,
		Message: "Answer deleted successfully",
	}, nil
}

// UpvoteAnswerByAnswerID implements the UpvoteAnswerByAnswerID RPC method
func (s *ContentService) UpvoteAnswerByAnswerID(ctx context.Context, req *contentPB.UpvoteAnswerByAnswerIDRequest) (*contentPB.UpvoteAnswerByAnswerIDResponse, error) {
	if req.AnswerID == "" {
		return &contentPB.UpvoteAnswerByAnswerIDResponse{
			Success: false,
			Message: "answer_id is required",
		}, errors.New("answer_id is required")
	}

	err := s.repo.UpvoteAnswer(ctx, req.QuestionId, req.AnswerID)
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

// DownvoteAnswerByAnswerID implements the DownvoteAnswerByAnswerID RPC method
func (s *ContentService) DownvoteAnswerByAnswerID(ctx context.Context, req *contentPB.DownvoteAnswerByAnswerIDRequest) (*contentPB.DownvoteAnswerByAnswerIDResponse, error) {
	if req.QuestionID == "" && req.AnswerID == "" {
		return &contentPB.DownvoteAnswerByAnswerIDResponse{
			Success: false,
			Message: "answer_id is required",
		}, errors.New("answer_id is required")
	}

	err := s.repo.DownvoteAnswer(ctx, req.QuestionID, req.AnswerID)
	if err != nil {
		return &contentPB.DownvoteAnswerByAnswerIDResponse{
			Success: false,
			Message: "Failed to downvote answer: " + err.Error(),
		}, err
	}

	return &contentPB.DownvoteAnswerByAnswerIDResponse{
		Success: true,
		Message: "Answer downvoted successfully",
	}, nil
}

// FlagQuestion implements the FlagQuestion RPC method
func (s *ContentService) FlagQuestion(ctx context.Context, req *contentPB.FlagQuestionRequest) (*contentPB.FlagQuestionResponse, error) {
	if req.QuestionID == "" || req.UserID == "" || req.Reason == "" {
		return &contentPB.FlagQuestionResponse{
			Success: false,
			Message: "question_id, user_id, and reason are required",
		}, errors.New("question_id, user_id, and reason are required")
	}

	err := s.repo.FlagQuestion(ctx, req.QuestionID, req.UserID, req.Reason)
	if err != nil {
		return &contentPB.FlagQuestionResponse{
			Success: false,
			Message: "Failed to flag question: " + err.Error(),
		}, err
	}

	return &contentPB.FlagQuestionResponse{
		Success: true,
		Message: "Question flagged successfully",
	}, nil
}

// FlagAnswer implements the FlagAnswer RPC method
func (s *ContentService) FlagAnswer(ctx context.Context, req *contentPB.FlagAnswerRequest) (*contentPB.FlagAnswerResponse, error) {
	if req.AnswerID == "" || req.UserID == "" || req.Reason == "" {
		return &contentPB.FlagAnswerResponse{
			Success: false,
			Message: "answer_id, user_id, and reason are required",
		}, errors.New("answer_id, user_id, and reason are required")
	}

	err := s.repo.FlagAnswer(ctx, req.QuestionID, req.AnswerID, req.UserID, req.AnswerID)
	if err != nil {
		return &contentPB.FlagAnswerResponse{
			Success: false,
			Message: "Failed to flag answer: " + err.Error(),
		}, err
	}

	return &contentPB.FlagAnswerResponse{
		Success: true,
		Message: "Answer flagged successfully",
	}, nil
}

// MarkQuestionAsAnswered implements the MarkQuestionAsAnswered RPC method
func (s *ContentService) MarkQuestionAsAnswered(ctx context.Context, req *contentPB.MarkQuestionAsAnsweredRequest) (*contentPB.MarkQuestionAsAnsweredResponse, error) {
	if req.QuestionID == "" {
		return &contentPB.MarkQuestionAsAnsweredResponse{
			Success: false,
			Message: "question_id is required",
		}, errors.New("question_id is required")
	}

	err := s.repo.MarkQuestionAsAnswered(ctx, req.QuestionID)
	if err != nil {
		return &contentPB.MarkQuestionAsAnsweredResponse{
			Success: false,
			Message: "Failed to mark question as answered: " + err.Error(),
		}, err
	}

	return &contentPB.MarkQuestionAsAnsweredResponse{
		Success: true,
		Message: "Question marked as answered successfully",
	}, nil
}

// GetUserFeed implements the GetUserFeed RPC method
func (s *ContentService) GetUserFeed(ctx context.Context, req *contentPB.GetUserFeedRequest) (*contentPB.GetUserFeedResponse, error) {
	if req.UserID == "" {
		return nil, errors.New("user_id is required")
	}

	questions, err := s.repo.GetUserFeed(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	pbQuestions := make([]*contentPB.Question, len(questions))
	for i, q := range questions {
		pbQuestions[i] = &contentPB.Question{
			QuestionID: q.ID.Hex(),
			Question:   q.Question,
			UserID:     q.UserID,
			CreatedAt:  q.CreatedAt.Unix(),
			Tags:       q.Tags,
			IsAnswered: q.IsAnswered,
		}
	}

	return &contentPB.GetUserFeedResponse{
		Questions: pbQuestions,
	}, nil
}
