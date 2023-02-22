package main

import (
	"context"
	"fmt"
	"time"

	"github.com/MontFerret/ferret/pkg/runtime/core"
	"github.com/MontFerret/ferret/pkg/runtime/values"
	"github.com/MontFerret/ferret/pkg/runtime/values/types"
)

func datePrefix(ctx context.Context, args ...core.Value) (core.Value, error) {
	// it's just a helper function which helps to validate a number of passed args
	err := core.ValidateArgs(args, 1, 1)

	if err != nil {
		// it's recommended to return built-in None type, instead of nil
		return values.ZeroInt, err
	}

	// i is another helper functions allowing to do type validation
	err = core.ValidateType(args[0], types.String)

	if err != nil {
		return values.EmptyString, err
	}
	year, month, day := time.Now().UTC().Date()

	prefix := fmt.Sprintf("%s/year=%d/month=%d/day=%d",
		args[0], year, int(month), day)

	return values.NewString(prefix), nil

}
