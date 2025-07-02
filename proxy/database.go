package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"strings"

	"github.com/tiredkangaroo/bigproxy/proxy/config"
	"github.com/tiredkangaroo/bigproxy/proxy/http"
	"github.com/tiredkangaroo/bigproxy/proxy/work"

	"github.com/ncruces/go-sqlite3"
	"github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
	"github.com/ncruces/go-sqlite3/ext/blobio"
)

// NOTE: bytes transferred

type FilterType string

const (
	FilterTypeUnknown FilterType = "unknown"
	FilterTypeString  FilterType = "string"
	FilterTypeNumber  FilterType = "number"
	FilterTypeBool    FilterType = "bool"
)

type Filter []FilterField
type FilterField struct {
	Name          string     `json:"name"`
	VerboseName   string     `json:"verboseName"`   // human readable name
	Type          FilterType `json:"type"`          // "string" | "int" | "bool"
	UniqueValues  []any      `json:"uniqueValues"`  // string | int
	SelectedValue any        `json:"selectedValue"` // string | int | bool
}

type Database struct {
	b          *sql.DB
	workerpool *work.WorkerPool
}

func (d *Database) Exec(query string, args ...any) (sql.Result, error) {
	var result sql.Result
	var err error
	d.workerpool.AddWait(func() {
		result, err = d.b.Exec(query, args...)
	})
	d.logError(query, args, err)
	return result, err
}

func (d *Database) Query(query string, args ...any) (*sql.Rows, error) {
	var rows *sql.Rows
	var err error
	d.workerpool.AddWait(func() {
		rows, err = d.b.Query(query, args...)
	})
	d.logError(query, args, err)
	return rows, err
}

func (d *Database) logError(query string, args []any, err error) {
	if err != nil {
		args := []any{
			"error", err.Error(),
			"query", query,
			"args", args,
		}
		if config.DefaultConfig.Debug { // only log with runtime.Caller in debug mode (this could be expensive im not sure, but defo not smth u want in prod)
			_, file, line, ok := runtime.Caller(2) // the actual caller -> exec/query -> this function
			if ok {
				args = append(args, "file", file, "line", line)
			}
		}
		slog.Error("database exec error", args...)
	}
}

func (d *Database) QueryRow(query string, args ...any) *sql.Row {
	var row *sql.Row
	d.workerpool.AddWait(func() {
		row = d.b.QueryRow(query, args...)
	})
	return row
}

type DatabaseAction struct {
	returnRows bool

	args []any
}

func (d *Database) Init(dirname string) error {
	d.workerpool.Start()

	var err error

	d.b, err = driver.Open(fmt.Sprintf("file:%s/cap.db?cache=shared&_journal_mode=WAL", dirname), blobio.Register)
	if err != nil {
		return fmt.Errorf("init: open: %w", err)
	}
	slog.Info("database open at cap.db", "dirname", dirname)

	// not null is present everywhere for my own sanity
	createRequestsTable := `CREATE TABLE IF NOT EXISTS requests (
		id TEXT PRIMARY KEY,
		starred BOOLEAN NOT NULL DEFAULT FALSE,
		secure BOOLEAN NOT NULL,
		datetime timestamp NOT NULL,
		host TEXT NOT NULL,
		clientIP TEXT NOT NULL,
		clientAuthorization TEXT,
		clientApplication TEXT NOT NULL,

		reqMethod INTEGER NOT NULL,
		reqPath TEXT NOT NULL,
		reqQuery BLOB NOT NULL,
		reqHeaders BLOB NOT NULL,
		reqBodyID TEXT NOT NULL,
		reqBodySize INTEGER NOT NULL,

		respStatusCode INTEGER NOT NULL,
		respHeaders BLOB NOT NULL,
		respBodyID TEXT NOT NULL,
		respBodySize INTEGER NOT NULL,

		timing BLOB NOT NULL,
		error TEXT
	);`
	_, err = d.Exec(createRequestsTable)
	if err != nil {
		return fmt.Errorf("init: failed to create requests table: %w", err)
	}
	bodyTable := `CREATE TABLE IF NOT EXISTS bodies (
		id TEXT PRIMARY KEY,
		body BLOB NOT NULL
	);`
	_, err = d.Exec(bodyTable)
	if err != nil {
		return fmt.Errorf("init: failed to create bodies table: %w", err)
	}

	return nil
}

