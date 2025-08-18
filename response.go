package fibernetia

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"maps"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

// TemplateData are data that will be available in the root template.
type TemplateData map[string]any

// TemplateFuncs are functions that will be available in the root template.
type TemplateFuncs map[string]any

// Props are the data that will be transferred
// and will be available in the front-end component.
type Props map[string]any

// OptionalProp is a property that will evaluate when needed.
//
// https://inertiajs.com/partial-reloads
type OptionalProp struct {
	ignoresFirstLoad
	Value any
}

func (p OptionalProp) Prop() any {
	return p.Value
}

func Optional(value any) OptionalProp {
	return OptionalProp{Value: value}
}

var _ ignoreFirstLoad = OptionalProp{}

type ignoreFirstLoad interface {
	shouldIgnoreFirstLoad() bool
}

type ignoresFirstLoad struct{}

func (i ignoresFirstLoad) shouldIgnoreFirstLoad() bool { return true }

// DeferProp is a property that will evaluate after page load.
//
// https://v2.inertiajs.com/deferred-props
type DeferProp struct {
	ignoresFirstLoad
	mergesProps
	Value any
	Group string
}

func (p DeferProp) Prop() any {
	return p.Value
}

func (p DeferProp) Merge() DeferProp {
	p.merge = true
	return p
}

func Defer(value any, group ...string) DeferProp {
	return DeferProp{
		Value: value,
		Group: firstOr(group, "default"),
	}
}

var _ ignoreFirstLoad = DeferProp{}

var _ mergeable = DeferProp{}

// AlwaysProp is a property that will always be evaluated.
//
// https://inertiajs.com/partial-reloads
type AlwaysProp struct {
	Value any
}

func (p AlwaysProp) Prop() any {
	return p.Value
}

func Always(value any) AlwaysProp {
	return AlwaysProp{Value: value}
}

// MergeProps is a property whose items will be merged instead of overwritten.
//
// https://v2.inertiajs.com/merging-props
type MergeProps struct {
	mergesProps
	Value any
}

func (p MergeProps) Prop() any {
	return p.Value
}

func (p MergeProps) Merge() MergeProps {
	p.merge = true
	return p
}

func Merge(value any) MergeProps {
	return MergeProps{
		Value:       value,
		mergesProps: mergesProps{merge: true},
	}
}

var _ mergeable = MergeProps{}

type mergeable interface {
	shouldMerge() bool
}

type mergesProps struct {
	merge bool
}

func (p mergesProps) shouldMerge() bool {
	return p.merge
}

// Proper is an interface for custom type, which provides property, that will be resolved.
type Proper interface {
	Prop() any
}

// ProperWithContext is an interface for custom type, which provides property,
// that will be resolved with context passing.
type ProperWithContext interface {
	PropWithContext(_ context.Context) any
}

// TryProper is an interface for custom type, which provides property and error.
type TryProper interface {
	TryProp() (any, error)
}

// TryProperWithContext resolves with context.
type TryProperWithContext interface {
	TryPropWithContext(_ context.Context) (any, error)
}

// ValidationErrors are messages, that will be stored in the "errors" prop.
type ValidationErrors map[string]any

// Location creates redirect response.
//
// If request was made by Inertia - sets status to 409 and url will be in "X-Inertia-Location" header.
// Otherwise, it will do an HTTP redirect with specified status (default is 302 for GET, 303 for POST/PUT/PATCH).
func (i *Inertia) Location(ctx *fiber.Ctx, url string, status ...int) {
	i.flashContext(ctx)

	if IsInertiaRequest(ctx) {
		setInertiaLocationInResponse(ctx, url)
		deleteInertiaInResponse(ctx)
		deleteVaryInResponse(ctx)
		setResponseStatus(ctx, fiber.StatusConflict)
		return
	}

	redirectResponse(ctx, url, status...)
}

// Back creates plain redirect response to the previous url.
func (i *Inertia) Back(ctx *fiber.Ctx, status ...int) {
	i.Redirect(ctx, backURL(ctx), status...)
}

func backURL(ctx *fiber.Ctx) string {
	// At the moment, it based only on the "Referer" HTTP header.
	return refererFromRequest(ctx)
}

// Redirect creates plain redirect response.
func (i *Inertia) Redirect(ctx *fiber.Ctx, url string, status ...int) {
	i.flashContext(ctx)
	redirectResponse(ctx, url, status...)
}

