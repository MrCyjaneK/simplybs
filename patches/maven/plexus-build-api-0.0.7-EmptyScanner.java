/*
Copyright (c) 2008 Sonatype, Inc. All rights reserved.

This program is licensed to you under the Apache License Version 2.0,
and you may not use this file except in compliance with the Apache License Version 2.0.
You may obtain a copy of the Apache License Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0.

Unless required by applicable law or agreed to in writing,
software distributed under the Apache License Version 2.0 is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the Apache License Version 2.0 for the specific language governing permissions and limitations there under.
*/

package org.sonatype.plexus.build.incremental;

import java.io.File;
import java.util.Comparator;

import org.codehaus.plexus.util.Scanner;

/**
 * Scanner implementation never finds any files/directories.
 */
public class EmptyScanner implements Scanner {

  private static final String[] EMPTY_STRING_ARRAY = new String[0];

  private final File basedir;

  public EmptyScanner(File basedir) {
    this.basedir = basedir;
  }

  public void addDefaultExcludes() {
  }

  public String[] getIncludedDirectories() {
    return EMPTY_STRING_ARRAY;
  }

  public String[] getIncludedFiles() {
    return EMPTY_STRING_ARRAY;
  }

  public void scan() {
  }

  public void setExcludes(String[] excludes) {
  }

  public void setIncludes(String[] includes) {
  }

  public void setFilenameComparator(Comparator filenameComparator) {
  }

  public File getBasedir() {
    return basedir;
  }

}
