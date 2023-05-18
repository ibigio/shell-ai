# ShellAI
A fantastic little AI Shell Assistant powered by GPT4.

_Turn natural language into shell commands!_

## Setup

1. Install Deno. (https://deno.land/manual/getting_started/installation)
```
brew install deno
```

2. Install Shell AI.
```
deno install --allow-read --allow-env --allow-net --name q https://raw.githubusercontent.com/ibigio/shell-ai/main/shell_ai.ts
```

8. Set the `SHELL_AI_KEY` environment variable to your user key. (Add this line to your `.zashrc` or `.bashrc` as well!)
```
export SHELL_AI_KEY="[insert key here]"
```

## Usage

Type `q` followed by a description of a shell command you want to write!

Nice for beginners...
```
$ q make a new git branch
git checkout -b "new-branch"
```

...and those who forget how to use `find`, like me.
```
$ q find files that contain "administrative" in the name
find . -name "*administrative*"
```
