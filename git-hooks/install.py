#!/usr/bin/env python
# Copyright 2014 Marc-Antoine Ruel. All rights reserved.
# Use of this source code is governed under the Apache License, Version 2.0
# that can be found in the LICENSE file.

import os
import subprocess
import sys


THIS_DIR = os.path.dirname(os.path.abspath(__file__))


def main():
  git_dir = subprocess.check_output(['git', 'rev-parse', '--git-dir']).strip()
  git_hook_dir = os.path.join(git_dir, 'hooks')
  relpath = os.path.relpath(THIS_DIR, git_hook_dir)
  os.symlink(
      os.path.join(relpath, 'pre-commit'),
      os.path.join(git_hook_dir, 'pre-commit'))
  return 0


if __name__ == '__main__':
  sys.exit(main())
