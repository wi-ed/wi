wi - right after vi
===================

*Experimental, do not look into.*

*Experimental, do not look into.*

*Experimental, do not look into.*

_Bringing text based editor technology past 1200 bauds._


Features
--------

  - Text based terminal for the 19.2kbps connected world.
  - Out of process plugins for today's 2Mb RAM systems.
  - Single language [Go](https://golang.org) to rule both the _program itself_ and _its macros_.
  - Fully asynchronous processing. No hang due to I/O ever.
  - Extremely extensible. Everything can be overriden.
  - i18n ready.
  - Auto-generated help.
  - `go get` (Go's native distribution mechanism) for both the editor and plugins.


Setup
-----


### Prerequisites

  - git
  - go


### Installation or updating

```
go get -u github.com/maruel/wi
```


### Installing or updating a plugin

Plugins are simple standalone executables that are transparently started by
`wi`.  They must be named `wi-plugin-*` (`wi-plugin-*.exe` on Windows). The
fact they exist in the same directory (`$GOPATH/bin`) as the `wi` executable is
enough to have the plugin started by `wi` automatically.

```
go get -u github.com/someone/wi-plugin-awesome
```


Vision
------

  - No I/O done in the UI thread. The UI should always be responsive even on a
    I/O saturated system.
  - Very extensible editor _with sane default settings_.
    - _Historical reasons are not good reasons_.
  - Out of process plugins. If a
    [web browser](http://dev.chromium.org/developers/design-documents/multi-process-architecture)
    can render web pages out of process, a text editor certainly can have its
    macros defined out of process.
  - Plugins written in the same language as the editor itself. No need to learn
    yet another language (vimscript? lisp? python?).
      - The only text editor with statically compiled macros!
  - Use the instrinsic Go distribution mechanism to distribute plugins. Use
    stable releases in `go1` branch by default.
  - Supports as many OSes as possible.
  - Handles everything internally as unicode characters.


Contributing
------------

Please run the presubmit check `./presubmit.py` first before doing a pull
request. Even better is to install the git pre-commit hook with
`./git-hooks/install.py`.

See online documentation on
[![GoDoc](https://godoc.org/github.com/maruel/wi?status.svg)](https://godoc.org/github.com/maruel/wi)
