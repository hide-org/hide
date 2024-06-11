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
- **Development containers**. Hide uses [devcontainers](https://containers.dev/) to create consistent and isolated development environments for agents. With devcontainers you can parallelize agent workloads and run them both on the local host and in the cloud. 
- **CodeGen Toolkit**. Hide provides a toolkit designed specifically for coding agents to interact with the codebase. The toolkit contains the tools for code editing, task running, searching and git operations.
  - **File API**. Allows agents to read and write files, including line numbers if needed, edit files with unified diffs, and receive helpful error messages to recover from failures.
  - **Search API**. Enables agents to search for project files, content, and code symbols like functions and classes. Also provides symbol usage information for better navigation.
  - **Git API**. Allows agents to pull, commit, and push changes to the remote git repository.
  - **Task API**. Enables agents to run tasks and shell commands. Tasks can be used for testing, linting, formatting, and more.
- **Code analysis**. Hide provides agents with the feedback from the code analysis tools, such as linters, formatters, and compilers, which improves the code generation quality and allows agents to recover from mistakes.
- **Framework Support**. Hide supports some of the major frameworks for building agents by providing toolkits for them. If frameworks is not your thing you can use the API directly. 

# Installation

Hide can be installed using Homebrew:

1. Install [Homebrew](https://brew.sh/) if you don't have it installed.

2. Add the Hide tap to your Homebrew:
   ```bash
   brew tap artmoskvin/hide
   ```
2. Install Hide using the brew install command:
   ```bash
   brew install hide
   ```

# Usage (Work in Progress)
This section will provide detailed instructions on how to use Hide once the documentation is complete.
