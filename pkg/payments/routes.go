// payments contains the http routes that perform CRUD
// operations on payments
package payments

import (
	"encoding/json"
	"github.com/go-chi/chi"
	. "github.com/pedro-gutierrez/form3/pkg/util"
	"net/http"
	"strconv"
)

// PaymentsService represents a payments service
// it defines the routes and the repo to operate
// with
type PaymentsService struct {
	// The database to operate with
	repo Repo
}

// New creates a new PaymentsService
func New(repo Repo) *PaymentsService {
	return &PaymentsService{repo: repo}
}

// Routes returns a router with all routes
// supported by this service
func (s *PaymentsService) Routes() *chi.Mux {
	router := chi.NewRouter()
	router.Get("/payments", s.List)
	router.Get("/payments/{id}", s.Fetch)
	router.Post("/payments", s.Create)
	router.Put("/payments/{id}", s.Update)
	router.Delete("/payments/{id}", s.Delete)
	return router
}

// List returns a list of payments. We return finite lists of payments
// so we need to check the from and to query params, and make sure
// they make sense
func (s *PaymentsService) List(w http.ResponseWriter, r *http.Request) {

	// convert the from param
	// into a integer or bad request
	from, err := strconv.Atoi(r.URL.Query().Get("from"))
	if err != nil {
		HandleHttpError(w, r, http.StatusBadRequest, err)
	}

	// convert the to param
	// into a integer or bad request
	to, err := strconv.Atoi(r.URL.Query().Get("to"))
	if err != nil {
		HandleHttpError(w, r, http.StatusBadRequest, err)
	}

	limit := to - from

	// from has to be less than to
	if limit <= 0 {
		HandleHttpError(w, r, http.StatusBadRequest, err)
	}

	// limit results
	if limit > 20 {
		limit = 20
	}

	repoItems, err := s.repo.List(from, limit)
	if err != nil {
		HandleHttpError(w, r, http.StatusInternalServerError, err)
		return
	} else {
		// Adapt repo data to payment data
		payments, err := NewPaymentsFromRepoItems(repoItems)
		if err != nil {
			HandleHttpError(w, r, http.StatusInternalServerError, err)
			return
		}

		links := NewLinks()

		// Send back the response
		RenderJSON(w, r, http.StatusOK, &PaymentsResponse{
			Data:  payments,
			Links: links,
		})
	}
}

// Fetch a payment by id
func (s *PaymentsService) Fetch(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	found, err := s.repo.Fetch(&RepoItem{Id: id})
	if err != nil {
		// Look for not found errors
		if s.repo.IsNotFound(err) {
			HandleHttpError(w, r, http.StatusNotFound, err)
		} else {
			HandleHttpError(w, r, http.StatusInternalServerError, err)
		}
		return
	}

	p, err := NewPaymentFromRepoItem(found)
	if err != nil {
		HandleHttpError(w, r, http.StatusInternalServerError, err)
		return
	}

	// Send back the response
	RenderJSON(w, r, http.StatusOK, &PaymentResponse{
		Data: p,
	})
}

// Delete a payment by id
func (s *PaymentsService) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// convert the version information
	// into a integer or bad request
	version, err := strconv.Atoi(r.URL.Query().Get("version"))
	if err != nil {
		HandleHttpError(w, r, http.StatusBadRequest, err)
	}

	// Do a lookup in order to return a proper 404 if
	// no record with that id exists
	_, err = s.repo.Fetch(&RepoItem{Id: id})
	if err != nil {
		// Look for not found errors
		if s.repo.IsNotFound(err) {
			HandleHttpError(w, r, http.StatusNotFound, err)
		} else {
			HandleHttpError(w, r, http.StatusInternalServerError, err)
		}
		return
	}

	// Delete the item from the repo assuming we are on the
	// right version
	err = s.repo.Delete(&RepoItem{Id: id, Version: version})
	if err != nil {
		// Look for not found errors again
		errorCode := http.StatusInternalServerError
		if s.repo.IsNotFound(err) || s.repo.IsConflict(err) {
			// The item was deleted in between, from a different goroutine
			// or was updated and the version increased. We treat both
			// cases as a concurren modification, that we
			// translate into a 409 Conflict
			errorCode = http.StatusConflict
		}
		HandleHttpError(w, r, errorCode, err)
		return
	}

	// Send back a 204
	RenderNoContent(w, r)
}

