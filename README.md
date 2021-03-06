# ShellAI
A fantastic little AI Shell Assistant for my friends, powered by GPT3. *(My friends aka pre-alpha testers.)*

## Setup

### tl;dr – if you know what you're doing
Compile `shell_ai.ts` using Deno, set `SHELL_AI_KEY` env var to your user key.

### Detailed Instructions

1. Clone repo.
```
git clone https://github.com/ibigio/shell-ai.git
cd shell-ai
```

2. Install Deno. (https://deno.land/manual/getting_started/installation) At time of writing, this is the command you'll find in that guide for macOS and Linux:
```
curl -fsSL https://deno.land/x/install/install.sh | sh
```

3. Compile `shell_ai.ts`. (This will create an executable called `shell_ai`.)
```
deno compile --allow-net --allow-env --allow-read shell_ai.ts
```


4. Save the executable in a bin. If you don't have one, follow steps 5 and 6.

5. Create a new bin.
```
mkdir -p ~/CustomBin/bin
mv shell_ai ~/CustomBin/bin
```

6. Add the bin to your path. You should also add the following line to your `.zshrc` or `.bashrc`:
```
export PATH=$PATH:~/CustomBin/bin
```

7. Rename the executable to a short name of your chosing. I like 'q'.
```
mv ~/CustomBin/bin/shell_ai ~/CustomBin/bin/q
```

8. Set the `SHELL_AI_KEY` environment variable to your user key. Add this line to your `.zashrc` or `.bashrc` as well!
```
export SHELL_AI_KEY="[insert key here]"
```

9. Open a new shell and have fun!

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