func (i *Inertia) flashContext(ctx *fiber.Ctx) {
	i.flashValidationErrorsFromContext(ctx)
	i.flashClearHistoryFromContext(ctx)
}

func (i *Inertia) flashValidationErrorsFromContext(ctx *fiber.Ctx) {
	if i.flash == nil {
		return
	}

	validationErrors := ValidationErrorsFromContext(ctx)
	if len(validationErrors) == 0 {
		return
	}

	err := i.flash.Flash(ctx, "errors", validationErrors)
	if err != nil {
		i.logger.Printf("cannot flash validation errors: %s", err)
	}
}

func (i *Inertia) flashClearHistoryFromContext(ctx *fiber.Ctx) {
	if i.flash == nil {
		return
	}

	clearHistory := ClearHistoryFromContext(ctx)
	if !clearHistory {
		return
	}

	err := i.flash.FlashClearHistory(ctx)
	if err != nil {
		i.logger.Printf("cannot flash clear history: %s", err)
	}
}

// Render returns response with Inertia data.
//
// If request was made by Inertia - it will return data in JSON format.
// Otherwise, it will return HTML with root template.
//
// If SSR is enabled, pre-renders JavaScript and return HTML (https://inertiajs.com/server-side-rendering).
func (i *Inertia) Render(ctx *fiber.Ctx, component string, props ...Props) (err error) {
	p, err := i.buildPage(ctx, component, firstOr(props, nil))
	if err != nil {
		return fmt.Errorf("build page: %w", err)
	}

	if IsInertiaRequest(ctx) {
		if err = i.doInertiaResponse(ctx, p); err != nil {
			return fmt.Errorf("inertia response: %w", err)
		}
		return nil
	}

	if err = i.doHTMLResponse(ctx, p); err != nil {
		return fmt.Errorf("html response: %w", err)
	}

	return nil
}

type page struct {
	Component      string              `json:"component"`
	Props          Props               `json:"props"`
	URL            string              `json:"url"`
	Version        string              `json:"version"`
	EncryptHistory bool                `json:"encryptHistory"`
	ClearHistory   bool                `json:"clearHistory"`
	DeferredProps  map[string][]string `json:"deferredProps,omitempty"`
	MergeProps     []string            `json:"mergeProps,omitempty"`
}

func (i *Inertia) buildPage(ctx *fiber.Ctx, component string, props Props) (*page, error) {
	props = i.collectProps(ctx, props)

	deferredProps := i.resolveDeferredProps(ctx, component, props)
	mergeProps := resolveMergeProps(ctx, props)

	props, err := i.resolveProps(ctx, component, props)
	if err != nil {
		return nil, fmt.Errorf("resolve props: %w", err)
	}

	return &page{
		Component:      component,
		Props:          props,
		URL:            ctx.OriginalURL(),
		Version:        i.version,
		EncryptHistory: i.resolveEncryptHistory(ctx),
		ClearHistory:   ClearHistoryFromContext(ctx),
		DeferredProps:  deferredProps,
		MergeProps:     mergeProps,
	}, nil
}

func (i *Inertia) resolveDeferredProps(ctx *fiber.Ctx, component string, props Props) map[string][]string {
	if isPartial(ctx, component) {
		return nil
	}

	keysByGroups := make(map[string][]string)

	for key, val := range props {
		if dp, ok := val.(DeferProp); ok {
			keysByGroups[dp.Group] = append(keysByGroups[dp.Group], key)
		}
	}

	return keysByGroups
}

func (i *Inertia) collectProps(ctx *fiber.Ctx, props Props) Props {
	result := make(Props)

	// Add validation errors.
	{
		result["errors"] = AlwaysProp{ValidationErrorsFromContext(ctx)}
	}

	if i.flash != nil {
		flashMap, _ := i.flash.GetAll(ctx)
		if len(flashMap) > 0 {
			result["flash"] = flashMap
		}
	}
	// Add shared props.
	i.sharedPropsMu.RLock()
	maps.Copy(result, i.sharedProps)
	i.sharedPropsMu.RUnlock()

	// Add props from context.
	maps.Copy(result, PropsFromContext(ctx))

	// Add passed props.
	maps.Copy(result, props)

	return result
}

