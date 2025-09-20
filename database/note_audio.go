package database

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/Gleipnir-Technology/fieldseeker-sync/database/models"

	"github.com/jackc/pgx/v5"
	//"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	//"github.com/stephenafamo/bob/dialect/psql/dialect"
	"github.com/stephenafamo/bob/dialect/psql/sm"
)

func NoteAudioGetLatest(ctx context.Context, uuid string) (*models.NoteAudio, error) {
	if PGInstance == nil {
		return nil, errors.New("You must initialize the DB first")
	}
	return models.NoteAudios.Query(
		sm.Where(psql.Quote("uuid").EQ(psql.Arg(uuid))),
		sm.OrderBy("version").Desc(),
		sm.Limit(1),
	).One(ctx, PGInstance.BobDB)
}

func NoteAudioUpdateTranscription(uuid string, transcription string, userUUID int) error {
	ctx := context.Background()
	var options pgx.TxOptions
	transaction, err := PGInstance.DB.BeginTx(ctx, options)
	if err != nil {
		return fmt.Errorf("Failed to begin transaction: %v", err)
	}

	args := pgx.NamedArgs{
		"creator":                         userUUID,
		"has_been_reviewed":               true,
		"transcription":                   transcription,
		"transcription_internally_edited": true,
		"uuid":                            uuid,
	}
	query := `
		WITH previous_row AS (
			SELECT *
			FROM note_audio
			WHERE uuid = @uuid
			AND version = (
				SELECT MAX(version)
				FROM note_audio
				WHERE uuid = @uuid
			)
		)
		INSERT INTO note_audio 
		(created, creator, deleted, duration, has_been_reviewed, is_audio_normalized, is_transcoded_to_ogg, needs_further_review, transcription, transcription_user_edited, transcription_internally_edited, version, uuid)
		SELECT
			created,
			@creator,
			deleted,
			duration,
			@has_been_reviewed,
			is_audio_normalized,
			is_transcoded_to_ogg,
			needs_further_review,
			@transcription,
			transcription_user_edited,
			@transcription_internally_edited,
			version + 1,
			@uuid
		FROM previous_row
	`
	row, err := PGInstance.DB.Exec(context.Background(), query, args)
	if err != nil {
		return fmt.Errorf("Failed to update transcription: %v\n", err)
	}
	log.Println("Saved updated transcription", uuid, row)

	args = pgx.NamedArgs{
		"completed_by":    userUUID,
		"note_audio_uuid": uuid,
	}
	query = `UPDATE task_audio_review SET completed_by=@completed_by WHERE note_audio_uuid=@note_audio_uuid`
	row, err = PGInstance.DB.Exec(context.Background(), query, args)
	if err != nil {
		return fmt.Errorf("Failed to complete audio review task: %v\n", err)
	}

	err = transaction.Commit(ctx)
	if err != nil {
		return fmt.Errorf("Failed to commit transaction: %v", err)
	}
	return nil
}
