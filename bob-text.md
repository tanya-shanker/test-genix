Configuring Bob Shell

Bob Shell offers extensive configuration options to customize your experience. This guide explains how to configure Bob Shell to match your workflow and preferences.
Getting started

You can configure Bob Shell using:

    Project-specific settings: Create .bob/settings.json in your project directory
    User-wide settings: Edit ~/.bob/settings.json in your home directory
    Command-line arguments: Run --sandbox when starting Bob Shell

# Example: Start Bob Shell with sandbox mode enabled
bob --sandbox

Configuration system
How configuration works

Bob Shell uses a layered configuration system where settings from different sources are combined according to a specific precedence order:
Priority	Source	Location	Scope
7 (Highest)	Command-line arguments	bob --option value	Current session
6	Environment variables	Shell or .env files	System/session
5	System settings file	/etc/bobshell/settings.json	All users, all projects
4	Project settings file	.bob/settings.json	Current project
3	User settings file	~/.bob/settings.json	Current user, all projects
2	System defaults file	/etc/bobshell/system-defaults.json	All users, all projects
1 (Lowest)	Hardcoded defaults	Built into Bob Shell	Always applied

When the same setting is defined in multiple places, the higher-priority source takes precedence.
Settings file locations

Bob Shell looks for settings files in these locations:

    Project settings: .bob/settings.json in your project directory
    User settings: ~/.bob/settings.json in your home directory
    System settings:

/etc/bobshell/settings.json

    System defaults:

/etc/bobshell/system-defaults.json

    Tip: You can reference environment variables in your settings files using $VAR_NAME or ${VAR_NAME} syntax. For example: "apiKey": "$MY_API_TOKEN".

Core settings categories

Bob Shell settings are organized into categories. Each category contains related settings that control specific aspects of Bob Shell's behavior.
General settings

Control basic Bob Shell behavior and preferences.

{
  "general": {
    "preferredEditor": "code",
    "vimMode": false,
    "disableAutoUpdate": false,
    "disableUpdateNag": false,
    "checkpointing": {
      "enabled": true
    }
  }
}

Setting	Type	Default	Description
preferredEditor	string	undefined	Editor to use when opening files
vimMode	boolean	false	Enable Vim keybindings
disableAutoUpdate	boolean	false	Prevent automatic updates
disableUpdateNag	boolean	false	Hide update notifications
checkpointing.enabled	boolean	false	Enable session checkpointing
UI settings

Customize Bob Shell's appearance and interface elements.

{
  "ui": {
    "theme": "GitHub",
    "hideBanner": true,
    "hideTips": false,
    "showLineNumbers": true
  }
}

Setting	Type	Default	Description
theme	string	undefined	UI color theme
customThemes	object	{}	Custom theme definitions
hideWindowTitle	boolean	false	Hide window title bar
hideTips	boolean	false	Hide helpful tips
hideBanner	boolean	false	Hide application banner
hideFooter	boolean	false	Hide footer
showMemoryUsage	boolean	false	Show memory usage stats
showLineNumbers	boolean	false	Show line numbers in chat
showCitations	boolean	false	Show citations for generated text
accessibility.disableLoadingPhrases	boolean	false	Disable loading phrases
Context settings

Control how Bob Shell manages project context and memory.

{
  "context": {
    "fileName": ["CONTEXT.md", "AGENTS.md"],
    "discoveryMaxDirs": 200,
    "includeDirectories": ["../shared-lib", "~/reference-code"],
    "fileFiltering": {
      "respectGitIgnore": true,
      "respectBobIgnore": true
    }
  }
}

Setting	Type	Default	Description
fileName	string/array	undefined	Context file name(s)
importFormat	string	undefined	Memory import format
discoveryMaxDirs	number	200	Max directories to search
includeDirectories	array	[]	Additional directories to include
loadFromIncludeDirectories	boolean	false	Load context from included dirs
fileFiltering.respectGitIgnore	boolean	true	Honor .gitignore files
fileFiltering.respectBobIgnore	boolean	true	Honor .bobignore files
fileFiltering.enableRecursiveFileSearch	boolean	true	Enable recursive file search
Tools settings

Configure how Bob Shell uses and manages tools.

{
  "tools": {
    "sandbox": "docker",
    "allowed": ["run_shell_command(git)", "run_shell_command(npm test)"],
    "exclude": ["write_file"]
  }
}

