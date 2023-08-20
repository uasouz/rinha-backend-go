package person

import (
	"encoding/json"
	"time"
)

type Person struct {
	ID        int       `json:"-"`
	UUID      string    `json:"uuid"`
	Name      string    `json:"name"`
	Nickname  string    `json:"apelido"`
	Birthdate time.Time `json:"nascimento"`
	Stack     []string  `json:"stack"`
	CreatedAt time.Time `json:"-"`
}

// MarshalJSON customizes the JSON output for the Birthdate field
func (p *Person) MarshalJSON() ([]byte, error) {
	type Alias Person
	return json.Marshal(&struct {
		Birthdate string `json:"nascimento"`
		*Alias
	}{
		Birthdate: p.Birthdate.Format("2006-01-02"),
		Alias:     (*Alias)(p),
	})
}

type People []*Person