// Create a new payment
func (s *PaymentsService) Create(w http.ResponseWriter, r *http.Request) {
	// decode the incoming json payload
	p, err := decodePayment(r)
	if err != nil {
		// Something wrong with the JSON
		// Translate this into a 400 Bad Request and finish
		// the request
		HandleHttpError(w, r, http.StatusBadRequest, err)
		return
	}

	// try to save it. The database
	// will do whatever integrity checks are necessary
	repoItem, err := p.ToRepoItem()
	if err != nil {
		// check for conflicts
		// or internal errors
		HandleHttpError(w, r, http.StatusInternalServerError, err)
		return
	}

	// Create the repo item for the payment
	// The store implementation does its own consistency
	// concurrency and locking strategy.
	createdItem, err := s.repo.Create(repoItem)
	if err != nil {
		if s.repo.IsConflict(err) {
			// We have a conflict, so return the appropiate
			// status code
			HandleHttpError(w, r, http.StatusConflict, err)
		} else {
			HandleHttpError(w, r, http.StatusInternalServerError, err)
		}

		return
	}

	p, err = NewPaymentFromRepoItem(createdItem)
	if err != nil {
		HandleHttpError(w, r, http.StatusInternalServerError, err)
		return
	}

	// Everything went fine. Confirm back to the client
	RenderJSON(w, r, http.StatusCreated, &PaymentResponse{
		Data: p,
	})
}

// Update an existing payment
func (s *PaymentsService) Update(w http.ResponseWriter, r *http.Request) {
	// decode the incoming json payload
	p, err := decodePayment(r)
	if err != nil {
		// Something is wrong with the JSON
		// Translate this into a 400 Bad Request and finish
		// the request
		HandleHttpError(w, r, http.StatusBadRequest, err)
		return
	}

	id := chi.URLParam(r, "id")
	// check the id of the payment body and the id
	// from the path parameters. Return a bad request if they differ
	if p.Id != "" && id != p.Id {
		HandleHttpError(w, r, http.StatusBadRequest, err)
		return
	}

	// Perform a lookup in order to return a proper 404
	// code if no record with that id exists
	_, err = s.repo.Fetch(&RepoItem{Id: id})
	if err != nil {
		// Look for not found errors
		if s.repo.IsNotFound(err) {
			HandleHttpError(w, r, http.StatusNotFound, err)
		} else {
			HandleHttpError(w, r, http.StatusInternalServerError, err)
		}
		return
	}

	// Convert the payment into a repo item
	// Further validations can be done here, so we need
	// to handle errors
	repoItem, err := p.ToRepoItem()
	if err != nil {
		HandleHttpError(w, r, http.StatusInternalServerError, err)
		return
	}

	// Update the payment. the repo implementation
	// will implement the most appropriate concurrency and locking
	// statregy
	updatedItem, err := s.repo.Update(repoItem)
	if err != nil {
		if s.repo.IsConflict(err) {
			// We have a conflict, so return the appropiate
			// status code
			HandleHttpError(w, r, http.StatusConflict, err)
		} else {
			HandleHttpError(w, r, http.StatusInternalServerError, err)
		}
		return
	}

	p, err = NewPaymentFromRepoItem(updatedItem)
	if err != nil {
		HandleHttpError(w, r, http.StatusBadRequest, err)
		return
	}

	// Everything went fine, Confirm by returning the payment
	// back to the client
	RenderJSON(w, r, http.StatusOK, &PaymentResponse{
		Data: p,
	})
}

// decodePayment is a convenience function that attempts to
// decode a payment from the HTTP request body.
func decodePayment(r *http.Request) (*Payment, error) {
	decoder := json.NewDecoder(r.Body)
	var pr PaymentRequest
	err := decoder.Decode(&pr)
	return pr.Payment, err
}
