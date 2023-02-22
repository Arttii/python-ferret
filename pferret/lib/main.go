package main

/*
struct COptions {
	char* cdp;
	char* proxy;
	char* user_agent;
	char* params;
};

struct CResult {
    char* data;
    char* err;
};
*/
import "C"

import (
	"context"
	"encoding/json"
	"io/ioutil"

	"github.com/MontFerret/ferret"
	"github.com/MontFerret/ferret/pkg/drivers"
	"github.com/MontFerret/ferret/pkg/drivers/cdp"
	"github.com/MontFerret/ferret/pkg/drivers/http"
	"github.com/MontFerret/ferret/pkg/runtime"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Options struct {
	Cdp       string
	Proxy     string
	UserAgent string
	Params    map[string]interface{}
}

type key string

//export Execute
func Execute(queryC *C.char, optsC C.struct_COptions) C.struct_CResult {
	query := C.GoString(queryC)

	opts := Options{}
	cdpS := C.GoString(optsC.cdp)

	if len(cdpS) != 0 {
		opts.Cdp = cdpS
	}

	proxyS := C.GoString(optsC.proxy)

	if len(proxyS) != 0 {
		opts.Proxy = proxyS
	}

	userAgentS := C.GoString(optsC.user_agent)

	if len(userAgentS) != 0 {
		opts.UserAgent = userAgentS
	}

	paramsS := C.GoString(optsC.params)

	if len(paramsS) != 0 {
		var params interface{}

		err := json.Unmarshal([]byte(paramsS), &params)

		if err != nil {
			return C.struct_CResult{
				err: C.CString(err.Error()),
			}
		}

		dict, ok := params.(map[string]interface{})

		if !ok {
			return C.struct_CResult{
				err: C.CString("Invalid params format"),
			}
		}

		opts.Params = dict
	}

	f := ferret.New()
	err := f.Drivers().Register(http.NewDriver(
		http.WithProxy(opts.Proxy),
		http.WithUserAgent(opts.UserAgent),
	), drivers.AsDefault())

	if err != nil {
		return C.struct_CResult{
			err: C.CString(err.Error()),
		}
	}
	if opts.Cdp != "" {
		err = f.Drivers().Register(cdp.NewDriver(
			cdp.WithProxy(opts.Proxy),
			cdp.WithUserAgent(opts.UserAgent),
			cdp.WithAddress(opts.Cdp),
		))

		if err != nil {
			return C.struct_CResult{
				err: C.CString(err.Error()),
			}
		}

	}

	err = f.Functions().Namespace("TRAFILATURA").RegisterFunction("EXTRACT", TrafilaturaExtract)

	if err != nil {
		return C.struct_CResult{
			err: C.CString(err.Error()),
		}
	}

	err = f.Functions().Namespace("SITEMAP").RegisterFunction("EXTRACT", ExtractSitemapUrls)

	if err != nil {
		return C.struct_CResult{
			err: C.CString(err.Error()),
		}
	}

	err = f.Functions().Namespace("URLS").RegisterFunction("EXTRACT", ExtractUrls)

	if err != nil {
		return C.struct_CResult{
			err: C.CString(err.Error()),
		}
	}

	cfg, err := awsConfig.LoadDefaultConfig(context.TODO())
	if err != nil {
		return C.struct_CResult{
			err: C.CString(err.Error()),
		}
	}
	client := s3.NewFromConfig(cfg)

	uploader := manager.NewUploader(client, func(u *manager.Uploader) {
		u.PartSize = 10 * 1024 * 1024
		u.Concurrency = 5
	})

	parquetWriter := NewParquetWriter(uploader)

	err = f.Functions().Namespace("PARQUET").RegisterFunction("WRITE_TRAFILATURA", parquetWriter.WriteParquetTrafilatura)

	if err != nil {
		return C.struct_CResult{
			err: C.CString(err.Error()),
		}
	}

	err = f.Functions().Namespace("PARALLEL").RegisterFunction("FETCH_TRAFILATURA", FetchTrafilatura)

	if err != nil {
		return C.struct_CResult{
			err: C.CString(err.Error()),
		}
	}

	out, err := f.Exec(
		context.Background(),
		query,
		runtime.WithParams(opts.Params),
	)

	if err != nil {
		return C.struct_CResult{
			err: C.CString(err.Error()),
		}
	}

	return C.struct_CResult{
		data: C.CString(string(out)),
	}
}

func main() {

	example, _ := ioutil.ReadFile("/home/arti/workdir/python-ferret/example/example.fql")
	Execute(C.CString(string(example)), C.struct_COptions{})
}
