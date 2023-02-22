package main

import (
	"context"
	"fmt"

	"github.com/MontFerret/ferret/pkg/drivers"
	"github.com/MontFerret/ferret/pkg/runtime/core"
	"github.com/MontFerret/ferret/pkg/runtime/values"
	"github.com/MontFerret/ferret/pkg/stdlib/html"
	"github.com/MontFerret/ferret/pkg/stdlib/strings"
)

// ELEMENTS finds Urls elements by a given CSS selector in a html/xml sitemap
// Returns an empty array if element not found.
// @param {HTMLPage | HTMLDocument | HTMLElement} node - Target html node.
// @param {String} Glob - Glob to match
// @param {String} Glob - Glob not to match
// @param {String} prefix - String prefix to append the urls to
// @param {String} selector - CSS selector. Default is <a>
// @return {HTMLElement[]} - An array of matched HTML elements.
func ExtractUrls(ctx context.Context, args ...core.Value) (core.Value, error) {

	err := core.ValidateArgs(args, 1, 5)

	if err != nil {
		return values.EmptyArray(), err
	}

	el, err := drivers.ToElement(args[0])

	if err != nil {
		return values.EmptyArray(), err
	}

	selector, _ := drivers.ToQuerySelector(values.NewString("a"))

	if len(args) > 4 {
		selector, err = drivers.ToQuerySelector(args[4])
		if err != nil {
			return values.EmptyArray(), err
		}
	}

	elements, err := html.Elements(ctx, el, selector)

	if err != nil {
		return values.EmptyArray(), err
	}

	arr := elements.(*values.Array)

	filtered := values.EmptyArray()

	arr.ForEach(func(value core.Value, idx int) bool {

		el, _ := drivers.ToElement(value)

		href, _ := el.GetAttribute(ctx, "href")

		if len(args) > 1 {
			if matches, _ := strings.Like(ctx, href, args[1]); !bool(matches.(values.Boolean)) {
				return true
			}
		}

		if len(args) > 2 {
			if matches, _ := strings.Like(ctx, href, args[2]); bool(matches.(values.Boolean)) {
				return true
			}
		}

		if len(args) > 3 {
			href = values.String(fmt.Sprintf("%s%s", args[3], href))
		}

		filtered.Push(href)
		return true
	})

	return filtered, nil

}
