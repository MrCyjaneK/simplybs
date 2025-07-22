package=native_rust_1_80_1
$(package)_version=1.80.1
$(package)_download_path=https://static.rust-lang.org/dist
$(package)_file_name=rustc-$($(package)_version)-src.tar.gz
$(package)_download_file=$($(package)_file_name)
$(package)_sha256_hash=2c0b8f643942dcb810cbcc50f292564b1b6e44db5d5f45091153996df95d2dc4
$(package)_build_dependencies=native_ccache native_python3 native_rust_1_79_0

define $(package)_preprocess_cmds
    echo '[build]' > config.toml && \
    echo 'full-bootstrap = true' >> config.toml && \
    echo 'vendor = true' >> config.toml && \
    echo 'extended = false' >> config.toml && \
    echo 'rustc = "$(host_prefix)/native/rust_1_79_0/bin/rustc"' >> config.toml && \
    echo 'cargo = "$(host_prefix)/native/rust_1_54_0/bin/cargo"' >> config.toml && \
    echo '[llvm]' >> config.toml && \
    echo 'ninja = false' >> config.toml && \
    echo 'download-ci-llvm = false' >> config.toml
endef

define $(package)_build_cmds
    python3 ./x.py build --stage 1 -j $(NUM_CORES)
endef

define $(package)_stage_cmds
    mkdir -p $($(package)_staging_prefix_dir)/rust_1_80_1/bin && \
    cp -a build/*/stage1/lib $($(package)_staging_prefix_dir)/rust_1_80_1 && \
    cp -a build/*/stage1/bin/rustc $($(package)_staging_prefix_dir)/rust_1_80_1/bin
endef
