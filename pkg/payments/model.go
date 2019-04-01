package payments

import (
	"encoding/json"
	. "github.com/pedro-gutierrez/form3/pkg/util"
	"github.com/pkg/errors"
	"strings"
)

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

	repoItem.Attributes = string(bytes)
	return repoItem, nil
}

// Converts a repo item into a payment
func NewPaymentFromRepoItem(item *RepoItem) (*Payment, error) {
	p := &Payment{
		Type:         "Payment",
		Id:           item.Id,
		Version:      item.Version,
		Organisation: item.Organisation,
	}

	// decode the attributes payload
	var attrs map[string]interface{}
	if item.Attributes != "" {
		err := json.NewDecoder(strings.NewReader(item.Attributes)).Decode(&attrs)
		if err != nil {
			return p, errors.Wrap(err, "Error parsing repo item attributes")
		}
	}
	p.Attributes = attrs
	return p, nil
}

// PaymentsFromRepoItems converts the given slice of repo
// items to a list of payments
func NewPaymentsFromRepoItems(items []*RepoItem) ([]*Payment, error) {
	payments := []*Payment{}
	for _, i := range items {
		p, err := NewPaymentFromRepoItem(i)
		if err != nil {
			return payments, err
		}
		payments = append(payments, p)
	}

	return payments, nil
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
