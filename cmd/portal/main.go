package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"www.github.com/ZinoKader/portal/constants"
	"www.github.com/ZinoKader/portal/models"
	"www.github.com/ZinoKader/portal/pkg/rendezvous"
	"www.github.com/ZinoKader/portal/tools"
)

type SendCommandOptions struct{}
type ReceiveCommandOptions struct{}
type AddCompletionsCommandOptions struct{}
type ServeCommandOptions struct{}

const SHELL_COMPLETION_SCRIPT = `_portal_completions() {
	args=("${COMP_WORDS[@]:1:$COMP_CWORD}")

	local IFS=$'\n'
	COMPREPLY=($(GO_FLAGS_COMPLETION=1 ${COMP_WORDS[0]} "${args[@]}"))
	return 1
}
complete -F _portal_completions portal
`

var (
	rootCmd = &cobra.Command{
		Use:   "portal",
		Short: "Portal is a quick and easy command-line file transfer utility from any computer to another.",
		Run:   func(cmd *cobra.Command, args []string) {},
	}
)

var programOptions struct {
	Verbose           string `short:"v" long:"verbose" optional:"true" optional-value:"no-file-specified" description:"Log detailed debug information (optional argument: specify output file with v=mylogfile or --verbose=mylogfile)"`
	RendezvousAddress string `short:"s" long:"server" description:"IP or hostname of the rendezvous server to use"`
	RendezvousPort    int    `short:"p" long:"port" description:"Port of the rendezvous server to use" default:"80"`
}

func init() {
	tools.RandomSeed()

	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(sendCmd)
	/*
		parser.AddCommand("send",
			"Send one or more files",
			"The send command adds one or more files to be sent. Files are archived and compressed before sending.",
			&sendCommand)

		parser.AddCommand("receive",
			"Receive files",
			"The receive command receives files from the sender with the matching password.",
			&receiveCommand)

		parser.AddCommand("serve",
			"Serve the Rendezvous server",
			"The serve command serves the Rendezvous server locally.",
			&serveCommand)

		parser.AddCommand("add-completions",
			"Add command line completions for bash and zsh",
			"The add-completions command adds command line completions to your shell. Uses the value from the $SHELL environment variable.",
			&addCompletionsCommand)

		parser.FindOptionByLongName("server").Default = []string{constants.DEFAULT_RENDEZVOUZ_ADDRESS}
	*/

	rootCmd.AddCommand(&cobra.Command{
		Use:   "receive",
		Short: "Receive files",
		Long:  "The receive command receives files from the sender with the matching password.",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 1 {
				fmt.Println("Provide a single password, for instance 1-cosmic-ray-quasar.")
				os.Exit(1)
			}
			if len(args) < 1 {
				fmt.Println("Provide the password that the file sender gave to you, for instance 1-galaxy-dust-aurora.")
				os.Exit(1)
			}
			handleReceiveCommand(models.ProgramOptions{RendezvousAddress: "1", RendezvousPort: 1}, args[0])
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "serve",
		Short: "Serve the rendezvous-server",
		Long:  "The serve command serves the rendezvous-server locally.",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 0 {
				fmt.Println("The serve command does not take any subcommands.")
				os.Exit(1)
			}
			server := rendezvous.NewServer(programOptions.RendezvousPort)
			server.Start()
		},
	})
}

func initConfig() {
	// Set default values
	viper.SetDefault("verbose", false)
	viper.SetDefault("rendezvousPort", constants.DEFAULT_RENDEZVOUS_PORT)
	viper.SetDefault("rendezvousAddress", constants.DEFAULT_RENDEZVOUS_ADDRESS)

	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// Search for config in home directory.
	viper.AddConfigPath(home)
	viper.SetConfigName(constants.CONFIG_FILE_NAME)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		// Create config file if not found
		//NOTE: perhaps should be an empty file initially, as we would not want defauy IP to be written to a file on the user host
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			configPath := filepath.Join(home, constants.CONFIG_FILE_NAME)
			configFile, err := os.Create(configPath)
			if err != nil {
				fmt.Println("Could not create config file:", err)
				os.Exit(1)
			}
			defer configFile.Close()
			_, err = configFile.Write([]byte(constants.DEFAULT_CONFIG_YAML))
			if err != nil {
				fmt.Println("Could not write defaults to config file:", err)
				os.Exit(1)
			}
		} else {
			fmt.Println("Could not read config file:", err)
			os.Exit(1)
		}
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// Execute is executed when the "send" command is invoked
// func (s *SendCommandOptions) Execute(args []string) error {
// 	if len(args) == 0 {
// 		return errors.New("No files provided. The send command takes file(s) delimited by spaces as arguments.")
// 	}

