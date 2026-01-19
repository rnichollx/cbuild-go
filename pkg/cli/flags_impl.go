package cli

type baseFlag struct {
	short     string
	long      string
	parameter Parameter
	greedy    bool
	overwrite OverwritePolicy
}

func (f *baseFlag) Short() string              { return f.short }
func (f *baseFlag) Long() string               { return f.long }
func (f *baseFlag) GetParameter() Parameter    { return f.parameter }
func (f *baseFlag) Greedy() bool               { return f.greedy }
func (f *baseFlag) Overwrite() OverwritePolicy { return f.overwrite }

type baseArgument struct {
	name      string
	parameter Parameter
	overwrite OverwritePolicy
	variadic  bool
	separator *string
}

func (a *baseArgument) Name() string               { return a.name }
func (a *baseArgument) GetParameter() Parameter    { return a.parameter }
func (a *baseArgument) Overwrite() OverwritePolicy { return a.overwrite }
func (a *baseArgument) Variadic() bool             { return a.variadic }
func (a *baseArgument) Separator() *string         { return a.separator }

func NewBoolFlag(short string, long string, param Parameter) Flag {
	return &baseFlag{short: short, long: long, parameter: param}
}

func NewStringFlag(short string, long string, param Parameter) Flag {
	return &baseFlag{short: short, long: long, parameter: param}
}

func NewIntFlag(short string, long string, param Parameter) Flag {
	return &baseFlag{short: short, long: long, parameter: param}
}

func NewStringArgument(name string, param Parameter) Argument {
	return &baseArgument{name: name, parameter: param}
}

func NewBoolArgument(name string, param Parameter) Argument {
	return &baseArgument{name: name, parameter: param}
}

func NewIntArgument(name string, param Parameter) Argument {
	return &baseArgument{name: name, parameter: param}
}

type genericParameter struct {
	key          ParameterKey
	defaultValue any
	paramType    ParameterType
	description  string
	required     bool
}

func (p *genericParameter) Key() ParameterKey   { return p.key }
func (p *genericParameter) Default() any        { return p.defaultValue }
func (p *genericParameter) Type() ParameterType { return p.paramType }
func (p *genericParameter) Description() string { return p.description }
func (p *genericParameter) Required() bool      { return p.required }

func NewParameter(key ParameterKey, paramType ParameterType, defaultValue any, description string, required bool) Parameter {
	return &genericParameter{
		key:          key,
		paramType:    paramType,
		defaultValue: defaultValue,
		description:  description,
		required:     required,
	}
}

func PBool(b bool) *bool {
	return &b
}

func PString(s string) *string {
	return &s
}

func PInt(i int64) *int64 {
	return &i
}
