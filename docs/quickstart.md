# Quickstart

In this quickstart, we will demonstrate how to create new Hide projects and let agents interact with them. We will explore Hide APIs and build a real coding agent using Langchain.

This quickstart assumes that you have already installed Hide Runtime and SDK, and have Runtime running on your local machine. If you haven't done so, please refer to the [Installation](installation.md) guide.

## Requirements

- Python 3.10+
- Hide Runtime running on your local machine
- Hide SDK installed (`pip install hide-py`)

## Creating a Client

To interact with the Hide Runtime, we need to create a client. We can do this by importing `hide` and creating an instance of `hide.Client`:

```python
import hide
from hide import model

hc = hide.Client()
```

By default, the client will connect to the runtime running on `http://localhost:8080`. If you have Hide running on a different host or port, you can specify it when creating the client:

```python
hc = hide.Client(base_url="https://my-hide-runtime:8081")
```

## Creating a Project

A project is a containerized development environment for a specific codebase. Creating a project consists of cloning the repository, setting up a devcontainer, and initializing the development environment. We can do this by calling the `create_project` method on the client and passing a URL of the project on GitHub:

```python
project = hc.create_project(
    repository=model.Repository(url="https://github.com/artmoskvin/tiny-math-service.git")
)
```

Here, we use the [Tiny Math Service](https://github.com/artmoskvin/tiny-math-service) which is a simple Python service performing basic mathematical operations. It has a devcontainer configuration file (`.devcontainer.json`) that is used to create a development environment for the project.

!!! note

    [Devcontainers](https://containers.dev/) is a specification for creating reproducible development environments.

If your repository doesn't have a devcontainer configuration, you can define one as part of the project creation request using the `devcontainer` parameter:

```python
from hide.devcontainer.model import ImageDevContainer

project = hc.create_project(
    repository=model.Repository(url="https://github.com/artmoskvin/tiny-math-service.git"),
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

Before we jump to the coding agents, let's take a look at how you can interact with the project created in the previous section. This will help you develop an intuition for how the Hide APIs work and how agents can interact with them.

### Running Tasks

You could notice that the devcontainer [configuration](https://github.com/artmoskvin/tiny-math-service/blob/main/.devcontainer.json) for the [Tiny Math Service](https://github.com/artmoskvin/tiny-math-service) contains a `customizations` section that defines a custom task called `test`. We can use this alias to run tests in our project:

```python
result = hc.run_task(
    project_id=project.id, 
    alias="test"
)

print(result.stdout)
# ============================= test session starts ==============================
# platform linux -- Python 3.12.5, pytest-8.0.1, pluggy-1.4.0
# rootdir: /workspace
# plugins: anyio-4.3.0
# collected 3 items
# 
# tests/test_api.py ...                                                    [100%]
# ======================== 3 passed, 5 warnings in 0.05s =========================
```

Running tasks is powered by the Task API which also allows us to run arbitrary shell commands by providing the `command` parameter:

```python
result = hc.run_task(
    project_id=project.id, 
    command="pwd"
)

print(result.stdout)
# /workspace
```

The tasks are executed from the project root so the print statement outputs the path to the project root directory.

### Reading and Updating Files

We can also read and update files in the project. For example, let's read the `maths.py` file and add a new endpoint in it. First, let's read the file:

```python
file = hc.get_file(
    project_id=project.id,
    path="my_tiny_service/api/routers/maths.py"
)

print(file)
#  1 | """Endpoint examples with input/output models and error handling."""
#  2 | import logging
#  3 | 
#  4 | import fastapi
#  5 | import pydantic
#  6 | import starlette.status
#  7 | 
#  8 | router = fastapi.APIRouter()
#... | ...
#112 |         raise fastapi.HTTPException(
#113 |             status_code=starlette.status.HTTP_400_BAD_REQUEST,
#114 |             detail="Division by zero is not allowed",
#115 |         ) from e
```

The file includes line numbers which can be useful when coding agents update files. Updating files can be done in three ways: by replacing the entire file, by manipulating lines, or by applying unified diffs.

!!! note

    The [Unified Diff](https://en.wikipedia.org/wiki/Diff_utility#Unified_format) is a format for comparing two files or versions of a file.

Let's see how the unified diff works:

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
+        result=maths_input.number1 ** maths_input.number,
+    )
"""

file = hc.update_file(
    project_id=project.id, 
    path='my_tiny_service/api/routers/maths.py',
    update=model.UdiffUpdate(patch=patch)
)

print(file)
#  1 | """Endpoint examples with input/output models and error handling."""
#  2 | import logging
#  3 | 
#  4 | import fastapi
#  5 | import pydantic
#  6 | import starlette.status
#  7 | 
#  8 | router = fastapi.APIRouter()
#... | ...
#123 | def exp(maths_input: MathsIn) -> MathsResult:
#124 |     """Calculates the exponent of two whole numbers."""
#125 |     return MathsResult(
#126 |         **maths_input.dict(),
#127 |         operation="exp",
#128 |         result=maths_input.number1 ** maths_input.number,
#                                                        ^^^^^^ Error: Cannot access attribute "number" for class "MathsIn"
#  Attribute "number" is unknown
#
#129 |     )
```

It turns out there was a typo in my patch but Hide noticed it and highlighted the line with the error. Like a normal IDE, Hide runs continuous diagnostics on the code using LSP servers and highlights errors. Currently, Hide provides diagnostics for Python, JavaScript, TypeScript, and Go, and we can add more languages if needed. Let us know in the GitHub Issues if you need support for other languages.

!!! note

    The [Language Server Protocol (LSP)](https://microsoft.github.io/language-server-protocol/) defines the protocol used between an editor or IDE and a language server that provides language features like auto complete, go to definition, find all references etc.

For more information on all the available update types and their parameters, see the [Files API](usage/files.md#updating-a-file) documentation.

## Using the Toolkit

Finally, let's take a look at how we can use the Hide toolkit to build a coding agent. For this quickstart, we will use the [Langchain](https://www.langchain.com/langchain) framework to build a simple coding agent that can solve coding problems based on a given prompt.

First, let's initialize the toolkit:

```python
from hide.toolkit import Toolkit

toolkit = Toolkit(project=project, client=hc)
lc_toolkit = toolkit.as_langchain()
```

The toolkit is a collection of pre-built tools that can be used by agent to solve coding problems. Let's take a look at the tools available in the toolkit:

```python
for tool in lc_toolkit.get_tools():
    print("Name:", tool.name)
    print("Description:", tool.description)
    print("Args:", tool.args)
    print("")

# Name: append_lines
# Description: append_lines(path: str, content: str) -> str - Append lines to a file in the project.
# Args: {'path': {'title': 'Path', 'type': 'string'}, 'content': {'title': 'Content', 'type': 'string'}}
# 
# ...
# 
# Name: run_task
# Description: run_task(command: Optional[str] = None, alias: Optional[str] = None) -> str - Run a task in the project. Provide either command or alias. Command will be executed in the shell.
#         For the list of available tasks and their aliases, use the `get_tasks` tool.
# Args: {'command': {'title': 'Command', 'type': 'string'}, 'alias': {'title': 'Alias', 'type': 'string'}}
```

Now, let's create an agent using the OpenAI's GPT-4o model. Make sure to replace `YOUR_OPENAI_API_KEY` with your actual OpenAI API key.

```python
from langchain import hub
from langchain.agents import AgentExecutor, create_tool_calling_agent
import os

os.environ["OPENAI_API_KEY"] = "YOUR_OPENAI_API_KEY"

from langchain_openai import ChatOpenAI

llm = ChatOpenAI(model="gpt-4o", seed=128)
prompt = hub.pull("hwchase17/openai-tools-agent")
tools = lc_toolkit.get_tools()

agent = create_tool_calling_agent(llm, tools, prompt)
agent_executor = AgentExecutor(agent=agent, tools=tools, verbose=True)
```

With the agent created, we can prompt it to add the tests for the new endpoint that we created earlier:

```python
prompt = """\
I created the new exponentiation endpoint in the `my_tiny_service/api/routers/maths.py` file. Could you add the tests for it in the `tests/test_api.py` file?
Run the tests and make sure they pass. If the tests fail, fix them until they pass.
"""

response = agent_executor.invoke({"input": prompt})

print(response["output"])
# > Entering new AgentExecutor chain...
# 
# Invoking: `get_file` with `{'path': 'my_tiny_service/api/routers/maths.py'}`
# 
# ...
# 
# > Finished chain.
# 
# All tests, including the new one for exponentiation, have passed.
```

This task will require the agent to use multiple tools from the Hide toolkit. The agent will first read the content of the `maths.py` and `test_api.py` files, then update them according to the instructions, and finally run the tests. If the tests fail the agent will try to fix them. It can take few rounds but eventually the agent will succeed.

## Next Steps

In this quickstart, we demonstrated how to create new Hide projects and let agents interact with them. For more details on how to use Hide, check out our [Guides](./usage/index.md) and [Tutorials](./tutorials/index.md).
