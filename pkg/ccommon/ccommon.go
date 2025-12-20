package ccommon

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Toolchain struct {
	CMakeToolchain map[string]struct {
		CMakeToolchainFile string `yaml:"cmake_toolchain_file"`
	} `yaml:"cmake_toolchain"`
	TargetArch   string `yaml:"target_arch"`
	TargetSystem string `yaml:"target_system"`
}

type BuildParameters struct {
	Toolchain     string
	ToolchainPath string
	BuildType     string
	DryRun        bool
}

type Workspace struct {
	WorkspacePath string

	Targets map[string]*Target `yaml:"targets"`

	CMakeBinary *string `yaml:"cmake_binary"`
	BuildType   string  `yaml:"build_type"`
	CXXVersion  string  `yaml:"cxx_version"`
	Generator   *string `yaml:"generator"`
}

type Target struct {
	Depends                []string `yaml:"depends"`
	ProjectType            string   `yaml:"project_type"`
	CMakePackageName       string   `yaml:"cmake_package_name"`
	FindPackageName        *string  `yaml:"find_package_name"`
	FindPackageRoot        *string  `yaml:"find_package_root"`
	SourceRoot             *string  `yaml:"source_root"`
	ConfigPath             *string  `yaml:"config_path"`
	CMakeAdditionalOptions []string `yaml:"cmake_additional_options"`
}

// CMakeConfigureArgs returns the arguments to pass to cmake when configuring the module
func (m *Target) CMakeConfigureArgs(workspace *Workspace, modname string, bp BuildParameters) ([]string, error) {

	args := []string{}

	src, err := m.CMakeSourcePath(workspace, modname)
	if err != nil {
		return nil, err
	}
	src, _ = filepath.Abs(src)

	args = append(args, "-S")
	args = append(args, src)

	bld, err := m.CMakeBuildPath(workspace, modname, bp)
	if err != nil {
		return nil, err
	}
	bld, _ = filepath.Abs(bld)

	args = append(args, "-B")
	args = append(args, bld)

	args = append(args, fmt.Sprintf("-DCMAKE_BUILD_TYPE=%s", bp.BuildType))

	if bp.ToolchainPath != "" {
		tc, tcPath, err := workspace.LoadToolchain(bp.Toolchain)
		if err == nil {
			// TODO: detect host platform
			hostPlatform := "host-linux-64"
			if tcf, ok := tc.CMakeToolchain[hostPlatform]; ok {
				tcfPath := filepath.Join(tcPath, tcf.CMakeToolchainFile)
				absTcfPath, err := filepath.Abs(tcfPath)
				if err == nil {
					tcfPath = absTcfPath
				}
				args = append(args, fmt.Sprintf("-DCMAKE_TOOLCHAIN_FILE=%s", tcfPath))
			}
		}
	}

	if workspace.Generator != nil {
		args = append(args, fmt.Sprintf("-G%s", *workspace.Generator))
	}

	for _, dep := range m.Depends {
		parts := strings.SplitN(dep, "/", 2)
		targetName := parts[0]

		mod, ok := workspace.Targets[targetName]
		if !ok {
			return nil, fmt.Errorf("depend on unknown target %s", targetName)
		}
		mod_args, err := mod.CMakeDependencyArgs(workspace, targetName, bp)
		if err != nil {
			return nil, err
		}
		args = append(args, mod_args...)
	}

	return args, nil

}

func (m *Target) CMakeSourcePath(workspace *Workspace, defaultRoot string) (string, error) {
	path := m.SourceRoot
	if path == nil {
		return filepath.Join(workspace.WorkspacePath, "sources", defaultRoot), nil
	}

	if filepath.IsAbs(*path) {
		return *path, nil
	}

	return filepath.Join(workspace.WorkspacePath, "sources", *path), nil
}

func (m *Target) CMakeConfigPath(workspace *Workspace, defaultRoot string, bp BuildParameters) (string, error) {

	buildPath, err := m.CMakeBuildPath(workspace, defaultRoot, bp)
	if err != nil {
		return "", err
	}

	if m.ConfigPath != nil {
		return filepath.Join(buildPath, *m.ConfigPath), nil
	}

	return buildPath, nil
}

func (m *Target) CMakeBuildPath(workspace *Workspace, defaultRoot string, bp BuildParameters) (string, error) {
	return filepath.Join(workspace.WorkspacePath, "buildspaces", defaultRoot, bp.Toolchain, bp.BuildType), nil
}

func (m *Target) CMakeExportPath(workspace *Workspace, defaultRoot string, bp BuildParameters) (string, error) {
	return filepath.Join(workspace.WorkspacePath, "exports", defaultRoot, bp.Toolchain, bp.BuildType), nil
}

