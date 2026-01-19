package cli

import (
	"context"
	"fmt"
	"strings"
)

type ParsingStyle int

const (
	// ParsingStyleRPNX is the default parsing style for the library
	ParsingStyleRPNX ParsingStyle = iota
	// ParsingStyleGNU is like ParsingStyleRPNX, but doesn't treat [ or ] specially
	ParsingStyleGNU
	// ParsingStylePOSIX treats single dash as arg+value, e.g. -abc is equivalent to `-a=bc` rather than `-a -b -c`
	ParsingStylePOSIX
	// ParsingStyleShort doesn't have long flags like `--foo`, instead
	// they are passed like `-foo`.
	ParsingStyleShort
	// like ParsingStyleShort, but uses `--%` for end of commands, and isn't case-sensitive
	ParsingStyleShortWindows
	// ParsingStyleWindows is like `/option:value` instead of `--option=value`
	ParsingStyleWindows
)

type ParseOptions struct {

	/// StrictOrderingArgs, if true, requires that any arguments appear strictly after all flags have been defined
	/// For example, given a command `foo` that accepts arguments, `foo bar.txt -f` is only valid if
	/// StrictOrderingArgs is false
	StrictOrderingArgs bool

	Style       ParsingStyle
	Subcommands map[string]SubcommandParseOptions
	Flags       []Flag
	Arguments   []Argument
}

type SubcommandParseOptions struct {
	/// If StopParsing is true, ParseFlagsAndArguments should immediately stop updating the context
	/// and leave any remaining values as UnparsedFlags or UnparsedArgs
	StopParsing bool
	/// If this option is set, the remaining flags are checked against the subcommand
	ParseOptions *ParseOptions
}

type ParseResult struct {
	Ctx context.Context

	// Any subcommands that were parsed have their name stored in Subcommands
	Subcommands []string

	// UnknownFlags contains any flags that were encountered during parsing but not understood.
	UnknownFlags []string
	// UnknownArgs contains any arguments that was encountered during parsing but not understood.
	UnknownArgs []string

	// If a subcommand was encountered, and it has StopParsing set, then the remainder of the input
	// is stored in Unparsed
	Unparsed []string
}

// ParseInput contains the input to the parser. To handle special cases like `--` in subcommands,
// we need to have an UnparsedArgs field for things that must always be arguments.
type ParseInput struct {
	Ctx      context.Context
	Unparsed []string
}

