# Contributing to Hide

We're excited that you're interested in contributing to Hide! This document outlines the process for contributing to this project and provides some guidelines to ensure a smooth collaboration.

## Getting Started

1. Fork the repository on GitHub.
2. Clone your fork locally:

   ```bash
   git clone https://github.com/your-username/hide.git
   cd hide
   ```

3. Create a new branch for your contribution:

   ```bash
   git checkout -b your-branch-name
   ```

## Making Changes

1. Make your changes in your feature branch.
2. Add or update tests as necessary (see the Testing section below).
3. Ensure your code follows the project's coding standards.
4. Commit your changes:

   ```bash
   git commit -m "Add a brief, descriptive commit message"
   ```

## Testing

We strongly emphasize the importance of testing. Please include tests for any new features or bug fixes. This helps maintain the project's quality and prevents regressions.

To run the tests:

```bash
go test ./...
```

See also the [Development](development.md) page for more information on running the tests.

Ensure all tests pass before submitting your pull request.

## Submitting Changes

1. Push your changes to your fork on GitHub:

   ```bash
   git push origin your-branch-name
   ```

2. Open a pull request against the main Hide repository.
3. Clearly describe your changes and the reasoning behind them in the pull request description.
4. Link any relevant issues in the pull request description.

## Code Review Process

The project maintainers will review your pull request. They may suggest some changes or improvements. This is a normal part of the contribution process, so don't be discouraged!

## Good First Issues

If you're new to the project, look for issues labeled `good first issue`. These are typically easier tasks that are suitable for newcomers to the project.

You can find these issues [here](https://github.com/hide-org/hide/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22).

## Style Guide

For formatting, we use the standard Go formatting guidelines. To run the formatter, use the following command:

```bash
go fmt ./...
```

or use the `make` command:

```bash
make format
```

## Community Guidelines

- Be respectful and considerate in your communications with other contributors.
- Provide constructive feedback and be open to receiving it as well.
- Focus on the best possible outcome for the project.

## Questions?

If you have any questions or need further clarification, don't hesitate to open an issue for discussion.

Thank you for contributing to Hide!
