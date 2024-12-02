package ccommon

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"path/filepath"
)

type Workspace struct {
	SourcePath string `yaml:"source_path"`
	BuildPath  string `yaml:"build_path"`

	Modules map[string]*Module `yaml:"modules"`

	CMakeBinary *string `yaml:"cmake_binary"`
	BuildType   string  `yaml:"build_type"`
	CXXVersion  string  `yaml:"cxx_version"`
	Generator   *string `yaml:"generator"`

	DryRun bool
}

type Module struct {
	Depends                []string `yaml:"depends"`
	FindPackageName        *string  `yaml:"find_package_name"`
	FindPackageRoot        *string  `yaml:"find_package_root"`
	BuildType              *string  `yaml:"build_type"`
	SourceRoot             *string  `yaml:"source_root"`
	BuildRoot              *string  `yaml:"build_root"`
	ConfigPath             *string  `yaml:"config_path"`
	CMakeAdditionalOptions []string `yaml:"cmake_additional_options"`
}

func (m *Module) CMakeBuildType(ws *Workspace) string {
	if m.BuildType != nil {
		return *m.BuildType
	}
	if ws.BuildType != "" {
		return ws.BuildType
	}
	return ""
}

// CMakeConfigureArgs returns the arguments to pass to cmake when configuring the module
func (m *Module) CMakeConfigureArgs(workspace *Workspace, modname string) ([]string, error) {

	args := []string{}

	src, err := m.CMakeSourcePath(workspace, modname)
	if err != nil {
		return nil, err
	}

	args = append(args, "-S")
	args = append(args, src)

	bld, err := m.CMakeBuildPath(workspace, modname)
	if err != nil {
		return nil, err
	}

	args = append(args, "-B")
	args = append(args, bld)

	args = append(args, fmt.Sprintf("-DCMAKE_BUILD_TYPE=%s", m.CMakeBuildType(workspace)))

	if workspace.Generator != nil {
		args = append(args, fmt.Sprintf("-G%s", *workspace.Generator))
	}

	for _, dep := range m.Depends {
		mod, ok := workspace.Modules[dep]
		if !ok {
			return nil, fmt.Errorf("depend on unknown module %s", dep)
		}
		mod_args, err := mod.CMakeDependencyArgs(workspace, dep)
		if err != nil {
			return nil, err
		}
		args = append(args, mod_args...)
	}

	return args, nil

}

func (m *Module) CMakeSourcePath(workspace *Workspace, defaultRoot string) (string, error) {
	var isAbs bool

	path := m.SourceRoot
	if path == nil {

		return workspace.SourcePath + "/" + defaultRoot, nil
	}

	isAbs = filepath.IsAbs(*path)

	var path_str string
	if !isAbs {
		path_str = filepath.Join(workspace.SourcePath, *path)
	} else {
		path_str = *path
	}

	return path_str, nil
}

func (m *Module) CMakeConfigPath(workspace *Workspace, defaultRoot string) (string, error) {

	buildPath, err := m.CMakeBuildPath(workspace, defaultRoot)
	if err != nil {
		return "", err
	}

	if m.ConfigPath != nil {
		return filepath.Join(buildPath, *m.ConfigPath), nil
	}

	return buildPath, nil
}

func (m *Module) CMakeBuildPath(workspace *Workspace, defaultRoot string) (string, error) {
	var isAbs bool

	path := m.BuildRoot
	if path == nil {
		return workspace.BuildPath + "/" + defaultRoot, nil
	}

	isAbs = filepath.IsAbs(*path)

	var path_str string
	if !isAbs {
		path_str = filepath.Join(workspace.BuildPath, *path)
	} else {
		path_str = *path
	}

	return path_str, nil
}

// CMakeDependencyArgs returns the arguments to pass to cmake when configuring another module that depends on this module
func (m *Module) CMakeDependencyArgs(workspace *Workspace, modname string) ([]string, error) {
	args := []string{}

	if m.FindPackageName != nil {
		dirname := *m.FindPackageName + "_DIR"

		configPath, err := m.CMakeConfigPath(workspace, modname)
		if err != nil {
			return nil, err
		}

		args = append(args, fmt.Sprintf("-D%s=%s", dirname, configPath))
	}

	if m.FindPackageRoot != nil {
		dirname := *m.FindPackageRoot + "_ROOT"

		sourcePath, err := m.CMakeSourcePath(workspace, modname)
		if err != nil {
			return nil, err
		}

		args = append(args, fmt.Sprintf("-D%s=%s", dirname, sourcePath))
	}

	return args, nil
}

func (w *Workspace) LoadConfig(path string) error {
	// Load the configuration from the file
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	err = yaml.Unmarshal(yamlFile, w)

	if err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	return nil
}

func (w *Workspace) Build() error {
	var builtModules = make(map[string]bool)

	for name, mod := range w.Modules {
		err := w.buildModule(mod, name, builtModules)
		if err != nil {
			return fmt.Errorf("failed to build module %s: %w", name, err)
		}
	}

	return nil

}

func (w *Workspace) Exec(command string, args []string) error {
	fmt.Printf("Executing: %s %s\n", command, args)
	return nil
}

func (w *Workspace) buildModule(mod *Module, modname string, builtModules map[string]bool) error {

	if builtModules[modname] {
		return nil
	}

	// Check if mod has dependencies that need to build first
	for _, dep := range mod.Depends {
		depMod, ok := w.Modules[dep]
		if !ok {
			return fmt.Errorf("depends on unknown module %s", dep)
		}
		err := w.buildModule(depMod, dep, builtModules)
		if err != nil {
			return fmt.Errorf("failed to build dependency %s: %w", dep, err)
		}
	}

	cmakeBinary := "cmake"
	if w.CMakeBinary != nil {
		cmakeBinary = *w.CMakeBinary
	}

	cMakeConfigureArgs, err := mod.CMakeConfigureArgs(w, modname)
	if err != nil {
		return fmt.Errorf("failed to get cmake configure args: %w", err)
	}

	w.Exec(cmakeBinary, cMakeConfigureArgs)

	buildPath, err := mod.CMakeBuildPath(w, modname)
	if err != nil {
		return fmt.Errorf("failed to get build path: %w", err)
	}
	// Build the module
	buildCmd := []string{"--build", buildPath}

	w.Exec(cmakeBinary, buildCmd)

	builtModules[modname] = true

	return nil
}
