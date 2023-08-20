package postgres

import (
	"context"
	"database/sql"
	"embed"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"

	"rinha-backend-go/persistence"
	"rinha-backend-go/persistence/postgres/models"
	"rinha-backend-go/person"

	_ "github.com/amacneil/dbmate/v2/pkg/driver/postgres"
	_ "github.com/lib/pq"
)

const (
	selectPeople = `
    SELECT id,uuid,name,nickname,birthdate,stack,created_at
    FROM people`
)

type PostgresStore struct {
	queries *models.Queries
	db      *sql.DB
}

func (s *PostgresStore) GetPeopleCount(ctx context.Context) (int64, error) {
	return s.queries.CountPeople(ctx)
}

func (s *PostgresStore) AddPerson(ctx context.Context, p person.Person) (int64, error) {
	personUUID, err := uuid.Parse(p.UUID)
	if err != nil {
		return 0, err
	}

	stack := p.Stack

	if stack == nil {
		stack = []string{}
	}

	res, err := s.queries.AddPerson(ctx, models.AddPersonParams{
		Name:      p.Name,
		Uuid:      personUUID,
		Nickname:  p.Nickname,
		Birthdate: p.Birthdate,
		Stack:     stack,
	},
	)

	if err != nil {
		return 0, err
	}

	return int64(res), nil
}

func convertPersonDBToPerson(p models.Person) (*person.Person, error) {
	return &person.Person{
		ID:        int(p.ID),
		UUID:      p.Uuid.String(),
		Name:      p.Name,
		Nickname:  p.Nickname,
		Birthdate: p.Birthdate,
		Stack:     p.Stack,
		CreatedAt: p.CreatedAt.Time,
	}, nil
}

func (s *PostgresStore) GetPerson(ctx context.Context, uid string) (*person.Person, error) {
	personUUID, err := uuid.Parse(uid)
	if err != nil {
		return nil, err
	}

	p, err := s.queries.GetPerson(ctx, personUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, persistence.ErrPersonNotFound
		}
		return nil, err
	}
	return convertPersonDBToPerson(p)
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

func (s *PostgresStore) GetPeople(_ context.Context, options *persistence.GetPeopleOptions) (person.People, error) {
	query := selectPeople

	optionsValues := []interface{}{}

	if options != nil && options.SearchQuery != "" {
		query += " WHERE name LIKE $1 OR nickname LIKE $2 "
		optionsValues = append(optionsValues, containsQuery(options.SearchQuery),
			containsQuery(options.SearchQuery))
	}

	if options != nil && options.PaginationToken != "" {
		if options.SearchQuery == "" {
			query += " WHERE "
		} else {
			query += " AND "
		}
		query += "(id > ? and created_at >= ?) "
		id, createdAt := extractValuesFromPaginationToken(options.PaginationToken)
		optionsValues = append(optionsValues, id, createdAt)

	}

	query += " ORDER BY id ASC LIMIT 5;"

	rows, err := s.db.Query(query, optionsValues...)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	defer rows.Close()

	var people person.People

	for rows.Next() {
		var p models.Person
		err := rows.Scan(
			&p.ID,
			&p.Uuid,
			&p.Name,
			&p.Nickname,
			&p.Birthdate,
			pq.Array(&p.Stack),
			&p.CreatedAt,
		)

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

func (s *PostgresStore) Close() error {
	return s.db.Close()
}

//go:embed migrations/*.sql
var fs embed.FS

func NewPostgresStore(dsn string) (*PostgresStore, error) {
	u, err := url.Parse(dsn)

	if err != nil {
		return nil, err
	}

	dbm := dbmate.New(u)

	dbm.FS = fs
	dbm.MigrationsDir = []string{"migrations"}

	err = dbm.CreateAndMigrate()
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(30)
	db.SetMaxIdleConns(5)

	q := models.New(db)

	return &PostgresStore{db: db, queries: q}, nil
}
