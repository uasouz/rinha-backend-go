package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"rinha-backend-go/persistence/sqlite"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/suite"
)

type APITestSuite struct {
	suite.Suite
	app *fiber.App
}

func (s *APITestSuite) SetupSuite() {
	store, err := sqlite.NewSQLiteStore()
	s.Require().NoError(err)

	s.app = fiber.New(
		fiber.Config{
			JSONEncoder: json.Marshal,
			JSONDecoder: json.Unmarshal,
		},
	)

	handler := PeopleHandler{store: store}

	s.app.Get("/pessoas", handler.GetPeople)
	s.app.Post("/pessoas", handler.AddPerson)
	s.app.Get("/pessoas/:id", handler.GetPerson)

}

func (s *APITestSuite) TestAddPerson() {
	request := AddPersonRequest{
		Name:      "John Doe",
		Nickname:  "johndoe",
		Birthdate: "1990-01-01",
		Stack:     []string{"Go", "Python"},
	}

	jsonRequest, _ := json.Marshal(request)

	req, err := http.NewRequest("POST", "/pessoas", bytes.NewReader(jsonRequest))
	req.Header.Add("Content-Type", "application/json")

	s.Require().NoError(err)

	resp, err := s.app.Test(req, 3)

	s.Require().NoError(err)

	bytes, _ := io.ReadAll(resp.Body)
	fmt.Println(string(bytes))

	s.Equal(http.StatusCreated, resp.StatusCode)
}

func (s *APITestSuite) TestAddPersonInvalidBirthdate() {
	request := AddPersonRequest{
		Name:      "John Doe",
		Nickname:  "johndoe",
		Birthdate: "1990-01-01 00:00:00",
		Stack:     []string{"Go", "Python"},
	}

	jsonRequest, _ := json.Marshal(request)

	req, err := http.NewRequest("POST", "/pessoas", bytes.NewReader(jsonRequest))
	req.Header.Add("Content-Type", "application/json")

	if err != nil {
		s.Require().NoError(err)
	}

	resp, err := s.app.Test(req, 2)

	s.Equal(http.StatusUnprocessableEntity, resp.StatusCode)
}

func (s *APITestSuite) TestAddPersonInvalidStack() {
	request := AddPersonRequest{
		Name:      "João Ninguém",
		Nickname:  "johndoe",
		Birthdate: "1990-01-01",
		Stack:     []string{"Go", "Python", "C++", "Java", "C#", "Javascript/TypeScript"},
	}

	jsonRequest, _ := json.Marshal(request)

	req, err := http.NewRequest("POST", "/pessoas", bytes.NewReader(jsonRequest))
	req.Header.Add("Content-Type", "application/json")

	if err != nil {
		s.Require().NoError(err)
	}

	resp, err := s.app.Test(req, 2)

	s.Equal(http.StatusUnprocessableEntity, resp.StatusCode)
}

func (s *APITestSuite) TestGetPerson() {
	request := AddPersonRequest{
		Name:      "John Doe",
		Nickname:  "johndoe",
		Birthdate: "1990-01-01",
		Stack:     []string{"Go", "Python"},
	}

	jsonRequest, _ := json.Marshal(request)

	req, err := http.NewRequest("POST", "/pessoas", bytes.NewReader(jsonRequest))
	req.Header.Add("Content-Type", "application/json")

	if err != nil {
		s.Require().NoError(err)
	}

	resp, err := s.app.Test(req, 2)

	s.Equal(http.StatusCreated, resp.StatusCode)

	req, err = http.NewRequest("GET", resp.Header.Get("Location"), nil)

	if err != nil {
		s.Require().NoError(err)
	}

	resp, err = s.app.Test(req, 2)

	s.Equal(http.StatusOK, resp.StatusCode)
}

func TestAPI(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}
