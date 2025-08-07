package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// Command represents a CLI command
type Command struct {
	Name        string
	Description string
	Flags       *flag.FlagSet
	Run         func(args []string)
}

// CLI holds registered commands
type CLI struct {
	commands map[string]*Command
}

// NewCLI creates a new CLI instance
func NewCLI() *CLI {
	return &CLI{commands: make(map[string]*Command)}
}

// RegisterCommand adds a command to the CLI
func (cli *CLI) RegisterCommand(cmd *Command) {
	cli.commands[cmd.Name] = cmd
}

// Run parses arguments and executes the right command
func (cli *CLI) Run() {
	if len(os.Args) < 2 {
		fmt.Println("No command provided")
		cli.PrintHelp()
		return
	}

	cmdName := os.Args[1]
	cmd, exists := cli.commands[cmdName]
	if !exists {
		fmt.Printf("Unknown command: %s\n", cmdName)
		cli.PrintHelp()
		return
	}

	cmd.Flags.Parse(os.Args[2:])
	cmd.Run(cmd.Flags.Args())
}

// PrintHelp prints all available commands
func (cli *CLI) PrintHelp() {
	fmt.Println("Available commands:")
	for name, cmd := range cli.commands {
		fmt.Printf("  %-10s %s\n", name, cmd.Description)
	}
}

func main() {
	cli := NewCLI()

	// Help
	help := &Command{
		Name:        "help",
		Description: "Show this text",
		Flags:       flag.NewFlagSet("help", flag.ExitOnError),
	}
	function := help.Flags.String("", "", "Gets more specific details on a given command")
	help.Run = func(args []string) {
		if function == nil || *function == "" {
			cli.PrintHelp()
		} else {
			// print some data for a given command
		}
	}
	cli.RegisterCommand(help)

	// Authentication command
	auth := &Command{
		Name:        "auth",
		Description: "Authenticate the user with the specified credentials",
		Flags:       flag.NewFlagSet("key", flag.ExitOnError),
	}
	details := auth.Flags.String("key", "", "Nylas API key")
	auth.Run = func(args []string) {
		// attempt authentication here later
		fmt.Printf("Hello, %s!\n", details)
	}
	cli.RegisterCommand(auth)

	// Webhook command
	webhook := &Command{
		Name:        "webhook",
		Description: "Manages various functions of a webhook",
		Flags:       flag.NewFlagSet("tunnel", flag.ExitOnError),
	}
	tunnelUrl := webhook.Flags.String("tunnel", "", "The locally hosted URL (http://localhost:PORT) to forward webhook messages to")
	webhook.Run = func(args []string) {
		client := &http.Client{
			Timeout: 0, // Disable timeout for streaming connections
		}

		req, err := http.NewRequest("GET", "http://localhost:8080/stream", nil)
		if err != nil {
			log.Fatalf("Error creating request: %v", err)
		}
		req.Header.Set("Accept", "text/event-stream")

		resp, err := client.Do(req)
		if err != nil {
			log.Fatalf("Error making request: %v", err)
		}

		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			log.Fatalf("Server returned non-200 status: %d", resp.StatusCode)
		}

		reader := bufio.NewReader(resp.Body)

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					fmt.Println("Server closed the connection.")
					break
				}
				log.Fatalf("Error reading from stream: %v", err)
			}

			// Remove leading/trailing whitespace
			line = strings.TrimSpace(line)

			if strings.HasPrefix(line, "data: ") { // Webhook events
				data := strings.TrimPrefix(line, "data: ")
				if tunnelUrl != nil && *tunnelUrl != "" {
					// Forward the data to the locally hosted address
					http.Post(*tunnelUrl, "application/json", strings.NewReader(data))
				} else {
					fmt.Printf("Received message: %s\n", data)
				}
			} else if strings.HasPrefix(line, ":") { // Comments
				fmt.Println("Received comment: " + line[1:])
			}
		}
	}
	cli.RegisterCommand(webhook)

	// Run the CLI
	cli.Run()
}
