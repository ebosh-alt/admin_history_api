package postgres

import (
	"admin_history/internal/repository/postgres/connection"
	domainPhoto "admin_history/internal/repository/postgres/domain/photo"
	domainPromocode "admin_history/internal/repository/postgres/domain/promocode"
	domainQuestionnaire "admin_history/internal/repository/postgres/domain/questionnaire"
	domainReview "admin_history/internal/repository/postgres/domain/review"
	domainUser "admin_history/internal/repository/postgres/domain/user"
	domainVideo "admin_history/internal/repository/postgres/domain/video"

	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"repository.postgres",
		fx.Provide(
			connection.NewPool,
			domainUser.New,
			domainQuestionnaire.New,
			domainPhoto.New,
			domainVideo.New,
			domainPromocode.New,
			domainReview.New,
		),
	)
}
