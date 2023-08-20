package persistence

import (
	"context"
	"errors"

	"rinha-backend-go/person"
)

type GetPeopleOptions struct {
	PaginationToken string
	SearchQuery     string
}

type Store interface {
	AddPerson(context.Context, person.Person) (int64, error)
	GetPeople(ctx context.Context, options *GetPeopleOptions) (person.People, error)
	GetPerson(context.Context, string) (*person.Person, error)
	GetPeopleCount(ctx context.Context) (int64, error)
}

var (
	ErrPersonNotFound = errors.New("Person not found")
)
