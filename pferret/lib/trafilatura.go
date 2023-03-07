package main

import (
	"context"
	"strconv"

	"strings"

	"github.com/MontFerret/ferret/pkg/drivers"
	"github.com/MontFerret/ferret/pkg/runtime/core"
	"github.com/MontFerret/ferret/pkg/runtime/values"
	"github.com/MontFerret/ferret/pkg/runtime/values/types"
	"github.com/go-shiori/dom"
	"github.com/markusmobius/go-trafilatura"
)

func TrafilaturaExtract(ctx context.Context, args ...core.Value) (core.Value, error) {
	// it's just a helper function which helps to validate a number of passed args
	err := core.ValidateArgs(args, 1, 2)

	if err != nil {
		// it's recommended to return built-in None type, instead of nil
		return values.None, err
	}

	opts := trafilatura.Options{
		Deduplicate:     true,
		ExcludeTables:   true,
		IncludeImages:   false,
		IncludeLinks:    true,
		ExcludeComments: true,
		Config: &trafilatura.Config{
			CacheSize:             4096,
			MinDuplicateCheckSize: 100,
			MaxDuplicateCount:     2,

			MinExtractedSize:        200,
			MinExtractedCommentSize: 10,
			MinOutputSize:           10,
			MinOutputCommentSize:    10,
		},
	}

	el, err := drivers.ToElement(args[0])

	if err != nil {
		return values.None, err
	}
	if len(args) == 2 {
		value := args[1]
		err := core.ValidateType(value, types.Object)
		if err != nil {
			return values.None, core.Error(
				err,
				"parse `params` argument",
			)
		}

		obj := value.(*values.Object)

		deduplicate, exists := obj.Get(values.NewString("deduplicate"))
		if exists {
			opts.Deduplicate = bool(deduplicate.(values.Boolean))
		}
		excludeTables, exists := obj.Get(values.NewString("excludeTables"))
		if exists {
			opts.ExcludeTables = bool(excludeTables.(values.Boolean))
		}
		includeImages, exists := obj.Get(values.NewString("includeImages"))
		if exists {
			opts.IncludeImages = bool(includeImages.(values.Boolean))
		}

		includeLinks, exists := obj.Get(values.NewString("includeLinks"))
		if exists {
			opts.IncludeLinks = bool(includeLinks.(values.Boolean))
		}

		excludeComments, exists := obj.Get(values.NewString("excludeComments"))
		if exists {
			opts.ExcludeComments = bool(excludeComments.(values.Boolean))
		}

	}

	reader := strings.NewReader(el.String())

	r, err := trafilatura.Extract(reader, opts)
	if err != nil {
		return values.None, err
	}

	obj := values.NewObject()
	obj.Set("contentHTML", values.NewString(dom.OuterHTML(r.ContentNode)))
	obj.Set("contentText", values.NewString(r.ContentText))

	if !opts.ExcludeComments {
		obj.Set("commentsHTML", values.NewString(r.CommentsText))
		obj.Set("commentsText", values.NewString(dom.OuterHTML(r.CommentsNode)))
	}

	obj.Set("title", values.NewString(r.Metadata.Title))
	obj.Set("author", values.NewString(r.Metadata.Author))
	obj.Set("url", values.NewString(r.Metadata.URL))
	obj.Set("hostname", values.NewString(r.Metadata.Hostname))
	obj.Set("description", values.NewString(r.Metadata.Description))
	obj.Set("sitename", values.NewString(r.Metadata.Sitename))
	obj.Set("date", values.NewString(strconv.FormatInt(r.Metadata.Date.UTC().UnixMilli(), 10)))
	obj.Set("license", values.NewString(r.Metadata.License))

	// obj.Set("metadata", values.NewObjectWith(
	// 	values.NewObjectProperty("title", values.NewString(r.Metadata.Title)),
	// 	values.NewObjectProperty("author", values.NewString(r.Metadata.Author)),
	// 	values.NewObjectProperty("url", values.NewString(r.Metadata.URL)),
	// 	values.NewObjectProperty("hostname", values.NewString(r.Metadata.Hostname)),
	// 	values.NewObjectProperty("description", values.NewString(r.Metadata.Description)),
	// 	values.NewObjectProperty("sitename", values.NewString(r.Metadata.Sitename)),
	// 	values.NewObjectProperty("date", values.NewString(strconv.FormatInt(r.Metadata.Date.UTC().UnixMilli(), 10))),

	// 	values.NewObjectProperty("license", values.NewString(r.Metadata.License))))
	return obj, nil

}

type TrafilaturaMetadata struct {
	Title       string `json:"title,omitempty" parquet:"title,dict,gzip"`
	Author      string `json:"author,omitempty" parquet:"author,dict,gzip"`
	Url         string `json:"url,omitempty" parquet:"url,dict,gzip"`
	Hostname    string `json:"hostname,omitempty" parquet:"hostname,dict,gzip"`
	Description string `json:"description,omitempty" parquet:"description,dict,gzip"`
	Sitename    string `json:"sitename,omitempty" parquet:"sitename,dict,gzip"`
	Date        string `json:"date,omitempty" parquet:"date,dict,gzip"`
}
type TrafilaturaItem struct {
	ContentHTML string `json:"content_html,omitempty" parquet:"content_html,dict,gzip"`
	ContentText string `json:"content_text,omitempty" parquet:"content_text,dict,gzip"`

	CommentsHTML string `json:"comments_html,omitempty" parquet:"comments_html,dict,gzip"`
	CommentsText string `json:"comments_text,omitempty" parquet:"comments_text,dict,gzip"`

	//Metadata TrafilaturaMetadata `json:"metadata,omitempty" parquet:"metadata,dict,gzip"`

	Title       string `json:"title,omitempty" parquet:"title,dict,gzip"`
	Author      string `json:"author,omitempty" parquet:"author,dict,gzip"`
	Url         string `json:"url,omitempty" parquet:"url,dict,gzip"`
	Hostname    string `json:"hostname,omitempty" parquet:"hostname,dict,gzip"`
	Description string `json:"description,omitempty" parquet:"description,dict,gzip"`
	Sitename    string `json:"sitename,omitempty" parquet:"sitename,dict,gzip"`
	Date        string `json:"date,omitempty" parquet:"date,dict,gzip"`
}
