# Development

## Testing

To run the tests, run the following command:

```bash
go test ./...
```

or use the `make` command:

```bash
make test
```

To run the tests with verbose output, run the following command:

```bash
go test -v ./...
```

To run a specific test suite, run the following command:

```bash
go test -v ./test_suite.go
```

To run a specific test, run the following command:

```bash
go test -v ./test_suite.go -run TestName
```

## Running Hide locally

To run Hide locally, run the following command:

```bash
go run ./cmd/hide
```

or use the `make` command:

```bash
make run
```

This will start a local server at `http://127.0.0.1:8080/`.

## Release

To release a new version of Hide, follow these steps:

1. Create a new release using the GitHub UI or the command line. For example, to create a new release with the command line, run the following command:

    ```bash
    gh release create vX.Y.Z --title "Hide vX.Y.Z" --generate-notes
    ```

    Replace `X.Y.Z` with the version number of the new release following the [semantic versioning](https://semver.org/) convention.

    For additional options and for UI instructions, refer to the [GitHub documentation](https://docs.github.com/en/repositories/releasing-projects-on-github/managing-releases-in-a-repository).

2. Update the version in the [hide brew](https://github.com/hide-org/homebrew-formulae/blob/main/Formula/hide.rb) formula:
  1. Copy the URL of the new release (tar.gz file) from the GitHub UI or the command line.
  2. Get the SHA256 checksum of the new release (tar.gz file) using the command line:

     ```bash
     sha256sum vX.Y.Z.tar.gz
     ```

  3. Update the `url` and the `sha256` fields in the `hide-brew.rb` file to the new release URL and checksum.

## Documentation

The documentation is built using [MkDocs](https://www.mkdocs.org/). To build the documentation, install MkDocs

```bash
pip install mkdocs
```

and then run the following command:

```bash
mkdocs build
```

The documentation will be built in the `site` directory.

To serve the documentation locally, run the following command:

```bash
mkdocs serve
```

This will start a local server at `http://127.0.0.1:8000/`.
