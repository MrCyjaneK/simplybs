package=native_cargo_0_68_0
$(package)_version=0.68.0
$(package)_download_path=https://github.com/rust-lang/cargo/archive/refs/tags
$(package)_file_name=cargo-$($(package)_version).tar.gz
$(package)_download_file=0.68.0.tar.gz
$(package)_sha256_hash=b60b794dfdca61dfad106dbbb9c7c2c077c1a4bd505b90988057be5fd1ae99b7
$(package)_build_dependencies=native_ccache native_python3 native_cargo_0_64_0 native_rust_1_67_1

define $(package)_env
    $(package)_config_env=PATH="$(host_prefix)/native/rust_1_67_1/bin:$(host_prefix)/native/cargo_0_64_0/bin:$(host_prefix)/native/bin:${PATH}" CARGO="$(host_prefix)/native/cargo_0_68_0/bin/cargo" RUSTC="$($(package)_host_prefix)/native/rust_1_67_1/bin/rustc"
endef

define $(package)_build_cmds
    cargo build --release
endef

define $(package)_stage_cmds
    mkdir -p $($(package)_staging_prefix_dir)/cargo_0_68_0/bin && \
    cp -a target/release/cargo $($(package)_staging_prefix_dir)/cargo_0_68_0/bin
endef
