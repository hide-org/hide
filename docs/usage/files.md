# Files

## Understanding Files in Hide

In Hide, the Files API allows coding agents to interact with the project's file system within the devcontainer environment. This enables operations such as creating, reading, updating, and deleting files. All file paths are relative to the project's root directory.

!!! note

    For all code examples, the server is assumed to be running on `localhost:8080`. Adjust the URL if your Hide server is running on a different host or port.

!!! note

    For all requests, replace `{project_id}` with your actual project ID.

### Creating a File

To create a new file in your project:

=== "curl"

    ```bash
    curl -X POST http://localhost:8080/projects/{project_id}/files \
         -H "Content-Type: application/json" \
         -d '{"path": "example.txt", "content": "Hello, World!"}'
    ```

=== "python"

    ```python
    # Coming soon
    ```

This will create a file named `example.txt` in your project's root directory with the content `Hello, World!`.

### Listing Files

To list all files in your project:

=== "curl"

    ```bash
    curl http://localhost:8080/projects/{project_id}/files
    ```

=== "python"

    ```python
    # Coming soon
    ```

This will return a list of all files recursively in your project's root directory.

### Reading a File

To read the contents of a specific file:

=== "curl"

    ```bash
    curl http://localhost:8080/projects/{project_id}/files/example.txt
    ```

=== "python"

    ```python
    # Coming soon
    ```

This will return the contents of the file `example.txt` in your project's root directory.

Reading files supports different parameters such as specifying a range of lines, or including the line numbers in the response. To include the line numbers, set the `showLineNumbers` parameter to `true`:

=== "curl"

    ```bash
    curl http://localhost:8080/projects/{project_id}/files/example.txt?showLineNumbers=true
    ```

=== "python"

    ```python
    # Coming soon
    ```

To specify a range of lines, set the `startLine` and `numLines` parameters:

=== "curl"

    ```bash
    curl http://localhost:8080/projects/{project_id}/files/example.txt?startLine=10&numLines=20
    ```

=== "python"

    ```python
    # Coming soon
    ```

### Updating a File

Updating files can be done in three ways: by replacing the entire file, by updating lines, or by applying unified diffs. We will look at each of these in the next sections.

#### Replacing the entire file

To replace the entire file, we can use the update type `overwrite` and provide the new content as the `content` parameter:

=== "python"

    ```python
    from hide.model import FileUpdateType, OverwriteUpdate

    result = hide_client.update_file(
        project_id="my-project",
        path="path/to/file.py",
        type=FileUpdateType.OVERWRITE,
        update=OverwriteUpdate(
            content="def hello_world():\n    print('Hello, World!')\n"
        )
    )
    ```

=== "curl"

    ```bash
    curl -X PUT http://localhost:8080/projects/my-project/files/path/to/file.py \
         -H "Content-Type: application/json" \
         -d '{
            "type": "overwrite",
            "overwrite": {
                "content": "def hello_world():\n    print('Hello, World!')\n"
            }
        }'
    ```

This will replace the entire file with the new content.

#### Updating lines

To update lines, we can use the update type `linediff` and provide the line diff as the `lineDiff` parameter:

=== "python"

    ```python
    from hide.model import FileUpdateType, LineDiffUpdate

    result = hide_client.update_file(
        project_id="my-project",
        path="path/to/file.py",
        type=FileUpdateType.LINEDIFF,
        update=LineDiffUpdate(
            start_line=1,
            end_line=2,
            content="def hello_world():\n    print('Hello, World!')\n"
        ),
    )
    ```

=== "curl"

    ```bash
    curl -X PUT http://localhost:8080/projects/my-project/files/path/to/file.py \
         -H "Content-Type: application/json" \
         -d '{
            "type": "linediff",
            "lineDiff": {
                "startLine": 1,
                "endLine": 2,
                "content": "def hello_world():\n    print('Hello, World!')\n"
            }
        }'
    ```

This will update the lines from line 1 to 3 with the new content.

#### Applying unified diffs

To apply unified diffs, we can use the update type `udiff` and provide the patch as the `patch` parameter:

=== "python"

    ```python
    from hide.model import FileUpdateType, UdiffUpdate

    result = hide_client.update_file(
        project_id="my-project",
        path="path/to/file.py",
        type=FileUpdateType.UDIFF,
        udiff=UdiffUpdate(
            patch="""--- path/to/file.py
    +++ path/to/file.py
    @@ -1,2 +1,2 @@
     def hello_world():
    -    print('Hello, World!')
    +    print('Hello, World!!!')
    """
        )
    )
    ```

=== "curl"

    ```bash
    curl -X PUT http://localhost:8080/projects/my-project/files/path/to/file.py \
         -H "Content-Type: application/json" \
         -d '{
            "type": "udiff",
            "udiff": {
                "patch": """--- path/to/file.py
        +++ path/to/file.py
        @@ -1,2 +1,2 @@
         def hello_world():
        -    print('Hello, World!')
        +    print('Hello, World!!!')
        """
            }
        }'
    ```

This will apply the unified diff to the file and update the lines accordingly.

### Deleting a File

To delete a specific file:

=== "curl"

    ```bash
    curl -X DELETE http://localhost:8080/projects/{project_id}/files/example.txt
    ```

=== "python"

    ```python
    # Coming soon
    ```

This will delete the file `example.txt` in your project's root directory.

## Error Handling

The API uses standard HTTP status codes to indicate the success or failure of requests:

- 200: Successful operation
- 404: File or project not found
- 400: Bad request (e.g., invalid input)
- 500: Internal server error

Always check the status code and response body for detailed error messages.
