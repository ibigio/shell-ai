# ShellAI
A fantastic little AI Shell Assistant powered by GPT.

_Turn natural language into shell commands, and ask open-ended questions!_

## Install

```bash
brew tap ibigio/ibigio
brew install shell-ai
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

You can get really specific too.
```
$ q print my local ip formatted like so "ip: [ip]" for mac
ifconfig | grep "inet " | grep -v 127.0.0.1 | awk '{print "ip: " $2}'
```

Or general.
```
$ q how do i set up a new nextjs project
To set up a new Next.js project, first make sure you have Node.js installed. Then, run the following command in your terminal:

npx create-next-app your-project-name
```

And deep.
```
$ q what is the meaning of life
42
```
