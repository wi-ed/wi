#!/usr/bin/env python
# Copyright 2014 Marc-Antoine Ruel. All rights reserved.
# Use of this source code is governed under the Apache License, Version 2.0
# that can be found in the LICENSE file.

"""Runs complete presubmit checks on this package."""

import logging
import optparse
import os
import subprocess
import sys
import time


THIS_FILE = os.path.abspath(__file__)
ROOT_DIR = os.path.dirname(THIS_FILE)


def call(cmd, reldir):
  logging.info('cwd=%s; %s', reldir, ' '.join(cmd))
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
    return call(tool, 'git-hooks')
  except OSError:
    print('Warning: installing %s' % url)
    subprocess.check_call(['go', 'get', '-u', url])
    return call(tool, 'git-hooks')


def main():
  parser = optparse.OptionParser(description=sys.modules[__name__].__doc__)
  parser.add_option(
      '-v', '--verbose', action='store_true', help='Logs what is being run')
  parser.add_option(
      '--goimport', action='store_true', help=optparse.SUPPRESS_HELP)
  options, args = parser.parse_args()
  if args:
    parser.error('Unknown args: %s' % args)

  if options.goimport:
    # goimports doesn't return non-zero even if some files need to be updated.
    out = subprocess.check_output(['goimports', '-l', '.'])
    if out:
      print('These files are improperly formmatted. Please run: goimports -w .')
      sys.stdout.write(out)
      return 1
    return 0

  logging.basicConfig(
      level=logging.DEBUG if options.verbose else logging.ERROR,
      format='%(levelname)-5s: %(message)s')

  start = time.time()
  procs = [
    check_or_install(['errcheck'], 'github.com/kisielk/errcheck'),
    check_or_install(['goimports', '.'], 'code.google.com/p/go.tools/cmd/goimports'),
    check_or_install(['golint'], 'github.com/golang/lint/golint'),
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
    call([sys.executable, THIS_FILE, '--goimport'], '.'),
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
