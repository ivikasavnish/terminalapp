#!/bin/bash

# Exit on any error
set -e

# Application details
APP_NAME="ServlociTerm"
VERSION="1.0.0"
MAINTAINER="ServlociTerm Team <support@servloci.com>"
DESCRIPTION="Easy to use terminal emulator for cloud VM, developer use cases"

# Build the application
echo "Building ServlociTerm for Linux..."
wails build -platform linux/amd64

# Create DEB package
echo "Creating DEB package..."
mkdir -p ./deb/DEBIAN ./deb/usr/bin
cp ./build/bin/$APP_NAME ./deb/usr/bin/
cat > ./deb/DEBIAN/control << EOF
Package: $APP_NAME
Version: $VERSION
Architecture: amd64
Maintainer: $MAINTAINER
Description: $DESCRIPTION
EOF

dpkg-deb --build ./deb ./$APP_NAME-$VERSION-amd64.deb

# Create RPM package
echo "Creating RPM package..."
mkdir -p ./rpm/{BUILD,RPMS,SOURCES,SPECS,SRPMS}
cp ./build/bin/$APP_NAME ./rpm/SOURCES/

cat > ./rpm/SPECS/$APP_NAME.spec << EOF
Name: $APP_NAME
Version: $VERSION
Release: 1
Summary: $DESCRIPTION
License: Proprietary

%description
$DESCRIPTION

%install
mkdir -p %{buildroot}/usr/bin
cp %{_sourcedir}/$APP_NAME %{buildroot}/usr/bin/$APP_NAME

%files
/usr/bin/$APP_NAME
EOF

rpmbuild -bb --define "_topdir $(pwd)/rpm" ./rpm/SPECS/$APP_NAME.spec

mv ./rpm/RPMS/x86_64/$APP_NAME-$VERSION-1.x86_64.rpm ./$APP_NAME-$VERSION-x86_64.rpm

echo "Linux builds created:"
echo "DEB: $APP_NAME-$VERSION-amd64.deb"
echo "RPM: $APP_NAME-$VERSION-x86_64.rpm"

# Clean up
rm -rf ./deb ./rpm