
## How to use
### Build

```sh
make build

# run:
./bin/gouploadserver
```

### Installation

```sh
make install

# run:
gouploadserver
```

## References

- https://www.digitalocean.com/community/tutorials/how-to-build-and-install-go-programs-pt
- https://golang.org/doc/tutorial/compile-install
- https://golang.org/ref/mod#go-install
- https://makefiletutorial.com/

```sh
➜  cmd go list -f '{{.Target}}'
/Users/guilherme/go/bin/cmd
```

```sh
go test -v -benchmem -bench=.
```