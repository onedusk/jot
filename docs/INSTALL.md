# Installing Jot

There are several ways to install Jot on your system. Choose the method that works best for your setup.

## Prerequisites

- Go 1.19 or later (for building from source)
- Git (for cloning the repository)

## Installation Methods

### Method 1: Using Go Install (Recommended)

The simplest way to install jot is using `go install`:

```bash
go install github.com/onedusk/jot/cmd/jot@latest
```

This installs jot to `$GOPATH/bin`. Make sure this directory is in your PATH:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

### Method 2: Using the Install Script

Clone the repository and run the install script:

```bash
git clone https://github.com/onedusk/jot.git
cd jot
./install.sh
```

The script will:
- Build the binary
- Install it to `/usr/local/bin` (may require sudo)
- Verify the installation

### Method 3: Using Make

If you have Make installed:

```bash
git clone https://github.com/onedusk/jot.git
cd jot
make install
```

This builds and installs jot to `/usr/local/bin` (requires sudo).

### Method 4: Manual Installation

1. Clone and build:
```bash
git clone https://github.com/onedusk/jot.git
cd jot
go build -o jot ./cmd/jot
```

2. Move to a directory in your PATH:
```bash
# Option A: User-specific installation (no sudo required)
mkdir -p ~/bin
mv jot ~/bin/
export PATH="$PATH:~/bin"  # Add to your shell profile

# Option B: System-wide installation (requires sudo)
sudo mv jot /usr/local/bin/
```

### Method 5: Download Pre-built Binary

Download the appropriate binary for your platform from the [releases page](https://github.com/onedusk/jot/releases):

```bash
# Example for macOS ARM64
curl -L https://github.com/onedusk/jot/releases/latest/download/jot-darwin-arm64.tar.gz | tar xz
sudo mv jot /usr/local/bin/
```

## Verifying Installation

After installation, verify that jot is working:

```bash
jot --version
jot --help
```

## Updating

To update jot to the latest version:

### If installed with go install:
```bash
go install github.com/onedusk/jot/cmd/jot@latest
```

### If installed from source:
```bash
cd jot
git pull
make clean install
```

## Uninstalling

### If installed to /usr/local/bin:
```bash
sudo rm /usr/local/bin/jot
```

### If installed with go install:
```bash
rm $(go env GOPATH)/bin/jot
```

### Using Make:
```bash
cd jot
make uninstall
```

## Troubleshooting

### Command not found

If you get "command not found" after installation, ensure the installation directory is in your PATH:

```bash
# For go install method
export PATH="$PATH:$(go env GOPATH)/bin"

# For /usr/local/bin installation
export PATH="$PATH:/usr/local/bin"

# For user bin directory
export PATH="$PATH:~/bin"
```

Add the appropriate export line to your shell profile (`~/.bashrc`, `~/.zshrc`, etc.) to make it permanent.

### Permission denied

If you get permission errors during installation:
- Use `sudo` for system-wide installation
- Choose a user-specific installation method that doesn't require sudo
- Use the `go install` method

### Build errors

Ensure you have:
- Go 1.19 or later: `go version`
- All dependencies: `go mod download`

## Platform-Specific Notes

### macOS

If you have Homebrew installed, you may want to ensure `/usr/local/bin` comes before system directories in your PATH to avoid conflicts with the BSD `jot` utility:

```bash
export PATH="/usr/local/bin:$PATH"
```

### Windows

On Windows, use the Windows binary from the releases page or build with:

```bash
go build -o jot.exe ./cmd/jot
```

Then add the directory containing `jot.exe` to your system PATH.

### Linux

Most Linux distributions include `/usr/local/bin` in the default PATH. If not, add it to your shell profile.