Setting	Type	Default	Description
sandbox	boolean/string	undefined	Sandbox execution environment
usePty	boolean	false	Use node-pty for shell commands
core	array	undefined	Restrict built-in tools (allowlist)
exclude	array	undefined	Tools to exclude from discovery
allowed	array	undefined	Tools that bypass confirmation
discoveryCommand	string	undefined	Command for tool discovery
callCommand	string	undefined	Command for calling tools
MCP settings

Configure Model Context Protocol server connections.

{
  "mcpServers": {
    "mainServer": {
      "command": "bin/mcp_server.py"
    },
    "remoteServer": {
      "url": "https://example.com/mcp",
      "headers": {
        "Authorization": "Bearer token123"
      }
    }
  }
}

Setting	Type	Default	Description
mcp.serverCommand	string	undefined	Command to start MCP server
mcp.allowed	array	undefined	Allowlist of MCP servers
mcp.excluded	array	undefined	Denylist of MCP servers
mcpServers.<SERVER_NAME>	object	-	Server-specific configuration

For each MCP server, you can configure:

    command: Command to execute (for local servers)
    args: Command-line arguments
    env: Environment variables
    cwd: Working directory
    url: Server-Sent Events (SSE) endpoint URL (for remote servers)
    httpUrl: HTTP endpoint URL (for remote servers)
    headers: HTTP headers for requests
    timeout: Request timeout in milliseconds
    trust: Trust server and bypass confirmations
    includeTools: Tool names to include
    excludeTools: Tool names to exclude

Command-line arguments

Pass these arguments when starting Bob Shell to override settings for that session:

# Start Bob Shell with specific settings
bob --sandbox --approval-mode auto_edit

Argument	Description	Example
--prompt, -p	Non-interactive prompt	bob -p "Explain this code"
--prompt-interactive, -i	Interactive initial prompt	bob -i "Help me debug"
--sandbox, -s	Enable sandbox mode	bob -s
--debug, -d	Enable debug mode	bob -d
--all-files, -a	Include all files as context	bob -a
--yolo	Auto-approve all tool calls	bob --yolo
--approval-mode	Set tool approval mode	bob --approval-mode=auto_edit
--allowed-tools	Tools to auto-approve	bob --allowed-tools="git status"
--checkpointing	Enable checkpointing	bob --checkpointing
--include-directories	Add directories to workspace	bob --include-directories=../lib
--chat-mode	Choose the mode for interaction	bob --chat-mode
--hide-intermediary-output	Output only the final task completion output	bob --hide-intermediary-output
--reset-bobshell-key-secret	Remove Bob Shell's API key from the secret store	bob --reset-bobshell-key-secret
Context files

Context files (like AGENTS.md) provide instructions to the AI model. These files are loaded hierarchically:

    Global context: ~/.bob/AGENTS.md (applies to all projects)
    Project context: AGENTS.md in project root and parent directories
    Local context: AGENTS.md in subdirectories (for component-specific instructions)

Example context file

# Project: My TypeScript Library

## General instructions

- Follow existing coding style
- Add JSDoc comments to all functions
- Prefer functional programming patterns
- Target TypeScript 5.0 and Node.js 20+

## Coding style

- Use 2 spaces for indentation
- Interface names should be prefixed with `I`
- Private class members should be prefixed with `_`
- Use strict equality (`===` and `!==`)

Managing context

    Use /memory refresh to reload all context files
    Use /memory show to view the current context

Sandboxing

Sandboxing provides security when running potentially unsafe operations:

# Enable sandboxing for a session
bob --sandbox

You can create custom sandbox environments:

    Create .bob/sandbox.Dockerfile in your project
    Base it on the bobshell-sandbox image
    Add your custom dependencies

FROM bobshell-sandbox

# Add custom dependencies
RUN apt-get update && apt-get install -y python3-dev

Build and use your custom sandbox:

BUILD_SANDBOX=1 bob -s

Note:

The create-pr command is not compatible with Sandbox sessions.
Usage statistics

Bob Shell collects anonymous usage statistics to improve the product. This includes:

    Tool usage patterns (names, success/failure, duration)
    API request metrics (model, duration, success)
    Session configuration information

No personal information, prompt content, or file content is collected.

