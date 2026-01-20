# cbuild & csetup

`cbuild` is a tool for building distributions by managing multiple projects simultaneously within a workspace. `csetup` is a companion tool used to initialize and manage these workspaces.

## Workspace Structure

A `cbuild` workspace typically follows this structure:

```
workspace/
  cbuild_workspace.yml        # Main workspace configuration
  toolchains/                 # Toolchain definitions
    <toolchain_name>/
      toolchain.yml
      <toolchain_file>.cmake
  sources/                    # Source code for projects
    <sourcename>/
  buildspaces/                # Build artifacts (temporary)
    <sourcename>/<toolchain>/<config>/
  exports/                    # Build outputs
    <sourcename>/<toolchain>/<config>/
```

## cbuild

`cbuild` is the main build tool. It uses the information in `cbuild_workspace.yml` to build projects.

### Commands

- **`build [target]`** (default): Build the specified target(s). If no target is specified, it builds all targets.
- **`clean [target]`**: Remove build artifacts from `buildspaces` for the specified target(s).
- **`build-deps <target>`**: Build only the dependencies for a specific target.

### Global Flags

- `-w, --workspace <path>`: Path to the workspace directory (default: current directory or nearest parent with `cbuild_workspace.yml`).
- `-d, --dry-run`: Show commands without executing them.
- `-h, --help`: Show help message.

### Common Flags (for build/clean/build-deps)

- `-c, --config <configs>`: Build configurations to use (e.g., `Debug,Release`), comma-separated.
- `-T, --toolchain <toolchain>`: Specific toolchain to use (default: `all`).
- `-t, --target <targets>`: Specific target(s) to build/clean, comma-separated.

## csetup

`csetup` is used for managing the workspace, including adding/removing sources, targets, and dependencies.

### Commands

- **`init [path] [--reinit]`**: Initialize a new workspace at the given path.
- **`git-clone <repo_url> <dest_name> [--download-deps] [--submodule] [--no-setup]`**: Clone a git repository into the `sources` directory and add it to the workspace.
- **`declare-git-source <repo_url> <dest_name>`**: Add git source information to the workspace without downloading.
- **`declare-local-source <local_path> <dest_name>`**: Add local source information to the workspace.
- **`download [source] [--download-deps] [--submodule] [--no-setup]`**: Download missing sources.
- **`load-defaults <source>`**: Load default configuration for a source from its `csetup.yml`.
- **`add-dependency <target> <dependency>`**: Add a dependency to a target.
- **`remove-dependency <target> <dependency>`**: Remove a dependency from a target.
- **`remove-source <source> [-X, --delete]`**: Remove a source from the workspace.
- **`remove-target <target>`**: Remove a target from the workspace.
- **`remove-project <source> [-X, --delete]`**: Remove a source and all its associated targets from the workspace.
- **`new-target <target> <source> [--project-type <type>] [--cmake-package-name <name>] [--overwrite]`**: Add a new target to the workspace.
- **`set-cxx-version <version> [target]`**: Set the C++ version for a target or the whole workspace.
- **`enable-staging <target>`**: Enable staging for a target. Staged targets are built against the installed outputs, instead of a build tree.
- **`disable-staging <target>`**: Disable staging for a target.
- **`list-sources`**: List all sources in the workspace.
- **`drop-files <source>`**: Delete local source files without removing them from configuration.
- **`get-args <target> [-c <config>] [-T <toolchain>]`**: Get the build arguments that would be passed to the build system (e.g., CMake).
- **`detect-toolchains`**: Automatically detect system toolchains and create definitions in `toolchains/`.
- **`add-config [-c <config>] [-T <toolchain>]`**: Add a build configuration.
- **`remove-config -c <config>`**: Remove a build configuration.

## YAML Formats

### `cbuild_workspace.yml`

Located at the root of the workspace.

```yaml
cmake_binary: "/usr/bin/cmake"    # Optional: Path to cmake binary
cxx_version: "20"                 # Default C++ standard for the workspace
configurations: ["Debug", "Release"] # Default build configurations

sources:
  <sourcename>:
    git:
      repository: "https://github.com/user/repo.git"
      revision: "main"            # Optional: Branch, tag, or commit hash
    # OR
    local: "/path/to/local/source"

targets:
  <targetname>:
    source: "<sourcename>"        # Optional: Defaults to target name
    project_type: "cmake"         # Currently only "cmake" is supported
    depends: ["dep1", "dep2/sub"] # List of dependencies
    cmake_package_name: "Name"    # Optional: For CMake's find_package()
    cxx_standard: "17"            # Optional: Override workspace C++ version
    staged: true                  # Optional: Use staging for this target
    extra_cmake_configure_args: ["-DFOO=BAR"] # Optional: Extra args for CMake
```

### Toolchain `toolchain.yml`

Located in `toolchains/<toolchain_name>/toolchain.yml`.

```yaml
target_arch: "x64"                # Target processor architecture
target_system: "linux"            # Target operating system
cmake_toolchain:
  <host_key>:                     # e.g., host-linux-x64
    cmake_toolchain_file: "path/to/toolchain.cmake"
    # OR
    generate:                     # Automatically generate a toolchain file
      c_compiler: "/usr/bin/gcc"
      cxx_compiler: "/usr/bin/g++"
      linker: "/usr/bin/ld"       # Optional
```

The `<host_key>` typically follows the format `host-<os>-<arch>` (e.g., `host-linux-x64`).



