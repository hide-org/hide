# Search

Hide offers powerful search capabilities to help you navigate and explore your projects efficiently. There are three main types of search available:

1. Content Search
2. File Search
3. Symbol Search

Let's dive into each of these search types and see how you can leverage them in your projects.

## Content Search

Content search allows you to find specific text patterns within your project files. This is incredibly useful when you need to locate particular code snippets, comments, or any textual content across your entire project.

### Basic Usage

To perform a content search:

=== "curl"

    ```bash
    curl -X GET "http://localhost:8080/projects/{projectId}/search?query=your_search_query"
    ```

=== "python"

    ```python
    # Coming soon
    ```

### Advanced Options

Hide's content search supports different search types:

- **Default**: Case-insensitive search
- **Exact**: Case-sensitive, exact match search
- **Regex**: Regular expression search

You can specify the search type using query parameters:

=== "curl"

    ```bash
    # Exact match
    curl -X GET "http://localhost:8080/projects/{projectId}/search?query=YourExactPhrase&exact"

    # Regex search
    curl -X GET "http://localhost:8080/projects/{projectId}/search?query=Your.*Regex&regex"
    ```

=== "python"

    ```python
    # Coming soon
    ```

## File Search

File search helps you find files within your project based on their names or paths. This is particularly useful when you're looking for specific files or want to filter files based on certain patterns.

### Basic Usage

To search for files:

=== "curl"

    ```bash
    curl -X GET "http://localhost:8080/projects/{projectId}/files"
    ```

=== "python"

    ```python
    # Coming soon
    ```

### Filtering Files

You can use `include` and `exclude` parameters to filter the search results:

=== "curl"

    ```bash
    # Include only Python files
    curl -X GET "http://localhost:8080/projects/{projectId}/files?include=*.py"

    # Exclude test files
    curl -X GET "http://localhost:8080/projects/{projectId}/files?exclude=test_*.py"

    # Combine include and exclude
    curl -X GET "http://localhost:8080/projects/{projectId}/files?include=*.py&exclude=test_*.py"
    ```

=== "python"

    ```python
    # Coming soon
    ```

## Symbol Search

Symbol search allows you to find specific symbols (like functions, classes, or variables) within your project. This is extremely helpful when you're trying to locate specific code elements without knowing their exact file location.

### Basic Usage

To search for symbols:

=== "curl"

    ```bash
    curl -X GET "http://localhost:8080/projects/{projectId}/symbols?query=your_symbol_name"
    ```

=== "python"

    ```python
    # Coming soon
    ```

### Advanced Options

You can customize the symbol search with additional parameters:

- `limit`: Specify the maximum number of results (default is 10, max is 100)
- Include or exclude specific symbol types

=== "curl"

    ```bash
    # Limit results to 20
    curl -X GET "http://localhost:8080/projects/{projectId}/symbols?query=your_symbol_name&limit=20"

    # Include only functions and classes (example, actual parameters may vary)
    curl -X GET "http://localhost:8080/projects/{projectId}/symbols?query=your_symbol_name&include=function&include=class"
    ```

=== "python"

    ```python
    # Coming soon
    ```

## Tips for Effective Searching

1. **Use specific queries**: The more specific your search query, the more accurate your results will be.
2. **Leverage regex**: For complex search patterns, use regex in content search.
3. **Combine search types**: Use file search to narrow down the scope, then use content search within those files.
4. **Utilize symbol search**: When looking for specific code elements, symbol search can be faster than content search.

By mastering these search capabilities, you'll be able to navigate your projects with ease and efficiency, making your development process smoother and more productive.