func resolveMergeProps(ctx *fiber.Ctx, props Props) []string {
	resetProps := setOf(resetFromRequest(ctx))

	var mergeProps []string
	for key, val := range props {
		if _, ok := resetProps[key]; ok {
			continue
		}

		if m, ok := val.(mergeable); ok && m.shouldMerge() {
			mergeProps = append(mergeProps, key)
		}
	}

	return mergeProps
}

//nolint:gocognit
func (i *Inertia) resolveProps(ctx *fiber.Ctx, component string, props Props) (Props, error) {
	// Partial reloads only work for visits made to the same page component.
	//
	// https://inertiajs.com/partial-reloads
	if isPartial(ctx, component) {
		// Only (include keys) and except (exclude keys) logic.
		only, except := getOnlyAndExcept(ctx)

		if len(only) > 0 {
			for key, val := range props {
				if _, ok := only[key]; ok {
					continue
				}
				if _, ok := val.(AlwaysProp); ok {
					continue
				}

				delete(props, key)
			}
		}
		for key := range except {
			if _, ok := props[key].(AlwaysProp); ok {
				continue
			}

			delete(props, key)
		}
	} else {
		// Props with ignoreFirstLoad should not be included.
		for key, val := range props {
			if ifl, ok := val.(ignoreFirstLoad); ok && ifl.shouldIgnoreFirstLoad() {
				delete(props, key)
			}
		}
	}

	// Resolve props concurrently.
	resolveCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	type result struct {
		key string
		val any
	}

	resultCh := make(chan result, len(props))
	errCh := make(chan error, 1)

	wg := sync.WaitGroup{}

	for key, val := range props {
		wg.Add(1)
		go func(ct context.Context, key string, val any) {
			defer wg.Done()

			resolvedVal, err := resolvePropVal(ct, val)
			if err != nil {
				select {
				case errCh <- fmt.Errorf("resolve prop %q: %w", key, err):
					cancel()
				default:
				}
				return
			}

			select {
			case resultCh <- result{key, resolvedVal}:
			case <-ct.Done():
			}
		}(resolveCtx, key, val)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	for res := range resultCh {
		props[res.key] = res.val
	}

	select {
	case err := <-errCh:
		return nil, err
	default:
	}

	return props, nil
}

func isPartial(ctx *fiber.Ctx, component string) bool {
	return partialComponentFromRequest(ctx) == component
}

func getOnlyAndExcept(ctx *fiber.Ctx) (only, except map[string]struct{}) {
	return setOf(onlyFromRequest(ctx)), setOf(exceptFromRequest(ctx))
}

func resolvePropVal(ct context.Context, val any) (_ any, err error) {
	switch proper := val.(type) {
	case Proper:
		val = proper.Prop()
	case TryProper:
		val, err = proper.TryProp()
		if err != nil {
			return nil, err
		}
	case ProperWithContext:
		val = proper.PropWithContext(ct)
	case TryProperWithContext:
		val, err = proper.TryPropWithContext(ct)
		if err != nil {
			return nil, err
		}
	}

	switch typed := val.(type) {
	case func() any:
		val = typed()
	case func(ctx context.Context) any:
		val = typed(ct)
	case func() (any, error):
		val, err = typed()
		if err != nil {
			return nil, fmt.Errorf("closure prop resolving: %w", err)
		}
	case func(ctx context.Context) (any, error):
		val, err = typed(ct)
		if err != nil {
			return nil, fmt.Errorf("closure prop resolving: %w", err)
		}
	}

	return val, nil
}

func (i *Inertia) resolveEncryptHistory(ctx *fiber.Ctx) bool {
	encryptHistory, ok := EncryptHistoryFromContext(ctx)
	if ok {
		return encryptHistory
	}

	return i.encryptHistory
}

func (i *Inertia) doInertiaResponse(ctx *fiber.Ctx, page *page) error {
	pageJSON, err := i.jsonMarshaller.Marshal(page)
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}

	setInertiaInResponse(ctx)
	setJSONResponse(ctx)
	setResponseStatus(ctx, fasthttp.StatusOK)

	if _, err = ctx.Write(pageJSON); err != nil {
		return fmt.Errorf("write bytes to response: %w", err)
	}

	return nil
}

