// SPDX-License-Identifier: Apache-2.0

// Package completions provides shell-completion scripts for the glacier binary.
// Supported shells: bash, zsh, fish, pwsh (PowerShell).
// Each script enumerates all nine glacier commands.
package completions

import "strings"

// commands lists all nine glacier sub-commands in the order they appear in help.
var commands = []string{
	"vibe",
	"version",
	"generate",
	"lint",
	"test",
	"init",
	"new",
	"completions",
	"explain",
}

// Bash returns the bash completion script for the glacier binary.
func Bash() string {
	cmds := strings.Join(commands, " ")
	return `# glacier bash completion
_glacier_completions() {
    local cur="${COMP_WORDS[COMP_CWORD]}"
    local commands="` + cmds + `"
    if [[ ${COMP_CWORD} -eq 1 ]]; then
        COMPREPLY=($(compgen -W "${commands}" -- "${cur}"))
    fi
}
complete -F _glacier_completions glacier
`
}

// Zsh returns the zsh completion script for the glacier binary.
func Zsh() string {
	var b strings.Builder
	b.WriteString("# glacier zsh completion\n")
	b.WriteString("#compdef glacier\n\n")
	b.WriteString("_glacier() {\n")
	b.WriteString("    local -a subcmds\n")
	b.WriteString("    subcmds=(\n")
	for _, cmd := range commands {
		b.WriteString("        '" + cmd + "'\n")
	}
	b.WriteString("    )\n")
	b.WriteString("    _describe 'glacier commands' subcmds\n")
	b.WriteString("}\n\n")
	b.WriteString("_glacier \"$@\"\n")
	return b.String()
}

// Fish returns the fish completion script for the glacier binary.
func Fish() string {
	var b strings.Builder
	b.WriteString("# glacier fish completion\n")
	for _, cmd := range commands {
		b.WriteString("complete -c glacier -f -n '__fish_use_subcommand' -a '" + cmd + "'\n")
	}
	return b.String()
}

// Pwsh returns the PowerShell completion script for the glacier binary.
func Pwsh() string {
	cmds := "'" + strings.Join(commands, "', '") + "'"
	return `# glacier PowerShell completion
Register-ArgumentCompleter -Native -CommandName glacier -ScriptBlock {
    param($wordToComplete, $commandAst, $cursorPosition)
    $commands = @(` + cmds + `)
    $commands | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
        [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
    }
}
`
}

// Script returns the completion script for the named shell.
// Recognised names: "bash", "zsh", "fish", "pwsh" (case-insensitive).
// Returns ("", false) for unrecognised shells.
func Script(shell string) (string, bool) {
	switch strings.ToLower(shell) {
	case "bash":
		return Bash(), true
	case "zsh":
		return Zsh(), true
	case "fish":
		return Fish(), true
	case "pwsh", "powershell":
		return Pwsh(), true
	default:
		return "", false
	}
}
