package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Subcommand struct {
	Name           string
	Description    string
	HelpText       string
	Flags          []Flag
	Arguments      []Argument
	RequiredParams []Parameter
	Exec           func(ctx context.Context, args []string) error
}

type Runner struct {
	Name          string
	Description   string
	GlobalFlags   []Flag
	Subcommands   map[string]*Subcommand
	DefaultSubcmd string
}

func (r *Runner) Run(ctx context.Context, args []string) error {
	// 1. Identify subcommand and parse its options
	subcmdParseOpts := make(map[string]SubcommandParseOptions)
	for name, subcmd := range r.Subcommands {
		mergedFlags := make([]Flag, 0, len(r.GlobalFlags)+len(subcmd.Flags))
		mergedFlagsMap := make(map[ParameterKey]Flag)
		for _, f := range r.GlobalFlags {
			mergedFlagsMap[f.GetParameter().Key()] = f
		}
		for _, f := range subcmd.Flags {
			mergedFlagsMap[f.GetParameter().Key()] = f
		}
		for _, f := range mergedFlagsMap {
			mergedFlags = append(mergedFlags, f)
		}

		subcmdParseOpts[name] = SubcommandParseOptions{
			ParseOptions: &ParseOptions{
				Flags:          mergedFlags,
				Arguments:      subcmd.Arguments,
				RequiredParams: subcmd.RequiredParams,
			},
		}
	}

	result, err := ParseFlagsAndArgs(ParseOptions{
		Flags:       r.GlobalFlags,
		Subcommands: subcmdParseOpts,
	}, ParseInput{Ctx: ctx, Tokens: args})

	if err != nil {
		// If we failed to parse, maybe we should try with the default subcommand if no subcommand was explicitly matched
		if r.DefaultSubcmd != "" {
			subcmd := r.Subcommands[r.DefaultSubcmd]
			mergedFlags := make([]Flag, 0, len(r.GlobalFlags)+len(subcmd.Flags))
			mergedFlagsMap := make(map[ParameterKey]Flag)
			for _, f := range r.GlobalFlags {
				mergedFlagsMap[f.GetParameter().Key()] = f
			}
			for _, f := range subcmd.Flags {
				mergedFlagsMap[f.GetParameter().Key()] = f
			}
			for _, f := range mergedFlagsMap {
				mergedFlags = append(mergedFlags, f)
			}

			resultDefault, errDefault := ParseFlagsAndArgs(ParseOptions{
				Flags:          mergedFlags,
				Arguments:      subcmd.Arguments,
				RequiredParams: subcmd.RequiredParams,
			}, ParseInput{Ctx: ctx, Tokens: args})
			if errDefault == nil {
				return subcmd.Exec(resultDefault.Ctx, resultDefault.Unparsed)
			}
		}
		return err
	}

	// 2. Help detection
	helpVal := GetBool(result.Ctx, NewParameter("help", ParameterTypeBool, PBool(false), ""))
	if helpVal != nil && *helpVal {
		if len(result.Subcommands) == 0 {
			r.PrintUsage("")
		} else {
			r.PrintUsage(result.Subcommands[0])
		}
		return nil
	}

	if len(result.Subcommands) == 0 {
		if r.DefaultSubcmd != "" {
			subcmd := r.Subcommands[r.DefaultSubcmd]
			// We already parsed it if it was successfully identified as NOT a subcommand
			// and NOT erroring. But since we use subcmdParseOpts, if it's not in result.Subcommands,
			// it means it didn't match any subcommand name in the tokens.

			// Try re-parsing with default subcommand options if it wasn't already successfully parsed
			// actually, if result was successful and no subcommand found, we can just use default.
			mergedFlags := make([]Flag, 0, len(r.GlobalFlags)+len(subcmd.Flags))
			mergedFlagsMap := make(map[ParameterKey]Flag)
			for _, f := range r.GlobalFlags {
				mergedFlagsMap[f.GetParameter().Key()] = f
			}
			for _, f := range subcmd.Flags {
				mergedFlagsMap[f.GetParameter().Key()] = f
			}
			for _, f := range mergedFlagsMap {
				mergedFlags = append(mergedFlags, f)
			}

			result, err = ParseFlagsAndArgs(ParseOptions{
				Flags:          mergedFlags,
				Arguments:      subcmd.Arguments,
				RequiredParams: subcmd.RequiredParams,
			}, ParseInput{Ctx: ctx, Tokens: args})
			if err != nil {
				return err
			}
			return subcmd.Exec(result.Ctx, result.Unparsed)
		}

		r.PrintUsage("")
		return nil
	}

	subcmdName := result.Subcommands[0]
	subcmd := r.Subcommands[subcmdName]

	return subcmd.Exec(result.Ctx, result.Unparsed)
}

