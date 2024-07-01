# Installation

Hide can be installed using [Homebrew](https://brew.sh/) or built from source.

## Homebrew Installation

1. Add the Hide tap to your Homebrew:

    ```bash
    brew tap artmoskvin/hide
    ```

2. Install Hide using the brew install command:

    ```bash
    brew install hide
    ```

## Building from Source

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

4. (Optional) Install Hide to your $GOPATH/bin directory:

    ```bash
    make install
    ```

After building from source, you can run Hide directly using `./hide` from the project directory, or hide if you've installed it to your `$GOPATH/bin`.
