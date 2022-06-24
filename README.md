# ShellAI
A fantastic little AI Shell Assistant for my friends, powered by GPT3.

## Setup

1. Install Deno. (https://deno.land/manual/getting_started/installation)
2. Compile the executable. (This will create an executable called `shell_ai`.)
```
deno compile --allow-net --allow-env --allow-read shell_ai.ts
```


3. Save the executable in a bin. If you don't have one, follow steps 4 and 5.

4. Create a new bin.
```
mkdir -p ~/CustomBin/bin
mv shell_ai ~/CustomBin/bin
```

5. Add the bin to your path. You should also add the following line to your `.zshrc` or `.bashrc`:
```
export PATH=$PATH:~/CustomBin/bin
```

6. Rename the executable to a short name of your chosing. I like 'q'.
```
mv ~/CustomBin/bin/shell_ai ~/CustomBin/bin/q
```

7. Set the `SHELL_AI_KEY` environment variable to your user key. Add this line to your `.zashrc` or `.bashrc` as well!
```
export SHELL_AI_KEY="[insert key here]"
```

8. Open a new shell and have fun!

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