// CMakeDependencyArgs returns the arguments to pass to cmake when configuring another module that depends on this module
func (m *Target) CMakeDependencyArgs(workspace *Workspace, modname string, bp BuildParameters) ([]string, error) {
	args := []string{}

	packageName := m.CMakePackageName
	if packageName == "" {
		packageName = modname
	}

	if m.FindPackageName != nil {
		packageName = *m.FindPackageName
	}

	if packageName != "" {
		dirname := packageName + "_DIR"

		configPath, err := m.CMakeConfigPath(workspace, modname, bp)
		if err != nil {
			return nil, err
		}
		configPath, err = filepath.Abs(configPath)
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
		sourcePath, _ = filepath.Abs(sourcePath)

		args = append(args, fmt.Sprintf("-D%s=%s", dirname, sourcePath))
	}

	return args, nil
}

func (w *Workspace) Load(path string) error {
	w.WorkspacePath = path
	// Load the configuration from the file
	yamlFile, err := os.ReadFile(filepath.Join(path, "cbuild_workspace.yml"))
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	err = yaml.Unmarshal(yamlFile, w)

	if err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	return nil
}

func (w *Workspace) Save() error {
	yamlFile, err := yaml.Marshal(w)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	err = os.WriteFile(filepath.Join(w.WorkspacePath, "cbuild_workspace.yml"), yamlFile, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func (w *Workspace) LoadToolchain(toolchainName string) (*Toolchain, string, error) {
	toolchainDir := filepath.Join(w.WorkspacePath, "toolchains", toolchainName)
	toolchainFile := filepath.Join(toolchainDir, "toolchain.yml")

	yamlFile, err := ioutil.ReadFile(toolchainFile)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read toolchain file: %w", err)
	}

	tc := &Toolchain{}
	err = yaml.Unmarshal(yamlFile, tc)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse toolchain file: %w", err)
	}

	return tc, toolchainDir, nil
}

func (w *Workspace) Build(bp BuildParameters) error {
	var builtModules = make(map[string]bool)

	for name, mod := range w.Targets {
		err := w.buildModule(mod, name, builtModules, bp)
		if err != nil {
			return fmt.Errorf("failed to build module %s: %w", name, err)
		}
	}

	return nil
}

func (w *Workspace) BuildTarget(targetName string, bp BuildParameters) error {
	var builtModules = make(map[string]bool)

	mod, ok := w.Targets[targetName]
	if !ok {
		return fmt.Errorf("unknown target %s", targetName)
	}

	return w.buildModule(mod, targetName, builtModules, bp)
}

func (w *Workspace) Exec(command string, args []string, dryRun bool) error {
	fmt.Printf("Executing: %s", command)
	for _, arg := range args {
		fmt.Printf(" %s", arg)
	}
	fmt.Println()

	if dryRun {
		return nil
	}

	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (w *Workspace) buildModule(mod *Target, modname string, builtModules map[string]bool, bp BuildParameters) error {

	if builtModules[modname] {
		return nil
	}

	// Check if mod has dependencies that need to build first
	for _, dep := range mod.Depends {
		parts := strings.SplitN(dep, "/", 2)
		targetName := parts[0]
		depMod, ok := w.Targets[targetName]
		if !ok {
			return fmt.Errorf("depends on unknown target %s", targetName)
		}
		err := w.buildModule(depMod, targetName, builtModules, bp)
		if err != nil {
			return fmt.Errorf("failed to build dependency %s: %w", targetName, err)
		}
	}

	cmakeBinary := "cmake"
	if w.CMakeBinary != nil {
		cmakeBinary = *w.CMakeBinary
	}

	if mod.ProjectType != "" && !strings.EqualFold(mod.ProjectType, "CMake") {
		return fmt.Errorf("unsupported project type: %s", mod.ProjectType)
	}

	cMakeConfigureArgs, err := mod.CMakeConfigureArgs(w, modname, bp)
	if err != nil {
		return fmt.Errorf("failed to get cmake configure args: %w", err)
	}

	err = w.Exec(cmakeBinary, cMakeConfigureArgs, bp.DryRun)
	if err != nil {
		return fmt.Errorf("failed to configure module %s: %w", modname, err)
	}

	buildPath, err := mod.CMakeBuildPath(w, modname, bp)
	if err != nil {
		return fmt.Errorf("failed to get build path: %w", err)
	}
	buildPath, _ = filepath.Abs(buildPath)

	// Build the module
	buildCmd := []string{"--build", buildPath}

	err = w.Exec(cmakeBinary, buildCmd, bp.DryRun)
	if err != nil {
		return fmt.Errorf("failed to build module %s: %w", modname, err)
	}

	builtModules[modname] = true

	return nil
}
