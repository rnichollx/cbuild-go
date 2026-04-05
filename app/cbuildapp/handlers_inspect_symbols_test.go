package cbuildapp

import (
	"reflect"
	"testing"
)

type symbolNameSize struct {
	Name string
	Size uint64
}

func toSymbolNameSizes(in []parsedSymbol) []symbolNameSize {
	out := make([]symbolNameSize, 0, len(in))
	for _, sym := range in {
		out = append(out, symbolNameSize{Name: sym.Name, Size: sym.Size})
	}
	return out
}

func TestParseNMDefinedSymbols(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []symbolNameSize
	}{
		{
			name: "gnu style with undefined symbol",
			input: "0000000000000000 T foo\n" +
				"                 U bar\n" +
				"0000000000000010 t local\n",
			want: []symbolNameSize{{Name: "foo", Size: 0x10}},
		},
		{
			name: "with object prefix and demangled symbol",
			input: "obj.o: 00000000 00000020 W std::vector<int, std::allocator<int> >::size() const\n" +
				"obj.o:                 U memcpy\n",
			want: []symbolNameSize{{Name: "std::vector<int, std::allocator<int> >::size() const", Size: 0x20}},
		},
		{
			name: "short format",
			input: "T exported\n" +
				"U imported\n",
			want: []symbolNameSize{{Name: "exported"}},
		},
		{
			name: "ignores non-letter type tokens",
			input: "obj.o: 00000000 (__TEXT,__text) external foo\n" +
				"obj.o: 00000010 T real\n",
			want: []symbolNameSize{{Name: "real"}},
		},
		{
			name: "filters junk local labels",
			input: "0000000000000000 t ltmp0\n" +
				"0000000000000010 S l_.str\n" +
				"0000000000000020 T real_symbol\n",
			want: []symbolNameSize{{Name: "real_symbol"}},
		},
		{
			name: "estimates sizes from symbol addresses when nm size is zero",
			input: "0000000000000000 T a\n" +
				"0000000000000010 T b\n" +
				"0000000000000030 T c\n",
			want: []symbolNameSize{
				{Name: "a", Size: 0x10},
				{Name: "b", Size: 0x20},
				{Name: "c", Size: 0},
			},
		},
		{
			name: "uses local symbols as boundaries for text size estimation",
			input: "0000000000000000 T a\n" +
				"0000000000000010 t ltmp0\n" +
				"0000000000000020 T b\n",
			want: []symbolNameSize{
				{Name: "a", Size: 0x10},
				{Name: "b", Size: 0},
			},
		},
		{
			name: "does not estimate sparse data symbol sizes from addresses",
			input: "0000000000000000 B guard_a\n" +
				"0000000000010000 B guard_b\n",
			want: []symbolNameSize{
				{Name: "guard_a", Size: 0},
				{Name: "guard_b", Size: 0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toSymbolNameSizes(parseNMDefinedSymbols([]byte(tt.input)))
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("parseNMDefinedSymbols() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsStdLikeSymbol(t *testing.T) {
	tests := []struct {
		sym  string
		want bool
	}{
		{sym: "std::__1::basic_string<char>::size() const", want: true},
		{sym: "unsigned long std::__1::__constexpr_strlen[abi:ne200100]<char>(char const*)", want: true},
		{sym: "_ZSt4cout", want: true},
		{sym: "_ZNSt3__112basic_stringIcNS_11char_traitsIcEENS_9allocatorIcEEE4sizeEv", want: true},
		{sym: "my_ns::Foo<std::__1::basic_string<char>>::bar() const", want: false},
		{sym: "my_ns::template_func<std::vector<int>>()", want: false},
		{sym: "my_project::api::DoThing()", want: false},
	}

	for _, tt := range tests {
		if got := isStdLikeSymbol(tt.sym); got != tt.want {
			t.Fatalf("isStdLikeSymbol(%q) = %v, want %v", tt.sym, got, tt.want)
		}
	}
}
