package entities

import (
	protos "admin_history/pkg/proto/gen/go"
	"encoding/json"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type User struct {
	ID            int64
	Username      string
	Status        bool
	AcceptedOffer bool
	CreatedAt     time.Time
	QTotal        int64
	QPaid         int64
	QUnpaid       int64
}

type UserDTO struct {
	ID            *int64
	Username      *string
	Status        *bool
	AcceptedOffer *bool
	CreatedAt     *time.Time
	QTotal        *int64
	QPaid         *int64
	QUnpaid       *int64
}

func (u *User) ToDTO() *UserDTO {
	return &UserDTO{
		ID:            &u.ID,
		Username:      &u.Username,
		Status:        &u.Status,
		AcceptedOffer: &u.AcceptedOffer,
		CreatedAt:     &u.CreatedAt,
		QTotal:        &u.QTotal,
		QPaid:         &u.QPaid,
		QUnpaid:       &u.QUnpaid,
	}
}

func (d *UserDTO) ToEntity() *User {
	u := &User{}
	if d.ID != nil {
		u.ID = *d.ID
	}
	if d.Username != nil {
		u.Username = *d.Username
	}
	if d.Status != nil {
		u.Status = *d.Status
	}
	if d.AcceptedOffer != nil {
		u.AcceptedOffer = *d.AcceptedOffer
	}
	if d.CreatedAt != nil {
		u.CreatedAt = *d.CreatedAt
	}
	if d.QTotal != nil {
		u.QTotal = *d.QTotal
	}
	if d.QPaid != nil {
		u.QPaid = *d.QPaid
	}
	if d.QUnpaid != nil {
		u.QUnpaid = *d.QUnpaid
	}
	return u
}

type Photo struct {
	ID              int64
	QuestionnaireID int64
	Path            string
	Scene           string
	TypePhoto       string
}

type PhotoDTO struct {
	ID              *int64
	QuestionnaireID *int64
	Path            *string
	Scene           *string
	TypePhoto       *string
}

func (p *Photo) ToDTO() *PhotoDTO {
	return &PhotoDTO{
		ID:              &p.ID,
		QuestionnaireID: &p.QuestionnaireID,
		Path:            &p.Path,
		Scene:           &p.Scene,
		TypePhoto:       &p.TypePhoto,
	}
}

func (d *PhotoDTO) ToEntity() *Photo {
	p := &Photo{}
	if d.ID != nil {
		p.ID = *d.ID
	}
	if d.QuestionnaireID != nil {
		d.QuestionnaireID = &p.QuestionnaireID
	}
	if d.Path != nil {
		d.Path = &p.Path
	}
	if d.Scene != nil {
		d.Scene = &p.Scene
	}
	if d.TypePhoto != nil {
		d.TypePhoto = &p.TypePhoto
	}
	return p
}

type Questionnaire struct {
	ID        int64
	UserID    int64
	History   string
	Status    bool
	Payment   bool
	CreatedAt time.Time

	Answers   []byte
	Photos    []Photo
	GenPhotos []Photo
}
type QuestionnaireDTO struct {
	ID        *int64
	UserID    *int64
	Answers   *[]byte
	History   *string
	Status    *bool
	Payment   *bool
	CreatedAt *time.Time

	Photos    *[]Photo
	GenPhotos *[]Photo
}

func (q *Questionnaire) ToDTO() *QuestionnaireDTO {
	return &QuestionnaireDTO{
		ID:        &q.ID,
		UserID:    &q.UserID,
		Answers:   &q.Answers,
		History:   &q.History,
		Status:    &q.Status,
		Payment:   &q.Payment,
		Photos:    &q.Photos,
		GenPhotos: &q.GenPhotos,
	}
}

func (q *Questionnaire) ToProto() *protos.Questionnaire {
	out := &protos.Questionnaire{
		Id:      q.ID,
		UserId:  q.UserID,
		History: q.History,
		Status:  q.Status,
		Payment: q.Payment,
	}

	if !q.CreatedAt.IsZero() && !q.CreatedAt.Equal(time.Time{}) {
		out.CreatedAt = timestamppb.New(q.CreatedAt)
	}

	if len(q.Answers) > 0 {
		var arr []*protos.Answer
		if err := json.Unmarshal(q.Answers, &arr); err == nil {
			out.Answers = arr
		}
	}

	return out
}
func (d *QuestionnaireDTO) ToEntity() *Questionnaire {
	var q Questionnaire
	if d.ID != nil {
		q.ID = *d.ID
	}
	if d.UserID != nil {
		q.UserID = *d.UserID
	}
	if d.History != nil {
		q.History = *d.History
	}
	if d.Status != nil {
		q.Status = *d.Status
	}
	if d.Payment != nil {
		q.Payment = *d.Payment
	}
	if d.CreatedAt != nil {
		q.CreatedAt = *d.CreatedAt
	}
	if d.Answers == nil {
		q.Answers = append([]byte(nil), *d.Answers...)
	}
	if d.Photos != nil {
		q.Photos = append([]Photo(nil), *d.Photos...)
	}
	if d.GenPhotos != nil {
		q.GenPhotos = append([]Photo(nil), *d.GenPhotos...)
	}

	return &q
}

type QuestionnaireFilter struct {
	Payment  *bool
	Status   *bool
	DateFrom *time.Time
	DateTo   *time.Time
}

type UsersFilter struct {
	Status        *bool
	AcceptedOffer *bool
	DateFrom      *time.Time
	DateTo        *time.Time
}
