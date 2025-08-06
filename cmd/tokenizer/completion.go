package main

import (
	"os"

	"github.com/spf13/cobra"
)

// completionCmd represents the completion command.
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script",
	Long: `Generate shell completion script for tokenizer.

To load completions:

Bash:
  $ source <(tokenizer completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ tokenizer completion bash > /etc/bash_completion.d/tokenizer
  # macOS:
  $ tokenizer completion bash > $(brew --prefix)/etc/bash_completion.d/tokenizer

Zsh:
  $ source <(tokenizer completion zsh)
  # To load completions for each session, execute once:
  $ tokenizer completion zsh > "${fpath[1]}/_tokenizer"

Fish:
  $ tokenizer completion fish | source
  # To load completions for each session, execute once:
  $ tokenizer completion fish > ~/.config/fish/completions/tokenizer.fish

PowerShell:
  PS> tokenizer completion powershell | Out-String | Invoke-Expression
  # To load completions for every new session, run:
  PS> tokenizer completion powershell > tokenizer.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			return cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			return cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
