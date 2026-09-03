#!/bin/bash
export JAVA_HOME=$NATIVEPREFIX/jdk-@JDK_VERSION@
exec "$JAVA_HOME/bin/@TOOL@" "$@"
