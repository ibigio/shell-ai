# ShellAI
A delightfully minimal, yet remarkably powerful AI Shell Assistant.

![2023-06-08 00 31 26](https://github.com/ibigio/shell-ai/assets/25421602/88a4aed3-4235-461a-ae1b-0a47c32d193b)

>
> "Ten minutes of Googling is now ten seconds in the terminal."
>
>   ~ Joe C.
>

## Install

```bash
brew tap ibigio/tap
brew install shell-ai
```

or

```bash
curl https://raw.githubusercontent.com/ibigio/shell-ai/main/install.sh | bash
```

and set your OPENAI key ([get one here](https://platform.openai.com/account/api-keys)) like so:

```bash
export OPENAI_API_KEY=[your key]
```

(Don't forget to set up billing with OpenAI.)

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


<details>
<summary>I want even more power.</summary>
<br>
By default this tool uses the `gpt-3.5-turbo` model, but if you want to use `gpt-4` you can override like so:

```bash
export OPENAI_MODEL_OVERRIDE="gpt-4"
```
</details>


## More Sample Use Cases
```
$ q print my local ip formatted like so "ip: [ip]" for mac
ifconfig | grep "inet " | grep -v 127.0.0.1 | awk '{print "ip: " $2}'
```
```
$ q how do i set up a new nextjs project
To set up a new Next.js project, first make sure you have Node.js installed. Then, run the following command in your terminal:

npx create-next-app your-project-name
```
```
$ q what is the meaning of life
42
```
