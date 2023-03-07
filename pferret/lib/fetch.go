package main

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/MontFerret/ferret/pkg/drivers"
	"github.com/MontFerret/ferret/pkg/runtime/core"
	"github.com/MontFerret/ferret/pkg/runtime/values"
	"github.com/MontFerret/ferret/pkg/runtime/values/types"
	"github.com/MontFerret/ferret/pkg/stdlib/html"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

// ELEMENTS parallel fetches docs from a list of urls
// Returns an empty array if element not found, none if error
// @param {String} Urls - Array of urls to fetch
// @param {Object} Driver params - Driver params
// @param {Int} Concurrency - Concurrency level
// @param {String} Wait duration - Wait duration between requests
// @return {HTMLElement[]} - An array of matched HTML elements.
func FetchTrafilatura(ctx context.Context, args ...core.Value) (core.Value, error) {

	err := core.ValidateArgs(args, 1, 3)

	if err != nil {
		return values.None, err
	}
	concunrrency := 5

	wait := time.Duration(0)

	if len(args) > 2 {
		concunrrency = int(args[2].(values.Int))
	}

	if len(args) > 3 {
		wait, _ = time.ParseDuration(args[3].(values.String).String())
		if err != nil {
			return values.None, err
		}

	}

	semaphore := semaphore.NewWeighted(int64(concunrrency))

	err = core.ValidateType(args[0], types.Array)

	if err != nil {
		return values.None, err
	}

	urls := args[0].(*values.Array)

	// g, _ := errgroup.WithContext(context.Background())

	g := new(errgroup.Group)

	docs := make(chan core.Value)
	results := values.EmptyArray()

	//logger := logging.FromContext(ctx)
	processedUrls := int32(urls.Length())

	urls.ForEach(func(url core.Value, idx int) bool {

		g.Go(func() error {

			defer func() {
				// Last one out closes shop
				if atomic.AddInt32(&processedUrls, -1) == 0 {
					close(docs)
				}
			}()

			// Acquire semaphore to limit concurrent download
			err := semaphore.Acquire(ctx, 1)
			if err != nil {
				return err
			}

			docValue := core.Value(nil)

			// Process URL
			if len(args) > 1 {
				docValue, err = html.Open(ctx, url)
			} else {
				docValue, err = html.Open(ctx, url, args[1])
			}

			semaphore.Release(1)
			if err != nil {

				return nil
			}

			doc, err := drivers.ToDocument(docValue)

			if err != nil {
				return nil

			}

			parsed, err := TrafilaturaExtract(ctx, doc)

			if err != nil {
				return nil
			}

			_ = doc

			// Add delay (to prevent too many request to target server)
			time.Sleep(wait)

			docs <- parsed
			return nil
		})

		return true
	})

	g.Go(func() error {
		for d := range docs {
			results.Push(d)
		}
		return nil
	})

	err = g.Wait()

	if err != nil {
		return values.None, err
	}

	return results, nil

}
