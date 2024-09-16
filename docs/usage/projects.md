# Projects

## Understanding Hide Projects

In Hide, a project represents a self-contained development environment for a specific codebase. Each project is built on top of a [devcontainer](https://containers.dev/), providing a consistent and isolated environment for coding agents to work in.

Key features of Hide projects:

1. **Devcontainer-based**: Each project runs in its own devcontainer, ensuring consistency across different machines and environments.
2. **GitHub Integration**: Currently, Hide supports creating projects from GitHub repositories. Support for local projects is coming soon.
3. **Flexible Configuration**: Projects can use a devcontainer.json file from the repository or accept container configuration as part of the project creation request.

!!! note

    For all code examples, the server is assumed to be running on `localhost:8080`. Adjust the URL if your Hide server is running on a different host or port.

## Creating a Project

To create a new project, you need to provide a GitHub repository URL. Hide will clone the repository, set up a devcontainer and associate it with the project id. The devcontainer configuration can come from two sources:

1. A `devcontainer.json` file in the repository (usually in the `.devcontainer` directory, for more options see [devcontainer.json specification](https://containers.dev/implementors/spec/#devcontainerjson)).
2. A configuration object in the project creation request.

If your repository contains a `devcontainer.json` file, omit the `devcontainer` field in the request. Hide will use the configuration from the repository. Otherwise provide configuration object in the `devcontainer` field of the project creation request. You can also use this field to override the existing `devcontainer.json` file.

### Using devcontainer.json

=== "curl"

    ```bash
    curl -X POST http://localhost:8080/projects \
      -H "Content-Type: application/json" \
      -d '{
        "repository": {
          "url": "https://github.com/your-username/your-repo.git",
        }
      }'
    ```

=== "python"

    ```python
    # Coming soon
    ```
### Using request body

=== "curl"

    ```bash
    curl -X POST http://localhost:8080/projects \
      -H "Content-Type: application/json" \
      -d '{
        "repository": {
          "url": "https://github.com/your-username/your-repo.git",
        },
        "devcontainer": {
          "name": "my-project",
          "image": "mcr.microsoft.com/devcontainers/python:3.10",
          "onCreateCommand": "pip install -r requirements.txt"
        }
      }'
    ```

=== "python"

    ```python
    # Coming soon
    ```

### Using commit hash

You can also specify a commit hash to checkout when cloning the repository. This is useful when you want your agent to work on a specific commit of a repository that is not the latest commit.

=== "curl"

    ```bash
    curl -X POST http://localhost:8080/projects \
      -H "Content-Type: application/json" \
      -d '{
        "repository": {
          "url": "https://github.com/your-username/your-repo.git",
          "commit": "your-commit-hash"
        }
      }'
    ```

=== "python"

    ```python
    # Coming soon
    ```

## Deleting a Project

Deleting a project will stop the project's devcontainer and delete the project.

To delete a project with id `123`:

=== "curl"


    ```bash
    curl -X DELETE http://localhost:8080/projects/123
    ```

=== "python"

    ```python
    # Coming soon
    ```

## Using images from Docker Hub

To use images from Docker Hub, you need to provide Docker Hub credentials when starting the server. You can do this by setting the `DOCKER_USER` and `DOCKER_TOKEN` environment variables.

```bash
export DOCKER_USER=your-docker-hub-username
export DOCKER_TOKEN=your-docker-hub-token
hide
```

You can also set these environment variables in the file and run Hide with the `env` flag:

```bash
touch .env
echo "DOCKER_USER=your-docker-hub-username" >> .env
echo "DOCKER_TOKEN=your-docker-hub-token" >> .env
hide -env .env
```

## Caveats

Hide tries to be as close to the devcontainer specification as possible. However some parts of the specification are not supported yet due to their complexity or ambiguity. For example:

- Devcontainer configurations based on Docker Compose are not supported yet.
- Devcontainer Features are not supported yet.
- Devcontainer image labels are not supported yet.

If you notice any other issues or have suggestions for improvements, please open an issue or submit a pull request on the [Hide repository](https://github.com/hide-org/hide).
