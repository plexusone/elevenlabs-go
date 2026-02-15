# Command Line Interface

The `elevenlabs` CLI provides command-line access to ElevenLabs text-to-speech services.

## Installation

```bash
go install github.com/agentplexus/go-elevenlabs/cmd/elevenlabs@latest
```

## Environment

Set your API key:

```bash
export ELEVENLABS_API_KEY=your-api-key
```

## Commands

| Command | Description |
|---------|-------------|
| [`tts`](tts.md) | Generate speech from a text file |
| [`ttsscript`](ttsscript.md) | Generate speech from a JSON script file |

## Quick Start

```bash
# Generate speech from text
elevenlabs tts -v <voice-id> speech.txt

# Use a preset
elevenlabs tts -v <voice-id> --preset oratory speech.txt

# Use a config file
elevenlabs tts --config tts-config.yaml speech.txt

# Generate from JSON script
elevenlabs ttsscript -lang en script.json
```

## Shell Completion

Generate shell completion scripts:

```bash
# Bash
elevenlabs completion bash > /etc/bash_completion.d/elevenlabs

# Zsh
elevenlabs completion zsh > "${fpath[1]}/_elevenlabs"

# Fish
elevenlabs completion fish > ~/.config/fish/completions/elevenlabs.fish
```

## Configuration

See [Configuration](configuration.md) for details on YAML config files and presets.
