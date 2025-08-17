package fibernetia

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"os"
	"sync"

	"github.com/valyala/fasthttp"
)

// Inertia is the main Gonertia structure, which contains all the logic for being an Inertia adapter.
type Inertia struct {
	rootTemplate     *template.Template
	rootTemplateHTML string

	sharedPropsMu sync.RWMutex
	sharedProps   Props

	sharedTemplateDataMu sync.RWMutex
	sharedTemplateData   TemplateData

	sharedTemplateFuncsMu sync.RWMutex
	sharedTemplateFuncs   TemplateFuncs

	flash FlashProvider

	ssrURL        string
	ssrHTTPClient *fasthttp.Client

	containerID    string
	version        string
	encryptHistory bool
	jsonMarshaller JSONMarshaller
	logger         Logger
}

// New initializes and returns Inertia.
func New(rootTemplateHTML string, opts ...Option) (*Inertia, error) {
	if rootTemplateHTML == "" {
		return nil, fmt.Errorf("blank root template")
	}

	i := newInertia(func(i *Inertia) {
		i.rootTemplateHTML = rootTemplateHTML
	})

	for _, opt := range opts {
		if err := opt(i); err != nil {
			return nil, fmt.Errorf("initialize inertia: %w", err)
		}
	}

	return i, nil
}

// NewFromTemplate receives a *template.Template and then initializes Inertia.
func NewFromTemplate(rootTemplate *template.Template, opts ...Option) (*Inertia, error) {
	if rootTemplate == nil {
		return nil, fmt.Errorf("nil root template")
	}

	i := newInertia(func(i *Inertia) {
		i.rootTemplate = rootTemplate
	})

	for _, opt := range opts {
		if err := opt(i); err != nil {
			return nil, fmt.Errorf("initialize inertia: %w", err)
		}
	}

	return i, nil
}

// NewFromFileFS reads all bytes from the root template file and then initializes Inertia.
func NewFromFileFS(rootFS fs.FS, rootTemplatePath string, opts ...Option) (*Inertia, error) {
	bs, err := fs.ReadFile(rootFS, rootTemplatePath)
	if err != nil {
		return nil, fmt.Errorf("read file %q: %w", rootTemplatePath, err)
	}

	return NewFromBytes(bs, opts...)
}

// NewFromFile reads all bytes from the root template file and then initializes Inertia.
func NewFromFile(rootTemplatePath string, opts ...Option) (*Inertia, error) {
	bs, err := os.ReadFile(rootTemplatePath)
	if err != nil {
		return nil, fmt.Errorf("read file %q: %w", rootTemplatePath, err)
	}

	return NewFromBytes(bs, opts...)
}

// NewFromReader reads all bytes from the reader with root template html and then initializes Inertia.
func NewFromReader(rootTemplateReader io.Reader, opts ...Option) (*Inertia, error) {
	bs, err := io.ReadAll(rootTemplateReader)
	if err != nil {
		return nil, fmt.Errorf("read root template: %w", err)
	}
	if closer, ok := rootTemplateReader.(io.Closer); ok {
		_ = closer.Close()
	}

	return NewFromBytes(bs, opts...)
}

// NewFromBytes receives bytes with root template html and then initializes Inertia.
func NewFromBytes(rootTemplateBs []byte, opts ...Option) (*Inertia, error) {
	return New(string(rootTemplateBs), opts...)
}

func newInertia(f func(i *Inertia)) *Inertia {
	i := &Inertia{
		jsonMarshaller:      jsonDefaultMarshaller{},
		containerID:         "app",
		logger:              log.New(io.Discard, "", 0),
		sharedProps:         make(Props),
		sharedTemplateData:  make(TemplateData),
		sharedTemplateFuncs: make(TemplateFuncs),
		ssrHTTPClient:       &fasthttp.Client{},
	}
	f(i)
	return i
}

// Logger defines an interface for debug messages.
type Logger interface {
	Printf(format string, v ...any)
	Println(v ...any)
}

// FlashProvider defines an interface for a flash data provider.
type FlashProvider interface {
	FlashErrors(ctx context.Context, errors ValidationErrors) error
	GetErrors(ctx context.Context) (ValidationErrors, error)
	ShouldClearHistory(ctx context.Context) (bool, error)
	FlashClearHistory(ctx context.Context) error
}

// ShareProp adds passed prop to shared props.
func (i *Inertia) ShareProp(key string, val any) {
	i.sharedPropsMu.Lock()
	defer i.sharedPropsMu.Unlock()

	i.sharedProps[key] = val
}

// SharedProps returns shared props.
func (i *Inertia) SharedProps() Props {
	i.sharedPropsMu.RLock()
	defer i.sharedPropsMu.RUnlock()

	return i.sharedProps
}

// SharedProp returns the shared prop.
func (i *Inertia) SharedProp(key string) (any, bool) {
	i.sharedPropsMu.RLock()
	defer i.sharedPropsMu.RUnlock()

	val, ok := i.sharedProps[key]
	return val, ok
}

// ShareTemplateData adds passed data to shared template data.
func (i *Inertia) ShareTemplateData(key string, val any) {
	i.sharedTemplateDataMu.Lock()
	defer i.sharedTemplateDataMu.Unlock()

	i.sharedTemplateData[key] = val
}

// ShareTemplateFunc adds the passed value to the shared template func map.
// If no root template HTML string has been defined, it returns an error.
func (i *Inertia) ShareTemplateFunc(key string, val any) error {
	if i.rootTemplateHTML == "" {
		return fmt.Errorf("undefined root template html string")
	}

	i.sharedTemplateFuncsMu.Lock()
	defer i.sharedTemplateFuncsMu.Unlock()

	i.sharedTemplateFuncs[key] = val
	return nil
}
