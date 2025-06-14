-- name: GetRequestByID :one
SELECT
    *
FROM
    requests
WHERE
    id = ?;

-- name: CreateRequest :exec
INSERT INTO
    requests (
        id,
        kind,
        datetime,
        host,
        clientIP,
        clientAuthorization,
        clientProcessName,
        requestMethod,
        requestURL,
        requestHeaders,
        requestBody,
        responseStatus,
        responseHeaders,
        responseBody,
        error
    )
VALUES
    (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
