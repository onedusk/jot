#!/bin/bash

# Jot Release Script
# This script builds and packages Jot for multiple platforms

set -e

VERSION=$(cat VERSION)
BUILD_DIR="build"
DIST_DIR="dist-release"

echo "Building Jot v${VERSION} releases..."

# Clean previous builds
rm -rf ${BUILD_DIR} ${DIST_DIR}
mkdir -p ${BUILD_DIR} ${DIST_DIR}

# Platforms to build
PLATFORMS=(
    "darwin amd64"
    "darwin arm64"
    "linux amd64"
    "linux arm64"
    "linux 386"
    "windows amd64"
    "windows 386"
)

# Build for each platform
for platform in "${PLATFORMS[@]}"; do
    read -r GOOS GOARCH <<< "$platform"
    output="${BUILD_DIR}/jot-${GOOS}-${GOARCH}"
    
    echo "Building for ${GOOS}/${GOARCH}..."
    
    GOOS=${GOOS} GOARCH=${GOARCH} go build \
        -ldflags "-X main.version=${VERSION} -s -w" \
        -trimpath \
        -o "${output}" \
        ./cmd/jot
    
    # Add .exe extension for Windows
    if [ "${GOOS}" = "windows" ]; then
        mv "${output}" "${output}.exe"
        output="${output}.exe"
    fi
    
    # Create archive
    archive_name="jot-v${VERSION}-${GOOS}-${GOARCH}"
    
    if [ "${GOOS}" = "windows" ]; then
        # Create zip for Windows
        (cd ${BUILD_DIR} && zip -q "../${DIST_DIR}/${archive_name}.zip" "$(basename ${output})")
    else
        # Create tar.gz for Unix
        (cd ${BUILD_DIR} && tar czf "../${DIST_DIR}/${archive_name}.tar.gz" "$(basename ${output})")
    fi
done

# Create checksums
echo "Generating checksums..."
(cd ${DIST_DIR} && shasum -a 256 * > checksums.txt)

# Count releases
release_count=$(ls -1 ${DIST_DIR}/*.{tar.gz,zip} 2>/dev/null | wc -l)

echo ""
echo " Release build complete!"
echo "   Version: v${VERSION}"
echo "   Packages: ${release_count}"
echo "   Location: ${DIST_DIR}/"
echo ""
echo "To create a GitHub release:"
echo "  1. git tag v${VERSION}"
echo "  2. git push origin v${VERSION}"
echo "  3. Upload files from ${DIST_DIR}/ to the GitHub release"