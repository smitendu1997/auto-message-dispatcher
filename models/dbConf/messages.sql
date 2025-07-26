-- name: GetPendingMessage :many
SELECT id, recipient_phone, content, status, messageid, sent_at, retry_count
FROM messages
WHERE status in ('pending','failed')
AND (retry_count < 3 OR status = 'pending')
LIMIT 2;

-- name: UpdateMessage :exec
UPDATE messages
SET
    content = COALESCE(sqlc.narg('content'), content),
    status = COALESCE(sqlc.narg('status'), status),
    messageId = COALESCE(sqlc.narg('messageId'), messageId),
    sent_at = COALESCE(sqlc.narg('sent_at'), sent_at),
    retry_count = CASE
        WHEN sqlc.narg('status') = 'failed' THEN retry_count + 1
        ELSE retry_count
    END
WHERE id = sqlc.arg('id');

-- name: GetSentMessages :many
SELECT id, recipient_phone, content, status, messageid, sent_at, createdon, retry_count FROM messages
WHERE status = 'sent'
ORDER BY sent_at DESC, id DESC
LIMIT ? OFFSET ?;

-- name: GetSentMessagesCount :one
SELECT COUNT(*) as total FROM messages WHERE status = 'sent';
