#!/bin/bash
#
# agmd release script
# Usage: ./scripts/release.sh v0.2.0
#

set -e

VERSION="$1"

if [ -z "$VERSION" ]; then
    echo "Usage: ./scripts/release.sh <version>"
    echo "Example: ./scripts/release.sh v0.2.0"
    exit 1
fi

# Validate version format
if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Error: Version must be in format v0.0.0"
    exit 1
fi

echo "==> Building agmd $VERSION"

# Clean dist directory
rm -rf dist
mkdir -p dist

# Platforms to build
PLATFORMS=(
    "darwin/arm64"
    "darwin/amd64"
    "linux/amd64"
    "linux/arm64"
)

# Build for each platform
for PLATFORM in "${PLATFORMS[@]}"; do
    GOOS="${PLATFORM%/*}"
    GOARCH="${PLATFORM#*/}"
    OUTPUT="dist/agmd_${GOOS}_${GOARCH}"

    echo "    Building $GOOS/$GOARCH..."
    GOOS=$GOOS GOARCH=$GOARCH go build -ldflags "-s -w -X main.version=$VERSION" -o "$OUTPUT" .
done

# Create tarballs
echo "==> Creating tarballs"
for PLATFORM in "${PLATFORMS[@]}"; do
    GOOS="${PLATFORM%/*}"
    GOARCH="${PLATFORM#*/}"
    BINARY="dist/agmd_${GOOS}_${GOARCH}"
    TARBALL="dist/agmd_${GOOS}_${GOARCH}.tar.gz"

    echo "    Creating $TARBALL..."
    cp "$BINARY" dist/agmd
    tar -czvf "$TARBALL" -C dist agmd > /dev/null
    rm dist/agmd
done

echo "==> Built artifacts:"
ls -la dist/*.tar.gz

# Check if release exists
echo ""
echo "==> Checking GitHub release $VERSION..."
if gh release view "$VERSION" > /dev/null 2>&1; then
    echo "    Release $VERSION exists, uploading assets..."
    gh release upload "$VERSION" dist/*.tar.gz --clobber
else
    echo "    Creating release $VERSION..."
    gh release create "$VERSION" dist/*.tar.gz \
        --title "$VERSION" \
        --notes "Release $VERSION" \
        --draft
    echo ""
    echo "    Draft release created. Edit and publish at:"
    echo "    https://github.com/GluonGrid/agmd/releases"
fi

echo ""
echo "==> Done! Release $VERSION ready."
echo ""
echo "Test installation with:"
echo "    curl -fsSL https://gluongrid.dev/agmd/install.sh | bash"
