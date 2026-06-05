## ADDED Requirements

### Requirement: CMD shell escaping handles metacharacters
The `shellQuote` function SHALL escape Windows CMD special characters (`&`, `|`, `<`, `>`, `^`, `"`) using the `^` escape character when the shell type is `ShellCmd`. Values containing spaces SHALL be wrapped in double quotes.

#### Scenario: URL with query parameters
- **WHEN** `shellQuote("cmd", "https://api.example.com/v1?key=val&other=2")` is called
- **THEN** the `&` is escaped as `^&`, producing a safe `set` command

#### Scenario: Value with pipe character
- **WHEN** `shellQuote("cmd", "token|value")` is called
- **THEN** the `|` is escaped as `^|`

#### Scenario: Value with spaces
- **WHEN** `shellQuote("cmd", "some value with spaces")` is called
- **THEN** the value is wrapped in double quotes with internal `"` escaped as `^"`

### Requirement: PowerShell shell escaping uses single-quote wrapping
The `shellQuote` function SHALL use single-quote wrapping for PowerShell values, with single quotes inside the value escaped as `''` (double-single-quote).

#### Scenario: Value with double quotes
- **WHEN** `shellQuote("powershell", `key"value`)` is called
- **THEN** the value is wrapped in single quotes: `'key"value'` (double quotes are literal in PowerShell single-quote strings)

#### Scenario: Value with single quotes
- **WHEN** `shellQuote("powershell", "it's a value")` is called
- **THEN** the single quote is escaped: `'it''s a value'`

#### Scenario: Value with dollar sign
- **WHEN** `shellQuote("powershell", "$env:FOO")` is called
- **THEN** the value is wrapped in single quotes: `'$env:FOO'` (dollar signs are literal in single-quote strings)

### Requirement: POSIX shell escaping remains correct
The existing single-quote escaping for bash, zsh, and fish SHALL remain unchanged. Values are wrapped in single quotes with embedded single quotes escaped as `'\''`.

#### Scenario: Value with single quote
- **WHEN** `shellQuote("bash", "it's here")` is called
- **THEN** the result is `'it'\''s here'`

#### Scenario: Value with special characters
- **WHEN** `shellQuote("zsh", "$HOME/path && echo foo")` is called
- **THEN** the result is `'$HOME/path && echo foo'` (all characters literal inside single quotes)

### Requirement: FormatForShell outputs correct syntax for all shells
The `FormatForShell` method SHALL produce syntactically correct output for bash, zsh, fish, PowerShell, and CMD with properly escaped values.

#### Scenario: Round-trip export and eval for bash
- **WHEN** environment variables with special characters are formatted for bash and evaluated
- **THEN** the resulting environment variables contain the exact original values

#### Scenario: Round-trip export for PowerShell
- **WHEN** environment variables with special characters are formatted for PowerShell and executed
- **THEN** the resulting environment variables contain the exact original values