func (r *Runner) PrintUsage(subcmdName string) {
	if subcmdName != "" {
		subcmd := r.Subcommands[subcmdName]
		argsSyn := ""

		for _, arg := range subcmd.Arguments {
			if isRequiredParam(subcmd.RequiredParams, arg.GetParameter().Key()) {
				argsSyn += " <" + arg.Name() + ">"
			} else {
				argsSyn += " [" + arg.Name() + "]"
			}
		}

		fmt.Printf("Usage: %s %s%s\n\n", r.Name, subcmdName, argsSyn)
		if subcmd.Description != "" {
			fmt.Printf("%s\n\n", subcmd.Description)
		}
		if subcmd.HelpText != "" {
			fmt.Printf("%s\n\n", subcmd.HelpText)
		}

		// Show flags for this subcommand
		mergedFlagsMap := make(map[ParameterKey]Flag)
		for _, f := range r.GlobalFlags {
			mergedFlagsMap[f.GetParameter().Key()] = f
		}
		for _, f := range subcmd.Flags {
			mergedFlagsMap[f.GetParameter().Key()] = f
		}

		maxFlagLen := 0
		formatFlag := func(f Flag) string {
			s := ""
			val := ""
			if f.GetParameter().Type() != ParameterTypeBool {
				val = " <value>"
			}
			if f.Short() != "" {
				s += "-" + f.Short() + val + ", "
			} else {
				s += "    "
			}
			s += "--" + f.Long() + val
			return s
		}

		var flags []Flag
		for _, f := range mergedFlagsMap {
			flags = append(flags, f)
			l := len(formatFlag(f))
			if l > maxFlagLen {
				maxFlagLen = l
			}
		}

		// Sort flags by key for consistent output
		for i := 0; i < len(flags); i++ {
			for j := i + 1; j < len(flags); j++ {
				if flags[i].GetParameter().Key() > flags[j].GetParameter().Key() {
					flags[i], flags[j] = flags[j], flags[i]
				}
			}
		}

		if len(flags) > 0 {
			fmt.Println("Flags:")
			for _, f := range flags {
				s := formatFlag(f)
				indent := "  "
				desc := f.GetParameter().Description()
				if gf, ok := findGlobalFlag(r.GlobalFlags, f.GetParameter().Key()); ok {
					if f.GetParameter().Description() != "" && f.GetParameter().Description() != gf.GetParameter().Description() {
						desc += fmt.Sprintf(" (overrides global: %s)", gf.GetParameter().Description())
					} else if isRequiredParam(subcmd.RequiredParams, f.GetParameter().Key()) {
						desc += fmt.Sprintf(" (required for %s)", subcmdName)
					}
				}
				fmt.Printf("%s%s%s  %s\n", indent, s, strings.Repeat(" ", maxFlagLen-len(s)), desc)
			}
		}
		return
	}

	fmt.Printf("Usage: %s <subcommand>\n\n", r.Name)
	if r.Description != "" {
		fmt.Printf("%s\n\n", r.Description)
	}

	// Calculate which flags should be in the "Flags:" section
	flagCount := make(map[ParameterKey]int)
	flagMap := make(map[ParameterKey]Flag)
	flagSubcmds := make(map[ParameterKey][]string)

	for _, f := range r.GlobalFlags {
		flagMap[f.GetParameter().Key()] = f
		// Global flags are always shown in "Flags:"
	}

	for _, name := range sortedSubcommandNames(r.Subcommands) {
		sub := r.Subcommands[name]
		for _, f := range sub.Flags {
			flagCount[f.GetParameter().Key()]++
			if _, exists := flagMap[f.GetParameter().Key()]; !exists {
				flagMap[f.GetParameter().Key()] = f
			}
			flagSubcmds[f.GetParameter().Key()] = append(flagSubcmds[f.GetParameter().Key()], name)
		}
	}

	var flagsToShowGlobal []Flag
	for key, f := range flagMap {
		isGlobal := false
		for _, gf := range r.GlobalFlags {
			if gf.GetParameter().Key() == key {
				isGlobal = true
				break
			}
		}

		if isGlobal || flagCount[key] >= 2 {
			flagsToShowGlobal = append(flagsToShowGlobal, f)
		}
	}

	maxFlagLen := 0
	formatFlag := func(f Flag) string {
		s := ""
		val := ""
		if f.GetParameter().Type() != ParameterTypeBool {
			val = " <value>"
		}
		if f.Short() != "" {
			s += "-" + f.Short() + val + ", "
		} else {
			s += "    "
		}
		s += "--" + f.Long() + val
		return s
	}

	for _, f := range flagsToShowGlobal {
		l := len(formatFlag(f))
		if l > maxFlagLen {
			maxFlagLen = l
		}
	}

	for _, name := range sortedSubcommandNames(r.Subcommands) {
		l := len(name)
		if l > maxFlagLen {
			maxFlagLen = l
		}
		sub := r.Subcommands[name]
		for _, f := range sub.Flags {
			showInSub := true
			for _, gf := range flagsToShowGlobal {
				if gf.GetParameter().Key() == f.GetParameter().Key() {
					showInSub = false
					break
				}
			}
			if showInSub {
				l := len(formatFlag(f))
				if l > maxFlagLen {
					maxFlagLen = l
				}
			}
		}
	}

	// Sort flagsToShowGlobal by key
	for i := 0; i < len(flagsToShowGlobal); i++ {
		for j := i + 1; j < len(flagsToShowGlobal); j++ {
			if flagsToShowGlobal[i].GetParameter().Key() > flagsToShowGlobal[j].GetParameter().Key() {
				flagsToShowGlobal[i], flagsToShowGlobal[j] = flagsToShowGlobal[j], flagsToShowGlobal[i]
			}
		}
	}

	if len(flagsToShowGlobal) > 0 {
		fmt.Println("Flags:")
		for _, f := range flagsToShowGlobal {
			s := formatFlag(f)
			isGlobal := false
			for _, gf := range r.GlobalFlags {
				if gf.GetParameter().Key() == f.GetParameter().Key() {
					isGlobal = true
					break
				}
			}

			note := ""
			if !isGlobal {
				note = fmt.Sprintf(" (for %s subcommands only)", strings.Join(flagSubcmds[f.GetParameter().Key()], ", "))
			}

			fmt.Printf("  %s%s  %s%s\n", s, strings.Repeat(" ", maxFlagLen-len(s)), f.GetParameter().Description(), note)
		}
		fmt.Println()
	}

	fmt.Println("Subcommands:")
	for _, name := range sortedSubcommandNames(r.Subcommands) {
		sub := r.Subcommands[name]
		fmt.Printf("  %s%s  %s\n", name, strings.Repeat(" ", maxFlagLen-len(name)), sub.Description)
		if len(sub.Flags) > 0 {
			for _, f := range sub.Flags {
				showInSub := false
				for _, gf := range flagsToShowGlobal {
					if gf.GetParameter().Key() == f.GetParameter().Key() {
						// Only show in sub if description is different
						if f.GetParameter().Description() != "" && f.GetParameter().Description() != gf.GetParameter().Description() {
							showInSub = true
						}
						break
					}
				}
				if showInSub || !isFlagInGlobalList(flagsToShowGlobal, f.GetParameter().Key()) {
					s := formatFlag(f)
					indent := "    "
					desc := f.GetParameter().Description()
					if gf, ok := findGlobalFlag(r.GlobalFlags, f.GetParameter().Key()); ok {
						if f.GetParameter().Description() != "" && f.GetParameter().Description() != gf.GetParameter().Description() {
							desc += fmt.Sprintf(" (overrides global: %s)", gf.GetParameter().Description())
						} else if isRequiredParam(sub.RequiredParams, f.GetParameter().Key()) {
							desc += fmt.Sprintf(" (required for %s)", name)
						}
					}
					fmt.Printf("%s%s%s  %s\n", indent, s, strings.Repeat(" ", maxFlagLen-len(s)), desc)
				}
			}
		}
	}
}

