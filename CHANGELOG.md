# Changelog

All notable changes to this project will be documented in this file.

## unreleased

### Added

- Added support for multiple AI providers:
  - Google Gemini (using the `-p gemini` flag)
  - Anthropic Claude (using the `-p claude` flag)
  - OpenAI (default, using the `-p openai` flag)
- Added provider selection in the CLI with the `-p` flag
- Added model specification with the `-model` flag 
- Added provider and model info display with the `-m` flag
- Added fully interactive numbered menu for provider and model selection
- Added user-friendly model descriptions for each provider
- Added support for latest Gemini models (2.5, 2.0, 1.5)
- Added error handling and default selections for improved user experience
- Fixed streaming issues to prevent chat mode from hanging
- Improved error handling in all providers to ensure proper stream completion
- Added support for provider-specific configuration

## 0.6.0

### Changed

- Changed project name from Yo to Yai, thanks to [@K-arch27](https://github.com/K-arch27) for the suggestion.

## 0.5.0

### Added

- Display help when starting REPL mode

## 0.4.0

### Added

- Configuration for OpenAI API model (default gpt-3.5-turbo) 

## 0.3.0

### Added

- Configuration for OpenAI API max-tokens (default 1000)
- Better feedback for install script

### Updated

- sashabaranov/go-openai to v1.8.0

## 0.2.0

### Added

- Support for pipe input

## 0.1.0

### Added

- Exec prompt mode
- Chat prompt mode
