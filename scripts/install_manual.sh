#!/bin/bash

# Manual install script for jotdoc
# Run this in your terminal with: ./install_manual.sh

echo "🔧 Manual installation of jotdoc"
echo ""
echo "This script will:"
echo "1. Remove the incorrectly installed jot from /usr/local/bin"
echo "2. Install your doc generator as 'jotdoc'"
echo ""
echo "Please run these commands manually:"
echo ""
echo "# 1. Remove the large jot file (our doc generator)"
echo "sudo rm /usr/local/bin/jot"
echo ""
echo "# 2. Install as jotdoc"
echo "sudo cp ./jot /usr/local/bin/jotdoc"
echo ""
echo "# 3. Verify installation"
echo "which jotdoc"
echo "jotdoc --help"
echo ""
echo "# 4. Test the original jot still works"
echo "which jot  # Should show /usr/bin/jot"
echo "jot 5      # Should print 1 2 3 4 5"