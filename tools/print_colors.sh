#!/bin/bash
# Copyright 2014 Marc-Antoine Ruel. All rights reserved.
# Use of this source code is governed under the Apache License, Version 2.0
# that can be found in the LICENSE file.

# Prints all 256 colors in a terminal.

reset=$(tput op)
y=$(printf %$((${COLUMNS}-6))s)
for i in {0..255}; do
  index=00$i
  echo -e ${index:${#index}-3:3} `tput setaf $i;tput setab $i`${y// /=}$reset
done
