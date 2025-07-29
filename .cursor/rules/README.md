# Cursor Rules

This directory contains coding rules and guidelines that are automatically applied by the AI assistant when helping with development tasks.

## Structure

- `coding/` - Contains coding-related rules
  - `generals.mdc` - General coding principles and language preferences
  - `go.mdc` - Go-specific coding patterns and best practices
  - `documentation.mdc` - Documentation standards and examples
  - `architecture.mdc` - Project architecture and design patterns
  - `change-mgmt.mdc` - Version control and change management guidelines

## How It Works

1. The AI assistant automatically checks these rules for every prompt
2. Rules are applied based on the context (file type, operation type)
3. No additional configuration is needed - just keep this directory in your project

## Language Preferences

- Code and variables: English
- Comments and documentation: Spanish
- Error messages: Spanish

## Key Guidelines

1. Follow Clean Architecture principles
2. Use dependency injection
3. Write comprehensive tests
4. Document all public APIs
5. Handle errors with Spanish messages
6. Keep code simple and maintainable

## Examples

The rules files contain Bitcoin Price Alert specific examples to demonstrate proper implementation of:
- Price monitoring
- Alert management
- WebSocket updates
- Notification strategies
- Error handling 