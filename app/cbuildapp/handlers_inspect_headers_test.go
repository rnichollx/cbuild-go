package cbuildapp

import (
	"reflect"
	"testing"
)

func TestSplitShellCommand(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{
			name:  "basic",
			input: "clang++ -I include -c src/main.cpp -o CMakeFiles/main.o",
			want:  []string{"clang++", "-I", "include", "-c", "src/main.cpp", "-o", "CMakeFiles/main.o"},
		},
		{
			name:  "quotes and escaped spaces",
			input: "clang++ -I\"/tmp/a b\" -DNAME='hello world' src/main.cpp",
			want:  []string{"clang++", "-I/tmp/a b", "-DNAME=hello world", "src/main.cpp"},
		},
		{
			name:    "unterminated",
			input:   "clang++ -I\"/tmp/a b src/main.cpp",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := splitShellCommand(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("splitShellCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildDependencyArgs(t *testing.T) {
	input := []string{"-I", "include", "-c", "src/main.cpp", "-o", "CMakeFiles/main.o", "-MD", "-MF", "main.d"}
	got := buildDependencyArgs(input)
	want := []string{"-I", "include", "src/main.cpp", "-w", "-MM"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("buildDependencyArgs() = %v, want %v", got, want)
	}
}

func TestParseMakeDependencies(t *testing.T) {
	output := []byte("main.o: src/main.cpp include/a.hpp include/with\\ name.hpp \\\n include/b.hpp\n")
	got, err := parseMakeDependencies(output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"src/main.cpp", "include/a.hpp", "include/with name.hpp", "include/b.hpp"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseMakeDependencies() = %v, want %v", got, want)
	}
}

func TestPathWithinRoot(t *testing.T) {
	root := "/tmp/work/sources/lib"
	if !pathWithinRoot("/tmp/work/sources/lib/include/a.hpp", root) {
		t.Fatalf("expected path to be within root")
	}
	if pathWithinRoot("/tmp/work/buildspaces/lib/include/a.hpp", root) {
		t.Fatalf("expected buildspaces path to be outside root")
	}
}

func TestIsHeaderFile(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{path: "foo/bar/a.h", want: true},
		{path: "foo/bar/a.HPP", want: true},
		{path: "foo/bar/a.tpp", want: true},
		{path: "foo/bar/a.cpp", want: false},
		{path: "foo/bar/CMakeLists.txt", want: false},
	}

	for _, tt := range tests {
		if got := isHeaderFile(tt.path); got != tt.want {
			t.Fatalf("isHeaderFile(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}
