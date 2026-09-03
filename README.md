# simply

> TODO: inserty catchy phrase later

## What is Simply?

Simply Build System (simplybs) is an effort to create a build system that fits everyone needs.

- No dependency on a single operating system
- Maintains a bootstrapable path for all dependencies
- Simple and easy to understand build definition
- Builds both native tools and target dependencies
- Maintains a familiar enviorment
- Language agnostic instructions
- okay this is a fancy shell script runner, what else do you want me to say?

## Package definition

All package definitions live inside of this repo (this is going to change soon **if** this build system suits my needs). Format as well as content and the way it is being interpreted will change (most likely dependency resolution will be reworked entirely), so I'll only give quick overview.

```json
// package/zlib.json
{
  // should match filename
  "package": "zlib",
  // used to identify built archives
  "version": "1.3.1",
  // type can be either 
  // "host" indicating a package that will run on the target device
  // "native" indicating a package that will be run on the builder
  // (soon) "source" indicationg a package that only contains source code (e.g. that was pulled using custom tools such as `repo` or are too complex for the built in system to handle)
  "type": "host",
  // where to find the source code
  "download": [{
    // "tar.gz" indicates a (who wouldn't have guessed) .tar.gz archive that will be extracted before build steps occur
    // "tar.bz2" indicares a (no way.. is it gonna be..) .tar.bz2 archive that will be extracted.. you get the drill
    // "git" indicates a Git repository being used
    // "none" means no source code is available (can be used for variety of packages to perform operations on existing packages without pulling anything from source)
    "kind": "tar.gz",
    // url should be pointing either to a file or .git repository, depending on .kind
    "url": "http://www.zlib.net/zlib-1.3.1.tar.gz",
    // sha256 is either file checksum or git hash
    "sha256": "9a93b2b7dfdac77ceba5a558a580e74667dd6fede4585b91eefb60f03b72df23"
  },
  "dependencies": [
    // *-android* is being checked against $HOST (always, even on type: native builds)
    // so here native/android_ndk is only going to be extracted into $PREFIX when the
    // build is targetting android.
    // Currently dependency system is not doing recursive resolution (it will properly
    // build all packages recursively but it won't inherit parent dependencies)
    "*-android*:native/android_ndk",
    // all is a magic keyword that works just like *
    "*:native/make",
    "*:native/libtool"
  ],
  "build": {
    "env": [
      // same logic as in dependencies applies, most variables are available during this phase (like $PREFIX or $HOST)
      "*:CFLAGS=$CFLAGS -fPIC",
      "*:config_opts=--prefix=$PREFIX --static",
      "*:LIBTOOL=$NATIVEPREFIX/bin/libtool",
      "*:CROSS_PREFIX=$HOST-"
    ],
    "steps": [
      // step-by-step instructions to build the package.
      "*:./configure $config_opts",
      "*:sed -i.bak s\\|^AR=.*\\|AR=$AR\\|g Makefile",
      "*:sed -i.bak s\\|^ARFLAGS=.*\\|ARFLAGS=$ARFLAGS\\|g Makefile",
      "*:make -j$NUM_CORES",
      "*:make DESTDIR=$STAGING_DIR install"
    ]
  }
}
```

## Usage

In order to build, let's say, `tor` for armv7a-linux-androideabi you would run the following command (on either a Mac or Linux x64 device).

```
$ go run . -host armv7a-linux-androideabi -package tor -build
```

## Recommended / "official" cache settings

```bash
SIMPLYBS_ENV_DIR=/opt/_
```

## Build cache (GitHub Releases)

Built package archives are content-addressed (`package-version-<8-char-hash>`).
They can be shared via a rolling GitHub Release without downloading the whole
cache on every run.

Cache is enabled only when **both** required variables are set:

| Variable | Required | Purpose |
| --- | --- | --- |
| `SIMPLYBS_CACHE_TAG` | yes | Release tag (e.g. `v0-sbs-$USER-$GOOS-$GOARCH`) |
| `SIMPLYBS_CACHE_REPO` | yes | `owner/repo` hosting the release |
| `SIMPLYBS_GH` | no | Optional path to `gh` |

```bash
# export SIMPLYBS_CACHE_TAG=v0-sbs-$USER-$(go env GOOS)-$(go env GOARCH)
# export SIMPLYBS_CACHE_REPO=mrcyjanex/simplybs_private
```

With those set, `-build` auto-pulls artifacts up front, pulls per-package
inside `EnsureBuilt` as a fallback, pushes each package after a successful
local build, and runs a final tree push for anything still missing on the
release. The up-front pull stops at a cache hit: if `native/rust@1_96_0`
is on the release, older rust bootstraps used only to *build* that compiler
are not downloaded.

```bash
go run . -host x86_64-linux-gnu -package zlib -build
```

Explicit pull/push still work:

```bash
go run . -host x86_64-linux-gnu -package zlib -cache-pull
go run . -cache-push                                          # all local built/
go run . -host x86_64-linux-gnu -package zlib -cache-push     # that package tree only
```

Push only uploads local artifacts that are not already on the release (a
package rebuild with a new hash is a new asset name, so “changed” caches
upload naturally).
