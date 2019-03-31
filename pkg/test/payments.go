package test

import (
	"fmt"
	"github.com/mdaverde/jsonpath"
	. "github.com/smartystreets/assertions"
	"reflect"
)

// TheServiceIsUp checks the health check is reponding propertly
func (w *World) TheServiceIsUp() error {
	return DoThen(w.IQueryTheHealthEndpoint(), func() error {
		return w.IShouldHaveStatusCode(200)
	})
}

// ThereAreNoPayments ensure there is no data. This requires the target
// system to be started with admin routes
func (w *World) ThereAreNoPayments() error {
	return DoThen(w.IDeleteAllData(), func() error {
		return DoThen(w.IShouldHaveStatusCode(200), func() error {
			return w.IShouldHavePayments(0)
		})
	})
}

// IDeleteAllData use the admin endpoints in order to delete all data
func (w *World) IDeleteAllData() error {
	path := w.versionedPath("/admin/repo")
	w.Client.Delete(path)
	return nil
}

// IQueryTheHealthEndpoint performs a GET on the healthcheck
// endpoint and stores the response details in the World
// context
func (w *World) IQueryTheHealthEndpoint() error {
	path := w.versionedPath("/health")
	w.Client.Get(path)
	return nil
}

// IGetAllPayments performs a GET on the payments
// endpoint and stores the response details in the World
// context
func (w *World) IGetAllPayments() error {
	path := w.versionedPath("/payments")
	w.Client.Get(path)
	return nil
}

// iQueryTheMetricsEndpoint performs a GET on the metrics
// endpoint and stores the response details in the World
// context
func (w *World) IQueryTheMetricsEndpoint() error {
	w.Client.Get("/metrics")
	return nil
}

// IShouldHaveStatusCode expects the client to have
// the given status code in its last response
func (w *World) IShouldHaveStatusCode(expected int) error {
	return ExpectThen(ShouldNotBeNil(w.Client.Resp), func() error {
		return Expect(ShouldEqual(w.Client.Resp.StatusCode, expected))
	})
}

// IShouldHaveContentType expects the client to have the
// given content type in its last response
func (w *World) IShouldHaveContentType(expected string) error {
	return ExpectThen(ShouldNotBeNil(w.Client.Resp), func() error {
		return Expect(ShouldContainSubstring(w.Client.Resp.Header.Get("content-type"), expected))
	})
}

// IShouldHaveAJson inspects the client's latest response and
// checks for a json document
func (w *World) IShouldHaveAJson() error {
	return ExpectThen(ShouldNotBeNil(w.Client.Json), func() error {
		w.Data.Subject = w.Client.Json
		return nil
	})
}

// IShouldHaveAText inspects the client's latest response and
// checks for a string
func (w *World) IShouldHaveAText() error {
	return ExpectThen(ShouldNotBeNil(w.Client.Text), func() error {
		w.Data.Subject = w.Client.Text
		return nil
	})
}

// ThatJsonShouldHaveString inspects the json, if any, in the client
// and looks for the given field, and verifies it is equal to the given
// expected string value
func (w *World) ThatJsonShouldHaveString(path string, expected string) error {
	return ExpectThen(ShouldNotBeNil(w.Data.Subject), func() error {
		actual, err := jsonpath.Get(w.Data.Subject, path)
		return ExpectThen(ShouldBeNil(err), func() error {
			return Expect(ShouldEqual(actual, expected))
		})
	})
}

// ThatJsonShouldHaveString inspects the json, if any, in the client
// and looks for the given field, and verifies it is equal to the given
// expected int value
func (w *World) ThatJsonShouldHaveInt(path string, expected int) error {
	return ExpectThen(ShouldNotBeNil(w.Data.Subject), func() error {
		actual, err := jsonpath.Get(w.Data.Subject, path)
		return ExpectThen(ShouldBeNil(err), func() error {
			return Expect(ShouldEqual(actual, expected))
		})
	})
}

// ThatTextShouldMatch inspects the json, if any, in the client
// and looks for the given field.
func (w *World) ThatTextShouldMatch(expected string) error {
	return ExpectThen(ShouldNotBeNil(w.Data.Subject), func() error {
		var text string
		return ExpectThen(ShouldEqual(reflect.TypeOf(w.Data.Subject), reflect.TypeOf(text)), func() error {
			text = w.Data.Subject.(string)
			return Expect(ShouldContainSubstring(text, expected))
		})
	})
}

