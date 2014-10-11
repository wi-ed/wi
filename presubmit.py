#!/usr/bin/env python
# Copyright 2014 Marc-Antoine Ruel. All rights reserved.
# Use of this source code is governed under the Apache License, Version 2.0
# that can be found in the LICENSE file.

"""Runs complete presubmit checks on this package."""

import os
import subprocess
import sys
import time


ROOT_DIR = os.path.dirname(os.path.abspath(__file__))


def call(cmd, reldir):
  return subprocess.Popen(
      cmd, cwd=os.path.join(ROOT_DIR, reldir),
      stdout=subprocess.PIPE, stderr=subprocess.STDOUT)


def drain(proc):
  if not proc:
    return 'Process failed'
  out = proc.communicate()[0]
  if proc.returncode:
    return out


def check_or_install(tool, url):
  try:
    # There's no .go files in git-hooks.
    return call([tool], 'git-hooks')
  except OSError:
    print('Warning: installing %s' % url)
    subprocess.check_call(['go', 'get', '-u', url])
    return call([tool], 'git-hooks')


def main():
  start = time.time()

  procs = [
    check_or_install('errcheck', 'github.com/kisielk/errcheck'),
    check_or_install('golint', 'github.com/golang/lint/golint'),
  ]
  while procs:
    drain(procs.pop(0))

  procs = [
    call(['go', 'build'], '.'),
    call(['go', 'test'], 'editor'),
    call(['go', 'test'], 'wi-plugin'),
    call(['go', 'build'], 'wi-plugin-sample'),
    #call(['go', 'test'], 'wi-plugin-sample'),
    call(['errcheck'], '.'),
    call(['errcheck'], 'editor'),
    call(['errcheck'], 'wi-plugin'),
    call(['errcheck'], 'wi-plugin-sample'),
    call(['golint'], '.'),
    call(['golint'], 'editor'),
    call(['golint'], 'wi-plugin'),
    call(['golint'], 'wi-plugin-sample'),
  ]
  failed = False
  out = drain(procs.pop(0))
  if out:
    failed = True
    print out

  for p in procs:
    out = drain(p)
    if out:
      failed = True
      print out

  end = time.time()
  if failed:
    print('Presubmit checks failed in %1.3fs!' % (end-start))
    return 1
  print('Presubmit checks succeeded in %1.3fs!' % (end-start))
  return 0


if __name__ == '__main__':
  sys.exit(main())
