# Langchain Tutorial

In this tutorial, we will use the Langchain framework to build a simple coding agent that can solve coding problems based on a given prompt. We will demonstrate how to use Hide to create a development environment, and how an agent can interact with it using the Hide toolkit. 

For this tutorial, we will use the [Tiny Math Service](https://github.com/artmoskvin/tiny-math-service) as our example project. This project is a simple Python service that performs basic mathematical operations. It is capable of performing addition, subtraction, multiplication, and division operations. It also has some tests that we can use to assert the code quality of the agent.

We choose OpenAI as the LLM provider for this tutorial, but you can use any other provider as well.

## Setting Up the Environment

To get started, you need to have the following tools installed:

- Docker
- Hide

We will use Python to build our agent, so make sure you have Python installed. We highly recommend creating a new virtual environment for this tutorial to avoid any dependency issues.

Also, make sure you have the OpenAI API key ready. You can get it [here](https://platform.openai.com/api-keys). You can either set it as an environment variable or a constant in your code.

### Installing Docker

If you don't have Docker installed, follow the instructions on the [Docker website](https://docs.docker.com/get-docker/) to install it.

### Installing Hide

To install Hide, run the following command:

```bash
brew install hide
```

For more installation options, refer to the [installation guide](../installation.md).

### Starting the Hide Server

Once you have Hide installed, you can start the server by running the following command:

```bash
hide
```

This will start the server on `http://localhost:8080`.

### Installing Python Dependencies

We will use Hide and the Langchain framework to build our agent. Let's install the necessary dependencies.

```bash
pip install langchain hide-py
```

## Creating a Project

Let's create a new project for our agent. Create a new Python file, e.g., `main.py`, and add the following code:

```python
import os

from hide.client.hide_client import HideClient
from hide.langchain.toolkit import HideToolkit
from langchain import hub
from langchain.agents import AgentExecutor, create_tool_calling_agent
from langchain_openai import ChatOpenAI

OPENAI_API_KEY = "ENTER YOUR KEY"
HIDE_BASE_URL = "http://localhost:8080"
PROJECT_GIT_URL = "https://github.com/artmoskvin/tiny-math-service.git"

if "OPENAI_API_KEY" not in os.environ:
    os.environ["OPENAI_API_KEY"] = OPENAI_API_KEY

hide_client = HideClient(base_url=HIDE_BASE_URL)
project = hide_client.create_project(url=PROJECT_GIT_URL)

print(f"Project ID: {project.id}")
```

Here, few things are happenning. First, we add all the required imports. Second, we define all the necessary constants and set the OpenAI API key as an environment variable if it is not already set. Finally, we create a Hide client and use it to create a new project.

Creating a project can take some time. Under the hood, Hide clones the repository and sets up a devcontainer using the configuration from the repository. This process can take a few minutes, so be patient.

## Building the Agent

Now that we have our project created, we can start building our agent. First, let's create a toolkit for our agent. Add the following code to the `main.py` file:

```python
toolkit = HideToolkit(project_id=project.id, hide_client=hide_client)
tools = toolkit.get_tools()

for tool in tools:
    print("Name:", tool.name)
    print("Description:", tool.description)
    print("Args:", tool.args)
    print("")
```

This code creates a Langchain toolkit for our agent and prints out the tool details for illustration purposes.

Next, let's create an agent. Add the following code to the `main.py` file:

```python
llm = ChatOpenAI(model="gpt-4o")
prompt = hub.pull("hwchase17/openai-tools-agent")
agent = create_tool_calling_agent(llm, tools, prompt)
agent_executor = AgentExecutor(agent=agent, tools=tools, verbose=True)
```

Now, our agent is ready and we can start testing it :tada:

## Simple Questions

Let's test our agent by asking it a simple question. For example, in the [devcontainer configuration](https://github.com/artmoskvin/tiny-math-service/blob/main/.devcontainer.json) for Tiny Math Service, we defined several tasks, one of which is for running tests. Agents can run tasks using the Hide toolkit. Let's ask our agent to run the tests and see if it can figure out how to do it:

```python
response = agent_executor.invoke({"input": "Run the tests for the math service"})
print("")
print(response["output"])
```

Using the Hide toolkit, the agent will list all the available tools, pick the one that matches the task, and run it. Sometimes, the agent tries to guess which task to run which can lead to a failure but the agent will recover by checking all the available tools and calling the right one. 

## Advanced Questions

Now let's try something real. Let's ask our agent to add a new endpoint that calculates the exponent of two numbers. Additionally, we want the endpoint to be covered by tests. The agent should be able to figure out how to do this. Our prompt will be more complicated here and include more detailed instructions but if you have ever worked with the coding agents none of this should come as a surprise to you.

!!! note

    We are calling the file names explicitly because the search functionality is not yet implemented in Hide. This will be fixed soon.

```python
prompt = """\
You are a helpful AI assistant.
Update the source code file following the instructions.
The user can't modify your code. So do not suggest incomplete code which requires users to modify.
Make sure that no comments and empty lines are removed from the file.
Check out the new exponentiation endpoint in the `my_tiny_service/api/routers/maths.py` file and add the tests for it in the `tests/test_api.py` file.
Run the tests and make sure they pass. If the tests fail, fix them until they pass.
"""
response = agent_executor.invoke({"input": prompt})
print("")
print(response["output"])
```

This task will require the agent to use multiple tools from the Hide toolkit. The agent will first read the content of the `maths.py` and `test_api.py` files, then update them according to the instructions, and finally run the tests. If the tests fail the agent will try to fix them. It can take few rounds but eventually the root cause of the problem will be fixed.

## Conclusion

In this tutorial, we demonstrated how to use the Langchain framework to build a simple coding agent that can solve coding problems based on a given prompt. We used Hide to create a development environment and the Hide toolkit to interact with it. Check out the [SWE-agent tutorial](./swe-agent.md) to learn how you can build a custom agent, like the SWE-agent, for your coding projects.
