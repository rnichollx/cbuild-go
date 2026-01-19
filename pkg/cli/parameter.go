package cli

import (
	"context"
	"errors"
	"net/url"
	"time"
)

type ParameterType int

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
	Key() any
	Default() any
	Type() ParameterType
	Description() string
	Required() bool
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
