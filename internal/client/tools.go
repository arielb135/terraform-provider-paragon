package client

import (
    "regexp"
    "strings"
    "encoding/json"
    "net/http"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "io/ioutil"
)

type WebhookBody struct {
    DataType string     `json:"dataType"`
    Type     string     `json:"type"`
    Parts    []BodyPart `json:"parts"`
}

type BodyPart struct {
    DataType string   `json:"dataType,omitempty"`
    Type     string   `json:"type"`
    Value    string   `json:"value,omitempty"`
    Path     []string `json:"path,omitempty"`
    Name     string   `json:"name,omitempty"`
}

func ConvertToWebhookAPIFormat(input string) (*WebhookBody, error) {
    tokenRegex := regexp.MustCompile(`{{\$\.(.*?)}}`)
    var parts []BodyPart
    lastIndex := 0

    for _, match := range tokenRegex.FindAllStringSubmatchIndex(input, -1) {
        beforeToken := input[lastIndex:match[0]]
        parts = append(parts, BodyPart{
            DataType: "STRING",
            Type:     "VALUE",
            Value:    beforeToken,
        })

        tokenContent := input[match[2]:match[3]]
        path := strings.Split(tokenContent, ".")
        parts = append(parts, BodyPart{
            Type: "OBJECT_VALUE",
            Path: path,
            Name: path[0],
        })

        lastIndex = match[1]
    }

    if lastIndex < len(input) {
        remaining := input[lastIndex:]
        parts = append(parts, BodyPart{
            DataType: "STRING",
            Type:     "VALUE",
            Value:    remaining,
        })
    }

    apiBody := WebhookBody{
        DataType: "ANY",
        Type:     "TOKENIZED",
        Parts:    parts,
    }

    return &apiBody, nil
}

func escapeNewlines(input string) string {
    return strings.ReplaceAll(input, "\n", "\\n")
}

func ConvertPartsToString(body WebhookBody) string {
	var result strings.Builder

	for _, part := range body.Parts {
		if part.Type == "VALUE" {
			result.WriteString(part.Value)
		} else if part.Type == "OBJECT_VALUE" {
			if len(part.Path) > 0 {
				result.WriteString("{{$.")
				result.WriteString(strings.Join(part.Path, "."))
				result.WriteString("}}")
			}
		}
	}

	return result.String()
}

func ConvertStringSliceToTypesStringSlice(slice []string) []types.String {
    result := make([]types.String, len(slice))
    for i, value := range slice {
        result[i] = types.StringValue(value)
    }
    return result
}


func formatErrorMessage(resp *http.Response) string {
   body, err := ioutil.ReadAll(resp.Body)
   if err != nil {
   	return "Error occurred"
   }
   defer resp.Body.Close()

   var errorResponse struct {
   	Message string `json:"message"`
   	Meta    struct {
   		Errors []struct {
   			Error string `json:"error"`
   		} `json:"errors"`
   	} `json:"meta"`
   }

   err = json.Unmarshal(body, &errorResponse)
   if err != nil {
   	return "Error occurred"
   }

   var errorMessage string
   if errorResponse.Message != "" {
   	errorMessage = errorResponse.Message
   } else {
   	errorMessage = "Error occurred"
   }

   if len(errorResponse.Meta.Errors) > 0 {
   	var errorMessages []string
   	for _, err := range errorResponse.Meta.Errors {
   		if err.Error != "" {
   			errorMessages = append(errorMessages, err.Error)
   		}
   	}

   	if len(errorMessages) > 0 {
   		errorMessage += ", " + strings.Join(errorMessages, ", ")
   	}
   }

   // Check for specific CONNECT_CREDENTIAL_FIELD error and provide clearer message
   if resp.StatusCode == 400 && strings.Contains(errorMessage, "CONNECT_CREDENTIAL_FIELD") && strings.Contains(errorMessage, "no ConnectCredential was supplied") {
   	return "Deploy failed - userSettings field was used in the workflow meaning - you MUST go to the integration -> Test Connect Portal and perform a one time connection through the connect portal wizard, then rerun this - This is a one time manual task needed to be done."
   }

   return errorMessage
}