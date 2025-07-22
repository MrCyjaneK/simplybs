package=native_rust_1_63_0
$(package)_version=1.63.0
$(package)_download_path=https://static.rust-lang.org/dist
$(package)_file_name=rustc-$($(package)_version)-src.tar.gz
$(package)_download_file=$($(package)_file_name)
$(package)_sha256_hash=1f9580295642ef5da7e475a8da2397d65153d3f2cb92849dbd08ed0effca99d0
$(package)_build_dependencies=native_ccache native_python3 native_rust_1_62_1 native_cargo_0_64_0
$(package)_patches=assembly.h.patch
# bump cargo here
define $(package)_preprocess_cmds
    cd $($(package)_extract_dir) && \
    patch -p1 < $($(package)_patch_dir)/assembly.h.patch && \
    echo '[build]' > config.toml && \
    echo 'full-bootstrap = true' >> config.toml && \
    echo 'vendor = true' >> config.toml && \
    echo 'extended = false' >> config.toml && \
    echo 'rustc = "$(host_prefix)/native/rust_1_62_1/bin/rustc"' >> config.toml && \
    echo 'cargo = "$(host_prefix)/native/cargo_0_64_0/bin/cargo"' >> config.toml && \
    echo '[llvm]' >> config.toml && \
    echo 'ninja = false' >> config.toml && \
    echo 'download-ci-llvm = false' >> config.toml
endef

define $(package)_build_cmds
    python3 ./x.py build --stage 1 -j $(NUM_CORES)
endef

define $(package)_stage_cmds
    mkdir -p $($(package)_staging_prefix_dir)/rust_1_63_0/bin && \
    cp -a build/*/stage1/lib $($(package)_staging_prefix_dir)/rust_1_63_0 && \
    cp -a build/*/stage1/bin/rustc $($(package)_staging_prefix_dir)/rust_1_63_0/bin
endef
