package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
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

func (s *ContentService) GetQuestionByID(ctx context.Context, req *contentPB.GetQuestionByIDRequest) (*contentPB.GetQuestionByIDResponse, error) {
	if req.QuestionID == "" {
		return nil, errors.New("question_id is required")
	}

	question, err := s.repo.GetQuestionByID(ctx, req.QuestionID)
	if err != nil {
		return nil, err
	}

	pbAnswers := make([]*contentPB.Answer, len(question.Answers))
	for i, answer := range question.Answers {
		pbAnswers[i] = &contentPB.Answer{
			Id:         answer.ID.Hex(),
			QuestionId: question.ID.Hex(),
			UserId:     answer.UserID,
			AnswerText: answer.Answer,
			Upvotes:    int32(answer.Upvotes),
			Downvotes:  int32(answer.Downvotes),
			IsFlagged:  answer.IsFlagged,
			CreatedAt:  answer.CreatedAt.Unix(),
			UpdatedAt:  answer.UpdatedAt.Unix(),
		}
	}

	pbQuestion := &contentPB.Question{
		QuestionID: question.ID.Hex(),
		Question:   question.Question,
		UserID:     question.UserID,
		CreatedAt:  question.CreatedAt.Unix(),
		Tags:       question.Tags,
		IsAnswered: question.IsAnswered,
		Details:    question.Details,
	}

	return &contentPB.GetQuestionByIDResponse{
		Question: pbQuestion,
		Answers:  pbAnswers,
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

func (s *ContentService) PostAnswerByQuestionID(ctx context.Context, req *contentPB.PostAnswerByQuestionIDRequest) (*contentPB.PostAnswerByQuestionIDResponse, error) {
	if err := validatePostAnswer(req); err != nil {
		return &contentPB.PostAnswerByQuestionIDResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	// // Check if user is trying to answer their own question
	// questionOwnerID, err := s.repo.GetUserIDFromQuestionID(ctx, req.QuestionID)
	// if err != nil {
	//     return &contentPB.PostAnswerByQuestionIDResponse{
	//         Success: false,
	//         Message: "Failed to verify question ownership",
	//     }, err
	// }

	// if questionOwnerID == req.UserID {
	//     return &contentPB.PostAnswerByQuestionIDResponse{
	//         Success: false,
	//         Message: "Cannot answer your own question",
	//     }, errors.New("self-answering not allowed")
	// }

	answer := &models.Answer{
		UserID:    req.UserID,
		Answer:    strings.TrimSpace(req.Answer),
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

func (s *ContentService) DeleteAnswerByAnswerID(ctx context.Context, req *contentPB.DeleteAnswerByAnswerIDRequest) (*contentPB.DeleteAnswerByAnswerIDResponse, error) {
	if req.AnswerID == "" || req.QuestionID == "" {
		return &contentPB.DeleteAnswerByAnswerIDResponse{
			Success: false,
			Message: "answer_id and question_id are required",
		}, errors.New("answer_id and question_id are required")
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

func (s *ContentService) FlagAnswer(ctx context.Context, req *contentPB.FlagAnswerRequest) (*contentPB.FlagAnswerResponse, error) {
	if req.AnswerID == "" || req.UserID == "" || req.Reason == "" || req.QuestionID == "" {
		return &contentPB.FlagAnswerResponse{
			Success: false,
			Message: "answer_id, question_id, user_id, and reason are required",
		}, errors.New("answer_id, question_id, user_id, and reason are required")
	}

	err := s.repo.FlagAnswer(ctx, req.QuestionID, req.AnswerID, req.UserID, req.Reason)
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

func (s *ContentService) GetUserFeed(ctx context.Context, req *contentPB.GetUserFeedRequest) (*contentPB.GetUserFeedResponse, error) {
	questions, err := s.repo.GetUserFeed(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	return &contentPB.GetUserFeedResponse{
		Questions: convertToProtoQuestions(questions),
	}, nil
}

func (s *ContentService) GetFlaggedQuestions(ctx context.Context,req *contentPB.GetFlaggedQuestionsRequest) (*contentPB.GetFlaggedQuestionsResponse, error) {

	questions, totalCount, err := s.repo.GetFlaggedQuestions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get flagged questions: %v", err)
	}

	return &contentPB.GetFlaggedQuestionsResponse{
		FlaggedQuestions:      convertToProtoQuestions(questions),
		TotalFlaggedQuestions: totalCount,
	}, nil
}

func (s *ContentService) GetFlaggedAnswers(ctx context.Context,req *contentPB.GetFlaggedAnswersRequest) (*contentPB.GetFlaggedAnswersResponse, error) {

	answers, totalCount, err := s.repo.GetFlaggedAnswers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get flagged answers: %v", err)
	}

	log.Println("answers and totalcount",answers,totalCount)

	protoAnswers := make([]*contentPB.Answer, len(answers))
	for i, answer := range answers {
		protoAnswers[i] = &contentPB.Answer{
			Id:         answer.ID.Hex(),
			UserId:     answer.UserID,
			AnswerText: answer.Answer,
			Upvotes:    int32(answer.Upvotes),
			Downvotes:  int32(answer.Downvotes),
			IsFlagged:  answer.IsFlagged,
			CreatedAt:  answer.CreatedAt.Unix(),
			UpdatedAt:  answer.UpdatedAt.Unix(),
		}
	}

	return &contentPB.GetFlaggedAnswersResponse{
		FlaggedAnswers:      protoAnswers,
		TotalFlaggedAnswers: totalCount,
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
		if tag != "" && len(tag) <= 10 {
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
