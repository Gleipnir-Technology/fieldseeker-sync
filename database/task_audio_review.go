package database

import (
	"context"
	"errors"
	//"fmt"
	//"time"

	"github.com/jackc/pgx/v5/stdlib"
	//"github.com/Gleipnir-Technology/fieldseeker-sync/shared"
	"github.com/Gleipnir-Technology/fieldseeker-sync/database/models"
	//"github.com/Gleipnir-Technology/fieldseeker-sync/database/sql"

	//"github.com/aarondl/opt/omit"
	"github.com/stephenafamo/bob"
	//"github.com/stephenafamo/bob/expr"
	//"github.com/stephenafamo/bob/dialect/psql"
	//"github.com/stephenafamo/bob/dialect/psql/sm"
	//"github.com/stephenafamo/bob/dialect/psql/im"
	//"github.com/stephenafamo/bob/dialect/psql/um"
	//"github.com/stephenafamo/bob/dialect/psql/dm"
	//"github.com/jackc/pgx/v5"
	//"github.com/jackc/pgx/v5/pgxpool"
)

//type TaskAudioReviewSlice []*TaskAudioReview
//type TaskAudioReviewSetter struct {
//CreatedBy omit.Val[int] `db:"created_by"`
//ReviewedBy omit.Val[int] `db:"reviewed_by"`
//
//orm.Setter[*taskAudioReview, *dialect.InsertQuery, *dialect.UpdateQuery]
//}

func TaskAudioReviewList() ([]*models.TaskAudioReview, error) {
	results := make([]*models.TaskAudioReview, 0)
	if PGInstance == nil {
		return results, errors.New("You must initialize the DB first")
	}
	ctx := context.Background()
	db := bob.NewDB(stdlib.OpenDBFromPool(PGInstance.DB))
	//query := sql.AllTaskAudioReview()
	//rows, err := query.All(ctx, db)
	rows, err := models.TaskAudioReviews.Query().All(ctx, db)
	if err != nil {
		return results, err
	}

	return rows, nil
}

/*
func NoteAudioQuery() ([]*shared.NoteAudio, error) {
	results := make([]*shared.NoteAudio, 0)
	if PGInstance == nil {
		return results, errors.New("You must initialize the DB first")
	}
	query := `
		SELECT
			created,
			creator,
			duration,
			has_been_reviewed,
			is_audio_normalized,
			is_transcoded_to_ogg,
			needs_further_review,
			transcription,
			transcription_user_edited,
			version,
			uuid
		FROM (
			SELECT *, ROW_NUMBER() OVER (PARTITION BY uuid ORDER BY version DESC) as version_rank
			FROM note_audio
		) ranked_rows
		WHERE version_rank = 1 AND deleted IS NULL;
	`
	rows, _ := PGInstance.DB.Query(context.Background(), query)

	if err := pgxscan.ScanAll(&results, rows); err != nil {
		log.Println("ScanAll on note_audio error:", err)
		return results, err
	}
	return results, nil
}
*/
