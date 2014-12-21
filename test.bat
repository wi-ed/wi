@echo off
:: Copyright 2014 Marc-Antoine Ruel. All rights reserved.
:: Use of this source code is governed under the Apache License, Version 2.0
:: that can be found in the LICENSE file.

:: Short test until we got something up and running.
cls
go build -tags debug
if errorlevel 1 goto :EOF
wi -c log_all editor_quit
if errorlevel 1 goto :EOF
type wi.log
