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
`wi`.  They must be named `wi-plugins-*` (`wi-plugins-*.exe` on Windows). The
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


History
-------

Marc-Antoine badly wanted a new editor that _wouldn't freeze_ when moving the
cursor around even if the HD has high I/O or the disk went to sleep, so he got
bored and wrote `svi` in 2010, which was a python prototype.

Obviously, using a text editor with a 3 letters name is unacceptable so the
project was renamed to `wi` in 2014. The old python was lost due to an HD crash
in 2012 and sanity prevailed, the whole project is being written in
[Go](https://golang.org). Then he realized that someone had done an emacs clone,
so he's doing vim. Because of reasons.


Thanks
------

  - Bram Moolenaar for vim.
  - no.smile.face@gmail.com for the inspiration
    [godit](https://github.com/nsf/godit). This project is using several
    libraries from him which helped bootstrapping this project much more
    quickly.
  - Kamil Kisiel's [vigo](https://github.com/kisielk/vigo) was also an
    inspiration.


Contributing
------------

Please run the presubmit check `./presubmit.py` first before doing a pull
request. Even better is to install the git pre-commit hook with
`./git-hooks/install.py`.

You'll also want to run:

    go get -u code.google.com/p/go.tools/cmd/godoc
    godoc -http=:6060

Then browse to http://localhost:6060/pkg/github.com/maruel/wi for a structure of
the source tree. You can also visit http://godoc.org/github.com/maruel/wi for
online browseable documentation.
