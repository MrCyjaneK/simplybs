packages:=boost openssl zeromq libiconv expat unbound polyseed

native_packages := native_ccache

rust_packages := native_mrustc native_rust_1_55_0 native_rust_1_56_1 native_rust_1_57_0 native_rust_1_58_1 native_rust_1_59_0 native_rust_1_60_0 native_rust_1_61_0 native_rust_1_62_1 native_rust_1_63_0 native_rust_1_64_0 native_rust_1_65_0 native_rust_1_66_1 native_rust_1_67_1 native_rust_1_68_2 native_rust_1_69_0 native_rust_1_70_0 native_rust_1_71_1 native_rust_1_72_1 native_rust_1_73_0 native_rust_1_74_1 native_rust_1_75_0 native_rust_1_76_0 native_rust_1_77_1 native_rust_1_78_0 native_rust_1_79_0 native_rust_1_80_1 native_rust_1_81_0 native_rust_1_82_0 native_rust_1_83_0 native_rust_1_84_1 native_rust_1_85_1 native_rust_1_86_0 native_rust_1_87_0 native_rust_1_88_0

cargo_packages := native_cargo_0_58_0 native_cargo_0_64_0 native_cargo_0_68_0 native_cargo_0_75_0 native_cargo_0_88_0

native_packages += $(cargo_packages)
native_packages += $(rust_packages)

native_packages += native_python3 native_nproc

hardware_packages := hidapi protobuf libusb
hardware_native_packages := native_protobuf

android_native_packages = android_ndk
android_packages = ncurses readline sodium

darwin_native_packages = $(hardware_native_packages)
darwin_packages = ncurses readline sodium $(hardware_packages)
ios_packages = sodium protobuf native_protobuf
iossimulator_packages = sodium protobuf native_protobuf

# not really native...
freebsd_native_packages = freebsd_base
freebsd_packages = ncurses readline sodium

linux_packages = eudev ncurses readline sodium $(hardware_packages)
linux_native_packages = $(hardware_native_packages)

ifeq ($(build_tests),ON)
packages += gtest
endif

ifneq ($(host_arch),riscv64)
linux_packages += unwind
endif

mingw32_packages = icu4c sodium $(hardware_packages)
mingw32_native_packages = $(hardware_native_packages)

ifneq ($(build_os),darwin)
darwin_native_packages += darwin_sdk native_clang native_cctools native_libtapi
endif

