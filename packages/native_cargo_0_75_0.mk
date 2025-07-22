package=native_cargo_0_75_0
$(package)_version=0.75.0
$(package)_download_path=https://github.com/rust-lang/cargo/archive/refs/tags
$(package)_file_name=cargo-$($(package)_version).tar.gz
$(package)_download_file=0.75.0.tar.gz
$(package)_sha256_hash=d6b9512bca4b4d692a242188bfe83e1b696c44903007b7b48a56b287d01c063b
$(package)_build_dependencies=native_ccache native_python3 native_cargo_0_68_0 native_rust_1_75_0

define $(package)_env
    $(package)_config_env=PATH="$(host_prefix)/native/rust_1_75_0/bin:$(host_prefix)/native/cargo_0_68_0/bin:$(host_prefix)/native/bin:${PATH}" CARGO="$(host_prefix)/native/cargo_0_76_0/bin/cargo" RUSTC="$($(package)_host_prefix)/native/rust_1_75_0/bin/rustc"
endef

define $(package)_build_cmds
    cargo build --release
endef

define $(package)_stage_cmds
    mkdir -p $($(package)_staging_prefix_dir)/cargo_0_76_0/bin && \
    cp -a target/release/cargo $($(package)_staging_prefix_dir)/cargo_0_76_0/bin
endef
