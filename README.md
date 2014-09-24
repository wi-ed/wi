wi - right after vi
===================

*Experimental, do not look into.*

*Experimental, do not look into.*

*Experimental, do not look into.*

_Bringing text based editor technology past 1200 bauds._


Features
--------

  - Editor for the 19.2kbps connected world.
  - Out of process plugins for today's 2Mb RAM systems.
  - [Go](https://golang.org) both for _program_ and _macros_.
  - Fully asynchronous processing. No hang due to I/O ever.
  - Extremely extensible. Everything can be overriden.
  - i18n ready.
  - Auto-generated help.
  - `go get` (Go's native distribution mechanism) for both the editor and plugins.


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

Plugins are standalone executables that are loaded by `wi`. `wi` discovers
plugins on startup by looking for `wi-plugin-*` / `wi-plugin-*.exe` in the same
directory (`$GOPATH/bin`) as the `wi` executable.

```
go get -u github.com/someone/wi-plugin-awesome
```


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
    yet another language (vimscript? lisp? python? Ah! ha!).
      - *The only text editor with statically compiled macros!*
  - Use instrinsic Go distribution mechanism to distribute plugins. Stable
    release is `go1` branch.
  - Broad OS support.
  - Unicode used internally.


Contributing
------------

Run the presubmit check `./presubmit.py` first before doing a pull request. Even
better is to install the git pre-commit hook with `./git-hooks/install.py`.

See online documentation on
[![GoDoc](https://godoc.org/github.com/maruel/wi?status.svg)](https://godoc.org/github.com/maruel/wi)
