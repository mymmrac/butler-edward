# Butler Edward

AI agent that aims to be simple and secure yet capable of being useful for day-to-day tasks.

> [!WARNING]
> This is work in progress software in its pre-alpha state. Use at your own risk.

## Features

- **Multiple Channels:** Interact with Edward via Terminal or Telegram.
- **Provider Agnostic:** Supports any OpenAI-compatible API (e.g., OpenAI, Groq, OpenRouter).
- **Tool Support:** Edward can interact with your local filesystem to read and write files.
- **Secure:** Designed with simplicity and security in mind.

## Configuration

To configure Butler Edward, create a `config.yaml` file in the project root. You can use `config.example.yaml` as a
template.

### Example Configuration

```yaml
defaults:
  provider: "groq"
  model: "llama-3.1-8b-instant"

workspace:
  root: "./workspace"

channels:
  terminal:
    enabled: true
  telegram:
    enabled: true
    bot-token: "YOUR_TELEGRAM_BOT_TOKEN"

providers:
  openai-compatible:
    groq:
      enabled: true
      base-url: "https://api.groq.com/openai/v1"
      chat-api: "/chat/completions"
      models-api: "/models"
      api-key: "YOUR_GROQ_API_KEY"
      models:
        - name: "llama-3.1-8b-instant"
```

### Environment Variables

You can also use environment variables to override configuration values. Use the prefix `BUTLER_EDWARD_` and replace
dots/hyphens with underscores.

Example: `BUTLER_EDWARD_CHANNELS_TELEGRAM_BOT_TOKEN=your_token`

## Usage

### Running the Agent

You can run Butler Edward using the Go CLI:

```bash
go run .
```

Or build and run the binary:

```bash
go build -o ./bin/butler-edward .
./bin/butler-edward
```

### Available Channels

- **Terminal:** Direct interaction through your terminal's standard input/output.
- **Telegram:** Interact with Edward via a Telegram Bot. Requires a bot token from [@BotFather](https://t.me/BotFather).

## Adding Providers

Butler Edward supports OpenAI-compatible providers. To add a new provider, update the `providers.openai-compatible`
section in your `config.yaml`:

```yaml
providers:
  openai-compatible:
    my-provider:
      enabled: true
      base-url: "https://api.my-provider.com/v1"
      chat-api: "/chat/completions"
      models-api: "/models"
      api-key: "YOUR_API_KEY"
      models:
        - name: "model-name-1"
        - name: "model-name-2"
```

## Tools

Edward currently has access to the following filesystem tools:

- `read_dir`: List files in the workspace.
- `read_file`: Read content of a file.
- `write_file`: Write content to a file.

All tool operations are restricted to the directory defined in `workspace.root`.
