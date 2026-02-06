package cli

type Flag struct {
	Short     string
	Long      string
	Parameter Parameter
	Greedy    bool
	Overwrite OverwritePolicy
}

type Argument struct {
	Name      string
	Parameter Parameter
	Overwrite OverwritePolicy
	Variadic  bool
	Separator *string
}

type OverwritePolicy int

const (
	// OverwritePolicyDefault is the same as OverwritePolicyAppend for lists and OverwritePolicyDisallowed otherwise.
	OverwritePolicyDefault OverwritePolicy = iota
	OverwritePolicyDisallowed
	OverwritePolicyAppend
	OverwritePolicyAllowed
)

// MixingPolicy defines the parsing policy for Parameters which are specified both as Flags and Arguments.
// This only applies when the same Parameter is available both as a flag and also as an argument, it has no influence
// for two different parameters are available as flags and arguments.
type MixingPolicy int

const (
	// MixingPolicySequencedArgs treats arguments strictly in their argument order.
	// Regardless of the order which flags appear, arguments will always be parsed in the order they are specified.
	// Given the Arguments A and B, with associated flags --a and --b then:
	// `abc --a def` this is treated as two attempts to set A.
	MixingPolicySequencedArgs MixingPolicy = iota

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
