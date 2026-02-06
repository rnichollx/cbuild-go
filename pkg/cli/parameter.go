package cli

import (
	"context"
	"errors"
	"fmt"
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

type Parameter struct {
	Key          ParameterKey
	DefaultValue any
	Type         ParameterType
	Description  string
}

func panicUnsetParameter(parm Parameter) {
	panic(fmt.Errorf("parameter %q is not set and has no default", parm.Key))
}

func GetBool(ctx context.Context, parm Parameter) bool {
	v := GetOptionalBool(ctx, parm)
	if v == nil {
		panicUnsetParameter(parm)
	}
	return *v
}

func GetInt(ctx context.Context, parm Parameter) int64 {
	v := GetOptionalInt(ctx, parm)
	if v == nil {
		panicUnsetParameter(parm)
	}
	return *v
}

func GetString(ctx context.Context, parm Parameter) string {
	v := GetOptionalString(ctx, parm)
	if v == nil {
		panicUnsetParameter(parm)
	}
	return *v
}

func GetPath(ctx context.Context, parm Parameter) string {
	v := GetOptionalPath(ctx, parm)
	if v == nil {
		panicUnsetParameter(parm)
	}
	return *v
}

func GetDatetime(ctx context.Context, parm Parameter) time.Time {
	v := GetOptionalDatetime(ctx, parm)
	if v == nil {
		panicUnsetParameter(parm)
	}
	return *v
}

func GetDuration(ctx context.Context, parm Parameter) time.Duration {
	v := GetOptionalDuration(ctx, parm)
	if v == nil {
		panicUnsetParameter(parm)
	}
	return *v
}

func GetURI(ctx context.Context, parm Parameter) url.URL {
	v := GetOptionalURI(ctx, parm)
	if v == nil {
		panicUnsetParameter(parm)
	}
	return *v
}

func GetBoolList(ctx context.Context, parm Parameter) []bool {
	v := GetOptionalBoolList(ctx, parm)
	if v == nil {
		panicUnsetParameter(parm)
	}
	return *v
}

func GetIntList(ctx context.Context, parm Parameter) []int64 {
	v := GetOptionalIntList(ctx, parm)
	if v == nil {
		panicUnsetParameter(parm)
	}
	return *v
}

func GetStringList(ctx context.Context, parm Parameter) []string {
	v := GetOptionalStringList(ctx, parm)
	if v == nil {
		panicUnsetParameter(parm)
	}
	return *v
}

func GetPathList(ctx context.Context, parm Parameter) []string {
	v := GetOptionalPathList(ctx, parm)
	if v == nil {
		panicUnsetParameter(parm)
	}
	return *v
}

func GetDatetimeList(ctx context.Context, parm Parameter) []time.Time {
	v := GetOptionalDatetimeList(ctx, parm)
	if v == nil {
		panicUnsetParameter(parm)
	}
	return *v
}

func GetDurationList(ctx context.Context, parm Parameter) []time.Duration {
	v := GetOptionalDurationList(ctx, parm)
	if v == nil {
		panicUnsetParameter(parm)
	}
	return *v
}

func GetURIList(ctx context.Context, parm Parameter) []url.URL {
	v := GetOptionalURIList(ctx, parm)
	if v == nil {
		panicUnsetParameter(parm)
	}
	return *v
}

func GetOptionalBool(ctx context.Context, parm Parameter) *bool {
	if parm.Type != ParameterTypeBool {
		panic(InvalidParameterTypeError)
	}
	val := ctx.Value(parm.Key)
	if val == nil {
		d := parm.DefaultValue
		if d == nil {
			return nil
		}
		return d.(*bool)
	}
	b := val.(bool)
	return &b
}

func GetBoolOr(ctx context.Context, parm Parameter, orValue *bool) *bool {
	v := GetOptionalBool(ctx, parm)
	if v != nil {
		return v
	}
	return orValue
}

func SetBool(ctx context.Context, parm Parameter, val bool) (context.Context, error) {
	if parm.Type != ParameterTypeBool {
		panic(InvalidParameterTypeError)
	}
	return context.WithValue(ctx, parm.Key, val), nil
}

func UnsetBool(ctx context.Context, parm Parameter) (context.Context, error) {
	if parm.Type != ParameterTypeBool {
		panic(InvalidParameterTypeError)
	}
	return context.WithValue(ctx, parm.Key, nil), nil
}

func GetOptionalInt(ctx context.Context, parm Parameter) *int64 {
	if parm.Type != ParameterTypeInt {
		panic(InvalidParameterTypeError)
	}
	val := ctx.Value(parm.Key)
	if val == nil {
		d := parm.DefaultValue
		if d == nil {
			return nil
		}
		return d.(*int64)
	}
	i := val.(int64)
	return &i
}

func GetIntOr(ctx context.Context, parm Parameter, orValue *int64) *int64 {
	v := GetOptionalInt(ctx, parm)
	if v != nil {
		return v
	}
	return orValue
}

func SetInt(ctx context.Context, parm Parameter, val int64) (context.Context, error) {
	if parm.Type != ParameterTypeInt {
		panic(InvalidParameterTypeError)
	}
	return context.WithValue(ctx, parm.Key, val), nil
}

func UnsetInt(ctx context.Context, parm Parameter) (context.Context, error) {
	if parm.Type != ParameterTypeInt {
		panic(InvalidParameterTypeError)
	}
	return context.WithValue(ctx, parm.Key, nil), nil
}

func GetOptionalString(ctx context.Context, parm Parameter) *string {
	if parm.Type != ParameterTypeString {
		panic(InvalidParameterTypeError)
	}
	val := ctx.Value(parm.Key)
	if val == nil {
		d := parm.DefaultValue
		if d == nil {
			return nil
		}
		return d.(*string)
	}
	s := val.(string)
	return &s
}

func GetStringOr(ctx context.Context, parm Parameter, orValue *string) *string {
	v := GetOptionalString(ctx, parm)
	if v != nil {
		return v
	}
	return orValue
}

func SetString(ctx context.Context, parm Parameter, val string) (context.Context, error) {
	if parm.Type != ParameterTypeString {
		panic(InvalidParameterTypeError)
	}
	return context.WithValue(ctx, parm.Key, val), nil
}

func UnsetString(ctx context.Context, parm Parameter) (context.Context, error) {
	if parm.Type != ParameterTypeString {
		panic(InvalidParameterTypeError)
	}
	return context.WithValue(ctx, parm.Key, nil), nil
}

func GetOptionalPath(ctx context.Context, parm Parameter) *string {
	if parm.Type != ParameterTypePath {
		panic(InvalidParameterTypeError)
	}
	val := ctx.Value(parm.Key)
	if val == nil {
		d := parm.DefaultValue
		if d == nil {
			return nil
		}
		return d.(*string)
	}
	s := val.(string)
	return &s
}

func GetPathOr(ctx context.Context, parm Parameter, orValue *string) *string {
	v := GetOptionalPath(ctx, parm)
	if v != nil {
		return v
	}
	return orValue
}

func SetPath(ctx context.Context, parm Parameter, val string) (context.Context, error) {
	if parm.Type != ParameterTypePath {
		panic(InvalidParameterTypeError)
	}
	return context.WithValue(ctx, parm.Key, val), nil
}

func UnsetPath(ctx context.Context, parm Parameter) (context.Context, error) {
	if parm.Type != ParameterTypePath {
		panic(InvalidParameterTypeError)
	}
	return context.WithValue(ctx, parm.Key, nil), nil
}

func GetOptionalDatetime(ctx context.Context, parm Parameter) *time.Time {
	if parm.Type != ParameterTypeDatetime {
		panic(InvalidParameterTypeError)
	}
	val := ctx.Value(parm.Key)
	if val == nil {
		d := parm.DefaultValue
		if d == nil {
			return nil
		}
		return d.(*time.Time)
	}
	t := val.(time.Time)
	return &t
}

func GetDatetimeOr(ctx context.Context, parm Parameter, orValue *time.Time) *time.Time {
	v := GetOptionalDatetime(ctx, parm)
	if v != nil {
		return v
	}
	return orValue
}

func SetDatetime(ctx context.Context, parm Parameter, val time.Time) (context.Context, error) {
	if parm.Type != ParameterTypeDatetime {
		panic(InvalidParameterTypeError)
	}
	return context.WithValue(ctx, parm.Key, val), nil
}

func UnsetDatetime(ctx context.Context, parm Parameter) (context.Context, error) {
	if parm.Type != ParameterTypeDatetime {
		panic(InvalidParameterTypeError)
	}
	return context.WithValue(ctx, parm.Key, nil), nil
}

func GetOptionalDuration(ctx context.Context, parm Parameter) *time.Duration {
	if parm.Type != ParameterTypeDuration {
		panic(InvalidParameterTypeError)
	}
	val := ctx.Value(parm.Key)
	if val == nil {
		d := parm.DefaultValue
		if d == nil {
			return nil
		}
		return d.(*time.Duration)
	}
	d := val.(time.Duration)
	return &d
}

func GetDurationOr(ctx context.Context, parm Parameter, orValue *time.Duration) *time.Duration {
	v := GetOptionalDuration(ctx, parm)
	if v != nil {
		return v
	}
	return orValue
}

func SetDuration(ctx context.Context, parm Parameter, val time.Duration) (context.Context, error) {
	if parm.Type != ParameterTypeDuration {
		panic(InvalidParameterTypeError)
	}
	return context.WithValue(ctx, parm.Key, val), nil
}

func UnsetDuration(ctx context.Context, parm Parameter) (context.Context, error) {
	if parm.Type != ParameterTypeDuration {
		panic(InvalidParameterTypeError)
	}
	return context.WithValue(ctx, parm.Key, nil), nil
}

func GetOptionalURI(ctx context.Context, parm Parameter) *url.URL {
	if parm.Type != ParameterTypeURI {
		panic(InvalidParameterTypeError)
	}
	val := ctx.Value(parm.Key)
	if val == nil {
		d := parm.DefaultValue
		if d == nil {
			return nil
		}
		return d.(*url.URL)
	}
	u := val.(url.URL)
	return &u
}

func GetURIOr(ctx context.Context, parm Parameter, orValue *url.URL) *url.URL {
	v := GetOptionalURI(ctx, parm)
	if v != nil {
		return v
	}
	return orValue
}

func SetURI(ctx context.Context, parm Parameter, val url.URL) (context.Context, error) {
	if parm.Type != ParameterTypeURI {
		panic(InvalidParameterTypeError)
	}
	return context.WithValue(ctx, parm.Key, val), nil
}

func UnsetURI(ctx context.Context, parm Parameter) (context.Context, error) {
	if parm.Type != ParameterTypeURI {
		panic(InvalidParameterTypeError)
	}
	return context.WithValue(ctx, parm.Key, nil), nil
}

func GetOptionalBoolList(ctx context.Context, parm Parameter) *[]bool {
	if parm.Type != ParameterTypeBoolList {
		panic(InvalidParameterTypeError)
	}
	val := ctx.Value(parm.Key)
	if val == nil {
		d := parm.DefaultValue
		if d == nil {
			return nil
		}
		return d.(*[]bool)
	}
	b := val.([]bool)
	return &b
}

func GetBoolListOr(ctx context.Context, parm Parameter, orValue *[]bool) *[]bool {
	v := GetOptionalBoolList(ctx, parm)
	if v != nil {
		return v
	}
	return orValue
}

func SetBoolList(ctx context.Context, parm Parameter, val []bool) (context.Context, error) {
	if parm.Type == ParameterTypeBoolList {
		return context.WithValue(ctx, parm.Key, val), nil
	}

	return ctx, InvalidParameterTypeError
}

func UnsetBoolList(ctx context.Context, parm Parameter) (context.Context, error) {
	if parm.Type == ParameterTypeBoolList {
		return context.WithValue(ctx, parm.Key, nil), nil
	}

	return ctx, InvalidParameterTypeError
}

func GetOptionalIntList(ctx context.Context, parm Parameter) *[]int64 {
	if parm.Type != ParameterTypeIntList {
		panic(InvalidParameterTypeError)
	}
	val := ctx.Value(parm.Key)
	if val == nil {
		d := parm.DefaultValue
		if d == nil {
			return nil
		}
		return d.(*[]int64)
	}
	i := val.([]int64)
	return &i
}

func GetIntListOr(ctx context.Context, parm Parameter, orValue *[]int64) *[]int64 {
	v := GetOptionalIntList(ctx, parm)
	if v != nil {
		return v
	}
	return orValue
}

func SetIntList(ctx context.Context, parm Parameter, val []int64) (context.Context, error) {
	if parm.Type == ParameterTypeIntList {
		return context.WithValue(ctx, parm.Key, val), nil
	}

	return ctx, InvalidParameterTypeError
}

func UnsetIntList(ctx context.Context, parm Parameter) (context.Context, error) {
	if parm.Type == ParameterTypeIntList {
		return context.WithValue(ctx, parm.Key, nil), nil
	}

	return ctx, InvalidParameterTypeError
}

func GetOptionalStringList(ctx context.Context, parm Parameter) *[]string {
	if parm.Type != ParameterTypeStringList {
		panic(InvalidParameterTypeError)
	}
	val := ctx.Value(parm.Key)
	if val == nil {
		d := parm.DefaultValue
		if d == nil {
			return nil
		}
		return d.(*[]string)
	}
	s := val.([]string)
	return &s
}

func GetStringListOr(ctx context.Context, parm Parameter, orValue *[]string) *[]string {
	v := GetOptionalStringList(ctx, parm)
	if v != nil {
		return v
	}
	return orValue
}

func SetStringList(ctx context.Context, parm Parameter, val []string) (context.Context, error) {
	if parm.Type == ParameterTypeStringList {
		return context.WithValue(ctx, parm.Key, val), nil
	}

	return ctx, InvalidParameterTypeError
}

func UnsetStringList(ctx context.Context, parm Parameter) (context.Context, error) {
	if parm.Type == ParameterTypeStringList {
		return context.WithValue(ctx, parm.Key, nil), nil
	}

	return ctx, InvalidParameterTypeError
}

func GetOptionalPathList(ctx context.Context, parm Parameter) *[]string {
	if parm.Type != ParameterTypePathList {
		panic(InvalidParameterTypeError)
	}
	val := ctx.Value(parm.Key)
	if val == nil {
		d := parm.DefaultValue
		if d == nil {
			return nil
		}
		return d.(*[]string)
	}
	s := val.([]string)
	return &s
}

func GetPathListOr(ctx context.Context, parm Parameter, orValue *[]string) *[]string {
	v := GetOptionalPathList(ctx, parm)
	if v != nil {
		return v
	}
	return orValue
}

func SetPathList(ctx context.Context, parm Parameter, val []string) (context.Context, error) {
	if parm.Type == ParameterTypePathList {
		return context.WithValue(ctx, parm.Key, val), nil
	}

	return ctx, InvalidParameterTypeError
}

func UnsetPathList(ctx context.Context, parm Parameter) (context.Context, error) {
	if parm.Type == ParameterTypePathList {
		return context.WithValue(ctx, parm.Key, nil), nil
	}

	return ctx, InvalidParameterTypeError
}

func GetOptionalDatetimeList(ctx context.Context, parm Parameter) *[]time.Time {
	if parm.Type != ParameterTypeDatetimeList {
		panic(InvalidParameterTypeError)
	}
	val := ctx.Value(parm.Key)
	if val == nil {
		d := parm.DefaultValue
		if d == nil {
			return nil
		}
		return d.(*[]time.Time)
	}
	t := val.([]time.Time)
	return &t
}

func GetDatetimeListOr(ctx context.Context, parm Parameter, orValue *[]time.Time) *[]time.Time {
	v := GetOptionalDatetimeList(ctx, parm)
	if v != nil {
		return v
	}
	return orValue
}

func SetDatetimeList(ctx context.Context, parm Parameter, val []time.Time) (context.Context, error) {
	if parm.Type == ParameterTypeDatetimeList {
		return context.WithValue(ctx, parm.Key, val), nil
	}

	return ctx, InvalidParameterTypeError
}

func UnsetDatetimeList(ctx context.Context, parm Parameter) (context.Context, error) {
	if parm.Type == ParameterTypeDatetimeList {
		return context.WithValue(ctx, parm.Key, nil), nil
	}

	return ctx, InvalidParameterTypeError
}

func GetOptionalDurationList(ctx context.Context, parm Parameter) *[]time.Duration {
	if parm.Type != ParameterTypeDurationList {
		panic(InvalidParameterTypeError)
	}
	val := ctx.Value(parm.Key)
	if val == nil {
		d := parm.DefaultValue
		if d == nil {
			return nil
		}
		return d.(*[]time.Duration)
	}
	d := val.([]time.Duration)
	return &d
}

func GetDurationListOr(ctx context.Context, parm Parameter, orValue *[]time.Duration) *[]time.Duration {
	v := GetOptionalDurationList(ctx, parm)
	if v != nil {
		return v
	}
	return orValue
}

func SetDurationList(ctx context.Context, parm Parameter, val []time.Duration) (context.Context, error) {
	if parm.Type == ParameterTypeDurationList {
		return context.WithValue(ctx, parm.Key, val), nil
	}

	return ctx, InvalidParameterTypeError
}

func UnsetDurationList(ctx context.Context, parm Parameter) (context.Context, error) {
	if parm.Type == ParameterTypeDurationList {
		return context.WithValue(ctx, parm.Key, nil), nil
	}

	return ctx, InvalidParameterTypeError
}

func GetOptionalURIList(ctx context.Context, parm Parameter) *[]url.URL {
	if parm.Type != ParameterTypeURIList {
		panic(InvalidParameterTypeError)
	}
	val := ctx.Value(parm.Key)
	if val == nil {
		d := parm.DefaultValue
		if d == nil {
			return nil
		}
		return d.(*[]url.URL)
	}
	u := val.([]url.URL)
	return &u
}

func GetURIListOr(ctx context.Context, parm Parameter, orValue *[]url.URL) *[]url.URL {
	v := GetOptionalURIList(ctx, parm)
	if v != nil {
		return v
	}
	return orValue
}

func SetURIList(ctx context.Context, parm Parameter, val []url.URL) (context.Context, error) {
	if parm.Type == ParameterTypeURIList {
		return context.WithValue(ctx, parm.Key, val), nil
	}

	return ctx, InvalidParameterTypeError
}

func UnsetURIList(ctx context.Context, parm Parameter) (context.Context, error) {
	if parm.Type == ParameterTypeURIList {
		return context.WithValue(ctx, parm.Key, nil), nil
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
	switch parm.Type {
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
	switch parm.Type {
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
	switch parm.Type {
	case ParameterTypeStringList:
		existing := GetOptionalStringList(ctx, parm)
		var list []string
		if existing != nil {
			list = *existing
		}
		return SetStringList(ctx, parm, append(list, inputs...))
	case ParameterTypeIntList:
		existing := GetOptionalIntList(ctx, parm)
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
		existing := GetOptionalBoolList(ctx, parm)
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
		existing := GetOptionalPathList(ctx, parm)
		var list []string
		if existing != nil {
			list = *existing
		}
		return SetPathList(ctx, parm, append(list, inputs...))
	case ParameterTypeDatetimeList:
		existing := GetOptionalDatetimeList(ctx, parm)
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
		existing := GetOptionalDurationList(ctx, parm)
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
		existing := GetOptionalURIList(ctx, parm)
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
