#!/usr/bin/env node
const fetch = (...args) => import('node-fetch').then(({default: fetch}) => fetch(...args));
const readline = require("readline");

// Command structure
class Command {
    constructor(name, shortcut, description, options, run) {
        this.name = name;
        this.shortcut = shortcut;
        this.description = description;
        this.options = options || {};
        this.run = run;
    }
}

class CLI {
    constructor() {
        this.commands = {};
    }

    registerCommand(cmd) {
        this.commands[cmd.name] = cmd;
        if (cmd.shortcut)
            this.commands[cmd.shortcut] = cmd;
    }

    run() {
        const args = process.argv.slice(2);
        if (args.length === 0) {
            console.log("No command provided");
            this.printHelp();
            return;
        }

        const cmdName = args[0];
        const cmd = this.commands[cmdName];
        if (!cmd) {
            console.log(`Unknown command: "${cmdName}".
See $ nylas --help for a list of available commands.`);
            return;
        }

        // Simple flag parsing
        const parsed = this.parseFlags(args.slice(1), cmd.options);
        if (parsed.err)
            console.log("unknown flag: " + parsed.err);
        else
            cmd.run(parsed.args, parsed.flags);
    }

    printHelp() {
        console.log(`The official CLI for Nylas.

Before using the CLI, you'll need to log into your Nylas account:
    $ nylas auth

\x1B[1mCommands:\x1B[22m`);
        for (const name in this.commands) {
            if (!name.startsWith("-")) {
                const cmd = this.commands[name];
                console.log(`   ${name.padEnd(14)} ${cmd.description}`);
            }
        }
        console.log(`
\x1B[1mFlags:\x1B[22m`);
        for (const name in this.commands) {
            if (name.startsWith("--")) {
                const cmd = this.commands[name];
                console.log(`   ${(cmd.shortcut ? cmd.shortcut : "").padEnd(3)} ${name.padEnd(10)} ${cmd.description}`);
            }
        }
    }

    parseFlags(argv, options) {
        const flags = {};
        const args = [];
        for (let i = 0; i < argv.length; i++) {
            if (argv[i].startsWith("--") || argv[i].startsWith("-") && argv[i].length == 2) {
                const flagName = argv[i].substring(argv[i].startsWith("--") ? 2 : 1);
                const value = argv[i + 1] && !argv[i + 1].startsWith("--") ? argv[++i] : true;
                flags[flagName] = value;
            } else {
                args.push(argv[i]);
            }
        }
        for (const arg of args)
            if (!options.contains(arg))
                return { err: arg };
        return { args, flags };
    }
}

// Main
const cli = new CLI();

// Help flag
cli.registerCommand(new Command(
    "--help", "-h",
    "Show this text",
    null,
    (args, flags) => 
        cli.printHelp()
));

// Version flag
cli.registerCommand(new Command(
    "--version", "-v",
    "See the CLI version",
    null,
    (args, flags) => 
        console.log("Nylas CLI version 1.0.0")
));

// Authentication command
cli.registerCommand(new Command(
    "auth", null,
    "Authenticate the user with the specified credentials",
    { key: "" },
    (args, flags) => {
        flags.key;
        // Skeleton command right now, doesn't do anything.
    }
));

// Webhook command
cli.registerCommand(new Command(
    "webhook", null,
    "Manages various functions of a webhook",
    { tunnel: "" },
    async (args, flags) => {
        try {
            const res = await fetch("http://localhost:8080/stream", {
                method: "GET",
                headers: { "Accept": "text/event-stream" },
                timeout: 0 // Keep alive for streaming
            });

            if (!res.ok) {
                console.error(`Server returned non-200 status: ${res.status}`);
                return;
            }

            const rl = readline.createInterface({
                input: res.body,
                crlfDelay: Infinity
            });

            rl.on("line", (line) => {
                line = line.trim();
                if (line.startsWith("data: ")) {
                    const data = line.slice(6);
                    if (flags.tunnel) {
                        fetch(flags.tunnel, {
                            method: "POST",
                            headers: { "Content-Type": "application/json" },
                            body: data
                        }).catch(err => {
                            // Ignore stream closing errors
                            console.log(err);
                            if (err.message !== "socket hang up")
                                console.error(`Forwarding error: ${err.message}`);
                        });
                    } else {
                        console.log(`Received message: ${data}`);
                    }
                } else if (line.startsWith(":")) {
                    console.log("Received comment: " + line.slice(1));
                }
            });

            rl.on("close", () => {
                console.log("Server closed the connection.");
            });

        } catch (err) {
            console.error("Error making request:", err);
        }
    }
));

// Run CLI
cli.run();