func (r *Runner) generateManpage(dir string) error {
	filename := filepath.Join(dir, r.Name+".1")
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	w := f
	date := time.Now().Format("2006-01-02")
	fmt.Fprintf(w, ".TH %s 1 \"%s\" \"\" \"\"\n", strings.ToUpper(r.Name), date)

	fmt.Fprintf(w, ".SH NAME\n%s \\- %s\n", r.Name, r.Description)

	fmt.Fprintf(w, ".SH SYNOPSIS\n")
	fmt.Fprintf(w, ".B %s\n<\\fIsubcommand\\fR>\n", r.Name)

	fmt.Fprintf(w, ".SH DESCRIPTION\n")
	fmt.Fprintf(w, "%s\n", r.Description)

	if len(r.GlobalFlags) > 0 {
		fmt.Fprintf(w, ".SH GLOBAL OPTIONS\n")
		for _, f := range r.GlobalFlags {
			fmt.Fprintf(w, ".TP\n")
			val := ""
			if f.GetParameter().Type() != ParameterTypeBool {
				val = " <value>"
			}
			if f.Short() != "" {
				fmt.Fprintf(w, ".B \\-%s%s, \\-\\-%s%s", f.Short(), val, f.Long(), val)
			} else {
				fmt.Fprintf(w, ".B \\-\\-%s%s", f.Long(), val)
			}
			fmt.Fprintf(w, "\n")
			fmt.Fprintf(w, "%s\n", f.GetParameter().Description())
		}
	}

	fmt.Fprintf(w, ".SH SUBCOMMANDS\n")
	for _, subName := range sortedSubcommandNames(r.Subcommands) {
		sub := r.Subcommands[subName]
		name := subName
		if sub.Name != "" {
			name = sub.Name
		}

		fmt.Fprintf(w, ".SS %s\n", name)
		fmt.Fprintf(w, "%s\n\n", sub.Description)
		if sub.HelpText != "" {
			fmt.Fprintf(w, "%s\n\n", sub.HelpText)
		}

		fmt.Fprintf(w, ".B Synopsis:\n")
		argsSyn := ""
		for _, arg := range sub.Arguments {
			if isRequiredParam(sub.RequiredParams, arg.GetParameter().Key()) {
				argsSyn += " <" + arg.Name() + ">"
			} else {
				argsSyn += " [" + arg.Name() + "]"
			}
		}
		fmt.Fprintf(w, "%s %s%s\n\n", r.Name, name, argsSyn)

		// Subcommand flags
		if len(sub.Flags) > 0 {
			fmt.Fprintf(w, ".B Options for %s:\n", name)
			for _, f := range sub.Flags {
				fmt.Fprintf(w, ".TP\n")
				val := ""
				if f.GetParameter().Type() != ParameterTypeBool {
					val = " <value>"
				}
				if f.Short() != "" {
					fmt.Fprintf(w, ".B \\-%s%s, \\-\\-%s%s", f.Short(), val, f.Long(), val)
				} else {
					fmt.Fprintf(w, ".B \\-\\-%s%s", f.Long(), val)
				}
				fmt.Fprintf(w, "\n")
				desc := f.GetParameter().Description()
				if gf, ok := findGlobalFlag(r.GlobalFlags, f.GetParameter().Key()); ok {
					if f.GetParameter().Description() != "" && f.GetParameter().Description() != gf.GetParameter().Description() {
						desc += fmt.Sprintf(" (overrides global: %s)", gf.GetParameter().Description())
					} else if isRequiredParam(sub.RequiredParams, f.GetParameter().Key()) {
						desc += fmt.Sprintf(" (required for %s)", name)
					}
				}
				fmt.Fprintf(w, "%s\n", desc)
			}
		}
	}

	return nil
}

func (r *Runner) GenerateManpages(dir string) error {
	return r.generateManpage(dir)
}

func isFlagInGlobalList(flags []Flag, key ParameterKey) bool {
	for _, f := range flags {
		if f.GetParameter().Key() == key {
			return true
		}
	}
	return false
}

func isRequiredParam(required []Parameter, key ParameterKey) bool {
	for _, p := range required {
		if p == nil {
			continue
		}
		if p.Key() == key {
			return true
		}
	}
	return false
}

func findGlobalFlag(globalFlags []Flag, key ParameterKey) (Flag, bool) {
	for _, f := range globalFlags {
		if f.GetParameter().Key() == key {
			return f, true
		}
	}
	return nil, false
}

func sortedSubcommandNames(subcommands map[string]*Subcommand) []string {
	var names []string
	for name := range subcommands {
		names = append(names, name)
	}
	// Simple bubble sort or similar to keep it dependency-free if possible,
	// or just use sort package if allowed.
	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			if names[i] > names[j] {
				names[i], names[j] = names[j], names[i]
			}
		}
	}
	return names
}
