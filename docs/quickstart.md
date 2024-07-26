# Quickstart

In this quickstart, we will demonstrate how to create new Hide projects and interact with them.

!!! note

    This quickstart assumes that you have already installed Hide and have it running on your local machine. If you haven't done so, please refer to the [Installation](installation.md) guide.

## Importing Hide and Creating a Client

To interact with the Hide server, we need to create a client. We can do this by importing `hide` and creating an instance of `hide.Client`:

```python
import hide
from hide.devcontainer.model import ImageDevContainer
from hide.model import FileUpdateType, UdiffUpdate, Repository
from hide.toolkit import Toolkit

hide_client = hide.Client()
```

By default, the client will connect to the Hide server running on `http://localhost:8080`. If you have Hide running on a different host or port, you can specify it when creating the client:

```python
hide_client = hide.Client(base_url="https://my-hide-server:8081")
```

## Creating a Project

A project is a containerized development environment for a specific codebase. Creating a project consists of cloning the repository and setting up a devcontainer. We can do this by calling the `create_project` method on the client and passing a URL of the project on GitHub:

```python
project = hide_client.create_project(
    Repository(url="https://github.com/artmoskvin/tiny-math-service.git")
)
```

Here, we use the [Tiny Math Service](https://github.com/artmoskvin/tiny-math-service) which is a simple Python service that performs basic mathematical operations. It has a devcontainer configuration file (`.devcontainer.json`) that is used to create a development environment for the project.

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

Creating a project can take some time. Under the hood, Hide clones the repository, pulls or builds the image, starts the container and installs the project dependencies.

## Using the Client

### Running Tasks

Having created a project, we can now start working with it. You could notice that the devcontainer [configuration](https://github.com/artmoskvin/tiny-math-service/blob/main/.devcontainer.json) for the [Tiny Math Service](https://github.com/artmoskvin/tiny-math-service) contains a `customizations` section that defines a custom task called `test`. We can use this task to run tests for our project:

```python
result = hide_client.run_task(
    project_id=project.id, 
    alias="test"
)

print(result.stdOut)
```

The print statement will output the test results.

Aliases are convenient when referring to frequently used commands. Running tasks is powered by Task API which also allows us to run arbitrary shell commands by providing the `command` parameter:

```python
result = hide_client.run_task(
    project_id=project.id, 
    command="pwd"
)

print(result.stdOut)
```

The tasks are executed from the project root so the print statement will output the path to the project root directory.

### Reading and Updating Files

We can also read and update files in the project. For example, we can read the `maths.py` file and add a new endpoint in it. First, let's read the file:

```python
result = hide_client.get_file(
    project_id=project.id,
    path="my_tiny_service/api/routers/maths.py"
)

print(result.content)
```

This will print the content of the `maths.py` file. 

Coding agents often update files by adding/replacing lines or by applying unified diffs. Therefore it can be important to include line numbers when reading files. With Hide, we can include line numbers when reading files by setting the `show_line_numbers` parameter to `True`:

```python
result = hide_client.get_file(
    project_id=project.id,
    path="my_tiny_service/api/routers/maths.py", 
    show_line_numbers=True
)

print(result.content)
```

This will print the content of the `maths.py` file with line numbers.

By default, Hide returns the first 100 lines of the file. We can change this by setting the `start_line` and `num_lines` parameters:

```python
result = hide_client.get_file(
    project_id=project.id,
    path="my_tiny_service/api/routers/maths.py",
    show_line_numbers=True,
    start_line=10,
    num_lines=200,
)

print(result.content)
```

This will print the content of the `maths.py` file with 200 lines starting from line 10.

Updating files can be done in three ways: by replacing the entire file, by updating lines, or by applying unified diffs. For this quickstart we will use the unified diff option:

```python

patch = """\
--- a/my_tiny_service/api/routers/maths.py
+++ b/my_tiny_service/api/routers/maths.py
@@ -113,3 +113,17 @@
             status_code=starlette.status.HTTP_400_BAD_REQUEST,
             detail="Division by zero is not allowed",
         ) from e
+
+
+@router.post(
+    "/exp",
+    summary="Calculate the exponent of two numbers",
+    response_model=MathsResult,
+)
+def exp(maths_input: MathsIn) -> MathsResult:
+    \"\"\"Calculates the exponent of two whole numbers.\"\"\"
+    return MathsResult(
+        **maths_input.dict(),
+        operation="exp",
+        result=maths_input.number1 ** maths_input.number2,
+    )
"""

result = hide_client.update_file(
    project_id=project.id, 
    path='my_tiny_service/api/routers/maths.py',
    type=FileUpdateType.UDIFF,
    update=UdiffUpdate(patch=patch)
)

print(result.content)
```

This will apply the unified diff to the file and return the updated content.

For more information on all the available update types and their parameters, see the [Files API](usage/files.md#updating-a-file) documentation.

