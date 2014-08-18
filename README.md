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
  - Gorgeous EGA colours!
  - Single language to rule both the program itself and its macros.
  - Fully asynchronous processing. No hang due to I/O ever.
  - Extremely extensible. Everything can be overriden.
  - Can be translated.
  - Integrated auto-generated help.
  - Uses go's native distribution mechanism for both the editor and plugins.
  - The only editor (worth installing) that has to be installed from sources!


Setup
-----


### Prerequisites

  - git
  - go


### Installation

```
go install github.com/maruel/wi
```


### Installing a plugin

Plugins are simple standalone executables that are transparently started by
`wi`.  They must be named `wi-plugins-*` (`wi-plugins-*.exe` on Windows). The
fact they exist in the same directory (`$GOPATH/bin`) as the `wi` executable is
enough to have the plugin started by `wi` automatically.

```
go install github.com/someone/wi-plugin-awesome
```


### Updating wi

Releases are going to be done in the `go1` branch, so updating `wi` to the
current stable release is trivial:

```
go get -u github.com/maruel/wi
go install github.com/maruel/wi
```


### Updating a plugin

It's also trivial to update wi plugins after updating `wi`. The APIs are
strongly versioned so after updating `wi`, the plugins may not start until they
are rebuilt:

```
go get -u github.com/someone/wi-plugin-awesome
go install github.com/someone/wi-plugin-awesome
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
    yet another language and reduces project complexity.
      - The only text editor with statically compiled macros!
  - Use the instrinsic go distribution mechanism to distribute plugins. Always
    use stable releases in `go1` branch by default. Easily (optionally) run at
    `HEAD` if desired.
      - No need to wait for years for Debian stable to be updated anymore.
  - Supports as many OSes as possible.
      - Because Windows users have pitying look.
  - Handles everything internally as unicode characters.
  - Open source editor with source code that is Write-Once-Read-Multi, e.g.
    readable source code that just works.


History
-------

Marc-Antoine badly wanted a new editor that _wouldn't freeze_ when moving the
cursor around even if the HD has high I/O or the disk went to sleep, so he got
bored and wrote 'svi' in 2010, which was a python prototype.

Obviously, using a text editor with a 3 letters name is unacceptable so the
project was renamed to 'wi' in 2014. The old python was lost due to an HD crash
in 2012 and sanity prevailed, the whole project is being written in
[Go](https://golang.org). Then he realized that someone had done an emacs clone,
so he's doing vim. Because of reasons.


Thanks
------

  - no.smile.face@gmail.com for the inspiration
    [godit](https://github.com/nsf/godit). This project is using several
    libraries from him which helped bootstrapping this project much more
    quickly.
  - Kamil Kisiel's [vigo](https://github.com/kisielk/vigo) was also an
    inspiration.


Contributing
------------

Please run `./presubmit.py` first before doing a pull request.
