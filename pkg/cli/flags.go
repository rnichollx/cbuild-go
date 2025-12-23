package cli

import (
	"context"
	"fmt"
	"strings"
)

type Flag interface {
	Short() string
	Long() string
	FromArgument() bool
	Key() FlagKey
	Description() string
	Valid(value string) error
	NeedsValue() bool
	Required() bool
}

type StringFlag struct {
	short string
	long  string
	key   FlagKey

	description  string
	fromArgument bool
	required     bool
}

func NewStringFlag(short, long string, key FlagKey, description string) *StringFlag {
	return &StringFlag{short: short, long: long, key: key, description: description}
}

func NewRequiredStringFlag(short, long string, key FlagKey, description string) *StringFlag {
	return &StringFlag{short: short, long: long, key: key, description: description, required: true}
}

func NewStringFlagFromArgument(short, long string, key FlagKey, description string) *StringFlag {
	return &StringFlag{short: short, long: long, key: key, description: description, fromArgument: true}
}

func (s *StringFlag) Short() string            { return s.short }
func (s *StringFlag) Long() string             { return s.long }
func (s *StringFlag) FromArgument() bool       { return s.fromArgument }
func (s *StringFlag) Key() FlagKey             { return s.key }
func (s *StringFlag) Description() string      { return s.description }
func (s *StringFlag) Valid(value string) error { return nil }
func (s *StringFlag) NeedsValue() bool         { return true }
func (s *StringFlag) Required() bool           { return s.required }

type BoolFlag struct {
	short        string
	long         string
	key          FlagKey
	description  string
	fromArgument bool
	required     bool
}

func NewBoolFlag(short, long string, key FlagKey, description string) *BoolFlag {
	return &BoolFlag{short: short, long: long, key: key, description: description}
}

func NewRequiredBoolFlag(short, long string, key FlagKey, description string) *BoolFlag {
	return &BoolFlag{short: short, long: long, key: key, description: description, required: true}
}

func NewBoolFlagFromArgument(short, long string, key FlagKey, description string) *BoolFlag {
	return &BoolFlag{short: short, long: long, key: key, description: description, fromArgument: true}
}

func (b *BoolFlag) Short() string            { return b.short }
func (b *BoolFlag) Long() string             { return b.long }
func (b *BoolFlag) FromArgument() bool       { return b.fromArgument }
func (b *BoolFlag) Key() FlagKey             { return b.key }
func (b *BoolFlag) Description() string      { return b.description }
func (b *BoolFlag) Valid(value string) error { return nil }
func (b *BoolFlag) NeedsValue() bool         { return false }
func (b *BoolFlag) Required() bool           { return b.required }

type FlagKey string

type ParseOptions struct {
	AllowUnknownFlags bool
	Flags             []Flag
}

func ParseFlags(ctx context.Context, opts ParseOptions, args []string) (context.Context, []string, error) {

	shortFlagMap := make(map[string]Flag)
	longFlagMap := make(map[string]Flag)

	var fromArgumentFlag Flag

	for _, flag := range opts.Flags {

		if flag.FromArgument() {
			if fromArgumentFlag != nil {
				return nil, nil, fmt.Errorf("only one flag can be a FromArgument flag")
			}
			fromArgumentFlag = flag
		}

		if flag.Short() != "" {
			if _, exists := shortFlagMap[flag.Short()]; exists {
				return nil, nil, fmt.Errorf("duplicate short flag: %s", flag.Short())
			}
			shortFlagMap[flag.Short()] = flag
		}

		if flag.Long() != "" {
			if _, exists := longFlagMap[flag.Long()]; exists {
				return nil, nil, fmt.Errorf("duplicate long flag: %s", flag.Long())
			}
			longFlagMap[flag.Long()] = flag
		}
	}

	seenFlags := make(map[FlagKey]bool)

	var nonFlagArgs []string
	for i := 0; i < len(args); i++ {

		arg := args[i]
		if arg == "--" {
			if opts.AllowUnknownFlags {
				nonFlagArgs = append(nonFlagArgs, args[i:]...)
			} else {
				nonFlagArgs = append(nonFlagArgs, args[i+1:]...)
			}
			break
		}

		if strings.HasPrefix(arg, "--") {
			name := arg[2:]
			flag, ok := longFlagMap[name]
			if !ok {
				if opts.AllowUnknownFlags {
					nonFlagArgs = append(nonFlagArgs, arg)
					continue
				}
				return nil, nil, fmt.Errorf("unknown flag: %s", arg)
			}

			if seenFlags[flag.Key()] {
				return nil, nil, fmt.Errorf("flag %s appeared multiple times", arg)
			}
			seenFlags[flag.Key()] = true

			if flag.NeedsValue() {
				if i+1 >= len(args) {
					return nil, nil, fmt.Errorf("missing value for flag: %s", arg)
				}
				val := args[i+1]
				if err := flag.Valid(val); err != nil {
					return nil, nil, fmt.Errorf("invalid value for flag %s: %w", arg, err)
				}
				ctx = context.WithValue(ctx, flag.Key(), val)
				i++
			} else {
				ctx = context.WithValue(ctx, flag.Key(), "true")
			}
		} else if strings.HasPrefix(arg, "-") && len(arg) > 1 {
			cluster := arg[1:]
			for j := 0; j < len(cluster); j++ {
				char := cluster[j]
				name := string(char)
				flag, ok := shortFlagMap[name]
				if !ok {
					if opts.AllowUnknownFlags {
						nonFlagArgs = append(nonFlagArgs, "-"+name)
						continue
					}
					return nil, nil, fmt.Errorf("unknown flag: -%s", name)
				}

				if seenFlags[flag.Key()] {
					return nil, nil, fmt.Errorf("flag -%s appeared multiple times", name)
				}
				seenFlags[flag.Key()] = true

				if flag.NeedsValue() {
					if j != len(cluster)-1 {
						return nil, nil, fmt.Errorf("flag -%s must be last in cluster as it requires a value", name)
					}
					if i+1 >= len(args) {
						return nil, nil, fmt.Errorf("missing value for flag: -%s", name)
					}
					val := args[i+1]
					if err := flag.Valid(val); err != nil {
						return nil, nil, fmt.Errorf("invalid value for flag -%s: %w", name, err)
					}
					ctx = context.WithValue(ctx, flag.Key(), val)
					i++
				} else {
					ctx = context.WithValue(ctx, flag.Key(), "true")
				}
			}
		} else {
			nonFlagArgs = append(nonFlagArgs, arg)
		}

	}

	if fromArgumentFlag != nil && len(nonFlagArgs) > 0 {
		if ctx.Value(fromArgumentFlag.Key()) == nil {
			val := nonFlagArgs[0]
			if err := fromArgumentFlag.Valid(val); err != nil {
				return nil, nil, fmt.Errorf("invalid value for FromArgument flag: %w", err)
			}
			ctx = context.WithValue(ctx, fromArgumentFlag.Key(), val)
			nonFlagArgs = nonFlagArgs[1:]
		}
	}

	for _, flag := range opts.Flags {
		if flag.Required() {
			if ctx.Value(flag.Key()) == nil {
				return nil, nil, fmt.Errorf("required flag not set: %s", flag.Key())
			}
		}
	}

	return ctx, nonFlagArgs, nil
}

func GetString(ctx context.Context, key FlagKey) string {
	val := ctx.Value(key)
	if val == nil {
		return ""
	}
	return val.(string)
}

func GetBool(ctx context.Context, key FlagKey) bool {
	val := ctx.Value(key)
	if val == nil {
		return false
	}
	return val.(string) == "true"
}
