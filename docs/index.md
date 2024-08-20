# Welcome to Hide
<div style="display: flex; justify-content: center;">
  <img src="assets/hide-quickstart.png" alt="Hide Quickstart"/>
</div>

Hide is a headless IDE for coding agents. It provides containerized development environments for codebases and exposes APIs for agents to interact with them.

## Features

- **Development Containers**: Isolated development environments for coding agents.
- **CodeGen Toolkit**: Designed for coding agents to interact with the codebase.
    - **File API**: Read, write files. Apply diffs. Error highlighting.
    - **Search API**: Search for files, content, and code symbols.
    - **Git API**: Pull, commit, and push changes to remote git repositories.
    - **Task API**: Run tests, linters, formatters, and other development tools.
- **Code Analysis**: Feedback from linters, formatters, and compilers.
- **Agents Integration**: Bring your own agent and create a custom toolkit from Hide APIs or use Hide's pre-built toolkits for popular frameworks.

## How it works

<figure markdown="span">
  ![Hide How It Works](assets/hide-how-it-works-light.png#only-light){ width="70%" }
  ![Hide How It Works](assets/hide-how-it-works-dark.png#only-dark){ width="70%" }
</figure>

Hide consists of two main components: Runtime and SDK.

### Runtime

Runtime is the backend system responsible for managing development containers and executing tasks. It can be run on any Docker host, whether it's your local machine or a remote server, providing flexibility in deployment options. 

Key responsibilities of Hide Runtime:

- Creating and managing containerized development environments
- Executing tasks within these environments
- Providing APIs for interacting with codebases

### SDK

SDK is a set of APIs and toolkits designed for coding agents to interact with the codebase. SDK simplifies the process of building custom toolkits for agents and offers pre-built toolkits to get started.

Key aspects of Hide SDK:

- Provides a high-level abstraction layer for Runtime's APIs
- Offers pre-built toolkits for popular frameworks and languages, which can be further customized
- Simplifies the process of building custom toolkits
