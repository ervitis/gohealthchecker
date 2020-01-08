# gohealthchecker

Go healhchecker is a library for building easy custom functions to use them as your health check in your microservice

This healthchecker library goes fine with [kubernetes](https://kubernetes.io/) or [openshift](https://www.openshift.com/) platform where you set the intervals of your health checkers calls.

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

Let's take a look at the example:

```go
func checkPort() gohealthchecker.Healthfunc {
	return func() (code int, e error) {
		conn, err := net.Dial("tcp", ":8185")
		if err != nil {
			return http.StatusInternalServerError, err
		}

		_ = conn.Close()
		return http.StatusOK, nil
	}
}

func main() {
	health := gohealthchecker.NewHealthchecker(http.StatusOK, http.StatusInternalServerError)

	health.Add(checkPort(), "checkPort")

	panic(http.ListenAndServe(":8085", health.ActivateHealthCheck("health")))
}
```

You can instance a `healthchecker` by using the `NewHealthchecker` function passing the two types of HTTP responses you want when it goes well or not.

Create a function for example `checkPort()` that will return a `Healthfunc` type. Inside that function add the logic you want for the healthchecker.

After that, add it using the `Add` method. You can pass a name if you want.

If you use the example it will return a KO message because it will check if the port 8185 is opened (or maybe yes, that depend of your opened ports)
The message would be like this:

```json
{
  "info":[
    {
      "message":"dial tcp :8185: connect: connection refused",
      "code":500,
      "service":"checkPort"
    }
  ],
  "code":500
}
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

See also the list of [contributors](./contributors.md) who participated in this project.

## License

This project is licensed under the Apache 2.0 - see the [LICENSE.md](LICENSE.md) file for details
