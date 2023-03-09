package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/MontFerret/ferret/pkg/runtime/core"
	"github.com/MontFerret/ferret/pkg/runtime/values"
	"github.com/MontFerret/ferret/pkg/runtime/values/types"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/segmentio/ksuid"
	"github.com/segmentio/parquet-go"
)

type ParquetWriter struct {
	uploader *manager.Uploader
}

// New returns ParquetWriter
func NewParquetWriter(uploader *manager.Uploader) *ParquetWriter {

	parquetWriter := &ParquetWriter{

		uploader: uploader,
	}

	return parquetWriter
}

func (w *ParquetWriter) WriteParquetTrafilatura(ctx context.Context, args ...core.Value) (core.Value, error) {
	// it's just a helper function which helps to validate a number of passed args
	err := core.ValidateArgs(args, 3, 3)

	if err != nil {
		// it's recommended to return built-in None type, instead of nil
		return values.ZeroInt, err
	}

	// i is another helper functions allowing to do type validation
	err = core.ValidateType(args[0], types.String)

	if err != nil {
		return values.ZeroInt, err
	}
	err = core.ValidateType(args[1], types.String)

	if err != nil {
		return values.ZeroInt, err
	}
	// i is another helper functions allowing to do type validation
	err = core.ValidateType(args[2], types.Array)

	if err != nil {
		return values.ZeroInt, err
	}

	// cast to built-in string type
	bucket := args[0].(values.String).String()
	path := args[1].(values.String).String()

	items := args[2].(*values.Array)

	reader, writer := io.Pipe()
	defer reader.Close()
	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Minute)
	urls := make(chan core.Value, items.Length())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer writer.Close()
		defer wg.Done()
		pWriter := parquet.NewWriter(writer, parquet.DataPageVersion(1), parquet.PageBufferSize(20))

		items.ForEach(func(value core.Value, idx int) bool {

			obj := value.(*values.Object)

			row := TrafilaturaItem{
				ContentHTML:  obj.MustGet("contentHTML").String(),
				ContentText:  obj.MustGet("contentText").String(),
				CommentsHTML: obj.MustGet("commentsHTML").String(),
				CommentsText: obj.MustGet("commentsText").String(),

				Title:    obj.MustGet("title").String(),
				Author:   obj.MustGet("author").String(),
				PageDate: obj.MustGet("date").String(),

				Url:         obj.MustGet("url").String(),
				Hostname:    obj.MustGet("hostname").String(),
				Description: obj.MustGet("description").String(),
				Sitename:    obj.MustGet("sitename").String(),
			}
			if row.ContentText != "" {
				if err := pWriter.Write(row); err != nil {
					writer.CloseWithError(err)
					return false
				}
				urls <- obj.MustGet("url")
			}

			return true
		})

		if err := pWriter.Close(); err != nil {
			writer.CloseWithError(err)

		}

	}()

	if strings.HasPrefix(bucket, "s3://") {
		_, err = w.uploader.Upload(ctxTimeout, &s3.PutObjectInput{
			Bucket: aws.String(strings.TrimPrefix(bucket, "s3://")),
			Key: aws.String(fmt.Sprintf("%s/%s.parquet",
				path, ksuid.New().String())),
			Body: reader,
		})

		if err != nil {
			return values.ZeroInt, err
		}
	} else {
		path = fmt.Sprintf("%s/%s", bucket, path)
		err := os.MkdirAll(path, 0700)
		if err != nil {
			return values.ZeroInt, err
		}
		file, err := os.Create(fmt.Sprintf("%s/%s.parquet",
			path, ksuid.New().String()))
		if err != nil {
			return values.ZeroInt, err
		}
		defer file.Close()

		// copy from reader data into writer file
		if _, err := io.Copy(file, reader); err != nil {
			return values.ZeroInt, err
		}
	}

	go func() {
		wg.Wait()
		close(urls)
	}()
	results := values.EmptyArray()
	for s := range urls {
		results.Push(s)
	}

	return results, nil
}
