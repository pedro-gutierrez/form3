package payments

import (
	"encoding/json"
	. "github.com/pedro-gutierrez/form3/pkg/util"
	"github.com/pkg/errors"
)

// Health basic health status info
type Health struct {
	Status string `json:"status"`
}

// Payment a payment
type Payment struct {
	Id           string      `json:"id"`
	Type         string      `json:"type"`
	Version      int         `json:"version"`
	Organisation string      `json:"organisation"`
	Attributes   interface{} `json:"attributes"`
}

// Converts a payment into something that
// can be saved into the database
func (p *Payment) ToRepoItem() (*RepoItem, error) {
	repoItem := &RepoItem{
		Id:           p.Id,
		Version:      p.Version,
		Organisation: p.Organisation,
	}
	// Try to serialize the payment attributes
	bytes, err := json.Marshal(p.Attributes)
	if err != nil {
		return repoItem, errors.Wrap(err, "Unable to serialize payment attributes")
	}

	repoItem.Attributes = bytes
	return repoItem, nil
}

// Converts a repo item into a payment
func NewPaymentFromRepoItem(item *RepoItem) *Payment {
	return &Payment{
		Type:         "Payment",
		Id:           item.Id,
		Version:      item.Version,
		Organisation: item.Organisation,
		Attributes:   item.Attributes,
	}
}

// PaymentsFromRepoItems converts the given slice of repo
// items to a list of payments
func NewPaymentsFromRepoItems(items []*RepoItem) []*Payment {
	payments := []*Payment{}
	for _, i := range items {
		payments = append(payments, NewPaymentFromRepoItem(i))
	}

	return payments
}

// PaymentRequest represents a http request that contains
// a payment in its field 'data'
type PaymentRequest struct {
	Payment *Payment `json:"data"`
}

// PaymentResponse represents a http response that contains
// a payment in its field 'data'
type PaymentResponse struct {
	Data *Payment `json:"data"`
}

type Links map[string]string

// NewLinks initializes a new set of links
func NewLinks() Links {
	return make(map[string]string)
}

// PaymentsResponse represents a http response that contains
// a list of payments in its field 'data'
type PaymentsResponse struct {
	Data  []*Payment `json:"data"`
	Links Links      `json:"links"`
}
