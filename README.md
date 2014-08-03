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
  - Fully asynchronous processing.
  - Uses go's native distribution mechanism for both the editor and plugins.
  - The only editor that has to be installed from sources!


Setup
-----


### Prerequisites

  - git
  - go


### Installation

```
go install github.com/maruel/wi
```


### Updating

```
go install -u github.com/maruel/wi
```


### Installing a plugin

Plugins are simple standalone executables that are transparently started by wi.

```
go install github.com/someone/wi-plugin-awesome
```


Vision
------

  - No I/O done in the UI thread. The UI should always be responsive even on a
    I/O saturated system.
  - Out of process plugins. If a web browser can render web pages out of
    process, a text editor certainly can have its macros defined out of process.
  - Plugins written in the same language as the editor itself. No need to learn
    yet another language and reduces project complexity. The only text editor
    with statically compiled macros!
  - Use the instrinsic go distribution mechanism to distribute plugins. Always
    run at HEAD. Set stable releases in go1 branch.
  - Supports as many OSes as possible.


History
-------

Marc-Antoine badly wanted a new editor that _wouldn't freeze_ when moving the
cursor around even if the HD has high I/O, so he got bored and wrote 'svi' in
2010, which was a python prototype. Obviously, using a text editor with a 3
letters name is unacceptable so the project was renamed to 'wi' in 2014. The old
python was lost due to an HD crash in 2012 and sanity prevailed, the whole
project is being written in #golang. Then he realized that someone had done an
emacs clone, so I'm doing vim. Because of reasons.

Thanks to no.smile.face@gmail.com for the inspiration (godit). This project is
using several libraries from him.
