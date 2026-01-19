package cli

import (
	"context"
	"fmt"
	"strings"
)

type Flag interface {
	Short() string
	Long() string

	GetParameter() Parameter

	Overwrite() OverwritePolicy
}

type Argument interface {
	Name() string
	GetParameter() Parameter
	Overwrite() OverwritePolicy
}

type OverwritePolicy int

const (
	OverwritePolicyDisallowed OverwritePolicy = iota
	OverwritePolicyAllowed
	OverwritePolicyAppend
)

// MixingPolicy defines the parsing policy for Parameters which are specified both as Flags and Arguments.
// This only applies when the same Parameter is available both as a flag and also as an argument, it has no influence
// for two different parameters are available as flags and arguments.
type MixingPolicy int

const (
	MixingPolicyNoDuplicates MixingPolicy = iota
	// MixingPolicySequencedArgs treats arguments strictly in their argument order.
	// Regardless of the order which flags appear, arguments will always be parsed in the order they are specified.
	// Given the Arguments A and B, with associated flags --a and --b then:
	// `abc --a def` this is treated as two attempts to set A.

	MixingPolicySequencedArgs

	// MixingPolicySequencedStrictArgs is the same MixingPolicySequencedArgs, except that any cross definitions
	// automatically trigger a parsing error, regardless of the mixing policy
	MixingPolicySequencedStrictArgs

	// MixingPolicyNoMixing requires all dual-method parameters be set using args or flags.
	// If at least one dual-method parameter is set using an argument, it is an error
	// to set any of them using a flag.
	MixingPolicyNoMixing

	// MixingPolicyFirstUnset parses flags, then arguments will fill argument parameters in order which are unset.
	// For example, given A, B then `abc --a def` sets A=def, B=abc.
	MixingPolicyFirstUnset
)

type ParseOptions struct {
	// AllowUnknownFlags causes the parser to not return an error if a flag is encountered which is not understood.
	AllowUnknownFlags bool

	// AllowUnknownArgs causes the parser to not return an error if an unknown argument is encountered.
	AllowUnknownArgs bool

	/// StrictOrderingFlags, if true, requires that flags for any subcommand appear after the subcommand,
	/// unless a parent command also accepts them.
	/// For example, if `foo` has the `bar` subcommand which accepts `-c`, then `foo bar -c` is always valid
	/// but `foo -c bar` is only valid when StrictOrderingFlags is false.
	StrictOrderingFlags bool

	/// StrictOrderingArgs, if true, requires that any arguments appear strictly after all flags have been defined
	/// For example, given a command `foo` that accepts arguments, `foo bar.txt -f` is only valid if
	/// StrictOrderingArgs is false
	StrictOrderingArgs bool

	Subcommands []SubcommandParseOptions
	Flags       []Flag
	Arguments   []Argument
}

type SubcommandParseOptions struct {
	/// The name of the subcommand, to be expected
	Name string
	/// If StopParsing is true, ParseFlagsAndArguments should immediately stop updating the context
	/// and leave any remaining values as UnparsedFlags or UnparsedArgs
	StopParsing bool
	/// If this option is set, the remaining flags are validated against the subcommand.
	/// Note that if StopParsing is set AND ParseOptions is non-null, then the arguments
	/// are to be validated and returned unparsed, and the context is not updated.
	ParseOptions *ParseOptions
}

type ParseResult struct {
	Ctx           context.Context
	UnparsedFlags []string
	UnparsedArgs  []string
}

// ParseInput contains the input to the parser. To handle special cases like `--` in subcommands,
// we need to have an UnparsedArgs field for things that must always be arguments.
type ParseInput struct {
	Ctx           context.Context
	Unparsed      []string
	UnparsedArgs  []string
	UnparsedFlags []string
}

func ParseFlagsAndArgs(input ParseInput) (ParseResult, error) {

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
