## ADDED Requirements

### Requirement: Provider use can persist to settings
The system SHALL support a `--persist` flag on `aix provider use` that writes the generated environment variables to `~/.claude/settings.json`.

#### Scenario: Persist activated provider
- **WHEN** user runs `aix provider use my-provider --persist`
- **THEN** the command writes `ANTHROPIC_BASE_URL`, the token variable, and any model overrides to `~/.claude/settings.json`
- **AND** the command still prints shell exports to stdout

#### Scenario: Persist without token fails
- **WHEN** user runs `aix provider use my-provider --persist` but no token is stored
- **THEN** the command returns a token error and does not modify `settings.json`

#### Scenario: Non-persist use remains unchanged
- **WHEN** user runs `aix provider use my-provider` without `--persist`
- **THEN** the command prints shell exports to stdout
- **AND** `~/.claude/settings.json` is not modified

#### Scenario: Config current shows persisted provider
- **WHEN** user has previously run `aix provider use my-provider --persist`
- **THEN** `aix config current` displays the persisted environment variables
