-- TaskAudioReviewCompletedBy
SELECT
	task.id AS task_id,
	task.note_audio_uuid AS note_audio_uuid,
	note.transcription AS transcription
FROM
	task_audio_review task
JOIN
	note_audio note  ON task.note_audio_uuid = note.uuid AND task.note_audio_version = (note.version-1)
JOIN
	user_ u ON task.completed_by = u.id
WHERE
	(task.completed_by = u.id AND u.username = ($1));
