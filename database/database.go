package database

import (
	"database/sql"
	"encoding/json"
	"leraProxy/request"
)

type Database struct {
	DB *sql.DB
}

func (d *Database) Save(request *request.Request, headers string) error {
	query := `insert into request (method, host, url, body, headers) 
values ($1, $2, $3, $4, $5) returning id`

	_, err := d.DB.Exec(query, request.Method, request.Host, request.URL, request.Body, headers)
	return err
}

func (d *Database) GetAllRequests() ([]request.Request, error) {
	requests := make([]request.Request, 0)
	query := `select id, method, host, url, body, headers from request`

	row, err := d.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	for row.Next() {
		headers := make([]byte, 0)
		request := request.Request{}

		row.Scan(&request.ID, &request.Method, &request.Host, &request.URL, &request.Body, &headers)

		json.Unmarshal(headers, &request.Headers)

		requests = append(requests, request)
	}

	return requests, nil
}

func (d *Database) GetRequest(id int) (*request.Request, error) {
	query := `select id, method, host, url, body, headers from request where id = $1`

	request := &request.Request{}
	headers := make([]byte, 0)

	err := d.DB.QueryRow(query, id).Scan(&request.ID, &request.Method, &request.Host,
		&request.URL, &request.Body, &headers)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(headers, &request.Headers)
	return request, err
}

