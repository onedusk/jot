#!/bin/bash

# Install jotdoc binary globally (renamed to avoid conflict with BSD jot)

echo "🔧 Installing jotdoc globally..."

# Check if jot binary exists
if [ ! -f "./jot" ]; then
    echo "❌ jot binary not found. Building first..."
    go build -o jot ./cmd/jot
    if [ $? -ne 0 ]; then
        echo "❌ Failed to build jot"
        exit 1
    fi
    echo "✅ Built jot binary"
fi

# First, remove the mistakenly installed jot if it's ours (larger than 1MB)
if [ -f "/usr/local/bin/jot" ]; then
    SIZE=$(stat -f%z /usr/local/bin/jot 2>/dev/null || stat -c%s /usr/local/bin/jot 2>/dev/null)
    if [ "$SIZE" -gt 1000000 ]; then
        echo "🔄 Removing previously installed jot (was our doc generator)..."
        sudo rm /usr/local/bin/jot
    fi
fi

# Install as jotdoc to avoid conflict with BSD jot
echo "📦 Installing to /usr/local/bin/jotdoc (requires sudo)..."
sudo cp ./jot /usr/local/bin/jotdoc

if [ $? -eq 0 ]; then
    echo "✅ Successfully installed!"
    echo ""
    echo "Verifying installation..."
    which jotdoc
    jotdoc --version 2>/dev/null || jotdoc --help | head -1
    echo ""
    echo "🎉 You can now use 'jotdoc' from anywhere!"
    echo ""
    echo "Quick start:"
    echo "  jotdoc init         # Initialize a new documentation project"
    echo "  jotdoc build        # Build documentation"
    echo "  jotdoc serve        # Serve documentation locally"
    echo "  jotdoc export       # Export to various formats"
else
    echo "❌ Installation failed. You can try manually:"
    echo "  sudo cp ./jot /usr/local/bin/jotdoc"
fi