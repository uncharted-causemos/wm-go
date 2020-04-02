# WM Go API

[![pipeline status](https://gitlab.uncharted.software/WM/wm-go/badges/master/pipeline.svg)](https://gitlab.uncharted.software/WM/wm-go/commits/master)

## Requirements

* Go v1.12 or higher. Make sure your `$GOPATH` is defined and that `$GOPATH/bin` is in your path.

## Instructions

Clone the repository and install the dependencies.

```
git clone git@gitlab.uncharted.software:WM/wm-go.git
cd wm-go
make install
```

Copy `sample.env` as `wm.env` and adjust the environment variables as needed.

Run the server:

```
make run
```
