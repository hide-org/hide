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

We are working on supporting different parameters for reading files, such as specifying a range of lines, or including the line numbers in the response. Stay tuned! :blush:

### Updating a File

To update the contents of an existing file:

=== "curl"

    ```bash
    curl -X PUT http://localhost:8080/projects/{project_id}/files/example.txt \
         -H "Content-Type: application/json" \
         -d '{"content": "Updated content!"}'
    ```

=== "python"

    ```python
    # Coming soon
    ```

This will update the contents of the file `example.txt` in your project's root directory to `Updated content!`. This API call will effectively overwrite the existing file.

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

### :construction: Applying Diffs

Coming soon

## Error Handling

The API uses standard HTTP status codes to indicate the success or failure of requests:

- 200: Successful operation
- 404: File or project not found
- 400: Bad request (e.g., invalid input)
- 500: Internal server error

Always check the status code and response body for detailed error messages.
