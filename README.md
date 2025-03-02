# OmniHook
OmniHook is a global Git hook manager that allows you to easily install and manage pre-commit, pre-push, and other Git hooks across multiple repositories. It provides a centralized way to enforce best practices and security policies.

![OmniHook Logo](static/omnihook.png)

## Features
- Install Git hooks from a remote Git repository or local file.
- Supports inline scripts as well as external script paths.
- Works with multiple Git repositories globally.
- Modular structure with YAML-based hook definitions.
- Runs all the hooks in parallel

## Installation

### Using Go
```sh
go install github.com/vjayajv/omnihook@latest
```

### Manual Installation
1. Download the latest release from [GitHub Releases](https://github.com/vjayajv/omnihook/releases).
2. Extract the binary and place it in a directory in your `PATH`.
3. Run `omnihook --help` to verify the installation.

## Configuration
You need to run omnihook configure first to initialize the tool. This setsup omnihook config file, hooks directory, setups global git hooks path and a pre-commit script that calls omnihook run:
```sh
omnihook configure
```

## Usage

### Install a Hook from a Git Repository
```sh
omnihook install --url https://github.com/example/hooks-repo.git
```

### Install Hooks from OmniHook Test Repository
```sh
omnihook install --url https://github.com/vjayajv/omnihook-test-hooks.git
```

### Install a Hook from a Local File
```sh
omnihook install --file /path/to/hook.yml
```

### Enable a Hook
```sh
omnihook enable --id <hook-id> --type <hook-type>
```

### Disable a Hook
```sh
omnihook disable --id <hook-id> --type <hook-type>
```

### Update Installed Hooks
```sh
omnihook update
```

### List Installed Hooks
```sh
omnihook list --all | --type <hook-type>
```

### Uninstall a Hook
```sh
omnihook uninstall --all | --type <hook-type> | --id <hook-id> --type <hook-type>
```

### Run Hooks Manually
```sh
omnihook run --type <hook-type> 
```

## Example Hook Configuration (`hook.yml`)
```yaml
id: pre-commit-linter
name: Pre-Commit Linter
description: Runs a linter before committing changes.
script: |
  #!/bin/sh
  echo "Running linter..."
  eslint .
hookType: "pre-commit"
```

## Contributing
Contributions are welcome! Feel free to open issues or submit pull requests.

## License
This project is licensed under the MIT License. See the [MIT License](https://opensource.org/licenses/MIT) for details. See the [LICENSE](LICENSE) file for details.
