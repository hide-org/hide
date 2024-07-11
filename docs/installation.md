# Installation

A typical setup for Hide consists of 2 parts: a server that runs on a local or remote Docker host, and a client that interacts with it.

## Prerequisites

### Docker

If you don't have Docker installed, follow the instructions on the [Docker website](https://docs.docker.com/get-docker/) to install it.

## Server

The server is responsible for managing the development environments, executing tasks, and providing APIs for interacting with the codebase. It can be installed using [Homebrew](https://brew.sh/) or built from source.

### Using Homebrew

1. Add the Hide tap to your Homebrew:

    ```bash
    brew tap artmoskvin/hide
    ```

2. Install Hide using the brew install command:

    ```bash
    brew install hide
    ```

After installing Hide, you can start the server by running the following command:

```bash
hide
```

You should see logs indicating that the server is running, something like: `Server started on :8080`.

### Building from Source

To build Hide from source, follow these steps:

1. Ensure you have [Go 1.22](https://go.dev/) or later installed on your system.
2. Clone the Hide repository:

    ```bash
    git clone https://github.com/artmoskvin/hide.git
    cd hide
    ```

3. Build the project:

    ```bash
    make build
    ```

4. (Optional) Install Hide to your `$GOPATH/bin` directory:

    ```bash
    make install
    ```

After building from source, you can run Hide by running the following command from the project directory:

```bash
./hide
```

or if you've installed it to your `$GOPATH/bin`:

```bash
hide
```

You should see logs indicating that the server is running, something like: `Server started on :8080`.

## Client

The client is responsible for interacting with the server. It is best used for creating new projects and implementing toolkits for coding agents.

We provide a Python package containing the client and some pre-built toolkits:

```bash
pip install hide
```

You can also implement your own client by calling the server's APIs directly (see [API Reference](api.md)).

For a quickstart on how to create new projects and interact with them, see the [Quickstart](quickstart.md) guide.
