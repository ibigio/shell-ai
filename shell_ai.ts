const user_key = Deno.env.get("SHELL_AI_KEY");
if (user_key == null) {
  console.log("======================================================");
  console.log("                        Hello!                        ");
  console.log("======================================================");
  console.log("");
  console.log(
    "You're seeing this greeting because the SHELL_AI_KEY\nenvironment variable is not set. Make sure to set it\nto your user key with the following command:"
  );
  console.log("");
  console.log('  export SHELL_AI_KEY="[insert key here]"');
  console.log("");
  console.log(
    "To avoid having to do this in future terminal sessions,\nyou can also add the above line to your .zshrc or\n.bashrc!"
  );
  console.log("");
  console.log(
    "I recommend putting this executable in a bin, and\nrenaming it to 'q' (or some other easy command). You\ncan do that like this:"
  );
  console.log("");
  console.log("  mkdir -p ~/CustomBin/bin");
  console.log("  mv " + Deno.execPath() + " ~/CustomBin/bin/g");
  console.log("  export PATH=$PATH:~/CustomBin/bin/g");
  console.log("");
  console.log(
    "You should add that last line to your .zshrc or\n.bashrc as well!"
  );
  console.log("");
  console.log("(If you don't have a user key, ask Ilan!)");
  Deno.exit(1);
}

if (Deno.args.length == 0) {
  console.log("usage: q desired command text");
  Deno.exit(1);
}

async function run_client() {
  const phrase = Deno.args.join(" ");
  const response = await fetch("https://shell-ai.deno.dev", {
    method: "POST",
    body: JSON.stringify({
      userID: user_key,
      phrase: phrase,
    }),
  });

  if (response.status != 200) {
    throw response.statusText;
  }

  try {
    const j = await response.json();
    const completion = j?.completion;
    console.log(completion);
  } catch {
    console.log("error: unable to parse response");
    Deno.exit(1);
  }
}

try {
  await run_client();
} catch (e) {
  console.log(e);
  console.log("error: command failed");
  Deno.exit(1);
}