func (d *Database) scanSingleRequest(row interface {
	Scan(dest ...any) error
}) (*Request, error) {
	req := &Request{
		req:  http.NewRequest(),
		resp: http.NewResponse(),
	}
	var reqQueryRaw, reqHeadersRaw, respHeadersRaw, timingDataRaw []byte
	var errorText sql.NullString
	err := row.Scan(
		&req.ID,
		&req.Starred,
		&req.Secure,
		sqlite3.TimeFormat4.Scanner(&req.Datetime),
		&req.Host,
		&req.ClientIP,
		&req.ClientAuthorization,
		&req.ClientApplication,
		&req.req.Method,
		&req.req.Path,
		&reqQueryRaw,
		&reqHeadersRaw,
		&req.reqBodyID,
		&req.req.ContentLength,
		&req.resp.StatusCode,
		&respHeadersRaw,
		&req.respBodyID,
		&req.resp.ContentLength,
		&timingDataRaw,
		&errorText,
	)
	if err != nil {
		return nil, fmt.Errorf("scan single request: %w", err)
	}
	// NOTE: streamline this, or use a helper func
	if err := json.Unmarshal(reqQueryRaw, &req.req.Query); err != nil {
		return nil, fmt.Errorf("scan single request: unmarshal query")
	}
	if err := json.Unmarshal(reqHeadersRaw, &req.req.Header); err != nil {
		return nil, fmt.Errorf("scan single request: unmarshal headers")
	}
	if err := json.Unmarshal(respHeadersRaw, &req.resp.Header); err != nil {
		return nil, fmt.Errorf("scan single request: unmarshal response headers")
	}
	if err := json.Unmarshal(timingDataRaw, &req.timing); err != nil {
		return nil, fmt.Errorf("scan single request: unmarshal timing data: %w", err)
	}
	if errorText.Valid {
		req.errorText = errorText.String
	} else {
		req.errorText = ""
	}
	req.reqBodyID = req.ID + "-req-body"
	req.respBodyID = req.ID + "-resp-body"

	return req, nil
}

func (d *Database) GetRequestByID(id string) (*Request, error) {
	query := `SELECT
		id,
		starred,
		secure,
		datetime,
		host,
		clientIP,
		clientAuthorization,
		clientApplication,
		reqMethod,
		reqPath,
		reqQuery,
		reqHeaders,
		reqBodyID,
		reqBodySize,
		respStatusCode,
		respHeaders,
		respBodyID,
		respBodySize,
		timing,
		error
	FROM requests WHERE id = ?`
	row := d.QueryRow(query, id)
	return d.scanSingleRequest(row)
}

// unilateral definition of filter selects here

// GetFilterCounts returns a map of fields to filter by to their unique values in the database.
func (d *Database) GetFilterUniqueValues(filter Filter) error {
	var err error
	for _, f := range filter {
		if f.Type != FilterTypeString && f.Type != FilterTypeNumber { // only string and int types are supported for unique values (bool unique values is just t/f lmao)
			continue
		}
		f.UniqueValues, err = d.uniqueValues(f.Name)
		if err != nil {
			return fmt.Errorf("get filter unique values for %s: %w", f.Name, err)
		}
	}
	return nil
}

// uniqueValuesAndCount returns a map of unique values for the specified column and how many time each value appears in the requests table.
func (d *Database) uniqueValues(by string) ([]any, error) {
	data := make([]any, 0, 8)
	query := fmt.Sprintf(`SELECT %s FROM requests GROUP BY %s;`, by, by)
	rows, err := d.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var uvalue any
		err := rows.Scan(&uvalue)
		if err != nil {
			return nil, err
		}
		if uvalue == "" {
			continue // skip empty values, they are not useful for filtering (also empty values are not allowed in radix ui selects)
		}
		data = append(data, uvalue)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return data, nil
}

