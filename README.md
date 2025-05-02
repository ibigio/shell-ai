<img width="1280" alt="Frame 7" src="https://github.com/ibigio/shell-ai/assets/25421602/8bbb4ed9-99e3-42df-9a79-dc101dc693ad">

# ShellAI

A delightfully minimal, yet remarkably powerful AI Shell Assistant.

![shell-ai-demo](https://github.com/ibigio/shell-ai/assets/25421602/f480db5d-3787-49d8-b1bc-a027f65858e6)

> "Ten minutes of Googling is now ten seconds in the terminal."
>
> ~ Joe C.

<p align="center">
<a href="https://twitter.com/ilanbigio"><img src="https://img.shields.io/twitter/follow/ilanbigio?style=for-the-badge&logo=twitter" alt="Twitter Follow"></a>
</p>

# About

For developers, referencing things online is inevitable – but one can only look up "how to do [X] in git" so many times before losing your mind.

<p align="center">
  <img src="https://imgs.xkcd.com/comics/tar.png">
</p>

**ShellAI** is meant to be a faster and smarter alternative to online reference: for shell commands, code examples, error outputs, and high-level explanations. We believe tools should be beautiful, minimal, and convenient, to let you get back to what you were doing as quickly and pleasantly as possible. That is the purpose of ShellAI.

_New: ShellAI now supports local models! See [Custom Model Configuration](#custom-model-configuration-new) for more._

# Install

### Homebrew

```bash
brew tap ibigio/tap
brew install shell-ai
```

### Linux

```bash
curl https://raw.githubusercontent.com/ibigio/shell-ai/main/install.sh | bash
```

# Usage

Type `q` followed by a description of a shell command, code snippet, or general question!

### Features

- Generate shell commands from a description.
- Reference code snippets for any programming language.
- Fast, syntax-highlighted, minimal UI.
- Auto-extract code from response and copy to clipboard.
- Follow up to refine command or explanation.
- Concise, helpful responses.
- Built-in support for GPT 3.5 and GPT 4.
- Support for [other providers and open source models](#custom-model-configuration-new)!

### Configuration

Set your [OpenAI API key](https://platform.openai.com/account/api-keys).

```bash
export OPENAI_API_KEY=[your key]
```

For more options (like setting the default model), run:

```bash
q config
```

(For more advanced config, like configuring open source models, check out the [Custom Model Configuration](#custom-model-configuration-new) section)

# Examples

### Shell Commands

`$ q make a new git branch`

```
git branch new-branch
```

`$ q find files that contain "administrative" in the name`

```
find /path/to/directory -type f -name "*administrative*"
```

### Code Snippets

`$ q initialize a static map in golang`

```golang
var staticMap = map[string]int{
    "key1": 1,
    "key2": 2,
    "key3": 3,
}
```

`$ q create a generator function in python for dates`

```python
def date_generator(start_date, end_date):
    current_date = start_date
    while current_date <= end_date:
        yield current_date
        current_date += datetime.timedelta(days=1)
```

# Custom Model Configuration (New!)

You can now configure model prompts and even add your own model setups in the `~/.shell-ai/config.yaml` file! ShellAI _should_ support any model that can be accessed through a chat-like endpoint... including local OSS models.

(I'm working on making config entirely possible through `q config`, but until then you'll have to edit the file directly.)

### Config File Syntax

````yaml
preferences:
  default_model: gpt-4-1106-preview

models:
  - name: gpt-4-1106-preview
    endpoint: https://api.openai.com/v1/chat/completions
    auth_env_var: OPENAI_API_KEY
    org_env_var: OPENAI_ORG_ID
    prompt:
      [
        {
          role: "system",
          content: "You are a terminal assistant. Turn the natural language instructions into a terminal command. By default always only output code, and in a code block. However, if the user is clearly asking a question then answer it very briefly and well.",
        },
        { role: "user", content: "print hi" },
        { role: "assistant", content: "```bash\necho \"hi\"\n```" },
      ]
    # other models ...

config_format_version: "1"
````

**Note:** The `auth_env_var` is set to `OPENAI_API_KEY` verbatim, not the key itself, so as to not keep sensitive information in the config file.

### Setting Up a Local Model

As a proof of concept I set up `stablelm-zephyr-3b.Q8_0` on my MacBook Pro (16GB) and it works decently well. (Mostly some formatting oopsies here and there.)

Here's what I did:

1. I cloned and set up `llama.cpp` ([repo](https://github.com/ggerganov/llama.cpp)). (Just follow the instructions.)
2. Then I downloaded the [`stablelm-zephyr-3b.Q8_0`](https://huggingface.co/TheBloke/stablelm-zephyr-3b-GGUF/blob/main/stablelm-zephyr-3b.Q8_0.gguf) GGUF [from hugging face](https://huggingface.co/TheBloke/stablelm-zephyr-3b-GGUF) (thanks, TheBloke) and saved it under `llama.cpp/models/`.
3. Then I ran the model in server mode with chat syntax (from `llama.cpp/`):

```bash
./server -m models/stablelm-zephyr-3b.Q8_0.gguf --host 0.0.0.0 --port 8080
```

4. Finally I added the new `model` config to my `~/.shell-ai/config.yaml`, and wrestled with the prompt until it worked – bet you can do better. (As you can see, YAML is [flexible](https://docs.ansible.com/ansible/latest/reference_appendices/YAMLSyntax.html).)

````yaml
models:
  - name: stablelm-zephyr-3b.Q8_0
    endpoint: http://127.0.0.1:8080/v1/chat/completions
    auth_env_var: OPENAI_API_KEY
    org_env_var: OPENAI_ORG_ID
    prompt:
      - role: system
        content:
          You are a terminal assistant. Turn the natural language instructions
          into a terminal command. By default always only output code, and in a code block.
          DO NOT OUTPUT ADDITIONAL REMARKS ABOUT THE CODE YOU OUTPUT. Do not repeat the
          question the users asks. Do not add explanations for your code. Do not output
          any non-code words at all. Just output the code. Short is better. However, if
          the user is clearly asking a general question then answer it very briefly and
          well. Indent code correctly.
      - role: user
        content: get the current time from some website
      - role: assistant
        content: |-
          ```bash
          curl -s http://worldtimeapi.org/api/ip | jq '.datetime'
          ```
      - role: user
        content: print hi
      - role: assistant
        content: |-
          ```bash
          echo "hi"
          ```

# other models ...
````

and also updated the default model (which you can also do from `q config`):

```yaml
preferences:
  default_model: stablelm-zephyr-3b.Q8_0
```

And huzzah! You can now use ShellAI on a plane.

(Fun fact, I implemented a good bit of the initial config TUI on a plane using this exact local model.)

### Setting Up Azure OpenAI endpoint

Define `AZURE_OPENAI_API_KEY` environment variable and make few changes to the config file.

```yaml
models:
  - name: azure-gpt-4
    endpoint: https://<resource_name>.openai.azure.com/openai/deployments/<deployment_name>/chat/completions?api-version=<api_version>
    auth_env_var: AZURE_OPENAI_API_KEY
```

### I Fucked Up The Config File

Great! Means you're having fun.

`q config revert` will revert it back to the latest working version you had. (ShellAI automatically saves backups on successful updates.) Wish more tools had this.

`q config reset` will nuke it to the (latest) default config.

# Contributing

Now that `~/.shell-ai/config.yaml` is set up, there's so much to do! I'm open to any feature ideas you might want to add, but am generally focused on two efforts:

1. Building out the config TUI so you never have to edit the file directly (along with other nice features like re-using prompts, etc), and
2. Setting up model install templates – think an (immutable) templates file where people can configure the model config and install steps, so someone can just go like `go config` -> `Install Model` -> pick one, and start using it.

Like I said, if you have other ideas, or just want to say hi, go ahead and reach out! [@ilanbigio](https://twitter.com/ilanbigio) :)
