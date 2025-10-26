package entities

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"time"

	protos "admin_history/pkg/proto/gen/go"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type User struct {
	ID            int64
	Username      string
	Language      *string
	RefBossID     *int64
	Status        bool
	AcceptedOffer bool
	CreatedAt     time.Time
	Promocode     *string
	Age           *int64
	Gender        *string
	MapBinding    *bool
	QTotal        int64
	QPaid         int64
	QUnpaid       int64
}

type UserDTO struct {
	ID            *int64
	Username      *string
	Language      *string
	RefBossID     *int64
	Status        *bool
	AcceptedOffer *bool
	CreatedAt     *time.Time
	Promocode     *string
	Age           *int64
	Gender        *string
	MapBinding    *bool
	QTotal        *int64
	QPaid         *int64
	QUnpaid       *int64
}

func (u *User) ToDTO() *UserDTO {
	return &UserDTO{
		ID:            &u.ID,
		Username:      &u.Username,
		Language:      u.Language,
		RefBossID:     u.RefBossID,
		Status:        &u.Status,
		AcceptedOffer: &u.AcceptedOffer,
		CreatedAt:     &u.CreatedAt,
		Promocode:     u.Promocode,
		Age:           u.Age,
		Gender:        u.Gender,
		MapBinding:    u.MapBinding,
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
	if d.Language != nil {
		u.Language = d.Language
	}
	if d.RefBossID != nil {
		u.RefBossID = d.RefBossID
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
	if d.Promocode != nil {
		u.Promocode = d.Promocode
	}
	if d.Age != nil {
		u.Age = d.Age
	}
	if d.Gender != nil {
		u.Gender = d.Gender
	}
	if d.MapBinding != nil {
		u.MapBinding = d.MapBinding
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

type Video struct {
	ID              int64
	QuestionnaireID int64
	Path            string
	TypeVideo       string
	CreatedAt       time.Time
}

type VideoDTO struct {
	ID              *int64
	QuestionnaireID *int64
	Path            *string
	TypeVideo       *string
	CreatedAt       *time.Time
}

func (v *Video) ToDTO() *VideoDTO {
	return &VideoDTO{
		ID:              &v.ID,
		QuestionnaireID: &v.QuestionnaireID,
		Path:            &v.Path,
		TypeVideo:       &v.TypeVideo,
		CreatedAt:       &v.CreatedAt,
	}
}

func (d *VideoDTO) ToEntity() *Video {
	v := &Video{}
	if d.ID != nil {
		v.ID = *d.ID
	}
	if d.QuestionnaireID != nil {
		v.QuestionnaireID = *d.QuestionnaireID
	}
	if d.Path != nil {
		v.Path = *d.Path
	}
	if d.TypeVideo != nil {
		v.TypeVideo = *d.TypeVideo
	}
	if d.CreatedAt != nil {
		v.CreatedAt = *d.CreatedAt
	}
	return v
}

type Questionnaire struct {
	ID         int64
	UserID     int64
	History    string
	Storyboard string
	Status     bool
	Payment    bool
	CreatedAt  time.Time

	Answers   []byte
	Photos    []Photo
	GenPhotos []Photo
}
type QuestionnaireDTO struct {
	ID         *int64
	UserID     *int64
	Answers    *[]byte
	History    *string
	Storyboard *string
	Status     *bool
	Payment    *bool
	CreatedAt  *time.Time

	Photos    *[]Photo
	GenPhotos *[]Photo
}

func (q *Questionnaire) ToDTO() *QuestionnaireDTO {
	return &QuestionnaireDTO{
		ID:         &q.ID,
		UserID:     &q.UserID,
		Answers:    &q.Answers,
		History:    &q.History,
		Storyboard: &q.Storyboard,
		Status:     &q.Status,
		Payment:    &q.Payment,
		Photos:     &q.Photos,
		GenPhotos:  &q.GenPhotos,
	}
}

func (q *Questionnaire) ToProto() *protos.Questionnaire {
	out := &protos.Questionnaire{
		Id:         q.ID,
		UserId:     q.UserID,
		History:    q.History,
		Storyboard: q.Storyboard,
		Status:     q.Status,
		Payment:    q.Payment,
	}

	if !q.CreatedAt.IsZero() && !q.CreatedAt.Equal(time.Time{}) {
		out.CreatedAt = timestamppb.New(q.CreatedAt)
	}

	if len(q.Answers) > 0 {
		var raw []map[string]any
		if err := json.Unmarshal(q.Answers, &raw); err == nil {
			toString := func(value any) string {
				switch v := value.(type) {
				case nil:
					return ""
				case string:
					return v
				default:
					return fmt.Sprint(v)
				}
			}

			out.Answers = make([]*protos.Answer, 0, len(raw))
			for _, item := range raw {
				question := ""
				if v, ok := item["question"]; ok {
					question = toString(v)
				} else if v, ok := item["q"]; ok {
					question = toString(v)
				}

				answer := ""
				if v, ok := item["answer"]; ok {
					answer = toString(v)
				} else if v, ok := item["a"]; ok {
					answer = toString(v)
				}

				out.Answers = append(out.Answers, &protos.Answer{
					Question: question,
					Answer:   answer,
				})
			}
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
	if d.Storyboard != nil {
		q.Storyboard = *d.Storyboard
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
	if d.Answers != nil {
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
	UserID   *int64
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
	Promocode     *string
	AgeFrom       *int64
	AgeTo         *int64
	Gender        *string
	MapBinding    *bool
}

type PromoCode struct {
	ID          int64
	Value       string
	NumberUses  *int
	Status      *bool
	Percent     int64
	Description string
}

type PromoCodeDTO struct {
	ID          *int64
	Value       *string
	NumberUses  *int
	Status      *bool
	Percent     *int64
	Description *string
}

func (p *PromoCode) ToDTO() *PromoCodeDTO {
	return &PromoCodeDTO{
		ID:          &p.ID,
		Value:       &p.Value,
		NumberUses:  p.NumberUses,
		Status:      p.Status,
		Percent:     &p.Percent,
		Description: &p.Description,
	}
}

func (d *PromoCodeDTO) ToEntity() *PromoCode {
	p := &PromoCode{}
	if d.ID != nil {
		p.ID = *d.ID
	}
	if d.Value != nil {
		p.Value = *d.Value
	}
	if d.NumberUses != nil {
		p.NumberUses = d.NumberUses
	}
	if d.Status != nil {
		p.Status = d.Status
	}
	if d.Percent != nil {
		p.Percent = *d.Percent
	}
	if d.Description != nil {
		p.Description = *d.Description
	}
	return p
}

type PromoCodeFilter struct {
	Status *bool
}

type Review struct {
	ID          int64
	UserID      int64
	Description string
	CreatedAt   time.Time
}

type ReviewDTO struct {
	ID          *int64
	UserID      *int64
	Description *string
	CreatedAt   *time.Time
}

func (r *Review) ToDTO() *ReviewDTO {
	return &ReviewDTO{
		ID:          &r.ID,
		UserID:      &r.UserID,
		Description: &r.Description,
		CreatedAt:   &r.CreatedAt,
	}
}

func (d *ReviewDTO) ToEntity() *Review {
	review := &Review{}
	if d.ID != nil {
		review.ID = *d.ID
	}
	if d.UserID != nil {
		review.UserID = *d.UserID
	}
	if d.Description != nil {
		review.Description = *d.Description
	}
	if d.CreatedAt != nil {
		review.CreatedAt = *d.CreatedAt
	}
	return review
}

type ReviewFilter struct {
	UserID   *int64
	DateFrom *time.Time
	DateTo   *time.Time
}

type MediaUpload struct {
	DemoPhotos       []*multipart.FileHeader
	FinalPhotos      []*multipart.FileHeader
	DemoVideo        *multipart.FileHeader
	GeneratedVideo   *multipart.FileHeader
	DeliveryPhoto    *multipart.FileHeader
	FinalPhotoScenes []string
}
