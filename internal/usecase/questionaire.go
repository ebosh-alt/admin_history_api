package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"admin_history/internal/entities"
	protos "admin_history/pkg/proto/gen/go"
)

func (u *Usecase) GetQuestionnairesList(ctx context.Context, req *protos.QuestionnairesListRequest) (*protos.QuestionnairesListResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("nil request")
	}
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}

	items, err := u.Postgres.GetQuestionnairesList(ctx, req.Page, req.Limit)
	if err != nil {
		return nil, err
	}

	resp := &protos.QuestionnairesListResponse{
		Questionnaires: make([]*protos.Questionnaire, 0, len(items)),
	}
	for i := range items {
		resp.Questionnaires = append(resp.Questionnaires, items[i].ToProto())
	}
	return resp, nil
}

func (u *Usecase) GetQuestionnaire(ctx context.Context, req *protos.QuestionnaireRequest) (*protos.QuestionnaireResponse, error) {
	if req == nil || req.Id <= 0 {
		return nil, fmt.Errorf("bad id")
	}
	q, err := u.Postgres.GetQuestionnaire(ctx, &entities.Questionnaire{ID: req.Id})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}
	return &protos.QuestionnaireResponse{Questionnaire: q.ToProto()}, nil
}

func (u *Usecase) UpdateQuestionnaire(ctx context.Context, req *protos.UpdateQuestionnaireRequest) (*protos.Status, error) {
	if req == nil || req.Questionnaire == nil || req.Questionnaire.Id == 0 {
		return nil, fmt.Errorf("invalid request")
	}

	ent, err := buildEntityForUpdate(req.Questionnaire)
	if err != nil {
		return nil, err
	}

	if err := u.Postgres.UpdateQuestionnaire(ctx, ent); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &protos.Status{Ok: false, Message: "questionnaire not found"}, nil
		}
		return nil, err
	}

	return &protos.Status{Ok: true, Message: "updated"}, nil
}

func buildEntityForUpdate(p *protos.Questionnaire) (*entities.Questionnaire, error) {
	q := &entities.Questionnaire{
		ID:      p.Id,
		UserID:  p.UserId,
		History: p.History,
		Status:  p.Status,
		Payment: p.Payment,
	}

	if p.Answers != nil {
		b, err := json.Marshal(p.Answers)
		if err != nil {
			return nil, fmt.Errorf("marshal answers: %w", err)
		}
		q.Answers = b
	}

	return q, nil
}
