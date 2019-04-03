package payments

import (
	"encoding/json"
	"fmt"
	. "github.com/pedro-gutierrez/form3/pkg/util"
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

// Payment attributes captures all detaled information
// about a payment. Here we are only capturing the Amount, for convenience,
// and we're leaving out the rest.
type PaymentAttributes struct {
	Amount string `json:"amount"`

	// TODO: add support for the rest of payment data
	// eg. beneficiary_party, charges_information, etc..
}

// Validate does semantic validation on the payment attributes
func (pa *PaymentAttributes) Validate() error {

	amount, err := strconv.ParseFloat(pa.Amount, 64)
	if err != nil {
		return errors.Wrap(err, "Invalid payment amount")
	}

	if amount <= 0 {
		return fmt.Errorf("Payment amount must be positive")
	}

	return nil
}

// Payment a payment
type Payment struct {
	Id           string            `json:"id"`
	Type         string            `json:"type"`
	Version      int               `json:"version"`
	Organisation string            `json:"organisation_id"`
	Attributes   PaymentAttributes `json:"attributes"`
}

// Validate does semantic validation on the payment
func (p *Payment) Validate() error {

	// check the id is not empty
	if len(strings.TrimSpace(p.Id)) == 0 {
		return errors.New("Id is empty")
	}

	// check the type
	if p.Type != "Payment" {
		return fmt.Errorf("Invalid type: %s", p.Type)
	}

	// check the organisation
	if len(strings.TrimSpace(p.Organisation)) == 0 {
		return errors.New("Organisation is empty")
	}

	// check the attributes
	return p.Attributes.Validate()
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
	var attrs PaymentAttributes
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
// a payment in its field 'data' and set of links
type PaymentResponse struct {
	Data  *Payment `json:"data"`
	Links Links    `json:"links"`
}

// A simple type to add restful links to our responses
type Links map[string]string

// PaymentsResponse represents a http response that contains
// a list of payments in its field 'data' and a set of links
type PaymentsResponse struct {
	Data  []*Payment `json:"data"`
	Links Links      `json:"links"`
}
