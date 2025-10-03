package usecase

import (
	"admin_history/internal/misc"
	"context"
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

	var f entities.QuestionnaireFilter
	if req.Payment != nil {
		v := req.Payment.Value
		f.Payment = &v
	}
	if req.Status != nil {
		v := req.Status.Value
		f.Status = &v
	}
	if req.DateFrom != nil {
		t := req.DateFrom.AsTime()
		f.DateFrom = &t
	}
	if req.DateTo != nil {
		t := req.DateTo.AsTime()
		f.DateTo = &t
	} // полузакрытый интервал: < date_to

	items, err := u.Postgres.GetQuestionnairesList(ctx, req.Page, req.Limit, f)
	if err != nil {
		return nil, err
	}

	resp := &protos.QuestionnairesListResponse{
		Questionnaires: make([]*protos.Questionnaire, 0, len(items)),
		Total:          int64(len(items)),
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

	ent, err := misc.BuildEntityForUpdate(req.Questionnaire)
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
