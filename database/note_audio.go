package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Gleipnir-Technology/fieldseeker-sync/database/models"
	"github.com/Gleipnir-Technology/fieldseeker-sync/shared"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	//"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	//"github.com/stephenafamo/bob/dialect/psql/dialect"
	"github.com/stephenafamo/bob/dialect/psql/sm"
)

func NoteAudioCreate(ctx context.Context, noteUUID uuid.UUID, payload shared.NoteAudioPayload, userID int) error {
	var options pgx.TxOptions
	transaction, err := PGInstance.DB.BeginTx(ctx, options)
	if err != nil {
		return fmt.Errorf("Failed to begin transaction: %v", err)
	}

	query := `INSERT INTO note_audio (
			created,
			creator,
			deleted,
			duration,
			has_been_reviewed,
			is_audio_normalized,
			is_transcoded_to_ogg,
			needs_further_review,
			transcription,
			transcription_internally_edited,
			transcription_user_edited,
			version,
			uuid
		) VALUES (
			@created,
			@creator,
			@deleted,
			@duration,
			@has_been_reviewed,
			@is_audio_normalized,
			@is_transcoded_to_ogg,
			@needs_further_review,
			@transcription,
			@transcription_internally_edited,
			@transcription_user_edited,
			@version,
			@uuid)`
	args := pgx.NamedArgs{
		"created":                         payload.Created,
		"creator":                         userID,
		"deleted":                         nil,
		"duration":                        payload.Duration,
		"has_been_reviewed":               false,
		"is_audio_normalized":             false,
		"is_transcoded_to_ogg":            false,
		"needs_further_review":            false,
		"transcription":                   payload.Transcription,
		"transcription_internally_edited": false,
		"transcription_user_edited":       payload.TranscriptionUserEdited,
		"version":                         payload.Version,
		"uuid":                            noteUUID,
	}
	row, err := PGInstance.DB.Exec(context.Background(), query, args)
	if err != nil {
		transaction.Rollback(ctx)
		return fmt.Errorf("Unable to insert row into note_audio: %v", err)
	}
	log.Println("Saved audio note", noteUUID, row)

	rows := make([][]interface{}, 0, len(payload.Breadcrumbs))
	for i, b := range payload.Breadcrumbs {
		rows = append(rows, []interface{}{
			b.Created,
			b.Cell,
			noteUUID,
			payload.Version,
			i,
		})
	}

	PGInstance.DB.CopyFrom(
		ctx,
		pgx.Identifier{"note_audio_breadcrumb"},
		[]string{"created", "cell", "note_audio_uuid", "note_audio_version", "position"},
		pgx.CopyFromRows(rows),
	)

	query = `INSERT INTO task_audio_review (
		completed_by,
		created,
		needs_review,
		note_audio_uuid,
		note_audio_version,
		reviewed_by
	) VALUES (
		@completed_by,
		@created,
		@needs_review,
		@note_audio_uuid,
		@note_audio_version,
		@reviewed_by)`
	args = pgx.NamedArgs{
		"completed_by":       nil,
		"created":            time.Now(),
		"needs_review":       false,
		"note_audio_uuid":    noteUUID,
		"note_audio_version": payload.Version,
		"reviewed_by":        nil,
	}
	row, err = PGInstance.DB.Exec(context.Background(), query, args)
	if err != nil {
		transaction.Rollback(ctx)
		return fmt.Errorf("Unable to insert row into task_audio_review: %v", err)
	}

	err = transaction.Commit(ctx)
	if err != nil {
		return fmt.Errorf("Failed to commit transaction: %v", err)
	}
	return nil
}

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

func NoteAudioNormalized(uuid string) error {
	if PGInstance == nil {
		return errors.New("You must initialize the DB first")
	}
	args := pgx.NamedArgs{
		"is_audio_normalized": true,
		"uuid":                uuid,
	}
	query := "UPDATE note_audio SET is_audio_normalized=@is_audio_normalized WHERE uuid=@uuid"
	_, err := PGInstance.DB.Exec(context.Background(), query, args)
	return err
}

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

func NoteAudioToNormalize() ([]*shared.NoteAudio, error) {
	results := make([]*shared.NoteAudio, 0)
	if PGInstance == nil {
		return results, errors.New("You must initialize the DB first")
	}
	args := pgx.NamedArgs{}
	query := "SELECT created, creator, duration, is_audio_normalized, transcription, transcription_user_edited, version, uuid FROM note_audio WHERE is_audio_normalized = FALSE"

	rows, _ := PGInstance.DB.Query(context.Background(), query, args)

	if err := pgxscan.ScanAll(&results, rows); err != nil {
		log.Println("ScanAll on note_audio error:", err)
		return results, err
	}
	return results, nil
}

func NoteAudioToTranscodeToOgg() ([]*shared.NoteAudio, error) {
	results := make([]*shared.NoteAudio, 0)
	if PGInstance == nil {
		return results, errors.New("You must initialize the DB first")
	}
	args := pgx.NamedArgs{}
	query := "SELECT created, creator, duration, is_audio_normalized, transcription, transcription_user_edited, version, uuid FROM note_audio WHERE is_transcoded_to_ogg = FALSE"

	rows, _ := PGInstance.DB.Query(context.Background(), query, args)

	if err := pgxscan.ScanAll(&results, rows); err != nil {
		log.Println("ScanAll on note_audio error:", err)
		return results, err
	}
	return results, nil
}

