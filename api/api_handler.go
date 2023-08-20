package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"rinha-backend-go/persistence"
	"rinha-backend-go/person"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/valyala/fasthttp"
)

type PeopleHandler struct {
	store persistence.Store
	cache *redis.Client
}

func (h *PeopleHandler) AddPerson(ctx *fiber.Ctx) error {
	var request AddPersonRequest

	err := ctx.BodyParser(&request)

	if err != nil {
		log.Println(err)
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(ErrorResponse{Error: err.Error()})
	}

	birthdate, err := time.Parse("2006-01-02", request.Birthdate)

	if err != nil {
		log.Println(err)
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(ErrorResponse{Error: ErrInvalidBirthdate.Error()})
	}

	err = request.Validate()

	if err != nil {
		log.Println(err)
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(ErrorResponse{Error: err.Error()})
	}

	personUUID, err := uuid.NewV4()

	if err != nil {
		log.Println(err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: err.Error()})
	}

	person := person.Person{
		Name:      request.Name,
		UUID:      personUUID.String(),
		Nickname:  request.Nickname,
		Birthdate: birthdate,
		Stack:     request.Stack,
	}

	_, err = h.store.AddPerson(ctx.Context(), person)

	if err != nil {
		log.Println(err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: err.Error()})
	}

	personJSONCache, err := json.Marshal(person)

	if err != nil {
		log.Println(err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: err.Error()})
	}

	err = h.cache.Set(ctx.Context(), person.UUID, personJSONCache, 0).Err()

	if err != nil {
		log.Println(err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: err.Error()})
	}

	ctx.Set(fiber.HeaderLocation, fmt.Sprintf("/pessoas/%v", person.UUID))

	return ctx.Status(fiber.StatusCreated).JSON(AddPersonResponse{UUID: person.UUID})
}

// generatePaginationToken generates a pagination token based on the last person in the slice
// The token is a string in the format: <person_id>-<unix_timestamp>
func generatePaginationToken(people person.People) string {
	if len(people) == 0 {
		return ""
	}

	lastPerson := people[len(people)-1]

	return fmt.Sprintf("%v-%v", lastPerson.ID, lastPerson.CreatedAt.Unix())
}

func (h *PeopleHandler) GetPeople(ctx *fiber.Ctx) error {

	t := ctx.Query("t")

	if t == "" {
		return ctx.Status(http.StatusBadRequest).SendString("O parâmetro 't' é obrigatório")
	}

	options := &persistence.GetPeopleOptions{
		PaginationToken: ctx.Query("pagina"),
		SearchQuery:     ctx.Query("t"),
	}

	people, err := h.store.GetPeople(ctx.Context(), options)

	if err != nil {
		log.Println(err)
		return ctx.Status(http.StatusInternalServerError).SendString(err.Error())
	}

	response := GetPeopleResponse{Resultados: people}

	if ctx.Query("pagina") != "" {
		var prevUrl = fasthttp.URI{}
		// ctx.Request().URI().CopyTo(&prevUrl)
		prevUrl.SetHost(string(ctx.Request().URI().Host()))
		prevUrl.SetScheme(string(ctx.Request().URI().Scheme()))
		prevUrl.SetPath("/pessoas")
		queryValues := prevUrl.QueryArgs()

		paginationStack := ctx.Query("paginationStack")

		if paginationStack != "" {
			previousPaginationToken := strings.Split(paginationStack, ",")
			queryValues.Set("pagina", previousPaginationToken[len(previousPaginationToken)-1])
			queryValues.Set("paginationStack", strings.Join(previousPaginationToken[:len(previousPaginationToken)-1], ","))
			if len(previousPaginationToken)-1 == 0 {
				queryValues.Del("paginationStack")
			}
		}

		anterior := prevUrl.String()
		response.Anterior = &anterior
	}

	if len(people) == 5 {
		var nextUrl = fasthttp.URI{}
		ctx.Request().URI().CopyTo(&nextUrl)
		queryValues := nextUrl.QueryArgs()

		if ctx.Query("pagina") != "" {
			queryValues.Set("paginationStack", ctx.Query("paginationStack")+","+ctx.Query("pagina"))
		}

		queryValues.Set("pagina", generatePaginationToken(people))
		proxima := nextUrl.String()
		response.Proxima = &proxima
	}

	return ctx.JSON(response)
}

func (h *PeopleHandler) GetPerson(ctx *fiber.Ctx) error {
	personID := ctx.Params("id")

	cachedPerson, err := h.cache.Get(ctx.Context(), personID).Result()

	if err != nil {
		log.Println(err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: err.Error()})
	}

	if cachedPerson != "" {
		ctx.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		_, err = ctx.Write([]byte(cachedPerson))
		return err
	}

	person, err := h.store.GetPerson(ctx.Context(), personID)

	if err != nil {
		if err == persistence.ErrPersonNotFound {
			return ctx.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: err.Error()})
		}
		log.Println(err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: err.Error()})
	}

	return ctx.JSON(person)
}

func (h *PeopleHandler) GetPeopleCount(ctx *fiber.Ctx) error {
	count, err := h.store.GetPeopleCount(ctx.Context())

	if err != nil {
		log.Println(err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: err.Error()})
	}

	_, err = ctx.Write([]byte(strconv.FormatInt(count, 10)))
	return err
}