// this function executes two queries, one for paginated []Request and one for total count as if there was no limit or offset
func (d *Database) GetRequestsMatchingFilter(f Filter, offset, limit int) ([]*Request, int, error) {
	// paginated query
	queryBase := `SELECT
		id,
		starred,
		secure,
		datetime,
		host,
		clientIP,
		clientAuthorization,
		clientApplication,
		reqMethod,
		reqPath,
		reqQuery,
		reqHeaders,
		reqBodyID,
		reqBodySize,
		respStatusCode,
		respHeaders,
		respBodyID,
		respBodySize,
		timing,
		error
	FROM requests`

	// count query
	countQueryBase := `SELECT COUNT(*) FROM requests`

	filtersUsed := []string{}
	args := []any{}
	countArgs := []any{}

	for _, f := range f {
		if f.SelectedValue == nil || f.SelectedValue == "" {
			continue // skip empty values, they are not useful for filtering
		}
		switch f.Type {
		case FilterTypeString, FilterTypeNumber:
			filtersUsed = append(filtersUsed, fmt.Sprintf("%s = ?", f.Name))
			args = append(args, f.SelectedValue)
			countArgs = append(countArgs, f.SelectedValue)
		case FilterTypeBool:
			sv, ok := f.SelectedValue.(bool)
			if !ok {
				panic("unreachable: how did u get here bro :/")
			}
			// this can be replaced with "f.Name (true)" and "NOT(f.Name)" (false)
			// but ts should work 🥀
			val := 0
			if sv {
				val = 1
			}
			filtersUsed = append(filtersUsed, fmt.Sprintf("%s = ?", f.Name))
			args = append(args, val)
			countArgs = append(countArgs, val)
		default:
			return nil, 0, fmt.Errorf("get requests with matching filter: unsupported filter type %s for field %s", f.Type, f.Name)
		}
	}

	// this where clause is used for both queries and represents the filters used
	whereClause := ""
	if len(filtersUsed) > 0 {
		whereClause = " WHERE " + strings.Join(filtersUsed, " AND ")
	}

	// count query building + execution
	var totalCount int
	countQuery := countQueryBase + whereClause
	err := d.QueryRow(countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("get requests count (query: %s): %w", countQuery, err)
	}

	// paginated data query building + execution
	query := queryBase + whereClause + " ORDER BY datetime DESC LIMIT ? OFFSET ?;"
	args = append(args, limit, offset)

	rows, err := d.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("get requests with matching filter (query: %s): %w", query, err)
	}

	// the requests with pagination
	reqs := make([]*Request, 0, limit)
	for rows.Next() {
		req, err := d.scanSingleRequest(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("get requests with matching filter (scan): %w", err)
		}
		reqs = append(reqs, req)
	}

	return reqs, totalCount, nil
}

func (d *Database) SaveRequest(req *Request, err error) error {
	query := `INSERT INTO requests (
		id,
		secure,
		datetime,
		host,

		clientIP,
		clientAuthorization,
		clientApplication,

		error`
	args := []any{
		req.ID,
		req.Secure,
		sqlite3.TimeFormat4.Encode(req.Datetime),
		req.Host,
		req.ClientIP,
		req.ClientAuthorization,
		req.ClientApplication,
	}
	if err != nil {
		args = append(args, err.Error())
	} else {
		args = append(args, nil)
	}

	// request
	query += `,
		reqMethod,
		reqPath,
		reqHeaders,
		reqQuery,
		reqBodyID,
		reqBodySize`
	if req.req != nil {
		args = append(args,
			req.req.Method,
			req.req.Path,
			marshal(req.req.Header),
			marshal(req.req.Query),
			req.reqBodyID,
			req.req.ContentLength,
		)
	} else {
		args = append(args,
			0,            // Method
			"",           // Path
			[]byte("{}"), // Headers
			[]byte("{}"), // Query
			"",           // Body ID
			0,            // Body Size
		)
	}

	// response
	query += `,
		respStatusCode,
		respHeaders,
		respBodyID,
		respBodySize`
	if req.resp != nil {
		args = append(args,
			req.resp.StatusCode,
			marshal(req.resp.Header),
			req.respBodyID,
			req.resp.ContentLength,
		)
	} else {
		args = append(args,
			0,            // StatusCode
			[]byte("{}"), // Headers
			"",           // Body ID
			0,            // Body Size
		)
	}

	// timing
	query += `,
		timing`
	if req.timing != nil {
		timingData, _ := json.Marshal(req.timing)
		args = append(args, timingData)
	} else {
		args = append(args, []byte("{}")) // empty timing
	}

	query += `) VALUES (`
	for i := range len(args) {
		query += "?"
		if i < len(args)-1 {
			query += ", "
		}
	}
	query += `);`
	_, err = d.Exec(query, args...)
	if err != nil {
		slog.Error("save request", "err", err.Error())
		return fmt.Errorf("save request: %w", err)
	}

	return nil
}

