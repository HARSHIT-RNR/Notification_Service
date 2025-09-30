-- name: CreateNotificationLog :exec
INSERT INTO notification_mails (
    recipient,
    template_name,
    status,
    details,
    data,
    attempted_at
) VALUES (
    $1, $2, $3, $4, $5, $6
);