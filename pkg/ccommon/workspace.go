package ccommon

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gitlab.com/rpnx/cbuild-go/pkg/cmake"
	"gitlab.com/rpnx/cbuild-go/pkg/host"
	"gitlab.com/rpnx/cbuild-go/pkg/system"

	"gopkg.in/yaml.v3"
)

type WorkspaceContext struct {
	Config        WorkspaceConfig
	WorkspacePath string
	DownloadDeps  bool
}

type WorkspaceConfig struct {
	Targets map[string]*TargetConfiguration `yaml:"targets"`

	CMakeBinary    *string  `yaml:"cmake_binary"`
	CXXVersion     string   `yaml:"cxx_version"`
	Configurations []string `yaml:"configurations"`
}

func (w *WorkspaceContext) Load(path string) error {
	w.WorkspacePath = path
	// Load the configuration from the file
	yamlFile, err := os.ReadFile(filepath.Join(path, "cbuild_workspace.yml"))
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	w.Config = WorkspaceConfig{}

	err = yaml.Unmarshal(yamlFile, &w.Config)

	if err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	if len(w.Config.Configurations) == 0 {
		w.Config.Configurations = []string{"Debug", "Release"}
	}

	return nil
}

func (w *WorkspaceContext) GetTarget(name string) (*TargetContext, error) {
	targetConfig, ok := w.Config.Targets[name]
	if !ok {
		return nil, fmt.Errorf("target %s not found in workspace", name)
	}

	return &TargetContext{
		Name:   name,
		Config: *targetConfig,
	}, nil
}

