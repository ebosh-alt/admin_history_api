package misc

import (
	"admin_history/internal/entities"
	protos "admin_history/pkg/proto/gen/go"
	"encoding/json"
	"fmt"
)

func BuildEntityForUpdate(p *protos.Questionnaire) (*entities.Questionnaire, error) {
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
