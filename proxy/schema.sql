CREATE TABLE IF NOT EXISTS requests (
    id text PRIMARY KEY,
    kind smallint NOT NULL,
    datetime TIMESTAMP NOT NULL,
    host text NOT NULL,
    clientIP text NOT NULL,
    clientAuthorization text NOT NULL,
    clientProcessName text,
    requestMethod text,
    requestURL text,
    requestHeaders blob,
    requestBody blob,
    responseStatus smallint,
    responseHeaders blob,
    responseBody blob,
    error text
);