func ParseFlagsAndArgs(opts ParseOptions, input ParseInput) (ParseResult, error) {

	var result ParseResult
	ctx := input.Ctx
	shortFlagMap := make(map[string]Flag)
	longFlagMap := make(map[string]Flag)

	arguments := opts.Arguments

	var fromArgumentFlag Flag

	for _, flag := range opts.Flags {
		if flag.Short() != "" {
			if _, exists := shortFlagMap[flag.Short()]; exists {
				return result, fmt.Errorf("duplicate short flag: %s", flag.Short())
			}
			shortFlagMap[flag.Short()] = flag
		}

		if flag.Long() != "" {
			if _, exists := longFlagMap[flag.Long()]; exists {
				return result, fmt.Errorf("duplicate long flag: %s", flag.Long())
			}
			longFlagMap[flag.Long()] = flag
		}
	}

	seenParameters := make(map[ParameterKey]bool)

	var encounterArg func(string) error
	var encounterFlag func(string) error

	unparsedTokens := input.Unparsed

	var onlyArgs bool
	var argIndex int

	argTerminator := "--"
	if opts.Style == ParsingStyleShortWindows {
		argTerminator = "--%"
	}

	setValueChar := "="
	if opts.Style == ParsingStyleWindows {
		setValueChar = ":"
	}

	isLongFlag := func(arg string) bool {
		if onlyArgs {
			return false
		}
		if opts.Style == ParsingStyleWindows {
			return arg[:1] == "/"
		} else if arg[:2] == "--" && arg != argTerminator {
			return true
		} else if arg[:1] == "-" && opts.Style == ParsingStyleShortWindows || opts.Style == ParsingStyleShort {
			return true
		}
		return false
	}

	isShortFlag := func(arg string) bool {
		if len(arg) <= 1 {
			// stdin/empty arg
			return false
		}
		if onlyArgs {
			return false
		}
		// Some targets don't support short args per se, in these targets short args are just aliases for long ones
		if isLongFlag(arg) || opts.Style == ParsingStyleWindows || opts.Style == ParsingStyleShortWindows || opts.Style == ParsingStyleShort {
			return false
		}
		if arg[:1] == "-" {
			return true
		}
		return false
	}

	longFlagName := func(arg string) string {
		switch opts.Style {
		case ParsingStyleWindows, ParsingStyleShortWindows, ParsingStyleShort:
			return arg[:1]
		case ParsingStyleRPNX, ParsingStyleGNU, ParsingStylePOSIX:
			return arg[:2]
		default:
			panic(fmt.Sprintf("invalid style option: %v", opts.Style))
		}
	}

	longFlagValue := func(arg string) *string {
		if strings.Contains(arg, setValueChar) {
			str := strings.SplitN(arg, setValueChar, 2)[1]
			return &str
		}
		return nil
	}

	for i := 0; i < len(unparsedTokens); i++ {

		arg := unparsedTokens[i]
		if arg == argTerminator {
			onlyArgs = true
			continue
		}

		if isLongFlag(arg) {

			var name string
			var value *string
			var values []string
			name = longFlagName(arg)
			value = longFlagValue(arg)
			flag, ok := longFlagMap[name]
			if !ok {
				return result, fmt.Errorf("unknown flag: %s", arg)
			}

			var requiresValue bool = false
			if flag.GetParameter().Type() != ParameterTypeBool {
				requiresValue = true
			}

			var isList bool
			switch flag.GetParameter().Type() {
			case ParameterTypeStringList:
			case ParameterTypeBoolList:
			case ParameterTypeIntList:
			case ParameterTypeDatetimeList:
			case ParameterTypeDurationList:
			case ParameterTypePathList:
				isList = true
			default:
				isList = false
			}

			// For flags that take a value, if we don't find one in the --foo=bar form,
			// then we assume the the argument(s) follow
			if requiresValue && value == nil {
				i++
				if i < len(unparsedTokens) {
					return result, fmt.Errorf("expected value for %s, unexpected end of input", arg)
				}

				value = &unparsedTokens[i]

				// A list specifier is for arguments that expect lists, using [ ] and +, for example
				// [a, b, c] can be passed as [ +a +b +c ]
				// This syntax is supported to avoid injection attacks.
				if opts.Style == ParsingStyleRPNX && *value == "[" && isList {
					i++
					for {
						if i >= len(unparsedTokens) {
							return result, fmt.Errorf("while parsing list spec for --%s, unexpected end of input", name)
						}
						arg2 := unparsedTokens[i]
						if arg2 == "]" {
							break
						} else if arg2[:1] == "+" {
							arg2 = arg2[1:]
							values = append(values, arg2)
						} else {
							return result, fmt.Errorf("while parsing list spec for --%s, unexpected argument format in list spec: %s", name, arg2)
						}

					}
				} else if isList && flag.Greedy() {
					// If greedy, accept as many normal args as appear in the input
					for {
						if i >= len(unparsedTokens) {
							break
						}
						arg2 := unparsedTokens[i]
						if isLongFlag(arg2) || isShortFlag(arg2) {
							// Stop and backoff this argument
							i--
							break
						} else if arg == argTerminator {
							onlyArgs = true
							break
						} else {
							values = append(values, arg2)
							i++
							continue
						}

					}
				} else if isShortFlag(*value) || isLongFlag(*value) {
					return result, fmt.Errorf("expected value for --%s, found %s", name, unparsedTokens[i])
				} else if isList {
					values = append(values, *value)
				}
			}

			policy := flag.Overwrite()
			if policy == OverwritePolicyDefault {
				if isList {
					policy = OverwritePolicyAppend
				} else {
					policy = OverwritePolicyDisallowed
				}
			}

			if seenParameters[flag.GetParameter().Key()] {

				if policy == OverwritePolicyDisallowed {
					return result, fmt.Errorf("duplicate flag --%s: parameter has already been set", flag.Long())
				} else if policy == OverwritePolicyAppend {
					if !isList {
						return result, fmt.Errorf("flagspec error, flag --% has OverwritePolicyAppend but is not a list", flag.Long())
					}
					var err error
					ctx, err = AppendParameter(ctx, flag.GetParameter(), values)
					if err != nil {
						return result, fmt.Errorf("flag --%s with args %v: %w", flag.Long(), values, err)
					}
				} else if policy == OverwritePolicyAllowed {
					var err error
					if isList {
						ctx, err = SetParameterList(ctx, flag.GetParameter(), values)
					} else {
						if value != nil {
							ctx, err = SetParameter(ctx, flag.GetParameter(), *value)
						} else {
							ctx, err = SetParameter(ctx, flag.GetParameter(), "enabled")
						}
					}
					if err != nil {
						return result, fmt.Errorf("flag --%s with args %v: %w", flag.Long(), values, err)
					}
				}
			}
		} else if isShortFlag(arg) {
			cluster := arg[1:]
			for j := 0; j < len(cluster); j++ {
				char := cluster[j]
				name := string(char)
				flag, ok := shortFlagMap[name]
				if !ok {
					return result, fmt.Errorf("unknown flag: -%s", name)
				}

				var requiresValue bool = false
				if flag.GetParameter().Type() != ParameterTypeBool {
					requiresValue = true
				}

				if seenParameters[flag.GetParameter().Key()] {
					return result, fmt.Errorf("duplicate flag --%s: parameter has already been set", flag.Long())
				}

				if requiresValue {
					if j != len(cluster)-1 {
						valueString := cluster[j+1:]
						if valueString[:1] == setValueChar {
							// if this looks like -D= or /D:value then we skip
							valueString = valueString[1:]
						}
						cluster = cluster[:j+1]
						var err error
						ctx, err = SetParameter(ctx, flag.GetParameter(), valueString)
						if err != nil {
							return result, fmt.Errorf("flag --%s with args %v: %w", flag.Long(), valueString, err)
						}
						break
					}
					if i+1 >= len(unparsedTokens) {
						return result, fmt.Errorf("missing value for flag: -%s", name)
					}
					val := unparsedTokens[i+1]
					var err error
					ctx, err = SetParameter(ctx, flag.GetParameter(), val)
					if err != nil {
						return result, fmt.Errorf("flag --%s with args %v: %w", flag.Long(), val, err)
					}
					i++
				} else {
					var err error
					ctx, err = SetParameter(ctx, flag.GetParameter(), "enabled")
					if err != nil {
						return result, fmt.Errorf("invalid value for flag -%s: %w", name, err)
					}
				}
			}
		} else {
			var argument *Argument
			if argIndex >= len(arguments) {
				if len(arguments) != 0 && arguments[len(arguments)-1].Variadic() {
					argument = &arguments[len(arguments)-1]
				}
				return result, fmt.Errorf("unexpected argument: %s", unparsedTokens[i])
			} else {
				argument = &arguments[argIndex]
			}

			var isList bool
			switch (*argument).GetParameter().Type() {
			case ParameterTypeStringList:
			case ParameterTypeBoolList:
			case ParameterTypeIntList:
			case ParameterTypeDatetimeList:
			case ParameterTypeDurationList:
			case ParameterTypePathList:
				isList = true
			default:
				isList = false
			}

			if isList {

			}

		}

	}
}
