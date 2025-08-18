package fibernetia

import (
	"bytes"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

//nolint:gocognit
func TestInertia_Middleware(t *testing.T) {
	t.Parallel()

	t.Run("plain request", func(t *testing.T) {
		t.Parallel()

		t.Run("do nothing, call next handler", func(t *testing.T) {
			t.Parallel()

			r := requestMock(fiber.MethodGet, "/")

			I().Middleware(assertHandlerServed(t))

			assertInertiaVary(t, r.Response())
			assertResponseStatusCode(t, r.Response(), fiber.StatusOK)
		})

		t.Run("flash", func(t *testing.T) {
			t.Parallel()

			t.Run("validation errors", func(t *testing.T) {
				t.Parallel()

				r := requestMock(fiber.MethodGet, "/")

				want := ValidationErrors{
					"foo": "baz",
					"baz": "quz",
				}

				store := session.New()
				flashProvider := NewSessionFlashProvider(store)

				i := I(func(i *Inertia) {
					i.flash = flashProvider
				})

				var got ValidationErrors
				i.Middleware(func(c *fiber.Ctx) error {
					got = ValidationErrorsFromContext(c)
					return c.Next()
				})(r)

				if !reflect.DeepEqual(got, want) {
					t.Fatalf("validation errors=%#v, want=%#v", got, want)
				}
			})

			t.Run("clear history", func(t *testing.T) {
				t.Parallel()

				r := requestMock(fiber.MethodGet, "/")

				store := session.New()
				flashProvider := NewSessionFlashProvider(store)

				i := I(func(i *Inertia) {
					i.flash = flashProvider
				})

				var got bool
				i.Middleware(func(c *fiber.Ctx) error {
					got = ClearHistoryFromContext(c)
					return c.Next()
				})(r)

				if !got {
					t.Fatalf("clear history=%v, want=true", got)
				}
			})
		})
	})

	t.Run("inertia request", func(t *testing.T) {
		t.Parallel()

		t.Run("assert versioning", func(t *testing.T) {
			t.Parallel()

			t.Run("diff version with GET, should change location with 409 and flash errors", func(t *testing.T) {
				t.Parallel()

				errors := ValidationErrors{
					"foo": "baz",
					"baz": "quz",
				}

				flashProvider := &flashProviderMock{}

				i := I(func(i *Inertia) {
					i.version = "foo"
					i.flash = flashProvider
				})

				r := requestMock(fiber.MethodGet, "https://example.com/home")
				asInertiaRequest(r)
				withInertiaVersion(r, "bar")

				i.Middleware(assertHandlerServed(t, setInertiaResponseHandler, successJSONHandler))(r)

				assertInertiaNotVary(t, r.Response())
				assertNotInertiaResponse(t, r.Response())
				assertResponseStatusCode(t, r.Response(), fiber.StatusConflict)
				assertInertiaLocation(t, r.Response(), "/home")

				if !reflect.DeepEqual(flashProvider.validationErrors, errors) {
					t.Fatalf("got validation errors=%#v, want=%#v", flashProvider.validationErrors, errors)
				}
			})

			t.Run("diff version with POST, do nothing", func(t *testing.T) {
				t.Parallel()

				i := I(func(i *Inertia) {
					i.version = "foo"
				})

				r := requestMock(fiber.MethodPost, "/home")
				asInertiaRequest(r)
				withInertiaVersion(r, "bar")

				i.Middleware(assertHandlerServed(t, successJSONHandler))(r)

				assertInertiaVary(t, r.Response())
				assertResponseStatusCode(t, r.Response(), fiber.StatusOK)
			})
		})

		t.Run("redirect back if empty response body", func(t *testing.T) {
			t.Parallel()

			t.Run("redirect back if empty request and status ok", func(t *testing.T) {
				t.Parallel()

				r := requestMock(fiber.MethodGet, "/")
				asInertiaRequest(r)
				withReferer(r, "/foo")

				I().Middleware(assertHandlerServed(t))(r)

				assertInertiaVary(t, r.Response())
				assertResponseStatusCode(t, r.Response(), fiber.StatusFound)
				assertLocation(t, r.Response(), "/foo")
				assertInertiaLocation(t, r.Response(), "")
			})

			t.Run("don't redirect back if empty request and status not ok", func(t *testing.T) {
				t.Parallel()

				r := requestMock(fiber.MethodGet, "/")
				asInertiaRequest(r)
				withReferer(r, "/foo")

				I().Middleware(assertHandlerServed(t, errorJSONHandler))(r)

				assertInertiaVary(t, r.Response())
				assertResponseStatusCode(t, r.Response(), fiber.StatusBadRequest)
				assertLocation(t, r.Response(), "")
				assertInertiaLocation(t, r.Response(), "")
			})
		})

		t.Run("POST, PUT and PATCH requests cannot have the status 302", func(t *testing.T) {
			t.Parallel()

			t.Run("GET can have 302 status", func(t *testing.T) {
				t.Parallel()

				r := requestMock(fiber.MethodGet, "/")
				asInertiaRequest(r)

				I().Middleware(assertHandlerServed(t, setStatusHandler(fiber.StatusFound)))(r)

				assertInertiaVary(t, r.Response())
				assertResponseStatusCode(t, r.Response(), fiber.StatusFound)
			})

			for _, method := range []string{fiber.MethodPut, fiber.MethodPatch, fiber.MethodDelete} {
				t.Run(method+" cannot have 302 status", func(t *testing.T) {
					t.Parallel()

					r := requestMock(fiber.MethodGet, "/")
					asInertiaRequest(r)

					I().Middleware(assertHandlerServed(t, setStatusHandler(fiber.StatusFound)))(r)

					assertInertiaVary(t, r.Response())
					assertResponseStatusCode(t, r.Response(), fiber.StatusSeeOther)
				})
			}
		})

		t.Run("success", func(t *testing.T) {
			t.Parallel()

			t.Run("with new response writer", func(t *testing.T) {
				t.Parallel()

				r := requestMock(fiber.MethodGet, "/")
				asInertiaRequest(r)

				handlers := []fiber.Handler{
					successJSONHandler,
					setHeadersHandler(map[string]string{
						"foo": "bar",
					}),
				}

				I().Middleware(assertHandlerServed(t, handlers...))(r)

				assertInertiaVary(t, r.Response())
				assertResponseStatusCode(t, r.Response(), fiber.StatusOK)

				if !reflect.DeepEqual(r.Response().Body(), successJSON) {
					t.Fatalf("JSON=%#v, want=%#v", r.Response().Body(), successJSON)
				}

				gotHeader := string(r.Response().Header.Peek("foo"))
				wantHeader := "bar"

				if gotHeader != wantHeader {
					t.Fatalf("header=%#v, want=%#v", gotHeader, wantHeader)
				}
			})

			t.Run("with passed response writer", func(t *testing.T) {
				t.Parallel()

				r := requestMock(fiber.MethodGet, "/")
				asInertiaRequest(r)

				buf := bytes.NewBufferString(successJSON)

				i := I()

				wrap := &inertiaResponseWrapper{
					statusCode: fiber.StatusNotFound,
					buf:        buf,
					header:     map[string]string{"foo": "bar"},
				}

				I().Middleware(assertHandlerServed(t, successJSONHandler))(r)
				i.copyWrapperResponse(r, wrap)

				assertInertiaVary(t, r.Response())
				assertResponseStatusCode(t, r.Response(), fiber.StatusNotFound)

				if !reflect.DeepEqual(r.Response().Body(), successJSON+successJSON) {
					t.Fatalf("JSON=%#v, want=%#v", r.Response().Body(), successJSON)
				}

				gotHeader := string(r.Response().Header.Peek("foo"))
				wantHeader := "bar"

				if gotHeader != wantHeader {
					t.Fatalf("header=%#v, want=%#v", gotHeader, wantHeader)
				}
			})
		})
	})
}

var (
	successJSON = `{"success": true}`
	errorJSON   = `{"success": false}`
)

func successJSONHandler(ctx *fiber.Ctx) error {
	_, _ = ctx.Write([]byte(successJSON))
	return nil
}

func errorJSONHandler(ctx *fiber.Ctx) error {
	_, _ = ctx.Write([]byte(errorJSON))
	ctx.Status(http.StatusBadRequest)
	return nil
}

func setStatusHandler(status int) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Response().Header.Set("X-Status", fmt.Sprintf("%d", status))
		return nil
	}
}

func setHeadersHandler(headers map[string]string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		for key, val := range headers {
			c.Response().Header.Set(key, val)
		}
		return nil
	}
}

func setInertiaResponseHandler(ctx *fiber.Ctx) error {
	setInertiaInResponse(ctx)
	return nil
}
