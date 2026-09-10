#!/bin/sh
set -e

case "$CMAKE_SYSTEM_NAME" in
Linux)
    sed -i \
        -e "s|@cmake_system_name@|Linux|" \
        -e "s|@target@|$HOST|" \
        -e "s|@host_prefix@|$PREFIX|" \
        -e "s|@sdk_path@||" \
        -e "s|@cc@||" \
        -e "s|@cxx@||" \
        -e "s|@osx_min_version@||" \
        -e "s|@cmake_c_flags@||" \
        -e "s|@cmake_cxx_flags@||" \
        -e "s|@cmake_ld_flags@||" \
        -e "s|@wmf_libs@||" \
        toolchain.cmake
    ;;
Darwin)
    sed -i \
        -e "s|@cmake_system_name@|Darwin|" \
        -e "s|@target@|$HOST|" \
        -e "s|@host_prefix@|$PREFIX|" \
        -e "s|@sdk_path@|$SDK_PATH|" \
        -e "s|@cc@|$NATIVEPREFIX/bin/$HOST-clang|" \
        -e "s|@cxx@|$NATIVEPREFIX/bin/$HOST-clang++|" \
        -e "s|@osx_min_version@|$OSX_MIN_VERSION|" \
        -e "s|@cmake_c_flags@|-mmacosx-version-min=$OSX_MIN_VERSION -isysroot $SDK_PATH|" \
        -e "s|@cmake_cxx_flags@|-mmacosx-version-min=$OSX_MIN_VERSION -isysroot $SDK_PATH -stdlib=libc++|" \
        -e "s|@cmake_ld_flags@|$LDFLAGS|" \
        -e "s|@wmf_libs@||" \
        toolchain.cmake
    ;;
Windows)
    sed -i \
        -e "s|@cmake_system_name@|Windows|" \
        -e "s|@target@|$HOST|" \
        -e "s|@host_prefix@|$PREFIX|" \
        -e "s|@sdk_path@||" \
        -e "s|@cc@||" \
        -e "s|@cxx@||" \
        -e "s|@osx_min_version@||" \
        -e "s|@cmake_c_flags@|$CFLAGS|" \
        -e "s|@cmake_cxx_flags@|$CXXFLAGS|" \
        -e "s|@cmake_ld_flags@|$LDFLAGS|" \
        -e "s|@wmf_libs@||" \
        toolchain.cmake
    ;;
*)
    echo "qt setup-toolchain: unsupported CMAKE_SYSTEM_NAME=$CMAKE_SYSTEM_NAME" >&2
    exit 1
    ;;
esac