func (w *WorkspaceContext) Save() error {
	yamlFile, err := yaml.Marshal(w.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	err = os.WriteFile(filepath.Join(w.WorkspacePath, "cbuild_workspace.yml"), yamlFile, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func (w *WorkspaceContext) GenerateToolchainFile(opts *CMakeGenerateToolchainFileOptions, systemName system.Platform, systemProcessor system.Processor, targetPath string) error {
	return cmake.GenerateToolchainFile(context.Background(), cmake.GenerateToolchainFileOptions{
		CCompiler:       opts.CCompiler,
		CXXCompiler:     opts.CXXCompiler,
		Linker:          opts.Linker,
		ExtraCXXFlags:   opts.ExtraCXXFlags,
		SystemPlatform:  systemName,
		SystemProcessor: systemProcessor,
		WorkspaceDir:    w.WorkspacePath,
		OutputFile:      targetPath,
	})
}

func (w *WorkspaceContext) LoadToolchain(toolchainName string) (*Toolchain, string, error) {
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

func (w *WorkspaceContext) ToolchainFilePath(modConfig *TargetConfiguration, bp TargetBuildParameters) (string, error) {
	tc, tcPath, err := w.LoadToolchain(bp.Toolchain)
	if err != nil {
		return "", fmt.Errorf("failed to load toolchain: %w", err)
	}

	hostPlatform := fmt.Sprintf("host-%s-%s", host.DetectHostPlatform().StringLower(), host.DetectHostProcessor().StringLower())
	if tcf, ok := tc.CMakeToolchain[hostPlatform]; ok {
		var tcfPath string
		if tcf.Generate != nil {
			tcfPath = filepath.Join(w.WorkspacePath, "buildspaces", bp.Toolchain, "generated_toolchain.cmake")
		} else {
			tcfPath = filepath.Join(tcPath, tcf.CMakeToolchainFile)
		}

		absTcfPath, err := filepath.Abs(tcfPath)
		if err != nil {
			return "", err
		}
		tcfPath = absTcfPath
		return tcfPath, nil
	}
	return "", nil
}

func (w *WorkspaceContext) Prebuild(bp TargetBuildParameters) (string, error) {
	tc, _, err := w.LoadToolchain(bp.Toolchain)
	if err != nil {
		return "", fmt.Errorf("failed to load toolchain: %w", err)
	}

	hostPlatform := fmt.Sprintf("host-%s-%s", host.DetectHostPlatform().StringLower(), host.DetectHostProcessor().StringLower())
	if tcf, ok := tc.CMakeToolchain[hostPlatform]; ok {
		tcfPath, err := w.ToolchainFilePath(nil, bp)
		if err != nil {
			return "", err
		}
		if tcf.Generate != nil {
			err := w.GenerateToolchainFile(tcf.Generate, tc.TargetSystem, tc.TargetArch, tcfPath)
			if err != nil {
				return "", fmt.Errorf("failed to generate toolchain file: %w", err)
			}
		}
		return tcfPath, nil
	}
	return "", nil
}

func (w *WorkspaceContext) Build(bp TargetBuildParameters) error {
	_, err := w.Prebuild(bp)
	if err != nil {
		return err
	}

	var builtModules = make(map[string]bool)

	for name := range w.Config.Targets {
		mod, err := w.GetTarget(name)
		if err != nil {
			return err
		}
		err = w.buildModule(mod, name, builtModules, bp)
		if err != nil {
			return fmt.Errorf("failed to build module %s: %w", name, err)
		}
	}

	return nil
}

func (w *WorkspaceContext) BuildTarget(targetName string, bp TargetBuildParameters) error {
	_, err := w.Prebuild(bp)
	if err != nil {
		return err
	}

	var builtModules = make(map[string]bool)

	mod, err := w.GetTarget(targetName)
	if err != nil {
		return err
	}

	return w.buildModule(mod, targetName, builtModules, bp)
}

func (w *WorkspaceContext) BuildDependencies(targetName string, bp TargetBuildParameters) error {
	_, err := w.Prebuild(bp)
	if err != nil {
		return err
	}

	var builtModules = make(map[string]bool)

	mod, err := w.GetTarget(targetName)
	if err != nil {
		return err
	}

	for _, dep := range mod.Config.Depends {
		parts := strings.SplitN(dep, "/", 2)
		depTargetName := parts[0]
		depMod, err := w.GetTarget(depTargetName)
		if err != nil {
			return err
		}
		err = w.buildModule(depMod, depTargetName, builtModules, bp)
		if err != nil {
			return fmt.Errorf("failed to build dependency %s: %w", depTargetName, err)
		}
	}

	return nil
}

func (w *WorkspaceContext) CleanTarget(targetName string, bp TargetBuildParameters) error {
	mod, err := w.GetTarget(targetName)
	if err != nil {
		return err
	}

	buildPath, err := mod.CMakeBuildPath(w, bp)
	if err != nil {
		return err
	}

	if bp.DryRun {
		fmt.Printf("dry-run: would delete build path %s\n", buildPath)
		return nil
	}
	return os.RemoveAll(buildPath)
}

func (w *WorkspaceContext) Exec(command string, args []string, dryRun bool) error {
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

func (w *WorkspaceContext) buildModule(mod *TargetContext, modname string, builtModules map[string]bool, bp TargetBuildParameters) error {

	if builtModules[modname] {
		return nil
	}

	for _, dep := range mod.Config.Depends {
		parts := strings.SplitN(dep, "/", 2)
		targetName := parts[0]
		depMod, err := w.GetTarget(targetName)
		if err != nil {
			return err
		}
		err = w.buildModule(depMod, targetName, builtModules, bp)
		if err != nil {
			return fmt.Errorf("failed to build dependency %s: %w", targetName, err)
		}
	}

	cmakeBinary := "cmake"
	if w.Config.CMakeBinary != nil {
		cmakeBinary = *w.Config.CMakeBinary
	}

	if mod.Config.ProjectType != "" && !strings.EqualFold(mod.Config.ProjectType, "CMake") {
		return fmt.Errorf("unsupported project type: %s", mod.Config.ProjectType)
	}

	cMakeConfigureArgs, err := mod.CMakeConfigureArgs(context.Background(), w, bp)
	if err != nil {
		return fmt.Errorf("failed to get cmake configure args: %w", err)
	}

	err = w.Exec(cmakeBinary, cMakeConfigureArgs, bp.DryRun)
	if err != nil {
		return fmt.Errorf("failed to configure module %s: %w", modname, err)
	}

	buildPath, err := mod.CMakeBuildPath(w, bp)
	if err != nil {
		return fmt.Errorf("failed to get build path: %w", err)
	}
	buildPath, err = filepath.Abs(buildPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute build path: %w", err)
	}

	// Build the module
	buildCmd := []string{"--build", buildPath, "--config", bp.BuildType}

	err = w.Exec(cmakeBinary, buildCmd, bp.DryRun)
	if err != nil {
		return fmt.Errorf("failed to build module %s: %w", modname, err)
	}

	if mod.Config.Staged != nil && *mod.Config.Staged {
		stagingPath, err := mod.CMakeStagingPath(w, bp)
		if err != nil {
			return fmt.Errorf("failed to get staging path: %w", err)
		}
		stagingPath, err = filepath.Abs(stagingPath)
		if err != nil {
			return fmt.Errorf("failed to get absolute staging path: %w", err)
		}

		installCmd := []string{"--install", buildPath, "--prefix", stagingPath, "--config", bp.BuildType}
		err = w.Exec(cmakeBinary, installCmd, bp.DryRun)
		if err != nil {
			return fmt.Errorf("failed to install module %s to staging: %w", modname, err)
		}
	}

	builtModules[modname] = true

	return nil
}

func (w *WorkspaceContext) ProcessCSetupFile(targetName string) error {
	targetConfig, ok := w.Config.Targets[targetName]
	if !ok {
		return fmt.Errorf("target %s not found in workspace", targetName)
	}

	target := &TargetContext{
		Name:   targetName,
		Config: *targetConfig,
	}

	sourcePath, err := target.CMakeSourcePath(w)
	if err != nil {
		return fmt.Errorf("failed to get source path for target %s: %w", targetName, err)
	}

	csetupFiles := []string{"csetup.yml", "csetuplists.yml", "CSetup.yml", "CSetupLists.yml"}
	var csetupFile string
	for _, f := range csetupFiles {
		path := filepath.Join(sourcePath, f)
		if _, err := os.Stat(path); err == nil {
			csetupFile = path
			break
		}
	}

	if csetupFile == "" {
		return nil // No csetup file to process
	}

	data, err := os.ReadFile(csetupFile)
	if err != nil {
		return fmt.Errorf("failed to read csetup file %s: %w", csetupFile, err)
	}

	var csetup CSetupLists
	err = yaml.Unmarshal(data, &csetup)
	if err != nil {
		return fmt.Errorf("failed to parse csetup file %s: %w", csetupFile, err)
	}

	reader := bufio.NewReader(os.Stdin)

	// Replace all fields of TargetConfiguration with the DefaultConfiguration from the CSetup file,
	// *except* ExternalSourceOverride if it is present, making sure to nil any existing fields not present
	// in the DefaultConfiguration, excluding ExternalSourceOverride.
	externalOverride := target.Config.ExternalSourceOverride
	target.Config = csetup.DefaultConfig
	target.Config.ExternalSourceOverride = externalOverride
	w.Config.Targets[targetName] = &target.Config

	// Process Suggested Dependencies
	for depName, sdep := range csetup.SuggestedSources {
		if err := sdep.ValidateWeb(); err != nil {
			return fmt.Errorf("invalid suggested source for dependency %s: %w", depName, err)
		}

		if _, exists := w.Config.Targets[depName]; !exists {
			autoDownload := w.DownloadDeps
			if !autoDownload {
				fmt.Printf("Dependency '%s' is not present in sources, module '%s' suggests getting it from '%s', download it? [Y/n] ", depName, targetName, sdep.From())
				response, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("error reading input: %w", err)
				}
				response = strings.ToLower(strings.TrimSpace(response))
				if response == "" || response == "y" || response == "yes" {
					autoDownload = true
				}
			}

			if autoDownload {
				err := w.Get(depName, sdep)
				if err != nil {
					return fmt.Errorf("failed to download '%s': %w", depName, err)
				}

				// Add to workspace targets
				w.Config.Targets[depName] = &TargetConfiguration{}
				fmt.Printf("Added target '%s' to workspace.\n", depName)

				// Also add it as a dependency to the current target
				target.Config.Depends = append(target.Config.Depends, depName)
				fmt.Printf("Added dependency '%s' to target '%s'.\n", depName, targetName)

				// Recursively process the new target's csetup file
				err = w.ProcessCSetupFile(depName)
				if err != nil {
					return fmt.Errorf("error processing csetup file for %s: %w", depName, err)
				}
			}
		}
	}

	return w.Save()
}

func (ws *WorkspaceContext) ListToolchains() ([]string, error) {
	toolchainsDir := filepath.Join(ws.WorkspacePath, "toolchains")
	entries, err := os.ReadDir(toolchainsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read toolchains directory: %w", err)
	}

	var toolchains []string
	for _, entry := range entries {
		if entry.IsDir() {
			toolchains = append(toolchains, entry.Name())
		}
	}

	return toolchains, nil
}

func (ws *WorkspaceContext) ListTargets() []string {
	var targets []string
	for k, _ := range ws.Config.Targets {
		targets = append(targets, k)
	}
	return targets
}
