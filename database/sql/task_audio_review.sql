-- TaskAudioReviewOutstanding
SELECT 
    task.id AS task_id,
    task.created AS task_created,
    task.needs_review,
    note.duration AS audio_duration,
    u.display_name AS creator_name
FROM 
    task_audio_review task
JOIN 
    note_audio note ON task.note_audio_uuid = note.uuid AND task.note_audio_version = note.version
JOIN 
    user_ u ON note.creator = u.id
WHERE 
    task.reviewed_by IS NULL
ORDER BY 
    task.created DESC;