To opt out, add this to your settings:

{
  "privacy": {
    "usageStatisticsEnabled": false
  }
}




Starting an interactive session

Interactive sessions provide a conversational interface to Bob directly in your terminal, allowing real-time assistance with your development tasks.
BobShell interactive session
Start an interactive session

Open a new terminal window. Ensure your BOBSHELL_API_KEY environment variable is set.

Navigate to the main directory of your project.

To start a Bob Shell interactive session, run:

bob

Basic usage
Interact with Bob Shell

    Type your instructions or questions directly in the terminal
    Press Enter to send your message to Bob
    Bob Shell will respond with its analysis and suggestions
    For tool usage (like reading or writing files), you'll be prompted to approve or decline each action

Reference files

Use the @ symbol to reference files in your project:

Explain the functionality in @src/main.js

This tells Bob Shell to read and analyze the specified file before responding.
Use slash commands

Type / to access a menu of available commands:

    Built-in commands like /help for assistance
    Mode-switching commands like /code or /ask
    Custom commands you've created

For a complete list of available commands, see Using slash commands in Bob Shell.
View file changes

When Bob Shell needs to modify files, it will show you the proposed changes:

    By default, changes are displayed in the terminal with a CLI diff view.
    To use an external editor for reviewing changes, configure your preferred editor with the /editor command. Then select the "Show diff in editor" option when prompted to review changes.

Advanced features
Tool approvals

For security, Bob Shell requires your approval before:

    Reading files.
    Writing or modifying files.
    Executing commands.

You can approve or decline each action individually when prompted.
Multi-turn conversations

Bob Shell maintains context throughout your conversation, allowing you to:

    Ask follow-up questions.
    Refine previous requests.
    Build on earlier responses.

Tips for effective use

    Be specific in your requests to get more targeted responses.
    Use @ references to provide code context when needed.
    For complex tasks, break them down into smaller steps.
    Use slash commands to quickly access common functionality.
    When working with large codebases, direct Bob Shell to the most relevant files.

When to use interactive session

Interactive session works best for:

    Exploratory coding sessions.
    Debugging and troubleshooting.
    Learning new concepts or technologies.
    Tasks that require multiple back-and-forth exchanges.
    Projects where you need to review changes before they're applied.



Starting a non-interactive session

Non-interactive session allows you to use Bob Shell directly from the command line without entering an interactive session. This approach works well for automation, scripting, and batch processing tasks.
Getting started

    Open a new terminal window. Ensure your BOBSHELL_API_KEY environment variable is set.
    Navigate to the main directory of your project.
    Run Bob Shell with your prompt using the -p command.

For example:

bob -p "Explain this project"

Basic usage
Providing prompts to Bob Shell

You can specify your prompt directly with the -p command:

bob -p "Explain this project"

Pipe content as input

You can pipe text content to Bob Shell:

cat buildError.txt | bob -p "Explain this build error"

Save results to a file

Redirect the output to save results:

bob -p "Review @bigFile.java" > review.md

Reference project files

Use the @ symbol to reference files in your project:

bob -p "Summarize the functionality in @src/main.js"

Advanced options
Enable file modifications

By default, Bob Shell only uses non-destructive tools (like reading files) in non-interactive session. To enable writing and updating files, add the --yolo flag:

bob -p "Fix bugs in @app.js" --yolo

    Note: Even with the --yolo flag enabled, Bob Shell will not write or update files outside the directory where it was started.

Format output for processing

The output contains both Bob Shell's answer and its thinking steps. For easier processing, add instructions to format the output:

bob -p "What Java version is this application using? Check @pom.xml. Enclose the answer in markdown tags" > analysis.md

When to use non-interactive session

Non-interactive session works best for:

    Integrating Bob Shell into automation scripts or CI/CD pipelines
    Processing multiple files with a single command
    Getting quick insights without starting an interactive session
    Generating documentation from code

Tips for effective use

    For complex tasks that might require multiple tool uses, interactive session usually works better.
    When processing large files or projects, be specific about which files to analyze.
    Use structured output instructions (like "Format the output as JSON") for easier parsing in scripts.
    Consider creating shell aliases or scripts for frequently used Bob Shell commands.
    For multi-line prompts, save them to a file and pipe to Bob Shell:

cat prompt.txt | bob