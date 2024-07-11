# Quickstart

In this quickstart, we will demonstrate how to create new Hide projects and interact with them.

!!! note

    This quickstart assumes that you have already installed Hide and have it running on your local machine. If you haven't done so, please refer to the [Installation](installation.md) guide.

## Importing Hide and Creating a Client

To interact with the Hide server, we need to create a client. We can do this by importing `hide` and creating an instance of `hide.Client`:

```python
import hide
from hide.devcontainer.model import ImageDevContainer
from hide.model import Repository
from hide.toolkit import Toolkit

hide_client = hide.Client()
```

By default, the client will connect to the Hide server running on `http://localhost:8080`. If you have Hide running on a different host or port, you can specify it when creating the client:

```python
hide_client = hide.Client(base_url="https://my-hide-server:8080")
```

## Creating a Project

A project is a containerized development environment for a specific codebase. Creating a project involves cloning the repository and setting up a devcontainer. We can do this by calling the `create_project` method on the client:

```python
project = hide_client.create_project(
    Repository(url="https://github.com/artmoskvin/tiny-math-service.git")
)
```

The [Tiny Math Service](https://github.com/artmoskvin/tiny-math-service) is a simple Python service that performs basic mathematical operations. It has a devcontainer configuration file (`.devcontainer.json`) that is used to create a development environment for the project.

!!! note

    [Devcontainers](https://containers.dev/) is a specification for creating development environments.

If your project doesn't have a devcontainer configuration, you can define one using the `devcontainer` parameter:

```python
project = hide_client.create_project(
    repository=Repository(url="https://github.com/artmoskvin/tiny-math-service.git"),
    devcontainer=ImageDevContainer(
        image="mcr.microsoft.com/devcontainers/python:3.12-bullseye",
        onCreateCommand="pip install poetry && poetry install",
        customizations={
            "hide": {
                "tasks": [
                    {"alias": "test", "command": "poetry run pytest"},
                    {"alias": "run", "command": "poetry run uvicorn main:main"},
                ]
            }
        },
    )
)
```

Creating a project can take some time. Under the hood, Hide clones the repository, pulls the image, and installs the project dependencies.

## Using the Client

### Running Tasks

Having created a project, we can now interact with it using the Hide client. You could notice that the devcontainer configuration contains a `customizations` section that defines a custom task called `test`. We can use this task to run tests for our project:

```python
result = hide_client.run_task(project.id, alias="test")
print(result.stdOut)
```

The print statement will output the test results.

Aliases are convenient when referring to a frequently used command. Task API also allows us to run arbitrary shell commands in the project root by providing the command in the `command` parameter:

```python
result = hide_client.run_task(project.id, command="pwd")
print(result.stdOut)
```

This will print the path to the project root directory.

### Updating Files

TBA
