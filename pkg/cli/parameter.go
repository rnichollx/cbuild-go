package cli

import (
	"context"
	"errors"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type ParameterType int

type ParameterKey string

const (
	ParameterTypeString ParameterType = iota
	ParameterTypeInt
	ParameterTypeBool
	ParameterTypePath
	ParameterTypeDatetime
	ParameterTypeDuration
	ParameterTypeURI
	ParameterTypeStringList
	ParameterTypeIntList
	ParameterTypeBoolList
	ParameterTypePathList
	ParameterTypeDatetimeList
	ParameterTypeDurationList
	ParameterTypeURIList
)

var InvalidParameterTypeError = errors.New("invalid parameter type")

type Parameter interface {
	Key() ParameterKey
	Default() any
	Type() ParameterType
	Description() string
}

func GetBool(ctx context.Context, parm Parameter) (*bool, error) {
	if parm.Type() == ParameterTypeBool {
		val := ctx.Value(parm.Key())
		if val == nil {
			d := parm.Default()
			if d == nil {
				return nil, nil
			}
			return d.(*bool), nil
		}
		b := val.(bool)
		return &b, nil
	}

	return nil, InvalidParameterTypeError
}

func SetBool(ctx context.Context, parm Parameter, val bool) (context.Context, error) {
	if parm.Type() == ParameterTypeBool {
		return context.WithValue(ctx, parm.Key(), val), nil
	}

	return ctx, InvalidParameterTypeError
}

func UnsetBool(ctx context.Context, parm Parameter) (context.Context, error) {
	if parm.Type() == ParameterTypeBool {
		return context.WithValue(ctx, parm.Key(), nil), nil
	}

	return ctx, InvalidParameterTypeError
}

func GetInt(ctx context.Context, parm Parameter) (*int64, error) {
	if parm.Type() == ParameterTypeInt {
		val := ctx.Value(parm.Key())
		if val == nil {
			d := parm.Default()
			if d == nil {
				return nil, nil
			}
			return d.(*int64), nil
		}
		i := val.(int64)
		return &i, nil
	}

	return nil, InvalidParameterTypeError
}

func SetInt(ctx context.Context, parm Parameter, val int64) (context.Context, error) {
	if parm.Type() == ParameterTypeInt {
		return context.WithValue(ctx, parm.Key(), val), nil
	}

	return ctx, InvalidParameterTypeError
}

func UnsetInt(ctx context.Context, parm Parameter) (context.Context, error) {
	if parm.Type() == ParameterTypeInt {
		return context.WithValue(ctx, parm.Key(), nil), nil
	}

	return ctx, InvalidParameterTypeError
}

func GetString(ctx context.Context, parm Parameter) (*string, error) {
	if parm.Type() == ParameterTypeString {
		val := ctx.Value(parm.Key())
		if val == nil {
			d := parm.Default()
			if d == nil {
				return nil, nil
			}
			return d.(*string), nil
		}
		s := val.(string)
		return &s, nil
	}

	return nil, InvalidParameterTypeError
}

func SetString(ctx context.Context, parm Parameter, val string) (context.Context, error) {
	if parm.Type() == ParameterTypeString {
		return context.WithValue(ctx, parm.Key(), val), nil
	}

	return ctx, InvalidParameterTypeError
}

func UnsetString(ctx context.Context, parm Parameter) (context.Context, error) {
	if parm.Type() == ParameterTypeString {
		return context.WithValue(ctx, parm.Key(), nil), nil
	}

	return ctx, InvalidParameterTypeError
}

func GetPath(ctx context.Context, parm Parameter) (*string, error) {
	if parm.Type() == ParameterTypePath {
		val := ctx.Value(parm.Key())
		if val == nil {
			d := parm.Default()
			if d == nil {
				return nil, nil
			}
			return d.(*string), nil
		}
		s := val.(string)
		return &s, nil
	}

	return nil, InvalidParameterTypeError
}

func SetPath(ctx context.Context, parm Parameter, val string) (context.Context, error) {
	if parm.Type() == ParameterTypePath {
		return context.WithValue(ctx, parm.Key(), val), nil
	}

	return ctx, InvalidParameterTypeError
}

func UnsetPath(ctx context.Context, parm Parameter) (context.Context, error) {
	if parm.Type() == ParameterTypePath {
		return context.WithValue(ctx, parm.Key(), nil), nil
	}

	return ctx, InvalidParameterTypeError
}

func GetDatetime(ctx context.Context, parm Parameter) (*time.Time, error) {
	if parm.Type() == ParameterTypeDatetime {
		val := ctx.Value(parm.Key())
		if val == nil {
			d := parm.Default()
			if d == nil {
				return nil, nil
			}
			return d.(*time.Time), nil
		}
		t := val.(time.Time)
		return &t, nil
	}

	return nil, InvalidParameterTypeError
}

func SetDatetime(ctx context.Context, parm Parameter, val time.Time) (context.Context, error) {
	if parm.Type() == ParameterTypeDatetime {
		return context.WithValue(ctx, parm.Key(), val), nil
	}

	return ctx, InvalidParameterTypeError
}

func UnsetDatetime(ctx context.Context, parm Parameter) (context.Context, error) {
	if parm.Type() == ParameterTypeDatetime {
		return context.WithValue(ctx, parm.Key(), nil), nil
	}

	return ctx, InvalidParameterTypeError
}

func GetDuration(ctx context.Context, parm Parameter) (*time.Duration, error) {
	if parm.Type() == ParameterTypeDuration {
		val := ctx.Value(parm.Key())
		if val == nil {
			d := parm.Default()
			if d == nil {
				return nil, nil
			}
			return d.(*time.Duration), nil
		}
		d := val.(time.Duration)
		return &d, nil
	}

	return nil, InvalidParameterTypeError
}

func SetDuration(ctx context.Context, parm Parameter, val time.Duration) (context.Context, error) {
	if parm.Type() == ParameterTypeDuration {
		return context.WithValue(ctx, parm.Key(), val), nil
	}

	return ctx, InvalidParameterTypeError
}

func UnsetDuration(ctx context.Context, parm Parameter) (context.Context, error) {
	if parm.Type() == ParameterTypeDuration {
		return context.WithValue(ctx, parm.Key(), nil), nil
	}

	return ctx, InvalidParameterTypeError
}

func GetURI(ctx context.Context, parm Parameter) (*url.URL, error) {
	if parm.Type() == ParameterTypeURI {
		val := ctx.Value(parm.Key())
		if val == nil {
			d := parm.Default()
			if d == nil {
				return nil, nil
			}
			return d.(*url.URL), nil
		}
		u := val.(url.URL)
		return &u, nil
	}

	return nil, InvalidParameterTypeError
}

func SetURI(ctx context.Context, parm Parameter, val url.URL) (context.Context, error) {
	if parm.Type() == ParameterTypeURI {
		return context.WithValue(ctx, parm.Key(), val), nil
	}

	return ctx, InvalidParameterTypeError
}

func UnsetURI(ctx context.Context, parm Parameter) (context.Context, error) {
	if parm.Type() == ParameterTypeURI {
		return context.WithValue(ctx, parm.Key(), nil), nil
	}

	return ctx, InvalidParameterTypeError
}

func GetBoolList(ctx context.Context, parm Parameter) (*[]bool, error) {
	if parm.Type() == ParameterTypeBoolList {
		val := ctx.Value(parm.Key())
		if val == nil {
			d := parm.Default()
			if d == nil {
				return nil, nil
			}
			return d.(*[]bool), nil
		}
		b := val.([]bool)
		return &b, nil
	}

	return nil, InvalidParameterTypeError
}

func SetBoolList(ctx context.Context, parm Parameter, val []bool) (context.Context, error) {
	if parm.Type() == ParameterTypeBoolList {
		return context.WithValue(ctx, parm.Key(), val), nil
	}

	return ctx, InvalidParameterTypeError
}

func UnsetBoolList(ctx context.Context, parm Parameter) (context.Context, error) {
	if parm.Type() == ParameterTypeBoolList {
		return context.WithValue(ctx, parm.Key(), nil), nil
	}

	return ctx, InvalidParameterTypeError
}

func GetIntList(ctx context.Context, parm Parameter) (*[]int64, error) {
	if parm.Type() == ParameterTypeIntList {
		val := ctx.Value(parm.Key())
		if val == nil {
			d := parm.Default()
			if d == nil {
				return nil, nil
			}
			return d.(*[]int64), nil
		}
		i := val.([]int64)
		return &i, nil
	}

	return nil, InvalidParameterTypeError
}

func SetIntList(ctx context.Context, parm Parameter, val []int64) (context.Context, error) {
	if parm.Type() == ParameterTypeIntList {
		return context.WithValue(ctx, parm.Key(), val), nil
	}

	return ctx, InvalidParameterTypeError
}

func UnsetIntList(ctx context.Context, parm Parameter) (context.Context, error) {
	if parm.Type() == ParameterTypeIntList {
		return context.WithValue(ctx, parm.Key(), nil), nil
	}

	return ctx, InvalidParameterTypeError
}

func GetStringList(ctx context.Context, parm Parameter) (*[]string, error) {
	if parm.Type() == ParameterTypeStringList {
		val := ctx.Value(parm.Key())
		if val == nil {
			d := parm.Default()
			if d == nil {
				return nil, nil
			}
			return d.(*[]string), nil
		}
		s := val.([]string)
		return &s, nil
	}

	return nil, InvalidParameterTypeError
}

func SetStringList(ctx context.Context, parm Parameter, val []string) (context.Context, error) {
	if parm.Type() == ParameterTypeStringList {
		return context.WithValue(ctx, parm.Key(), val), nil
	}

	return ctx, InvalidParameterTypeError
}

func UnsetStringList(ctx context.Context, parm Parameter) (context.Context, error) {
	if parm.Type() == ParameterTypeStringList {
		return context.WithValue(ctx, parm.Key(), nil), nil
	}

	return ctx, InvalidParameterTypeError
}

func GetPathList(ctx context.Context, parm Parameter) (*[]string, error) {
	if parm.Type() == ParameterTypePathList {
		val := ctx.Value(parm.Key())
		if val == nil {
			d := parm.Default()
			if d == nil {
				return nil, nil
			}
			return d.(*[]string), nil
		}
		s := val.([]string)
		return &s, nil
	}

	return nil, InvalidParameterTypeError
}

func SetPathList(ctx context.Context, parm Parameter, val []string) (context.Context, error) {
	if parm.Type() == ParameterTypePathList {
		return context.WithValue(ctx, parm.Key(), val), nil
	}

	return ctx, InvalidParameterTypeError
}

func UnsetPathList(ctx context.Context, parm Parameter) (context.Context, error) {
	if parm.Type() == ParameterTypePathList {
		return context.WithValue(ctx, parm.Key(), nil), nil
	}

	return ctx, InvalidParameterTypeError
}

func GetDatetimeList(ctx context.Context, parm Parameter) (*[]time.Time, error) {
	if parm.Type() == ParameterTypeDatetimeList {
		val := ctx.Value(parm.Key())
		if val == nil {
			d := parm.Default()
			if d == nil {
				return nil, nil
			}
			return d.(*[]time.Time), nil
		}
		t := val.([]time.Time)
		return &t, nil
	}

	return nil, InvalidParameterTypeError
}

func SetDatetimeList(ctx context.Context, parm Parameter, val []time.Time) (context.Context, error) {
	if parm.Type() == ParameterTypeDatetimeList {
		return context.WithValue(ctx, parm.Key(), val), nil
	}

	return ctx, InvalidParameterTypeError
}

func UnsetDatetimeList(ctx context.Context, parm Parameter) (context.Context, error) {
	if parm.Type() == ParameterTypeDatetimeList {
		return context.WithValue(ctx, parm.Key(), nil), nil
	}

	return ctx, InvalidParameterTypeError
}

func GetDurationList(ctx context.Context, parm Parameter) (*[]time.Duration, error) {
	if parm.Type() == ParameterTypeDurationList {
		val := ctx.Value(parm.Key())
		if val == nil {
			d := parm.Default()
			if d == nil {
				return nil, nil
			}
			return d.(*[]time.Duration), nil
		}
		d := val.([]time.Duration)
		return &d, nil
	}

	return nil, InvalidParameterTypeError
}

func SetDurationList(ctx context.Context, parm Parameter, val []time.Duration) (context.Context, error) {
	if parm.Type() == ParameterTypeDurationList {
		return context.WithValue(ctx, parm.Key(), val), nil
	}

	return ctx, InvalidParameterTypeError
}

func UnsetDurationList(ctx context.Context, parm Parameter) (context.Context, error) {
	if parm.Type() == ParameterTypeDurationList {
		return context.WithValue(ctx, parm.Key(), nil), nil
	}

	return ctx, InvalidParameterTypeError
}

func GetURIList(ctx context.Context, parm Parameter) (*[]url.URL, error) {
	if parm.Type() == ParameterTypeURIList {
		val := ctx.Value(parm.Key())
		if val == nil {
			d := parm.Default()
			if d == nil {
				return nil, nil
			}
			return d.(*[]url.URL), nil
		}
		u := val.([]url.URL)
		return &u, nil
	}

	return nil, InvalidParameterTypeError
}

func SetURIList(ctx context.Context, parm Parameter, val []url.URL) (context.Context, error) {
	if parm.Type() == ParameterTypeURIList {
		return context.WithValue(ctx, parm.Key(), val), nil
	}

	return ctx, InvalidParameterTypeError
}

func UnsetURIList(ctx context.Context, parm Parameter) (context.Context, error) {
	if parm.Type() == ParameterTypeURIList {
		return context.WithValue(ctx, parm.Key(), nil), nil
	}

	return ctx, InvalidParameterTypeError
}

func parseBool(str string) (bool, error) {
	str = strings.ToLower(str)
	switch str {
	case "true", "enabled", "y", "yes", "on":
		return true, nil
	case "false", "disabled", "n", "no", "off":
		return false, nil
	}
	return false, InvalidParameterTypeError
}

func SetParameter(ctx context.Context, parm Parameter, input string) (context.Context, error) {
	switch parm.Type() {
	case ParameterTypeString:
		return SetString(ctx, parm, input)
	case ParameterTypeInt:
		val, err := strconv.ParseInt(input, 10, 64)
		if err != nil {
			return ctx, err
		}
		return SetInt(ctx, parm, val)
	case ParameterTypeBool:
		val, err := parseBool(input)
		if err != nil {
			return ctx, err
		}
		return SetBool(ctx, parm, val)
	case ParameterTypePath:
		return SetPath(ctx, parm, input)
	case ParameterTypeDatetime:
		val, err := time.Parse(time.RFC3339, input)
		if err != nil {
			return ctx, err
		}
		return SetDatetime(ctx, parm, val)
	case ParameterTypeDuration:
		val, err := time.ParseDuration(input)
		if err != nil {
			return ctx, err
		}
		return SetDuration(ctx, parm, val)
	case ParameterTypeURI:
		val, err := url.Parse(input)
		if err != nil {
			return ctx, err
		}
		return SetURI(ctx, parm, *val)
	}

	return ctx, InvalidParameterTypeError
}

func SetParameterList(ctx context.Context, parm Parameter, inputs []string) (context.Context, error) {
	switch parm.Type() {
	case ParameterTypeStringList:
		return SetStringList(ctx, parm, inputs)
	case ParameterTypeIntList:
		var list []int64
		for _, input := range inputs {
			val, err := strconv.ParseInt(input, 10, 64)
			if err != nil {
				return ctx, err
			}
			list = append(list, val)
		}
		return SetIntList(ctx, parm, list)
	case ParameterTypeBoolList:
		var list []bool
		for _, input := range inputs {
			val, err := parseBool(input)
			if err != nil {
				return ctx, err
			}
			list = append(list, val)
		}
		return SetBoolList(ctx, parm, list)
	case ParameterTypePathList:
		return SetPathList(ctx, parm, inputs)
	case ParameterTypeDatetimeList:
		var list []time.Time
		for _, input := range inputs {
			val, err := time.Parse(time.RFC3339, input)
			if err != nil {
				return ctx, err
			}
			list = append(list, val)
		}
		return SetDatetimeList(ctx, parm, list)
	case ParameterTypeDurationList:
		var list []time.Duration
		for _, input := range inputs {
			val, err := time.ParseDuration(input)
			if err != nil {
				return ctx, err
			}
			list = append(list, val)
		}
		return SetDurationList(ctx, parm, list)
	case ParameterTypeURIList:
		var list []url.URL
		for _, input := range inputs {
			val, err := url.Parse(input)
			if err != nil {
				return ctx, err
			}
			list = append(list, *val)
		}
		return SetURIList(ctx, parm, list)
	}

	return ctx, InvalidParameterTypeError
}

func AppendParameter(ctx context.Context, parm Parameter, inputs []string) (context.Context, error) {
	switch parm.Type() {
	case ParameterTypeStringList:
		existing, err := GetStringList(ctx, parm)
		if err != nil {
			return ctx, err
		}
		var list []string
		if existing != nil {
			list = *existing
		}
		return SetStringList(ctx, parm, append(list, inputs...))
	case ParameterTypeIntList:
		existing, err := GetIntList(ctx, parm)
		if err != nil {
			return ctx, err
		}
		var list []int64
		if existing != nil {
			list = *existing
		}
		for _, input := range inputs {
			val, err := strconv.ParseInt(input, 10, 64)
			if err != nil {
				return ctx, err
			}
			list = append(list, val)
		}
		return SetIntList(ctx, parm, list)
	case ParameterTypeBoolList:
		existing, err := GetBoolList(ctx, parm)
		if err != nil {
			return ctx, err
		}
		var list []bool
		if existing != nil {
			list = *existing
		}
		for _, input := range inputs {
			val, err := parseBool(input)
			if err != nil {
				return ctx, err
			}
			list = append(list, val)
		}
		return SetBoolList(ctx, parm, list)
	case ParameterTypePathList:
		existing, err := GetPathList(ctx, parm)
		if err != nil {
			return ctx, err
		}
		var list []string
		if existing != nil {
			list = *existing
		}
		return SetPathList(ctx, parm, append(list, inputs...))
	case ParameterTypeDatetimeList:
		existing, err := GetDatetimeList(ctx, parm)
		if err != nil {
			return ctx, err
		}
		var list []time.Time
		if existing != nil {
			list = *existing
		}
		for _, input := range inputs {
			val, err := time.Parse(time.RFC3339, input)
			if err != nil {
				return ctx, err
			}
			list = append(list, val)
		}
		return SetDatetimeList(ctx, parm, list)
	case ParameterTypeDurationList:
		existing, err := GetDurationList(ctx, parm)
		if err != nil {
			return ctx, err
		}
		var list []time.Duration
		if existing != nil {
			list = *existing
		}
		for _, input := range inputs {
			val, err := time.ParseDuration(input)
			if err != nil {
				return ctx, err
			}
			list = append(list, val)
		}
		return SetDurationList(ctx, parm, list)
	case ParameterTypeURIList:
		existing, err := GetURIList(ctx, parm)
		if err != nil {
			return ctx, err
		}
		var list []url.URL
		if existing != nil {
			list = *existing
		}
		for _, input := range inputs {
			val, err := url.Parse(input)
			if err != nil {
				return ctx, err
			}
			list = append(list, *val)
		}
		return SetURIList(ctx, parm, list)
	}

	return ctx, InvalidParameterTypeError
}
