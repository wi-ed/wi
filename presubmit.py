#!/usr/bin/env python
# Copyright 2014 Marc-Antoine Ruel. All rights reserved.
# Use of this source code is governed under the Apache License, Version 2.0
# that can be found in the LICENSE file.

"""Copy this script in your package if desired."""

import os
import sys


THIS_DIR = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, os.path.join(THIS_DIR, 'git-hooks-go'))

import presubmit_impl


if __name__ == '__main__':
  # Disable golint or govet if desired.
  sys.exit(presubmit_impl.main(tags='debug', run_golint=True, run_govet=True))
