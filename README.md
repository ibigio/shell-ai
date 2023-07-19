<img width="1280" alt="Frame 7" src="https://github.com/ibigio/shell-ai/assets/25421602/8bbb4ed9-99e3-42df-9a79-dc101dc693ad">

# ShellAI

A delightfully minimal, yet remarkably powerful AI Shell Assistant.

![shell-ai-demo](https://github.com/ibigio/shell-ai/assets/25421602/f480db5d-3787-49d8-b1bc-a027f65858e6)

> "Ten minutes of Googling is now ten seconds in the terminal."
>
> ~ Joe C.

# About

For developers, referencing things online is inevitable – but one can only look up "how to do [X] in git" so many times before losing your mind.

**ShellAI** is meant to be a faster and smarter alternative to online reference: for shell commands, code examples, error outputs, and high-level explanations. We believe tools should be beautiful, minimal, and convenient, to let you get back to what you were doing as quickly and pleasantly as possible. That is the purpose of ShellAI.

## Install

### Homebrew

```bash
brew tap ibigio/tap
brew install shell-ai
```

### Linux

```bash
curl https://raw.githubusercontent.com/ibigio/shell-ai/main/install.sh | bash
```

## Usage

Type `q` followed by a description of a shell command, code snippet, or general question!

### Features

- Generate shell commands from a description.
- Reference code snippets for any programming language.
- Fast, syntax-highlighted, minimal UI.
- Auto-extract code from response and copy to clipboard.
- Follow up to refine command or explanation.
- Concise, helpful responses.
- Tailored support for GPT 3.5 and GPT 4.

### Configuration

Set your [OpenAI API key](https://platform.openai.com/account/api-keys).

```bash
export OPENAI_API_KEY=[your key]
```

Configure the default model (or even add your own) at `~/.shell-ai/config.yaml`.

(We currently support `gpt-3.5-turbo` and `gpt-4`, but OSS models coming soon!)

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

# Contributing

ShellAI is polished but by no means finished, feel free to open issues or pull requests on the repo! If you have other ideas or just want to say hi, go ahead and reach out!

Also I'll be adding support for OSS models in the coming weeks, adding your own is now easier! Give it a shot in `~/.shell-ai/config.yaml`. DM for questions [@ilanbigio](https://twitter.com/ilanbigio) :)
