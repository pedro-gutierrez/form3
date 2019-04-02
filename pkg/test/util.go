package test

import (
	"errors"
	"fmt"
)

// expectThen is a convenience function that helps to chain
// an expectation after another. If the given msg is not empty
// then that means an expectation failed, and we translate it into
// an error. Otherwise we proceed with next.
func ExpectThen(msg string, next func() error) error {
	if msg != "" {
		return errors.New(msg)
	} else {
		return next()
	}
}

// expect is a convenience function that denotes this is the
// last expectation from the chain
func Expect(msg string) error {
	if msg != "" {
		return errors.New(msg)
	} else {
		return nil
	}
}

// doThen is similar to expectThen except that it works with
// actions that either return an error or nil. This is useful when
// chaining steps
func DoThen(err error, next func() error) error {
	if err != nil {
		return err
	} else {
		return next()
	}
}

// DoSequence performs the given step function, count times,
// in a sequence.
func DoSequence(step func(it int) error, count int) error {
	for i := 0; i < count; i++ {
		err := step(i)
		// break the loop as soon as we get an error
		if err != nil {
			return err
		}
	}
	return nil
}

// PaymentData is a simplified representation of
// a payment, to be used in BDDs. Most of fields will be set
// by default, but here we declare the ones that are
// really significant in our tests
type PaymentData struct {
	Id           string
	Version      int
	Organisation string
	Amount       string
}

// ToJSON returns a json string from the payment data
// Most values that are not critical for our tests
// will be set to arbitrary defaults
func (p *PaymentData) ToJSON() string {
	return fmt.Sprintf(`{ 
		"data": {
			"id": "%s",
			"type": "Payment",
			"version": %v,
			"organisation": "%s",
			"attributes": {
				"amount": "%s"
			}
		}
	}`, p.Id, p.Version, p.Organisation, p.Amount)
}

// a ScenarioData struct is data for a particular scenario
// Each scenario needs to have a clean struct, in order to avoid
// side effects
type ScenarioData struct {
	// Holds a simplified representation of a Payment
	PaymentData *PaymentData

	// Generic datastructure where steps might store data
	// and read from it
	Subject interface{}
}

// a World is a simple context or container for features
type World struct {
	serverUrl  string
	apiVersion string
	Client     *Client
	Data       *ScenarioData
}

// Return a new World container for the given server
// url and api version.
func NewWorld(serverUrl string, apiVersion string) *World {
	return &World{
		serverUrl:  serverUrl,
		apiVersion: apiVersion,
	}
}

// NewData creates a new, blank scenario data structure
func (w *World) NewData() {
	w.Data = &ScenarioData{}
	w.Client = NewClient(w.serverUrl)
}

// versionedPath returns the given path, prefixed with
// the configured api version
func (w *World) versionedPath(path string) string {
	return fmt.Sprintf("/%s%s", w.apiVersion, path)
}