func (i *Inertia) doHTMLResponse(ctx *fiber.Ctx, page *page) (err error) {
	// If root template is already created - we'll use it to save some time.
	if i.rootTemplate == nil {
		i.rootTemplate, err = i.buildRootTemplate()
		if err != nil {
			return fmt.Errorf("build root template: %w", err)
		}
	}

	templateData, err := i.buildTemplateData(ctx, page)
	if err != nil {
		return fmt.Errorf("build template data: %w", err)
	}

	setHTMLResponse(ctx)

	if err = i.rootTemplate.Execute(ctx, templateData); err != nil {
		return fmt.Errorf("execute root template: %w", err)
	}

	return nil
}

func (i *Inertia) buildRootTemplate() (*template.Template, error) {
	i.sharedTemplateFuncsMu.RLock()
	defer i.sharedTemplateFuncsMu.RUnlock()

	tmpl := template.New("").Funcs(template.FuncMap(i.sharedTemplateFuncs))
	return tmpl.Parse(i.rootTemplateHTML)
}

func (i *Inertia) buildTemplateData(ctx *fiber.Ctx, page *page) (TemplateData, error) {
	// Defaults.
	inertia, inertiaHead, err := i.buildInertiaHTML(page)
	if err != nil {
		return nil, fmt.Errorf("build inertia html: %w", err)
	}

	templateData := TemplateData{
		"inertia":     inertia,
		"inertiaHead": inertiaHead,
	}

	// Add the shared template data to the result.
	i.sharedTemplateDataMu.RLock()
	for key, val := range i.sharedTemplateData {
		templateData[key] = val
	}
	i.sharedTemplateDataMu.RUnlock()

	// Add template data from context to the result.
	for key, val := range TemplateDataFromContext(ctx) {
		templateData[key] = val
	}

	return templateData, nil
}

func (i *Inertia) buildInertiaHTML(page *page) (inertia, inertiaHead template.HTML, _ error) {
	pageJSON, err := i.jsonMarshaller.Marshal(page)
	if err != nil {
		return "", "", fmt.Errorf("json marshal page into json: %w", err)
	}

	if i.isSSREnabled() {
		inertia, inertiaHead, err = i.htmlContainerSSR(pageJSON)
		if err == nil {
			return inertia, inertiaHead, nil
		}

		i.logger.Printf("ssr rendering error: %s", err)
	}

	return i.htmlContainer(pageJSON)
}

func (i *Inertia) isSSREnabled() bool {
	return i.ssrURL != ""
}

// htmlContainerSSR will send request with json marshaled page payload to ssr render endpoint.
// That endpoint will return head and body html, which will be returned and then rendered.
func (i *Inertia) htmlContainerSSR(pageJSON []byte) (inertia, inertiaHead template.HTML, _ error) {
	url := i.prepareSSRURL()

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.Header.SetMethod(fasthttp.MethodPost)
	req.SetRequestURI(url)
	req.Header.SetContentType("application/json")
	req.SetBody(pageJSON)

	if err := i.ssrHTTPClient.Do(req, resp); err != nil {
		return "", "", fmt.Errorf("execute http request: %w", err)
	}

	if resp.StatusCode() >= fasthttp.StatusBadRequest {
		return "", "", fmt.Errorf("invalid response status code: %d", resp.StatusCode())
	}

	var ssr struct {
		Head []string `json:"head"`
		Body string   `json:"body"`
	}

	err := i.jsonMarshaller.Decode(bytes.NewReader(resp.Body()), &ssr)
	if err != nil {
		return "", "", fmt.Errorf("json decode ssr render response: %w", err)
	}

	inertia = template.HTML(ssr.Body)
	inertiaHead = template.HTML(strings.Join(ssr.Head, "\n"))

	return inertia, inertiaHead, nil
}

func (i *Inertia) prepareSSRURL() string {
	return strings.ReplaceAll(i.ssrURL, "/render", "") + "/render"
}

func (i *Inertia) htmlContainer(pageJSON []byte) (inertia, _ template.HTML, _ error) {
	var sb strings.Builder

	sb.WriteString(`<div id="`)
	sb.WriteString(i.containerID)
	sb.WriteString(`" data-page="`)
	template.HTMLEscape(&sb, pageJSON)
	sb.WriteString(`"></div>`)

	return template.HTML(sb.String()), "", nil
}
