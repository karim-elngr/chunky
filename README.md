# Chunky
A fast file downloader

## Overview
This program is a command-line tool that downloads a file over HTTP in parallel. 

## Features
1. **Parallel Downloads:** Splits the file into chunks and downloads them concurrently using goroutines.
2. **HTTP Range Requests:** Utilizes HTTP range requests to download specific parts of the file.

## Usage
To test the program, use the following example URL:

```shell
  chunky download --url=https://files.testfile.org/PDF/50MB-TESTFILE.ORG.pdf --directory=. --parallelism=4 --size=1048576
```