package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	_ "modernc.org/sqlite"
)

type Database struct {
	u *sql.DB
}

func (d *Database) Init() error {
	var err error
	d.u, err = sql.Open("sqlite", "cap.db")
	if err != nil {
		return fmt.Errorf("init: open: %w", err)
	}
	createRequestsTable := `CREATE TABLE IF NOT EXISTS requests (
		id TEXT PRIMARY KEY,
		kind INTEGER NOT NULL,
		datetime timestamp NOT NULL,
		host TEXT NOT NULL,
		clientIP TEXT NOT NULL,
		clientAuthorization TEXT,
		clientProcessName TEXT NOT NULL,
		reqURL TEXT NOT NULL,
		reqMethod TEXT NOT NULL,
		reqHeaders BLOB NOT NULL,
		reqBody BLOB NOT NULL,
		respStatusCode INTEGER NOT NULL,
		respHeaders BLOB NOT NULL,
		respBody BLOB NOT NULL,
		error TEXT
	);`
	_, err = d.u.Exec(createRequestsTable)
	if err != nil {
		return fmt.Errorf("init: failed to create requests table: %w", err)
	}
	return nil
}

func (d *Database) GetRequestByID(id string) (*Request, error) {
	query := `SELECT
		id,
		kind,
		datetime,
		host,
		clientIP,
		clientAuthorization,
		clientProcessName,
		reqMethod,
		reqURL,
		reqHeaders,
		reqBody,
		respStatusCode,
		respHeaders,
		respBody,
		error
	FROM requests WHERE id = ?`
	row := d.u.QueryRow(query, id)

	req := new(Request)
	req.req = new(http.Request)
	req.req.URL = new(url.URL)
	req.resp = new(http.Response)

	// var datetime string
	var requestQuery []byte
	var requestHeaders []byte
	var requestBody []byte
	var responseHeaders []byte
	var responseBody []byte
	var errText sql.NullString
	var requrl string

	err := row.Scan(
		&req.ID,
		&req.Kind,
		&req.Datetime,
		&req.Host,
		&req.ClientIP,
		&req.ClientAuthorization,
		&req.ClientProcessName,
		&req.req.Method,
		&requrl,
		&requestHeaders,
		&requestBody,
		&req.resp.StatusCode,
		&responseHeaders,
		&responseBody,
		&errText,
	)
	if err != nil {
		return nil, fmt.Errorf("GetRequestByID: %w", err)
	}
	var urlQuery url.Values
	err = json.Unmarshal(requestQuery, &urlQuery)
	if err == nil {
		req.req.URL.RawQuery = urlQuery.Encode()
	}
	json.Unmarshal(requestHeaders, &req.req.Header)
	json.Unmarshal(responseHeaders, &req.resp.Header)
	req.req.Body = io.NopCloser(bytes.NewBuffer(requestBody))
	req.resp.Body = io.NopCloser(bytes.NewBuffer(responseBody))
	req.req.URL, _ = url.Parse(requrl)
	req.errorText = errText.String
	// NOTE: handle error text here

	return req, nil
}

func (d *Database) SaveRequest(req *Request, err error) error {
	query := `INSERT INTO requests (
		id,
		kind,
		datetime,
		host,

		clientIP,
		clientAuthorization,
		clientProcessName,

		error`
	args := []any{
		req.ID,
		req.Kind,
		req.Datetime,
		req.Host,
		req.ClientIP,
		req.ClientAuthorization,
		req.ClientProcessName,
	}
	if err != nil {
		args = append(args, err.Error())
	} else {
		args = append(args, nil)
	}

	// request
	query += `,
		reqMethod,
		reqURL,
		reqHeaders,
		reqBody`
	if req.req != nil {
		reqbody := req.body()
		if reqbody == nil {
			reqbody = []byte("")
		}
		args = append(args,
			req.req.Method,
			req.req.URL.String(),
			marshal(req.req.Header),
			reqbody,
		)
	} else {
		args = append(args,
			"",           // Method
			"",           // URL
			[]byte("{}"), // Headers
			[]byte(""),   // Body
		)
	}

	// response
	query += `,
		respStatusCode,
		respHeaders,
		respBody`
	if req.resp != nil {
		args = append(args,
			req.resp.StatusCode,
			marshal(req.resp.Header),
			marshal(req.respbody()),
		)
	} else {
		args = append(args,
			0,            // StatusCode
			[]byte("{}"), // Headers
			[]byte(""),   // Body
		)
	}

	query += `) VALUES (`
	for i := range len(args) {
		query += "?"
		if i < len(args)-1 {
			query += ", "
		}
	}
	query += `);`
	_, err = d.u.Exec(query, args...)
	if err != nil {
		slog.Error("save request", "err", err.Error())
		return fmt.Errorf("save request: %w", err)
	}

	return nil
}

func NewDatabase(db *sql.DB) *Database {
	return &Database{u: db}
}
