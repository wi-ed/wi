wi - right after vi
===================

*Experimental, do not look into.*

*Experimental, do not look into.*

*Experimental, do not look into.*

_Bringing text based editor technology past 1200 bauds._


[![GoDoc](https://godoc.org/github.com/maruel/wi?status.svg)](https://godoc.org/github.com/maruel/wi)
[![Build Status](https://travis-ci.org/maruel/wi.svg?branch=master)](https://travis-ci.org/maruel/wi)
[![Coverage Status](https://img.shields.io/coveralls/maruel/wi.svg)](https://coveralls.io/r/maruel/wi?branch=master)


Features
--------

  - Editor for the 19.2kbps connected world.
  - Out of process plugins for today's 2Mb RAM systems.
  - [Go](https://golang.org) for both _program_ and _macros_.
  - Fully asynchronous processing. No hang due to I/O ever.
  - Extremely extensible. Everything can be overriden.
  - i18n ready.
  - Auto-generated help.
  - `go get` (Go's native distribution mechanism) for both the editor and
    plugins.
  - Integrated debugging and good test coverage.


Setup
-----


### Prerequisites

  - [git](http://git-scm.com)
  - [Go](https://golang.org)


### Installation or updating

```
go get -u github.com/maruel/wi
```


### Installing or updating a plugin

Plugins are standalone executables or source files that are loaded by `wi`. `wi`
discovers plugins on startup by looking for `wi-plugin-*` / `wi-plugin-*.exe`
and `wi-plugin-*.go` in the same directory (`$GOPATH/bin`) as the `wi`
executable.

```
go get -u github.com/someone/wi-plugin-awesome
```

`*.go` files are sent to `go run` for _on-the-fly_ compilation at
the cost of slower startup time so there's no native updating support.


Vision
------

  - No I/O done in the UI thread. UI must always be responsive even on a I/O
    saturated system.
  - Very extensible editor _with sane default settings_.
    - _Historical reasons are not good reasons_.
  - Out of process plugins. If a
    [web browser](http://dev.chromium.org/developers/design-documents/multi-process-architecture)
    can render web pages out of process, an editor can do the same.
  - Plugins written in the same language as the editor itself. No need to learn
    yet another language (vimscript? lisp? python? javascript?).
      - *The only text editor with statically compiled macros!*
  - Use instrinsic Go distribution mechanism to distribute plugins. Stable
    release is `go1` branch.
  - Broad OS support, including seamless Windows support.
  - Unicode used internally.


Contributing
------------

Run the presubmit check `./presubmit.py` first before doing a pull request. Even
better is to install the git pre-commit hook with `./git-hooks/install.py`.

A CLA (_form to be determined_) will be required for contribution.
