package database

import (
	"context"
	"errors"

	"github.com/Gleipnir-Technology/fieldseeker-sync/database/sql"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/stephenafamo/bob"
)

func TaskAudioReviewList() ([]sql.TaskAudioReviewOutstandingRow, error) {
	results := make([]sql.TaskAudioReviewOutstandingRow, 0)
	if PGInstance == nil {
		return results, errors.New("You must initialize the DB first")
	}
	ctx := context.Background()
	db := bob.NewDB(stdlib.OpenDBFromPool(PGInstance.DB))
	query := sql.TaskAudioReviewOutstanding()
	rows, err := query.All(ctx, db)
	if err != nil {
		return results, err
	}

	return rows, nil
}
