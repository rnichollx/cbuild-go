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
	Name                  string
	Description           string
	HelpText              string
	AcceptsFlags          []Flag
	Arguments             []Argument
	AllowUnrecognizedArgs bool
	AllowUnknownFlags     bool
	Exec                  func(ctx context.Context, args []string) error
}

type Runner struct {
	Name          string
	Description   string
	GlobalFlags   []Flag
	Subcommands   map[string]*Subcommand
	DefaultSubcmd string
}

func (r *Runner) Run(ctx context.Context, args []string) error {
	// 1. Identify subcommand
	// We need to find the subcommand name in args, considering it might be preceded by flags.
	subcmdName := ""
	subcmdIdx := -1

	// Preliminary parse of global flags to find where the subcommand might be
	preliminaryResult, _ := ParseFlagsAndArgs(ParseOptions{
		Flags: r.GlobalFlags,
	}, ParseInput{Ctx: ctx, Tokens: args})
	preliminaryArgs := preliminaryResult.Unparsed

	if len(preliminaryArgs) > 0 {
		potentialSubcmd := preliminaryArgs[0]
		if _, ok := r.Subcommands[potentialSubcmd]; ok {
			subcmdName = potentialSubcmd
			// Find the index of this subcommand in the original args
			for i, arg := range args {
				if arg == subcmdName {
					subcmdIdx = i
					break
				}
			}
		}
	}

	if subcmdName == "" && r.DefaultSubcmd != "" {
		subcmdName = r.DefaultSubcmd
	}

	// 2. Help detection
	// We check for help flag in the whole arguments list.
	// If help is requested, we print usage and exit.
	helpFlags := []Flag{}
	for _, f := range r.GlobalFlags {
		if f.GetParameter().Key() == "help" {
			helpFlags = append(helpFlags, f)
		}
	}
	// Also check subcommand flags for help if a subcommand is identified
	if subcmdName != "" {
		for _, f := range r.Subcommands[subcmdName].AcceptsFlags {
			if f.GetParameter().Key() == "help" {
				helpFlags = append(helpFlags, f)
			}
		}
	}

	helpResult, _ := ParseFlagsAndArgs(ParseOptions{
		Flags: helpFlags,
	}, ParseInput{Ctx: ctx, Tokens: args})
	helpCtx := helpResult.Ctx

	helpVal, _ := GetBool(helpCtx, NewParameter("help", ParameterTypeBool, nil, "", false))
	if helpVal != nil && *helpVal {
		// If help is requested and no subcommand was explicitly found,
		// show general help even if there is a default subcommand.
		if subcmdIdx == -1 {
			r.PrintUsage("")
		} else {
			r.PrintUsage(subcmdName)
		}
		return nil
	}

	if subcmdName == "" {
		r.PrintUsage("")
		if len(preliminaryArgs) > 0 {
			return fmt.Errorf("unknown subcommand: %s", preliminaryArgs[0])
		}
		return nil
	}

	subcmd := r.Subcommands[subcmdName]

	// 3. Merge flags
	// Subcommand flags override global flags with the same Key
	mergedFlagsMap := make(map[ParameterKey]Flag)
	for _, f := range r.GlobalFlags {
		mergedFlagsMap[f.GetParameter().Key()] = f
	}
	for _, f := range subcmd.AcceptsFlags {
		mergedFlagsMap[f.GetParameter().Key()] = f
	}

	var mergedFlags []Flag
	for _, f := range mergedFlagsMap {
		mergedFlags = append(mergedFlags, f)
	}

	// 4. Parse all flags together
	// We need to remove the subcommand name from args if it was explicitly provided
	var finalArgs []string
	if subcmdIdx != -1 {
		finalArgs = append(args[:subcmdIdx], args[subcmdIdx+1:]...)
	} else {
		finalArgs = args
	}

	ctx, remainingArgs, err := ParseFlags(ctx, ParseOptions{
		Flags: mergedFlags,
	}, finalArgs)
	if err != nil {
		return err
	}

	if !subcmd.AllowUnrecognizedArgs && len(remainingArgs) > len(subcmd.Arguments) {
		// If we have more arguments than explicitly defined, and unrecognized ones aren't allowed
		return fmt.Errorf("subcommand %s does not accept unrecognized arguments", subcmdName)
	}

	return subcmd.Exec(ctx, remainingArgs)
}

func (r *Runner) PrintUsage(subcmdName string) {
	if subcmdName != "" {
		subcmd := r.Subcommands[subcmdName]
		argsSyn := ""

		for _, arg := range subcmd.Arguments {
			if arg.GetParameter().Required() {
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
		for _, f := range subcmd.AcceptsFlags {
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
					} else if f.GetParameter().Required() && !gf.GetParameter().Required() {
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
		for _, f := range sub.AcceptsFlags {
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
		for _, f := range sub.AcceptsFlags {
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
		if len(sub.AcceptsFlags) > 0 {
			for _, f := range sub.AcceptsFlags {
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
						} else if f.GetParameter().Required() && !gf.GetParameter().Required() {
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
			if arg.GetParameter().Required() {
				argsSyn += " <" + arg.Name() + ">"
			} else {
				argsSyn += " [" + arg.Name() + "]"
			}
		}
		fmt.Fprintf(w, ".B %s %s\n%s\n\n", r.Name, name, argsSyn)

		// Subcommand flags
		if len(sub.AcceptsFlags) > 0 {
			fmt.Fprintf(w, ".B Options for %s:\n", name)
			for _, f := range sub.AcceptsFlags {
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
					} else if f.GetParameter().Required() && !gf.GetParameter().Required() {
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
