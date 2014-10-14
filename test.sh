#!/bin/bash
# Copyright 2014 Marc-Antoine Ruel. All rights reserved.
# Use of this source code is governed under the Apache License, Version 2.0
# that can be found in the LICENSE file.

# Short test until we got something up and running.

set -e

go build -tags debug
./wi -c log_all editor_quit
cat wi.log
