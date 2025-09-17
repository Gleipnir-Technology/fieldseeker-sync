package database

import (
	"context"
	"errors"

	"github.com/Gleipnir-Technology/fieldseeker-sync/database/sql"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/dialect"
	"github.com/stephenafamo/bob/dialect/psql/sm"
)

type TaskAudioReviewOutstandingSort int

const (
	SortNeedsReview TaskAudioReviewOutstandingSort = iota
	SortCreated
	SortAudioDuration
	SortCreatorName
)

// copied from sql.task_audio_review.bob.go
type taskAudioReviewOutstandingTransformer = bob.SliceTransformer[sql.TaskAudioReviewOutstandingRow, []sql.TaskAudioReviewOutstandingRow]

func TaskAudioReviewList(sort TaskAudioReviewOutstandingSort, isAscending bool) ([]sql.TaskAudioReviewOutstandingRow, error) {
	results := make([]sql.TaskAudioReviewOutstandingRow, 0)
	if PGInstance == nil {
		return results, errors.New("You must initialize the DB first")
	}
	//rows, err := query.All(ctx, db)
	ctx := context.Background()
	thing := sql.TaskAudioReviewOutstanding()
	var orderBy dialect.OrderBy[*dialect.SelectQuery]
	if isAscending {
		orderBy = sm.OrderBy("task_created").Desc()
	} else {
		orderBy = sm.OrderBy("task_created").Asc()
	}
	selector := psql.Select(
		thing,
		orderBy,
	)
	db := bob.NewDB(stdlib.OpenDBFromPool(PGInstance.DB))
	var rows []sql.TaskAudioReviewOutstandingRow
	var err error
	rows, err = bob.Allx[taskAudioReviewOutstandingTransformer](ctx, db, selector, thing.Scanner)

	if err != nil {
		return results, err
	}

	return rows, nil
}
