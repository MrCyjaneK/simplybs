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
  "download": {
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
    "all:native/make",
    "all:native/libtool"
  ],
  "build": {
    "env": [
      // same logic as in dependencies applies, most variables are available during this phase (like $PREFIX or $HOST)
      "all:CFLAGS=$CFLAGS -fPIC",
      "all:config_opts=--prefix=$PREFIX --static",
      "all:LIBTOOL=$PREFIX/native/bin/libtool",
      "all:CROSS_PREFIX=$HOST-"
    ],
    "steps": [
      // step-by-step instructions to build the package.
      "all:./configure $config_opts",
      "all:sed -i.bak s\\|^AR=.*\\|AR=$AR\\|g Makefile",
      "all:sed -i.bak s\\|^ARFLAGS=.*\\|ARFLAGS=$ARFLAGS\\|g Makefile",
      "all:make -j$NUM_CORES",
      "all:make DESTDIR=$STAGING_DIR install"
    ]
  }
}
```

## Usage

In order to build, let's say, `libtor` for armv7a-linux-androideabi you would run the following command (on either a Mac or Linux x64 device).

```
$ go run . -host armv7a-linux-androideabi -package libtor -build
```