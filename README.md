# PatchNote

AI-powered Git assistant that generates commit messages, PR descriptions, and code reviews from your repository changes.

## Features

- **Commit Message Generation** - Conventional Commit format with title and body
- **PR Description Generation** - Markdown PR descriptions with summary, motivation, testing, and checklist
- **Code Review** - Structured reviews with bug detection, security analysis, and suggestions
- **Multi-language** - Generate output in any language
- **Interactive Workflow** - Commit, edit, or copy results from the terminal
- **Gitignore-aware** - Automatically filters ignored files from diffs

## Installation

```bash
go install github.com/Sephy314/patchnote@latest
```

Or build from source:

```bash
git clone https://github.com/Sephy314/patchnote.git
cd patchnote
go build -o patchnote .
```

## Quick Start

```bash
# 1. Register your Groq API key
patchnote register

# 2. Stage your changes
git add .

# 3. Generate a commit message
patchnote commit
```

Get a free API key at [console.groq.com](https://console.groq.com).

## Usage

### Generate a commit message

```bash
patchnote commit
# or simply
patchnote
```

Analyzes staged and unstaged changes, generates a Conventional Commit message, then lets you commit, edit, copy, or cancel.

### Generate a PR description

```bash
patchnote pr
```

Generates a Markdown PR description with title, summary, motivation, main changes, testing notes, breaking changes, and a checklist.

### Generate a code review

```bash
patchnote review
```

Generates a structured code review covering summary, positive observations, potential bugs, security, performance, and suggestions.

## Commands

| Command | Description |
|---------|-------------|
| `patchnote` | Generate a commit message (default) |
| `patchnote commit` | Generate and create a commit |
| `patchnote pr` | Generate a PR description |
| `patchnote review` | Generate a code review |
| `patchnote register` | Register and verify a Groq API key |

## Configuration

Configuration is stored at `~/.config/patchnote/config.yaml`.

| Key | Default | Description |
|-----|---------|-------------|
| `provider` | `groq` | AI provider |
| `model` | `llama-3.3-70b-versatile` | Model to use |
| `temperature` | `0.2` | Generation temperature (0.0 - 1.0) |
| `language` | `english` | Output language |
| `api_key` | _(empty)_ | Groq API key |

## Project Structure

```
patchnote/
├── main.go                      # Entry point
├── cmd/
│   ├── root.go                  # Default command (commit flow)
│   ├── commit.go                # commit subcommand
│   ├── pr.go                    # pr subcommand
│   ├── review.go                # review subcommand
│   ├── register.go              # API key registration
│   └── clipboard.go             # Cross-platform clipboard
├── internal/
│   ├── ai/
│   │   ├── client.go            # AI client interface
│   │   └── groq.go              # Groq API implementation
│   ├── config/
│   │   └── config.go            # Config load/save (Viper)
│   ├── git/
│   │   └── client.go            # Git operations
│   ├── output/
│   │   └── output.go            # Output formatting
│   ├── prompts/
│   │   └── prompts.go           # Prompt construction
│   └── ui/
│       └── ui.go                # Terminal UI helpers
└── configs/                     # (empty)
```

## Testing

```bash
go test ./...
```

## License

Apache License 2.0 - see [LICENSE](LICENSE) for details.
