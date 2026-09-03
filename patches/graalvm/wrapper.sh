#!/bin/bash
export JAVA_HOME=$NATIVEPREFIX/graalvm
export GRAALVM_HOME=$JAVA_HOME
exec "$JAVA_HOME/bin/@TOOL@" "$@"