// NOTE: combine next two last parts of func into one for wblob

func (d *Database) SaveBody(id string, body *http.Body) error {
	cl := body.ContentLength()
	if cl == 0 {
		_, err := d.Exec(`INSERT INTO bodies (id, body) VALUES (?, ?)`, id, []byte{})
		return err
	}
	query := `INSERT INTO bodies (id, body) VALUES (:id, :cl);`
	res, err := d.Exec(query, sql.Named("id", id), sql.Named("cl", sqlite3.ZeroBlob(cl)))
	if err != nil {
		return fmt.Errorf("save body: %w", err)
	}
	rowid, _ := res.LastInsertId()

	_, err = d.Exec(
		`SELECT writeblob('main', 'bodies', 'body', :rowid, :offset, :message)`,
		sql.Named("rowid", rowid), sql.Named("offset", 0), sql.Named("message", sqlite3.Pointer(body)),
	)
	if err != nil {
		return fmt.Errorf("save body: writeblob: %w", err)
	}
	return nil
}

func (d *Database) UpdateBody(id string, body *http.Body) error {
	cl := body.ContentLength()
	query := `UPDATE bodies SET body = :cl WHERE id = :id;`
	res, err := d.Exec(query, sql.Named("cl", sqlite3.ZeroBlob(cl)), sql.Named("id", id))
	if err != nil {
		return fmt.Errorf("update body: %w", err)
	}
	rowid, _ := res.LastInsertId()

	_, err = d.Exec(
		`SELECT writeblob('main', 'bodies', 'body', :rowid, :offset, :message)`,
		sql.Named("rowid", rowid), sql.Named("offset", 0), sql.Named("message", sqlite3.Pointer(body)),
	)
	if err != nil {
		return fmt.Errorf("save body: writeblob: %w", err)
	}
	return nil
}

func (d *Database) WriteRequestBody(id string, writer io.Writer) error {
	query := `SELECT rowid, length(body) FROM bodies WHERE id = ?;`
	row := d.QueryRow(query, id+"-req-body")
	if row == nil {
		return fmt.Errorf("write request body: no body found for id %s", id)
	}
	return d.readBlobToWriterWithRowID(row, writer)
}

func (d *Database) WriteResponseBody(id string, writer io.Writer) error {
	query := `SELECT rowid, length(body) FROM bodies WHERE id = ?;`
	row := d.QueryRow(query, id+"-resp-body")
	if row == nil {
		return fmt.Errorf("write response body: no body found for id %s", id)
	}
	return d.readBlobToWriterWithRowID(row, writer)
}

func (d *Database) SetRequestStarred(id string, starred bool) error {
	query := `UPDATE requests SET starred = ? WHERE id = ?;`
	_, err := d.Exec(query, starred, id)
	if err != nil {
		return fmt.Errorf("set request starred: %w", err)
	}
	return nil
}

func (d *Database) readBlobToWriterWithRowID(row *sql.Row, writer io.Writer) error {
	var rowid int64
	var length int64
	err := row.Scan(&rowid, &length)
	if err != nil {
		return fmt.Errorf("read blob to writer: scan row: %w", err)
	}
	writer.Write(fmt.Appendf([]byte{}, "Content-Length: %d\r\n\r\n", length))

	query := `SELECT readblob('main', 'bodies', 'body', :rowid, :offset, :writer);`
	_, err = d.Exec(query, sql.Named("rowid", rowid), sql.Named("offset", 0), sql.Named("writer", sqlite3.Pointer(writer)))
	if err != nil {
		return fmt.Errorf("write request body: %w", err)
	}

	return nil
}

func NewDatabase() *Database {
	return &Database{workerpool: work.NewWorkerPool(1)}
}