func NoteAudioTranscodedToOgg(uuid string) error {
	if PGInstance == nil {
		return errors.New("You must initialize the DB first")
	}
	args := pgx.NamedArgs{
		"is_transcoded_to_ogg": true,
		"uuid":                 uuid,
	}
	query := "UPDATE note_audio SET is_transcoded_to_ogg=@is_transcoded_to_ogg WHERE uuid=@uuid"
	_, err := PGInstance.DB.Exec(context.Background(), query, args)
	return err
}

func NoteAudioUpdateDelete(uuid string, userID int) error {
	ctx := context.Background()
	var options pgx.TxOptions
	transaction, err := PGInstance.DB.BeginTx(ctx, options)

	if err != nil {
		return fmt.Errorf("Failed to begin transaction: %v", err)
	}
	args := pgx.NamedArgs{
		"deleted":    time.Now(),
		"deleted_by": userID,
		"uuid":       uuid,
	}
	query := `
		UPDATE note_audio
		SET deleted=@deleted, deleted_by=@deleted_by
		WHERE uuid=@uuid
	`
	row, err := PGInstance.DB.Exec(context.Background(), query, args)
	if err != nil {
		return fmt.Errorf("Failed to update note_audio to deleted: %v\n", err)
	}

	query = `UPDATE task_audio_review SET
		completed_by=@completed_by
		WHERE note_audio_uuid=@note_audio_uuid`
	args = pgx.NamedArgs{
		"completed_by":    userID,
		"note_audio_uuid": uuid,
	}
	row, err = PGInstance.DB.Exec(context.Background(), query, args)
	if err != nil {
		return fmt.Errorf("Failed to update note_audio to deleted: %v\n", err)
	}

	log.Printf("Marked task_audio_review for %s %s completed", uuid, row)
	err = transaction.Commit(ctx)
	if err != nil {
		return fmt.Errorf("Failed to commit transaction: %v", err)
	}
	return nil
}

func NoteAudioUpdateReviewed(uuid string, userID int) error {
	args := pgx.NamedArgs{
		"has_been_reviewed": true,
		"user_id":           userID,
		"uuid":              uuid,
	}
	query := `
		UPDATE note_audio
		SET has_been_reviewed=@has_been_reviewed
		WHERE uuid=@uuid
	`
	row, err := PGInstance.DB.Exec(context.Background(), query, args)
	if err != nil {
		return fmt.Errorf("Failed to update transcription: %v\n", err)
	}
	log.Printf("Marked note_audio %s %s reviewed", uuid, row)
	query = `UPDATE task_audio_review SET
		completed_by=@user_id WHERE note_audio_uuid=@uuid`
	row, err = PGInstance.DB.Exec(context.Background(), query, args)
	if err != nil {
		return fmt.Errorf("Failed to update transcription: %v\n", err)
	}
	log.Printf("Marked review task for %s %s completed by %s", uuid, userID, row)

	return nil
}

func NoteAudioUpdateFurtherReviewed(taskID int32, userID int) error {
	args := pgx.NamedArgs{
		"id":      taskID,
		"user_id": userID,
	}
	query := `
		UPDATE task_audio_review SET
		reviewed_by=@user_id
		WHERE id=@id
	`
	_, err := PGInstance.DB.Exec(context.Background(), query, args)
	if err != nil {
		return fmt.Errorf("Failed to update further reviewed on review task: %v\n", err)
	}
	return nil
}

func NoteAudioUpdateNeedsFurtherReview(uuid string, userID int) error {
	args := pgx.NamedArgs{
		"needs_further_review": true,
		"user_id":              userID,
		"uuid":                 uuid,
	}
	query := `
		UPDATE note_audio
		SET needs_further_review=@needs_further_review
		WHERE uuid=@uuid
	`
	row, err := PGInstance.DB.Exec(context.Background(), query, args)
	if err != nil {
		return fmt.Errorf("Failed to update needs further review: %v\n", err)
	}
	log.Printf("Marked note_audio %s %s needs further review", uuid, row)

	query = `
		UPDATE task_audio_review SET
		completed_by=@user_id,
		needs_review=@needs_further_review
		WHERE note_audio_uuid=@uuid
	`
	row, err = PGInstance.DB.Exec(context.Background(), query, args)
	if err != nil {
		return fmt.Errorf("Failed to update needs further review on review task: %v\n", err)
	}
	return nil
}
func NoteAudioUpdateTranscription(uuid string, transcription string, userUUID int) error {
	ctx := context.Background()
	var options pgx.TxOptions
	transaction, err := PGInstance.DB.BeginTx(ctx, options)
	if err != nil {
		return fmt.Errorf("Failed to begin transaction: %v", err)
	}

	args := pgx.NamedArgs{
		"created":                         time.Now(),
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
			@created,
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
