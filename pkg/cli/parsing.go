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

	Style          ParsingStyle
	Subcommands    map[string]SubcommandParseOptions
	Flags          []Flag
	Arguments      []Argument
	RequiredParams []Parameter
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

	// Unparsed contains any tokens that were encountered if a subcommand has StopParsing set.
	Unparsed []string
}

// ParseInput contains the input to the parser. To handle special cases like `--` in subcommands,
// we need to have an UnparsedArgs field for things that must always be arguments.
type ParseInput struct {
	Ctx    context.Context
	Tokens []string
}

func ParseFlags(ctx context.Context, opts ParseOptions, args []string) (context.Context, []string, error) {
	if ctx == nil {
		panic("ctx must not be nil")
	}
	result, err := ParseFlagsAndArgs(opts, ParseInput{
		Ctx:    ctx,
		Tokens: args,
	})
	if err != nil {
		return result.Ctx, nil, err
	}
	return result.Ctx, result.Unparsed, nil
}

func ParseFlagsAndArgs(opts ParseOptions, input ParseInput) (ParseResult, error) {
	if input.Ctx == nil {
		panic("ctx must not be nil")
	}
	var result ParseResult

	shortFlagMap := make(map[string]Flag)
	longFlagMap := make(map[string]Flag)

	arguments := opts.Arguments

	allParameters := make(map[ParameterKey]Parameter)

	result.Ctx = input.Ctx
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

		param := flag.GetParameter()
		if param == nil {
			return result, fmt.Errorf("missing parameter for flag: %s", flag.Short())
		}

		allParameters[param.Key()] = param
	}

	for _, arg := range opts.Arguments {
		param := arg.GetParameter()
		if param == nil {
			return result, fmt.Errorf("missing parameter for argument: %s", arg.Name())
		}
		allParameters[param.Key()] = param
	}

	seenParameters := make(map[ParameterKey]bool)

	unparsedTokens := input.Tokens

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
		} else if len(arg) >= 2 && arg[:2] == "--" && arg != argTerminator {
			return true
		} else if len(arg) >= 1 && arg[:1] == "-" && (opts.Style == ParsingStyleShortWindows || opts.Style == ParsingStyleShort) {
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
		name := arg
		if strings.Contains(arg, setValueChar) {
			name = strings.SplitN(arg, setValueChar, 2)[0]
		}
		switch opts.Style {
		case ParsingStyleWindows:
			return name[1:]
		case ParsingStyleShortWindows, ParsingStyleShort:
			return name[1:]
		case ParsingStyleRPNX, ParsingStyleGNU, ParsingStylePOSIX:
			if len(name) >= 2 && name[:2] == "--" {
				return name[2:]
			}
			return name[1:]
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
			if i+1 < len(unparsedTokens) {
				unparsed := unparsedTokens[i+1:]
				for _, u := range unparsed {
					var argument Argument
					if argIndex >= len(arguments) {
						if len(arguments) != 0 && arguments[len(arguments)-1].Variadic() {
							argument = arguments[len(arguments)-1]
						} else {
							return result, fmt.Errorf("unexpected argument: %s", u)
						}
					} else {
						argument = arguments[argIndex]
					}

					var isList bool
					switch argument.GetParameter().Type() {
					case ParameterTypeStringList, ParameterTypeBoolList, ParameterTypeIntList, ParameterTypeDatetimeList, ParameterTypeDurationList, ParameterTypePathList, ParameterTypeURIList:
						isList = true
					default:
						isList = false
					}

					if isList {
						var values []string
						if argument.Separator() != nil {
							values = strings.Split(u, *argument.Separator())
						} else {
							values = []string{u}
						}

						policy := argument.Overwrite()
						if policy == OverwritePolicyDefault {
							policy = OverwritePolicyAppend
						}

						if seenParameters[argument.GetParameter().Key()] && policy == OverwritePolicyDisallowed {
							return result, fmt.Errorf("duplicate argument %s: parameter has already been set", argument.Name())
						}

						var err error
						if policy == OverwritePolicyAppend {
							result.Ctx, err = AppendParameter(result.Ctx, argument.GetParameter(), values)
						} else {
							result.Ctx, err = SetParameterList(result.Ctx, argument.GetParameter(), values)
						}
						if err != nil {
							return result, fmt.Errorf("argument %s with args %v: %w", argument.Name(), values, err)
						}
					} else {
						if seenParameters[argument.GetParameter().Key()] {
							policy := argument.Overwrite()
							if policy == OverwritePolicyDefault || policy == OverwritePolicyDisallowed {
								return result, fmt.Errorf("duplicate argument %s: parameter has already been set", argument.Name())
							}
						}
						var err error
						result.Ctx, err = SetParameter(result.Ctx, argument.GetParameter(), u)
						if err != nil {
							return result, fmt.Errorf("argument %s with args %s: %w", argument.Name(), u, err)
						}
					}
					seenParameters[argument.GetParameter().Key()] = true
					if !argument.Variadic() {
						argIndex++
					}
				}
			}
			break
		}

		if isLongFlag(arg) {
			var name string
			var value *string
			var values []string
			name = longFlagName(arg)
			value = longFlagValue(arg)
			flag, ok := longFlagMap[name]
			if !ok && opts.Style == ParsingStyleWindows {
				// Try matching short flag if long flag fails in Windows style
				// to support things like /Dfoo or /D:foo
				if len(name) > 0 {
					shortName := name[:1]
					if f, sok := shortFlagMap[shortName]; sok {
						flag = f
						ok = true
						if value == nil && len(name) > 1 {
							// /Dfoo case
							valStr := name[1:]
							value = &valStr
						}
					}
				}
			}

			if !ok {
				return result, fmt.Errorf("unknown flag: %s", arg)
			}

			var requiresValue bool = false
			if flag.GetParameter().Type() != ParameterTypeBool {
				requiresValue = true
			}

			var isList bool
			switch flag.GetParameter().Type() {
			case ParameterTypeStringList, ParameterTypeBoolList, ParameterTypeIntList, ParameterTypeDatetimeList, ParameterTypeDurationList, ParameterTypePathList, ParameterTypeURIList:
				isList = true
			default:
				isList = false
			}

			// For flags that take a value, if we don't find one in the --foo=bar form,
			// then we assume the the argument(s) follow
			if requiresValue && value == nil {
				i++
				if i >= len(unparsedTokens) {
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
						i++
					}
				} else if isList && flag.Greedy() {
					values = append(values, *value)
					// If greedy, accept as many normal args as appear in the input
					for {
						if i+1 >= len(unparsedTokens) {
							break
						}
						arg2 := unparsedTokens[i+1]
						if isLongFlag(arg2) || isShortFlag(arg2) {
							break
						} else if arg2 == argTerminator {
							break
						} else {
							i++
							values = append(values, arg2)
						}
					}
				} else if isShortFlag(*value) || isLongFlag(*value) {
					prefix := "--"
					if opts.Style == ParsingStyleShort || opts.Style == ParsingStyleShortWindows || opts.Style == ParsingStyleWindows {
						prefix = "-"
					}
					if opts.Style == ParsingStyleWindows {
						prefix = "/"
					}
					return result, fmt.Errorf("expected value for %s%s, found %s", prefix, name, *value)
				} else if isList {
					values = append(values, *value)
				}
			} else if value != nil {
				if isList {
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
				}
			}

			if isList {
				if policy == OverwritePolicyAppend {
					var err error
					result.Ctx, err = AppendParameter(result.Ctx, flag.GetParameter(), values)
					if err != nil {
						return result, fmt.Errorf("flag --%s with args %v: %w", flag.Long(), values, err)
					}
				} else {
					var err error
					result.Ctx, err = SetParameterList(result.Ctx, flag.GetParameter(), values)
					if err != nil {
						return result, fmt.Errorf("flag --%s with args %v: %w", flag.Long(), values, err)
					}
				}
			} else {
				var err error
				if value != nil {
					result.Ctx, err = SetParameter(result.Ctx, flag.GetParameter(), *value)
				} else {
					result.Ctx, err = SetParameter(result.Ctx, flag.GetParameter(), "enabled")
				}
				if err != nil {
					return result, fmt.Errorf("flag --%s with args %v: %w", flag.Long(), value, err)
				}
			}
			seenParameters[flag.GetParameter().Key()] = true
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

				var isList bool
				switch flag.GetParameter().Type() {
				case ParameterTypeStringList, ParameterTypeBoolList, ParameterTypeIntList, ParameterTypeDatetimeList, ParameterTypeDurationList, ParameterTypePathList, ParameterTypeURIList:
					isList = true
				default:
					isList = false
				}

				var values []string
				var value *string

				if requiresValue {
					if j != len(cluster)-1 {
						valueString := cluster[j+1:]
						if valueString[:1] == setValueChar {
							// if this looks like -D= or /D:value then we skip
							valueString = valueString[1:]
						}
						value = &valueString
						j = len(cluster) // consume the rest of the cluster
					} else {
						if i+1 >= len(unparsedTokens) {
							return result, fmt.Errorf("missing value for flag: -%s", name)
						}
						i++
						val := unparsedTokens[i]
						value = &val
					}
					if isList {
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
						return result, fmt.Errorf("duplicate flag -%s: parameter has already been set", name)
					}
				}

				if isList {
					if policy == OverwritePolicyAppend {
						var err error
						result.Ctx, err = AppendParameter(result.Ctx, flag.GetParameter(), values)
						if err != nil {
							return result, fmt.Errorf("flag -%s with args %v: %w", name, values, err)
						}
					} else {
						var err error
						result.Ctx, err = SetParameterList(result.Ctx, flag.GetParameter(), values)
						if err != nil {
							return result, fmt.Errorf("flag -%s with args %v: %w", name, values, err)
						}
					}
				} else {
					var err error
					if value != nil {
						result.Ctx, err = SetParameter(result.Ctx, flag.GetParameter(), *value)
					} else {
						result.Ctx, err = SetParameter(result.Ctx, flag.GetParameter(), "enabled")
					}
					if err != nil {
						return result, fmt.Errorf("flag -%s with args %v: %w", name, value, err)
					}
				}
				seenParameters[flag.GetParameter().Key()] = true
			}
		} else {
			if argIndex == 0 && !onlyArgs {
				if subcmd, ok := opts.Subcommands[arg]; ok {
					result.Subcommands = append(result.Subcommands, arg)
					if subcmd.StopParsing {
						result.Unparsed = unparsedTokens[i+1:]

						return result, nil
					}

					if subcmd.ParseOptions != nil {
						// Update current options for the next level
						opts.Flags = subcmd.ParseOptions.Flags
						longFlagMap = make(map[string]Flag)
						shortFlagMap = make(map[string]Flag)
						for _, flag := range opts.Flags {
							if flag.Short() != "" {
								shortFlagMap[flag.Short()] = flag
							}
							if flag.Long() != "" {
								longFlagMap[flag.Long()] = flag
							}
							parm := flag.GetParameter()
							allParameters[parm.Key()] = parm
						}

						opts.Arguments = subcmd.ParseOptions.Arguments
						for _, arg := range opts.Arguments {
							parm := arg.GetParameter()
							allParameters[parm.Key()] = parm
						}
						arguments = opts.Arguments
						argIndex = 0
						opts.Subcommands = subcmd.ParseOptions.Subcommands
						opts.Style = subcmd.ParseOptions.Style
						opts.StrictOrderingArgs = subcmd.ParseOptions.StrictOrderingArgs
						opts.RequiredParams = subcmd.ParseOptions.RequiredParams
						seenParameters = make(map[ParameterKey]bool)
						continue
					}
					return result, fmt.Errorf("subcommand %s has no parse options and StopParsing is not set", arg)
				}
			}

			var argument Argument
			if argIndex >= len(arguments) {
				if len(arguments) != 0 && arguments[len(arguments)-1].Variadic() {
					argument = arguments[len(arguments)-1]
				} else {
					return result, fmt.Errorf("unexpected argument: %s", unparsedTokens[i])
				}
			} else {
				argument = arguments[argIndex]
			}

			var isList bool
			switch argument.GetParameter().Type() {
			case ParameterTypeStringList, ParameterTypeBoolList, ParameterTypeIntList, ParameterTypeDatetimeList, ParameterTypeDurationList, ParameterTypePathList, ParameterTypeURIList:
				isList = true
			default:
				isList = false
			}

			if isList {
				var values []string
				if argument.Separator() != nil {
					values = strings.Split(unparsedTokens[i], *argument.Separator())
				} else {
					values = []string{unparsedTokens[i]}
				}

				policy := argument.Overwrite()
				if policy == OverwritePolicyDefault {
					policy = OverwritePolicyAppend
				}

				if seenParameters[argument.GetParameter().Key()] && policy == OverwritePolicyDisallowed {
					return result, fmt.Errorf("duplicate argument %s: parameter has already been set", argument.Name())
				}

				var err error
				if policy == OverwritePolicyAppend {
					result.Ctx, err = AppendParameter(result.Ctx, argument.GetParameter(), values)
				} else {
					result.Ctx, err = SetParameterList(result.Ctx, argument.GetParameter(), values)
				}
				if err != nil {
					return result, fmt.Errorf("argument %s with args %v: %w", argument.Name(), values, err)
				}
			} else {
				if seenParameters[argument.GetParameter().Key()] {
					policy := argument.Overwrite()
					if policy == OverwritePolicyDefault || policy == OverwritePolicyDisallowed {
						return result, fmt.Errorf("duplicate argument %s: parameter has already been set", argument.Name())
					}
				}
				var err error
				result.Ctx, err = SetParameter(result.Ctx, argument.GetParameter(), unparsedTokens[i])
				if err != nil {
					return result, fmt.Errorf("argument %s with args %s: %w", argument.Name(), unparsedTokens[i], err)
				}
			}
			seenParameters[argument.GetParameter().Key()] = true
			if !argument.Variadic() {
				argIndex++
			}
		}

	}

	for _, param := range opts.RequiredParams {
		if param == nil {
			return result, fmt.Errorf("required parameter cannot be nil")
		}
		key := param.Key()
		if _, ok := allParameters[key]; !ok {
			return result, fmt.Errorf("required parameter %q is not defined in flags or arguments", key)
		}
		if _, have := seenParameters[key]; !have {
			return result, fmt.Errorf("missing value for required parameter %q", key)
		}
	}

	return result, nil
}
