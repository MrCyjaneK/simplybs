package=native_rust_1_81_0
$(package)_version=1.81.0
$(package)_download_path=https://static.rust-lang.org/dist
$(package)_file_name=rustc-$($(package)_version)-src.tar.gz
$(package)_download_file=$($(package)_file_name)
$(package)_sha256_hash=872448febdff32e50c3c90a7e15f9bb2db131d13c588fe9071b0ed88837ccfa7
$(package)_build_dependencies=native_ccache native_python3 native_rust_1_80_1

define $(package)_preprocess_cmds
    echo '[build]' > config.toml && \
    echo 'full-bootstrap = true' >> config.toml && \
    echo 'vendor = true' >> config.toml && \
    echo 'extended = false' >> config.toml && \
    echo 'rustc = "$(host_prefix)/native/rust_1_80_1/bin/rustc"' >> config.toml && \
    echo 'cargo = "$(host_prefix)/native/rust_1_54_0/bin/cargo"' >> config.toml && \
    echo '[llvm]' >> config.toml && \
    echo 'ninja = false' >> config.toml && \
    echo 'download-ci-llvm = false' >> config.toml
endef

define $(package)_build_cmds
    python3 ./x.py build --stage 1 -j $(NUM_CORES)
endef

define $(package)_stage_cmds
    mkdir -p $($(package)_staging_prefix_dir)/rust_1_81_0/bin && \
    cp -a build/*/stage1/lib $($(package)_staging_prefix_dir)/rust_1_81_0 && \
    cp -a build/*/stage1/bin/rustc $($(package)_staging_prefix_dir)/rust_1_81_0/bin
endef
