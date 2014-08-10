# Copyright 2014 Marc-Antoine Ruel. All rights reserved.
# Use of this source code is governed under the Apache License, Version 2.0
# that can be found in the LICENSE file.

set -e

cd "$(dirname $0)"

# First, ensure everything is buildable.
go build
cd wi-plugin-sample
go build
cd ..

# TODO(maruel): Run all the following in parallel!

# Second, ensure tests pass.
go test
go test ./wi-plugin

# Third, ensure error return values are checked.
# go install github.com/kisielk/errcheck
errcheck
