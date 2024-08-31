#!/bin/bash

# Exit on any error
set -e

# Function to check if a command exists
command_exists() {
  command -v "$1" >/dev/null 2>&1
}

# Check for required tools
for cmd in go wails docker; do
  if ! command_exists $cmd; then
    echo "Error: $cmd is not installed. Please install it and try again."
    exit 1
  fi
done

# Set variables
APP_NAME="ServlociTerm"
VERSION="1.0.0"
OUTPUT_DIR="./build"

# Create output directory
mkdir -p $OUTPUT_DIR

# Function to find and move the built file
move_built_file() {
  local source_pattern=$1
  local destination=$2
  local found_file=$(find . -name "$source_pattern" -type f -print -quit)

  if [ -n "$found_file" ]; then
    mv "$found_file" "$destination"
    echo "Moved $found_file to $destination"
  else
    echo "Error: Could not find file matching $source_pattern"
    return 1
  fi
}

# Build for Windows
echo "Building for Windows..."
wails build -platform windows/amd64
move_built_file "${APP_NAME}.exe" "$OUTPUT_DIR/${APP_NAME}-${VERSION}-windows-amd64.exe"

# Create Windows installer
echo "Creating Windows installer..."
# You'll need to create an InnoSetup script (ServlociTerm.iss) for this step
# iscc.exe ServlociTerm.iss

# Build for macOS
echo "Building for macOS..."
wails build -platform darwin/amd64
move_built_file "${APP_NAME}.app" "$OUTPUT_DIR/${APP_NAME}-${VERSION}-macos.app"

# Create macOS package
echo "Creating macOS package..."
pkgbuild --root "$OUTPUT_DIR/${APP_NAME}-${VERSION}-macos.app" --identifier com.servloci.${APP_NAME} --version $VERSION "$OUTPUT_DIR/${APP_NAME}-${VERSION}-macos.pkg"

# Build for Linux
echo "Building for Linux..."
wails build -platform linux/amd64
move_built_file "${APP_NAME}" "$OUTPUT_DIR/${APP_NAME}-${VERSION}-linux-amd64"

# Create AppImage
echo "Creating AppImage..."
# You'll need to set up the AppDir structure for this step
# linuxdeploy --appdir AppDir --executable $OUTPUT_DIR/${APP_NAME}-${VERSION}-linux-amd64 --desktop-file=ServlociTerm.desktop --icon-file=ServlociTerm.png --output appimage

# Build Ubuntu DEB package
echo "Building Ubuntu DEB package..."
docker run --rm -v "$(pwd):/app" -w /app ubuntu:latest bash -c "
  apt-get update && apt-get install -y dpkg-dev
  mkdir -p /app/deb/DEBIAN /app/deb/usr/bin
  cp $OUTPUT_DIR/${APP_NAME}-${VERSION}-linux-amd64 /app/deb/usr/bin/${APP_NAME}
  echo 'Package: ${APP_NAME}
Version: ${VERSION}
Architecture: amd64
Maintainer: ServlociTerm Team <support@servloci.com>
Description: Easy to use terminal emulator for cloud VM, developer use cases
' > /app/deb/DEBIAN/control
  dpkg-deb --build /app/deb $OUTPUT_DIR/${APP_NAME}-${VERSION}-ubuntu-amd64.deb
"
echo "Ubuntu DEB package created: $OUTPUT_DIR/${APP_NAME}-${VERSION}-ubuntu-amd64.deb"

# Build Fedora RPM package
echo "Building Fedora RPM package..."
docker run --rm -v "$(pwd):/app" -w /app fedora:latest bash -c "
  dnf install -y rpm-build
  mkdir -p /app/rpm/{BUILD,RPMS,SOURCES,SPECS,SRPMS}
  cp $OUTPUT_DIR/${APP_NAME}-${VERSION}-linux-amd64 /app/rpm/SOURCES/${APP_NAME}
  echo 'Name: ${APP_NAME}
Version: ${VERSION}
Release: 1
Summary: Easy to use terminal emulator for cloud VM, developer use cases
License: Proprietary

%description
ServlociTerm is an easy to use terminal emulator designed for cloud VM and developer use cases.

%install
mkdir -p %{buildroot}/usr/bin
cp %{_sourcedir}/${APP_NAME} %{buildroot}/usr/bin/${APP_NAME}

%files
/usr/bin/${APP_NAME}
' > /app/rpm/SPECS/${APP_NAME}.spec
  rpmbuild -bb --define '_topdir /app/rpm' /app/rpm/SPECS/${APP_NAME}.spec
  cp /app/rpm/RPMS/x86_64/${APP_NAME}-${VERSION}-1.x86_64.rpm $OUTPUT_DIR/${APP_NAME}-${VERSION}-fedora-x86_64.rpm
"
echo "Fedora RPM package created: $OUTPUT_DIR/${APP_NAME}-${VERSION}-fedora-x86_64.rpm"

echo "Build process completed. Check the $OUTPUT_DIR directory for the output files."
echo "Ubuntu DEB package: $OUTPUT_DIR/${APP_NAME}-${VERSION}-ubuntu-amd64.deb"
echo "Fedora RPM package: $OUTPUT_DIR/${APP_NAME}-${VERSION}-fedora-x86_64.rpm"