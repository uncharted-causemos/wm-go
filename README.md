# WM Go API

[![pipeline status](https://gitlab.uncharted.software/WM/wm-go/badges/master/pipeline.svg)](https://gitlab.uncharted.software/WM/wm-go/commits/master)  [![coverage report](https://gitlab.uncharted.software/WM/wm-go/badges/master/coverage.svg)](https://gitlab.uncharted.software/WM/wm-go/commits/master)

## Requirements

* Go v1.12 or higher. Make sure your `$GOPATH` is defined and that `$GOPATH/bin` is in your path.
* Ensure you have Go Lint installed; eg: `go get -u golang.org/x/lint/golint`

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

## Note on CI/CD Workflows
  - Linting and test runs when there's a merge requests or push to master.
  - Docker image with latest tag will be created and pushed to the registry when changes are committed to master.
  - Docker image with a tag (eg. 0.1.1) will be created (and pushed to the registry) and will be deployed to openstack instance when a commit is tagged.
