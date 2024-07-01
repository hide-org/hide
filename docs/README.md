<div style="display: flex; justify-content: center;">
<pre style="color: rgb(241, 250, 152); background-color: rgb(41, 42, 53); line-height: 1.3;">
          _____                    _____                    _____                    _____
         /\    \                  /\    \                  /\    \                  /\    \
        /::\____\                /::\    \                /::\    \                /::\    \
       /:::/    /                \:::\    \              /::::\    \              /::::\    \
      /:::/    /                  \:::\    \            /::::::\    \            /::::::\    \
     /:::/    /                    \:::\    \          /:::/\:::\    \          /:::/\:::\    \
    /:::/____/                      \:::\    \        /:::/  \:::\    \        /:::/__\:::\    \
   /::::\    \                      /::::\    \      /:::/    \:::\    \      /::::\   \:::\    \
  /::::::\    \   _____    ____    /::::::\    \    /:::/    / \:::\    \    /::::::\   \:::\    \
 /:::/\:::\    \ /\    \  /\   \  /:::/\:::\    \  /:::/    /   \:::\ ___\  /:::/\:::\   \:::\    \
/:::/  \:::\    /::\____\/::\   \/:::/  \:::\____\/:::/____/     \:::|    |/:::/__\:::\   \:::\____\
\::/    \:::\  /:::/    /\:::\  /:::/    \::/    /\:::\    \     /:::|____|\:::\   \:::\   \::/    /
 \/____/ \:::\/:::/    /  \:::\/:::/    / \/____/  \:::\    \   /:::/    /  \:::\   \:::\   \/____/
          \::::::/    /    \::::::/    /            \:::\    \ /:::/    /    \:::\   \:::\    \
           \::::/    /      \::::/____/              \:::\    /:::/    /      \:::\   \:::\____\
           /:::/    /        \:::\    \               \:::\  /:::/    /        \:::\   \::/    /
          /:::/    /          \:::\    \               \:::\/:::/    /          \:::\   \/____/
         /:::/    /            \:::\    \               \::::::/    /            \:::\    \
        /:::/    /              \:::\____\               \::::/    /              \:::\____\
        \::/    /                \::/    /                \::/____/                \::/    /
         \/____/                  \/____/                  ~~                       \/____/
</pre>
</div>

Hide is a headless IDE for coding agents. It provides scalable development environments with comprehensive tools for enhanced code generation quality.

# Features

## Development Containers

Hide leverages [devcontainers](https://containers.dev/) to create consistent and isolated development environments for agents.

- **Parallelization**: Run multiple agent workloads simultaneously.
- **Flexibility**: Execute environments both locally and in the cloud.
- **Consistency**: Ensure uniform development setups across different machines.

## CodeGen Toolkit

Our comprehensive toolkit is specifically designed for coding agents to interact with the codebase efficiently.

### File API

Enables agents to manipulate files with precision:

- Read and write files, including specific line numbers.
- Edit files using unified diffs.
- Receive helpful error messages for better error recovery.

Example usage:
```python
# Coming soon
```

### Search API

Powerful search capabilities for improved codebase navigation:

- Search for project files, content, and code symbols (functions, classes, etc.).
- Retrieve symbol usage information.

Example usage:
```python
# Coming soon
```

### Git API

Seamless integration with version control:

- Pull, commit, and push changes to remote git repositories.
- Manage branches and resolve conflicts.

Example usage:
```python
# Coming soon
```

### Task API

Execute various development tasks:

- Run tests, linters, formatters, and other development tools.
- Execute shell commands within the development environment.

Example usage:
```python
# Coming soon
```

### Code Analysis Integration

Improve code generation quality with integrated analysis tools:

- Receive feedback from linters, formatters, and compilers.
- Enable agents to self-correct and learn from mistakes.

### Framework Support

Hide is designed to work with major agent development frameworks:

- [Langchain](https://www.langchain.com/)
- [Autogen](https://microsoft.github.io/autogen/)

