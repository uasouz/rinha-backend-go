package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"rinha-backend-go/persistence"
	"rinha-backend-go/person"

	_ "github.com/mattn/go-sqlite3"
)

const (
	createPeopleTable = `
    CREATE TABLE IF NOT EXISTS people (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      uuid TEXT not null,
      name TEXT not null,
      nickname TEXT not null,
      birthdate INTEGER not null,
      stack TEXT,
      created_at INTEGER not null
    );

    CREATE INDEX IF NOT EXISTS idx_people_name ON people (name);
    CREATE INDEX IF NOT EXISTS idx_people_nickname ON people (nickname);
    CREATE INDEX IF NOT EXISTS idx_people_created_at ON people (created_at);
    CREATE INDEX IF NOT EXISTS idx_people_uuid ON people (uuid);
  `
	insertPerson = `
    insert into people (uuid,name,nickname,birthdate,stack,created_at)
    values (?,?,?,?,?,?);
  `

	selectPeople = `
    SELECT id,uuid,name,nickname,birthdate,stack,created_at
    FROM people
  `
	selectPerson = `
    SELECT id,uuid,name,nickname,birthdate,stack,created_at
    FROM people
    WHERE uuid= ?;
  `
)

type SQLiteStore struct {
	db *sql.DB
}

func (s *SQLiteStore) GetPeopleCount(ctx context.Context) (int64, error) {
	var count int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM people").Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

type PersonDB struct {
	ID        int
	UUID      string
	Name      string
	Nickname  string
	Birthdate int64
	Stack     string
	CreatedAt int64
}

func convertPersonToPersonDB(p person.Person) (*PersonDB, error) {
	stackJson, err := json.Marshal(p.Stack)

	if err != nil {
		return nil, err
	}

	return &PersonDB{
		ID:        p.ID,
		UUID:      p.UUID,
		Name:      p.Name,
		Nickname:  p.Nickname,
		Birthdate: p.Birthdate.Unix(),
		Stack:     string(stackJson),
		CreatedAt: time.Now().Unix(),
	}, nil
}

func convertPersonDBToPerson(p PersonDB) (*person.Person, error) {
	var stack []string
	err := json.Unmarshal([]byte(p.Stack), &stack)
	if err != nil {
		return nil, err
	}

	return &person.Person{
		ID:        p.ID,
		UUID:      p.UUID,
		Name:      p.Name,
		Nickname:  p.Nickname,
		Birthdate: time.Unix(p.Birthdate, 0),
		Stack:     stack,
		CreatedAt: time.Unix(p.CreatedAt, 0),
	}, nil
}

func (s *SQLiteStore) AddPerson(_ context.Context, p person.Person) (int64, error) {
	dbPerson, err := convertPersonToPersonDB(p)
	if err != nil {
		return 0, err
	}

	result, err := s.db.Exec(insertPerson, dbPerson.UUID, dbPerson.Name, dbPerson.Nickname, dbPerson.Birthdate, dbPerson.Stack, dbPerson.CreatedAt)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func extractValuesFromPaginationToken(token string) (int64, int64) {
	var id int64
	var createdAt int64

	if token != "" {
		values := strings.Split(token, "-")
		id, _ = strconv.ParseInt(values[0], 10, 64)
		createdAt, _ = strconv.ParseInt(values[1], 10, 64)
	}

	return id, createdAt
}

func containsQuery(s string) string {
	return "%" + s + "%"
}

func (s *SQLiteStore) GetPeople(_ context.Context, options *persistence.GetPeopleOptions) (person.People, error) {

	query := selectPeople

	optionsValues := []interface{}{}

	if options != nil && options.SearchQuery != "" {
		query += "WHERE name LIKE ? OR nickname LIKE ? "
		optionsValues = append(optionsValues, containsQuery(options.SearchQuery),
			containsQuery(options.SearchQuery))
	}

	if options != nil && options.PaginationToken != "" {
		if options.SearchQuery == "" {
			query += "WHERE "
		} else {
			query += "AND "
		}
		query += "(id > ? and created_at >= ?) "
		id, createdAt := extractValuesFromPaginationToken(options.PaginationToken)
		optionsValues = append(optionsValues, id, createdAt)

	}

	query += "ORDER BY id ASC LIMIT 5;"

	rows, err := s.db.Query(query, optionsValues...)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	defer rows.Close()

	var people person.People

	for rows.Next() {
		var p PersonDB
		err := rows.Scan(&p.ID, &p.UUID, &p.Name, &p.Nickname, &p.Birthdate, &p.Stack, &p.CreatedAt)
		if err != nil {
			return nil, err
		}

		person, err := convertPersonDBToPerson(p)

		if err != nil {
			return nil, err
		}

		people = append(people, person)

	}

	return people, nil
}

func (s *SQLiteStore) GetPerson(_ context.Context, id string) (*person.Person, error) {
	var p PersonDB
	err := s.db.QueryRow(selectPerson, id).Scan(&p.ID, &p.UUID, &p.Name, &p.Nickname, &p.Birthdate, &p.Stack, &p.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, persistence.ErrPersonNotFound
		}
		return nil, err
	}

	return convertPersonDBToPerson(p)
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

func NewSQLiteStore() (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", "./people.db")

	if err != nil {
		return nil, err
	}

	// Create the people table if it doesn't exist
	_, err = db.Exec(createPeopleTable)

	if err != nil {
		return nil, err
	}

	return &SQLiteStore{
		db: db,
	}, nil
}
