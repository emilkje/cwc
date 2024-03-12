package cmd

import (
	"strings"

	"github.com/emilkje/cwc/pkg/config"
	"github.com/emilkje/cwc/pkg/errors"
	"github.com/emilkje/cwc/pkg/ui"
	"github.com/spf13/cobra"
)

var (
	apiKeyFlag          string
	endpointFlag        string
	apiVersionFlag      string
	modelDeploymentFlag string
	modelFlag           string
	providerFlag        string
)

// Declaration of a new cobra command for login
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with LLM provider",
	Long: `Login will prompt you to enter your provider name, API key and other relevant information required for authentication.
Your credentials will be stored securely in your keyring and will never be exposed on the file system directly. 
`,
	RunE: func(cmd *cobra.Command, args []string) error {

		// Prompt for other required authentication details (apiKey, endpoint, version, and deployment)
		apiKey := apiKeyFlag
		endpoint := endpointFlag
		apiVersion := apiVersionFlag
		modelDeployment := modelDeploymentFlag
		model := modelFlag
		provider := strings.ToLower(providerFlag)

		if provider == "" {
			ui.PrintMessage("Enter Provider name: ", ui.MessageTypeInfo)
			provider = config.SanitizeInput(ui.ReadUserInput())
		}

		if apiKeyFlag == "" {
			ui.PrintMessage("Enter the API Key: ", ui.MessageTypeInfo)
			apiKey = config.SanitizeInput(ui.ReadUserInput())
		}

		if endpointFlag == "" {
			ui.PrintMessage("Enter the API Endpoint: ", ui.MessageTypeInfo)
			endpoint = config.SanitizeInput(ui.ReadUserInput())
		}

		if apiVersionFlag == "" {
			ui.PrintMessage("Enter the API Version: ", ui.MessageTypeInfo)
			apiVersion = config.SanitizeInput(ui.ReadUserInput())
		}
		if provider == "azure" {
			if modelDeploymentFlag == "" {
				ui.PrintMessage("Enter the Azure OpenAI Model Deployment: ", ui.MessageTypeInfo)
				modelDeployment = config.SanitizeInput(ui.ReadUserInput())
			}
		}

		if provider == "openai" {
			if modelFlag == "" {
				ui.PrintMessage("Enter the Model name: ", ui.MessageTypeInfo)
				model = config.SanitizeInput(ui.ReadUserInput())
			}
		}

		cfg := config.NewConfig(provider, endpoint, apiVersion, modelDeployment, model)
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
	loginCmd.Flags().StringVarP(&providerFlag, "provider", "p", "", "Provider name. Supported providers: "+strings.Join(config.SupportedProviders, " "))
	loginCmd.Flags().StringVarP(&modelFlag, "model", "m", "", "OpenAI (compatible) Model")
	loginCmd.Flags().StringVarP(&apiKeyFlag, "api-key", "k", "", "API Key")
	loginCmd.Flags().StringVarP(&endpointFlag, "endpoint", "e", "", "API Endpoint")
	loginCmd.Flags().StringVarP(&apiVersionFlag, "api-version", "v", "", "API Version")
	loginCmd.Flags().StringVarP(&modelDeploymentFlag, "model-deployment", "d", "", "Azure OpenAI Model Deployment")
}
