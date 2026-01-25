# OpenAPI Specification

This directory contains the OpenAPI 3.0 specification for the Wish List Application API, split into multiple files for better maintainability.

## Structure

The specification is divided into the following sections:

- `openapi.json` - Main specification file that references all other files
- `paths/` - API endpoint definitions
- `components/schemas/` - Data model definitions
- `components/responses/` - Common response definitions
- `components/securitySchemes/` - Security scheme definitions

## Commands

Use the following Make commands to work with the specification:

```bash
# Validate the specification
make openapi-validate

# Bundle all split files into a single file (output to dist/openapi.json)
make openapi-bundle

# Preview the documentation in a browser
make openapi-preview

# Show information about the specification structure
make openapi-info
```

## Tools

The specification uses [Redocly CLI](https://redocly.com/docs/cli/) for validation and bundling:

- Validation: `npx @redocly/cli lint api/openapi.json`
- Bundling: `npx @redocly/cli bundle api/openapi.json --output dist/openapi.json`
- Preview: `npx @redocly/cli preview-docs api/openapi.json`

## Benefits

Splitting the specification into multiple files provides:

- Better organization and readability
- Easier collaboration between team members
- Modularity - individual components can be updated independently
- Improved version control with smaller, focused commits

## Configuration

The `.redocly.yaml` file in the project root contains the configuration for Redocly tools.