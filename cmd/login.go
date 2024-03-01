package cmd

import (
	"github.com/emilkje/cwc/pkg/config"
	"github.com/emilkje/cwc/pkg/errors"
	"github.com/emilkje/cwc/pkg/ui"
	"github.com/spf13/cobra"
)

const (
	serviceName = "cwc"
)

var (
	endpointFlag        string
	apiVersionFlag      string
	modelDeploymentFlag string
)

// Declaration of a new cobra command for login
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Azure OpenAI",
	Long: `Login will prompt you to enter your Azure OpenAI API key and other relevant information required for authentication. 
Your credentials will be stored securely in your keyring and will never be exposed on the file system directly.
You can also pass the API key through stdin by using the '-' argument making the command non-interactive. For example:

> echo $API_KEY | cwc login -e "https://my-deployment.openai.azure.com/" -v "2023-12-01-preview" -m "gpt-4-turbo" -
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// if the first argument is '-' or is empty, then we should read from stdin
		var apiKey string
		if len(args) == 0 {
			ui.PrintMessage("Enter your Azure OpenAI API Key: ", ui.MessageTypeInfo)
			apiKey = config.SanitizeInput(ui.ReadUserInput())
		} else if args[0] == "-" {
			reader := cmd.InOrStdin()
			var buffer []byte
			buffer = make([]byte, 1024)
			n, err := reader.Read(buffer)
			if err != nil {
				return err
			}
			apiKey = string(buffer[:n])
		}

		// Prompt for other required authentication details (endpoint, version, and deployment)
		endpoint := endpointFlag
		apiVersion := apiVersionFlag
		modelDeployment := modelDeploymentFlag

		if endpointFlag == "" {
			ui.PrintMessage("Enter the Azure OpenAI API Endpoint: ", ui.MessageTypeInfo)
			endpoint = config.SanitizeInput(ui.ReadUserInput())
		}

		if apiVersionFlag == "" {
			ui.PrintMessage("Enter the Azure OpenAI API Version: ", ui.MessageTypeInfo)
			apiVersion = config.SanitizeInput(ui.ReadUserInput())
		}

		if modelDeploymentFlag == "" {
			ui.PrintMessage("Enter the Azure OpenAI Model Deployment: ", ui.MessageTypeInfo)
			modelDeployment = config.SanitizeInput(ui.ReadUserInput())
		}

		cfg := config.NewConfig(endpoint, apiVersion, modelDeployment)
		cfg.SetAPIKey(apiKey)

		err := config.SaveConfig(cfg)
		if err != nil {
			if validationErr, ok := errors.AsConfigValidationError(err); ok {
				for _, e := range validationErr.Errors {
					ui.PrintMessage(e+"\n", ui.MessageTypeError)
				}
				return nil // suppress the error
			}
			return err
		}

		ui.PrintMessage("config saved successfully\n", ui.MessageTypeSuccess)

		return nil
	},
}

func init() {

	// Add flags to the login command
	loginCmd.Flags().StringVarP(&endpointFlag, "endpoint", "e", "", "Azure OpenAI API Endpoint")
	loginCmd.Flags().StringVarP(&apiVersionFlag, "api-version", "v", "", "Azure OpenAI API Version")
	loginCmd.Flags().StringVarP(&modelDeploymentFlag, "model-deployment", "m", "", "Azure OpenAI Model Deployment")
}
