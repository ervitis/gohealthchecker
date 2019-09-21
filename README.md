# gohealthchecker

Go healhchecker is a library for building easy custom functions to use them as your health check in your microservice

## Getting Started

Install it with

> dep

```bash
dep ensure -add github.com/ervitis/gohealthchecker
```

> go modules

```bash
go get -u github.com/ervitis/gohealthchecker
```

If you want to test it:

```bash
go test -v -race ./...
```

### Prerequisites

Install the prerequisites using go modules

```bash
go mod install
```

### Installing

Use the library in your microservice app.

You can see an example inside the folder `examples`

### And coding style tests

No library is needed for making additional tests. Use the `testing` library from Golang

## Contributing

Please read [CONTRIBUTING.md](https://gist.github.com/PurpleBooth/b24679402957c63ec426) for details on our code of conduct, and the process for submitting pull requests to us.

## Versioning

We use [SemVer](http://semver.org/) for versioning. For the versions available, see the [tags on this repository](https://github.com/your/project/tags). 

## Authors

* **ervitis** - *Initial work* - [ervitis](https://github.com/ervitis)

See also the list of [contributors](https://github.com/your/project/contributors) who participated in this project.

## License

This project is licensed under the Apache 2.0 - see the [LICENSE.md](LICENSE.md) file for details
