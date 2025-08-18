package fibernetia

import (
	"io"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

func I(opts ...func(i *Inertia)) *Inertia {
	i := &Inertia{
		containerID:        "app",
		jsonMarshaller:     jsonDefaultMarshaller{},
		sharedProps:        make(Props),
		sharedTemplateData: make(TemplateData),
		logger:             log.New(io.Discard, "", 0),
	}

	for _, opt := range opts {
		opt(i)
	}

	return i
}

// requestMock creates a Fiber-compatible request context
func requestMock(method, target string) *fiber.Ctx {
	app := fiber.New()

	req := app.AcquireCtx(&fasthttp.RequestCtx{})

	req.Request().Header.SetMethod(method)
	req.Request().SetRequestURI(target)

	return req
}

func asInertiaRequest(c *fiber.Ctx) {
	c.Request().Header.Set("X-Inertia", "true")
}

func withOnly(c *fiber.Ctx, data []string) {
	c.Request().Header.Set("X-Inertia-Partial-Data", strings.Join(data, ","))
}

func withExcept(c *fiber.Ctx, data []string) {
	c.Request().Header.Set("X-Inertia-Partial-Except", strings.Join(data, ","))
}

func withReset(c *fiber.Ctx, data []string) {
	c.Request().Header.Set("X-Inertia-Reset", strings.Join(data, ","))
}

func withPartialComponent(c *fiber.Ctx, component string) {
	c.Request().Header.Set("X-Inertia-Partial-Component", component)
}

func withInertiaVersion(c *fiber.Ctx, ver string) {
	c.Request().Header.Set("X-Inertia-Version", ver)
}

func withReferer(c *fiber.Ctx, referer string) {
	c.Request().Header.Set("Referer", referer)
}

func withValidationErrors(c *fiber.Ctx, errors ValidationErrors) {
	c.Locals("validationErrors", errors)
}

func withClearHistory(c *fiber.Ctx) {
	c.Locals("clearHistory", true)
}

func assertResponseStatusCode(t *testing.T, resp *fiber.Response, want int) {
	t.Helper()

	if resp.StatusCode() != want {
		t.Fatalf("status=%d, want=%d", resp.StatusCode(), want)
	}
}

func assertHeader(t *testing.T, resp *fiber.Response, key, want string) {
	t.Helper()

	if got := string(resp.Header.Peek(key)); got != want {
		t.Fatalf("header %s=%s, want=%s", strings.ToLower(key), got, want)
	}
}

func assertHeaderMissing(t *testing.T, resp *fiber.Response, key string) {
	t.Helper()

	if got := string(resp.Header.Peek(key)); got != "" {
		t.Fatalf("unexpected header %s=%s, want=empty", key, got)
	}
}

func assertLocation(t *testing.T, resp *fiber.Response, want string) {
	t.Helper()

	assertHeader(t, resp, "Location", want)
}

func assertInertiaResponse(t *testing.T, resp *fiber.Response) {
	t.Helper()

	assertHeader(t, resp, "X-Inertia", "true")
}

func assertNotInertiaResponse(t *testing.T, resp *fiber.Response) {
	t.Helper()

	assertHeaderMissing(t, resp, "X-Inertia")
}

func assertInertiaLocation(t *testing.T, resp *fiber.Response, want string) {
	t.Helper()

	assertHeader(t, resp, "X-Inertia-Location", want)
}

func assertJSONResponse(t *testing.T, resp *fiber.Response) {
	t.Helper()

	assertHeader(t, resp, "Content-Type", "application/json")
}

func assertHTMLResponse(t *testing.T, resp *fiber.Response) {
	t.Helper()

	assertHeader(t, resp, "Content-Type", "text/html")
}

func assertInertiaVary(t *testing.T, resp *fiber.Response) {
	t.Helper()

	assertHeader(t, resp, "Vary", "X-Inertia")
}

func assertInertiaNotVary(t *testing.T, resp *fiber.Response) {
	t.Helper()

	assertHeaderMissing(t, resp, "Vary")
}

func assertHandlerServed(t *testing.T, handlers ...fiber.Handler) fiber.Handler {
	t.Helper()

	called := false

	t.Cleanup(func() {
		if !called {
			t.Fatal("handler was not called")
		}
	})

	return func(c *fiber.Ctx) error {
		for _, handler := range handlers {
			_ = handler(c)
		}

		called = true
		return nil
	}
}

func tmpFile(t *testing.T, content string) *os.File {
	t.Helper()

	f, err := os.CreateTemp("", "fibernetia")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	closed := false

	if _, err = f.WriteString(content); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if err = f.Close(); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	closed = true

	t.Cleanup(func() {
		if !closed {
			if err = f.Close(); err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
		}

		if err = os.Remove(f.Name()); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
	})

	return f
}

type flashProviderMock struct {
	FlashMock              func(c *fiber.Ctx, key string, val any) error
	GetMock                func(c *fiber.Ctx, key string) (any, error)
	GetAllMock             func(c *fiber.Ctx) (map[string]any, error)
	ShouldClearHistoryMock func(c *fiber.Ctx) (bool, error)
	FlashClearHistoryMock  func(c *fiber.Ctx) error
	clearHistory           bool
	validationErrors       map[string]any
}

func (p *flashProviderMock) Flash(c *fiber.Ctx, key string, val any) error {
	if key == "errors" {
		if m, ok := val.(map[string]any); ok {
			p.validationErrors = m
		}
	}
	if p.FlashMock != nil {
		return p.FlashMock(c, key, val)
	}
	return nil
}

func (p *flashProviderMock) Get(c *fiber.Ctx, key string) (any, error) {
	if p.GetMock != nil {
		return p.GetMock(c, key)
	}
	return nil, nil
}

func (p *flashProviderMock) GetAll(c *fiber.Ctx) (map[string]any, error) {
	if p.GetAllMock != nil {
		return p.GetAllMock(c)
	}
	return nil, nil
}

func (p *flashProviderMock) FlashClearHistory(c *fiber.Ctx) error {
	p.clearHistory = true
	if p.FlashClearHistoryMock != nil {
		return p.FlashClearHistoryMock(c)
	}
	return nil
}

func (p *flashProviderMock) ShouldClearHistory(c *fiber.Ctx) (bool, error) {
	if p.ShouldClearHistoryMock != nil {
		return p.ShouldClearHistoryMock(c)
	}
	return p.clearHistory, nil
}
