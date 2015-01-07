@echo off
:: Copyright 2014 Marc-Antoine Ruel. All rights reserved.
:: Use of this source code is governed under the Apache License, Version 2.0
:: that can be found in the LICENSE file.

setlocal

:: Short test until we got something up and running.
cd %~dp0\..
cls
go build -race -tags debug
if errorlevel 1 goto :EOF
set WIPLUGINSPATH=.
wi -c log_all editor_quit
if errorlevel 1 goto :EOF
type wi.log
