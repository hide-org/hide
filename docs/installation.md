# Installation

A typical installation for Hide consists of 2 parts: a runtime that runs on a local or remote Docker host, and an SDK that interacts with it.

Runtime is the backend system responsible for managing development containers and executing tasks.

SDK is a set of APIs and toolkits designed for coding agents to interact with the codebase.

## Prerequisites

### Docker

Hide Runtime requires Docker Engine to be installed on your system. Note that if you intend to use Hide with a remote Docker host, you will need to install Docker Engine on that host.

For installation instructions for your OS, see the [Docker Engine documentation](https://docs.docker.com/engine/install/).

## Runtime

The runtime is responsible for managing the development environments, executing tasks, and providing APIs for interacting with the codebase. Runtime can be installed using [Homebrew](https://brew.sh/) or built from source.

### Using Homebrew

1. Add the Hide tap to your Homebrew:

    ```bash
    brew tap hide-org/formulae
    ```

2. Install Hide using the brew install command:

    ```bash
    brew install hide
    ```

### Building from Source

To build Hide from source, follow these steps:

1. Ensure you have [Go 1.22+](https://go.dev/) or later installed on your system.
2. Clone the repository:

    ```bash
    git clone https://github.com/hide-org/hide.git
    cd hide
    ```

3. Build Hide and install it to your `$HOME/go/bin` directory:

    ```bash
    make install
    ```

    !!! note

        Make sure that `$HOME/go/bin` is in your `$PATH` environment variable e.g. `export PATH=$PATH:$HOME/go/bin`.


4. Install LSP server for your language of choice.

    For Python, install the `pyright` package:

    ```bash
    pipx install pyright
    ```

    Note that we use [pipx](https://pipx.pypa.io/stable/) to install the package globally in isolated environment. After installation run `pyright` command in the shell to install `nodejs` if it's missing (Pyright is written in Typescript and it will install `nodejs` for you). 

    For JavaScript and TypeScript, install the `typescript-language-server` package:

    ```bash
    npm install -g typescript-language-server
    ```

    For Go, install the `gopls` package:

    ```bash
    go install golang.org/x/tools/gopls@latest
    ```

### Running Hide

After installing Hide, you can start the runtime by running the following command:

```bash
hide run
```
You should see logs indicating that the server is running, something like: `Server started on 127.0.0.1:8080`. For more options, including how to specify the port, see help:

```bash
hide --help
```

## SDK

The SDK is a set of APIs and toolkits designed for coding agents to interact with the codebase. It is best used for creating new projects and implementing toolkits for coding agents.

We provide a Python package containing the SDK and some pre-built toolkits:

```bash
pip install hide-py
```

You can also implement your own toolkit by calling the Runtime's APIs directly (see [API Reference](api.md)).

For a quickstart on how to create new projects and interact with them, see the [Quickstart](quickstart.md) guide.
