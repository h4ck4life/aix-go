## ADDED Requirements

### Requirement: Provider edit command exists
The system SHALL provide an `aix provider edit <name>` command that modifies an existing provider's properties.

#### Scenario: Edit provider URL
- **WHEN** user runs `aix provider edit my-provider --url https://api.new.com`
- **THEN** the provider's base URL is updated in the registry

#### Scenario: Edit provider token type
- **WHEN** user runs `aix provider edit my-provider --token-type auth-token`
- **THEN** the provider's token variable is updated to `ANTHROPIC_AUTH_TOKEN`

#### Scenario: Edit provider model
- **WHEN** user runs `aix provider edit my-provider --model claude-sonnet-4-6`
- **THEN** the provider's custom model is updated

#### Scenario: Edit default model alias
- **WHEN** user runs `aix provider edit my-provider --default-model sonnet=claude-sonnet-4-6`
- **THEN** the provider's default model alias is updated

#### Scenario: Edit multiple properties at once
- **WHEN** user runs `aix provider edit my-provider --url https://api.new.com --token-type api-key`
- **THEN** both the URL and token type are updated

#### Scenario: Edit non-existent provider fails
- **WHEN** user runs `aix provider edit nonexistent --url https://api.new.com`
- **THEN** the command returns a validation error with exit code 2
