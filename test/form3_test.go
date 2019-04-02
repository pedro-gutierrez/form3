package main

import (
	"flag"
	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/colors"
	. "github.com/pedro-gutierrez/form3/pkg/test"
	"log"
	"os"
	"testing"
)

var (
	opt        = godog.Options{Output: colors.Colored(os.Stdout)}
	serverUrl  *string
	apiVersion *string
)

func init() {
	serverUrl = flag.String("server-url", "http://localhost:8080", "the payments server url to test against")
	apiVersion = flag.String("api-version", "v1", "the api version")
	godog.BindFlags("godog.", flag.CommandLine, &opt)
}

func TestMain(m *testing.M) {
	flag.Parse()
	opt.Paths = flag.Args()

	status := godog.RunWithOptions("form3", func(s *godog.Suite) {
		FeatureContext(s)
	}, opt)

	if st := m.Run(); st > status {
		status = st
	}
	os.Exit(status)
}

// FeatureContext is a Godog hook that bind Gherkin
// steps to Golang functions from our World structure.
// All test related types and step implementations are located in package
// github.com/pedro-gutierrez/form3/pkg/test
func FeatureContext(s *godog.Suite) {

	// Build a new World. This instance will be shared accross all scenarios
	w := NewWorld(*serverUrl, *apiVersion)

	// Make sure our scenario data is reset before each scenario
	// so that we do not incurr into side effects
	s.BeforeScenario(func(interface{}) {

		// Initialize new scenario data
		w.NewData()

		// Also ensure the service is up, and there is no data
		// from previous scenarios or run
		// All the existing data will be lost, so use with care
		err := DoThen(w.TheServiceIsUp(), func() error {
			return w.ThereAreNoPayments()
		})

		// Best effort.
		if err != nil {
			log.Printf("Error before scenario:")
			log.Printf("%v", err)
		}
	})

	s.Step(`^the service is up$`, w.TheServiceIsUp)
	s.Step(`^there are no payments$`, w.ThereAreNoPayments)
	s.Step(`^I query the health endpoint$`, w.IQueryTheHealthEndpoint)
	s.Step(`^I query the metrics endpoint$`, w.IQueryTheMetricsEndpoint)
	s.Step(`^I should have a json$`, w.IShouldHaveAJson)
	s.Step(`^I should have a text$`, w.IShouldHaveAText)
	s.Step(`^I should have status code (\d+)$`, w.IShouldHaveStatusCode)
	s.Step(`^I should have content-type (.*)$`, w.IShouldHaveContentType)
	s.Step(`^that json should have string at (.*) equal to (.*)$`, w.ThatJsonShouldHaveString)
	s.Step(`^that json should have int at (.*) equal to (.*)$`, w.ThatJsonShouldHaveInt)
	s.Step(`^that json should have (\d+) items$`, w.ThatJsonShouldHaveItems)
	s.Step(`^that json should have an (.*)$`, w.ThatJsonShouldHaveA)
	s.Step(`^that json should have a (.*)$`, w.ThatJsonShouldHaveA)
	s.Step(`^that text should match (.*)$`, w.ThatTextShouldMatch)
	s.Step(`^I get all payments$`, w.IGetAllPayments)
	s.Step(`^I get payments (\d+) to (\d+)$`, w.IGetPaymentsFromTo)
	s.Step(`^I get payments without from/to$`, w.IGetPaymentsWithoutFromTo)
	s.Step(`^a payment with id (.*)$`, w.APaymentWithId)
	s.Step(`^a payment with id (.*), no organisation$`, w.APaymentWithIdNoOrganisation)
	s.Step(`^a payment with id (.*), amount (.*)$`, w.APaymentWithIdAmount)
	s.Step(`^I create that payment$`, w.ICreateThatPayment)
	s.Step(`^I update that payment$`, w.IUpdateThatPayment)
	s.Step(`^I delete that payment$`, w.IDeleteThatPayment)
	s.Step(`^I get that payment$`, w.IGetThatPayment)
	s.Step(`^I created a new payment with id (.*)$`, w.ICreatedANewPaymentWithId)
	s.Step(`^I created (\d+) payments$`, w.ICreatedPayments)
	s.Step(`^I should have (\d+) payment\(s\)$`, w.IShouldHavePayments)
	s.Step(`^I deleted that payment$`, w.IDeletedThatPayment)
	s.Step(`^I updated that payment$`, w.IUpdatedThatPayment)
	s.Step(`^I delete version (\d+) of that payment$`, w.IDeleteVersionOfThatPayment)
	s.Step(`^I delete that payment, without saying which version$`, w.IDeleteThatPaymentWithoutSayingWhichVersion)
	s.Step(`^I update version (\d+) of that payment$`, w.IUpdateVersionOfThatPayment)
	s.Step(`^that payment has version (\d+)$`, w.ThatPaymentHasVersion)
}
