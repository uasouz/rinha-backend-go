package api

import "rinha-backend-go/person"

type ErrorResponse struct {
	Error string `json:"error"`
}

type GetPeopleResponse struct {
	Qtd        int           `json:"qtd"`
	Pagina     int           `json:"pagina"`
	Anterior   *string       `json:"anterior"`
	Proxima    *string       `json:"proxima"`
	Resultados person.People `json:"resultados"`
}

type AddPersonResponse struct {
	UUID string `json:"uuid"`
}