// ThatJsonShouldHaveItems inspects the json, if any, and checks whether
// the data field contains an array with the expected number of items
func (w *World) ThatJsonShouldHaveItems(expected int) error {
	return ExpectThen(ShouldNotBeNil(w.Data.Subject), func() error {
		actual, err := jsonpath.Get(w.Data.Subject, "data")
		return ExpectThen(ShouldBeNil(err), func() error {
			var items []interface{}
			return ExpectThen(ShouldEqual(reflect.TypeOf(actual), reflect.TypeOf(items)), func() error {
				items = actual.([]interface{})
				return Expect(ShouldEqual(len(items), expected))
			})
		})
	})
}

// APaymentWithId defines a new payment in the current scenario context
// with the given client defined id
func (w *World) APaymentWithId(id string) error {
	w.Data.PaymentData = &PaymentData{
		Id:      id,
		Version: 1,
	}
	return nil
}

// ThatPaymentHasVersion updates the version of the payment data in the
// current scenario data
func (w *World) ThatPaymentHasVersion(v int) error {
	return ExpectThen(ShouldNotBeNil(w.Data.PaymentData), func() error {
		w.Data.PaymentData.Version = v
		return nil
	})
}

// ICreateThatPayment actually creates the payment defined in the world
// request data (as a string) by posting it to the payments endpoint, as json
func (w *World) ICreateThatPayment() error {
	path := w.versionedPath("/payments")
	w.Client.Post(path, w.Data.PaymentData.ToJSON())
	return nil
}

// IUpdateThatPayment sends a PUT request for the payment defined in the
// scenario data.
func (w *World) IUpdateThatPayment() error {
	return ExpectThen(ShouldNotBeNil(w.Data.PaymentData), func() error {
		p := w.Data.PaymentData
		path := w.versionedPath(fmt.Sprintf("/payments/%s", p.Id))
		w.Client.Put(path, p.ToJSON())
		return nil
	})
}

// IUpdateVersionOfThatPayment updates a specific version of payment
func (w *World) IUpdateVersionOfThatPayment(v int) error {
	return ExpectThen(ShouldNotBeNil(w.Data.PaymentData), func() error {
		p := w.Data.PaymentData
		p.Version = v
		return w.IUpdateThatPayment()
	})
}

// IDeleteThatPayment sends a DELETE request for the payment defined in the
// scenario data. The delete operation requires a version information
func (w *World) IDeleteThatPayment() error {
	return ExpectThen(ShouldNotBeNil(w.Data.PaymentData), func() error {
		p := w.Data.PaymentData
		path := w.versionedPath(fmt.Sprintf("/payments/%s?version=%v", p.Id, p.Version))
		w.Client.Delete(path)
		return nil
	})
}

// ICreatedANewPaymentWithId combines logic from previous steps in order
// to provide a convenience Given step for payment fixtures in more complex
// scenarios
func (w *World) ICreatedANewPaymentWithId(id string) error {
	return DoThen(w.APaymentWithId(id), func() error {
		return DoThen(w.ICreateThatPayment(), func() error {
			return w.IShouldHaveStatusCode(201)
		})
	})
}

// IDeletedThatPayment combines logic from previous steps in order
// to provide a convenience Given step for payment fixtures in more complex
// scenarios
func (w *World) IDeletedThatPayment() error {
	return DoThen(w.IDeleteThatPayment(), func() error {
		return w.IShouldHaveStatusCode(204)
	})
}

// IDeleteVersionOfThatPayment deletes a specific version of payment
func (w *World) IDeleteVersionOfThatPayment(v int) error {
	return ExpectThen(ShouldNotBeNil(w.Data.PaymentData), func() error {
		p := w.Data.PaymentData
		p.Version = v
		return w.IDeleteThatPayment()
	})
}

// IUpdatedThatPayment combines logic from previous steps in order
// to provide a convenience Given step for payment fixtures in more complex
// scenarios
func (w *World) IUpdatedThatPayment() error {
	return DoThen(w.IUpdateThatPayment(), func() error {
		return w.IShouldHaveStatusCode(200)
	})
}

// IShouldHavePayments fetches a list of payments and ensure the number
// of items returned is the expected
func (w *World) IShouldHavePayments(expected int) error {
	return DoThen(w.IGetAllPayments(), func() error {
		return DoThen(w.IShouldHaveAJson(), func() error {
			return w.ThatJsonShouldHaveItems(expected)
		})
	})
}