// 	err := validateRendezvousAddress()
// 	if err != nil {
// 		return err
// 	}

// 	if len(programOptions.Verbose) != 0 {
// 		logFileName := programOptions.Verbose
// 		if programOptions.Verbose == "no-file-specified" {
// 			logFileName = "portal-send.log"
// 		}
// 		f, err := tea.LogToFile(logFileName, "portal-send: ")
// 		if err != nil {
// 			return errors.New("Could not log to the provided file.")
// 		}
// 		defer f.Close()
// 	} else {
// 		log.SetOutput(io.Discard)
// 	}

// 	handleSendCommand(models.ProgramOptions{
// 		RendezvousAddress: programOptions.RendezvousAddress,
// 		RendezvousPort:    programOptions.RendezvousPort,
// 	}, args)
// 	return nil
// }

// // Execute is executed when the "receive" command is invoked
// func (r *ReceiveCommandOptions) Execute(args []string) error {
// 	if len(args) > 1 {
// 		return errors.New("Provide a single password, for instance 1-cosmic-ray-quasar.")
// 	}
// 	if len(args) < 1 {
// 		return errors.New("Provide the password that the file sender gave to you, for instance 1-galaxy-dust-aurora.")
// 	}

// 	err := validateRendezvousAddress()
// 	if err != nil {
// 		return err
// 	}

// 	if len(programOptions.Verbose) != 0 {
// 		logFileName := programOptions.Verbose
// 		if programOptions.Verbose == "no-file-specified" {
// 			logFileName = "portal-receive.log"
// 		}
// 		f, err := tea.LogToFile(logFileName, "portal-receive: ")
// 		if err != nil {
// 			return errors.New("Could not log to the provided file.")
// 		}
// 		defer f.Close()
// 	} else {
// 		log.SetOutput(io.Discard)
// 	}

// 	handleReceiveCommand(models.ProgramOptions{
// 		RendezvousAddress: programOptions.RendezvousAddress,
// 		RendezvousPort:    programOptions.RendezvousPort,
// 	}, args[0])
// 	return nil
// }

// // Execute is executed when the "add-completions" command is invoked
// func (a *AddCompletionsCommandOptions) Execute(args []string) error {
// 	shellBinPath := os.Getenv("SHELL")
// 	if len(shellBinPath) == 0 {
// 		return fmt.Errorf(
// 			"Completions not added - could not find which shell is used.\nTo add completions manually, add the following to your config:\n\n%s", SHELL_COMPLETION_SCRIPT)
// 	}

// 	shellPathComponents := strings.Split(os.Getenv("SHELL"), "/")
// 	usedShell := shellPathComponents[len(shellPathComponents)-1]
// 	if !tools.Contains([]string{"bash", "zsh"}, usedShell) {
// 		return fmt.Errorf("Unsupported shell \"%s\" at path: \"%s\".", usedShell, shellBinPath)
// 	}

// 	err := writeShellCompletionScript(usedShell)
// 	if err != nil {
// 		return fmt.Errorf("Failed when adding script to shell config file: %e", err)
// 	}

// 	fmt.Println("Successfully added completions to your shell config. Run 'source' on your shell config or restart your shell.")
// 	return nil
// }

// writeShellCompletionScript writes the completion script to the specified shell name
func writeShellCompletionScript(shellName string) error {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	shellConfigName := fmt.Sprintf(".%src", shellName)
	shellConfigPath := path.Join(homedir, shellConfigName)
	f, err := os.OpenFile(shellConfigPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.WriteString(fmt.Sprintf("\n# portal shell completion\n%s\n", SHELL_COMPLETION_SCRIPT)); err != nil {
		return err
	}

	return nil
}

func validateRendezvousAddress() error {
	rendezvouzAdress := net.ParseIP(programOptions.RendezvousAddress)
	err := tools.ValidateHostname(programOptions.RendezvousAddress)
	// neither a valid IP nor a valid hostname was provided
	if (rendezvouzAdress == nil) && err != nil {
		return errors.New("Invalid IP or hostname provided.")
	}
	return nil
}
