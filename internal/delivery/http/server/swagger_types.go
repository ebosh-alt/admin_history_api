// Package server содержит HTTP обработчики.
//
// @title           Admin History API
// @version         1.0
// @description     REST API для управления пользователями, анкетами, фотографиями и промокодами.
// @host            localhost:3000
// @BasePath        /api
// @schemes         http
package server

import (
	protos "admin_history/pkg/proto/gen/go"
)

// ErrorResponse describes a generic error payload returned by the API.
type ErrorResponse struct {
	Message string `json:"message"`
}

type (
	UserResponse                    = protos.UserResponse
	UsersListResponse               = protos.UsersListResponse
	UpdateUserRequest               = protos.UpdateUserRequest
	Status                          = protos.Status
	QuestionnaireResponse           = protos.QuestionnaireResponse
	QuestionnairesListResponse      = protos.QuestionnairesListResponse
	UpdateQuestionnaireRequest      = protos.UpdateQuestionnaireRequest
	SubmitQuestionnaireMediaRequest = protos.SubmitQuestionnaireMediaRequest
	PhotoResponse                   = protos.PhotoResponse
	VideoResponse                   = protos.VideoResponse
	CreatePromoCodeRequest          = protos.CreatePromoCodeRequest
	PromoCodeResponse               = protos.PromoCodeResponse
	PromoCodesListResponse          = protos.PromoCodesListResponse
	UpdatePromoCodeRequest          = protos.UpdatePromoCodeRequest
	ReviewResponse                  = protos.ReviewResponse
	ReviewsListResponse             = protos.ReviewsListResponse
	Photo                           = protos.Photo
	Video                           = protos.Video
	PhotoRequest                    = protos.PhotoRequest
	VideoRequest                    = protos.VideoRequest
	PromoCode                       = protos.PromoCode
	Review                          = protos.Review
	Questionnaire                   = protos.Questionnaire
)
